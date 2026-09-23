// Lab 4-A · 单机事务：隔离级别、MVCC 与写偏斜
//
// 运行：  cd go/lab04a && go run .
// 调参：  go run . -trials 500 -doctors 3 -txns 12 -ops 4
//
// 配套课件：courseware/ch04a-transactions.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"dsc/internal/tui"
)

var (
	trialsF  = flag.Int("trials", 200, "4A-3：并发试验次数")
	doctorsF = flag.Int("doctors", 2, "4A-3：在岗医生数")
	leaversF = flag.Int("leavers", 2, "4A-3：同时请假的人数")
	txnsF    = flag.Int("txns", 12, "4A-4：并发事务数")
	opsF     = flag.Int("ops", 4, "4A-4：每事务操作数")
	seedF    = flag.Int64("seed", 42, "随机种子基准")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：医生 %d 人 ｜ 同时请假 %d 人 ｜ 试验 %d 次 ｜ 并发事务 %d × %d 操作 ｜ 种子 %d\n",
		*doctorsF, *leaversF, *trialsF, *txnsF, *opsF, *seedF)
	lab4a1()
	lab4a2()
	lab4a3()
	lab4a4()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 4A-1：异常 × 隔离级别矩阵
// ═══════════════════════════════════════════════════════════════════════
func lab4a1() {
	tui.HeadN("4A-1", "并发异常 × 隔离级别矩阵", "这张表不是抄来的，每一格都是迷你事务引擎真的跑了一遍")

	w := []int{22, 4, 12, 12, 12, 14, 12}
	al := "LLRRRRR"
	hdr := []string{"异常", ""}
	for _, L := range Levels {
		hdr = append(hdr, shortName(L))
	}
	tui.TableHeadA(hdr, w, al)

	for _, sc := range Scenarios {
		row := []string{sc.Name, ""}
		if !sc.ANSI {
			row[1] = "*"
		}
		for _, L := range Levels {
			bad, _, e := RunScenario(sc, L)
			cell := "挡住"
			if bad {
				cell = "异常发生"
			}
			if e.Deadlock {
				cell += "(死锁)"
			}
			row = append(row, cell)
		}
		tui.TableRowA(row, w, al)
	}
	fmt.Println("\n    * = 不在 ANSI SQL-92 的异常清单里 —— 而它们恰恰是线上真正出事的那几种。")

	// 把最关键的两行时序打出来
	for _, id := range []string{"lu", "ws"} {
		for _, sc := range Scenarios {
			if sc.ID != id {
				continue
			}
			for _, L := range []string{RR, RRLOCK} {
				bad, detail, e := RunScenario(sc, L)
				fmt.Printf("\n    ── %s × %s ──────────────────────────────\n", sc.Name, LevelName[L])
				for _, t := range e.Trace {
					pad := ""
					if t.T == 2 {
						pad = strings.Repeat(" ", 44)
					}
					fmt.Printf("      T%d %s%s\n", t.T, pad, t.Txt)
				}
				verdict := "正确"
				if bad {
					verdict = "★ 不变量被破坏"
				}
				fmt.Printf("      ⇒ %s　%s\n", detail, verdict)
			}
		}
	}

	fmt.Println(`
    结论
      · 看「丢失更新」和「写偏斜」两行：RR 那一列是红的。
        把 MySQL 从 RC 调到 RR，你买到的是「不可重复读」和「幻读」，
        【没有】买到「丢失更新」和「写偏斜」——而后两者才是业务不变量被破坏的那一类。
      · RR 和 RR+FOR UPDATE 的全部差别，是把隐式的读变成了显式的锁。
        数据库看不见你代码里那个 if，加锁是唯一能让它看见的办法。`)
}

