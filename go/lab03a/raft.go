package main

import (
	"fmt"
	"math/rand"
)

// ════════════════════════════════════════════════════════════════════════
// Raft 选举状态机
//
//	这一节只做三件事：任期递增、投票、心跳压制。
//	日志复制与安全性属性留给 Lab 3-B。
//	核心逻辑不到 150 行 —— Raft 之所以出名，正是因为它能这么短。
//
// ════════════════════════════════════════════════════════════════════════

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

func (s State) String() string { return [...]string{"Follower", "Candidate", "Leader"}[s] }

const (
	HeartbeatMs = 100 // Leader 发心跳的间隔
	NoVote      = -1
)

type Node struct {
	id       int
	state    State
	term     int
	votedFor int          // 本任期把票投给了谁；NoVote 表示还没投
	votes    map[int]bool // Candidate 收到的选票
	timer    int64        // 选举计时器已走过的毫秒
	timeout  int64        // 本轮的选举超时阈值
	hb       int64        // Leader 距上次心跳的毫秒
}

type Cluster struct {
	nodes  []*Node
	net    *Network
	rnd    *rand.Rand
	base   int64 // 选举超时下限
	window int64 // 随机区间宽度；0 表示不随机化
	unsafe bool  // 演示用：去掉「每个任期只投一票」这条规则

	Elections, Splits int
	MaxLeadersPerTerm int
	Violations        []string
	Log               []string
	Verbose           bool
}

func New(n int, net *Network, base, window, seed int64, unsafe bool) *Cluster {
	c := &Cluster{net: net, rnd: rand.New(rand.NewSource(seed + 999)),
		base: base, window: window, unsafe: unsafe, MaxLeadersPerTerm: 0}
	for i := 0; i < n; i++ {
		nd := &Node{id: i, state: Follower, votedFor: NoVote, votes: map[int]bool{}}
		nd.timeout = c.newTimeout()
		c.nodes = append(c.nodes, nd)
	}
	return c
}

func (c *Cluster) newTimeout() int64 {
	if c.window <= 0 {
		return c.base // 固定超时 ⇒ 分裂投票
	}
	return c.base + c.rnd.Int63n(c.window)
}

// Majority 是过半票数 —— Raft 全部安全性的地基。
func (c *Cluster) majority() int { return len(c.nodes)/2 + 1 }

func (c *Cluster) logf(f string, a ...any) {
	s := fmt.Sprintf("[%6dms] ", c.net.Now()) + fmt.Sprintf(f, a...)
	c.Log = append(c.Log, s)
	if c.Verbose {
		fmt.Println("    " + s)
	}
}

// stepDown 是 Raft 里最重要的一条规则：
// 任何节点，任何时候，看到更大的任期号就立刻退回 Follower。
// 旧 Leader 因此会自动退位，不需要任何人通知它。
func (c *Cluster) stepDown(n *Node, term int) {
	if n.state == Leader {
		c.logf("N%d 看到 term=%d，从 Leader 退位", n.id+1, term)
	}
	n.term = term
	n.state = Follower
	n.votedFor = NoVote
	n.votes = map[int]bool{}
	n.timer = 0
	n.timeout = c.newTimeout()
}

func (c *Cluster) startElection(n *Node) {
	n.term++
	n.state = Candidate
	n.votedFor = n.id // 先投自己一票
	n.votes = map[int]bool{n.id: true}
	n.timer = 0
	n.timeout = c.newTimeout()
	c.Elections++
	c.logf("N%d 选举超时 → Candidate，term=%d", n.id+1, n.term)
	for j := range c.nodes {
		if j != n.id {
			c.net.Send(Msg{From: n.id, To: j, Type: MsgRequestVote, Term: n.term})
		}
	}
}

func (c *Cluster) becomeLeader(n *Node) {
	n.state = Leader
	n.hb = HeartbeatMs // 立刻发一轮心跳宣示主权
	c.logf("★ N%d 当选 Leader，term=%d，得票 %d/%d", n.id+1, n.term, len(n.votes), len(c.nodes))
}

