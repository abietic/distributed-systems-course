package main

import (
	"fmt"
	"math/rand"
	"strings"
)

// ════════════════════════════════════════════════════════════════════════
// Raft 日志复制（在 Lab 3-A 的选举之上）
//
//	在 Lab 3-B 的基础上再加四件事（Part 3-C）：
//	  ① 持久化：currentTerm / votedFor / log 落盘，崩溃重启后恢复
//	  ② 快照：日志压缩 + InstallSnapshot
//	  ③ 成员变更：每个节点持有自己的 config，切换时刻可以不同
//	  ④ 三种读：本地读 / ReadIndex / Lease Read
//
// ════════════════════════════════════════════════════════════════════════

type Entry struct {
	Index, Term int
	Cmd         string
}

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

func (s State) String() string { return [...]string{"Follower", "Candidate", "Leader"}[s] }

const (
	HeartbeatMs = 100
	NoVote      = -1
)

type Node struct {
	id       int
	state    State
	term     int
	votedFor int
	votes    map[int]bool
	timer    int64
	timeout  int64
	hb       int64

	// —— 日志相关 ——
	log         []Entry // log[0] 是哨兵 {Index:0, Term:0}
	commitIndex int
	lastApplied int
	applied     map[int]string // 状态机：index -> 已执行的命令

	nextIndex  []int
	matchIndex []int

	// —— Part 3-C ——
	disk      Persisted         // 模拟磁盘：只有这三样能挺过重启
	snapIdx   int               // lastIncludedIndex
	snapTerm  int               // lastIncludedTerm
	snapshot  map[string]string // 状态机镜像
	kv        map[string]string // 状态机本体
	config    map[int]bool      // 本节点当前认为的集群成员
	leaseTill int64             // Lease Read 的租约到期时刻
	ackTicks  int               // 本轮心跳收到的响应数（用于 CheckQuorum / 续租）
}

// Persisted 就是 Raft 论文 Figure 2 里「Persistent state on all servers」那三样。
// 其余状态（commitIndex / lastApplied / nextIndex / matchIndex）都是易失的。
type Persisted struct {
	CurrentTerm int
	VotedFor    int
	Log         []Entry
}

func (n *Node) last() Entry { return n.log[len(n.log)-1] }

// at 按【逻辑 index】取条目。有快照之后数组下标不再等于日志 index，
// 所有访问都必须走它，不能再直接 n.log[i]。
func (n *Node) at(i int) (Entry, bool) {
	e := n.entryAt(i)
	if e.Index != i {
		return Entry{}, false
	}
	return e, true
}

type Cluster struct {
	nodes  []*Node
	net    *Network
	rnd    *rand.Rand
	base   int64
	window int64

	NaiveCommit bool // ★ 关掉「只提交当前任期」的限制 ⇒ 复现 Figure 8
	SlowBackoff bool // 关掉快速回退，每次只退 1
	NoLogCheck  bool // ★ 关掉「日志至少一样新」的投票条件 ⇒ 退化成 Lab 3-A 的模型

	// —— Part 3-C：故意"丢失"某一样持久化状态 ——
	LoseTerm bool
	LoseVote bool
	LoseLog  bool

	SnapThreshold int   // 日志超过这么多条就做快照；0 表示不做
	LeaseMs       int64 // Lease Read 的租约时长

	StaleReads, ReadErrors, GoodReads int
	Snapshots, InstallSnaps           int
	twoLeadersEver                    bool

	// —— 上帝视角：全局已提交的日志。任何节点提交一条就登记在这里 ——
	globalCommitted map[int]Entry

	Elections, Splits, Rejects, Appends int
	Violations                          []string
	Log                                 []string
	Verbose                             bool
	cmdSeq                              int
}

