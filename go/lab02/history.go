package main

import "sort"

// ════════════════════════════════════════════════════════════════════════
// 一致性判定器：和课件「实验 2」跑的是同一套算法。
// 这是 Jepsen 做的事情的极简版 —— 穷举所有可能的串行化顺序，
// 看是否存在一个满足对应约束的顺序。
// ════════════════════════════════════════════════════════════════════════

const Init = "⊥"

type Op struct {
	C    int    // 客户端编号
	Kind string // "w" 写 / "r" 读
	V    string // 值
	S, E int    // 真实时间上的 [发出, 返回] 区间
}

type History struct {
	Name string
	Desc string
	Ops  []Op
}

// dfsOrder 带剪枝的穷举：在 idx 这些操作里找一个合法串行化顺序。
//
//	canPlace 决定"此刻能不能把 j 放在下一个位置"——不同一致性模型的差别全在这里。
//	读的合法性检查是共同的：读必须返回当前寄存器的值。
func dfsOrder(ops []Op, idx []int, canPlace func(j int, placed map[int]bool) bool) []int {
	placed := map[int]bool{}
	var out []int
	steps := 0

	var go_ func(cur string) bool
	go_ = func(cur string) bool {
		steps++
		if steps > 500000 {
			return false
		}
		if len(out) == len(idx) {
			return true
		}
		for _, j := range idx {
			if placed[j] || !canPlace(j, placed) {
				continue
			}
			if ops[j].Kind == "r" && ops[j].V != cur {
				continue
			}
			placed[j] = true
			out = append(out, j)
			next := cur
			if ops[j].Kind == "w" {
				next = ops[j].V
			}
			if go_(next) {
				return true
			}
			out = out[:len(out)-1]
			placed[j] = false
		}
		return false
	}
	if go_(Init) {
		res := make([]int, len(out))
		copy(res, out)
		return res
	}
	return nil
}

func allIdx(ops []Op) []int {
	idx := make([]int, len(ops))
	for i := range ops {
		idx[i] = i
	}
	return idx
}

// CheckLinearizable：必须尊重真实时间序（op_i 返回早于 op_j 发出 ⇒ i 在前）。
func CheckLinearizable(ops []Op) []int {
	return dfsOrder(ops, allIdx(ops), func(j int, placed map[int]bool) bool {
		for i := range ops {
			if !placed[i] && i != j && ops[i].E < ops[j].S {
				return false
			}
		}
		return true
	})
}

// CheckSequential：去掉真实时间约束，只保留每个客户端的程序序。
// 它和线性一致的唯一区别就在这一个函数里。
func CheckSequential(ops []Op) []int {
	return dfsOrder(ops, allIdx(ops), func(j int, placed map[int]bool) bool {
		for i := range ops {
			if !placed[i] && i != j && ops[i].C == ops[j].C && ops[i].S < ops[j].S {
				return false
			}
		}
		return true
	})
}

// causalReach 计算因果序的传递闭包：程序序 + 写→读依赖。
func causalReach(ops []Op) [][]bool {
	n := len(ops)
	adj := make([][]int, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			if ops[i].C == ops[j].C && ops[i].S < ops[j].S { // 程序序
				adj[i] = append(adj[i], j)
			}
			if ops[i].Kind == "w" && ops[j].Kind == "r" && ops[j].V == ops[i].V { // 写 → 读
				adj[i] = append(adj[i], j)
			}
		}
	}
	R := make([][]bool, n)
	for s := 0; s < n; s++ {
		R[s] = make([]bool, n)
		stack, seen := []int{s}, map[int]bool{s: true}
		for len(stack) > 0 {
			x := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, y := range adj[x] {
				if !seen[y] {
					seen[y] = true
					R[s][y] = true
					stack = append(stack, y)
				}
			}
		}
	}
	return R
}

// CheckCausal：允许每个客户端有各自的串行化，但都必须尊重全局因果序。
// 返回 nil 表示不满足；否则返回 客户端 -> 该客户端眼中的顺序。
func CheckCausal(ops []Op) map[int][]int {
	R := causalReach(ops)
	out := map[int][]int{}
	for _, c := range clients(ops) {
		var idx []int
		for i, o := range ops {
			if o.Kind == "w" || o.C == c { // 所有写 + 这个客户端自己的读
				idx = append(idx, i)
			}
		}
		res := dfsOrder(ops, idx, func(j int, placed map[int]bool) bool {
			for _, i := range idx {
				if !placed[i] && i != j && R[i][j] {
					return false
				}
			}
			return true
		})
		if res == nil {
			return nil
		}
		out[c] = res
	}
	return out
}

