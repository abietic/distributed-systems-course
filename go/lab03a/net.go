package main

import (
	"math/rand"
	"sort"
)

// ════════════════════════════════════════════════════════════════════════
// 确定性网络模拟器
//
//	共识算法极难调试，因为 bug 往往依赖特定的消息交错顺序。
//	所以第一件事不是写 Raft，而是写一个「同一个种子必定复现同一次执行」的网络。
//	这里没有 goroutine、没有真实时间 —— 只有一个虚拟时钟和一个按时间排序的消息队列。
//
// ════════════════════════════════════════════════════════════════════════

type MsgType int

const (
	MsgRequestVote MsgType = iota
	MsgRequestVoteReply
	MsgAppendEntries // 本节只用它做心跳，3-B 才会带上日志
)

func (t MsgType) String() string {
	return [...]string{"RequestVote", "VoteReply", "Heartbeat"}[t]
}

type Msg struct {
	From, To  int
	Type      MsgType
	Term      int
	Granted   bool
	DeliverAt int64 // 虚拟时钟上的送达时刻（毫秒）
	seq       int   // 同一时刻的稳定排序键，保证完全确定
}

type Network struct {
	n             int
	rnd           *rand.Rand
	delay, jitter int64
	loss          float64

	group []int  // 分区分组；全 0 表示网络连通
	down  []bool // 节点是否宕机

	queue []Msg
	now   int64
	seq   int

	Sent, Dropped, Delivered int
}

func NewNetwork(n int, delay, jitter int64, loss float64, seed int64) *Network {
	return &Network{n: n, rnd: rand.New(rand.NewSource(seed)),
		delay: delay, jitter: jitter, loss: loss,
		group: make([]int, n), down: make([]bool, n)}
}

func (nw *Network) Now() int64      { return nw.now }
func (nw *Network) Down(i int) bool { return nw.down[i] }
func (nw *Network) Kill(i int)      { nw.down[i] = true }
func (nw *Network) Revive(i int)    { nw.down[i] = false }

// Partition 把节点分成若干组，只有同组之间能通信。groups[i] 是节点 i 的组号。
func (nw *Network) Partition(groups []int) { copy(nw.group, groups) }
func (nw *Network) Heal() {
	for i := range nw.group {
		nw.group[i] = 0
	}
}

// Send 把消息放进队列。丢包、分区、宕机在这里统一裁决 ——
// 注意发送方永远不知道这三件事中的哪一件发生了（Part 0 的第三态）。
func (nw *Network) Send(m Msg) {
	nw.Sent++
	if nw.down[m.From] || nw.down[m.To] || nw.group[m.From] != nw.group[m.To] {
		nw.Dropped++
		return
	}
	if nw.rnd.Float64() < nw.loss {
		nw.Dropped++
		return
	}
	jit := int64(0)
	if nw.jitter > 0 {
		jit = nw.rnd.Int63n(2*nw.jitter) - nw.jitter
	}
	d := nw.delay + jit
	if d < 1 {
		d = 1
	}
	m.DeliverAt = nw.now + d
	m.seq = nw.seq
	nw.seq++
	nw.queue = append(nw.queue, m)
}

// Advance 把虚拟时钟推进 dt 毫秒，返回这段时间内应当送达的全部消息（按时间排序）。
func (nw *Network) Advance(dt int64) []Msg {
	nw.now += dt
	sort.Slice(nw.queue, func(i, j int) bool {
		if nw.queue[i].DeliverAt != nw.queue[j].DeliverAt {
			return nw.queue[i].DeliverAt < nw.queue[j].DeliverAt
		}
		return nw.queue[i].seq < nw.queue[j].seq
	})
	var due []Msg
	keep := nw.queue[:0]
	for _, m := range nw.queue {
		if m.DeliverAt <= nw.now {
			if !nw.down[m.To] && nw.group[m.From] == nw.group[m.To] {
				due = append(due, m) // 飞行途中发生的宕机/分区同样会让消息落空
				nw.Delivered++
			} else {
				nw.Dropped++
			}
		} else {
			keep = append(keep, m)
		}
	}
	nw.queue = keep
	return due
}