func New(n int, net *Network, base, window, seed int64) *Cluster {
	c := &Cluster{net: net, rnd: rand.New(rand.NewSource(seed + 999)),
		base: base, window: window, globalCommitted: map[int]Entry{}}
	for i := 0; i < n; i++ {
		cfg := map[int]bool{}
		for j := 0; j < n; j++ {
			cfg[j] = true
		}
		nd := &Node{id: i, state: Follower, votedFor: NoVote, votes: map[int]bool{},
			log: []Entry{{Index: 0, Term: 0}}, applied: map[int]string{},
			nextIndex: make([]int, n), matchIndex: make([]int, n),
			kv: map[string]string{}, config: cfg, snapshot: map[string]string{}}
		nd.persist()
		nd.timeout = c.newTimeout()
		c.nodes = append(c.nodes, nd)
	}
	return c
}

func (c *Cluster) newTimeout() int64 {
	if c.window <= 0 {
		return c.base
	}
	return c.base + c.rnd.Int63n(c.window)
}
func (c *Cluster) majority() int { return len(c.nodes)/2 + 1 }

func (c *Cluster) logf(f string, a ...any) {
	s := fmt.Sprintf("[%6dms] ", c.net.Now()) + fmt.Sprintf(f, a...)
	c.Log = append(c.Log, s)
	if c.Verbose {
		fmt.Println("    " + s)
	}
}

func (c *Cluster) violate(f string, a ...any) {
	v := fmt.Sprintf(f, a...)
	for _, x := range c.Violations {
		if x == v {
			return
		}
	}
	if len(c.Violations) < 8 {
		c.Violations = append(c.Violations, v)
	}
}

// ─── 选举（3-A 的逻辑 + 「日志至少一样新」这条投票条件）────────────────────
func (c *Cluster) stepDown(n *Node, term int) {
	n.term, n.state, n.votedFor = term, Follower, NoVote
	n.votes = map[int]bool{}
	n.timer, n.timeout = 0, c.newTimeout()
	n.persist() // ★ 响应任何 RPC 之前必须落盘
}

func (c *Cluster) startElection(n *Node) {
	n.term++
	n.state = Candidate
	n.votedFor = n.id
	n.votes = map[int]bool{n.id: true}
	n.timer, n.timeout = 0, c.newTimeout()
	n.persist() // ★ 先落盘，再发 RequestVote
	c.Elections++
	last := n.last()
	for j := range c.nodes {
		if j != n.id {
			c.net.Send(Msg{From: n.id, To: j, Type: MsgRequestVote, Term: n.term,
				LastLogIndex: last.Index, LastLogTerm: last.Term})
		}
	}
}

// upToDate 实现「candidate 的日志至少和我一样新」：先比 term，再比长度。
func upToDate(candTerm, candIdx int, mine Entry) bool {
	if candTerm != mine.Term {
		return candTerm > mine.Term
	}
	return candIdx >= mine.Index
}

func (c *Cluster) becomeLeader(n *Node) {
	n.state = Leader
	n.hb = HeartbeatMs
	for j := range c.nodes {
		n.nextIndex[j] = n.last().Index + 1 // 乐观猜测
		n.matchIndex[j] = 0                 // 保守事实
	}
	n.matchIndex[n.id] = n.last().Index
	c.logf("★ N%d 当选 Leader，term=%d，日志长度=%d", n.id+1, n.term, n.last().Index)

	// Leader 完整性断言：新 Leader 必须包含所有已提交的日志
	for idx, e := range c.globalCommitted {
		got, ok := n.at(idx)
		if !ok || got.Term != e.Term || got.Cmd != e.Cmd {
			c.violate("【Leader 完整性被破坏】N%d 在 term=%d 当选，但它缺少已提交的 index=%d (term=%d, cmd=%s)",
				n.id+1, n.term, idx, e.Term, e.Cmd)
		}
	}
}

// ─── 日志复制 ──────────────────────────────────────────────────────────
func (c *Cluster) sendAppend(n *Node, to int) {
	ni := n.nextIndex[to]
	if ni < 1 {
		ni = 1
	}
	// ★ 要发的日志已经被压缩掉了 ⇒ 改发快照
	if ni <= n.snapIdx {
		c.sendSnapshot(n, to)
		return
	}
	prev, _ := n.at(ni - 1)
	var ents []Entry
	if ni <= n.last().Index {
		base := n.log[0].Index
		ents = append(ents, n.log[ni-base:]...)
	}
	c.net.Send(Msg{From: n.id, To: to, Type: MsgAppendEntries, Term: n.term,
		PrevLogIndex: prev.Index, PrevLogTerm: prev.Term,
		Entries: ents, LeaderCommit: n.commitIndex})
}

