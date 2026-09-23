// Lab 3-C · 持久化、快照、成员变更与线性一致读
//
// 运行：  cd go/lab03c && go run .
// 调参：  go run . -lose votedfor     ← 只看丢掉 votedFor 的后果
//
//	go run . -lose term | log
//
// 配套课件：courseware/ch03c-production-raft.html
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
	seedF   = flag.Int64("seed", 42, "随机种子")
	loseF   = flag.String("lose", "", "只跑某一种失去持久化的场景：term | votedfor | log")
	trialsF = flag.Int("trials", 60, "每组实验的种子数")
)

const tick = 5

func mk(seed int64, cfg func(*Cluster)) *Cluster {
	return mkw(seed, 300, 0, cfg)
}

// mkw 允许指定选举超时窗口与丢包率。
// 窗口设成 0（固定超时）会让多个节点【同时】发起选举 ——
// 这正是暴露「votedFor 没落盘」的必要条件：得先有两个 candidate 在抢同一个任期。
func mkw(seed int64, window int64, loss float64, cfg func(*Cluster)) *Cluster {
	net := NewNetwork(*nF, *delayF, *jitterF, loss, seed)
	c := New(*nF, net, 400, window, seed)
	c.LeaseMs = 250
	if cfg != nil {
		cfg(c)
	}
	return c
}
func run(c *Cluster, ms int64) {
	for t := int64(0); t < ms; t += tick {
		c.Step(tick)
	}
}

