package main

import "math/rand"

// ════════════════════════════════════════════════════════════════════════
// Quorum KV：N 个副本的键值存储，W/R 可配，带复制延迟
//
//	核心命题：W + R > N  ⟺  任意读集合与任意写集合必有交集  ⟹  读一定能看到最新写
//	Lab 2-1 会用实测数据验证它。
//
// ════════════════════════════════════════════════════════════════════════

type Versioned struct {
	Ver int
	Val string
}

type pending struct {
	node      int
	v         Versioned
	deliverAt int // 第几次操作之后送达（模拟复制延迟）
}

type QuorumKV struct {
	N, W, R int
	Lag     int // 未被写 quorum 覆盖的副本，要延迟这么多次操作才收到
	rnd     *rand.Rand

	replicas []Versioned
	queue    []pending
	clock    int // 全局版本号（上帝视角）
	opCount  int

	StaleReads, TotalReads, FailedWrites int
}

func NewQuorumKV(n, w, r, lag int, seed int64) *QuorumKV {
	return &QuorumKV{N: n, W: w, R: r, Lag: lag,
		rnd: rand.New(rand.NewSource(seed)), replicas: make([]Versioned, n)}
}

// sample 随机挑 k 个不同的副本下标。
func (q *QuorumKV) sample(k int) []int {
	p := q.rnd.Perm(q.N)
	return p[:k]
}

// tick 推进"时间"，把到期的复制消息投递出去。
func (q *QuorumKV) tick() {
	q.opCount++
	keep := q.queue[:0]
	for _, p := range q.queue {
		if p.deliverAt <= q.opCount {
			if p.v.Ver > q.replicas[p.node].Ver {
				q.replicas[p.node] = p.v
			}
		} else {
			keep = append(keep, p)
		}
	}
	q.queue = keep
}

// Write 向随机挑选的 W 个副本同步写入；其余副本经过 Lag 次操作后才异步收到。
func (q *QuorumKV) Write(val string) bool {
	q.tick()
	q.clock++
	v := Versioned{Ver: q.clock, Val: val}

	target := q.sample(q.W)
	inW := map[int]bool{}
	for _, i := range target {
		inW[i] = true
		q.replicas[i] = v
	}
	for i := 0; i < q.N; i++ {
		if !inW[i] {
			q.queue = append(q.queue, pending{node: i, v: v, deliverAt: q.opCount + q.Lag})
		}
	}
	return true
}

// Read 从随机挑选的 R 个副本收集，取版本号最大的那个（这正是 Dynamo 风格的做法）。
func (q *QuorumKV) Read() Versioned {
	q.tick()
	best := Versioned{}
	for _, i := range q.sample(q.R) {
		if q.replicas[i].Ver > best.Ver {
			best = q.replicas[i]
		}
	}
	q.TotalReads++
	if best.Ver < q.clock { // 上帝视角：已提交的最新版本是 q.clock
		q.StaleReads++
	}
	return best
}

func (q *QuorumKV) Guaranteed() bool { return q.W+q.R > q.N }