func (c *Cluster) handleAppend(n *Node, m Msg) {
	if m.Term < n.term { // 过时的 Leader
		c.net.Send(Msg{From: n.id, To: m.From, Type: MsgAppendEntriesReply, Term: n.term, Success: false})
		return
	}
	n.term, n.state, n.votedFor = m.Term, Follower, m.From
	n.timer, n.timeout = 0, c.newTimeout()

	// ★ 一致性检查：prevLogIndex 处的 term 必须等于 prevLogTerm
	prev, ok := n.at(m.PrevLogIndex)
	if !ok || prev.Term != m.PrevLogTerm {
		c.Rejects++
		reply := Msg{From: n.id, To: m.From, Type: MsgAppendEntriesReply, Term: n.term, Success: false}
		if !ok { // 日志太短
			reply.ConflictTerm = -1
			reply.ConflictIndex = n.last().Index + 1
		} else { // 该位置 term 不同 ⇒ 报告冲突任期的第一条
			reply.ConflictTerm = prev.Term
			i := m.PrevLogIndex
			for i > n.snapIdx+1 && n.entryAt(i-1).Term == prev.Term {
				i--
			}
			reply.ConflictIndex = i
		}
		c.net.Send(reply)
		return
	}

	// 检查通过：截断冲突部分，追加新条目
	for k, e := range m.Entries {
		idx := m.PrevLogIndex + 1 + k
		if cur, ok := n.at(idx); ok {
			if cur.Term == e.Term {
				continue // 已有且一致，跳过
			}
			n.log = n.log[:idx-n.log[0].Index] // ★ 冲突 ⇒ 删掉该位置及之后的全部日志
		}
		n.log = append(n.log, e)
	}
	c.Appends++

	// ★ 上限是「这次 RPC 里最后一条新条目」，而不是「我自己的最后一条」。
	//   我自己更靠后的条目还没被这次 RPC 验证过，可能和 Leader 的不一样——
	//   拿它当上限，就可能把一条错的条目标成已提交。（Figure 2 原文：index of last new entry）
	//   在「每次都把日志发到末尾」时两者等价，一旦给单条消息设大小上限就不再等价。
	if lastNew := m.PrevLogIndex + len(m.Entries); m.LeaderCommit > n.commitIndex {
		n.commitIndex = min(m.LeaderCommit, lastNew)
		c.apply(n)
	}
	n.persist()
	c.net.Send(Msg{From: n.id, To: m.From, Type: MsgAppendEntriesReply, Term: n.term,
		Success: true, MatchIndex: m.PrevLogIndex + len(m.Entries)})
}

func (c *Cluster) handleAppendReply(n *Node, m Msg) {
	if m.Term > n.term {
		c.stepDown(n, m.Term)
		return
	}
	if n.state != Leader || m.Term != n.term {
		return
	}
	if m.Success {
		if m.MatchIndex > n.matchIndex[m.From] {
			n.matchIndex[m.From] = m.MatchIndex
		}
		n.nextIndex[m.From] = n.matchIndex[m.From] + 1
		c.maybeCommit(n)
		return
	}
	// 被拒绝：回退 nextIndex
	if c.SlowBackoff {
		if n.nextIndex[m.From] > 1 {
			n.nextIndex[m.From]--
		}
		return
	}
	// 快速回退：一次跳过整个冲突任期
	if m.ConflictTerm == -1 {
		n.nextIndex[m.From] = m.ConflictIndex
		return
	}
	last := -1
	for i := len(n.log) - 1; i >= 1; i-- {
		if n.log[i].Term == m.ConflictTerm {
			last = n.log[i].Index
			break
		}
	}
	if last >= 0 {
		n.nextIndex[m.From] = last + 1 // Leader 也有这个任期 ⇒ 分歧点在它之后
	} else {
		n.nextIndex[m.From] = m.ConflictIndex // Leader 没有 ⇒ 整段都要覆盖
	}
	if n.nextIndex[m.From] < 1 {
		n.nextIndex[m.From] = 1
	}
}

