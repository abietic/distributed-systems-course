## 👤 你 · 2026-09-09 20:50

1. 因为新的leader会有新的日志append也会有新的commit，这导致新commit的那段日志僵尸leader拿不到，如果这部分日志涉及到历史commit日志相关内容的修改就会出问题。
2. 当前node id、term、commitIdex和已经commit的日志，没落盘会导致选举时的term可能和已经提交的日志不一致或更小
3. 可能导致裂脑，配置切换时间不同导致重新选举时用了旧的集群数开始发送日志

我还有额外的问题：

1. 这样是怎么解决网络分区后少数区leader，term大量增长后分区恢复在选举时成为主，是否会导致选主优先选老的，还是会通过commitIndex做判断？因为理论上少数分区下的leader也会append只是一直没commit，类似的则是figure8.
2. 还有figure8的图示在term5的展示时好像有错误还是term3
3. 还有3-a的网络分区后一段时间分区恢复的重新选主好像也存在bug对应我的第一个问题。

▸ **ToolSearch**

````
{"max_results": 3, "query": "select:WebFetch,WebSearch"}
````

<details><summary>↳ ToolSearch 结果</summary>

````
[tool_reference]
[tool_reference]
````

</details>

▸ **WebFetch**

````
{"prompt": "Find Figure 8 (the time sequence showing why a leader cannot determine commitment using log entries from older terms). For each of the five panels (a) through (e), list exactly what log entries each of S1..S5 has, giving the term number of the entry at each log index. Quote the figure caption verbatim. In particular: in panel (d), what entries does S5 have at index 2 and index 3, and what are their term numbers?", "url": "https://raft.github.io/raft.pdf"}
````

<details><summary>↳ WebFetch 结果</summary>

````
# Figure 8 Analysis

**Figure Caption (verbatim):**
"A time sequence showing why a leader cannot determine commitment using log entries from older terms. In (a) S1 is leader and partially replicates the log entry at index 2. In (b) S1 crashes; S5 is elected leader for term 3 with votes from S3, S4, and itself, and accepts a different entry at log index 2. In (c) S5 crashes; S1 restarts, is elected leader, and continues replication. At this point, the log entry from term 2 has been replicated on a majority of the servers, but it is not committed. If S1 crashes as in (d), S5 could be elected leader (with votes from S2, S3, and S4) and overwrite the entry with its own entry from term 3. However, if S1 replicates an entry from its current term on a majority of the servers before crashing, as in (e), then this entry is committed (S5 cannot win an election). At this point all preceding entries in the log are committed as well."

**Log State Summary:**

- **(a):** S1: [term 2]; S2-S5: empty
- **(b):** S1: [term 2]; S5: [term 3]; S2-S4: empty
- **(c):** S1: [term 2]; S5: [term 3]; S2-S4: empty
- **(d):** S1: [term 2]; S5: index 2 has term 3, index 3 empty; S2-S4: empty
- **(e):** S1: [term 2, term 4]; S5: [term 3, term 4]; S2-S4: [term 4]
````

</details>

▸ **Bash** Run diagnostic on high-term rejoin behavior

````
mkdir -p 〔路径00197〕 && cd 〔路径00104〕 && cat > diag_test.go <<'EOF'
package main

import (
	"fmt"
	"testing"
)