func shortName(l string) string {
	switch l {
	case RU:
		return "RU"
	case RC:
		return "RC"
	case RR:
		return "RR"
	case RRLOCK:
		return "RR+FOR UPD"
	}
	return "SERIAL"
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 4A-2：ReadView 可见性判定
// ═══════════════════════════════════════════════════════════════════════
func lab4a2() {
	tui.HeadN("4A-2", "MVCC 的 ReadView 可见性判定", "RC 和 RR 的实现差异只有一行：ReadView 什么时候创建")

	fmt.Println(`
    同一条版本链、同一套四步判定算法，只改「ReadView 何时创建」这一件事。`)

	chain := []Version{{32, 500}, {28, 400}, {24, 300}, {18, 200}, {12, 100}}
	fmt.Print("\n    balance 这一行的版本链（新 → 旧）：")
	for i, v := range chain {
		if i > 0 {
			fmt.Print("  →")
		}
		fmt.Printf("  [trx %d = %d]", v.Trx, v.Val)
	}
	fmt.Println()

	try := func(label string, mids []int, creator, next int) {
		rv := &ReadView{Mids: mids, Max: next, Creator: creator}
		rv.Min = next
		if len(mids) > 0 {
			rv.Min = mids[0]
		}
		e := &Engine{}
		fmt.Printf("\n    %s\n", label)
		fmt.Printf("      ReadView{ m_ids=%v  min=%d  max=%d  creator=%d }\n", mids, rv.Min, rv.Max, rv.Creator)
		picked := false
		for _, v := range chain {
			ok := e.visible(v.Trx, rv)
			mark := "不可见"
			why := ""
			switch {
			case v.Trx == rv.Creator:
				why = "trx == creator ⇒ 我自己改的"
			case v.Trx < rv.Min:
				why = fmt.Sprintf("trx < min(%d) ⇒ 建 ReadView 前就已提交", rv.Min)
			case v.Trx >= rv.Max:
				why = fmt.Sprintf("trx >= max(%d) ⇒ 在我之后才开始", rv.Max)
			case contains(mids, v.Trx):
				why = "trx 在 m_ids 里 ⇒ 我建 ReadView 时它还没提交"
			default:
				why = "不在 m_ids 且 < max ⇒ 已提交"
			}
			if ok && !picked {
				mark = "★ 可见"
				picked = true
				why += "　← 第一个可见的，就是答案"
			} else if ok {
				mark = "可见"
			}
			fmt.Printf("      trx %-3d = %-4d %s  %s\n", v.Trx, v.Val, tui.PadR(mark, 8), why)
		}
	}

	try("① 事务 25 建 ReadView 时，trx 28 和 32 都还没提交", []int{28, 32}, 25, 40)
	try("② trx 28 提交了 —— RC 会在下一条语句重新采集活跃集合", []int{32}, 25, 40)

	fmt.Println(`
    结论
      · 同一条版本链，仅仅因为「活跃事务集合」采集的时刻不同，
        返回值就从 300 跳到了 400。
      · RC：每条语句都新建 ReadView ⇒ 上面两种情况会在同一个事务里先后出现 ⇒ 不可重复读。
        RR：整个事务只建一次并复用     ⇒ 永远停留在情况 ①     ⇒ 可重复读。
      · 【这就是 RC 和 RR 的全部实现差异。】版本链、判定算法、undo log 结构，两者完全一样。
      · 顺带纠一个极常见的错误：版本链存在 undo log 里，不是 redo log。
        redo log 是物理日志，只用于崩溃恢复的 roll-forward，和可见性判断毫无关系。`)
}

func contains(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 4A-3：写偏斜 —— 医院值班表
// ═══════════════════════════════════════════════════════════════════════
func skewSteps(c int, rnd *Rand) []*Op {
	per := make([][]*Op, c)
	for t := 1; t <= c; t++ {
		per[t-1] = []*Op{
			op(t, "begin"),
			{T: t, Kind: "scan", Pred: POnCall, Lockable: true},
			{T: t, Kind: "write", Key: "doc:" + itoa(t), Val: 0, GuardMin: 2, HasGuardMin: true},
			op(t, "commit"),
		}
	}
	idx := make([]int, c)
	out := []*Op{}
	left := 4 * c
	for left > 0 {
		avail := []int{}
		for i := 0; i < c; i++ {
			if idx[i] < len(per[i]) {
				avail = append(avail, i)
			}
		}
		p := avail[rnd.Intn(len(avail))]
		out = append(out, per[p][idx[p]])
		idx[p]++
		left--
	}
	return out
}

func lab4a3() {
	tui.HeadN("4A-3", "写偏斜：医院值班表", "不变量「任何时刻至少 1 个医生在岗」，看谁守得住")

	D, C, N := *doctorsF, *leaversF, *trialsF
	if C > D {
		C = D
	}
	fmt.Printf(`
    %d 个医生在岗，%d 个人同时发起请假。每个人的事务都是：
      BEGIN; SELECT count(*) WHERE on_call=true;  if (>= 2) UPDATE 自己 = 不在岗; COMMIT
    每次试验随机交错这些步骤，跑 %d 次。
`, D, C, N)

	w := []int{26, 18, 16, 18, 18}
	al := "LRRRR"
	tui.TableHeadA([]string{"隔离配置", "不变量被破坏", "平均最终在岗", "出现过锁等待", "主动放弃请假"}, w, al)

	for _, lv := range []string{RC, RR, RRLOCK, SER} {
		bad, sumOn, blocked, gaveUp := 0, 0, 0, 0
		for s := 1; s <= N; s++ {
			rnd := NewRand(int64(s) * 2654435761)
			init := map[string]int{}
			for i := 1; i <= D; i++ {
				init["doc:"+itoa(i)] = 1
			}
			e := NewEngine(lv, init)
			steps := skewSteps(C, rnd)
			if lv == SER {
				steps = Serialize(steps)
			}
			e.Run(steps)
			on := e.CountWhere(POnCall)
			sumOn += on
			if on < 1 {
				bad++
			}
			if e.Blocked() {
				blocked++
			}
			for _, t := range e.Trace {
				if strings.Contains(t.Txt, "放弃写入") {
					gaveUp++
				}
			}
		}
		tui.TableRowA([]string{LevelName[lv],
			fmt.Sprintf("%d / %d", bad, N),
			fmt.Sprintf("%.2f 人", float64(sumOn)/float64(N)),
			fmt.Sprintf("%d / %d", blocked, N),
			fmt.Sprintf("%d 次", gaveUp)}, w, al)
	}

	fmt.Println(`
    结论
      · RC 和 RR 的违规次数【几乎一样】。把隔离级别从 RC 调到 RR，对写偏斜没有任何帮助——
        RR 加强的是「我读到的不变」，而这里的问题是【你读到的确实没变，只是它已经过时了】。
      · 加上 FOR UPDATE 后违规降到 0，代价是大量锁等待，以及若干次
        「有人查到人不够，主动放弃请假」——后者正是我们想要的行为：
        不是数据库报错，而是应用自己发现前提不成立。
      · 加锁把「读」这个动作物化成了一件别人看得见的事，
        于是数据库终于知道你的读和写之间有因果关系。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 4A-4：2PL vs SSI
// ═══════════════════════════════════════════════════════════════════════
func lab4a4() {
	tui.HeadN("4A-4", "2PL vs SSI：两张不同的账单", "一个付「等待」，一个付「白做的功」")

	C, O := *txnsF, *opsF
	fmt.Printf(`
    %d 个并发事务，每个 %d 个操作，系统每 tick 最多执行 %d 个操作。
      2PL：拿不到锁就等（等待不占 CPU，做过的功不浪费），等待图成环 = 死锁中止。
      SSI：谁也不等，全速跑完再验证；验证不过就整个事务从头再来（白做的功照样烧 CPU）。
`, C, O, PAR)

	w := []int{12, 3, 12, 12, 14, 12, 3, 12, 12, 14, 12}
	al := "LLRRRRLRRRR"
	tui.TableHeadA([]string{"热点", "", "2PL 耗时", "中止", "白做操作", "白做占比",
		"", "SSI 耗时", "中止", "白做操作", "白做占比"}, w, al)

	for hh := 0; hh <= 100; hh += 20 {
		a := Sim2PL(MkTxns(C, O, float64(hh)/100, NewRand(*seedF*2654435761)))
		b := SimSSI(MkTxns(C, O, float64(hh)/100, NewRand(*seedF*2654435761)))
		pctOf := func(r CCResult) string {
			if r.Work == 0 {
				return "—"
			}
			return fmt.Sprintf("%.0f%%", float64(r.Wasted)/float64(r.Work)*100)
		}
		tui.TableRowA([]string{fmt.Sprintf("%d%%", hh), "",
			fmt.Sprintf("%d", a.Tick), fmt.Sprintf("%d", a.Aborts), fmt.Sprintf("%d", a.Wasted), pctOf(a), "",
			fmt.Sprintf("%d", b.Tick), fmt.Sprintf("%d", b.Aborts), fmt.Sprintf("%d", b.Wasted), pctOf(b)}, w, al)
	}

	fmt.Println(`
    结论
      · 看两列「白做占比」：热点集中度往右拉，2PL 的几乎不涨，SSI 的一路涨到有效功的几倍。
        而两者的【墙上时间】相差不大——这正是这个对比的重点：
        SSI 省下来的等待时间，是用真金白银的 CPU 换的。
      · 所以没有哪一个「总是更好」：
          冲突低、事务短 ⇒ SSI 更合适（谁也不等，读事务完全无锁，延迟分布好看得多）
          冲突高、事务长 ⇒ 2PL 更稳  （等待难看，但做过的功不会被扔掉，行为可预测）
      · 还有一个数字上看不见的代价：SSI 要求应用层有【正确的】重试逻辑。
        重试时如果复用了上一次读到的值，那比不用 SSI 还危险——
        你把一个本来会被数据库拦下的错误，变成了一个静默的错误。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  Part 4-A 到此结束

  这一章的一句话：【隔离级别是打折出售的「世界是静止的」这个假设，
  而折扣的具体内容，各家数据库同名不同物。】

  最该带走的三条：
    1. 版本链在 undo log，不在 redo log；RC 与 RR 的实现差异只有
       「ReadView 何时创建」这一行。
    2. InnoDB 的 RR 防幻读靠两套机制：快照读靠 ReadView，当前读靠 Next-Key Lock。
       但它【防不住写偏斜】，Oracle 的 SERIALIZABLE 也防不住。
    3. 能变成约束的不变量，一律变成约束；剩下的，把读物化成锁。

  下一站 Part 4-B · 分布式事务的提交协议：
  2PC 与 XA、它的阻塞问题、TCC、Saga、本地消息表、事务消息——
  以及一个贯穿始终的问题：当「提交」这个动作本身要跨机器时，
  Part 0 的那个「第三态」又回来了。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