// maybeCommit 就是 §3.12 的提交规则，两个条件缺一不可。
func (c *Cluster) maybeCommit(n *Node) {
	for N := n.last().Index; N > n.commitIndex; N-- {
		cnt := 0
		for j := range c.nodes {
			if n.matchIndex[j] >= N {
				cnt++
			}
		}
		if cnt < n.majorityOf() { // ★ 同上
			continue
		}
		// ★★★ 这一行就是 Figure 8 的那条限制 ★★★
		if !c.NaiveCommit && n.entryAt(N).Term != n.term {
			continue
		}
		n.commitIndex = N
		c.apply(n)
		return
	}
}

// apply 把已提交的日志喂给状态机，并在这里做两条安全性断言。
func (c *Cluster) apply(n *Node) {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		e := n.entryAt(n.lastApplied)
		if e.Index != n.lastApplied {
			n.lastApplied--
			break
		}
		n.applied[e.Index] = e.Cmd
		applyKV(n, e.Cmd) // 喂给状态机

		// 断言一：状态机安全性 —— 同一 index 上不能有两个节点执行不同命令
		if prev, ok := c.globalCommitted[e.Index]; ok {
			if prev.Cmd != e.Cmd || prev.Term != e.Term {
				c.violate("【状态机安全性被破坏】index=%d 上出现两条不同的已提交日志：先前 (term=%d,%s)，现在 N%d 执行了 (term=%d,%s)",
					e.Index, prev.Term, prev.Cmd, n.id+1, e.Term, e.Cmd)
			}
		} else {
			c.globalCommitted[e.Index] = e
		}
	}
	c.MaybeSnapshot(n)
}

// applyKV 把 "SET k=v" 这样的命令作用到状态机上。
func applyKV(n *Node, cmd string) {
	if !strings.HasPrefix(cmd, "SET ") {
		return
	}
	kv := strings.SplitN(cmd[4:], "=", 2)
	if len(kv) == 2 {
		n.kv[kv[0]] = kv[1]
	}
}

// ─── 主循环 ────────────────────────────────────────────────────────────
func (c *Cluster) handle(m Msg) {
	n := c.nodes[m.To]
	switch m.Type {
	case MsgRequestVote:
		if m.Term > n.term {
			c.stepDown(n, m.Term)
		}
		grant := m.Term == n.term && n.state != Leader &&
			(n.votedFor == NoVote || n.votedFor == m.From) &&
			(c.NoLogCheck || upToDate(m.LastLogTerm, m.LastLogIndex, n.last())) // ★ 日志至少一样新
		if grant {
			n.votedFor = m.From
			n.timer, n.timeout = 0, c.newTimeout()
			n.persist() // ★ 投票必须先落盘再回复 —— 否则重启后会忘记，同任期投两次
		}
		c.net.Send(Msg{From: n.id, To: m.From, Type: MsgRequestVoteReply, Term: n.term, Granted: grant})

	case MsgRequestVoteReply:
		if m.Term > n.term {
			c.stepDown(n, m.Term)
			return
		}
		if n.state == Candidate && m.Granted && m.Term == n.term {
			n.votes[m.From] = true
			if len(n.votes) >= n.majorityOf() { // ★ 用【本节点自己的】配置算过半
				c.becomeLeader(n)
			}
		}

	case MsgAppendEntries:
		c.handleAppend(n, m)
	case MsgAppendEntriesReply:
		n.ackTicks++
		c.handleAppendReply(n, m)
	case MsgInstallSnapshot:
		c.handleInstallSnapshot(n, m)
	case MsgInstallSnapshotReply:
		if n.state == Leader && m.Success {
			n.ackTicks++
			if m.MatchIndex > n.matchIndex[m.From] {
				n.matchIndex[m.From] = m.MatchIndex
			}
			n.nextIndex[m.From] = n.matchIndex[m.From] + 1
		}
	}
}

