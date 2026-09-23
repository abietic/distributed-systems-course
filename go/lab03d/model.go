// Lab 3-D 的四个模型。
//
// 这一章和 3-A/3-B/3-C 不同：那三章要跑完整的 Raft 状态机，
// 因为它们研究的是【算法本身】会不会出错。
// 3-D 研究的是【算法与系统之间的接缝】——
// 读路径怎么和状态机对齐、新 Leader 的 commitIndex 从哪来、
// quorum 和日志复制到底差在哪、配置从哪来。
// 这些问题用一个聚焦的小模型讲得更清楚，所以这里不复用 lab03c 的 raft.go。
package main

import (
	"fmt"
	"sort"
)

// ═══════════════════════════════════════════════════════════════════════
// 模型 1：ReadIndex 的第三步
// ═══════════════════════════════════════════════════════════════════════

// ReadResult 记录一次读的结果。
type ReadResult struct {
	Got    int // 读到的值（= 该 index 处写入的值；0 表示还没有值）
	Want   int // 此刻已提交的最新值
	Waited int // 第三步阻塞等了多少条 apply
	Stale  bool
}

// LeaderState 是一个被极度简化的 Leader：
// 我们只保留这条读路径真正依赖的三个量。
type LeaderState struct {
	commitIndex int
	lastApplied int
}

// DoRead 执行一次 ReadIndex 读。
//
// 三步全在这个函数里，一步不少：
//
//	step1  readIndex := commitIndex
//	step2  发一轮心跳确认自己还是 Leader          （本模型里恒成功——我们要隔离出第三步）
//	step3  等 lastApplied >= readIndex 再读状态机  ← waitApply 控制开不开
//
// 之所以让 step2 恒成功，是为了证明一件事：
// 【前两步全部正确通过，读出来依然可能是旧数据。】
func (l *LeaderState) DoRead(waitApply bool) ReadResult {
	readIndex := l.commitIndex // step 1
	// step 2：心跳确认（本模型中无分区，恒通过）
	r := ReadResult{Want: readIndex}
	if waitApply {
		// step 3：阻塞，直到状态机追上 readIndex
		r.Waited = readIndex - l.lastApplied
		if r.Waited < 0 {
			r.Waited = 0
		}
		l.lastApplied = readIndex
		r.Got = readIndex
	} else {
		// 跳过 step 3，直接读状态机当前的内容
		r.Got = l.lastApplied
	}
	r.Stale = r.Got != r.Want
	return r
}

// ═══════════════════════════════════════════════════════════════════════
// 模型 2：no-op —— 新 Leader 上任后 commitIndex 从哪来
// ═══════════════════════════════════════════════════════════════════════

// NoopOutcome 是一轮场景的结果。
type NoopOutcome int

const (
	OutcomeFresh   NoopOutcome = iota // 返回了最新的已提交值
	OutcomeStale                      // 读到旧值，而客户端收到过「写入成功」⇒ 线性一致性破裂
	OutcomeQueued                     // no-op 还没提交，读被挂起；提交后再处理，返回最新值 ⇒ 安全
	OutcomeAllowed                    // 返回旧值，但那次写从未 ack ⇒ 允许（Part 0 的「第三态」）
)