func main() {
	flag.Parse()
	fmt.Printf("\n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 每组 %d 个种子 ｜ 种子基准 %d\n",
		*nF, *delayF, *jitterF, *trialsF, *seedF)
	lab3c1()
	lab3c2()
	lab3c3()
	lab3c4()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3C-1：持久化 —— 丢掉哪一样，哪条安全性倒下
// ═══════════════════════════════════════════════════════════════════════
func lab3c1() {
	tui.HeadN("3C-1", "持久化与崩溃重启", "分别丢掉三样状态中的一样，看哪条安全性倒下")

	w := []int{28, 14, 14, 14, 18}
	al := "LRRRR"
	tui.TableHeadA([]string{"磁盘上丢了什么", "违规的种子", "选举安全性", "状态机安全性", "结论"}, w, al)

	try := func(label string, lt, lv, ll bool) {
		bad, elec, sm := 0, 0, 0
		for seed := int64(1); seed <= int64(*trialsF); seed++ {
			// 固定超时 + 丢包 ⇒ 经常出现两个 candidate 抢同一个任期，
			// 这是让「忘记投过票」真正造成脑裂的必要土壤。
			c := mkw(*seedF+seed, 0, 0.12, func(c *Cluster) {
				c.LoseTerm, c.LoseVote, c.LoseLog = lt, lv, ll
			})
			run(c, 1200)
			for i := 0; i < 80; i++ {
				c.Submit(fmt.Sprintf("SET k%d=%d", i%5, i))
				run(c, 30)
				if i%8 == 7 {
					// 先杀 Leader：固定超时下，剩下的 follower 会几乎同时超时、
					// 同时变成 candidate 抢同一个任期 —— 这是脑裂的土壤。
					if l := c.Leader(); l != nil {
						c.net.Kill(l.id)
					}
					run(c, 420)
					// 选举正酣时，把一个刚投过票的节点掐掉又立刻拉起来 ——
					// 它会带着"失忆"回到同一场选举里。
					k := c.rnd.Intn(*nF)
					if !c.net.Down(k) {
						c.net.Kill(k)
						run(c, 30)
						c.Restart(k)
					}
					run(c, 250)
					for j := 0; j < *nF; j++ {
						if c.net.Down(j) {
							c.Restart(j)
						}
					}
					run(c, 350)
				}
			}
			run(c, 3000)
			if len(c.Violations) > 0 {
				bad++
				for _, v := range c.Violations {
					if strings.Contains(v, "选举安全性") {
						elec++
						break
					}
				}
				for _, v := range c.Violations {
					if strings.Contains(v, "状态机") || strings.Contains(v, "Leader 完整性") {
						sm++
						break
					}
				}
			}
		}
		verdict := "✓ 安全"
		if bad > 0 {
			verdict = "✗ 被破坏"
		}
		tui.TableRowA([]string{label, fmt.Sprintf("%d/%d", bad, *trialsF),
			fmt.Sprint(elec), fmt.Sprint(sm), verdict}, w, al)
	}

	// —— 构造场景：直接把「两个 candidate 抢同一任期」摆出来，不靠随机调度撞 ——
	constructed := func(loseVote bool) bool {
		net := NewNetwork(5, 20, 0, 0, 1)
		c := New(5, net, 400, 0, 1)
		c.LoseVote = loseVote
		for _, n := range c.nodes { // 全体处于任期 5
			n.term = 5
			n.persist()
		}
		n1, n2, n5 := c.nodes[0], c.nodes[1], c.nodes[4]
		for _, cd := range []*Node{n1, n2} { // N1、N2 同时成为 Candidate
			cd.state = Candidate
			cd.votedFor = cd.id
			cd.votes = map[int]bool{cd.id: true}
			cd.persist()
		}
		c.nodes[2].votedFor = 0 // N3 投 N1
		c.nodes[2].persist()
		n1.votes[2] = true
		c.nodes[3].votedFor = 1 // N4 投 N2
		c.nodes[3].persist()
		n2.votes[3] = true

		n5.votedFor = 0 // N5 投 N1 ⇒ N1 拿到 3 票当选
		n5.persist()
		n1.votes[4] = true
		c.becomeLeader(n1)

		c.net.Kill(4) // ★ N5 崩溃重启
		c.Restart(4)

		// N2 再来索票。votedFor 有没有恢复，决定了 N5 会不会投第二次。
		c.handle(Msg{From: 1, To: 4, Type: MsgRequestVote, Term: 5,
			LastLogIndex: n2.last().Index, LastLogTerm: n2.last().Term})
		if n5.votedFor == 1 {
			n2.votes[4] = true
			if len(n2.votes) >= n2.majorityOf() {
				c.becomeLeader(n2)
			}
		}
		c.checkSafety()
		return c.twoLeadersEver
	}

	if *loseF == "" {
		try("（什么都不丢，正确）", false, false, false)
		try("★ 丢失 votedFor", false, true, false)
		try("丢失 currentTerm", true, false, false)
		try("丢失 log[]", false, false, true)
	} else {
		try("（什么都不丢，正确）", false, false, false)
		switch *loseF {
		case "votedfor":
			try("★ 丢失 votedFor", false, true, false)
		case "term":
			try("丢失 currentTerm", true, false, false)
		case "log":
			try("丢失 log[]", false, false, true)
		}
	}
	fmt.Println()
	tui.Row("构造场景 · votedFor 已落盘", map[bool]string{
		true: "✗ 出现两个 Leader", false: "✓ N5 记得投过票，拒绝第二次索票 ⇒ 只有一个 Leader"}[constructed(false)])
	tui.Row("构造场景 · votedFor 丢失", map[bool]string{
		true:  "✗ 同一任期出现两个 Leader —— 票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次",
		false: "✓ 安全"}[constructed(true)])
	fmt.Println(`
    ▸ 上面两行是【构造】出来的场景：直接把"两个 candidate 抢同一任期"摆出来，
      不依赖随机调度去撞。这也说明随机故障注入为什么不够 ——
      有些 bug 需要非常特定的交错才会现形，这正是 TLA+ 和 Jepsen 存在的理由（Part 6）。
    ▸ 表格里 votedFor 那一行的随机试验可能是 0，不代表它安全，只代表没撞上。
    ▸ votedFor 是最容易被漏掉的一样，也是后果最直接的：节点重启后忘了自己投过票，
      同一任期可以再投一次 ⇒ 两个 candidate 各自凑齐过半 ⇒ 脑裂。
    ▸ log 丢失破坏的是 Leader 完整性："持有已提交日志的节点构成过半"这个事实变成了假的，
      于是后续选举可能选出不含该条目的 Leader，已提交的数据凭空消失。
    ▸ currentTerm 丢失让节点失去时间感，无法识别过时的 Leader 和 candidate。
    ▸ 判断什么必须落盘的通用直觉：【可推导的结论不用存，对外许下的承诺必须存】。
      commitIndex 是结论（Leader 会告诉你），votedFor 是承诺（只有自己知道）。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3C-2：快照与 InstallSnapshot
// ═══════════════════════════════════════════════════════════════════════
func lab3c2() {
	tui.HeadN("3C-2", "日志压缩与 InstallSnapshot", "把一个节点隔离到落后超过快照点，再放回来")

	c := mk(*seedF, func(c *Cluster) { c.SnapThreshold = 30 })
	run(c, 1500)
	c.net.Kill(4) // N5 掉线，错过一大段日志
	for i := 0; i < 220; i++ {
		c.Submit(fmt.Sprintf("SET k%d=%d", i%8, i))
		run(c, 20)
	}
	run(c, 1500)
	l := c.Leader()
	beforeSnaps := c.InstallSnaps
	logLen := 0
	snapAt := 0
	if l != nil {
		logLen = l.last().Index - l.snapIdx
		snapAt = l.snapIdx
	}

	c.net.Revive(4) // N5 回来追赶
	caught := int64(-1)
	t0 := c.net.Now()
	for t := int64(0); t < 20000; t += tick {
		c.Step(tick)
		if l != nil && c.nodes[4].lastApplied >= l.commitIndex && l.commitIndex > 0 {
			caught = c.net.Now() - t0
			break
		}
	}

	w := []int{30, 16, 4, 32}
	al := "LRLL"
	tui.TableHeadA([]string{"指标", "数值", "", "说明"}, w, al)
	tui.TableRowA([]string{"提交的命令数", fmt.Sprint(c.cmdSeq), "", "客户端发起"}, w, al)
	tui.TableRowA([]string{"Leader 做过的快照", fmt.Sprint(c.Snapshots), "", "每 30 条压缩一次"}, w, al)
	tui.TableRowA([]string{"快照点 lastIncludedIndex", fmt.Sprint(snapAt), "", "这之前的日志已丢弃"}, w, al)
	tui.TableRowA([]string{"Leader 保留的日志条数", fmt.Sprint(logLen), "", "而不是全部 " + fmt.Sprint(c.cmdSeq) + " 条"}, w, al)
	tui.TableRowA([]string{"发出的 InstallSnapshot", fmt.Sprint(c.InstallSnaps - beforeSnaps), "", "N5 落后太多，走快照追赶"}, w, al)
	tui.TableRowA([]string{"N5 追平耗时", map[bool]string{true: fmt.Sprint(caught) + " ms", false: "未追平"}[caught >= 0], "", ""}, w, al)
	ok, why := c.LogsIdentical()
	tui.TableRowA([]string{"日志仍然一致", map[bool]string{true: "✓ 是", false: "✗ 否"}[ok], "",
		map[bool]string{true: "快照没有破坏日志匹配性质", false: why}[ok]}, w, al)
	tui.Row("安全性违反", fmt.Sprintf("%d 条", len(c.Violations)))
	fmt.Println(`
    ▸ Leader 只保留了最近几十条日志，其余压缩进了快照 —— 磁盘和重放时间都是常数级。
    ▸ N5 落后到快照点之前，Leader 发现 nextIndex ≤ lastIncludedIndex，
      于是改发 InstallSnapshot 而不是逐条补日志。
    ▸ 快照里必须带 (lastIncludedIndex, lastIncludedTerm)：日志被截断后，
      AppendEntries 的一致性检查还要拿它们当"被截断部分的代表"来比较。
      少了这两个字段，日志匹配性质在截断处就断了。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3C-3：成员变更裂脑
// ═══════════════════════════════════════════════════════════════════════
func lab3c3() {
	tui.HeadN("3C-3", "成员变更：直接跳 vs 单节点", "让不同节点在不同时刻切换配置，看会不会选出两个 Leader")

	// 直接从 {0,1,2} 跳到 {0,1,2,3,4}：先让 2,3,4 切到新配置，0,1 还留在旧配置。
	// 此时 {0,1} 是旧配置的过半，{2,3,4} 是新配置的过半 —— 两组不相交。
	direct := func(seed int64) bool {
		c := mk(seed, nil)
		for i := 0; i < *nF; i++ {
			c.SetConfig(i, []int{0, 1, 2}) // 起始配置：只有 3 个节点
		}
		c.net.Kill(3)
		c.net.Kill(4)
		run(c, 2500)
		c.net.Revive(3)
		c.net.Revive(4)
		// ★ 只有一部分节点切到了新配置
		for _, i := range []int{2, 3, 4} {
			c.SetConfig(i, []int{0, 1, 2, 3, 4})
		}
		// 制造分区，让两组各自选举
		c.net.Partition([]int{0, 0, 1, 1, 1})
		run(c, 9000)
		return len(c.Violations) > 0 || c.twoLeadersEver
	}

	// 单节点变更 {0,1,2} → {0,1,2,3}：同样只有一部分节点先切，
	// 而且用完全相同的分区方式去"制造"裂脑 —— 结果会告诉你它造不出来。
	single := func(seed int64) bool {
		c := mk(seed, nil)
		for i := 0; i < *nF; i++ {
			c.SetConfig(i, []int{0, 1, 2})
		}
		c.net.Kill(3)
		c.net.Kill(4) // N5 不属于新旧任何一个配置，全程离线
		run(c, 2500)
		c.net.Revive(3)
		for _, i := range []int{2, 3} { // 只有 N3、N4 切到了新配置
			c.SetConfig(i, []int{0, 1, 2, 3})
		}
		// 同样切成 {N1,N2} | {N3,N4}：前者按旧配置已过半(2/3)，
		// 后者按新配置只有 2 个可达却需要 3 个 ⇒ 选不出来。
		c.net.Partition([]int{0, 0, 1, 1, 1})
		run(c, 9000)
		return len(c.Violations) > 0 || c.twoLeadersEver
	}

	w := []int{34, 18, 18}
	al := "LRR"
	tui.TableHeadA([]string{"变更方式", "出现两个 Leader", "结论"}, w, al)
	cnt := func(f func(int64) bool) int {
		k := 0
		for seed := int64(1); seed <= int64(*trialsF); seed++ {
			if f(*seedF + seed*13) {
				k++
			}
		}
		return k
	}
	d := cnt(direct)
	sg := cnt(single)
	tui.TableRowA([]string{"① 直接跳 {A,B,C} → {A..E}", fmt.Sprintf("%d/%d", d, *trialsF),
		map[bool]string{true: "✗ 会裂脑", false: "✓"}[d > 0]}, w, al)
	tui.TableRowA([]string{"② 单节点 {A,B,C} → {A,B,C,D}", fmt.Sprintf("%d/%d", sg, *trialsF),
		map[bool]string{true: "✗ 会裂脑", false: "✓ 一次都造不出来"}[sg > 0]}, w, al)
	fmt.Println()
	tui.Row("① 的两个过半集合", "旧 {A,B}（2/3 够）与 新 {C,D,E}（3/5 够）—— 不相交")
	tui.Row("① 的算式", "maj_old + maj_new = 2+3 = 5，|并集| = 5，5 > 5 ✗ 不成立")
	tui.Row("② 的算式", "maj_old + maj_new = 2+3 = 5，|并集| = 4，5 > 4 ✓ 成立 ⇒ 必相交")
	fmt.Println(`
    ▸ ① 里 {A,B} 用旧配置算已经过半（2/3），{C,D,E} 用新配置算也已经过半（3/5），
      两组没有任何共同节点 ⇒ 可以在同一任期各选一个 Leader ⇒ 脑裂。
    ▸ ② 里要凑出两个不相交的过半集合需要 5 个不同节点，而并集只有 4 个 —— 鸽笼原理，
      凑不出来。所以「一次只动一个节点」不是谨慎习惯，是可以证明的充分条件。
    ▸ 一般式：加一个节点时 maj_old + maj_new = n+2 > n+1 = |并集|，恒成立；
      减一个时和为 n+1 > n，也恒成立。跳两格时等号刚好不成立，缝就开了。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3C-4：三种读，同一个僵尸 Leader
// ═══════════════════════════════════════════════════════════════════════
func lab3c4() {
	tui.HeadN("3C-4", "三种读的对照", "制造僵尸 Leader，看哪种读会静默返回陈旧数据")

	type r struct{ stale, errs, good int }
	res := map[ReadMode]*r{LocalRead: {}, ReadIndexRead: {}, LeaseRead: {}}

	for seed := int64(1); seed <= int64(*trialsF); seed++ {
		c := mk(*seedF+seed*7, nil)
		run(c, 1500)
		l := c.Leader()
		if l == nil {
			continue
		}
		c.Submit("SET x=1")
		run(c, 600)

		// 把 Leader 隔离到少数派：它 + 一个邻居
		other := (l.id + 1) % *nF
		grp := make([]int, *nF)
		grp[l.id], grp[other] = 1, 1
		c.net.Partition(grp)
		run(c, 6000) // 多数派选出新 Leader

		// 客户端向【新】Leader 写 x=99
		var newL *Node
		for _, n := range c.nodes {
			if n.state == Leader && grp[n.id] == 0 {
				newL = n
			}
		}
		if newL == nil {
			continue
		}
		newL.log = append(newL.log, Entry{Index: newL.last().Index + 1, Term: newL.term, Cmd: "SET x=99"})
		newL.matchIndex[newL.id] = newL.last().Index
		run(c, 2500)
		if newL.kv["x"] != "99" {
			continue // 这一轮没写成功，跳过
		}

		// 现在从僵尸 Leader 上用三种方式读
		zombie := c.nodes[l.id]
		if zombie.state != Leader {
			continue // 这一轮它已经退位了，不是僵尸
		}
		for _, m := range []ReadMode{LocalRead, ReadIndexRead, LeaseRead} {
			v, ok := c.Read(zombie.id, "x", m)
			switch {
			case !ok:
				res[m].errs++
			case v != "99":
				res[m].stale++
			default:
				res[m].good++
			}
		}
	}

	w := []int{20, 20, 20, 20}
	al := "LRRR"
	tui.TableHeadA([]string{"读的实现", "★ 静默返回陈旧值", "正确报错", "返回正确值"}, w, al)
	for _, m := range []ReadMode{LocalRead, ReadIndexRead, LeaseRead} {
		tui.TableRowA([]string{m.String(), fmt.Sprint(res[m].stale),
			fmt.Sprint(res[m].errs), fmt.Sprint(res[m].good)}, w, al)
	}
	fmt.Println(`
    ▸ 本地读会静默返回旧值 —— 不报错、不留日志。客户端刚从新 Leader 拿到"写入成功"，
      转头从僵尸 Leader 读到旧值，线性一致性当场破裂，而监控上什么都看不出来。
    ▸ ReadIndex 在读之前先确认过半节点仍认自己是主，僵尸 Leader 凑不齐 ⇒ 报错。
      【给你错误，而不是给你错的数据】—— 这个区别在排障时价值巨大。
    ▸ Lease Read 的租约在收不到过半心跳响应后就续不上了，过期后同样拒绝本地读。
      它比 ReadIndex 快（省掉一次心跳往返），但安全性建立在【时钟漂移有界】上：
      如果这台机器的时钟突然变慢，它会以为租约还没到期而继续本地读。
      Part 1 那个坑，在这里变成了一致性问题。
    ▸ 所以"Raft 集群是线性一致的"准确的说法是：写入是线性一致的；
      读是否线性一致，取决于你选了上面哪一种。etcd 默认 ReadIndex，TiKV 默认 Lease Read。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  三个思考题（答案在课件 §3.18 / §3.20 / §3.21）：

    1. 三样状态里，丢掉哪一样造成的违规最多？为什么是它？
    2. Lab 3C-3 的 ① 里两个 Leader 分别靠哪些节点的票当选？
       把那两组列出来，验证它们确实不相交。
    3. Lab 3C-4 里 Lease Read 一次陈旧读都没有 —— 这是否意味着它和 ReadIndex 一样安全？
       先想清楚它依赖什么假设。

  Part 3 到此结束。下一站：Part 4 · 分布式事务
  （ACID、隔离级别、MVCC、2PC/XA、TCC、Saga、事务消息、Percolator、Calvin、Spanner）。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