func (c *Cluster) Submit(cmd string) bool {
	l := c.Leader()
	if l == nil {
		return false
	}
	c.cmdSeq++
	l.log = append(l.log, Entry{Index: l.last().Index + 1, Term: l.term, Cmd: cmd})
	l.matchIndex[l.id] = l.last().Index
	l.persist()
	return true
}

func (c *Cluster) Step(dt int64) {
	for _, n := range c.nodes {
		if c.net.Down(n.id) {
			continue
		}
		if n.state == Leader {
			n.hb += dt
			if n.hb >= HeartbeatMs {
				// 上一轮心跳收到了过半响应 ⇒ 续租约（Lease Read 的依据）
				if n.ackTicks+1 >= n.majorityOf() && c.LeaseMs > 0 {
					n.leaseTill = c.net.Now() + c.LeaseMs
				}
				n.ackTicks = 0
				n.hb = 0
				n.matchIndex[n.id] = n.last().Index
				for j := range c.nodes {
					if j != n.id && n.config[j] {
						c.sendAppend(n, j)
					}
				}
				c.maybeCommit(n)
			}
		} else {
			n.timer += dt
			if n.timer >= n.timeout {
				if n.state == Candidate {
					c.Splits++
				}
				c.startElection(n)
			}
		}
	}
	for _, m := range c.net.Advance(dt) {
		if !c.net.Down(m.To) {
			c.handle(m)
		}
	}
	c.checkSafety()
}

// checkSafety 每个 tick 检查选举安全性与日志匹配性质。
func (c *Cluster) checkSafety() {
	byTerm := map[int]int{}
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && n.state == Leader {
			byTerm[n.term]++
		}
	}
	for t, k := range byTerm {
		if k > 1 {
			c.twoLeadersEver = true
			c.violate("【选举安全性被破坏】term %d 同时有 %d 个 Leader", t, k)
		}
	}
	// 成员变更场景下的裂脑：两个 Leader【各自都能在自己的配置下凑齐过半】，
	// 也就是两个都能独立提交日志。只是自称 Leader 的僵尸不算。
	if c.ActiveLeaders() > 1 {
		c.twoLeadersEver = true
	}
	// 日志匹配性质：任意两个节点，若某 index 上 term 相同，则之前全部相同
	for i := 0; i < len(c.nodes); i++ {
		for j := i + 1; j < len(c.nodes); j++ {
			a, b := c.nodes[i], c.nodes[j]
			lo := max(a.snapIdx, b.snapIdx) + 1
			mn := min(a.last().Index, b.last().Index)
			for k := mn; k >= lo; k-- {
				if a.entryAt(k).Term == b.entryAt(k).Term {
					for x := lo; x <= k; x++ {
						if a.entryAt(x).Term != b.entryAt(x).Term || a.entryAt(x).Cmd != b.entryAt(x).Cmd {
							c.violate("【日志匹配性质被破坏】N%d 与 N%d 在 index=%d 上 term 相同，但 index=%d 不同",
								i+1, j+1, k, x)
						}
					}
					break
				}
			}
		}
	}
}

// ActiveLeaders 数一数有几个 Leader【能在自己的配置下凑齐过半】——
// 也就是有几个 Leader 真的能提交日志。
//
//	这才是「裂脑」的准确定义。一个被隔离到少数派的僵尸 Leader 虽然也自称 Leader，
//	但它一条日志都提交不了，不构成裂脑。
func (c *Cluster) ActiveLeaders() int {
	k := 0
	for _, n := range c.nodes {
		if c.net.Down(n.id) || n.state != Leader {
			continue
		}
		reach := 0
		for j := range c.nodes {
			if n.config[j] && !c.net.Down(j) && c.net.SameGroup(n.id, j) {
				reach++
			}
		}
		if reach >= n.majorityOf() {
			k++
		}
	}
	return k
}

