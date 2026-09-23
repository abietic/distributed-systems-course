// Lab 3-D · Raft 附录：索引、读、配置与日志
//
// 运行：  cd go/lab03d && go run .
// 调参：  go run . -applylag 8 -reads 300 -seeds 60 -writers 4
//
// 配套课件：courseware/ch03d-faq.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"dsc/internal/tui"
)

var (
	applylagF = flag.Int("applylag", 4, "3D-1：apply 线程最多滞后多少条")
	readsF    = flag.Int("reads", 500, "3D-1：每种配置读多少次")
	seedsF    = flag.Int("seeds", 40, "3D-2 / 3D-3：跑多少个种子")
	writersF  = flag.Int("writers", 2, "3D-3：并发写的客户端数")
	nF        = flag.Int("n", 5, "3D-3：副本数 N")
	wF        = flag.Int("w", 3, "3D-3：写 quorum W")
	rF        = flag.Int("r", 3, "3D-3：读 quorum R")
	lossF     = flag.Float64("loss", 0.22, "3D-3：丢包率")
	seedF     = flag.Int64("seed", 42, "随机种子基准")
	entriesF  = flag.Int("entries", 3, "3D-2：旧 Leader 崩溃前复制了几条")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：N=%d W=%d R=%d ｜ applylag=%d ｜ 种子基准 %d\n",
		*nF, *wF, *rF, *applylagF, *seedF)
	lab3d1()
	lab3d2()
	lab3d3()
	lab3d4()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3D-1：ReadIndex 的第三步