// RunNoop 复现这样一个场景：
//
//	任期 2：Leader L 把 k 条写复制到了过半节点。
//	        ① 有时它来得及推进自己的 commitIndex 并【回复客户端成功】，然后才崩；
//	        ② 有时它还没来得及推进 commitIndex 就崩了。
//	        无论哪种，它都【没来得及广播 leaderCommit】——
//	        所以「已提交」这个事实随它一起消失。
//	任期 3：某个持有这些日志的节点当选。它的 commitIndex = 0。
//	        受 Figure 8 限制，它【不能】直接提交任期 2 的条目。
//	然后客户端来读。
//
// ackedBeforeCrash 对应上面的 ①：这是唯一会造成真违规的分支。
// noop 决定新 Leader 上任后发不发空条目。
// readEarly 表示读请求到得很早（no-op 还没提交完）。
func RunNoop(k int, ackedBeforeCrash, noop, readEarly bool) NoopOutcome {
	// 客户端认为的最新值：只有 ① 分支里它才收到过成功
	var acked int
	if ackedBeforeCrash {
		acked = k
	}

	newLeaderCommit := 0 // 新 Leader 的 commitIndex —— 它是易失的，不会从旧 Leader 传过来

	if noop {
		if readEarly {
			// no-op 还在复制中。etcd raft 此时把读请求【挂起排队】：
			//   if !r.committedEntryInCurrentTerm() {
			//       r.pendingReadIndexMessages = append(r.pendingReadIndexMessages, m)
			//   }
			// 等本任期第一条提交后再放行 —— 那时 commitIndex 已经到位，返回的就是最新值。
			// （也有实现选择直接拒绝、让客户端重试；两种都安全，区别只在延迟。）
			return OutcomeQueued
		}
		// no-op 是【当前任期】的条目 ⇒ 可以直接提交；
		// 提交是前缀性的 ⇒ 任期 2 的 k 条被顺带提交。
		newLeaderCommit = k + 1
	}
	// 不发 no-op：没有新写入进来，commitIndex 就永远停在 0。

	// ReadIndex：step1 拿 commitIndex，step2 心跳确认（通过），step3 等 apply（本模型瞬时）
	readIndex := newLeaderCommit
	got := readIndex
	if got > k {
		got = k // no-op 本身没有值，真正的最新值是第 k 条
	}

	if got == k {
		return OutcomeFresh // 返回了最新的已提交值
	}
	if acked > got {
		// 客户端收到过「x=acked 写入成功」，现在却读不到它 ⇒ 线性一致性破裂
		return OutcomeStale
	}
	// 返回旧值，但那次写从来没有 ack 过 ⇒ 它的结果本就是「未知」，返回旧值是允许的
	return OutcomeAllowed
}

// ═══════════════════════════════════════════════════════════════════════
// 模型 3：Dynamo 式 Quorum vs Raft 日志复制
// ═══════════════════════════════════════════════════════════════════════

// QuorumOutcome 描述一轮并发写之后，N 个副本手里到底是什么。
type QuorumOutcome struct {
	Replicas []int // 每个副本最终保存的值（-1 = 什么都没收到）
	Distinct int   // 副本间不同的最终值有几种
	ReadVers int   // 读 R 个副本拿回几个不同版本
	WriteOK  int   // 有几个写收到了 >= W 个 ack
}

// RunDynamo 模拟 c 个客户端【同时】写同一个 key。
//
// 每个写都发往全部 n 个副本（Dynamo 的做法），收到 w 个 ack 就返回成功。
// 关键在于：各副本收到这些写的【顺序是不同的】，而副本本身不带版本信息，
// 于是每个副本保存的是「最后到达的那个」。
// 这就是为什么并发写会在副本间留下分歧——和 W 设多大毫无关系。
func RunDynamo(n, w, r, c int, loss float64, rnd *Rand) QuorumOutcome {
	deliv := make([][]bool, c)
	for i := range deliv {
		deliv[i] = make([]bool, n)
		for j := range deliv[i] {
			deliv[i][j] = rnd.Float() > loss
		}
	}
	out := QuorumOutcome{Replicas: make([]int, n)}
	for i := 0; i < c; i++ {
		acks := 0
		for j := 0; j < n; j++ {
			if deliv[i][j] {
				acks++
			}
		}
		if acks >= w {
			out.WriteOK++
		}
	}
	for j := 0; j < n; j++ {
		got := []int{}
		for i := 0; i < c; i++ {
			if deliv[i][j] {
				got = append(got, i)
			}
		}
		rnd.Shuffle(len(got), func(a, b int) { got[a], got[b] = got[b], got[a] })
		if len(got) == 0 {
			out.Replicas[j] = -1
		} else {
			out.Replicas[j] = got[len(got)-1] // 最后到达的胜出
		}
	}
	seen := map[int]bool{}
	for _, v := range out.Replicas {
		if v >= 0 {
			seen[v] = true
		}
	}
	out.Distinct = len(seen)
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	rnd.Shuffle(n, func(a, b int) { idx[a], idx[b] = idx[b], idx[a] })
	rs := map[int]bool{}
	for i := 0; i < r && i < n; i++ {
		if v := out.Replicas[idx[i]]; v >= 0 {
			rs[v] = true
		}
	}
	out.ReadVers = len(rs)
	return out
}