func (c *Cluster) Leader() *Node {
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && n.state == Leader {
			return n
		}
	}
	return nil
}

func (c *Cluster) LogsIdentical() (bool, string) {
	base := c.nodes[0]
	for _, n := range c.nodes[1:] {
		mn := min(base.commitIndex, n.commitIndex)
		for i := max(base.snapIdx, n.snapIdx) + 1; i <= mn; i++ {
			if base.entryAt(i).Cmd != n.entryAt(i).Cmd || base.entryAt(i).Term != n.entryAt(i).Term {
				return false, fmt.Sprintf("N1 与 N%d 在 index=%d 不同", n.id+1, i)
			}
		}
	}
	return true, ""
}

func (c *Cluster) MaxCommit() int {
	m := 0
	for _, n := range c.nodes {
		if n.commitIndex > m {
			m = n.commitIndex
		}
	}
	return m
}

func (c *Cluster) LogStr(i int) string {
	s := ""
	for _, e := range c.nodes[i].log[1:] {
		s += fmt.Sprint(e.Term) + " "
	}
	if s == "" {
		s = "(空)"
	}
	return s
}

// ════════════════════════════════════════════════════════════════════════
// Part 3-C ①：持久化与崩溃重启
// ════════════════════════════════════════════════════════════════════════

// persist 在响应任何 RPC 之前调用 —— 论文的要求就是这么严格。
func (n *Node) persist() {
	n.disk.CurrentTerm = n.term
	n.disk.VotedFor = n.votedFor
	n.disk.Log = append([]Entry(nil), n.log...)
}

// Restart 模拟一次崩溃重启：
//
//	易失状态全部清零，持久化状态从「磁盘」恢复 ——
//	但 c.LoseXxx 打开时，对应的那一样恢复不回来，用来演示后果。
func (c *Cluster) Restart(i int) {
	n := c.nodes[i]
	d := n.disk

	n.term = d.CurrentTerm
	n.votedFor = d.VotedFor
	n.log = append([]Entry(nil), d.Log...)

	if c.LoseTerm {
		n.term = 0 // ★ 丢了 currentTerm：节点失去时间感，无法识别过时消息
	}
	if c.LoseVote {
		n.votedFor = NoVote // ★ 丢了 votedFor：忘了自己投过票 ⇒ 可以在同一任期投两次
	}
	if c.LoseLog {
		n.log = []Entry{{Index: 0, Term: 0}} // ★ 丢了 log：已提交的条目在这个副本上消失
	}

	// 易失状态一律重置
	n.state = Follower
	n.votes = map[int]bool{}
	n.commitIndex, n.lastApplied = n.snapIdx, n.snapIdx
	n.timer, n.timeout, n.hb = 0, c.newTimeout(), 0
	n.leaseTill = 0
	for j := range c.nodes {
		n.nextIndex[j], n.matchIndex[j] = 1, 0
	}
	n.persist()
	c.net.Revive(i)
}

// ════════════════════════════════════════════════════════════════════════
// Part 3-C ②：快照
// ════════════════════════════════════════════════════════════════════════

// MaybeSnapshot 把 lastApplied 之前的日志压缩掉。
// 快照里必须带上 (lastIncludedIndex, lastIncludedTerm) —— 它们是被截断部分的「代表」，
// 让 AppendEntries 的一致性检查在截断之后依然能工作。
func (c *Cluster) MaybeSnapshot(n *Node) {
	if c.SnapThreshold <= 0 || n.lastApplied-n.snapIdx < c.SnapThreshold {
		return
	}
	cut := n.lastApplied
	e := n.entryAt(cut)
	n.snapIdx, n.snapTerm = cut, e.Term
	n.snapshot = map[string]string{}
	for k, v := range n.kv {
		n.snapshot[k] = v
	}
	// 只保留 cut 之后的日志；0 号位换成快照的「代表」条目
	var kept []Entry
	kept = append(kept, Entry{Index: n.snapIdx, Term: n.snapTerm})
	for _, x := range n.log {
		if x.Index > cut {
			kept = append(kept, x)
		}
	}
	n.log = kept
	n.persist()
	c.Snapshots++
}

