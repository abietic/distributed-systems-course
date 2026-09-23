// Lab 3-B · 日志复制与安全性
//
// 运行：  cd go/lab03b && go run .
// ★关键： go run . -naivecommit    ← 关掉「只提交当前任期」，看 Figure 8 怎么炸
//
//	go run . -slowbackoff    ← 关掉快速回退，对比 RPC 轮数
//
// 配套课件：courseware/ch03b-log-replication.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"dsc/internal/tui"
)

var (
	nF      = flag.Int("n", 5, "节点数")
	delayF  = flag.Int64("delay", 20, "单向网络延迟（毫秒）")
	jitterF = flag.Int64("jitter", 10, "延迟抖动")
	lossF   = flag.Float64("loss", 0.0, "丢包率")
	seedF   = flag.Int64("seed", 42, "随机种子")
	naiveF  = flag.Bool("naivecommit", false, "★ 去掉「只提交当前任期」的限制")
	slowF   = flag.Bool("slowbackoff", false, "关掉快速回退，每次只退 1")
	nolcF   = flag.Bool("nologcheck", false, "★ 关掉「日志至少一样新」的投票条件（退化成 Lab 3-A 的模型）")
	vF      = flag.Bool("v", false, "打印细节")
)

const tick = 5

func build(seed int64) *Cluster {
	net := NewNetwork(*nF, *delayF, *jitterF, *lossF, seed)
	c := New(*nF, net, 400, 300, seed)
	c.NaiveCommit = *naiveF
	c.SlowBackoff = *slowF
	c.NoLogCheck = *nolcF
	c.Verbose = *vF
	return c
}
func run(c *Cluster, ms int64) {
	for t := int64(0); t < ms; t += tick {
		c.Step(tick)
	}
}