func (c *Cluster) handle(m Msg) {
	n := c.nodes[m.To]
	switch m.Type {

	case MsgRequestVote:
		if m.Term > n.term {
			c.stepDown(n, m.Term)
		}
		// 投票的条件：任期不过时、本任期还没投过票。
		//
		// ★ 注意这里缺了真实 Raft 的第三个条件：「candidate 的日志至少和我一样新」。
		//   本 Lab 的模型里没有日志，所以无法实现它。这个简化有可观测的后果：
		//   分区恢复后，任期很高但日志陈旧的少数派节点会有相当高的概率当选 ——
		//   真实 Raft 里这不可能发生。Lab 3B-5 用 A/B 对照把这件事测了出来。
		grant := m.Term == n.term && n.state != Leader &&
			(n.votedFor == NoVote || n.votedFor == m.From)
		if c.unsafe {
			// ★ 故意去掉「每个任期只投一票」——用来演示安全性怎么被打破
			grant = m.Term >= n.term && n.state != Leader
			if m.Term > n.term {
				n.term = m.Term
			}
		}
		if grant {
			n.votedFor = m.From
			n.timer = 0
			n.timeout = c.newTimeout()
		}
		c.net.Send(Msg{From: n.id, To: m.From, Type: MsgRequestVoteReply, Term: n.term, Granted: grant})

	case MsgRequestVoteReply:
		if m.Term > n.term {
			c.stepDown(n, m.Term)
			return
		}
		if n.state == Candidate && m.Granted && m.Term == n.term {
			n.votes[m.From] = true
			if len(n.votes) >= c.majority() {
				c.becomeLeader(n)
			}
		}

	case MsgAppendEntries: // 心跳
		if m.Term >= n.term {
			if n.state == Leader && m.From != n.id {
				c.logf("N%d 收到 term=%d 的心跳，从 Leader 退位", n.id+1, m.Term)
			}
			n.term = m.Term
			n.state = Follower
			n.votedFor = m.From
			n.votes = map[int]bool{}
			n.timer = 0
			n.timeout = c.newTimeout()
		}
		// m.Term < n.term：发送者是过时的 Leader，直接忽略（它会从别处知道自己该退位）
	}
}

// Step 把整个集群向前推进 dt 毫秒。
func (c *Cluster) Step(dt int64) {
	for _, n := range c.nodes {
		if c.net.Down(n.id) {
			continue
		}
		if n.state == Leader {
			n.hb += dt
			if n.hb >= HeartbeatMs {
				n.hb = 0
				for j := range c.nodes {
					if j != n.id {
						c.net.Send(Msg{From: n.id, To: j, Type: MsgAppendEntries, Term: n.term})
					}
				}
			}
		} else {
			n.timer += dt
			if n.timer >= n.timeout {
				if n.state == Candidate {
					c.Splits++
					c.logf("N%d 超时仍未过半 → 分裂投票", n.id+1)
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

// checkSafety 断言 Raft 的第一条安全性属性：
//
//	Election Safety —— 任何一个任期内，至多只有一个 Leader。
//
// 它的证明只有两步：当选需要过半票 + 每个任期每人只投一票 ⇒ 两个过半集合必相交 ⇒ 矛盾。
func (c *Cluster) checkSafety() {
	byTerm := map[int][]int{}
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && n.state == Leader {
			byTerm[n.term] = append(byTerm[n.term], n.id)
		}
	}
	for term, ids := range byTerm {
		if len(ids) > c.MaxLeadersPerTerm {
			c.MaxLeadersPerTerm = len(ids)
		}
		if len(ids) > 1 {
			v := fmt.Sprintf("term %d 同时存在 %d 个 Leader：%v", term, len(ids), plusOne(ids))
			if len(c.Violations) < 5 {
				c.Violations = append(c.Violations, v)
			}
		}
	}
}

func plusOne(ids []int) []int {
	out := make([]int, len(ids))
	for i, x := range ids {
		out[i] = x + 1
	}
	return out
}

// ─── 查询辅助 ────────────────────────────────────────────────────────────
func (c *Cluster) Leader() *Node {
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && n.state == Leader {
			return n
		}
	}
	return nil
}

func (c *Cluster) LeadersIn(group []int, want int) []*Node {
	var out []*Node
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && n.state == Leader && group[n.id] == want {
			out = append(out, n)
		}
	}
	return out
}

func (c *Cluster) MaxTerm() int {
	m := 0
	for _, n := range c.nodes {
		if n.term > m {
			m = n.term
		}
	}
	return m
}

func (c *Cluster) Summary() string {
	s := ""
	for _, n := range c.nodes {
		tag := n.state.String()[:1]
		if c.net.Down(n.id) {
			tag = "☠"
		}
		s += fmt.Sprintf("N%d:%s(t%d) ", n.id+1, tag, n.term)
	}
	return s
}