// entryAt 按逻辑 index 取条目；index == snapIdx 时返回快照的代表条目。
func (n *Node) entryAt(idx int) Entry {
	base := n.log[0].Index
	if idx < base {
		return Entry{Index: -1, Term: -1} // 已被压缩掉，取不到
	}
	off := idx - base
	if off >= len(n.log) {
		return Entry{Index: -1, Term: -1}
	}
	return n.log[off]
}

func (n *Node) hasIndex(idx int) bool { return n.entryAt(idx).Index == idx }

func (c *Cluster) sendSnapshot(n *Node, to int) {
	snap := map[string]string{}
	for k, v := range n.snapshot {
		snap[k] = v
	}
	c.net.Send(Msg{From: n.id, To: to, Type: MsgInstallSnapshot, Term: n.term,
		LastIncludedIndex: n.snapIdx, LastIncludedTerm: n.snapTerm, Snapshot: snap})
	c.InstallSnaps++
}

func (c *Cluster) handleInstallSnapshot(n *Node, m Msg) {
	if m.Term < n.term {
		return
	}
	n.term, n.state, n.votedFor = m.Term, Follower, m.From
	n.timer, n.timeout = 0, c.newTimeout()
	if m.LastIncludedIndex <= n.snapIdx {
		return // 已经有更新的快照了
	}
	n.snapIdx, n.snapTerm = m.LastIncludedIndex, m.LastIncludedTerm
	n.snapshot = m.Snapshot
	n.kv = map[string]string{}
	for k, v := range m.Snapshot {
		n.kv[k] = v
	}
	n.log = []Entry{{Index: n.snapIdx, Term: n.snapTerm}}
	n.commitIndex, n.lastApplied = n.snapIdx, n.snapIdx
	n.persist()
	c.net.Send(Msg{From: n.id, To: m.From, Type: MsgInstallSnapshotReply,
		Term: n.term, Success: true, MatchIndex: n.snapIdx})
}

// ════════════════════════════════════════════════════════════════════════
// Part 3-C ③：成员变更
//
//	关键：每个节点持有【自己的】config，切换时刻可以不同。
//	过半的判断用节点自己的 config —— 这正是裂脑的来源。
// ════════════════════════════════════════════════════════════════════════

func (n *Node) majorityOf() int { return len(n.config)/2 + 1 }

// SetConfig 让指定节点切换到新配置（模拟"不同节点在不同时刻切换"）。
func (c *Cluster) SetConfig(i int, members []int) {
	cfg := map[int]bool{}
	for _, m := range members {
		cfg[m] = true
	}
	c.nodes[i].config = cfg
}

// ════════════════════════════════════════════════════════════════════════
// Part 3-C ④：三种读
// ════════════════════════════════════════════════════════════════════════

type ReadMode int

const (
	LocalRead ReadMode = iota
	ReadIndexRead
	LeaseRead
)

func (r ReadMode) String() string { return [...]string{"本地读", "ReadIndex", "Lease Read"}[r] }

// Read 在节点 i 上按指定模式读一个 key。
// 返回 (值, 是否成功)。失败表示节点正确地拒绝了服务 —— 这是好事，不是 bug。
func (c *Cluster) Read(i int, key string, mode ReadMode) (string, bool) {
	n := c.nodes[i]
	if c.net.Down(n.id) || n.state != Leader {
		return "", false
	}
	switch mode {
	case LocalRead:
		// 直接读状态机。最快，但无法证明自己此刻还是 Leader。
		return n.kv[key], true
	case ReadIndexRead:
		// 读之前先确认过半节点仍然认自己是主。
		if n.ackTicks+1 < n.majorityOf() { // +1 是它自己
			return "", false
		}
		return n.kv[key], true
	case LeaseRead:
		// 租约期内直接本地读；租约过期就拒绝。
		if c.net.Now() > n.leaseTill {
			return "", false
		}
		return n.kv[key], true
	}
	return "", false
}