func main() {
	flag.Parse()
	fmt.Printf("\n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 丢包 %.0f%% ｜ 种子 %d%s%s\n",
		*nF, *delayF, *jitterF, *lossF*100, *seedF,
		map[bool]string{true: " ｜ ★ 已关闭提交限制", false: ""}[*naiveF],
		map[bool]string{true: " ｜ 朴素回退", false: ""}[*slowF])
	lab3b1()
	lab3b2()
	lab3b3()
	lab3b4()
	lab3b5()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3B-1：日志复制的基本正确性
// ═══════════════════════════════════════════════════════════════════════
func lab3b1() {
	tui.HeadN("3B-1", "日志复制", "提交 200 条命令，验证所有节点的日志逐条相同")

	c := build(*seedF)
	run(c, 1500) // 先选出 Leader
	submitted := 0
	for round := 0; round < 200; round++ {
		if c.Submit(fmt.Sprintf("SET k%d=%d", round%10, round)) {
			submitted++
		}
		run(c, 30)
	}
	run(c, 2000)

	ok, why := c.LogsIdentical()
	w := []int{26, 14, 4, 44}
	al := "LRLL"
	tui.TableHeadA([]string{"指标", "数值", "", "说明"}, w, al)
	tui.TableRowA([]string{"提交的命令数", fmt.Sprint(submitted), "", "客户端发起"}, w, al)
	tui.TableRowA([]string{"最高 commitIndex", fmt.Sprint(c.MaxCommit()), "", "已达成共识的日志条数"}, w, al)
	tui.TableRowA([]string{"AppendEntries 成功", fmt.Sprint(c.Appends), "", "次"}, w, al)
	tui.TableRowA([]string{"AppendEntries 被拒", fmt.Sprint(c.Rejects), "", "一致性检查未通过"}, w, al)
	tui.TableRowA([]string{"所有节点日志一致", map[bool]string{true: "✓ 是", false: "✗ 否"}[ok], "",
		map[bool]string{true: "逐条比对通过", false: why}[ok]}, w, al)
	fmt.Println()
	for i := 0; i < *nF && i < 3; i++ {
		tui.Row(fmt.Sprintf("N%d 日志前 20 条的 term", i+1), truncate(c.LogStr(i), 60))
	}
	fmt.Println(`
    ▸ 所有节点的日志逐条相同，且状态机在每个 index 上执行了相同的命令。
    ▸ 「被拒」是 0，因为这是个全新集群、没有任何故障，所有节点的日志从一开始就一致，
      Leader 的 nextIndex 乐观猜测一次就猜对了。
      只有当日志真的分叉过（换过 Leader、发生过分区），一致性检查才会拒绝 ——
      Lab 3B-4 会故意制造那种局面。`)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3B-2：三条安全性断言，在故障注入下
// ═══════════════════════════════════════════════════════════════════════
func lab3b2() {
	tui.HeadN("3B-2", "三条安全性断言", "边跑边杀节点、边制造分区，每个 tick 都检查")

	w := []int{30, 14, 12, 12, 16}
	al := "LRRRR"
	tui.TableHeadA([]string{"场景", "提交的命令", "已提交", "违反", "安全性"}, w, al)

	scenario := func(label string, f func(c *Cluster)) {
		c := build(*seedF)
		run(c, 1200)
		f(c)
		nv := len(c.Violations)
		tui.TableRowA([]string{label, fmt.Sprint(c.cmdSeq), fmt.Sprint(c.MaxCommit()),
			fmt.Sprint(nv), map[bool]string{true: "✗ 被破坏", false: "✓ 成立"}[nv > 0]}, w, al)
		if nv > 0 {
			for _, v := range c.Violations {
				tui.Row("  ↳", v)
			}
		}
	}

	scenario("① 无故障，持续写入", func(c *Cluster) {
		for i := 0; i < 120; i++ {
			c.Submit(fmt.Sprintf("cmd%d", i))
			run(c, 25)
		}
		run(c, 1500)
	})

	scenario("② 边写边杀 Leader", func(c *Cluster) {
		for i := 0; i < 120; i++ {
			c.Submit(fmt.Sprintf("cmd%d", i))
			run(c, 25)
			if i%30 == 29 {
				if l := c.Leader(); l != nil {
					c.net.Kill(l.id)
					run(c, 900)
					c.net.Revive(l.id)
				}
			}
		}
		run(c, 2500)
	})

	scenario("③ 反复 3|2 分区", func(c *Cluster) {
		grp := []int{0, 0, 0, 1, 1}
		for i := 0; i < 120; i++ {
			c.Submit(fmt.Sprintf("cmd%d", i))
			run(c, 25)
			if i%40 == 20 {
				c.net.Partition(grp)
			}
			if i%40 == 39 {
				c.net.Heal()
			}
		}
		c.net.Heal()
		run(c, 3000)
	})

	scenario("④ 30% 丢包 + 随机杀节点", func(c *Cluster) {
		c.net.loss = 0.3
		for i := 0; i < 150; i++ {
			c.Submit(fmt.Sprintf("cmd%d", i))
			run(c, 25)
			if i%25 == 24 {
				k := c.rnd.Intn(*nF)
				if c.net.Down(k) {
					c.net.Revive(k)
				} else {
					c.net.Kill(k)
				}
			}
		}
		for i := 0; i < *nF; i++ {
			c.net.Revive(i)
		}
		c.net.loss = 0
		run(c, 4000)
	})

	fmt.Println(`
    ▸ 断言的是三条：选举安全性、日志匹配性质、状态机安全性。
      任何一条被破坏都会记录下来并打印现场。
    ▸ 注意场景③④里「提交的命令」远多于「已提交」—— 分区和丢包期间客户端
      发出的请求进了 Leader 的日志却没能提交，最后被截断丢弃。
      客户端看到的是超时，然后数据消失。这就是 §3.10 说的第三态。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3B-3：Figure 8 精确复现
// ═══════════════════════════════════════════════════════════════════════
func lab3b3() {
	tui.HeadN("3B-3", "Figure 8 复现", "手工构造论文里那五个节点的状态，然后按规则推进")

	// 直接构造 Figure 8 阶段 (c) 的局面，比等随机调度撞出来可靠得多。
	mk := func(naive bool) *Cluster {
		net := NewNetwork(5, 20, 5, 0, 1)
		c := New(5, net, 400, 300, 1)
		c.NaiveCommit = naive
		e := func(idx, term int, cmd string) Entry { return Entry{Index: idx, Term: term, Cmd: cmd} }

		// (a)(b) 的结果：S1/S2 有 term2 的 index2，S5 有 term3 的 index2
		logs := [][]Entry{
			{e(1, 1, "A"), e(2, 2, "X")}, // S1
			{e(1, 1, "A"), e(2, 2, "X")}, // S2
			{e(1, 1, "A")},               // S3
			{e(1, 1, "A")},               // S4
			{e(1, 1, "A"), e(2, 3, "Y")}, // S5 —— 同一个 index 上的不同内容
		}
		for i, lg := range logs {
			c.nodes[i].log = append([]Entry{{Index: 0, Term: 0}}, lg...)
			c.nodes[i].term = 4
		}
		c.nodes[4].term = 3
		// (c)：S1 在任期 4 当选，已把 term2 那条复制给 S3，并在 index3 写下当前任期的日志
		s1 := c.nodes[0]
		s1.state = Leader
		s1.term = 4
		c.nodes[2].log = append(c.nodes[2].log, e(2, 2, "X")) // S3 也拿到了 term2 那条
		s1.log = append(s1.log, e(3, 4, "Z"))                 // S1 自己的当前任期日志
		for j := range c.nodes {
			s1.nextIndex[j] = s1.last().Index + 1
			s1.matchIndex[j] = 0
		}
		s1.matchIndex[0], s1.matchIndex[1], s1.matchIndex[2] = 3, 2, 2 // S1/S2/S3 都有 index2
		return c
	}

	// 为了并排展示，把两次结果收集起来
	type res struct {
		committed, overwritten bool
		viol                   []string
	}
	collect := func(naive bool) res {
		c := mk(naive)
		s1 := c.nodes[0]
		c.maybeCommit(s1)
		committed := s1.commitIndex >= 2
		saved := ""
		if e, ok := c.globalCommitted[2]; ok {
			saved = e.Cmd
		}
		c.net.Kill(0)
		s5 := c.nodes[4]
		s5.term, s5.state = 5, Leader
		for j := range c.nodes {
			s5.nextIndex[j], s5.matchIndex[j] = 1, 0
		}
		s5.matchIndex[4] = s5.last().Index
		for i := 1; i <= 3; i++ {
			n := c.nodes[i]
			n.term, n.state = 5, Follower
			n.log = append([]Entry{{Index: 0, Term: 0}}, s5.log[1:]...)
			s5.matchIndex[i] = n.last().Index
		}
		c.becomeLeader(s5)
		c.checkSafety()
		c.maybeCommit(s5)
		for i := 1; i <= 4; i++ {
			c.apply(c.nodes[i])
		}
		return res{committed, saved != "" && c.nodes[1].log[2].Cmd != saved, c.Violations}
	}
	a, b := collect(false), collect(true)

	w := []int{40, 18, 18}
	al := "LRR"
	tui.TableHeadA([]string{"阶段 (c)→(d) 发生了什么", "① 完整 Raft", "② 关掉限制"}, w, al)
	yn := func(v bool) string {
		if v {
			return "是"
		}
		return "否"
	}
	tui.TableRowA([]string{"index2 (term2) 复制到过半了吗", "是 (3/5)", "是 (3/5)"}, w, al)
	tui.TableRowA([]string{"S1 把它标记为已提交了吗", yn(a.committed), yn(b.committed)}, w, al)
	tui.TableRowA([]string{"S5 当选后覆盖掉它了吗", "是", "是"}, w, al)
	tui.TableRowA([]string{"被覆盖的是「已提交」的日志吗", yn(a.overwritten), yn(b.overwritten)}, w, al)
	tui.TableRowA([]string{"安全性断言", map[bool]string{true: "✗ 被破坏", false: "✓ 成立"}[len(a.viol) > 0],
		map[bool]string{true: "✗ 被破坏", false: "✓ 成立"}[len(b.viol) > 0]}, w, al)
	fmt.Println()
	if len(b.viol) > 0 {
		tui.Row("② 报出的违反", "")
		for _, v := range b.viol {
			fmt.Println("      " + v)
		}
	}
	fmt.Println(`
    ▸ 两边的物理事实完全一样：那条 term2 的日志都复制到了 3/5 个节点，
      也都被 S5 覆盖掉了。唯一的区别是——① 从没把它叫做「已提交」。
    ▸ Raft 没有去阻止覆盖（那需要改选举规则，代价大得多），
      而是确保「被覆盖的东西从来没被承诺过」。这是一个非常克制的修补。
    ▸ 对客户端的含义：① 里客户端收到的是超时（第三态），数据可能在也可能不在；
      ② 里客户端收到的是"成功"，然后数据消失了。后者才是真正的事故。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3B-4：回退策略对照
// ═══════════════════════════════════════════════════════════════════════
func lab3b4() {
	tui.HeadN("3B-4", "冲突回退：朴素 vs 快速", "让一个 follower 落后很多，数它追平要几轮 RPC")

	w := []int{26, 16, 16, 18}
	al := "LRRR"
	tui.TableHeadA([]string{"回退策略", "被拒次数", "追平耗时(ms)", "最终是否追平"}, w, al)

	try := func(label string, slow bool) {
		net := NewNetwork(*nF, *delayF, *jitterF, 0, *seedF)
		c := New(*nF, net, 400, 300, *seedF)
		c.SlowBackoff = slow
		run(c, 1200)
		c.net.Kill(4) // 把 N5 隔离，让它错过一大段日志
		for i := 0; i < 260; i++ {
			c.Submit(fmt.Sprintf("cmd%d", i))
			run(c, 12)
			if i == 90 || i == 180 { // 中途换两次 Leader，制造多个任期
				if l := c.Leader(); l != nil {
					c.net.Kill(l.id)
					run(c, 900)
					c.net.Revive(l.id)
				}
			}
		}
		before := c.Rejects
		t0 := c.net.Now()
		c.net.Revive(4) // N5 回来，开始追赶
		caught := int64(-1)
		for t := int64(0); t < 40000; t += tick {
			c.Step(tick)
			if c.nodes[4].last().Index >= c.MaxCommit() && c.nodes[4].commitIndex >= c.MaxCommit() {
				caught = c.net.Now() - t0
				break
			}
		}
		tui.TableRowA([]string{label, fmt.Sprint(c.Rejects - before),
			map[bool]string{true: fmt.Sprint(caught), false: "未追平"}[caught >= 0],
			map[bool]string{true: "✓", false: "✗"}[caught >= 0]}, w, al)
	}
	try("朴素回退（每次 −1）", true)
	try("快速回退（按任期跳）", false)

	fmt.Println(`
    ▸ 一个任期内的日志是同一个 Leader 连续写下的，要么整段一致要么整段不一致
      （日志匹配性质）。所以在冲突任期内部逐条回退是纯粹的浪费。
    ▸ 真实系统里日志常有几万条却只跨越几个任期，这个优化能把几万轮 RPC 降到个位数。
      落后节点追不上，等于集群实际少了一个副本 —— 容错能力悄悄下降了。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3B-5：高任期的少数派节点回归时，能不能抢到 Leader
//
//	这一节补上 Lab 3-A 的一个已知简化：3-A 的模型没有日志，
//	所以投票时缺了「candidate 的日志至少和我一样新」这一条。
//	下面用 A/B 对照把那个简化的后果测出来。
//
// ═══════════════════════════════════════════════════════════════════════
func lab3b5() {
	tui.HeadN("3B-5", "高任期节点回归", "3|2 分区让少数派任期涨到几十，再恢复，看谁当选")

	const trials = 40
	type out struct{ minorityWon, majorityWon, noLeader, viol, maxMinTerm int }

	run40 := func(noLogCheck, writeDuringPartition bool) out {
		var o out
		for seed := int64(1); seed <= trials; seed++ {
			net := NewNetwork(5, 30, 15, 0, seed)
			c := New(5, net, 400, 300, seed)
			c.NoLogCheck = noLogCheck
			adv := func(ms int64) {
				for t := int64(0); t < ms; t += tick {
					c.Step(tick)
				}
			}
			adv(2000)
			for i := 0; i < 20; i++ { // 分区前先提交一批日志
				c.Submit(fmt.Sprintf("pre%d", i))
				adv(30)
			}
			adv(800)
			grp := []int{0, 0, 0, 1, 1} // 多数派 {N1,N2,N3}｜少数派 {N4,N5}
			net.Partition(grp)
			if writeDuringPartition {
				// 多数派继续接收写入 ⇒ 它的日志会长出新任期的条目
				for i := 0; i < 25; i++ {
					c.Submit(fmt.Sprintf("during%d", i))
					adv(40)
				}
			}
			adv(12000) // 少数派在这里不停超时，任期疯涨；日志则完全冻结
			if c.nodes[3].term > o.maxMinTerm {
				o.maxMinTerm = c.nodes[3].term
			}
			net.Heal()
			adv(9000)
			l := c.Leader()
			switch {
			case l == nil:
				o.noLeader++
			case grp[l.id] == 1:
				o.minorityWon++
			default:
				o.majorityWon++
			}
			if len(c.Violations) > 0 {
				o.viol++
			}
		}
		return o
	}

	a := run40(false, true)  // 完整 Raft，分区期间多数派持续写入
	b := run40(true, true)   // 去掉日志检查（退化成 Lab 3-A 的模型）
	d := run40(false, false) // 完整 Raft，但分区期间多数派没有任何写入

	w := []int{32, 16, 16, 18}
	al := "LRRR"
	tui.TableHeadA([]string{"恢复后由谁当选（各 40 次）", "① 完整 Raft", "② 无日志检查", "③ 多数派没写入"}, w, al)
	tui.TableRowA([]string{"多数派节点（日志更新）", fmt.Sprint(a.majorityWon), fmt.Sprint(b.majorityWon), fmt.Sprint(d.majorityWon)}, w, al)
	tui.TableRowA([]string{"★ 少数派节点（日志陈旧）", fmt.Sprint(a.minorityWon), fmt.Sprint(b.minorityWon), fmt.Sprint(d.minorityWon)}, w, al)
	tui.TableRowA([]string{"没能选出 Leader", fmt.Sprint(a.noLeader), fmt.Sprint(b.noLeader), fmt.Sprint(d.noLeader)}, w, al)
	tui.TableRowA([]string{"安全性违反", fmt.Sprint(a.viol), fmt.Sprint(b.viol), fmt.Sprint(d.viol)}, w, al)
	fmt.Println()
	tui.Row("分区期间少数派任期最高涨到", fmt.Sprintf("%d　← 而它们的日志一条都没长", a.maxMinTerm))
	fmt.Println(`
    ▸ ① 少数派一次都赢不了。它们的任期虽然涨到几十（回来时确实会逼现任 Leader 退位、
      触发一次不必要的选举），但日志停在分区前 —— 投票时那条「至少和我一样新」把它们全挡住了。
      高任期节点回归是「有破坏力，没有危险」：损失的是可用性，不是安全性。
    ▸ ② 去掉日志检查后陈旧节点频繁当选 —— 这正是 Lab 3-A 那个模型的行为。
      3-A 没有日志，那一条无法实现；这个对照就是补上那个缺口。
    ▸ ③ 是最微妙的一列：分区期间多数派【没有任何写入】，于是两边日志一模一样 ——
      少数派根本不"陈旧"，它当选完全合法，也毫无危害。
      【所以"陈旧节点赢不了"这句话的准确版本是："日志真的落后的节点赢不了"。】
      而真实系统里 ③ 这种窗口几乎不存在，因为新 Leader 一当选就会追加一条 no-op ——
      它的 lastLogTerm 立刻变成当前任期，把这个窗口直接关掉。
      这就是 §3.12 说的 no-op 的第二重作用：既帮前任日志过提交线，又堵住旧节点回来抢位子的缝。
    ▸ 顺带纠正一个常见误解：选举【从不看 commitIndex】。
      比较的只有 (lastLogTerm, lastLogIndex)，而且先比 term。
      一个节点无法验证别人报上来的 commitIndex 是否可信，所以它不能参与投票判断。
    ▸ 真实系统用 PreVote 让那次不必要的选举也别发生：candidate 在真正 term++ 之前
      先试探性问一轮"我发起选举你们会投我吗"，被隔离的节点得不到肯定答复，任期就不会涨。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  三个思考题（答案在课件 §3.12 / §3.13）：

    1. -naivecommit 下被覆盖的那条日志，在被覆盖之前有几个节点持有它？
       "过半"为什么救不了它？
    2. 如果同时打开 -naivecommit 并让新 Leader 上任就追加一条 no-op 日志，
       Figure 8 还会发生吗？先想清楚再改代码试。
    3. 把 -loss 调到 0.4，Lab 3B-2 的日志匹配断言会不会被破坏？为什么？
    4. Lab 3B-5 的 ① 里少数派一次都没赢。但如果少数派那边有一个【僵尸 Leader】
       （它会持续追加日志，只是提交不了），情况会变吗？什么条件下它能赢？
       赢了之后会不会破坏安全性？（提示：想想 no-op）

  下一节：Part 3-C · 持久化、快照、成员变更与线性一致读。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
