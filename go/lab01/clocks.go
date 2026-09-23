package main

import (
	"fmt"
	"strings"
)

// ════════════════════════════════════════════════════════════════════════
// Lamport 逻辑时钟
//
//	保证：a → b  ⟹  L(a) < L(b)
//	不保证：L(a) < L(b) ⇒ a → b     ← 这就是它的根本局限
//
// ════════════════════════════════════════════════════════════════════════
type Lamport struct{ t uint64 }

// Local 处理一个本地事件（发送消息也算本地事件）。
func (l *Lamport) Local() uint64 { l.t++; return l.t }

// Recv 处理一次消息接收，remote 是消息里携带的发送方时间戳。
func (l *Lamport) Recv(remote uint64) uint64 {
	if remote > l.t {
		l.t = remote
	}
	l.t++
	return l.t
}

// ════════════════════════════════════════════════════════════════════════
// 向量时钟
//
//	保证：a → b  ⟺  V(a) < V(b)      ← 充要条件，所以能精确检测并发
//	代价：O(N) 空间，N = 节点数
//
// ════════════════════════════════════════════════════════════════════════
type Vector []uint64

func NewVector(n int) Vector { return make(Vector, n) }

func (v Vector) Clone() Vector { c := make(Vector, len(v)); copy(c, v); return c }

// Local：进程 i 发生一个本地事件。
func (v Vector) Local(i int) Vector { c := v.Clone(); c[i]++; return c }

// Merge：进程 i 收到携带向量 o 的消息 —— 先逐位取最大，再自增自己那一位。
func (v Vector) Merge(o Vector, i int) Vector {
	c := v.Clone()
	for k := range c {
		if o[k] > c[k] {
			c[k] = o[k]
		}
	}
	c[i]++
	return c
}

func (v Vector) String() string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprint(x)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

type Ord int

const (
	Equal Ord = iota
	Before
	After
	Concurrent
)

func (o Ord) String() string {
	return [...]string{"相同", "a → b", "b → a", "a ∥ b（并发）"}[o]
}

func le(a, b Vector) bool {
	for k := range a {
		if a[k] > b[k] {
			return false
		}
	}
	return true
}

// Compare 是向量时钟的全部精髓，只有五行。
func Compare(a, b Vector) Ord {
	ab, ba := le(a, b), le(b, a)
	switch {
	case ab && ba:
		return Equal
	case ab:
		return Before
	case ba:
		return After
	default:
		return Concurrent // 各自都知道对方不知道的事 ⇒ 谁都没影响谁 ⇒ 冲突
	}
}

// ════════════════════════════════════════════════════════════════════════
// 混合逻辑时钟 HLC（Kulkarni et al., 2014）
//
//	时间戳是二元组 (l, c)，按字典序比较。
//	l 尽量贴着物理时钟走，c 只在物理时钟"不够用"时递增。
//	性质：单调 + 满足 Lamport 时钟条件 + |l − 物理时间| 有界
//
// ════════════════════════════════════════════════════════════════════════
type HLC struct{ L, C int64 }

// Local：产生一个本地事件 / 发送消息。pt 是当前物理时钟读数。
func (h *HLC) Local(pt int64) (int64, int64) {
	prev := h.L
	if pt > h.L {
		h.L = pt
	}
	if h.L == prev {
		h.C++ // 物理时钟没有前进（含回拨），靠逻辑位撑住单调性
	} else {
		h.C = 0 // 物理时钟前进了，逻辑位归零
	}
	return h.L, h.C
}

// Recv：收到携带 (lm, cm) 的消息。
func (h *HLC) Recv(pt, lm, cm int64) (int64, int64) {
	prev := h.L
	h.L = max3(prev, lm, pt)
	switch {
	case h.L == prev && h.L == lm:
		h.C = maxI(h.C, cm) + 1
	case h.L == prev:
		h.C++
	case h.L == lm:
		h.C = cm + 1
	default:
		h.C = 0
	}
	return h.L, h.C
}

func (h HLC) String() string { return fmt.Sprintf("(%d, %d)", h.L, h.C) }

func maxI(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func max3(a, b, c int64) int64 { return maxI(maxI(a, b), c) }
