package main

import (
	"fmt"
	"math/rand"
)

// ════════════════════════════════════════════════════════════════════════
// Raft 日志复制（在 Lab 3-A 的选举之上）
//
//	新增三件事：
//	  ① AppendEntries 的 prevLog 一致性检查 ⇒ 日志匹配性质
//	  ② 提交规则：过半 matchIndex + 【当前任期】 ⇒ Figure 8 的那条限制
//	  ③ 三条安全性断言，每个 tick 都检查
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
}

func (n *Node) last() Entry { return n.log[len(n.log)-1] }

func (n *Node) at(i int) (Entry, bool) {
	if i < 0 || i >= len(n.log) {
		return Entry{}, false
	}
	return n.log[i], true
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
		nd := &Node{id: i, state: Follower, votedFor: NoVote, votes: map[int]bool{},
			log: []Entry{{Index: 0, Term: 0}}, applied: map[int]string{},
			nextIndex: make([]int, n), matchIndex: make([]int, n)}
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
}

func (c *Cluster) startElection(n *Node) {
	n.term++
	n.state = Candidate
	n.votedFor = n.id
	n.votes = map[int]bool{n.id: true}
	n.timer, n.timeout = 0, c.newTimeout()
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
	prev, _ := n.at(ni - 1)
	var ents []Entry
	if ni <= n.last().Index {
		ents = append(ents, n.log[ni:]...)
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
			for i > 1 && n.log[i-1].Term == prev.Term {
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
			n.log = n.log[:idx] // ★ 冲突 ⇒ 删掉该位置及之后的全部日志
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
			last = i
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
		if cnt < c.majority() {
			continue
		}
		// ★★★ 这一行就是 Figure 8 的那条限制 ★★★
		if !c.NaiveCommit && n.log[N].Term != n.term {
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
		e := n.log[n.lastApplied]
		n.applied[e.Index] = e.Cmd

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

	case MsgAppendEntries:
		c.handleAppend(n, m)
	case MsgAppendEntriesReply:
		c.handleAppendReply(n, m)
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
				n.hb = 0
				n.matchIndex[n.id] = n.last().Index
				for j := range c.nodes {
					if j != n.id {
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
			c.violate("【选举安全性被破坏】term %d 同时有 %d 个 Leader", t, k)
		}
	}
	// 日志匹配性质：任意两个节点，若某 index 上 term 相同，则之前全部相同
	for i := 0; i < len(c.nodes); i++ {
		for j := i + 1; j < len(c.nodes); j++ {
			a, b := c.nodes[i], c.nodes[j]
			mn := min(a.last().Index, b.last().Index)
			for k := mn; k >= 1; k-- {
				if a.log[k].Term == b.log[k].Term {
					for x := 1; x <= k; x++ {
						if a.log[x].Term != b.log[x].Term || a.log[x].Cmd != b.log[x].Cmd {
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
		for i := 1; i <= mn; i++ {
			if base.log[i].Cmd != n.log[i].Cmd || base.log[i].Term != n.log[i].Term {
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
