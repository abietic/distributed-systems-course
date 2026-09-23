// Lab 3-A · 共识与 Raft 选举
//
// 运行：  cd go/lab03a && go run .
// 调参：  go run . -loss 0.3 -delay 80 -norandom
//
//	go run . -unsafe        ← 去掉「每任期一票」，看安全性怎么炸
//
// 配套课件：courseware/ch03a-consensus-election.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"dsc/internal/tui"
)

var (
	nF      = flag.Int("n", 5, "节点数（取奇数）")
	delayF  = flag.Int64("delay", 30, "单向网络延迟（毫秒）")
	jitterF = flag.Int64("jitter", 15, "延迟抖动（毫秒）")
	lossF   = flag.Float64("loss", 0.0, "丢包率 0~1")
	baseF   = flag.Int64("base", 400, "选举超时下限（毫秒）")
	winF    = flag.Int64("window", 300, "选举超时随机区间宽度（毫秒）")
	seedF   = flag.Int64("seed", 42, "随机种子，同种子完全可复现")
	noRandF = flag.Bool("norandom", false, "关闭随机化选举超时（制造分裂投票）")
	unsafeF = flag.Bool("unsafe", false, "★ 去掉「每个任期只投一票」规则，演示安全性被破坏")
	vF      = flag.Bool("v", false, "打印每一条状态转换")
)

const tick = 5 // 虚拟时钟每次推进的毫秒数