// RunRaft 把同样的 c 个并发写灌进 Raft。
//
// 它们全都先到 Leader，Leader 按到达先后分配 index —— 顺序在写入之前就定死了。
// 复制过程中的丢包靠重试和 nextIndex 回退补齐，最终每个副本的日志完全相同。
// 所以这个函数不需要随机：分歧数恒为 1，这不是概率，是结构。
func RunRaft(n, c int, rnd *Rand) QuorumOutcome {
	order := make([]int, c)
	for i := range order {
		order[i] = i
	}
	rnd.Shuffle(c, func(a, b int) { order[a], order[b] = order[b], order[a] })
	winner := order[c-1]
	out := QuorumOutcome{Replicas: make([]int, n), Distinct: 1, ReadVers: 1, WriteOK: c}
	for i := range out.Replicas {
		out.Replicas[i] = winner
	}
	return out
}

// ═══════════════════════════════════════════════════════════════════════
// 模型 4：冷启动 —— 配置从哪来
// ═══════════════════════════════════════════════════════════════════════

// NodeCfg 是一个节点的启动参数，以及它 data dir 的状态。
type NodeCfg struct {
	ID      string
	Cluster []string // --initial-cluster 里的成员列表
	State   string   // --initial-cluster-state: new | existing
	Token   string   // --initial-cluster-token
	HasData bool     // data dir 里已有 WAL/快照 —— 有的话，所有 --initial-* 参数都被忽略
	Wiped   bool     // 曾经是集群成员，但 data dir 被清空了（currentTerm / votedFor / log 一起丢了）
	Term    int      // data dir 里记录的 currentTerm（HasData 时才有意义）
}

// Step 是运维按顺序做的一个动作。
type Step struct {
	Kind string // start：启动节点 | add：在集群上执行 member add | write：集群确认了 N 条写
	Node string
	N    int
}

// Cluster 是实际形成的一个 Raft 集群。
type Cluster struct {
	// Key 是集群身份。etcd 里它是【确定性】算出来的：
	//   成员 ID = hash(排序后的 peer URL + --initial-cluster-token)   （静态引导时不带时间戳）
	//   集群 ID = hash(所有成员 ID)
	// 所以「同一个 token + 同一份成员列表」⇒ 同一个集群 ID。
	// 之后 member add 会改变成员配置，但【不会】改变集群 ID。
	Key     string
	Members []string        // 投票成员配置
	Started map[string]bool // 已经启动的成员
	Term    int             // 当前任期
	Reset   bool            // 是否由「清空 data dir 后重新引导」的节点建起来的 —— 任期从头开始
	Writes  int             // 在当前这个任期里确认过的写
	Winner  string          // 如果有旧节点回归并当选，记录是谁
	Idx     int             // 创建顺序，只用于打印
}

func (c *Cluster) started() int { return len(c.Started) }
func (c *Cluster) need() int    { return len(c.Members)/2 + 1 }
func (c *Cluster) Leader() bool { return c.started() >= c.need() }
func (c *Cluster) StartedList() []string {
	out := []string{}
	for _, m := range c.Members {
		if c.Started[m] {
			out = append(out, m)
		}
	}
	return out
}

// BootResult 是一个场景跑完之后的样子。
type BootResult struct {
	Clusters []*Cluster
	Refused  []string // 被 etcd 拒绝启动的节点
	Rejected []string // 被拒绝的 member add
	Lost     int      // 被覆盖掉的、已经确认过的写
	Log      []string // 每一步发生了什么
}

