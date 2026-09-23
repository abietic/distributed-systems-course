package main

import (
	"fmt"
	"sort"
)

// ════════════════════════════════════════════════════════════════════════
// 事件轨迹：和课件「实验 1 · 时空图编辑器」里的「载入经典示例」完全一致。
// 你在浏览器里看到的每一个数字，这里都能跑出来。
// ════════════════════════════════════════════════════════════════════════

const NP = 3 // 三个进程

var PN = [NP]string{"P1", "P2", "P3"}

type Kind int

const (
	KLocal Kind = iota
	KSend
	KRecv
)

type Ev struct {
	ID   int
	P    int // 所在进程
	T    int // 时空图上的列（只用于排版和拓扑排序，不是物理时钟！）
	Kind Kind
	Peer int // 配对事件的 ID（发送 ↔ 接收）

	L uint64 // Lamport 时间戳
	V Vector // 向量时间戳
}

func (e *Ev) Name() string { return fmt.Sprintf("%s@t%d", PN[e.P], e.T) }

func (e *Ev) KindName() string {
	return [...]string{"本地事件", "发送消息", "接收消息"}[e.Kind]
}

// BuildTrace 构造经典示例场景并计算两种时钟。
//
//	P1: t1 本地 ─ t3 发送 ──────────────┐        t8 接收 ←┐   t11 本地
//	P2: t0 本地 ───────────── t5 接收 ←┘ t7 发送 ─┐       │
//	P3: t2 本地 ───────── t6 发送 ──────────────────┘─→ t9 接收   t12 本地
func BuildTrace() []*Ev {
	var evs []*Ev
	id := 0
	local := func(p, t int) *Ev {
		id++
		e := &Ev{ID: id, P: p, T: t, Kind: KLocal}
		evs = append(evs, e)
		return e
	}
	msg := func(sp, st, rp, rt int) {
		s, r := local(sp, st), local(rp, rt)
		s.Kind, s.Peer = KSend, r.ID
		r.Kind, r.Peer = KRecv, s.ID
	}

	local(0, 1)     // P1 上的第一个事件
	local(1, 0)     // P2 上的第一个事件（与上面并发）
	local(2, 2)     // P3 上的第一个事件
	msg(0, 3, 1, 5) // P1 → P2
	msg(2, 6, 0, 8) // P3 → P1
	msg(1, 7, 2, 9) // P2 → P3
	local(0, 11)    // P1 最后一个事件
	local(2, 12)    // P3 最后一个事件

	computeClocks(evs)
	return evs
}

// computeClocks 按 (t, p) 排序后依次处理。
// 因为消息一定满足「接收列 > 发送列」，这个顺序天然是一个合法的拓扑序。
func computeClocks(evs []*Ev) {
	byID := map[int]*Ev{}
	for _, e := range evs {
		byID[e.ID] = e
	}
	order := append([]*Ev(nil), evs...)
	sort.Slice(order, func(i, j int) bool {
		if order[i].T != order[j].T {
			return order[i].T < order[j].T
		}
		return order[i].P < order[j].P
	})

	lam := make([]Lamport, NP)
	vec := make([]Vector, NP)
	for i := range vec {
		vec[i] = NewVector(NP)
	}

	for _, e := range order {
		if e.Kind == KRecv {
			s := byID[e.Peer]
			e.L = lam[e.P].Recv(s.L)
			vec[e.P] = vec[e.P].Merge(s.V, e.P)
		} else {
			e.L = lam[e.P].Local()
			vec[e.P] = vec[e.P].Local(e.P)
		}
		e.V = vec[e.P].Clone()
	}
}

// causalPath 用 happens-before 的三条规则做一次可达性搜索，
// 用来独立验证向量时钟的判定是对的（而不是自己验证自己）。
func causalPath(a, b *Ev, evs []*Ev) bool {
	byID := map[int]*Ev{}
	for _, e := range evs {
		byID[e.ID] = e
	}
	seen := map[int]bool{}
	var dfs func(x *Ev) bool
	dfs = func(x *Ev) bool {
		if x.ID == b.ID {
			return true
		}
		if seen[x.ID] {
			return false
		}
		seen[x.ID] = true
		for _, n := range evs { // 规则 ①：同进程内的后继事件
			if n.P == x.P && n.T > x.T && dfs(n) {
				return true
			}
		}
		if x.Kind == KSend { // 规则 ②：发送 → 接收
			if dfs(byID[x.Peer]) {
				return true
			}
		}
		return false
	}
	return dfs(a)
}