func clients(ops []Op) []int {
	m := map[int]bool{}
	for _, o := range ops {
		m[o.C] = true
	}
	var cs []int
	for c := range m {
		cs = append(cs, c)
	}
	sort.Ints(cs)
	return cs
}

// versions 按真实开始时间给写编版本号，初始值为版本 0。
func versions(ops []Op) map[string]int {
	m := map[string]int{Init: 0}
	var ws []Op
	for _, o := range ops {
		if o.Kind == "w" {
			ws = append(ws, o)
		}
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].S < ws[j].S })
	for i, w := range ws {
		m[w.V] = i + 1
	}
	return m
}

// CheckReadYourWrites：客户端读到的值不能比它自己此前写入的更旧。
func CheckReadYourWrites(ops []Op) []string {
	V, bad := versions(ops), []string{}
	for _, c := range clients(ops) {
		mine := clientOps(ops, c)
		lastW := -1
		for _, o := range mine {
			if o.Kind == "w" {
				lastW = V[o.V]
			} else if lastW >= 0 && V[o.V] < lastW {
				bad = append(bad, "C"+itoa(c+1)+" 写入后读到了更旧的 "+o.V)
			}
		}
	}
	return bad
}

// CheckMonotonicReads：同一客户端的连续读，版本不能倒退。
func CheckMonotonicReads(ops []Op) []string {
	V, bad := versions(ops), []string{}
	for _, c := range clients(ops) {
		var rs []Op
		for _, o := range clientOps(ops, c) {
			if o.Kind == "r" {
				rs = append(rs, o)
			}
		}
		for i := 1; i < len(rs); i++ {
			if V[rs[i].V] < V[rs[i-1].V] {
				bad = append(bad, "C"+itoa(c+1)+" 先读到 "+rs[i-1].V+" 又读回了更旧的 "+rs[i].V)
			}
		}
	}
	return bad
}

func clientOps(ops []Op, c int) []Op {
	var r []Op
	for _, o := range ops {
		if o.C == c {
			r = append(r, o)
		}
	}
	sort.Slice(r, func(i, j int) bool { return r[i].S < r[j].S })
	return r
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// Histories 与课件「实验 2」的六个场景完全一致。
var Histories = []History{
	{"① 教科书式的线性一致", "C1 写完并返回后 C2 才开始读，必须读到新值",
		[]Op{{0, "w", "A", 0, 3}, {1, "r", "A", 5, 8}}},
	{"② 重叠区间里读到旧值", "读与写在时间上重叠，线性一致允许读到旧值",
		[]Op{{0, "w", "A", 2, 8}, {1, "r", Init, 3, 6}, {2, "r", "A", 10, 12}}},
	{"③ 陈旧读", "写早已返回，之后开始的读却看到初始值",
		[]Op{{0, "w", "A", 0, 3}, {1, "r", Init, 5, 8}}},
	{"④ 两个观察者，相反的顺序", "两个并发写，C3 看到 A→B，C4 看到 B→A",
		[]Op{{0, "w", "A", 0, 2}, {1, "w", "B", 1, 3},
			{2, "r", "A", 5, 6}, {2, "r", "B", 8, 9},
			{3, "r", "B", 5, 6}, {3, "r", "A", 8, 9}}},
	{"⑤ 因果倒置", "C2 读到 A 后才写 B（A 因果先于 B），C3 却先看到 B 又看回 A",
		[]Op{{0, "w", "A", 0, 2}, {1, "r", "A", 3, 4}, {1, "w", "B", 5, 7},
			{2, "r", "B", 9, 10}, {2, "r", "A", 12, 13}}},
	{"⑥ 读不到自己的写", "C1 自己写完紧接着自己读，却读到初始值",
		[]Op{{0, "w", "A", 0, 2}, {0, "r", Init, 4, 6}, {1, "r", "A", 8, 10}}},
}