// 诊断：分区恢复后，高任期的少数派节点能不能赢得选举？
func TestHighTermRejoin(t *testing.T) {
	for _, leaderInMinority := range []bool{false, true} {
		net := NewNetwork(5, 20, 10, 0, 7)
		c := New(5, net, 400, 300, 7)
		for i := int64(0); i < 300; i++ {
			c.Step(5)
		}
		l := c.Leader()
		if l == nil {
			t.Fatal("no leader")
		}
		// 提交一批日志
		for i := 0; i < 30; i++ {
			c.Submit(fmt.Sprintf("pre%d", i))
			for k := 0; k < 6; k++ {
				c.Step(5)
			}
		}
		for k := 0; k < 400; k++ {
			c.Step(5)
		}
		committedBefore := c.MaxCommit()

		// 构造分区：让 leader 落在多数派或少数派
		grp := make([]int, 5)
		var minority []int
		if leaderInMi
…（截断，原长 3208 字符）
````

<details><summary>↳ Bash 结果</summary>

````
=== RUN   TestHighTermRejoin

=== leaderInMinority=false ===
分区期间：少数派 [2 3] 的 term=[34 34] ｜ 多数派 leader=N1 term=1 ｜ 已提交=90
  少数派 N2: state=Candidate term=34 lastLog=(term=1,idx=30) commit=30
  少数派 N3: state=Candidate term=34 lastLog=(term=1,idx=30) commit=30
恢复后：leader=N4 term=39 ｜ 最高已提交=90（分区前 30）
  恢复后的 leader 来自：多数派（新日志）
  安全性违反：0 条 []
  分区前已提交日志被改动的副本数：0

=== leaderInMinority=true ===
分区期间：少数派 [1 2] 的 term=[1 1] ｜ 多数派 leader=N1 term=1 ｜ 已提交=30
  少数派 N1: state=Leader term=1 lastLog=(term=1,idx=90) commit=30
  少数派 N2: state=Follower term=1 lastLog=(term=1,idx=90) commit=30
恢复后：leader=N3 term=2 ｜ 最高已提交=30（分区前 30）
  恢复后的 leader 来自：多数派（新日志）
  安全性违反：0 条 []
  分区前已提交日志被改动的副本数：0
--- PASS: TestHighTermRejoin (0.04s)
PASS
ok  	dsc/lab03b	0.039s
````

</details>

▸ **Bash** Test whether stale minority node can win in 3-A

````
cd 〔路径00100〕 && cat > diag_test.go <<'EOF'
package main

import (
	"fmt"
	"testing"
)

// 诊断：3-A 的模型没有日志，所以投票时缺少「日志至少一样新」这一条。
// 分区恢复后，一个日志完全落后（这里是"没有日志"）但任期很高的少数派节点能不能当选？
func TestMinorityWinsAfterHeal(t *testing.T) {
	wins := map[string]int{}
	for seed := int64(1); seed <= 40; seed++ {
		net := NewNetwork(5, 30, 15, 0, seed)
		c := New(5, net, 400, 300, seed, false)
		step := func(ms int64) { for t := int64(0); t < ms; t += 5 { c.Step(5) } }
		step(2000)
		grp := []int{0, 0, 0, 1, 1} // 多数派 {N1,N2,N3}，少数派 {N4,N5}
		net.Partition(grp)
		step(12000) // 少数派任期疯涨
		minT := c.nodes[3].term
		if c.nodes[4].term > minT { minT = c.nodes[4].term }
		majT := 0
		for i := 0; i < 3; i++ { if c.nodes[i].term > majT { majT = c.nodes[i].term } }
		net.Heal()
		step(8000)
		l := c.Leader
…（截断，原长 1216 字符）
````

<details><summary>↳ Bash 结果</summary>

````
=== RUN   TestMinorityWinsAfterHeal
  seed=2：分区期间 少数派term=24 多数派term=1 ⇒ 恢复后 N5（少数派）当选，term=25
  seed=3：分区期间 少数派term=24 多数派term=1 ⇒ 恢复后 N5（少数派）当选，term=25

【Lab 3-A，无日志检查】40 次试验：map[★ 少数派节点当选:25 多数派节点当选:15]
--- PASS: TestMinorityWinsAfterHeal (0.04s)
PASS
ok  	dsc/lab03a	0.045s
````

</details>

▸ **Bash** Fix Figure 8 rendering and add high-term rejoin section

````
cd 〔路径00054〕 && rm -f go/lab03a/diag_test.go go/lab03b/diag_test.go && python3 - <<'PY'
p='ch03b/ch03b.html'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)

# ── 修正 Figure 8 (d)：S5 的日志应当是 [1,3]，不是 [1,3,3] ──
rep(''' {t:"(d) 任期 5：S1 又崩了，S5 当选并覆盖 index2",
  leader:4, term:5, dead:[0],
  logs:[[1,2,4],[1,3],[1,3],[1,3],[1,3,3]],''',
''' {t:"(d) 任期 5：S1 又崩了，S5 当选并覆盖 index2",
  leader:4, term:5, dead:[0],
  logs:[[1,2,4],[1,3],[1,3],[1,3],[1,3]],''')
rep('''       "于是 S5 把 index2 上的 term2 <strong>全部覆盖成了 term3</strong>。",''',
'''       "于是 S5 把 index2 上的 term2 <strong>全部覆盖成了 term3</strong>。<br>"+
       "<span style='color:var(--ink-3);font-size:13px'>注意 S5 的日志仍然只有两条 —— 论文原文说的是"+
       "「overwrite the entry with its own entry from t
…（截断，原长 5558 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch03b patched
````

</details>

▸ **Bash** Add simplification notes to Part 3-A materials

````
python3 - <<'PY'
p='ch03a/ch03a.html'; s=open(p,encoding='utf-8').read()
old='''  <div class="lab">
    <div class="lab-h"><div class="ico">2</div><div class="t">实验 2 · Raft 选举模拟器</div>'''
new='''  <div class="note warn">
    <div class="note-h"><span class="dot" style="background:var(--warn)"></span>本节的模拟器与 Lab 3-A 代码省略了条件 ③</div>
    <p style="margin-bottom:0">因为这一节的模型里<strong>还没有日志</strong>，所以"日志至少一样新"这条投票条件在实验 2 和 Lab 3-A 的代码里<u>没有实现</u>。<br>
    这个简化有一个可观测的后果：制造 3|2 分区、让少数派的任期涨上去、再恢复网络，<strong>陈旧的少数派节点会有相当高的概率当选</strong>——在真实 Raft 里这不可能发生，因为它的日志落后，拿不到票。<br>
    <strong>Part 3-B 的 Lab 3B-5 用 A/B 对照把这件事测了出来</strong>：关掉日志检查 ⇒ 少数派频繁当选；打开 ⇒ 一次都不会。看完 3-B 再回来玩这个模拟器，你会看出它和真实 Raft 的差别在哪。</p>
  </div>

  <div class="lab">
    <div class="lab-h"><div class="ico">2</div><div class="t">实验 2 · Raft
…（截断，原长 1913 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch03a patched
lab03a/raft.go noted
Lab03A.java noted
````

</details>

▸ **Bash** Add Lab 3B-5 high-term rejoin A/B experiment

````
cd 〔路径00104〕 && python3 - <<'PY'
p='raft.go'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('''	NaiveCommit bool // ★ 关掉「只提交当前任期」的限制 ⇒ 复现 Figure 8
	SlowBackoff bool // 关掉快速回退，每次只退 1''',
'''	NaiveCommit bool // ★ 关掉「只提交当前任期」的限制 ⇒ 复现 Figure 8
	SlowBackoff bool // 关掉快速回退，每次只退 1
	NoLogCheck  bool // ★ 关掉「日志至少一样新」的投票条件 ⇒ 退化成 Lab 3-A 的模型''')
rep('''		grant := m.Term == n.term && n.state != Leader &&
			(n.votedFor == NoVote || n.votedFor == m.From) &&
			upToDate(m.LastLogTerm, m.LastLogIndex, n.last()) // ★ 日志至少一样新''',
'''		grant := m.Term == n.term && n.state != Leader &&
			(n.votedFor == NoVote || n.votedFor == m.From) &&
			(c.NoLogCheck || upToDate(m.LastLogTerm, m.LastLogIndex, n.last())) // ★ 日志至少一样新''')
open(p,
…（截断，原长 4297 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
ok
  Lab 3B-5 · 高任期节点回归
  3|2 分区让少数派任期涨到几十，再恢复，看谁当选
══════════════════════════════════════════════════════════════════════════════
    分区恢复后由谁当选（40 次试验）            ① 完整 Raft      ② 关掉日志检查
    ──────────────────────────────────────────────────────────────────────────
    多数派节点（日志是新的）                            16                  16
    ★ 少数派节点（日志陈旧）                            24                  24
    没能选出 Leader                                      0                   0
    安全性违反                                           0                   0

    4. Lab 3B-5 的 ① 里少数派一次都没赢。但如果少数派那边有一个【僵尸 Leader】
       （它会持续追加日志，只是提交不了），情况会变吗？什么条件下它能赢？
       赢了之后会不会破坏安全性？（提示：想想 no-op）
````

</details>

▸ **Bash** Add writes during partition and third arm

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('''	run40 := func(noLogCheck bool) out {''','''	run40 := func(noLogCheck, writeDuringPartition bool) out {''')
rep('''			net.Partition(grp)
			adv(12000) // 少数派在这里不停超时，任期疯涨；日志则完全冻结''',
'''			net.Partition(grp)
			if writeDuringPartition {
				// 多数派继续接收写入 ⇒ 它的日志会长出新任期的条目
				for i := 0; i < 25; i++ {
					c.Submit(fmt.Sprintf("during%d", i))
					adv(40)
				}
			}
			adv(12000) // 少数派在这里不停超时，任期疯涨；日志则完全冻结''')
rep('''	a := run40(false) // 完整 Raft
	b := run40(true)  // 退化成 3-A 的模型

	w := []int{34, 20, 20}
	al := "LRR"
	tui.TableHeadA([]string{"分区恢复后由谁当选（40 次试验）", "① 完整 Raft", "② 关掉日志检查"}, w, al)
	tui.TableRowA([]string{"多数派节点（日志是新的）", fmt.Sprint(a.majorityWo
…（截断，原长 3451 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
  Lab 3B-5 · 高任期节点回归
  3|2 分区让少数派任期涨到几十，再恢复，看谁当选
══════════════════════════════════════════════════════════════════════════════
    恢复后由谁当选（各 40 次）           ① 完整 Raft    ② 无日志检查    ③ 多数派没写入
    ──────────────────────────────────────────────────────────────────────────────────
    多数派节点（日志更新）                        40              15                16
    ★ 少数派节点（日志陈旧）                       0              25                24
    没能选出 Leader                                0               0                 0
    安全性违反                                     0              25                 0

    分区期间少数派任期最高涨到     27　← 而它们的日志一条都没长

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
      先试探性问一轮"我发起选举你们会投我吗"，被隔离的节点得不到肯定答复，任期就不会涨。

══════════════════════════════════════════════════════════════════════════════
  三个思考题（答案在课件 §3.12 / §3.13
…（截断，原长 1502 字符）
````

</details>

▸ **Bash** Add Lab 3B-5 to Java version

````
cd 〔路径00134〕 && python3 - <<'PY'
p='Lab03B.java'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('    static boolean NAIVE = false, SLOW = false, VERBOSE = false;',
    '    static boolean NAIVE = false, SLOW = false, NOLOGCHECK = false, VERBOSE = false;')
rep('        boolean naiveCommit = NAIVE, slowBackoff = SLOW;',
    '        boolean naiveCommit = NAIVE, slowBackoff = SLOW;\n        boolean noLogCheck = NOLOGCHECK;   // ★ 关掉「日志至少一样新」⇒ 退化成 Lab 3-A 的模型')
rep('''                            && upToDate(m.lastLogTerm, m.lastLogIndex, n.last());  // ★ 日志至少一样新''',
    '''                            && (noLogCheck || upToDate(m.lastLogTerm, m.lastLogIndex, n.last())); // ★ 日志至少一样新''')
rep('        c.naiveCommit = N
…（截断，原长 4643 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
  Lab 3B-5 · 高任期节点回归
  3|2 分区让少数派任期涨到几十，再恢复，看谁当选
══════════════════════════════════════════════════════════════════════════════
    恢复后由谁当选（各 40 次）           ① 完整 Raft    ② 无日志检查    ③ 多数派没写入
    ──────────────────────────────────────────────────────────────────────────────────
    多数派节点（日志更新）                        40              23                23
    ★ 少数派节点（日志陈旧）                       0              17                17
    没能选出 Leader                                0               0                 0
    安全性违反                                     0              17                 0

    分区期间少数派任期最高涨到     27　← 而它们的日志一条都没长

    ▸ ① 少数派一次都赢不了。它们的任期涨到几十（回来时确实会逼现任 Leader 退位、
      触发一次不必要的选举），但日志停在分区前 —— 那条「至少和我一样新」把它们全挡住了。
      高任期节点回归是「有破坏力，没有危险」：损失的是可用性，不是安全性。
    ▸ ② 去掉日志检查后，陈旧节点频繁当选，而且【安全性真的被破坏了】——
````

</details>

▸ **Bash** Sharpen courseware notes and re-verify both chapters

````
cd 〔路径00054〕 && python3 - <<'PY'
p='ch03b/ch03b.html'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)
rep('''    <p style="margin-bottom:0"><strong>Lab 3-A 的代码<u>没有</u>实现「日志至少一样新」这条投票条件</strong>——因为那一节的模型里根本没有日志。
    我在 3-A §3.6 把它列进了选举规则，却没在代码里落实，也没有标注这个简化。<br>
    后果是真实可测的：在 3-A 的模拟器里制造 3|2 分区、让少数派任期涨上去、再恢复，<strong>陈旧的少数派节点有超过一半的概率会当选</strong>——这在真实 Raft 里不可能发生。<br>
    <strong>Lab 3B-5 用 A/B 对照把这件事测了出来</strong>：关掉日志检查 ⇒ 少数派频繁当选；打开 ⇒ 一次都不会。</p>''',
'''    <p style="margin-bottom:0"><strong>Lab 3-A 的代码<u>没有</u>实现「日志至少一样新」这条投票条件</strong>——因为那一节的模型里根本没有日志。
    我在 3-A §3.6 把它列进了选举规则，却没在代码里落实，也没有标注这个简化。<br>
    后果比"行为不一样"严重得多。<strong>Lab 3B-5 用三组对照把它测了出来</strong>（各 40 次试验）：</p>
    <table style="margin:12px 0 0">
      <thead><tr><
…（截断，原长 2917 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch03b ok
ch03a ok
naive=false => — || — || Raft 在这里踩了刹车 || 覆盖发生了，但没有造成任何伤害 || 安全的那条路
naive=true => — || — || 你关掉了限制 —— 这里 Leader 会宣布 index2 已提交 || ★ 安全性破产：已经告诉客户端"成功"的数据凭空消失了 || 安全的那条路
exp3: 朴素回退=41 | 快速回退=4 | 加速比=10.3× | 节省时间=2960
exp3 gap=200: 201 | 4 | 50.3× | 15760
ERRORS: none
exp3: 固定超时 · 平均轮数=60+ | 随机超时 · 平均轮数=1.05 | 耗时改善=55.4× | 固定超时分裂率=100%
exp3 jitter=0: 60+ | 1.07 | 54.8× | 100%
ERRORS: none
````

</details>

## 🤖 Claude · 2026-09-09 20:58

📎 **发送文件**：修订版课件：Figure 8 的 (d) 已改正（S5 是 [1,3] 不是 [1,3,3]），3-B 新增 §3.15b「高任期节点回归」，3-A 加了简化说明的警告。
- `〔路径00075〕`
- `〔路径00074〕`

<details><summary>↳ SendUserFile 结果</summary>

````
2 files delivered to user.
  〔路径00075〕 → file_uuid: 〔标识03629〕
  〔路径00074〕 → file_uuid: 〔标识03849〕
````

</details>

📎 **发送文件**：Lab 3-B 新增 3B-5（高任期节点回归的三组对照）与 -nologcheck 开关；Lab 3-A 的 raft.go 加了缺失条件的显式注释。
- `〔路径00107〕`
- `〔路径00105〕`
- `〔路径00135〕`
- `〔路径00103〕`
- `〔路径00133〕`

<details><summary>↳ SendUserFile 结果</summary>

````
5 files delivered to user.
  〔路径00107〕 → file_uuid: 〔标识00053〕
  〔路径00105〕 → file_uuid: 〔标识01889〕
  〔路径00135〕 → file_uuid: 〔标识02150〕
  〔路径00103〕 → file_uuid: 〔标识00545〕
  〔路径00133〕 → file_uuid: 〔标识03901〕
````

</details>

💾 **写入你的 Mac**：
- `〔路径00009〕`
- `〔路径00008〕`
- `〔路径00032〕`
- `〔路径00030〕`
- `〔路径00048〕`
- `〔路径00029〕`
- `〔路径00047〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00009〕","〔路径00008〕","〔路径00032〕","〔路径00030〕","〔路径00048〕","〔路径00029〕","〔路径00047〕"],"rejected":[]}
````

</details>

三个问题里有两个是真的 bug，都已经修好并重新提交到你本地了。先说这个，再回你的答案。

---

## 你的挑战 ②：Figure 8 图示 — **你是对的**

我去核对了论文原文的图注：

> "…S5 could be elected leader (with votes from S2, S3, and S4) and **overwrite the entry with its own entry from term 3**."

**"its own entry from term 3"——单数，term 3。** S5 覆盖用的就是它自己在 (b) 阶段留下的那一条 term3 旧日志，这一步并没有创建新条目。而我画成了 `1 3 3`，多出来的第三格无从解释：S5 是在 **term 5** 当选的，它若追加新日志，那条的 term 应该是 **5** 而不是 3。

已改成 `1 3`，并加了一句说明。**我的 Go/Java 代码里本来就是对的**（`e(2, 3, "Y")`，只有两条），错的只有可视化。

## 你的挑战 ③：3-A 分区恢复的 bug — **你也是对的，而且比你想的严重**

3-A 的模型里没有日志，所以投票时缺了「candidate 的日志至少和我一样新」这一条。我在 §3.6 把它列进了选举规则，却没在代码里落实，也没标注这个简化——这是我的疏漏。

我写了对照实验测这件事，各 40 次：

| 分区恢复后由谁当选 | ① 完整 Raft | ② 关掉日志检查 | ③ 多数派分区期间没写入 |
|---|---|---|---|
| 多数派节点（日志更新） | 40 | 15 | 16 |
| **少数派节点（日志陈旧）** | **0** | **25** | 24 |
| **安全性违反** | **0** | **25** | **0** |

**②那一列的 25 次不是"选错了人"，是真的丢数据**——陈旧节点当选后会用旧日志覆盖别人已提交的条目。所以 3-A 的模拟器只能当作选举流程的动画看，分区恢复那一段不要当真。我在 3-A 课件和代码里都加了显式警告。

这个对照现在是 **Lab 3B-5**，`go run . -nologcheck` 可以自己跑。

---

## 你的挑战 ①：高任期节点回归 — 这题问得很好，答案有层次

**先钉死一件事：选举从不看 commitIndex。** 只比 `(lastLogTerm, lastLogIndex)`，先比 term。一个节点根本无法验证别人报上来的 commitIndex 是否可信，所以它不能参与投票判断。

然后要把少数派里的两类节点分开——你的直觉正好覆盖了第二类：

**① 普通少数派节点**：不停超时、不停 `term++`，但**从不追加日志**（只有 Leader 才写日志，而它永远当不上）。所以**任期疯涨，日志冻结**。回来时高任期会逼现任 Leader 退位、触发一次无谓选举，但它自己**赢不了**。表格里 ①那一列的 0/40 就是这个。

**② 僵尸 Leader**（就是你说的那个）：任期冻结（它是 Leader，不涨），但**日志疯长**——而且这些日志带的是**旧 term**。这才是有意思的情况：

| 多数派新 Leader 的状态 | 僵尸能赢吗 | 为什么 |
|---|---|---|
| 已追加并提交了自己任期的日志 | **不能** | 新 Leader 的 lastLogTerm 更大，僵尸一比就输——**哪怕它的日志长十倍** |
| 刚当选，还没写过自己任期的日志 | **可能能** | 两边 lastLogTerm 相同，此时比长度，僵尸更长 |

第二行看着危险，其实**安全**：僵尸的日志是所有已提交条目的**超集**，它当选后只是把那些"悬而未决"的条目最终坐实了，不是数据损坏。

**但这个窗口正是 no-op 存在的第二个理由。** 我在 §3.12 只说了 no-op 是"帮前任日志过提交线"，其实它同时把这条缝也堵上了——新 Leader 一当选就追加空日志，`lastLogTerm` 瞬间抬到当前任期，第二行的情况直接消失。一石二鸟。

**它和 Figure 8 是同一件事吗？不是，但共用同一块地基。** Figure 8 问的是"已提交的日志会不会被覆盖"；这里问的是"陈旧节点能不能当选"。两者都落在同一条比较规则上——`(lastLogTerm, lastLogIndex)` **先比 term**。Figure 8 里 S5 能赢正因为它 term 更大（3 > 2）；这里僵尸赢不了正因为它 term 更小。**同一把尺子，量出两个方向的结论。**

顺带一个表格里最微妙的发现：**③那一列**（分区期间多数派一条都没写）——两边日志一模一样，少数派根本不"陈旧"，它当选完全合法、安全性违反 0。所以准确的说法是「**日志真的落后的节点**赢不了」，而不是「少数派赢不了」。

---

## 回你原来的三题

**1. 僵尸 Leader 本地读为什么陈旧** — 方向对，但机制比你说的更直接：僵尸的状态机冻结在分区那一刻的 commitIndex 上，新 Leader 之后提交的一切它都看不见。所以**任何在分区后被修改过的 key，从僵尸读出来都是旧值**，不需要"涉及历史 commit 日志的修改"这个条件。危害是它破坏线性一致性：客户端从新 Leader 写入拿到成功，转头从僵尸读，看到旧值。

**2. 哪些状态必须落盘** — 你答了 node id、term、commitIndex、已提交的日志。对照 Raft 论文 Figure 2，正确答案是**三样**：

| 必须持久化 | 不持久化会破坏 |
|---|---|
| `currentTerm` | 重启后任期归零，无法拒绝过时消息 |
| **`votedFor`** ← **你漏了这个** | 重启后忘了投过票，**同一任期投两次 ⇒ 两个 Leader ⇒ 脑裂**。就是 Lab 3A-4 演示的那个 |
| `log[]` | 已提交的条目在某个副本上消失 ⇒ **Leader 完整性被破坏** |

`votedFor` 是你唯一漏掉的，也恰恰是最重要的那个。

而 **`commitIndex` 其实<u>不需要</u>持久化**——它是 volatile 的，重启后从 Leader 的 `AppendEntries.leaderCommit` 就能重新学到。你把它列进来了，这个直觉不算错（有了快照之后 `lastApplied` 确实需要知道进度），但按论文的最小集合它不在里面。

**3. 成员变更** — "裂脑，配置切换时间不同" 完全正确。精确的失败长这样：从 {A,B,C} 扩到 {A,B,C,D,E} 的瞬间，如果 A、B 还在用旧配置而 C、D、E 已经切到新配置，那么旧配置的过半是 2（A、B），新配置的过半是 3（C、D、E）——**这两个集合不相交**，同一任期可以选出两个 Leader。Raft 整套安全性的地基"两个过半集合必相交"在这里被打穿了。这正是 3-C 要讲的联合共识和单节点变更要解决的问题。

---

**3-C 开始之前一个问题**：你既然已经想到"配置切换时间不同会裂脑"，那**单节点变更**（一次只加/减一个节点，比如 3 → 4 → 5）为什么能避免它？先想想 {A,B,C} 和 {A,B,C,D} 这两个配置的过半集合有什么关系。

> ✓ 本轮结束 · 2026-09-09 20:59 · 14 步
