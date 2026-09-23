// Lab 1 · 时间、顺序与因果
//
// 运行：  cd go/lab01 && go run .
// 调参：  go run . -skew 200 -gap 50 -n 500
//
// 配套课件：courseware/ch01-time-order.html
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"dsc/internal/tui"
)

var (
	nFlag    = flag.Int("n", 200, "Lab 1-3 的并发写轮数")
	skewFlag = flag.Int64("skew", 80, "副本 R1 的时钟偏移（毫秒）")
	gapFlag  = flag.Int64("gap", 60, "两次写入的真实间隔（毫秒）")
	jitFlag  = flag.Int64("jit", 60, "时钟偏移的随机抖动（毫秒）")
	seedFlag = flag.Int64("seed", 42, "随机种子")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：并发写 %d 轮 ｜ R1 时钟偏移 %+dms（抖动 ±%dms）｜ 写入真实间隔 %dms ｜ 种子 %d\n",
		*nFlag, *skewFlag, *jitFlag, *gapFlag, *seedFlag)
	lab11()
	lab12()
	lab13()
	lab14()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 1-1：把时空图算出来
// ═══════════════════════════════════════════════════════════════════════
func lab11() {
	tui.Head(1, 1, "Lamport 时钟与向量时钟",
		"和课件「实验 1」的经典示例完全一致 —— 浏览器里看到的数字，这里能跑出来")

	evs := BuildTrace()
	w := []int{12, 12, 12, 12, 16}
	tui.TableHead([]string{"事件", "类型", "对端", "Lamport", "向量时钟"}, w)
	for _, e := range evs {
		peer := "—"
		if e.Peer != 0 {
			for _, x := range evs {
				if x.ID == e.Peer {
					peer = x.Name()
				}
			}
		}
		tui.TableRow([]string{e.Name(), e.KindName(), peer, fmt.Sprint(e.L), e.V.String()}, w)
	}
	fmt.Println(`
    ▸ 注意 P1@t8 的向量是 [3,0,2] —— 第 2 位是 2，说明 P1 通过那条来自 P3 的消息
      "得知"了 P3 已经走了 2 步。向量时钟的每一位就是"我所知道的它走到哪了"。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 1-2：枚举所有事件对，抓 Lamport 的盲区
// ═══════════════════════════════════════════════════════════════════════
func lab12() {
	tui.Head(1, 2, "Lamport 的盲区", "枚举所有事件对，看两种时钟在哪些对上给出不同答案")

	evs := BuildTrace()
	var total, causal, conc, misled int
	var examples [][]string

	for i := 0; i < len(evs); i++ {
		for j := i + 1; j < len(evs); j++ {
			a, b := evs[i], evs[j]
			total++
			ord := Compare(a.V, b.V)

			// 独立验证：用 happens-before 的定义做可达性搜索，不依赖向量时钟
			fwd, bwd := causalPath(a, b, evs), causalPath(b, a, evs)
			want := Concurrent
			if fwd {
				want = Before
			} else if bwd {
				want = After
			}
			if ord != want {
				panic(fmt.Sprintf("向量时钟判定与 happens-before 定义不符: %s vs %s → %v / %v",
					a.Name(), b.Name(), ord, want))
			}

			if ord == Concurrent {
				conc++
				if a.L != b.L { // Lamport 给出了顺序，但它们其实并发
					misled++
					if len(examples) < 6 {
						first, second := a, b
						if a.L > b.L {
							first, second = b, a
						}
						examples = append(examples, []string{
							a.Name() + " ∥ " + b.Name(),
							fmt.Sprintf("L=%d", a.L), fmt.Sprintf("L=%d", b.L),
							"Lamport 误以为 " + first.Name() + " 在 " + second.Name() + " 之前",
						})
					}
				}
			} else {
				causal++
			}
		}
	}

	tui.Row("事件对总数", total)
	tui.Row("有因果关系（a→b 或 b→a）", causal, " 对　← 这些 Lamport 判断正确")
	tui.Row("并发（互不影响）", conc, " 对")
	tui.Row("其中被 Lamport 误导的", fmt.Sprintf("%d 对　← 占并发对的 %.0f%%", misled, float64(misled)/float64(conc)*100))
	fmt.Println()

	w := []int{24, 8, 8, 2, 46}
	al := "LRRLL"
	tui.TableHeadA([]string{"并发的事件对", "L(a)", "L(b)", "", "Lamport 会得出的错误结论"}, w, al)
	for _, ex := range examples {
		tui.TableRowA([]string{ex[0], ex[1], ex[2], "", ex[3]}, w, al)
	}
	fmt.Println(`
    ▸ 向量时钟的判定通过了独立校验：程序另外用 happens-before 的三条规则做了
      一次可达性搜索，两者结论完全一致（不一致会直接 panic）。
    ▸ Lamport 在所有"有因果关系"的对上都判断正确 —— 它从不把因果判反。
      它只是在"并发"的对上会编造出一个不存在的顺序。这就是"只能证伪，不能证实"。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 1-3：LWW 静默丢写 vs 版本向量
// ═══════════════════════════════════════════════════════════════════════
func lab13() {
	tui.Head(1, 3, "last-write-wins 到底丢了多少",
		"两个副本各收到一次写，互不知情 —— 这是一次真正的并发写")

	rnd := rand.New(rand.NewSource(*seedFlag))
	var lwwDiscard, lwwDiscardLater, vvConflict, vvDiscard int

	for i := 0; i < *nFlag; i++ {
		trueA := int64(i) * 1000
		trueB := trueA + *gapFlag // B 在真实时间上确确实实晚于 A

		jit := int64(0)
		if *jitFlag > 0 {
			jit = rnd.Int63n(2**jitFlag) - *jitFlag
		}
		stampA := trueA + *skewFlag + jit // R1 用自己（歪的）时钟打戳
		stampB := trueB                   // R2 时钟准确

		// —— 策略一：last-write-wins ——
		// 无论如何都只能留一条，另一条被丢弃。
		lwwDiscard++
		if stampA > stampB {
			lwwDiscardLater++ // 丢掉的是真实更晚的 B —— 肉眼可见的错误
		}

		// —— 策略二：版本向量 ——
		// A 写在 R1 上：[1,0]；B 写在 R2 上：[0,1]。互不 ≤ ⇒ 并发。
		vA, vB := Vector{1, 0}, Vector{0, 1}
		if Compare(vA, vB) == Concurrent {
			vvConflict++ // 检测出冲突，两条都保留为 siblings
		}
	}

	w := []int{34, 12, 12}
	tui.TableHead([]string{"", "last-write-wins", "版本向量"}, w)
	tui.TableRow([]string{"并发写轮数", fmt.Sprint(*nFlag), fmt.Sprint(*nFlag)}, w)
	tui.TableRow([]string{"检测出冲突", "0", fmt.Sprint(vvConflict)}, w)
	tui.TableRow([]string{"被丢弃的写入", fmt.Sprint(lwwDiscard), fmt.Sprint(vvDiscard)}, w)
	tui.TableRow([]string{"其中丢的是真实更晚那条", fmt.Sprint(lwwDiscardLater), "—"}, w)
	fmt.Println()
	tui.Row("LWW 的\"明显错误率\"", fmt.Sprintf("%.1f%%（丢掉真实更晚写入的比例）",
		float64(lwwDiscardLater)/float64(*nFlag)*100))
	fmt.Println(`
    ▸ 最容易被忽略的一行是「被丢弃的写入」：LWW 每一轮都丢掉一条，
      因为它必须二选一。把 -skew 设成 0，"明显错误率"会降到 0，
      但丢弃数依然是 100% —— 时钟准不准，改变的只是"丢哪一条"。
    ▸ 版本向量不假装知道答案：它把两条都留下来（siblings），
      把"该怎么合并"这个只有业务代码知道的问题交还给业务代码。
    ▸ 试试 -skew 0 和 -skew 300，看两个数字怎么变。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 1-4：HLC 在时钟回拨下的表现
// ═══════════════════════════════════════════════════════════════════════
func lab14() {
	tui.Head(1, 4, "混合逻辑时钟 HLC", "注入一次 NTP 回拨，看三种时钟谁活了下来")

	rnd := rand.New(rand.NewSource(*seedFlag + 1))
	pt := int64(12000) // 模拟的物理时钟（毫秒）
	var lam Lamport
	var hlc HLC

	type row struct {
		n          int
		pt         int64
		lam        uint64
		l, c       int64
		back, jump bool
	}
	var rows []row
	prevPT := int64(-1)
	var backCnt, maxDrift int64
	var prevHLC HLC
	monotonic := true

	step := func(jump int64) {
		if jump != 0 {
			pt += jump
		} else {
			pt += 20 + rnd.Int63n(70)
		}
		l, c := hlc.Local(pt)
		back := prevPT >= 0 && pt < prevPT
		if back {
			backCnt++
		}
		// HLC 单调性校验：(l,c) 字典序必须严格递增
		if l < prevHLC.L || (l == prevHLC.L && c <= prevHLC.C) {
			monotonic = false
		}
		prevHLC = HLC{l, c}
		if d := l - pt; d > maxDrift {
			maxDrift = d
		}
		rows = append(rows, row{len(rows) + 1, pt, lam.Local(), l, c, back, jump != 0})
		prevPT = pt
	}

	for i := 0; i < 3; i++ {
		step(0)
	}
	step(-800) // ★ NTP 把时钟往回拨了 800ms
	for i := 0; i < 4; i++ {
		step(0)
	}

	w := []int{12, 12, 11, 14, 16, 4}
	tui.TableHeadA([]string{"事件", "物理时钟", "Lamport", "HLC (l, c)", "l − 物理时钟", ""}, w, "LRRRRL")
	for _, r := range rows {
		tag := ""
		if r.jump {
			tag = " *回拨*"
		}
		mark := ""
		if r.back {
			mark = "  ⟵ 物理时钟倒退了"
		}
		tui.TableRowA([]string{fmt.Sprintf("#%d%s", r.n, tag), fmt.Sprint(r.pt),
			fmt.Sprint(r.lam), fmt.Sprintf("(%d, %d)", r.l, r.c),
			fmt.Sprintf("+%d ms", r.l-r.pt), mark}, w, "LRRRRL")
	}
	fmt.Println()
	tui.Row("物理时钟倒退次数", backCnt, " 次　← 任何基于它的排序 / TTL / 快照都会出错")
	tui.Row("HLC 是否严格单调", map[bool]string{true: "是 ✓", false: "否 ✗"}[monotonic])
	tui.Row("HLC 与物理时钟的最大偏差", maxDrift, " ms　← 有界，且回拨结束后会自动收敛回 0")
	fmt.Println(`
    ▸ 回拨发生后，HLC 的 l 冻结在回拨前的值不动，靠 c 递增维持严格单调；
      等物理时钟重新追上来，c 自动归零，l 继续贴着物理时间走。
    ▸ Lamport 也单调，但它的数值和现实时间毫无关系 —— 你没法用它做
      "给我 10:00 那一刻的快照"这类时间范围查询。HLC 两样都要到了，
      代价只是每个时间戳多一个整数。CockroachDB / MongoDB 用的就是它。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  三个思考题（答案在课件 §1.3 / §1.5）：

    1. -skew 设成 0，LWW 还会丢数据吗？先猜再跑。
    2. Lab 1-2 里"被 Lamport 误导的对数"能不能降到 0？
       如果把所有进程都两两互发一次消息会怎样？（提示：还剩多少并发对）
    3. 向量时钟能判定并发，那它能告诉你"该保留哪一条"吗？
       如果不能，谁能？

  下一站：Part 2 · 复制与一致性模型 —— 把"一致性"拆成一个精确的谱系。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
