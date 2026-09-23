package main

// ════════════════════════════════════════════════════════════════════════
// CRDT：合并操作满足交换律、结合律、幂等律 ⇒ 冲突在数学上不可能存在
// ════════════════════════════════════════════════════════════════════════

// GCounter 只增计数器。结构和 Part 1 的向量时钟完全一样。
type GCounter []int

func NewGCounter(n int) GCounter { return make(GCounter, n) }

func (g GCounter) Inc(node int) { g[node]++ }

// Merge 逐位取最大 —— 交换、结合、幂等三律全部满足。
func (g GCounter) Merge(o GCounter) {
	for i := range g {
		if o[i] > g[i] {
			g[i] = o[i]
		}
	}
}

func (g GCounter) Value() int {
	s := 0
	for _, x := range g {
		s += x
	}
	return s
}

func (g GCounter) Clone() GCounter { c := make(GCounter, len(g)); copy(c, g); return c }

// PNCounter 可增可减：两个 GCounter，一个记增一个记减。
type PNCounter struct{ P, N GCounter }

func NewPNCounter(n int) *PNCounter { return &PNCounter{NewGCounter(n), NewGCounter(n)} }
func (c *PNCounter) Inc(node int)   { c.P.Inc(node) }
func (c *PNCounter) Dec(node int)   { c.N.Inc(node) }
func (c *PNCounter) Merge(o *PNCounter) {
	c.P.Merge(o.P)
	c.N.Merge(o.N)
}
func (c *PNCounter) Value() int { return c.P.Value() - c.N.Value() }

// ORSet 观察-移除集合：每个元素带唯一 tag，删除只能删掉"已观察到的" tag，
// 于是并发的「加」和「删」中，加优先（add-wins）。
type ORSet struct {
	adds map[string]map[int64]bool // 元素 -> tag 集合
	rems map[string]map[int64]bool // 元素 -> 已删除的 tag 集合
}

func NewORSet() *ORSet {
	return &ORSet{adds: map[string]map[int64]bool{}, rems: map[string]map[int64]bool{}}
}

func (s *ORSet) Add(e string, tag int64) {
	if s.adds[e] == nil {
		s.adds[e] = map[int64]bool{}
	}
	s.adds[e][tag] = true
}

func (s *ORSet) Remove(e string) {
	if s.rems[e] == nil {
		s.rems[e] = map[int64]bool{}
	}
	for t := range s.adds[e] { // 只能删掉自己已经观察到的 tag
		s.rems[e][t] = true
	}
}

func (s *ORSet) Has(e string) bool {
	for t := range s.adds[e] {
		if !s.rems[e][t] {
			return true
		}
	}
	return false
}

func (s *ORSet) Merge(o *ORSet) {
	mergeInto := func(dst, src map[string]map[int64]bool) {
		for e, tags := range src {
			if dst[e] == nil {
				dst[e] = map[int64]bool{}
			}
			for t := range tags {
				dst[e][t] = true
			}
		}
	}
	mergeInto(s.adds, o.adds)
	mergeInto(s.rems, o.rems)
}

func (s *ORSet) Size() int {
	n := 0
	for e := range s.adds {
		if s.Has(e) {
			n++
		}
	}
	return n
}

// LWWRegister 也是合法的 CRDT（必然收敛），但它会丢数据 ——
// 这正说明「收敛」和「不丢数据」是两件互相独立的事。
type LWWRegister struct {
	Val int
	TS  int64
}

func (r *LWWRegister) Set(v int, ts int64) { r.Val, r.TS = v, ts }
func (r *LWWRegister) Merge(o LWWRegister) {
	if o.TS > r.TS {
		*r = o
	}
}