// ═══════════════════════════════════════════════════════════════════════
func lab3d1() {
	tui.HeadN("3D-1", "ReadIndex 的第三步", "前两步全部通过，少了第三步会读到什么")

	fmt.Println(`
    模型：一个没有分区的 Leader。它一直合法，心跳永远能收到过半响应
    ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。`)

	w := []int{14, 3, 14, 16, 3, 14, 16}
	al := "LLRRLRR"
	tui.TableHeadA([]string{"apply 滞后", "", "跳过第3步", "└ 陈旧读", "", "执行第3步", "└ 平均等待"}, w, al)

	for lag := 0; lag <= *applylagF*2; lag++ {
		if lag > 0 && lag != *applylagF && lag%2 == 1 && lag != 1 {
			continue
		}
		staleNo, staleYes, waited := 0, 0, 0
		rnd := NewRand(*seedF*1000 + int64(lag)*17 + 3)
		for i := 0; i < *readsF; i++ {
			commit := 10 + i
			// apply 线程此刻落后多少条，随机落在 [0, lag]
			l := &LeaderState{commitIndex: commit, lastApplied: commit - rnd.Intn(lag+1)}
			snapshot := *l
			if r := l.DoRead(false); r.Stale {
				staleNo++
			}
			l2 := snapshot
			r2 := l2.DoRead(true)
			if r2.Stale {
				staleYes++
			}
			waited += r2.Waited
		}
		mark := ""
		if lag == *applylagF {
			mark = "◀"
		}
		tui.TableRowA([]string{
			fmt.Sprintf("%d 条以内", lag), mark,
			fmt.Sprintf("%d/%d", staleNo, *readsF),
			pct(staleNo, *readsF),
			"",
			fmt.Sprintf("%d/%d", staleYes, *readsF),
			fmt.Sprintf("%.2f 条", float64(waited)/float64(*readsF)),
		}, w, al)
	}

	fmt.Println(`
    结论
      · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
      · 执行第 3 步的陈旧读恒为 0，代价是平均多等几条 apply。
      · 生产上这条曲线对应的现象就是：写接口返回成功，立刻查却查不到。
        滞后被撑大的典型原因是状态机执行慢（大事务 / 写放大 / compaction 抢 IO）。
      · 再说一遍：这一整张表里，ReadIndex 的第 1、2 步【全部通过】。
        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3D-2：no-op —— 新 Leader 的 commitIndex 从哪来
// ═══════════════════════════════════════════════════════════════════════
func lab3d2() {
	tui.HeadN("3D-2", "no-op 条目的必要性", "新 Leader 不发空日志，ReadIndex 三步走完照样读到旧值")

	fmt.Printf(`
    场景（每个种子随机决定两个分支）：
      任期 2：Leader 把 %d 条写复制到了过半节点，然后崩溃。
              分支 ①（约一半概率）它崩之前【已经推进 commitIndex 并回复客户端成功】；
              分支 ②                它还没来得及推进 commitIndex。
              两种情况下它都没来得及广播 leaderCommit ——
              「已提交」这个事实随它一起消失了。
      任期 3：某个持有这些日志的节点当选，它的 commitIndex = 0。
              受 Figure 8 限制，它【不能】直接提交任期 2 的条目。
      然后客户端用完整的 ReadIndex（三步一步不少）来读。
`, *entriesF)

	w := []int{22, 18, 16, 18, 22}
	al := "LRRRR"
	tui.TableHeadA([]string{"新 Leader 上任后", "读到旧值(违规)", "返回最新值", "排队后返回最新", "旧值但未ack(允许)"}, w, al)

	type res struct{ stale, fresh, queued, allowed int }
	run := func(noop bool) res {
		var r res
		rnd := NewRand(*seedF * 31)
		for s := 0; s < *seedsF; s++ {
			acked := rnd.Float() < 0.5  // 分支 ① / ②
			early := rnd.Float() < 0.35 // 读请求到得早不早
			switch RunNoop(*entriesF, acked, noop, early) {
			case OutcomeStale:
				r.stale++
			case OutcomeFresh:
				r.fresh++
			case OutcomeQueued:
				r.queued++
			case OutcomeAllowed:
				r.allowed++
			}
		}
		return r
	}
	a := run(false)
	b := run(true)
	tui.TableRowA([]string{"不发 no-op", fmt.Sprintf("%d", a.stale), fmt.Sprintf("%d", a.fresh),
		fmt.Sprintf("%d", a.queued), fmt.Sprintf("%d", a.allowed)}, w, al)
	tui.TableRowA([]string{"发 no-op（正确）", fmt.Sprintf("%d", b.stale), fmt.Sprintf("%d", b.fresh),
		fmt.Sprintf("%d", b.queued), fmt.Sprintf("%d", b.allowed)}, w, al)

	fmt.Printf(`
    结论
      · 不发 no-op：%d/%d 个种子读到了旧值，而客户端【收到过写入成功】——
        这是货真价实的线性一致性破裂，且 ReadIndex 三步一步没少、全部通过。
        问题出在第 1 步拿到的 commitIndex 本身就是错的。
      · 发 no-op：违规 0 次。no-op 是【当前任期】的条目，可以直接提交；
        提交是前缀性的 ⇒ 任期 2 那几条被顺带提交 ⇒ commitIndex 到位。
      · 那 %d 次「排队」也很重要：no-op 还没提交时，etcd 把读请求挂起
        （pendingReadIndexMessages），等本任期第一条提交后再处理——那时 commitIndex 已经到位。
        也有实现选择直接拒绝、让客户端重试。两种都安全，
        绝不能做的是拿一个还没到位的 commitIndex 去凑合。
      · 「旧值但未 ack」的那些不算违规：那次写从未收到成功响应，
        它的结果本来就是 Part 0 说的「第三态」——客户端必须自己去查证。
`, a.stale, *seedsF, b.queued)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3D-3：Quorum 读写 vs Raft 日志复制
// ═══════════════════════════════════════════════════════════════════════
func lab3d3() {
	tui.HeadN("3D-3", "Quorum 读写 ≠ Raft 日志复制", "同一组并发写灌进两个世界，数副本间的分歧")

	fmt.Printf(`
    %d 个客户端【同时】写同一个 key，N=%d，丢包率 %.0f%%。
      ① Dynamo 式：每个写发往全部副本、收到 W 个 ack 即成功，副本不带版本信息。
      ② Raft    ：全部请求先到 Leader，Leader 按到达先后分配 index 再复制。
`, *writersF, *nF, *lossF*100)

	w := []int{8, 2, 16, 16, 18, 3, 14, 14}
	al := "LLRRRLRR"
	tui.TableHeadA([]string{"写 W", "", "Dynamo 写成功", "└ 平均分歧", "└ 出现分歧的种子",
		"", "Raft 写成功", "└ 平均分歧"}, w, al)

	for ww := 1; ww <= *nF; ww++ {
		sumD, seedsWithD, okD := 0, 0, 0
		sumR, seedsWithR, okR := 0, 0, 0
		for s := 1; s <= *seedsF; s++ {
			rnd := NewRand(*seedF + int64(s)*2654435761)
			d := RunDynamo(*nF, ww, *rF, *writersF, *lossF, rnd)
			sumD += d.Distinct
			okD += d.WriteOK
			if d.Distinct > 1 {
				seedsWithD++
			}
			rnd2 := NewRand(*seedF + int64(s)*2654435761)
			r := RunRaft(*nF, *writersF, rnd2)
			sumR += r.Distinct
			okR += r.WriteOK
			if r.Distinct > 1 {
				seedsWithR++
			}
		}
		mark := ""
		if ww == *wF {
			mark = " ◀"
		}
		tui.TableRowA([]string{
			fmt.Sprintf("%d%s", ww, mark), "",
			fmt.Sprintf("%.2f/%d", float64(okD)/float64(*seedsF), *writersF),
			fmt.Sprintf("%.2f 种", float64(sumD)/float64(*seedsF)),
			fmt.Sprintf("%d/%d", seedsWithD, *seedsF), "",
			fmt.Sprintf("%.2f/%d", float64(okR)/float64(*seedsF), *writersF),
			fmt.Sprintf("%.2f 种", float64(sumR)/float64(*seedsF)),
		}, w, al)
	}

	// 展示某一个种子下的细节
	rnd := NewRand(*seedF + 7*2654435761)
	d := RunDynamo(*nF, *wF, *rF, *writersF, *lossF, rnd)
	rnd2 := NewRand(*seedF + 7*2654435761)
	r := RunRaft(*nF, *writersF, rnd2)
	fmt.Printf("\n    随便挑一个种子看细节（W=%d）：\n", *wF)
	fmt.Printf("      Dynamo 各副本最终保存： %s   ⇒ %d 种不同的值，读 R=%d 拿回 %d 个版本\n",
		vals(d.Replicas), d.Distinct, *rF, d.ReadVers)
	fmt.Printf("      Raft   各副本最终保存： %s   ⇒ %d 种，读 Leader 一个拿回 %d 个版本\n",
		vals(r.Replicas), r.Distinct, r.ReadVers)

	fmt.Println(`
    结论
      · 看第一列和第三列的对比：W 从 1 调到 N，【写成功数在掉】（要求的 ack 更多了），
        但【分歧数一点没变】。W 决定的只是写端何时收到 ack，
        和副本之间是否一致毫无关系——把 W 调到 N 也救不了。
        分歧的成因是【各副本看到的到达顺序不同】，而副本没有版本信息去裁决。
      · Raft 的分歧恒为 1。不是概率上恰好，是结构上不可能：
        顺序在请求到达 Leader 的那一刻就定死了，复制的是【同一个 index 上的同一条 entry】。
      · 所以「过半」在两边的含义完全不同：
          Dynamo：读集合 ∩ 写集合 ≠ ∅  ⇒ 至少碰到一个持有最新值的副本（但不知道哪个是最新）
          Raft  ：任意两次成功操作的参与集合必相交 ⇒ 新 Leader 必持有全部已提交日志
        底层数学是同一条，用法完全相反。
      · 顺便：Raft 里没有「quorum 读」。ReadIndex 那轮心跳拿回的不是数据，
        是【我还是不是 Leader】这一个 bit。`)
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 3D-4：冷启动 —— 配置写错会长出几个集群
// ═══════════════════════════════════════════════════════════════════════
func lab3d4() {
	tui.HeadN("3D-4", "冷启动沙盒", "初始配置只能从外部给——给错了会长出第二个集群，或者更糟")

	type scen struct {
		name  string
		nodes []NodeCfg
		steps []Step
		note  string
	}
	full := []string{"n1", "n2", "n3"}
	start := func(ids ...string) []Step {
		out := []Step{}
		for _, id := range ids {
			out = append(out, Step{Kind: "start", Node: id})
		}
		return out
	}
	scens := []scen{
		{"A · 正确的静态引导", []NodeCfg{
			{ID: "n1", Cluster: full, State: "new", Token: "T1"},
			{ID: "n2", Cluster: full, State: "new", Token: "T1"},
			{ID: "n3", Cluster: full, State: "new", Token: "T1"},
		}, start("n1", "n2", "n3"),
			"期望结果：1 个集群、1 个 Leader、容错 1 台。"},

		{"B · 一个节点的列表写错", []NodeCfg{
			{ID: "n1", Cluster: full, State: "new", Token: "T1"},
			{ID: "n2", Cluster: full, State: "new", Token: "T1"},
			{ID: "n3", Cluster: []string{"n3"}, State: "new", Token: "T1"},
		}, start("n1", "n2", "n3"),
			"最危险的一种：两个集群、两个 Leader，【都能写】，数据分叉且永远无法自动合并。\n" +
				"      n3 的成员列表不同 ⇒ 集群 ID 不同 ⇒ 两边互不承认。从 n3 的视角看，它的集群完美无瑕。"},

		{"C · token 不同", []NodeCfg{
			{ID: "n1", Cluster: full, State: "new", Token: "T1"},
			{ID: "n2", Cluster: full, State: "new", Token: "T1"},
			{ID: "n3", Cluster: full, State: "new", Token: "T2"},
		}, start("n1", "n2", "n3"),
			"比 B 好，因为它失败得很响：token 是成员 ID 的一部分 ⇒ n3 眼里的集群 ID 和 n1/n2 的不同 ⇒ 被拒之门外；\n" +
				"      它自己认为集群有 3 个成员，凑不齐过半 ⇒ 卡在启动中。n1/n2 能工作，但容错能力是 0。"},

		{"D1 · 清空两台后用 state=new 重启（n1 在线）", []NodeCfg{
			{ID: "n1", Cluster: full, State: "existing", Token: "T1", HasData: true, Term: 12},
			{ID: "n2", Cluster: full, State: "new", Token: "T1", Wiped: true},
			{ID: "n3", Cluster: full, State: "new", Token: "T1", Wiped: true},
		}, start("n1", "n2", "n3"),
			"同 token、同成员列表 ⇒ 集群 ID【完全相同】，n2/n3 不是在建「另一个集群」，而是在冒充已经存在的成员。\n" +
				"      etcd 的 isMemberBootstrapped 检查拦住了它们：n1 记得它们发布过客户端地址 ⇒ 拒绝启动。\n" +
				"      结果是停写，但数据还在。过半已经丢了，member remove 也提交不了，\n" +
				"      正确的恢复路径是在 n1 上用 --force-new-cluster 重建单成员集群，再逐个 member add。"},

		{"D2 · 同上，但 n1 当时不在线", []NodeCfg{
			{ID: "n1", Cluster: full, State: "existing", Token: "T1", HasData: true, Term: 12},
			{ID: "n2", Cluster: full, State: "new", Token: "T1", Wiped: true},
			{ID: "n3", Cluster: full, State: "new", Token: "T1", Wiped: true},
		}, []Step{{Kind: "start", Node: "n2"}, {Kind: "start", Node: "n3"},
			{Kind: "write", N: 5}, {Kind: "start", Node: "n1"}},
			"最阴险的一种。检查问不到任何知情者就放行了（源码里联系不上就返回 false）。\n" +
				"      n2/n3 用同一个集群 ID 引导，任期从 0 重来，在任期 2 里确认了 5 条写。n1 回来时任期 12：\n" +
				"      现任 Leader 收到它带着任期 12 的回复就下台，n1 的日志更新 ⇒ 当选 ⇒ 那 5 条写被覆盖。\n" +
				"      根因回到 3-C：清空 data dir = 同时丢掉 currentTerm、votedFor、log。任期倒退，\n" +
				"      才会出现「任期 2 的 Leader」这种本不该存在的东西，它提交的内容不受 Raft 保护。\n" +
				"      如果 n1 永远不回来：集群健康、能读能写，半年的数据没了。"},

		{"E · 单节点引导 + member add", []NodeCfg{
			{ID: "n1", Cluster: []string{"n1"}, State: "new", Token: "T1"},
			{ID: "n2", Cluster: []string{"n1", "n2"}, State: "existing", Token: "T1"},
			{ID: "n3", Cluster: full, State: "existing", Token: "T1"},
		}, []Step{{Kind: "start", Node: "n1"},
			{Kind: "add", Node: "n2"}, {Kind: "start", Node: "n2"},
			{Kind: "add", Node: "n3"}, {Kind: "start", Node: "n3"}},
			"同样正确，而且不需要提前知道最终拓扑。每个 existing 节点的 --initial-cluster\n" +
				"      = 当时的成员 ∪ 自己（正是 member add 打印出来的内容），它不组建任何东西，只是加入。\n" +
				"      n1 的 --initial-cluster 一直是「只有自己」，这没关系：集群跑起来后成员信息在 data dir 里。"},

		{"F · E 的最后一步，n3 的列表漏了 n2", []NodeCfg{
			{ID: "n1", Cluster: []string{"n1"}, State: "new", Token: "T1"},
			{ID: "n2", Cluster: []string{"n1", "n2"}, State: "existing", Token: "T1"},
			{ID: "n3", Cluster: []string{"n1", "n3"}, State: "existing", Token: "T1"},
			{ID: "n4", Cluster: []string{"n1", "n2", "n3", "n4"}, State: "existing", Token: "T1"},
		}, []Step{{Kind: "start", Node: "n1"},
			{Kind: "add", Node: "n2"}, {Kind: "start", Node: "n2"},
			{Kind: "add", Node: "n3"}, {Kind: "start", Node: "n3"},
			{Kind: "add", Node: "n4"}},
			"n3 的参数和集群当前成员对不上 ⇒ 被拒绝启动。但 member add n3 已经提交了：\n" +
				"      集群现在是 3 个成员、2 个在线——还能服务，容错能力却是 0。\n" +
				"      这时想再加 n4，etcd 会拒绝（加完后已启动的成员凑不齐过半）——这就是「一次只加一个、\n" +
				"      确认正常了再加下一个」的原因，也是先用 --learner 加入、追平后再 promote 的原因。"},
	}

	w := []int{46, 8, 8, 14, 10, 30}
	al := "LRRRRL"
	tui.TableHeadA([]string{"场景", "集群数", "Leader", "在线/成员", "被拒启动", "  结果"}, w, al)
	type outcome struct {
		r       BootResult
		leaders int
		verdict string
	}
	results := make([]outcome, len(scens))
	for i, sc := range scens {
		r := Bootstrap(sc.nodes, sc.steps)
		leaders := 0
		online := []string{}
		for _, c := range r.Clusters {
			if c.Leader() {
				leaders++
			}
			online = append(online, fmt.Sprintf("%d/%d", c.started(), len(c.Members)))
		}
		degraded := false
		for _, c := range r.Clusters {
			if c.Leader() && c.started() < len(c.Members) {
				degraded = true
			}
		}
		v := "正常"
		switch {
		case leaders > 1:
			v = "脑裂：两个集群都能写"
		case r.Lost > 0:
			v = fmt.Sprintf("%d 条已确认的写被覆盖", r.Lost)
		case leaders == 0:
			v = "停写（但数据还在）"
		case len(r.Rejected) > 0 || degraded:
			v = "降级：容错能力归零"
		}
		refused := "—"
		if len(r.Refused) > 0 {
			refused = join(r.Refused, ",")
		}
		results[i] = outcome{r, leaders, v}
		tui.TableRowA([]string{sc.name, fmt.Sprintf("%d", len(r.Clusters)), fmt.Sprintf("%d", leaders),
			join(online, " + "), refused, "  " + v}, w, al)
	}

	fmt.Println()
	for i, sc := range scens {
		fmt.Printf("    %s\n", sc.name)
		for _, line := range results[i].r.Log {
			fmt.Printf("      %s\n", line)
		}
		fmt.Printf("      ── %s\n\n", strings.ReplaceAll(sc.note, "\n      ", "\n         "))
	}

	fmt.Println(`    结论
      · 「有几台机器」和「怎么访问到对方」全部来自那份 --initial-cluster ——
        Raft 自己不做服务发现，DNS SRV / K8s Service 只是帮你【生成】这份列表。
      · 集群 ID 是由 token 和成员列表【确定性】算出来的：同 token 同列表 ⇒ 同一个集群。
        所以清空 data dir 后用 state=new 重启，不是「建了个新集群」，而是「冒充旧成员」。
      · 只有 state=new 的节点会按自己的参数引导集群；existing 节点只加入、从不组建，
        它的 --initial-cluster 必须和集群当前成员一致，否则拒绝启动。
      · --initial-* 只在 data dir 为空的首次启动生效。集群跑起来后成员信息就在本地状态里了，
        【改配置文件不会改变集群成员】——只能走 member add / remove / update。
        成员数据丢了的节点，正确做法是 member remove 再 member add 一个新成员（新 ID 带时间戳，不会撞）。`)
}

// ═══════════════════════════════════════════════════════════════════════
func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  Part 3 到此真正结束

  3-A 选谁当 Leader，3-B Leader 怎么把日志复制对，
  3-C 怎么活到生产，3-D 算法与系统之间的接缝。

  贯穿四章的还是那一条：【两个过半集合必然相交】。
  它在 Raft 里一共出现了四次——选举安全性、Leader 完整性、
  quorum 读写、成员变更——你已经全部见过了。

  下一站 Part 4 · 事务。先补单机的功课（ACID、隔离级别、
  四类并发异常、MVCC 与快照隔离），那正是 3D 里说的「第二层版本」。
  然后才进分布式：2PC/XA、TCC、Saga、事务消息、Percolator、Calvin、Spanner。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func pct(a, b int) string {
	if b == 0 {
		return "—"
	}
	return fmt.Sprintf("%.0f%%", float64(a)/float64(b)*100)
}

func vals(rs []int) string {
	names := []string{"A", "B", "C", "D"}
	out := ""
	for i, v := range rs {
		if i > 0 {
			out += " "
		}
		if v < 0 {
			out += "∅"
		} else {
			out += names[v%4]
		}
	}
	return out
}
