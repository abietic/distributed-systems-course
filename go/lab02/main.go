// Lab 2 · 复制与一致性模型
//
// 运行：  cd go/lab02 && go run .
// 调参：  go run . -n 5 -w 3 -r 3 -ops 500 -lag 200
//
// 配套课件：courseware/ch02-replication.html
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"dsc/internal/tui"
)

var (
	nF    = flag.Int("n", 5, "副本数 N")
	wF    = flag.Int("w", 3, "写 quorum W")
	rF    = flag.Int("r", 3, "读 quorum R")
	opsF  = flag.Int("ops", 400, "每组实验的读写次数")
	lagF  = flag.Int("lag", 3, "复制延迟（多少次操作之后异步副本才收到）")
	seedF = flag.Int64("seed", 42, "随机种子")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：N=%d ｜ W=%d ｜ R=%d ｜ 操作数 %d ｜ 复制延迟 %d 次操作 ｜ 种子 %d\n",
		*nF, *wF, *rF, *opsF, *lagF, *seedF)
	lab21()
	lab22()
	lab23()
	lab24()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 2-1：W + R > N 到底保证了什么
// ═══════════════════════════════════════════════════════════════════════
func lab21() {
	tui.Head(2, 1, "Quorum：W + R > N", "同一份工作负载，换不同的 W/R，数陈旧读")

	w := []int{22, 7, 7, 7, 13, 16}
	al := "LRRRRR"
	tui.TableHeadA([]string{"配置", "N", "W", "R", "W+R>N", "陈旧读比例"}, w, al)

	run := func(label string, n, ww, rr int) float64 {
		q := NewQuorumKV(n, ww, rr, *lagF, *seedF)
		for i := 0; i < *opsF; i++ {
			q.Write(fmt.Sprintf("v%d", i))
			q.Read()
		}
		ratio := float64(q.StaleReads) / float64(q.TotalReads) * 100
		tui.TableRowA([]string{label, fmt.Sprint(n), fmt.Sprint(ww), fmt.Sprint(rr),
			map[bool]string{true: "是", false: "否"}[q.Guaranteed()],
			fmt.Sprintf("%.1f%%", ratio)}, w, al)
		return ratio
	}

	N := *nF
	maj := N/2 + 1
	strict := run("多数派 W=R=⌊N/2⌋+1", N, maj, maj)
	run("W=N, R=1", N, N, 1)
	run("W=1, R=N", N, 1, N)
	loose := run("W=R=⌊N/2⌋（故意不足）", N, max(1, N/2), max(1, N/2))
	run("W=1, R=1（最快最弱）", N, 1, 1)

	fmt.Println()
	tui.Row("W+R>N 的三行陈旧读", fmt.Sprintf("%.1f%%　← 应当恒为 0", strict))
	tui.Row("W+R≤N 的陈旧读", fmt.Sprintf("%.1f%%　← 交集可能为空，读不到最新", loose))
	fmt.Println(`
    ▸ 前三行的 W+R 都 > N，陈旧读恒为 0 —— 交集性质是一条数学保证，不是概率。
    ▸ 后两行故意让 W+R ≤ N，陈旧读立刻出现。它未必是错误配置：
      Cassandra 的 ONE/ONE 就是这么设的，用一致性换最低延迟和最高可用性。
    ▸ 试试 -lag 20，看后两行的比例怎么变、前三行会不会变。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 2-2：分区来了，CP 和 AP 各自付出什么代价
// ═══════════════════════════════════════════════════════════════════════
func lab22() {
	tui.Head(2, 2, "CAP：把取舍变成两个数字",
		fmt.Sprintf("%d 个副本被切成 %d|%d 两侧，同一份写入负载跑两种策略", *nF, (*nF+1)/2, *nF/2))

	majSize := (*nF + 1) / 2 // 多数派一侧的副本数
	quorum := *nF/2 + 1

	type result struct{ ok, fail, conflict int }
	sim := func(cp bool) result {
		rnd := rand.New(rand.NewSource(*seedF))
		var res result
		writtenMajority, writtenMinority := map[int]bool{}, map[int]bool{}
		for i := 0; i < *opsF; i++ {
			key := rnd.Intn(*opsF / 4)            // 少量 key，制造冲突机会
			toMajority := rnd.Intn(*nF) < majSize // 请求随机落到某一侧
			if cp {
				// CP：只有能凑齐 quorum 的一侧才接受写，少数派直接拒绝
				if toMajority && majSize >= quorum {
					res.ok++
					writtenMajority[key] = true
				} else {
					res.fail++
				}
			} else {
				// AP：两侧都接受写，分区恢复后再合并
				res.ok++
				if toMajority {
					writtenMajority[key] = true
				} else {
					writtenMinority[key] = true
				}
			}
		}
		for k := range writtenMajority { // 分区恢复：同一个 key 两侧都写过 ⇒ 冲突
			if writtenMinority[k] {
				res.conflict++
			}
		}
		return res
	}

	cp, ap := sim(true), sim(false)
	w := []int{16, 14, 14, 16, 18}
	al := "LRRRR"
	tui.TableHeadA([]string{"策略", "写入成功", "写入失败", "成功率", "恢复后的冲突 key"}, w, al)
	tui.TableRowA([]string{"CP（少数派拒绝）", fmt.Sprint(cp.ok), fmt.Sprint(cp.fail),
		fmt.Sprintf("%.1f%%", float64(cp.ok)/float64(*opsF)*100), fmt.Sprint(cp.conflict) + "（不可能有）"}, w, al)
	tui.TableRowA([]string{"AP（两侧都收）", fmt.Sprint(ap.ok), fmt.Sprint(ap.fail),
		fmt.Sprintf("%.1f%%", float64(ap.ok)/float64(*opsF)*100), fmt.Sprint(ap.conflict) + " 个待合并"}, w, al)
	fmt.Println(`
    ▸ 这就是 CAP 的全部内容：分区期间，你要么损失一部分写入的可用性（CP），
      要么接下一堆需要合并的冲突（AP）。没有第三个选项，因为 P 不可选。
    ▸ 注意 CP 那一行的失败率 ≈ 少数派副本占比 —— 少数派侧的请求全部被拒。
      这不是 bug，是设计：宁可停下，不可出错。
    ▸ AP 那一行的冲突数就是你欠下的债，分区恢复后必须用 LWW / siblings / CRDT 还上。
      Lab 2-4 会告诉你这三种还法的差别有多大。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 2-3：一致性判定器（Jepsen 的极简版）
// ═══════════════════════════════════════════════════════════════════════
func lab23() {
	tui.Head(2, 3, "一致性判定器", "穷举所有串行化顺序，判定每条历史属于谱系的哪一档")

	w := []int{26, 10, 10, 10, 10, 10}
	al := "LRRRRR"
	tui.TableHeadA([]string{"执行历史", "线性一致", "顺序一致", "因果一致", "读己之写", "单调读"}, w, al)

	yn := func(b bool) string {
		if b {
			return "✓"
		}
		return "✗"
	}
	for _, h := range Histories {
		lin := CheckLinearizable(h.Ops) != nil
		seq := CheckSequential(h.Ops) != nil
		cau := CheckCausal(h.Ops) != nil
		ryw := len(CheckReadYourWrites(h.Ops)) == 0
		mono := len(CheckMonotonicReads(h.Ops)) == 0

		// 谱系必须是嵌套的：线性 ⟹ 顺序 ⟹ 因果。不成立说明判定器有 bug。
		if lin && !seq {
			panic("判定器自相矛盾：线性一致却不顺序一致 —— " + h.Name)
		}
		if seq && !cau {
			panic("判定器自相矛盾：顺序一致却不因果一致 —— " + h.Name)
		}

		tui.TableRowA([]string{h.Name, yn(lin), yn(seq), yn(cau), yn(ryw), yn(mono)}, w, al)
	}

	fmt.Println()
	// 展开一条最有意思的历史
	h := Histories[3]
	tui.Row("细看 "+h.Name, h.Desc)
	if cau := CheckCausal(h.Ops); cau != nil {
		for _, c := range clients(h.Ops) {
			if order, ok := cau[c]; ok {
				parts := make([]string, len(order))
				for i, j := range order {
					o := h.Ops[j]
					parts[i] = fmt.Sprintf("%s(%s)@C%d", o.Kind, o.V, o.C+1)
				}
				tui.Row("  C"+itoa(c+1)+" 眼中的顺序", strings.Join(parts, " → "))
			}
		}
	}
	fmt.Println(`
    ▸ 判定器内置了自检：谱系必须满足 线性 ⟹ 顺序 ⟹ 因果，违反直接 panic。
    ▸ ④ 那一行是最值得盯的：不存在任何单一全序能同时解释 C3 和 C4，
      但允许每人有自己的顺序之后就都说得通了 —— 因为那两个写是并发的。
      这正是因果一致比顺序一致可用性高的全部原因。
    ▸ ⑥ 违反读己之写，同时也违反了因果一致 —— 因果一致蕴含读己之写。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 2-4：CRDT 的三律，以及"收敛 ≠ 正确"
// ═══════════════════════════════════════════════════════════════════════
func lab24() {
	tui.Head(2, 4, "CRDT：收敛不等于正确", "随机分区、随机合并顺序，跑一千遍")

	const R = 3
	rnd := rand.New(rand.NewSource(*seedF))

	// —— 一次实验：三个副本在分区期间各自累加，然后以随机顺序两两合并 ——
	trial := func(rnd *rand.Rand) (total, g, lww int) {
		gs := make([]GCounter, R)
		lw := make([]LWWRegister, R)
		for i := range gs {
			gs[i] = NewGCounter(R)
		}
		ts := int64(0)
		for i := 0; i < R; i++ {
			k := 1 + rnd.Intn(5)
			for j := 0; j < k; j++ {
				ts++
				gs[i].Inc(i)               // G-Counter：只增自己那一槽
				lw[i].Set(lw[i].Val+1, ts) // LWW：本地读改写
				total++
			}
		}
		// 随机顺序做若干轮两两合并
		for round := 0; round < 6; round++ {
			a, b := rnd.Intn(R), rnd.Intn(R)
			if a != b {
				gs[a].Merge(gs[b])
				lw[a].Merge(lw[b])
			}
		}
		for i := 0; i < R; i++ { // 确保全连通
			for j := 0; j < R; j++ {
				if i != j {
					gs[i].Merge(gs[j])
					lw[i].Merge(lw[j])
				}
			}
		}
		return total, gs[0].Value(), lw[0].Val
	}

	var trials, gExact, lwwLost, totalClicks, lwwKept int
	var diverged int
	for t := 0; t < 1000; t++ {
		tot, g, l := trial(rnd)
		trials++
		totalClicks += tot
		lwwKept += l
		if g == tot {
			gExact++
		} else {
			diverged++
		}
		lwwLost += tot - l
	}

	w := []int{28, 14, 20, 24}
	al := "LRRR"
	tui.TableHeadA([]string{"数据类型", "实验次数", "结果精确", "累计丢失"}, w, al)
	tui.TableRowA([]string{"G-Counter（逐位取最大）", fmt.Sprint(trials),
		fmt.Sprintf("%d 次（%.0f%%）", gExact, float64(gExact)/float64(trials)*100), "0"}, w, al)
	tui.TableRowA([]string{"LWW-Register（取时间戳大）", fmt.Sprint(trials),
		"0 次", fmt.Sprintf("%d / %d 次点击", lwwLost, totalClicks)}, w, al)
	fmt.Println()
	tui.Row("G-Counter 与真值不符", fmt.Sprintf("%d 次　← 应当恒为 0", diverged))
	tui.Row("LWW 丢失率", fmt.Sprintf("%.1f%%", float64(lwwLost)/float64(totalClicks)*100))

	// —— 三律验证：交换、结合、幂等 ——
	a, b, c := NewGCounter(3), NewGCounter(3), NewGCounter(3)
	a.Inc(0)
	a.Inc(0)
	b.Inc(1)
	c.Inc(2)
	c.Inc(2)
	c.Inc(2)
	eq := func(x, y GCounter) bool {
		for i := range x {
			if x[i] != y[i] {
				return false
			}
		}
		return true
	}
	m1 := a.Clone()
	m1.Merge(b)
	m1.Merge(c) // (a∪b)∪c
	m2 := c.Clone()
	m2.Merge(a)
	m2.Merge(b) // (c∪a)∪b —— 换个顺序
	m3 := m1.Clone()
	m3.Merge(m1)
	m3.Merge(b)
	m3.Merge(b) // 重复合并
	fmt.Println()
	tui.Row("交换律 + 结合律", map[bool]string{true: "✓ 两种合并顺序结果相同", false: "✗"}[eq(m1, m2)])
	tui.Row("幂等律", map[bool]string{true: "✓ 重复合并结果不变", false: "✗"}[eq(m1, m3)])

	// —— OR-Set：并发的加与删，加优先 ——
	s1, s2 := NewORSet(), NewORSet()
	s1.Add("牛奶", 1)
	s2.Merge(s1)    // 两个副本都看到了"牛奶"
	s2.Remove("牛奶") // 副本 2 删掉它
	s1.Add("牛奶", 2) // 副本 1 并发地又加了一次（新 tag）
	s1.Merge(s2)
	s2.Merge(s1)
	tui.Row("OR-Set 并发加删", fmt.Sprintf("两副本收敛=%v，元素存在=%v（add-wins）",
		s1.Size() == s2.Size(), s1.Has("牛奶")))

	fmt.Println(`
    ▸ G-Counter 在 1000 次随机分区 + 随机合并顺序下，结果 100% 精确 ——
      因为逐位取最大满足交换律、结合律、幂等律，合并顺序根本不影响结果。
    ▸ LWW 每次也都"收敛"了（三个副本值相同），但值是错的。
      收敛是关于「最终一致」的性质，不丢数据是关于「语义」的性质，两者正交。
    ▸ OR-Set 的 add-wins：并发的"加"和"删"中加优先，因为删除只能删掉
      自己已经观察到的那些 tag，删不掉自己没见过的新 tag。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  三个思考题（答案在课件 §2.5 / §2.6）：

    1. 把 -w 2 -r 2 -n 5（W+R=4 ≤ 5）跑一遍，陈旧读比例是多少？
       再把 -lag 从 3 调到 30，这个比例怎么变？为什么 W+R>N 那几行不受影响？
    2. Lab 2-2 里，如果把分区改成 4|1 而不是 3|2，CP 的成功率会怎么变？
       AP 的冲突数呢？先猜再改代码。
    3. LWW-Register 是一个合法的 CRDT（满足三律、必然收敛），
       但它丢数据。那"CRDT"这个保证到底值多少钱？

  下一站：Part 3 · 共识算法 —— 从 FLP 到 Raft，Lab 3 用 Go 从零写一个能跑的 Raft。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