func clusterKey(token string, list []string) string {
	cl := append([]string{}, list...)
	sort.Strings(cl)
	return token + "/" + join(cl, ",")
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string{}, a...)
	y := append([]string{}, b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// Bootstrap 按顺序执行一串运维动作，模拟 etcd 实际会做的判断。
//
// 三条规则（都来自 etcd 源码，见课件 §3.28）：
//  1. data dir 里有数据 ⇒ --initial-* 全部忽略，按 data dir 回到原来的集群。
//  2. state=new 且 data dir 为空 ⇒ 先问列表里的其他成员「我是不是已经被引导过」
//     （isMemberBootstrapped）。有知情者说「是」⇒ 拒绝启动；问不到人 ⇒ 放行，按参数引导。
//  3. state=existing 且 data dir 为空 ⇒ 不引导任何东西，去 peer 那里拉当前成员配置，
//     和自己的 --initial-cluster 比对（ValidateClusterAndAssignIDs），对得上才加入。
//     Raft 启动时【不带初始成员】（RestartNode），配置等 Leader 发过来。
func Bootstrap(nodes []NodeCfg, steps []Step) BootResult {
	cfg := map[string]NodeCfg{}
	for _, n := range nodes {
		cfg[n.ID] = n
	}
	byKey := map[string]*Cluster{}
	var r BootResult
	logf := func(f string, a ...any) { r.Log = append(r.Log, fmt.Sprintf(f, a...)) }
	getOrCreate := func(key string, members []string) *Cluster {
		if c, ok := byKey[key]; ok {
			return c
		}
		c := &Cluster{Key: key, Members: append([]string{}, members...), Started: map[string]bool{},
			Idx: len(r.Clusters) + 1}
		byKey[key] = c
		r.Clusters = append(r.Clusters, c)
		return c
	}
	clusterOf := func(id string) *Cluster {
		for _, c := range r.Clusters {
			if c.Started[id] {
				return c
			}
		}
		return nil
	}
	// 一个集群刚凑齐过半时选出 Leader：由清空后重新引导的节点建起来的，任期从 1 开始（引导条目），
	// 第一次选举是任期 2 —— 而原来那个集群早就走到了任期 12。这就是「任期倒退」。
	elect := func(c *Cluster, before bool) {
		if !before && c.Leader() && c.Term == 0 {
			c.Term = 2
			logf("        ⇒ %s 凑齐过半（%d/%d），选出 Leader，任期 %d", c.label(), c.started(), len(c.Members), c.Term)
		}
	}

	for _, st := range steps {
		n := cfg[st.Node]
		switch st.Kind {
		case "start":
			switch {
			case n.HasData:
				key := clusterKey(n.Token, n.Cluster)
				c := getOrCreate(key, n.Cluster)
				before := c.Leader()
				c.Started[n.ID] = true
				logf("启动 %s：data dir 有数据（任期 %d）⇒ 忽略 --initial-*，回到原集群", n.ID, n.Term)
				if c.Reset && c.Leader() && n.Term > c.Term {
					// 旧节点回归：它的任期更高 ⇒ 现任 Leader 收到它的回复后立刻下台；
					// 它的日志最后一条任期也更高 ⇒ 它当选；然后把任期倒退期间写下的东西全部覆盖。
					logf("        ⇒ %s 的任期 %d > 现任 Leader 的 %d ⇒ 现任 Leader 下台", n.ID, n.Term, c.Term)
					logf("        ⇒ %s 的日志更新（最后一条任期 %d > %d）⇒ 当选，覆盖其他节点的日志", n.ID, n.Term, c.Term)
					if c.Writes > 0 {
						logf("        ⇒ 任期 %d 里确认过的 %d 条写【全部被覆盖】", c.Term, c.Writes)
						r.Lost += c.Writes
					}
					c.Term = n.Term + 1
					c.Writes = 0
					c.Winner = n.ID
				} else if !before && c.Leader() {
					if c.Term == 0 {
						c.Term = n.Term + 1
					}
					logf("        ⇒ %s 凑齐过半（%d/%d），选出 Leader", c.label(), c.started(), len(c.Members))
				} else if !c.Leader() {
					logf("        ⇒ %s 只有 %d/%d 在线，凑不齐过半，无法选主", c.label(), c.started(), len(c.Members))
				}

			case n.State == "new":
				key := clusterKey(n.Token, n.Cluster)
				// isMemberBootstrapped：问列表里的其他成员，有没有谁记得我已经被引导过
				if c, ok := byKey[key]; ok {
					knows := ""
					for _, m := range c.Members {
						if m != n.ID && c.Started[m] && cfg[m].HasData {
							knows = m
							break
						}
					}
					if knows != "" {
						logf("启动 %s（state=new）：先问 %s「我是不是已经被引导过」⇒ %s 说是", n.ID, knows, knows)
						logf("        ⇒ 拒绝启动：member %s has already been bootstrapped", n.ID)
						r.Refused = append(r.Refused, n.ID)
						continue
					}
				}
				_, existed := byKey[key]
				others := len(r.Clusters)
				c := getOrCreate(key, n.Cluster)
				if n.Wiped {
					c.Reset = true
				}
				before := c.Leader()
				c.Started[n.ID] = true
				if len(n.Cluster) == 1 {
					logf("启动 %s（state=new，列表只有自己）⇒ 引导一个单成员集群 %s", n.ID, c.label())
				} else {
					logf("启动 %s（state=new）：列表里没有能证明它被引导过的在线成员 ⇒ 按参数引导 %s", n.ID, c.label())
				}
				if !existed && others > 0 {
					logf("        ⇒ token 或成员列表和已有的集群不同 ⇒ 成员 ID、集群 ID 都不同 ⇒ 这是另一个集群，两边互不承认")
				}
				elect(c, before)
				if before && c.Leader() {
					logf("        ⇒ 加入，%d/%d 在线", c.started(), len(c.Members))
				}
				if !c.Leader() {
					logf("        ⇒ %s 只有 %d/%d 在线，凑不齐过半，无法选主", c.label(), c.started(), len(c.Members))
				}

			case n.State == "existing":
				var c *Cluster
				for _, p := range n.Cluster {
					if p != n.ID {
						if cc := clusterOf(p); cc != nil {
							c = cc
							break
						}
					}
				}
				if c == nil {
					logf("启动 %s（state=existing）：列表里的 peer 都连不上 ⇒ 拒绝启动：cannot fetch cluster info", n.ID)
					r.Refused = append(r.Refused, n.ID)
					continue
				}
				if !sameSet(n.Cluster, c.Members) {
					logf("启动 %s（state=existing）：自己的列表 %v ≠ 集群当前成员 %v", n.ID, n.Cluster, c.Members)
					if len(n.Cluster) != len(c.Members) {
						logf("        ⇒ 拒绝启动：member count is unequal")
					} else {
						logf("        ⇒ 拒绝启动：PeerURLs: no match found for existing member")
					}
					r.Refused = append(r.Refused, n.ID)
					continue
				}
				before := c.Leader()
				c.Started[n.ID] = true
				logf("启动 %s（state=existing）：从 peer 拉到成员配置 %v，与自己的参数一致 ⇒ 加入", n.ID, c.Members)
				logf("        ⇒ Raft 启动时不带初始成员，配置随 Leader 的快照/日志到达（%d/%d 在线）", c.started(), len(c.Members))
				if !before && c.Leader() {
					logf("        ⇒ 重新凑齐过半，恢复服务")
				}
			}

		case "add":
			var c *Cluster
			for _, cc := range r.Clusters {
				if cc.Leader() {
					c = cc
					break
				}
			}
			if c == nil {
				logf("member add %s ⇒ 没有能服务的 Leader，配置变更提交不了", st.Node)
				r.Rejected = append(r.Rejected, st.Node)
				continue
			}
			// IsReadyToAddVotingMember：加完之后，已启动的成员必须还能凑齐过半
			nmembers, nstarted := len(c.Members)+1, c.started()
			if !(nstarted == 1 && nmembers == 2) && nstarted < nmembers/2+1 {
				logf("member add %s ⇒ 拒绝：加完后成员 %d、已启动 %d < 过半 %d（ErrNotEnoughStartedMembers）",
					st.Node, nmembers, nstarted, nmembers/2+1)
				r.Rejected = append(r.Rejected, st.Node)
				continue
			}
			c.Members = append(c.Members, st.Node)
			logf("member add %s ⇒ 一条配置变更日志，提交后成员变为 %v（集群 ID 不变）", st.Node, c.Members)
			if !c.Leader() {
				logf("        ⇒ 现在过半要 %d 个、在线只有 %d 个：在 %s 启动之前集群停写（etcd 对 1→2 特许这种情况）",
					c.need(), c.started(), st.Node)
			}

		case "write":
			for _, c := range r.Clusters {
				if c.Leader() {
					c.Writes += st.N
					logf("客户端写入 %d 条 ⇒ %s 在任期 %d 提交并【确认】", st.N, c.label(), c.Term)
					break
				}
			}
		}
	}
	return r
}

func (c *Cluster) label() string {
	return "集群#" + itoa(c.Idx) + "{" + join(c.Members, ",") + "}"
}