func build() *Cluster {
	win := *winF
	if *noRandF {
		win = 0
	}
	net := NewNetwork(*nF, *delayF, *jitterF, *lossF, *seedF)
	c := New(*nF, net, *baseF, win, *seedF, *unsafeF)
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
	win := *winF
	if *noRandF {
		win = 0
	}
	fmt.Printf("\n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 丢包 %.0f%% ｜ 选举超时 %d~%dms ｜ 种子 %d%s\n",
		*nF, *delayF, *jitterF, *lossF*100, *baseF, *baseF+win, *seedF,
		map[bool]string{true: " ｜ ★ UNSAFE 模式", false: ""}[*unsafeF])

	lab3a1()
	lab3a2()
	lab3a3()
	lab3a4()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3A-1：确定性网络 —— 同一个种子必定复现同一次执行
// ═══════════════════════════════════════════════════════════════════════
func lab3a1() {
	tui.HeadN("3A-1", "确定性网络模拟", "调共识算法的第一件事：让 bug 能复现")

	sig := func(seed int64) string {
		save := *seedF
		*seedF = seed
		c := build()
		run(c, 3000)
		*seedF = save
		return fmt.Sprintf("Leader=%v maxTerm=%d 选举=%d 分裂=%d",
			leaderName(c), c.MaxTerm(), c.Elections, c.Splits)
	}
	a, b, d := sig(42), sig(42), sig(7)
	tui.Row("种子 42 · 第一次", a)
	tui.Row("种子 42 · 第二次", b)
	tui.Row("种子 7", d)
	tui.Row("同种子结果一致", map[bool]string{true: "✓ 完全可复现", false: "✗ 不可复现（有 bug）"}[a == b])
	if a == b {
		fmt.Println(`
    ▸ 没有 goroutine、没有真实时间，只有一个虚拟时钟和一个按时间排序的消息队列。
      这样任何一次诡异的执行都能靠种子复现，而不是"跑一百遍偶尔挂一次"。
    ▸ 这是 MIT 6.5840 的 labrpc、以及 etcd 的 raft 测试框架采用的同一套思路：
      把并发与时间从被测逻辑里彻底剥离出去。`)
	}
}

func leaderName(c *Cluster) string {
	if l := c.Leader(); l != nil {
		return fmt.Sprintf("N%d(t%d)", l.id+1, l.term)
	}
	return "无"
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3A-2：五个场景
// ═══════════════════════════════════════════════════════════════════════
func lab3a2() {
	tui.HeadN("3A-2", "五个场景", "每个 tick 都断言「任一任期至多一个 Leader」")

	w := []int{28, 20, 10, 9, 8, 16}
	al := "LLRRRR"
	tui.TableHeadA([]string{"场景", "结果 Leader", "最大term", "选举数", "分裂", "选举安全性"}, w, al)

	safe := func(c *Cluster) string {
		if len(c.Violations) > 0 {
			return "✗ 被破坏！"
		}
		return "✓ 成立"
	}

	// —— 场景 1：冷启动选主 ——
	c1 := build()
	run(c1, 3000)
	tui.TableRowA([]string{"① 冷启动", leaderName(c1), fmt.Sprint(c1.MaxTerm()),
		fmt.Sprint(c1.Elections), fmt.Sprint(c1.Splits), safe(c1)}, w, al)

	// —— 场景 2：杀死 Leader，观察重新选主 ——
	c2 := build()
	run(c2, 2000)
	if l := c2.Leader(); l != nil {
		c2.net.Kill(l.id)
		c2.logf("外力：杀死 Leader N%d", l.id+1)
	}
	run(c2, 4000)
	tui.TableRowA([]string{"② 杀死 Leader 后重选", leaderName(c2), fmt.Sprint(c2.MaxTerm()),
		fmt.Sprint(c2.Elections), fmt.Sprint(c2.Splits), safe(c2)}, w, al)

	// —— 场景 3：3|2 分区 ——
	c3 := build()
	run(c3, 2000)
	grp := []int{0, 0, 0, 1, 1}
	c3.net.Partition(grp)
	c3.logf("外力：网络分区 {N1,N2,N3} | {N4,N5}")
	run(c3, 6000)
	majL := c3.LeadersIn(grp, 0)
	minL := c3.LeadersIn(grp, 1)
	tui.TableRowA([]string{"③ 3|2 分区", fmt.Sprintf("多数派%d 少数派%d", len(majL), len(minL)),
		fmt.Sprint(c3.MaxTerm()), fmt.Sprint(c3.Elections), fmt.Sprint(c3.Splits), safe(c3)}, w, al)

	// —— 场景 4：分区恢复 ——
	c3.net.Heal()
	c3.logf("外力：分区修复")
	run(c3, 5000)
	terms := map[int]bool{}
	for _, n := range c3.nodes {
		terms[n.term] = true
	}
	tui.TableRowA([]string{"④ 分区恢复后收敛", leaderName(c3), fmt.Sprint(c3.MaxTerm()),
		fmt.Sprint(c3.Elections), fmt.Sprint(c3.Splits), safe(c3)}, w, al)

	// —— 场景 5：连续随机杀节点 ——
	c5 := build()
	rnd := c5.rnd
	for round := 0; round < 12; round++ {
		run(c5, 1500)
		i := rnd.Intn(*nF)
		if c5.net.Down(i) {
			c5.net.Revive(i)
		} else if aliveCount(c5) > *nF/2+1 { // 始终保留过半存活，否则集群本来就该停摆
			c5.net.Kill(i)
		}
	}
	run(c5, 4000)
	tui.TableRowA([]string{"⑤ 反复随机杀/救节点", leaderName(c5), fmt.Sprint(c5.MaxTerm()),
		fmt.Sprint(c5.Elections), fmt.Sprint(c5.Splits), safe(c5)}, w, al)

	fmt.Println()
	tui.Row("③ 多数派侧的 Leader", fmt.Sprintf("%d 个（能提交日志）", len(majL)))
	tui.Row("③ 少数派侧的 Leader", fmt.Sprintf("%d 个 —— %s", len(minL),
		map[bool]string{true: "★ 僵尸 Leader：自认为是主，但一条日志都提交不了", false: "已退位"}[len(minL) > 0]))
	tui.Row("③ 少数派能联系到几个节点", fmt.Sprintf("%d 个，过半需要 %d 个 ⇒ 提交不了任何东西",
		reachable(c3, grp, 1), c3.majority()))
	tui.Row("④ 恢复后集群的任期数", fmt.Sprintf("%d 种　← 应当收敛到 1~2 种", len(terms)))
	tui.Row("④ 最终状态", c3.Summary())

	fmt.Println(`
    ▸ 场景③是 CAP 的具体样子，但它给出的答案比教科书更微妙：
      如果分区前的 Leader 恰好落在少数派一侧，基础 Raft 不会让它主动退位 ——
      它继续给同侧的 Follower 发心跳，自认为还是 Leader。这叫「僵尸 Leader」。
      它不违反安全性（拿不到过半确认，一条日志都提交不了），但客户端把写请求
      发给它会一直超时，读请求还可能读到陈旧数据。
    ▸ 工程上用 CheckQuorum 解决：Leader 定期确认自己还能联系到过半节点，
      否则主动退位。etcd 默认开启它。这也是「用了 Raft 就一定线性一致」
      这句话的又一个反例 —— 线性一致读还需要 ReadIndex 或 Lease Read（3-C）。
    ▸ 场景④注意少数派带回来的高任期：它一接触集群就会逼现任 Leader 退位，
      引发一次不必要的选举。这正是 Raft 论文用 PreVote 优化解决的问题。
    ▸ 五个场景、上万个 tick，选举安全性一次都没被破坏 —— 这不是运气，
      是「过半票 + 每任期一票 ⇒ 两个过半集合必相交」这条数学事实。`)
}

// reachable 数一数某个分区组里还有多少个存活节点 —— 用来判断它能不能凑齐过半。
func reachable(c *Cluster, grp []int, g int) int {
	k := 0
	for _, n := range c.nodes {
		if !c.net.Down(n.id) && grp[n.id] == g {
			k++
		}
	}
	return k
}

func aliveCount(c *Cluster) int {
	k := 0
	for _, n := range c.nodes {
		if !c.net.Down(n.id) {
			k++
		}
	}
	return k
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3A-3：随机化选举超时的作用
// ═══════════════════════════════════════════════════════════════════════
func lab3a3() {
	tui.HeadN("3A-3", "随机化选举超时", "关掉它，看分裂投票怎么把集群拖住")

	w := []int{24, 18, 12, 12, 16}
	al := "LLRRR"
	tui.TableHeadA([]string{"选举超时策略", "选出 Leader", "耗时(ms)", "分裂投票", "发起过的选举"}, w, al)

	try := func(label string, window int64) {
		net := NewNetwork(*nF, *delayF, *jitterF, *lossF, *seedF)
		c := New(*nF, net, *baseF, window, *seedF, false)
		var elected int64 = -1
		for t := int64(0); t < 20000; t += tick {
			c.Step(tick)
			if elected < 0 && c.Leader() != nil {
				elected = net.Now()
			}
		}
		et := "未能选出"
		if elected >= 0 {
			et = fmt.Sprint(elected)
		}
		tui.TableRowA([]string{label, map[bool]string{true: "✓ " + leaderName(c), false: "✗ 一直没有"}[c.Leader() != nil],
			et, fmt.Sprint(c.Splits), fmt.Sprint(c.Elections)}, w, al)
	}
	try("固定超时（窗口 0）", 0)
	try(fmt.Sprintf("随机超时（窗口 %d）", *winF), *winF)

	fmt.Println(`
    ▸ 固定超时下所有节点同时醒来、同时给自己投票，谁也拿不到过半票，
      超时后再来一轮 —— 任期一路飙升却选不出 Leader。这就是活锁。
    ▸ 随机化只用一个随机数就打破了对称性：总有人先醒，赶在别人之前收齐票。
    ▸ 注意这不是"随机化让选举更快"，而是"随机化让选举能够终止"。
      FLP 说确定性算法在异步系统里不能保证终止，Raft 换到了随机化这条赛道上，
      代价是只保证「以概率 1 终止」而非「N 步内一定终止」。工程上这就够了。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3A-4：去掉「每任期一票」会怎样
// ═══════════════════════════════════════════════════════════════════════
func lab3a4() {
	tui.HeadN("3A-4", "把安全性拆掉给你看", "去掉「每个任期只投一票」这一条，其他不变")

	w := []int{28, 22, 18, 16, 14}
	al := "LRRRR"
	tui.TableHeadA([]string{"规则", "同任期最多 Leader 数", "违反的种子", "选出了 Leader", "选举安全性"}, w, al)

	trial := func(label string, unsafe bool) {
		worst, violations, elected := 0, 0, 0
		const trials = 60
		for seed := int64(1); seed <= trials; seed++ { // 多跑几个种子，撞上并发选举的概率才够高
			// 窗口很窄 + 有丢包 ⇒ 经常出现两个 Candidate 同时索票
			net := NewNetwork(*nF, *delayF, 40, 0.15, seed)
			c := New(*nF, net, 300, 90, seed, unsafe)
			run(c, 6000)
			if c.MaxLeadersPerTerm > worst {
				worst = c.MaxLeadersPerTerm
			}
			if len(c.Violations) > 0 {
				violations++
			}
			if c.Leader() != nil {
				elected++
			}
		}
		tui.TableRowA([]string{label, fmt.Sprintf("%d 个", worst),
			fmt.Sprintf("%d/%d 个种子", violations, trials),
			fmt.Sprintf("%d/%d", elected, trials),
			map[bool]string{true: "✗ 被破坏", false: "✓ 成立"}[violations > 0]}, w, al)
		return
	}
	trial("① 完整 Raft（每任期一票）", false)
	trial("② 去掉「每任期一票」", true)

	fmt.Println(`
    ▸ ① 60 个种子、上万个 tick，同一任期的 Leader 数最多就是 1 —— 因为两个过半集合必然相交，
      交集里那个节点在同一任期投了两票，这是不可能的。
    ▸ ② 一旦允许一个节点在同一任期投多票，两个 Candidate 就能各自凑齐过半，
      同一任期出现两个 Leader —— 这就是脑裂，两个 Leader 会各写各的日志，
      数据从此分叉且无法自动合并。
    ▸ 所以「每任期只投一票」不是实现细节，它和「过半」一起构成了
      Raft 全部安全性的地基。而且 votedFor 必须持久化 ——
      节点重启后如果忘了自己投过票，同样会破坏它（Lab 3-C 会处理）。
      试试 go run . -unsafe -v 看具体是怎么炸的。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  三个思考题（答案在课件 §3.6 / §3.7）：

    1. 把 -loss 调到 0.5，集群还能选出 Leader 吗？需要多久？
       再调到 0.8 呢？在什么丢包率下它彻底选不出来？
    2. -norandom 时任期会一路飙升。如果这时候网络突然恢复正常，
       这个高任期会对集群造成什么影响？（提示：PreVote）
    3. 场景③里少数派的两个节点，任期涨到了多少？
       如果分区持续一小时，它们的任期会涨到什么量级？这有害吗？

  下一节：Part 3-B · 日志复制与五条安全性属性 —— 包括论文 Figure 8
  那个反直觉的结论：Leader 不能靠"数副本数"来提交前任任期的日志。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
