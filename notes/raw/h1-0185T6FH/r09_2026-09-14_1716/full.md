## 👤 你 · 2026-09-14 17:16

1. 都不用跳过，作为查缺补漏了，可重复读可以防住幻读因为MySQL会使用其MVCC进行事务对应版本快照的读，其通过redo log中记录对应事务所处版本来锁定对应版本事务应该读到的修改有哪些，超出这个版本的修改对于这个事务不可见。
2. 第3章主要讲了raft但是我怎么感觉raft好像读写都是leader负责的？虽然你讲了follower look但是好像这是特定产品的特性，是因为raft作为共识算法主要是保证一致性对于读性能没有特别优化？然后quorum读可以算是一种相对更一致的但是不止依赖leader的读方式？这样会不会对集群吞吐造成比较大的负担？去quorum写又是什么？就是raft的日志append吗？
3. 还有raft的配置是如何启用的？比如每次集群增减机器都会重新选举吗？配置是从leader下发的吗？如果是冷启动怎么办？怎么知道有几个机器，机器都怎么访问到？
4. raft做的都是日志的记录，这个是不是和redo log类似？但是对于想kv和DB这种方式是不是代表执行日志本身还有延迟，真的涉及到读对应内容时是怎么确定具体内容已经和版本对应上？

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

> —— 上下文在此处被自动压缩 ——

<details><summary>📋 上下文压缩摘要（点击展开）</summary>

This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.

Summary:
## 1. Primary Request and Intent

The user (〔用户名〕, 〔邮箱〕, 〔公司〕) asked in Chinese for a **systematic from-zero course on distributed algorithms and their applications**. Their proposed 3-stage roadmap was (1) basic definitions like visibility/eventual consistency, (2) algorithms like 2PC/TCC/Lamport/Raft, (3) applications like MapReduce/Bigtable/Spanner/ZooKeeper/Kafka/RocketMQ/Vitess — but they explicitly said I need not follow it and could organize it as I saw fit.

Requirements they stated:
- Mix of **description, Q&A, interaction, and hands-on practice** (描述、问答、互动、实操相结合)
- Each stage must produce a **reusable, reviewable interactive learning artifact** ("变成一个实际可复用可复习的课程")
- Start with the beginner-oriented preface/introduction

Clarified via AskUserQuestion:
- **Hands-on language: Go and Java** (both)
- **Starting level: complete beginner** (完全零基础)
- **Artifact form (multi-select): interactive HTML courseware + runnable code lab repo + per-chapter Q&A quizzes**
- **Depth: engineering intuition first, theory to the point** (工程直觉优先)

Later explicit redirections:
- "先不用管我机器上的服务，先着重在分布式系统原理和流行程序和中间件" — **stop analyzing their own codebase; focus on principles and popular systems/middleware**
- "拆开来做吧" — split Part 3 into three sub-parts (3-A election, 3-B log replication, 3-C production concerns)
- For Part 4: "都不用跳过，作为查缺补漏了" — **don't skip anything in 4-A (single-machine transaction fundamentals)**

## 2. Key Technical Concepts

**Distributed systems theory:** partial failure; the three "free assumptions" of single-machine programming (shared now / binary call outcome / single copy of data); unreliable network (loss, delay, reorder, duplicate, partition), clock, process; crash-stop / crash-recovery / omission / Byzantine failure models; safety vs liveness; two generals problem; FLP impossibility; at-most-once vs at-least-once vs effectively-once; idempotency keys; the "third state" (indeterminate).

**Time and order:** wall clock vs monotonic clock vs logical clock; happens-before partial order; Lamport clocks (only falsify, never verify); vector clocks (iff condition); LWW dangers; HLC (l,c); Spanner TrueTime + commit wait.

**Replication and consistency:** leader-based / multi-leader / leaderless topologies; three replication-lag anomalies (read-your-writes, monotonic reads, consistent prefix reads); consistency ladder (linearizable ⊃ sequential ⊃ causal ⊃ eventual); CAP misreadings + PACELC; Quorum W+R>N and the **intersection property**; sloppy quorum/hinted handoff breaking W+R>N; CRDTs (G-Counter, PN-Counter, OR-Set, LWW-Register); "convergence ≠ correctness".

**Consensus:** consensus ≡ total order broadcast ≡ state machine replication; 2PC blocking and why it is NOT consensus (all-agree vs majority; the decision is not replicated); 3PC trades safety for liveness; Raft roles/terms (term = Lamport clock); election rules (4 rules); randomized election timeout as FLP workaround; split vote; log matching property; commit rule with the **current-term restriction (Figure 8)**; five safety properties; conflict fast backoff by term; zombie leader; PreVote; CheckQuorum; persistence of currentTerm/votedFor/log; snapshots + InstallSnapshot; single-node membership change proof (maj_old + maj_new = n+2 > n+1 = |C_old ∪ C_new|); joint consensus; ReadIndex / Lease Read / Follower Read.

**Recurring meta-theme:** "two majority sets must intersect" appears four times in Raft (election safety, leader completeness, quorum read/write, membership change).

**Tooling:** Go 1.24.7, JDK 21 (single-file source launch `java X.java`, JDK 17+ required), Playwright + preinstalled Chromium at `〔路径00178〕`, dataviz skill validated palette.

## 3. Files and Code Sections

All files exist in **both** `〔路径00055〕` (cloud workspace) and `〔路径00003〕` (user's Mac, via device_commit_files).

### Courseware (self-contained HTML, dark-first, shared CSS extracted to `〔路径00240〕`)
- **`courseware/ch00-intro.html`** — Ch0 preface/intro. 6 interactive experiments: availability calculator (log-scale bar chart), unreliable network sandbox (rAF message animation showing client-believed vs server-actual), clock drift/order inversion, "you be the judge" timeout quiz, retry hell vs idempotency key, two generals infinite regress. 20-term glossary, 7-part roadmap, 12 quiz questions.
- **`courseware/ch01-time-order.html`** — Part 1. Spacetime diagram editor (click to add events, drag messages, live Lamport + vector clocks, relation verdict panel exposing Lamport blind spots), LWW silent-loss reproduction, HLC clock-rollback table. 10 quiz questions.
- **`courseware/ch02-replication.html`** — Part 2. Replication-lag anomaly replay, **consistency checker running real exhaustive DFS** (linearizable/sequential/causal/RYW/monotonic over 6 histories), Quorum configurator, CRDT vs LWW counter. 10 quiz questions. Opens with the user's own `ON CONFLICT` question.
- **`courseware/ch03a-consensus-election.html`** — Part 3-A. 2PC blocking replay (4 scenarios), **Raft election simulator** (real state machine + message passing, click-to-kill nodes, partition, safety counter), fixed vs randomized timeout batch comparison. Contains an added `note warn` block admitting the missing log-freshness check.
- **`courseware/ch03b-log-replication.html`** — Part 3-B. Log replication + conflict backoff (incl. paper Figure 7's six followers), **Figure 8 interactive replay with a "turn off the restriction" toggle**, backoff strategy comparison. New section **§3.15b 高任期节点回归**. 10 quiz questions.
- **`courseware/ch03c-production-raft.html`** — Part 3-C. Crash-restart persistence scenarios (4 cases), **membership-change split-brain simulator** (sliders for old/new size, shows the two majority sets and whether they can be disjoint), three-reads comparison on a zombie leader. Complete Raft rules table. 10 quiz questions.

### Go labs (module `dsc`)
- **`go/internal/tui/tui.go`** — CJK-aware terminal alignment shared by all labs:
```go
func Dispw(s string) int   // CJK chars count as 2 columns
func PadR/PadL(s string, w int) string
func Head(part, n int, title, subtitle string)
func HeadN(label, title, subtitle string)   // e.g. "3A-1"
func Row(k string, v ...any)
func TableHead/TableRow(cols []string, w []int)
func TableHeadA/TableRowA(cols []string, w []int, align string) // align e.g. "LRRRR"
```
- **`go/lab00/`** — net.go (`Call(timeout, exec)` with 4 annotated paths — the third-state teaching core), bank.go (DeductNaive / DeductIdempotent / DeductSplitDedup), client.go (`DoWithRetry`), main.go (4 experiments).
- **`go/lab01/`** — clocks.go (Lamport, Vector with 5-line `Compare`, HLC), trace.go (event trace matching the courseware demo exactly + `causalPath` independent verification that panics on mismatch), main.go.
- **`go/lab02/`** — quorum.go (N-replica KV with W/R/lag), crdt.go, history.go (`CheckLinearizable`/`CheckSequential` differ only in the `canPlace` closure), main.go.
- **`go/lab03a/`** — net.go (deterministic virtual-clock network), raft.go (election state machine ~150 lines + `checkSafety`), main.go.
- **`go/lab03b/`** — raft.go with log replication; the Figure 8 fix is one line in `maybeCommit`:
```go
// ★★★ 这一行就是 Figure 8 的那条限制 ★★★
if !c.NaiveCommit && n.log[N].Term != n.term {
    continue
}
```
  Flags: `-naivecommit`, `-slowbackoff`, `-nologcheck`. Lab 3B-5 added for high-term rejoin.
- **`go/lab03c/`** — raft.go extended with `Persisted{CurrentTerm,VotedFor,Log}`, `persist()`, `Restart(i)` with `LoseTerm/LoseVote/LoseLog`, `MaybeSnapshot`/`entryAt`/`sendSnapshot`/`handleInstallSnapshot`, per-node `config` + `majorityOf()`, `leaseTill`/`ackTicks`, `Read(i, key, mode)` with LocalRead/ReadIndexRead/LeaseRead, and:
```go
// ActiveLeaders 数一数有几个 Leader【能在自己的配置下凑齐过半】——
// 这才是「裂脑」的准确定义。僵尸 Leader 不算。
func (c *Cluster) ActiveLeaders() int
```
  main.go has `constructed(loseVote bool)` — a deterministic double-vote construction (N1/N2 both candidates in term 5, N5 votes N1, crashes, restarts, votes N2).

### Java labs (single-file, JDK 17+, `System.setOut(... UTF_8)` for CJK)
`java/lab00/Lab00.java`, `lab01/Lab01.java`, `lab02/Lab02.java`, `lab03a/Lab03A.java`, `lab03b/Lab03B.java`, `lab03c/Lab03C.java` — each mirrors its Go counterpart with matching output format and identical conclusions.

### README.md
Quick-start commands, per-lab parameter tables, 7-part roadmap status table, directory tree, "what each lab reproduces" tables.

## 4. Errors and Fixes

1. **FALSE CLAIM about artifacts** — I twice told the user the courseware was "saved as a desktop artifact (`ds-course-ch00` / `ds-course-ch01`)". There is **no artifact tool in this session**. I corrected this openly: "这个说法是错的，我没有真的发布过 artifact… 抱歉给了你一个不存在的东西." Courseware lives only as HTML files in their repo + conversation cards.

2. **ASCII double quotes inside Chinese text in JS string literals (happened 3 times)** — broke page load with `Unexpected identifier`:
   - ch03a: `它们只知道"我俩都投了 yes、都没收到指令"` → `「…」`
   - ch03c: `回复客户端"成功"` ×2 and `Leader 完整性依赖"持有已提交日志的节点构成过半"` → `「…」`
   Built a precise checker that walks from each `key:"` to the first unescaped `"` and flags a following CJK char.

3. **CJK table misalignment in Go/Java output** — `%-20s` counts runes not display columns. Fixed with `Dispw`/`PadR`/`PadL` and per-column alignment strings; added spacer columns where right-aligned and left-aligned columns abutted.

4. **Java text blocks stripping indentation** — added `.indent(4)`. Once I sliced `""");` with `e[:-3]` producing `"".indent(4));` (broken terminator); fixed by replacing `"".indent(4));` → `""".indent(4));`.

5. **Playwright not installed / wrong browser path** — `npm install playwright` then `executablePath:'〔路径00178〕'`.

6. **USER-REPORTED BUG #1 (correct): Figure 8 panel (d) rendering** — "figure8的图示在term5的展示时好像有错误还是term3". I WebFetched raft.pdf and confirmed the caption: *"S5 could be elected leader (with votes from S2, S3, and S4) and **overwrite the entry with its own entry from term 3**"* — singular. My `logs[4] = [1,3,3]` was internally inconsistent (S5 elected in term 5 would create term-5 entries). **Fixed to `[1,3]`** plus an explanatory note. The Go/Java code was already correct (`e(2, 3, "Y")`).

7. **USER-REPORTED BUG #2 (correct, and worse than they thought): Part 3-A post-partition re-election** — "3-a的网络分区后一段时间分区恢复的重新选主好像也存在bug". Confirmed by diagnostic: **25/40 seeds the stale minority node wins**. Root cause: Lab 3-A has no logs so the "candidate's log at least as up-to-date" vote condition was listed in §3.6 but never implemented or flagged. Fixes applied:
   - Added `-nologcheck` flag + **Lab 3B-5** three-arm comparison: ① full Raft 40/40 majority, 0 violations; ② no log check 25/40 minority wins with **25 safety violations**; ③ majority made no writes 24/40 minority wins but 0 violations (they aren't actually stale)
   - Added §3.15b to ch03b with the results table
   - Added `note warn` to ch03a §3.6 and explicit comments in `lab03a/raft.go` and `Lab03A.java`

8. **Lab 3C-3 initially reported 7/25 split brain for single-node change (should be 0)** — two causes: (a) node 4's config wasn't set correctly in the `single` scenario; (b) `twoLeadersEver` counted any two live leaders including harmless zombies. Fixed by killing node 4 (not in either config) and introducing `ActiveLeaders()` (leaders that can actually reach a majority under their own config). Result became **30/30 vs 0/30**.

9. **Lab 3C-1 random fault injection found 0 violations for votedFor loss** — needed a very specific interleaving. Fixed by adding a **deterministic constructed scenario**, which turned the failure into a meta-lesson: "随机跑几十遍没出事不等于安全，只等于没撞上 —— 这正是 TLA+ 和 Jepsen 存在的理由".

## 5. Problem Solving

- Reorganized the user's 3-stage roadmap into 7 parts, adding a Part 0 mental model (three free assumptions) and a single-machine transaction primer before distributed transactions.
- Every claim in the courseware is backed by a runnable experiment producing hard numbers (e.g. LWW loses 66.5% of writes; fixed timeout = 245 split votes and no leader; naive commit = 37/60 seeds with multiple same-term leaders; local read = 25/25 stale).
- Built deterministic virtual-clock network simulators so consensus bugs reproduce from a seed.
- Verified every courseware page in headless Chromium with scripted interaction and zero console errors before delivery.

## 6. All user messages

1. *"我想系统地了解并深入学习分布式算法及其对应应用。请你帮助我从零开始进行学习。我预估的大致学习路线是这样的： 1. 分布式的各种基础定义，如可见性，最终一致性等 2. 分布式的各种算法，如两阶段提交、TCC、lamport、raft等 3. 分布式的各种应用及，如mapreduce，bigtable，spanner，zookeeper、Kafka、rocketmq、vitess等 但是你不需要遵循我的思路进行，你可以按照你认为合适的方式进行展露。我们最好以描述、问答、互动、实操相结合的方式完成学习，并且希望每阶段的学习完成都能总结出对应的交互式学习的产物，变成一个实际可复用可复习的课程。请你先开始最基础的面向初学者的序言和介绍吧。"*

2. AskUserQuestion answers: 实操语言="go和java吧"; 起点="完全零基础"; 产物="交互式 HTML 课件,可运行的代码实验仓库,每章一套问答测验"; 深度="工程直觉优先，理论点到为止"

3. *"1. 不能，也可能是本身服务就没执行完（如执行过程中遇到stop the world）或者是由于网络原因下游的返回没有送到，甚至是执行中下游服务直接宕机。这个时候如果下游可以保证接口幂等或者是有其它验证下游是否触发的方式（如共享的数据库表状态），可以等待一段时间后重试或者返回执行任务失败。2. 因为这些分布式系统一般涉及到投票选主和数据同步多数完成这些动作，如果4台那么就相当于要有3台达成一致，总体来说不太经济。所以一般是3台或5台这样多数就是2台和3台。3. 这个我一时没有印象，只记得之前看redis和nginx好像都提到过时钟偏移和单调时钟防止时间明明向后但是实际值却回拨了的情况。我想这个在实际场景会影响到可见顺序或者超时剔除？"*

4. *"1. 用的是updated_at，这个时间戳是数据库主库打的 2. 由于lamport时钟只能用来证伪即当a happens before b的时候一定不存在L(a)>=L(b)，因此想要一定有a happens before b需要进一步增加配合进程ID，形成（PID， Lamport）对通过排序确定，即使用lamport的另一个特性在单个进程下全序的能力来实现。3. 有，主要是〔内部库C〕和〔内部服务A〕，其中后者负责统一对用户录入、提取的指标进行指标的标准化，〔内部表Y〕到〔内部表X〕；这个表只由〔内部服务A〕来写从而达到〔内部服务A〕统一盖戳的目的。同时现在这些服务中大部分冲突通过on conflict语义使用数据库来解决，这应该算是一种"数据库唯一约束"+"LWW"？"*

5. *(interrupted, then re-sent)* **"1. 先不用管我机器上的服务，先着重在分布式系统原理和流行程序和中间件；应该是单实例，在更新过程中确实会有由于提交时间导致的cronjob扫描不全的问题，这个通过增加一个时间跨度更长周期更长的cronjob做sweep去兜底 2. 拆开来做吧"**

6. *"1. 因为不知道是否有其他的参与者 2. 应该是旧数据，应该只影响不能再做写了 3. 我觉得可以了"*

7. *"1. 因为新的leader会有新的日志append也会有新的commit，这导致新commit的那段日志僵尸leader拿不到… 2. 当前node id、term、commitIdex和已经commit的日志，没落盘会导致选举时的term可能和已经提交的日志不一致或更小 3. 可能导致裂脑，配置切换时间不同导致重新选举时用了旧的集群数开始发送日志 我还有额外的问题：1. 这样是怎么解决网络分区后少数区leader，term大量增长后分区恢复在选举时成为主，是否会导致选主优先选老的，还是会通过commitIndex做判断？… 2. 还有figure8的图示在term5的展示时好像有错误还是term3 3. 还有3-a的网络分区后一段时间分区恢复的重新选主好像也存在bug对应我的第一个问题。"*

8. *"1. {A,B,C} 和 {A,B,C,D}配置过半集合一个是要两个算leader一个是要三个算leader，新增一个导致了过半集合的数量定义变化，理论上会触发重新选举 2. 3->4应该由于过半数量的变化导致重新选举但是4->5应该不会触发，新加入；一次加减一个可以方便leader同步当前集群过半集合数量状态？"*

9. **(MOST RECENT)** *"1. 都不用跳过，作为查缺补漏了，可重复读可以防住幻读因为MySQL会使用其MVCC进行事务对应版本快照的读，其通过redo log中记录对应事务所处版本来锁定对应版本事务应该读到的修改有哪些，超出这个版本的修改对于这个事务不可见。 2. 第3章主要讲了raft但是我怎么感觉raft好像读写都是leader负责的？虽然你讲了follower look但是好像这是特定产品的特性，是因为raft作为共识算法主要是保证一致性对于读性能没有特别优化？然后quorum读可以算是一种相对更一致的但是不止依赖leader的读方式？这样会不会对集群吞吐造成比较大的负担？去quorum写又是什么？就是raft的日志append吗？ 3. 还有raft的配置是如何启用的？比如每次集群增减机器都会重新选举吗？配置是从leader下发的吗？如果是冷启动怎么办？怎么知道有几个机器，机器都怎么访问到？ 4. raft做的都是日志的记录，这个是不是和redo log类似？但是对于想kv和DB这种方式是不是代表执行日志本身还有延迟，真的涉及到读对应内容时是怎么确定具体内容已经和版本对应上？"*

No security-relevant constraints or credential-handling instructions were stated by the user at any point.

## 7. Pending Tasks

1. **Answer the four questions in the most recent user message** (RR/phantom correction, Raft read/write + quorum confusion, Raft config/bootstrap, Raft log vs redo log + version alignment).
2. **Build Part 4-A**: ACID, isolation levels, four concurrency anomalies (dirty read / non-repeatable read / phantom / **write skew**), MVCC, snapshot isolation — with interactive courseware + Lab 4-A in Go and Java + quiz. User said **"都不用跳过，作为查缺补漏了"** (don't skip any of it).
3. Later: **Part 4-B** (2PC/XA, TCC, Saga, local message table, RocketMQ transactional messages) and **Part 4-C** (Percolator/TiDB, Calvin, Spanner external consistency, how 2PC + Raft removes blocking).
4. Remaining roadmap: Part 5 (classic systems: GFS/MapReduce/Bigtable/Dynamo/Spanner/ZooKeeper/Kafka/RocketMQ/Vitess), Part 6 (chaos engineering, Jepsen, TLA+, distributed locks/fencing tokens).

## 8. Current Work

Part 3 was fully completed and delivered (courseware + Go/Java labs committed to `〔路径00003〕`, README updated, all tasks marked complete). I then posed a Part 4-A warm-up question: **"MySQL 的可重复读（RR）能防住幻读吗？"**

The user replied with the four questions listed above. I had loaded `WebFetch`/`WebSearch` intending to verify etcd bootstrap flag names (`--initial-cluster`, `--initial-cluster-state`, `--initial-cluster-token`) before asserting them in the answer to their question 3, because I had previously made a false factual claim (the artifact claim) and wanted to avoid repeating that.

My planned answers:
- **Q1 correction:** MVCC version info lives in the **undo log** (plus hidden columns `DB_TRX_ID`, `DB_ROLL_PTR`), **not the redo log** — redo log is for crash recovery/roll-forward. RR prevents phantoms via **two** mechanisms: snapshot reads use a ReadView held for the whole transaction; current reads (`SELECT ... FOR UPDATE`, `UPDATE`, `DELETE`) use **Next-Key Locks**. The user only named the first. And critically: **RR does NOT prevent write skew** — that's the hook into Part 4-A.
- **Q2:** Raft's strong leader is definitional (all writes via leader); Follower Read is not in the paper but is a safe extension (it still obtains readIndex from the leader); Quorum W+R>N is a *different* (leaderless/Dynamo) system — **quorum write ≠ Raft log append** (Raft's append has strict ordering and the log matching property; Dynamo's quorum write is unordered and produces conflicts requiring vector clocks/LWW); throughput scaling answer is **Multi-Raft**.
- **Q3:** config is a special log entry replicated via the normal AppendEntries path, adopted on receipt (not on commit); adding/removing does not trigger an election unless the leader itself is removed; bootstrap via static initial member list vs single-node bootstrap + `member add`; member entries carry addresses; snapshots must include the config.
- **Q4:** Raft log = **logical, deterministic commands** for state-machine replication; redo log = physical/physiological for crash recovery — real systems have both (TiKV: Raft log + RocksDB WAL). `commitIndex ≠ lastApplied`; **ReadIndex step (c) "wait until lastApplied ≥ readIndex" IS the version-alignment mechanism**; a second orthogonal version layer exists in MVCC storage (Percolator `start_ts`), analogous to MySQL's ReadView but with a global timestamp.

I had also decided to capture Q2/Q3/Q4 as a durable **"Part 3 附录 / FAQ"** courseware page since they are exactly the gaps most Raft tutorials leave, before moving on to Part 4-A.

## 9. Optional Next Step

Answer the user's four questions in full (no tool calls needed for Q1/Q2/Q4; optionally verify etcd bootstrap flag names before asserting specifics in Q3), then build the **Part 3 附录 (FAQ)** courseware page covering questions 2–4, and then begin **Part 4-A**.

Direct quotes anchoring this:
- User: *"1. 都不用跳过，作为查缺补漏了，可重复读可以防住幻读因为MySQL会使用其MVCC进行事务对应版本快照的读，其通过redo log中记录对应事务所处版本来锁定对应版本事务应该读到的修改有哪些"*
- User: *"3. 还有raft的配置是如何启用的？比如每次集群增减机器都会重新选举吗？配置是从leader下发的吗？如果是冷启动怎么办？怎么知道有几个机器，机器都怎么访问到？"*
- My prior message: *"Part 4 分布式事务，我建议这样切… **4-A** 补单机课：ACID、隔离级别、四类并发异常（脏读/不可重复读/幻读/**写偏斜**）、MVCC、快照隔离 … 先给我一个判断题热身：**MySQL 的可重复读（RR）能防住幻读吗？**"*

If you need specific details from before compaction (like exact code snippets, error messages, or content you generated), read the full transcript at: /root/.claude/projects/-home-claude/〔标识00226〕.jsonl
Continue the conversation from where it left off without asking the user any further questions. Resume directly — do not acknowledge the summary, do not recap what was happening, do not preface with "I'll continue" or similar. Pick up the last task as if the break never happened.

</details>

▸ **ToolSearch**

````
{"max_results": 5, "query": "select:WebFetch,WebSearch,TaskCreate,TaskUpdate"}
````

<details><summary>↳ ToolSearch 结果</summary>

````
[tool_reference]
[tool_reference]
[tool_reference]
[tool_reference]
````

</details>

▸ **Bash** Check course workspace state

````
ls -la 〔路径00055〕 && echo "---courseware---" && ls -la 〔路径00070〕 && echo "---go---" && ls 〔路径00080〕 && echo "---java---" && ls 〔路径00122〕
````

<details><summary>↳ Bash 结果</summary>

````
total 60
drwxr-xr-x 11 root   root    4096 Sep 10 04:52 .
drwxr-xr-x  9 claude claude  4096 Aug 20 02:21 ..
-rw-r--r--  1 root   root   13585 Sep 10 05:08 README.md
drwxr-xr-x  2 root   root    4096 Aug 20 02:35 ch00
drwxr-xr-x  2 root   root    4096 Aug 20 04:57 ch01
drwxr-xr-x  2 root   root    4096 Aug 31 05:50 ch02
drwxr-xr-x  2 root   root    4096 Sep  1 11:27 ch03a
drwxr-xr-x  2 root   root    4096 Sep  5 12:16 ch03b
drwxr-xr-x  2 root   root    4096 Sep 10 04:52 ch03c
drwxr-xr-x  2 root   root    4096 Sep 10 05:08 courseware
drwxr-xr-x  9 root   root    4096 Sep 10 04:59 go
drwxr-xr-x  8 root   root    4096 Sep 10 05:06 java
---courseware---
total 496
drwxr-xr-x  2 root root  4096 Sep 10 05:08 .
drwxr-xr-x 11 root root  4096 Sep 10 04:52 ..
-rw-r--r--  1 root root 99625 Aug 20 02:45 ch00-intro.html
-rw-r--r--  1 root root 72176 Aug 20 04:57 ch01-time-order.html
-rw-r--r--  1 root root 86634 Aug 31 06:05 ch02-replication.html
-rw-r--r--  1 root root 81269 Sep  9 12:57 ch03a-consensus-election.html
-rw-r--r--  1 root root 75441 Sep  9 12:57 ch03b-log-replication.html
-rw-r--r--  1 root root 71707 Sep 10 05:08 ch03c-production-raft.html
---go---
go.mod
internal
lab00
lab01
lab02
lab03a
lab03b
lab03c
---java---
lab00
lab01
lab02
lab03a
lab03b
lab03c
````

</details>

▸ **WebFetch**

````
{"prompt": "List the exact flag names used for static bootstrap of an etcd cluster: initial cluster member list, cluster state (new vs existing), cluster token, advertise peer urls. Also describe what happens when a new member is added to an existing cluster (member add) and what --initial-cluster-state=existing means.", "url": "https://etcd.io/docs/v3.5/op-guide/clustering/"}
````

▸ **Bash** Read README head

````
cd 〔路径00054〕 && sed -n '1,60p' README.md
````

<details><summary>↳ Bash 结果</summary>

````
# 分布式系统：从零到共识

一套「讲解 + 交互课件 + 可运行实验 + 自测」四位一体的学习课程。
每一章都包含：一份可在浏览器打开的交互式课件，一组 Go 与 Java 双语言实验，
以及一套自测题。

---

## 快速开始

```bash
# 1. 打开交互课件（浏览器里直接玩模拟实验）
open courseware/ch00-intro.html          # 第 0 章：序言与导论
open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
open courseware/ch02-replication.html   # Part 2：复制与一致性模型
open courseware/ch03a-consensus-election.html  # Part 3-A：共识与 Raft 选举
open courseware/ch03b-log-replication.html    # Part 3-B：日志复制与安全性
open courseware/ch03c-production-raft.html   # Part 3-C：持久化 / 快照 / 成员变更 / 线性一致读
# Linux 用 xdg-open

# 2. 跑实验（Go 版，需要 Go 1.21+）
cd go/lab00 && go run .      # Lab 0：不可靠信道与第三态
cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC
cd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT
cd go/lab03a && go run .     # Lab 3-A：确定性网络 + Raft 选举 + 安全性断言
cd go/lab03b && go run .     # Lab 3-B：日志复制 + Figure 8 复现
cd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读

# 3. 跑实验（Java 版，需要 JDK 17+，无需 Maven/Gradle）
cd java/lab00 && java Lab00.java
cd java/lab01 && java Lab01.java
cd java/lab02 && java Lab02.java
cd java/lab03a && java Lab03A.java
cd java/lab03b && java Lab03B.java
cd java/lab03c && java Lab03C.java
```

两个语言版本实现的是**完全相同的四个实验**，输出格式也一致，可以直接对照阅读。
所有实验都接受相同的参数：

```bash
go run .        -loss 0.5 -retry 10 -n 500 -seed 7 -crash 0.2
java Lab00.java -loss 0.5 -retry 10 -n 500 -seed 7 -crash 0.2
```

**Lab 0 参数**

| 参数 | 含义 | 默认 |
|---|---|---|
| `-loss` | 单向丢包率（请求方向和响应方向各判定一次） | `0.30` |
| `-retry` | 客户端最大重试次数
…（截断，原长 1743 字符）
````

</details>

<details><summary>↳ WebFetch 结果</summary>

````
# etcd Cluster Bootstrap: Static Configuration

## Key Flags for Static Bootstrap

The primary flags for statically bootstrapping an etcd cluster are:

1. **`--initial-cluster`** - Specifies all cluster members with their peer URLs (e.g., `infra0=http://〔内网地址00001〕:2380,infra1=http://〔内网地址00002〕:2380`)

2. **`--initial-cluster-state`** - Set to `new` for fresh cluster initialization

3. **`--initial-cluster-token`** - A unique identifier for the cluster, recommended when creating multiple clusters with identical configurations

4. **`--initial-advertise-peer-urls`** - The peer URL each member advertises to other cluster members

## Important Details

The documentation emphasizes that "the URLs specified in `initial-cluster` are the *advertised peer URLs*, i.e. they should match the value of `initial-advertise-peer-urls` on the respective nodes."

## Adding Members to Existing Clusters

When adding a new member to an already-running cluster, you don't use static bootstrap flags. Instead, the process involves runtime reconfiguration through the member API. The `--initial-cluster-state=existing` parameter would be used if restarting an established cluster node (rather than `new` for bootstrap), but new members are typically added dynamically after cluster formation.
````

</details>

▸ **TaskCreate** 回答四个问题（RR/幻读、Raft 读写与 quorum、配置与冷启动、日志与版本对齐）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #21 created successfully: 回答四个问题（RR/幻读、Raft 读写与 quorum、配置与冷启动、日志与版本对齐）
````

</details>

▸ **TaskCreate** 构建 Part 3-D 附录课件 ch03d-faq.html

<details><summary>↳ TaskCreate 结果</summary>

````
Task #22 created successfully: 构建 Part 3-D 附录课件 ch03d-faq.html
````

</details>

▸ **TaskCreate** 实现 Lab 3-D（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #23 created successfully: 实现 Lab 3-D（Go + Java）
````

</details>

▸ **TaskCreate** 验证 3-D 课件与实验并交付

<details><summary>↳ TaskCreate 结果</summary>

````
Task #24 created successfully: 验证 3-D 课件与实验并交付
````

</details>

▸ **TaskCreate** Part 4-A：单机事务基础（课件 + Lab + 测验）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #25 created successfully: Part 4-A：单机事务基础（课件 + Lab + 测验）
````

</details>

▸ **TaskUpdate** in_progress

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #21 status
````

</details>

## 🤖 Claude · 2026-09-14 17:26

## 1. RR 与幻读：方向对，但有三处要修

你的答案里**有一个明确的知识点错误，两个缺口**。

**错误：不是 redo log，是 undo log。**

版本链存在 undo log 里，行上有两个隐藏列：

```
| id | name | DB_TRX_ID | DB_ROLL_PTR |
                 ↑            ↑
        最后改这行的事务 id   指向 undo log 里的上一版本
```

沿着 `DB_ROLL_PTR` 一路回溯，就是这一行的历史版本链。判断"哪个版本对我可见"靠的是 **ReadView**（活跃事务 id 集合 `m_ids` + `min_trx_id` + `max_trx_id` + `creator_trx_id`），不是日志。

redo log 是**物理/物理逻辑**日志，内容长这样："把表空间 5 第 137 页 offset 92 处的 4 个字节改成 0x0000002A"。它只用于崩溃后 roll-forward，**跟可见性判断一点关系都没有**。这两个搞混很常见，但代价大——搞混了就想不通"为什么 redo log 刷盘了事务还能回滚"。

顺带一个漂亮的对照：**RC 和 RR 在 InnoDB 里的实现差异只有一行**——RC 每次快照读都新建一个 ReadView，RR 只在事务第一次快照读时建一个、整个事务复用。就这一个区别。

**缺口一：MVCC 只挡住了一半，另一半是锁。**

MVCC 管的是**快照读**（普通 `SELECT`）。当前读——`SELECT ... FOR UPDATE`、`LOCK IN SHARE MODE`、`UPDATE`、`DELETE`——读的是**最新版本**，MVCC 完全帮不上忙。InnoDB 在 RR 下靠 **Next-Key Lock（记录锁 + 间隙锁）** 锁住范围，不让别人往缝里插数据。

所以 InnoDB 防幻读是**两套独立机制**：快照读靠 ReadView，当前读靠 Next-Key Lock。你只说了前一套。

**缺口二：「RR 能防幻读」这句话本身要加限定。**

ANSI SQL-92 里，RR 是**明确允许**幻读的——这正是 RR 和 Serializable 的定义性区别。InnoDB 的 RR 比标准强。所以正确说法是：**InnoDB 的 RR 能防，标准的 RR 不保证**。PostgreSQL 的 RR（实际是快照隔离）也能防，但机制不一样。

**然后是真正的钩子——RR 防不住写偏斜（write skew）。**

医院值班表，约束是「任何时刻至少 1 个医生在岗」。现在 Alice 和 Bob 都在岗，两人同时想请假：

```
T1 (Alice)                       T2 (Bob)
SELECT count(*) WHERE on_call
  → 2 ✓ 满足 >= 2                SELECT count(*) WHERE on_call
                                   → 2 ✓ 满足 >= 2
UPDATE doctors
  SET on_call=false
  WHERE name='Alice'             UPDATE doctors
                                   SET on_call=false
                                   WHERE name='Bob'
COMMIT                           COMMIT

结果：0 个医生在岗，约束被破坏。
```

RR 挡不住，快照隔离也挡不住。因为两个事务**改的是不同的行**，没有写写冲突；Next-Key Lock 也锁不到对方（T1 根本没打算锁 Bob 那行）。它们是靠一个「读到的前提」做决策，而那个前提被对方悄悄推翻了。

这就是 Part 4-A 的主线：**快照隔离不等于可串行化，差的就是写偏斜**。防它需要显式加锁物化冲突、或 Serializable、或 PostgreSQL 的 SSI。

---

## 2. Raft 的读写、以及你把两个世界混在一起了

**（a）强 Leader 是定义，不是性能妥协。**

论文里的词是 Strong Leader：日志只从 leader 单向流向 follower，客户端所有写都经过 leader。这不是「没来得及优化读」，而是整个正确性论证的地基——Raft 的 Log Matching Property（两个日志在同一 index 同一 term 则此前全部相同）建立在「只有 leader 追加」上。允许 follower 接写，这条性质立刻崩。

Raft 是**用限制换可理解性**。Paxos 允许任意节点提议，代价是要处理乱序 chosen 的洞，难懂得多。

**（b）读走 leader 也不是白读的——你已经在 3-C 见过了。**

Leader 自己也可能已被罢免而不自知（僵尸 leader）。所以「读走 leader」其实有三档，**区别只在"我怎么确认我还是 leader"**：

| | 怎么确认 | 代价 | 你的 Lab 3C-3 跑出来 |
|---|---|---|---|
| Local Read | 不确认 | 0 | **25/25 全脏读** |
| ReadIndex Read | 发一轮心跳收过半 | 1 RTT，不写日志 | 0 脏读 |
| Lease Read | 租约没到期就不确认 | 0，但依赖时钟 | 0 脏读（时钟不飘的前提下） |

**（c）Follower Read 不是"某个产品的私货"。**

论文第 8 节其实讨论了 read-only 优化（ReadIndex + lease），follower read 是后来的工程扩展，但它的安全做法有标准形态：

```
follower 收到读请求
  → 向 leader 要一个 readIndex（leader 走上面的确认流程）
  → follower 等自己的 lastApplied >= readIndex
  → 本地读
```

关键：**一致性保证仍然来自 leader 的那次确认**，follower 只承担了「读数据」的 CPU 和带宽。所以它省的是 leader 的数据传输开销，**不省那一轮确认 RTT**。

**（d）这是你最大的混淆点：Quorum 读写 ≠ Raft 日志复制。**

这是两个不同世界的东西，只是都用了「过半」这个词：

| | Dynamo 式 Quorum（Part 2） | Raft（Part 3） |
|---|---|---|
| 谁能接受写 | **任何副本**（无 leader） | **只有 leader** |
| 写有全局顺序吗 | **没有**，并发写产生多版本 | **有**，每条日志唯一 index |
| 冲突怎么办 | 甩给读端：向量时钟 / LWW / CRDT | **不会有冲突**，leader 定序 |
| 读要干什么 | 读 R 个副本，自己合并/挑最新 | 读 leader 一个（加确认）就够 |
| 一致性 | 最终一致 / 可调 | 线性一致 |
| 代表 | Dynamo、Cassandra、Riak | etcd、TiKV、Consul |

所以**「quorum 写」不是 Raft 的日志 append**。表面都是"等过半确认"，含义完全不同：

- Dynamo 的 W：把值发给 N 个副本，等 W 个回 ok。**这 W 个副本上的值可能互相不一致**（并发写各写各的），靠读修复。
- Raft 的 commit：leader 在 index=5 放了一条 entry，等过半 follower **复制了这一条**才算提交。**过半节点的 index=5 一定是同一条 entry**。

「过半」在两边的作用也不同：
- Dynamo：W+R>N 保证读集合和写集合**至少交一个节点**，所以能读到最新值——但不知道哪个是最新，要靠版本号判断。
- Raft：过半保证**任意两次成功操作（选举 / 提交）的参与集合必然相交**，从而新 leader 必然持有全部已提交日志。

底层数学是同一条「两个过半集合必相交」，用法完全不同。你在 Part 2 学的 W+R>N 不要往 Raft 上套。

**（e）吞吐怎么办？**

单个 Raft group 的写吞吐确实被 leader 卡死。生产上主要靠：

1. **Multi-Raft / 分片**——这是最主要的手段。把 key space 切成成千上万个 Region，每个 Region 一个独立 Raft group，leader 打散到全集群。TiKV、CockroachDB 都是这样。单 group 不扩展，但 group 数量可以扩展。
2. **Batch + Pipeline**：一次 AppendEntries 带多条 entry；不等上一轮 ack 就发下一轮。
3. **Follower / Lease Read** 卸载读。
4. **并行 fsync**：leader 本地落盘和发给 follower 可以同时做（论文没要求先本地后远程）。

顺带纠正一个常见误解：3 台加到 7 台，写延迟**不一定变差**——取决于第 4 快的节点，而不是最慢的。真正变差的是 leader 的出向带宽（O(N)）和网络总流量。所以 5 或 7 是上限，不是因为延迟，是因为带宽和收益递减。

---

## 3. Raft 的配置是怎么启用的

**（a）是从 leader 下发的。** 配置变更 `C_new` 就是**一条特殊的日志条目**，走完全相同的 AppendEntries 路径。

**（b）但这里有个反直觉的点：配置一收到就生效，不等提交。**

论文原话：a server always uses the latest configuration in its log, **regardless of whether the entry is committed**.

为什么？因为如果等提交才生效，而「提交」的判定又依赖「过半是几个」，就死循环了——你得先用旧配置提交新配置，可提交的那一刻配置就该变了。Raft 选择「立即采用」，代价是配置条目可能被新 leader 覆盖回滚。之所以敢这么干，正是因为**单节点变更保证了 C_old 和 C_new 的过半集合必相交**（你上一轮自己推过：maj_old + maj_new = n+2 > n+1）。

**（c）每次增减机器都会重新选举吗？不会。**

配置变更本身不触发选举，leader 继续是 leader。两个例外：

- **被移除的是 leader 自己**：论文的做法是 leader 在 `C_new` 提交之后才 step down。有意思的是这段时间它在管理一个**不包含自己的集群**——复制日志、算过半时不把自己算进去。
- **新节点捣乱**：刚加入的节点日志是空的，要很久才追上。如果它在追赶期间选举超时，它的 term 更高会把现任 leader 打下来，但它日志不够新又选不上，白白浪费一次选举。两个解法：**learner / non-voting member**（先以不投票身份加入，追上了再升为 voter，etcd 和 TiKV 都有）和 **PreVote**（先问一轮"如果我选你会投吗"，不实际自增 term）。

所以要修正你上一轮的推理：你说 `{A,B,C} → {A,B,C,D}` 过半集合从 2 变 3「理论上会触发重新选举」——**前半句对，后半句错**。过半数确实变了，但现任 leader 只是从"等 2 个 ack"改成"等 3 个 ack"，它还是 leader。触发选举的是 leader 被移除或新节点乱发选举，不是过半数变化本身。

**（d）冷启动——这是个先有鸡还是先有蛋的问题。**

配置存在日志里，但要复制日志得先知道配置。Raft 的答案是：**初始配置不能从日志来，必须从外部给**。两条工程路径：

**路径 1：静态引导。** 每个节点启动时用参数写死完整成员列表。etcd 的写法（我刚去官方文档核对过）：

```bash
etcd --name infra0 \
  --initial-advertise-peer-urls http://〔内网地址00001〕:2380 \
  --initial-cluster infra0=http://〔内网地址00001〕:2380,infra1=http://〔内网地址00002〕:2380,infra2=http://〔内网地址00003〕:2380 \
  --initial-cluster-state new \
  --initial-cluster-token my-etcd-token
```

直接回答你的三个子问题：
- **怎么知道有几台机器**：`--initial-cluster` 里数出来的。
- **机器怎么互相访问到**：地址就写在 `--initial-cluster` 里，**是配置的一部分，不靠发现**。配置日志条目存的是 `id → 地址` 的映射，不只是 id 集合。
- `--initial-cluster-state`：`new` = 我在创建新集群；`existing` = 我在加入已存在的集群（重启节点、或先 `member add` 再启动的新节点）。
- `--initial-cluster-token`：防止你同时搭多套配置相同的集群时节点互相串台。

**关键细节**：这些 `--initial-*` 参数**只在 data dir 为空的首次启动时生效**。集群一旦跑起来，成员信息就存进 data dir 了，后续重启从本地状态读，`--initial-cluster` 被忽略。改配置文件不改集群成员——这是运维踩坑重灾区。

**路径 2：单节点引导 + 动态加入。** 先起一台「只有我一个成员」的集群（自己就是过半，立刻选自己当 leader），然后一台一台 `member add`。PD、Consul 常用这个模式。

**（e）一个连回 3-C 的推论**：因为成员地址存在配置日志里，而快照会截断日志，所以**快照必须包含当前配置**——否则做完快照重启，节点不知道同伴是谁。这就是 3-C 里 InstallSnapshot 要带 config 的原因。

服务发现（DNS SRV、K8s Service）只是帮你**生成**那份初始列表，Raft 自己不做发现。

---

## 4. Raft log 和 redo log，以及版本怎么对上

**（a）都是 WAL，但目的和内容完全不同。**

| | Raft Log | InnoDB redo log |
|---|---|---|
| 解决什么 | **多机之间达成一致的顺序** | **单机崩溃不丢已提交事务** |
| 内容 | 逻辑命令（`SET x=1`），必须**确定性** | 物理/物理逻辑（"第 7 页 offset 12 改成..."） |
| 谁读 | 所有副本的状态机 | 只有本机崩溃恢复 |
| 能重放到别的机器吗 | **能，这就是它存在的意义** | 不能，页号只对本机数据文件有意义 |
| 顺序标识 | index，**跨机器一致** | LSN，只在本机有意义 |
| 怎么截断 | 快照（snapshot） | checkpoint（刷脏页） |

「确定性」这条特别要命：Raft log 里**不能**写 `SET t = NOW()` 或 `SET x = rand()`，不同副本重放会得到不同结果，状态机就发散了。真实系统要么在 leader 上先把非确定值算好再写进日志，要么禁掉这类命令。——这和 MySQL binlog 从 statement 格式改成 row 格式是**同一个原因**。

**（b）真实系统两个都有，而且会打两次盘。** TiKV：

```
写请求 → Raft log（复制到 3 副本，自己也要落盘 = raftdb 的 WAL）
       → 过半确认 = committed
       → apply 到状态机 = 写 RocksDB kvdb（它自己又有一层 WAL）
```

所以 TiKV 花了很大力气做 Raft Engine（专用的 raft log 存储）来省掉一层。MySQL Group Replication / 阿里 X-Paxos 走的是另一条路：让 binlog 兼任 Raft log。

**（c）「执行日志本身还有延迟」——完全对，这就是 `commitIndex` 和 `lastApplied` 的差。**

Raft 里有几个 index 必须分清，这是理解你这个问题的钥匙：

```
     ┌─ lastLogIndex：已 append 到本地（可能没提交，可能被覆盖）
     │          ┌─ commitIndex：已提交（过半复制，永不回滚）
     │          │        ┌─ lastApplied：已 apply 到状态机（真正能读到）
     ▼          ▼        ▼
  [ ......  |  .....  |  ..... ]
              ↑         ↑
     这段：已保证不丢，但还没生效 ← 你说的「延迟」就在这里
```

apply 是后台线程异步做的，所以 `commitIndex` 和 `lastApplied` 之间永远有个窗口。

**（d）那读的时候怎么保证版本对上？靠等。**

完整的 ReadIndex 是**三步**，很多教程只讲前两步：

1. `readIndex := commitIndex`
2. 发一轮心跳，收到过半响应 → 确认我还是 leader，`readIndex` 是真提交点
3. **等到 `lastApplied >= readIndex`，然后才读状态机** ← 你问的就是这一步

第 3 步做的事情就是**把「日志的版本」（index）和「状态机的版本」（lastApplied）对齐**。

第 1 步还有个附加条件很多人不知道：**leader 必须在本 term 内至少提交过一条日志才能服务读**。原因你已经学过了——Figure 8 的限制让新 leader 不能直接提交旧 term 的日志，所以它刚上任时 `commitIndex` 可能还没推到该在的位置，直接拿来当 readIndex 会读到旧值。解法是新 leader 上任立刻提交一条空的 **no-op entry**，把 commitIndex 顶上去。

**（e）还有正交的第二层版本：存储引擎的 MVCC。**

Raft 的 index 回答的是「这台机器的状态机重放到哪一步了」。但一个事务型 KV 还要回答「这次读该看哪个快照」，这是**另一层**：

- **TiKV / Percolator**：key 存成 `(key, ts) → value`，事务从 TSO 拿全局时间戳 `start_ts`，读取 `commit_ts <= start_ts` 的最大版本。
- **Spanner**：用 TrueTime 的时间戳，`read at timestamp t` 会等到所有 `t' <= t` 的事务都完成（safe time）。
- **MySQL**：ReadView + undo 版本链（就是第 1 题那套），但时间戳是本地事务 id，不是全局的。

所以一次 TiDB 的快照读要过**两道版本关**：

```
① Raft 层：等 lastApplied >= readIndex   —— 这台机器的日志重放够不够新
② MVCC 层：取 commit_ts <= start_ts 的版本 —— 这次读该看哪个快照
```

一个漂亮的对照：**Raft index 之于集群，相当于 LSN 之于单机；MVCC 的 start_ts 之于快照读，相当于 ReadView 之于 InnoDB 快照读。** 第二层正是 Part 4 要展开的内容——你的第 1 题和第 4 题最后接上了同一根线。

---

问题 2、3、4 恰好是绝大多数 Raft 教程留下的三个洞，值得沉淀成课件而不是只留在对话里。我先补一个 **Part 3-D 附录**（四个 index 的可视化 + ReadIndex 三步、Quorum vs Raft 对照沙盒、冷启动脑裂模拟），然后开 Part 4-A。

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #21 status
````

</details>

▸ **Bash** Check shared CSS and page structure

````
ls -la 〔路径00240〕 2>/dev/null; echo "---"; cd 〔路径00069〕 && grep -n "</style>" ch03c-production-raft.html | head -3; echo "--- head structure ---"; sed -n '1,40p' ch03c-production-raft.html
````

<details><summary>↳ Bash 结果</summary>

````
-rw-r--r-- 1 root root 10516 Aug 20 04:38 〔路径00240〕
---
156:</style><style>
175:</style>
--- head structure ---
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 3-C · 持久化、快照、成员变更与线性一致读</title>
<style>
/* ========== 设计令牌 ========== */
:root{
  --plane:#0d0d0d;      --surface:#1a1a19;   --surface-2:#212120;
  --ink:#ffffff;        --ink-2:#c3c2b7;     --ink-3:#898781;
  --grid:#2c2c2a;       --axis:#383835;      --border:rgba(255,255,255,0.10);
  --s1:#3987e5;         --s2:#d95926;        --s3:#199e70;
  --s4:#c98500;         --s7:#9085e9;
  --ok:#0ca30c;         --warn:#fab219;      --serious:#ec835a;  --crit:#d03b3b;
  --mono:ui-monospace,SFMono-Regular,"SF Mono",Menlo,Consolas,monospace;
  --sans:system-ui,-apple-system,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;
}
html[data-theme="light"]{
  --plane:#f9f9f7;      --surface:#fcfcfb;   --surface-2:#f2f1ed;
  --ink:#0b0b0b;        --ink-2:#52514e;     --ink-3:#898781;
  --grid:#e1e0d9;       --axis:#c3c2b7;      --border:rgba(11,11,11,0.10);
  --s1:#2a78d6;         --s2:#eb6834;        --s3:#1baf7a;
  --s4:#eda100;         --s7:#4a3aa7;
}
*{box-sizing:border-box}
html{scroll-behavior:smooth;scroll-padding-top:24px}
body{
  margin:0;background:var(--plane);color:var(--ink);font-family:var(--sans);
  font-size:16px;line-height:1.75;-webkit-font-smoothing:antialiase
…（截断，原长 2246 字符）
````

</details>

▸ **Bash** Read page-specific CSS and body start

````
sed -n '156,260p' ch03c-production-raft.html
````

<details><summary>↳ Bash 结果</summary>

````
</style><style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  background:var(--surface-2);border:1px solid var(--border);color:var(--ink)}
.axm{display:grid;grid-template-columns:26px 1fr;gap:10px;align-items:start;margin:10px 0}
.axm .b{font-family:var(--mono);font-size:11px;font-weight:700;color:var(--plane);background:var(--s1);
  width:24px;height:24px;border-radius:6px;display:grid;place-items:center;margin-top:2px}
.ladder{display:grid;gap:6px;margin:16px 0}
.lrung{display:grid;grid-template-columns:168px 1fr;gap:14px;align-items:start;padding:11px 15px;
  background:var(--surface);border:1px solid var(--border);border-radius:9px}
.lrung .nm{font-weight:660;color:var(--ink);font-size:14px}
.lrung .en{font-size:11px;color:var(--ink-3);font-family:var(--mono)}
.lrung .de{font-size:13.5px;color:var(--ink-2)}
.two{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.three{display:grid;grid-template-columns:repeat(3,1fr);gap:14px}
@media(max-width:900px){.two,.three{grid-template-columns:1fr}}
.grid-wrap{overflow-x:auto}
</style>
</head>
<body>
<div class="progress" id="prog"></div>
<button class="themebtn" id="tbtn" title="切换明暗">◐</button>
<div class="wrap">
<aside>
  <div class="toc-brand">分布式系统 · Part 3-C</div>
  <nav class=
…（截断，原长 6165 字符）
````

</details>

▸ **Bash** Locate quiz engine

````
grep -n 'id="quiz"\|class="quiz\|function.*[Qq]uiz\|QUIZ\|renderQuiz\|const QS\|labfoot\|<script>' ch03c-production-raft.html | head -40
````

<details><summary>↳ Bash 结果</summary>

````
583:<script>
````

</details>

▸ **Bash** Find quiz markup

````
grep -n 'quiz\|Quiz\|自测' ch03c-production-raft.html | head -30; echo "=== tail ==="; tail -30 ch03c-production-raft.html
````

<details><summary>↳ Bash 结果</summary>

````
193:    <a href="#s6">3.23 自测题（10 题）</a>
235:  <p>3-B 的自测题 9 提过一句，这里展开。Raft 的正确性证明依赖一个前提：<em>写进"稳定存储"的东西不会丢</em>。而现实里这个前提比想象中脆弱：</p>
483:  <h2>3.23 自测题</h2>
=== tail ===
    'N1 仍然自认为是 Leader（基础 Raft 不会让它主动退位），它的状态机<strong>冻结在 x=1</strong>。<br>'+
    '现在点第 ② 步，让客户端向<em>新</em> Leader 写入。</p></div>';
  else if(!S3.read) h='<div class="note ok" style="margin:0"><div class="note-h"><span class="dot" style="background:var(--ok)"></span>写入成功</div>'+
    '<p style="margin-bottom:0">客户端向 N3 写 <code>x=99</code>，复制到 {N3,N4,N5} 过半，<strong>提交并回复客户端成功</strong>。<br>'+
    '而 N1、N2 那边完全不知情，它们的状态机还停在 <code>x=1</code>。<br>'+
    '现在点第 ③ 步：客户端（比如被 DNS 或连接池路由到了 N1）从僵尸 Leader 读 x。</p></div>';
  else h='<div class="note crit" style="margin:0"><div class="note-h"><span class="dot" style="background:var(--crit)"></span>三种读，三种结果</div>'+
    '<table style="margin:10px 0 0"><thead><tr><th style="width:22%">读的实现</th><th style="width:22%">返回什么</th><th>为什么</th></tr></thead><tbody>'+
    '<tr><td><strong>① 本地读</strong></td><td><strong style="color:var(--crit)">x=1（陈旧）</strong></td>'+
    '<td>N1 直接读自己的状态机就返回。<strong>不报错、无日志、无痕迹</strong>——客户端刚从 N3 拿到"写入成功"，转头就读到旧值，线性一致性当场破裂。</td></tr>'+
    '<tr><td><strong>② ReadIndex</strong></td><td><strong style="color:var(--ok)">报错</strong></td>'+
    '<td>N1 读之前先发一轮心跳确认自己还是主，但只有 N2 响应（1/5 &lt; 过半 3）⇒ <strong>它发现自己已经不是合法 Leader，拒绝服务</strong>。<em>给你错误，而不是给你错的数据。</em></td></tr>'+
    '<tr><td><strong>③ Lease Read</strong></td><td><strong style="color:var(--ok)">报错</stro
…（截断，原长 2368 字符）
````

</details>

▸ **Bash** Read quiz section markup

````
sed -n '483,582p' ch03c-production-raft.html
````

<details><summary>↳ Bash 结果</summary>

````
  <h2>3.23 自测题</h2>
  <p>先自己想 30 秒再展开。</p>

  <details class="q"><summary>1. 哪三样状态必须持久化？为什么 commitIndex 不在其中？</summary>
    <div class="a"><p><code>currentTerm</code>、<code>votedFor</code>、<code>log[]</code>。</p>
    <p style="margin-bottom:0"><code>commitIndex</code> 是一个<strong>可推导的结论</strong>——重启后置 0，Leader 的 <code>AppendEntries.leaderCommit</code> 会立刻把它重新推上去。而 <code>votedFor</code> 是一个<strong>不可推导的承诺</strong>：你对外许下的诺言，除了自己记着没有任何办法找回来。<em>这个"可推导 vs 承诺"的区分，是判断什么必须落盘的通用直觉。</em></p></div></details>

  <details class="q"><summary>2. 如果 votedFor 不持久化，具体会怎样？</summary>
    <div class="a"><p>节点在任期 5 投票给了 A，然后崩溃重启，<strong>"忘了"自己投过票</strong>。此时 B 也在任期 5 索票，它就再投一次。</p>
    <p style="margin-bottom:0">于是 A 和 B 可能<u>各自</u>凑齐过半 ⇒ <strong>同一任期两个 Leader ⇒ 脑裂</strong>，两个 Leader 各写各的日志，数据分叉且无法自动合并。<br>
    这和 Lab 3A-4 里"去掉每任期一票"演示的是同一个后果，只不过触发方式从"改代码"变成了"重启一下"——<em>后者在生产环境里天天发生</em>。</p></div></details>

  <details class="q"><summary>3. 快照里为什么必须存 lastIncludedIndex 和 lastIncludedTerm？</summary>
    <div class="a"><p>因为 <code>AppendEntries</code> 的一致性检查要比较 <code>prevLogIndex</code> 处的 term。日志被压缩后，那个位置可能已经不在日志数组里了。</p>
    <p style="margin-bottom:0"><code>(lastIncludedIndex, lastIncludedTerm)</code> 充当被截断部分的<strong>代表</strong>：当 <code>prevLogIndex == lastIncludedIndex</code> 时，直接拿 <code>lastIncludedTerm</code> 去比较。<em>没有它们，日志匹配性质在截断之后就断了。</em>另外快照还必须包含成员配置，否则节点恢复后不知道集群里有谁。</p></div></details>

  <details class="q"><summary>4. 为什么单节点成员变更是安全的？给出证明。</summary>
    <div class="a"><p>因为新旧配置的任意两个过半集合<s
…（截断，原长 7077 字符）
````

</details>

▸ **Bash** List available CSS classes

````
grep -o '^\.[a-z0-9-]*\|^[a-z]*\.[a-z0-9-]*\|^\.[a-z]* \.[a-z-]*' 〔路径00240〕 | sort -u | head -60
````

<details><summary>↳ Bash 结果</summary>

````
.card
.card-t
.cm
.ctrl
.dot
.eyebrow
.f
.kv
.lab
.lab-b
.lab-h
.lead
.legend
.log
.log .tsp
.note
.note-h
.path
.progress
.pstep
.pstep .de
.pstep .n
.pstep .ti
.q
.q .a
.stage
.stat
.stat .k
.stat .u
.stat .v
.stats
.sw
.sw .tr
.tag
.themebtn
.toc
.toc-brand
.two
.wrap
button.primary
````

</details>

▸ **Bash** Inspect key CSS class definitions

````
grep -n -A3 '^\.stats\|^\.stat{\|^\.log{\|^\.stage\|^\.pstep{\|^\.kv\|^\.sw{\|^\.legend\|^\.card{\|^\.lab{\|^\.note{\|^\.ctrl' 〔路径00240〕 | head -60
````

<details><summary>↳ Bash 结果</summary>

````
57:.card{background:var(--surface);border:1px solid var(--border);border-radius:12px;padding:20px 22px;margin:20px 0}
58-.card-t{font-size:12px;letter-spacing:.1em;text-transform:uppercase;color:var(--ink-3);margin-bottom:12px;font-weight:600}
59:.note{border-left:2px solid var(--s1);background:var(--surface);padding:14px 18px;margin:20px 0;border-radius:0 8px 8px 0}
60-.note.warn{border-left-color:var(--warn)}
61-.note.crit{border-left-color:var(--crit)}
62-.note.ok{border-left-color:var(--ok)}
--
79:.lab{background:var(--surface);border:1px solid var(--border);border-radius:14px;overflow:hidden;margin:26px 0}
80-.lab-h{padding:13px 20px;border-bottom:1px solid var(--border);display:flex;align-items:center;gap:10px;
81-  background:var(--surface-2)}
82-.lab-h .ico{font-size:12px;font-weight:700;color:var(--plane);background:var(--s1);width:20px;height:20px;
--
87:.ctrl{display:flex;flex-wrap:wrap;gap:18px;align-items:flex-end;margin-bottom:18px}
88-.f{display:flex;flex-direction:column;gap:5px;min-width:150px}
89-.f label{font-size:11.5px;color:var(--ink-3);letter-spacing:.05em;text-transform:uppercase;font-weight:600;
90-  display:flex;justify-content:space-between;gap:10px}
--
102:.sw{display:flex;align-items:center;gap:9px;font-size:13.5px;color:var(--ink-2);cursor:pointer;user-select:none}
103-.sw input{display:none}
104-.sw .tr{width:36px;height:20px;border-radius:12px;background:var(--axis);position:relative;transition:.2s;flex:none}
105-.sw .tr::after{content:"";posit
…（截断，原长 3507 字符）
````

</details>

▸ **TaskUpdate** in_progress

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #22 status
````

</details>

▸ **Bash** Write courseware part 1

````
cat > 〔路径00213〕 <<'HTMLEOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 3-D · Raft 附录：索引、读、配置与日志的真相</title>
HTMLEOF
cat 〔路径00240〕 >> 〔路径00213〕
cat >> 〔路径00213〕 <<'HTMLEOF'
<style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  background:var(--surface-2);border:1px solid var(--border);color:var(--ink)}
.two{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.three{display:grid;grid-template-col
…（截断，原长 10121 字符）
````

<details><summary>↳ Bash 结果</summary>

````
22996 〔路径00213〕
````

</details>

▸ **Bash** Write courseware part 2

````
cat > 〔路径00216〕 <<'HTMLEOF'

<section id="s2">
  <h2>3.25 no-op：新 Leader 上任的第一件事</h2>
  <p>这条规则藏在 Raft 论文第 8 节的一句话里，但它是 Figure 8 那条限制的<strong>直接后果</strong>——你在 3-B 已经推过一半了。</p>

  <p>回忆 3-B 的提交规则：</p>
<pre><code><span class="cm">// Leader 只能提交<strong>当前任期</strong>的日志条目；</span>
<span class="cm">// 旧任期的条目只能靠「被当前任期的条目带着一起提交」而间接提交。</span>
if log[N].term != currentTerm { continue }</code></pre>

  <p>现在看新 Leader 刚当选的那一刻。它的日志里可能有一大段<strong>已经复制到过半、但 <code>commitIndex</code> 没标记</strong>的旧任期条目——上一任 Leader 复制完就崩了，还没来得及推进 commitIndex 并广播。</p>

  <p>新 Leader 想提交它们吗？想。能直接提交吗？<strong>不能</strong>，Figure 8 那条限制不让。于是形成一个尴尬的僵局：</p>

  <div class="note crit">
    <div class="note-h"><span class="dot" style="background:var(--crit)"></span>僵局</div>
    <p style="margin-bottom:0">新 Leader 的 <code>commitInd
…（截断，原长 8404 字符）
````

<details><summary>↳ Bash 结果</summary>

````
11759 〔路径00216〕
````

</details>

▸ **Bash** Write courseware part 3

````
cat > 〔路径00219〕 <<'HTMLEOF'

<section id="s5">
  <h2>3.28 配置：从哪来、何时生效、会不会触发选举</h2>

  <h3>① 配置是从 Leader 下发的吗？是。</h3>
  <p>成员配置 <code>C_new</code> 就是<strong>一条特殊的日志条目</strong>，走完全相同的 <code>AppendEntries</code> 路径复制出去。没有第二套通道。</p>
  <p>而且它存的不只是「节点 id 的集合」，是 <strong>id → 地址 的映射</strong>：</p>
<pre><code>{ "type": "config",
  "voters": {
    "n1": "〔内网地址00001〕:2380",
    "n2": "〔内网地址00002〕:2380",
    "n3": "〔内网地址00003〕:2380" },
  "learners": {} }</code></pre>
  <p>这直接回答了「机器怎么互相访问到」：<strong>地址是配置的一部分，Raft 自己不做服务发现</strong>。DNS SRV、K8s Service 这些只是帮你<em>生成</em>那份初始列表。</p>

  <div class="note ok">
    <div class="note-h"><span class="dot" style="background:var(--ok)"></span>连回 3-C：为什么快照必须包含配置</div>
    <p style="margin-bottom:0">配置存在日志里 → 快照会截断日志 → 如果快照不含配置，节点重启后<strong>不知道同伴是谁、地址是什么</strong>，连选举都发不出
…（截断，原长 11152 字符）
````

<details><summary>↳ Bash 结果</summary>

````
16323 〔路径00219〕
````

</details>

▸ **Bash** Write courseware part 4 (quiz + lab)

````
cat > 〔路径00222〕 <<'HTMLEOF'

<section id="s8">
  <h2>3.31 自测题</h2>
  <p>先自己想 30 秒再展开。</p>

  <details class="q"><summary>1. Raft 的四个 index 分别是什么？哪些会被回滚？</summary>
    <div class="a"><p><code>lastLogIndex</code>（已 append）、<code>matchIndex[i]</code>（Leader 记的各 follower 复制进度）、<code>commitIndex</code>（已提交）、<code>lastApplied</code>（已执行）。恒有 <code>lastLogIndex &gt;= commitIndex &gt;= lastApplied</code>。</p>
    <p style="margin-bottom:0">只有 <code>lastLogIndex</code> 对应的那段尾巴<strong>会被回滚</strong>——未提交的日志随时可能被新 Leader 截断覆盖。<code>commitIndex</code> 以下的部分按定义永不回滚。<code>matchIndex</code> 是 Leader 的易失记账，新 Leader 上任会重置为 0 重新试探（3-B 的冲突回退）。</p></div></details>

  <details class="q"><summary>2. 为什么 ReadIndex 必须有第三步？只做前两步会发生什么？</summary>
    <div class="a"><p>前两步只保证「<code>readIndex</code> 这个<strong>数字</str
…（截断，原长 8442 字符）
````

<details><summary>↳ Bash 结果</summary>

````
13800 〔路径00222〕
````

</details>

▸ **Bash** Read shared JS boilerplate

````
sed -n '583,640p' ch03c-production-raft.html
````

<details><summary>↳ Bash 结果</summary>

````
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(t,a,x){const e=document.createElementNS(NS,t);for(const k in a)e.setAttribute(k,a[k]);if(x!=null)e.textContent=x;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}

$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");draw1();draw2();draw3();};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHeight)*100)+"%";
  let cur=secs[0].id;for(const s of secs)if(s.getBoundingClientRect().top<=120)cur=s.id;
  links.forEach(a=>a.classList.toggle("on",a.getAttribute("href")==="#"+cur));}
document.addEventListener("scroll",onScroll,{passive:true});onScroll();

/* ══════════════════════════════════════════════════════════════════
   实验 1 · 崩溃重启：漏掉一样持久化状态会怎样
   ══════════════════════════════════════════════════════════════════ */
const N1=["N1","N2","N3","N4","N5"];
function nd(role,term,vote,dead){return {role,term,vote,dead:!!dead};}
const SC1=[
 {n:"全部持久化（正确）",disk:"currentTerm ✓　votedFor ✓　log ✓",steps:[
  {t:"任期 5：N1 和 N2 同时成为 Candidate",
   s:[nd("C",5,"N1"),nd("C",5,"N2"),nd("F",5,"—"),nd("F",5,"—"),nd("F",5,"—")],
   note:"两个节点同时超时，各自 t
…（截断，原长 3718 字符）
````

</details>

▸ **Bash** Write JS demos 1 and 2

````
cat > 〔路径00223〕 <<'JSEOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));

$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHeight)*100)+"%";
  let cur=secs[0].id;for(const s of secs)if(s.getBoundingClientRect().top<=120)cur=s.id;
  links.forEach(a=>a.classList.toggle("on",a.getAttribute("href")==="#"+cur));}
document.addEventListener("scroll",onScroll,{passive:true});onScroll();

/* ═════════════════════════════════════════
…（截断，原长 10388 字符）
````

<details><summary>↳ Bash 结果</summary>

````
12496 〔路径00223〕
````

</details>

▸ **Bash** Write JS demo 3

````
cat > 〔路径00224〕 <<'JSEOF'

/* ══════════════════════════════════════════════════════════════════
   实验 3 · Quorum vs Raft
   ══════════════════════════════════════════════════════════════════ */
function rng(seed){let s=(seed>>>0)||1;return()=>{s^=s<<13;s>>>=0;s^=s>>>17;s^=s<<5;s>>>=0;return s/4294967296;};}
const VN=["A","B","C","D"], VC=["var(--s1)","var(--s2)","var(--s3)","var(--s4)"];
const LOSS=0.22;
function bind3(){
  const ids=[["i3n","l3n"],["i3w","l3w"],["i3r","l3r"],["i3c","l3c"],["i3s","l3s"]];
  for(const [a,b] of ids){$("#"+a).oninput=()=>{
    if(a==="i3n"){const n=+$("#i3n").value;$("#i3w").max=n;$("#i3r").max=n;
      if(+$("#i3w").value>n)$("#i3w").value=n; if(+$("#i3r").value>n)$("#i3r").value=n;
      $("#l3w").textContent=$("#i3w").value; $("#l3r").textContent=$("#i
…（截断，原长 6957 字符）
````

<details><summary>↳ Bash 结果</summary>

````
8301 〔路径00224〕
````

</details>

▸ **Bash** Write JS demo 4 and assemble page

````
cat > 〔路径00225〕 <<'JSEOF'

/* ══════════════════════════════════════════════════════════════════
   实验 4 · 冷启动沙盒
   ══════════════════════════════════════════════════════════════════ */
const SC4=[
 {n:"A · 正确的静态引导",
  desc:"三台机器拿到<strong>同一份</strong>成员列表、同一个 token，data dir 都是空的。",
  nodes:[
   {id:"n1",cl:["n1","n2","n3"],st:"new",tk:"T1",gen:1,dd:"空"},
   {id:"n2",cl:["n1","n2","n3"],st:"new",tk:"T1",gen:1,dd:"空"},
   {id:"n3",cl:["n1","n2","n3"],st:"new",tk:"T1",gen:1,dd:"空"}],
  concl:"<strong style='color:var(--ok)'>这就是期望的结果。</strong>一个集群、一个 Leader、容错 1 台。<br>"+
    "注意「有几台机器」和「怎么访问到对方」<u>全部来自那份 <code>--initial-cluster</code></u>——Raft 自己不做服务发现。"},
 {n:"B · 一个节点的列表写错了",
  desc:"运维在 n3 的启动脚本里漏掉了另外两台，<code>--initial-cluster</code> 只写了它自己。<strong>没有任何报错。</strong>",
  nodes:[
   {id:"n
…（截断，原长 7201 字符）
````

<details><summary>↳ Bash 结果</summary>

````
94793 〔路径00077〕
````

</details>

▸ **Bash** Check for broken string literals

````
cat > 〔路径00196〕 <<'PY'
import re,sys
p=sys.argv[1]
s=open(p,encoding='utf-8').read()
# only inspect the <script> region
i=s.rindex('<script>')
js=s[i:]
bad=[]
n=len(js)
k=0
while k<n:
    c=js[k]
    if c in '"\'':
        q=c; j=k+1; buf=''
        while j<n:
            if js[j]=='\\': j+=2; continue
            if js[j]==q: break
            if js[j]=='\n': break
            buf+=js[j]; j+=1
        if j<n and js[j]==q:
            k=j+1; continue
        else:
            # unterminated on this line -> suspicious
            line=js[:k].count('\n')+1
            bad.append((line, js[k:k+90].replace('\n','\\n')))
            k+=1; continue
    k+=1
off=s[:i].count('\n')
for l,t in bad[:20]:
    print("LINE",l+off,":",t)
print("suspects:",len(bad))
PY
python3 〔路径00196〕 ch03d-faq.html
````

<details><summary>↳ Bash 结果</summary>

````
suspects: 0
````

</details>

▸ **Bash** Check playwright availability

````
cd 〔路径00052〕 && ls node_modules/playwright 2>/dev/null >/dev/null && echo "playwright ok" || echo "need install"; ls 〔路径00175〕
````

<details><summary>↳ Bash 结果</summary>

````
need install
chromium
chromium-1194
chromium_headless_shell-1194
ffmpeg-1011
````

</details>

▸ **Bash** Install playwright

````
npm install playwright --silent 2>&1 | tail -3; echo "exit=$?"
````

<details><summary>↳ Bash 结果</summary>

````
exit=0
````

</details>

▸ **Bash** Verify 3-D courseware in headless Chromium

````
cat > 〔路径00254〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const p=await b.newPage({viewport:{width:1400,height:1000}});
  const errs=[];
  p.on('console',m=>{if(m.type()==='error')errs.push('CONSOLE: '+m.text());});
  p.on('pageerror',e=>errs.push('PAGEERROR: '+e.message));
  await p.goto('file://〔路径00077〕');
  await p.waitForTimeout(500);

  // ---- Demo 1 ----
  for(let i=0;i<5;i++) await p.click('#b1w');
  await p.click('#b1rep'); await p.click('#b1rep'); await p.click('#b1rep');
  await p.click('#b1cm');
  await p.click('#b1ap');
  let ib=await p.textContent('#ib1');
  console.log('D1 indexes:', ib.r
…（截断，原长 3384 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
D1 indexes: lastLogIndex = 5commitIndex = 3lastApplied = 1matchIndex[]N2=3 N3=3 N4=3 N5=3 （过半达到 3）
node:internal/process/promises:394
    triggerUncaughtException(err, true /* fromPromise */);
    ^

page.uncheck: Timeout 30000ms exceeded.
Call log:
[2m  - waiting for locator('#b1wait')[22m
[2m    - locator resolved to <input checked id="b1wait" type="checkbox"/>[22m
[2m  - attempting click action[22m
[2m    2 × waiting for element to be visible, enabled and stable[22m
[2m      - element is not visible[22m
[2m    - retrying click action[22m
[2m    - waiting 20ms[22m
[2m    2 × waiting for element to be visible, enabled and stable[22m
[2m      - element is not visible[22m
[2m    - retrying click action[22m
[2m      - waiting 100ms[22m
[2m    58 × waiting for element to be visible, enabled and stable[22m
[2m       - element is not visible[22m
[2m     - retrying click action[22m
[2m       - waiting 500ms[22m

    at 〔路径00255〕 {
  log: [
    "  - waiting for locator('#b1wait')",
    '    - locator resolved to <input checked id="b1wait" type="checkbox"/>',
    '  - attempting click action',
    '    2 × waiting for element to be visible, enabled and stable',
    '      - element is not visible',
    '    - retrying click action',
    '    - waiting 20ms',
    '    2 × waiting for element to be visible, enabled and stable',
    '      - element is not visible',
    '    - retrying click action',
    '      - waiting 100ms',

…（截断，原长 2299 字符）
````

</details>

▸ **Bash** Re-run verification

````
cd /tmp && python3 - <<'PY'
import re
s=open('〔路径00254〕').read()
s=s.replace("await p.uncheck('#b1wait');","await p.evaluate(()=>{document.querySelector('#b1wait').checked=false;});")
s=s.replace("await p.check('#b1wait');","await p.evaluate(()=>{document.querySelector('#b1wait').checked=true;});")
s=s.replace("await p.uncheck('#b2noop');","await p.evaluate(()=>{const e=document.querySelector('#b2noop');e.checked=false;e.dispatchEvent(new Event('change'));});")
s=s.replace("await p.check('#b2noop');","await p.evaluate(()=>{const e=document.querySelector('#b2noop');e.checked=true;e.dispatchEvent(new Event('change'));});")
open('〔路径00254〕','w').write(s)
PY
node 〔路径00254〕
````

<details><summary>↳ Bash 结果</summary>

````
D1 indexes: lastLogIndex = 5commitIndex = 3lastApplied = 1matchIndex[]N2=3 N3=3 N4=3 N5=3 （过半达到 3）
D1 read(no step3): 丢了 2 条已提交的写
D1 read(step3): 返回 x = 3 | 线性一致: true
D2 no-noop final: 客户端 ReadIndex 读 x① readIndex := commitIndex = 0 ② 心跳过半确认 ✓ 通过 ③ lastApplied=0 >= 0 ✓ 通过返回 （初始值，x 不存在）可客户端明明收到过「x=3 写入成功」
D2 with-noop final: 客户端 ReadIndex 读 x① readIndex := commitIndex = 4 ② 心跳过半确认 ✓ ③ lastApplied=4 >= 4 ✓返回 x = 3 —— 正是客户端上次写成功的那个值。线性一致 ✓
D3 dynamo divergence over 12 seeds: 2,2,2,2,2,2,2,2,2,2,2,2
D3 with W=N=5, C=4: 3,2,3,2,4,3,4,3
D4 scenario A: clusters=1 leaders=1 effective=3
D4 scenario B: clusters=2 leaders=2 effective=3
D4 scenario C: clusters=2 leaders=1 effective=2
D4 scenario D: clusters=2 leaders=1 effective=2
D4 scenario E: clusters=1 leaders=1 effective=3
quiz items: 10
sections: 10
NO CONSOLE ERRORS
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Capture demo screenshots

````
cat > 〔路径00241〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const p=await b.newPage({viewport:{width:1380,height:1100},deviceScaleFactor:2});
  await p.goto('file://〔路径00077〕');
  await p.waitForTimeout(400);
  for(let i=0;i<6;i++) await p.click('#b1w');
  for(let i=0;i<4;i++) await p.click('#b1rep');
  await p.click('#b1cm'); await p.click('#b1ap');
  await p.evaluate(()=>{document.querySelector('#b1wait').checked=false;});
  await p.click('#b1read');
  await p.locator('#s1 .lab').screenshot({path:'〔路径00236〕'});
  await p.evaluate(()=>{const e=document.querySelector('#b2noop');e.checked=false;e.dispatchEv
…（截断，原长 1713 字符）
````

<details><summary>↳ Bash 结果</summary>

````
mobile horizontal overflow: true
-rw-r--r-- 1 root root 272011 Sep 14 09:38 〔路径00236〕
-rw-r--r-- 1 root root 196755 Sep 14 09:38 〔路径00237〕
-rw-r--r-- 1 root root 444507 Sep 14 09:38 〔路径00238〕
-rw-r--r-- 1 root root 363680 Sep 14 09:38 〔路径00239〕
-rw-r--r-- 1 root root 152091 Aug 20 02:35 〔路径00243〕
-rw-r--r-- 1 root root 222881 Aug 20 02:35 〔路径00244〕
-rw-r--r-- 1 root root 280935 Aug 20 02:35 〔路径00245〕
-rw-r--r-- 1 root root 237954 Aug 20 02:35 〔路径00246〕
````

</details>

▸ **Read** `〔路径00236〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Read** `〔路径00239〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

## 👤 你 · 2026-09-14 17:38

🖼️（附图，图片本身未缓存）

▸ **Bash** Diagnose mobile overflow

````
cat > 〔路径00211〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  for(const f of ['ch03c-production-raft','ch03d-faq']){
    const p=await b.newPage({viewport:{width:400,height:900}});
    await p.goto('file://〔路径00070〕'+f+'.html');
    await p.waitForTimeout(300);
    const r=await p.evaluate(()=>{
      const W=document.documentElement.clientWidth, out=[];
      document.querySelectorAll('main *').forEach(e=>{
        const b=e.getBoundingClientRect();
        if(b.right>W+2) out.push((e.tagName+'.'+(e.className||'')).slice(0,50)+' right='+Math.round(b.right)+' w='+Math.round(b.width));
      });
      return {sw:document.docume
…（截断，原长 974 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch03c-production-raft {
 "sw": 400,
 "cw": 400,
 "out": [
  "svg.[object SVGAnimatedString] right=744 w=680",
  "rect.[object SVGAnimatedString] right=476 w=108",
  "text.[object SVGAnimatedString] right=431 w=18",
  "text.[object SVGAnimatedString] right=445 w=46",
  "text.[object SVGAnimatedString] right=438 w=34",
  "rect.[object SVGAnimatedString] right=595 w=108",
  "text.[object SVGAnimatedString] right=550 w=18",
  "text.[object SVGAnimatedString] right=564 w=46",
  "text.[object SVGAnimatedString] right=558 w=34",
  "rect.[object SVGAnimatedString] right=715 w=108",
  "text.[object SVGAnimatedString] right=670 w=18",
  "text.[object SVGAnimatedString] right=684 w=46"
 ]
}
ch03d-faq {
 "sw": 412,
 "cw": 400,
 "out": [
  "TABLE. right=412 w=384",
  "THEAD. right=412 w=384",
  "TR. right=412 w=384",
  "TH. right=412 w=72",
  "TBODY. right=412 w=384",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384"
 ]
}
````

</details>

▸ **Bash** Add table scroll wrapper and recheck

````
cd 〔路径00069〕 && python3 - <<'PY'
p='ch03d-faq.html'
s=open(p,encoding='utf-8').read()
helper='''function wrapT(root){(root||document).querySelectorAll("table").forEach(t=>{
  const pa=t.parentElement; if(pa&&pa.classList.contains("grid-wrap"))return;
  const d=document.createElement("div"); d.className="grid-wrap";
  pa.insertBefore(d,t); d.appendChild(t);});}
'''
anchor='document.addEventListener("scroll",onScroll,{passive:true});onScroll();\n'
assert s.count(anchor)==1
s=s.replace(anchor, anchor+helper)
# call after each render
reps=[
 ('  if(S1.last)$("#o1").innerHTML=S1.last; \n}', '  if(S1.last)$("#o1").innerHTML=S1.last;\n  wrapT($("#o1"));\n}'),
 ('  $("#b2step").disabled = i>=st.length-1;\n}', '  $("#b2step").disabled = i>=st.length-1;\n  wrapT($("#o2"));\n}'),
 (
…（截断，原长 1312 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
 ]
}
ch03d-faq {
 "sw": 400,
 "cw": 400,
 "out": [
  "TABLE. right=412 w=384",
  "THEAD. right=412 w=384",
  "TR. right=412 w=384",
  "TH. right=412 w=72",
  "TBODY. right=412 w=384",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384",
  "TD. right=412 w=72",
  "TR. right=412 w=384"
 ]
}
````

</details>

▸ **Bash** Re-verify after table wrapping

````
cd 〔路径00052〕 && node 〔路径00254〕 2>&1 | tail -25
````

<details><summary>↳ Bash 结果</summary>

````
D1 indexes: lastLogIndex = 5commitIndex = 3lastApplied = 1matchIndex[]N2=3 N3=3 N4=3 N5=3 （过半达到 3）
D1 read(no step3): 丢了 2 条已提交的写
D1 read(step3): 返回 x = 3 | 线性一致: true
D2 no-noop final: 客户端 ReadIndex 读 x① readIndex := commitIndex = 0 ② 心跳过半确认 ✓ 通过 ③ lastApplied=0 >= 0 ✓ 通过返回 （初始值，x 不存在）可客户端明明收到过「x=3 写入成功」
D2 with-noop final: 客户端 ReadIndex 读 x① readIndex := commitIndex = 4 ② 心跳过半确认 ✓ ③ lastApplied=4 >= 4 ✓返回 x = 3 —— 正是客户端上次写成功的那个值。线性一致 ✓
D3 dynamo divergence over 12 seeds: 2,2,2,2,2,2,2,2,2,2,2,2
D3 with W=N=5, C=4: 3,2,3,2,4,3,4,3
D4 scenario A: clusters=1 leaders=1 effective=3
D4 scenario B: clusters=2 leaders=2 effective=3
D4 scenario C: clusters=2 leaders=1 effective=2
D4 scenario D: clusters=2 leaders=1 effective=2
D4 scenario E: clusters=1 leaders=1 effective=3
quiz items: 10
sections: 10
NO CONSOLE ERRORS
````

</details>

▸ **Bash** Inspect lab03c and tui API

````
cd 〔路径00079〕 && ls -la lab03c/ && wc -l lab03c/*.go internal/tui/tui.go && echo "=== tui api ===" && grep -n "^func " internal/tui/tui.go
````

<details><summary>↳ Bash 结果</summary>

````
total 60
drwxr-xr-x 2 root root  4096 Sep 10 05:04 .
drwxr-xr-x 9 root root  4096 Sep 10 04:59 ..
-rw-r--r-- 1 root root 18474 Sep 10 05:04 main.go
-rw-r--r-- 1 root root  4406 Sep 10 05:04 net.go
-rw-r--r-- 1 root root 23920 Sep 10 05:04 raft.go
  444 lab03c/main.go
  149 lab03c/net.go
  786 lab03c/raft.go
  101 internal/tui/tui.go
 1480 total
=== tui api ===
11:func Dispw(s string) int {
29:func PadR(s string, w int) string {
36:func PadL(s string, w int) string {
44:func Head(part, n int, title, subtitle string) {
51:func Row(k string, v ...any) {
56:func TableHead(cols []string, w []int) { TableHeadA(cols, w, defaultAlign(len(cols))) }
57:func TableRow(cells []string, w []int) { TableRowA(cells, w, defaultAlign(len(cells))) }
59:func defaultAlign(n int) string {
67:func TableHeadA(cols []string, w []int, align string) {
76:func TableRowA(cells []string, w []int, align string) { fmt.Println(layout(cells, w, align)) }
78:func layout(cells []string, w []int, align string) string {
94:func Bar(n int) string { return strings.Repeat("═", n) }
97:func HeadN(label, title, subtitle string) {
````

</details>

▸ **Bash** Read lab03c main head and tui helpers

````
sed -n '1,80p' lab03c/main.go && echo "...." && sed -n '40,101p' internal/tui/tui.go
````

<details><summary>↳ Bash 结果</summary>

````
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
…（截断，原长 3619 字符）
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #22 status
````

</details>

▸ **TaskUpdate** in_progress

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #23 status
````

</details>

▸ **Bash** Write Go lab03d model

````
mkdir -p 〔路径00111〕 && cat > 〔路径00113〕 <<'GOEOF'
// Lab 3-D 的四个模型。
//
// 这一章和 3-A/3-B/3-C 不同：那三章要跑完整的 Raft 状态机，
// 因为它们研究的是【算法本身】会不会出错。
// 3-D 研究的是【算法与系统之间的接缝】——
// 读路径怎么和状态机对齐、新 Leader 的 commitIndex 从哪来、
// quorum 和日志复制到底差在哪、配置从哪来。
// 这些问题用一个聚焦的小模型讲得更清楚，所以这里不复用 lab03c 的 raft.go。
package main

import "sort"

// ═══════════════════════════════════════════════════════════════════════
// 模型 1：ReadIndex 的第三步
// ═══════════════════════════════════════════════════════════════════════

// ReadResult 记录一次读的结果。
type ReadResult struct {
	Got      int // 读到的值（= 该 index 处写入的值；0 表示还没有值）
	Want     int // 此刻已提交的最新值
	Waited   int // 第三步阻塞等了多少条 apply
	Stale    bool
}

// LeaderState 是一个被极度简化的 Leader：
// 我们只保留这条读路径真正依赖的三个量。
type LeaderState struct {
	commitIndex int
…（截断，原长 7027 字符）
````

<details><summary>↳ Bash 结果</summary>

````
done
````

</details>

▸ **Bash** Write Go rand helper

````
cat > 〔路径00114〕 <<'GOEOF'
package main

import "strconv"

// Rand 是一个最简单的 xorshift32。
//
// 为什么不用 math/rand：Java 版要产出【逐字节相同】的输出，
// 两边各自的标准库随机数是对不上的。自己写 8 行，两边一模一样。
type Rand struct{ s uint32 }

func NewRand(seed int64) *Rand {
	v := uint32(seed)
	if v == 0 {
		v = 1
	}
	return &Rand{s: v}
}

func (r *Rand) Next() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}

func (r *Rand) Float() float64    { return float64(r.Next()) / 4294967296.0 }
func (r *Rand) Intn(n int) int    { return int(r.Next() % uint32(n)) }
func (r *Rand) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		swap(i, r.Intn(i+1))
	}
}

func join(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x

…（截断，原长 881 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Build and run Go lab03d

````
cat > 〔路径00112〕 <<'GOEOF'
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
	readsF    = flag.Int("reads", 200, "3D-1：每种配置读多少次")
	seedsF    = flag.Int("seeds", 40, "3D-2 / 3D-3：跑多少个种子")
	writersF  = flag.Int("writers", 2, "3D-3：并发写的客户端数")
	nF        = flag.Int("n", 5, "3D-3：副本数 N")
	wF        = flag.Int("w", 3, "3D-3：写 quorum W")
	rF        = flag.Int("r", 3, "3D-3：读 quorum R")
	lossF     = flag.Float64("loss", 0.22, "3D-3：丢包率")
	seedF     = flag.Int64("seed", 42, "随机种子基准")
	entriesF  = flag.Int("entries", 3, "3D-
…（截断，原长 11344 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：N=5 W=3 R=3 ｜ applylag=4 ｜ 种子基准 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-1 · ReadIndex 的第三步
  前两步全部通过，少了第三步会读到什么
══════════════════════════════════════════════════════════════════════════════

    模型：一个没有分区的 Leader。它一直合法，心跳永远能收到过半响应
    ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。
    apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
    ────────────────────────────────────────────────────────────────────────────────
    0 条以内                  0/200              0%            0/200         0.00 条
    1 条以内                 99/200             50%            0/200         0.49 条
    2 条以内                147/200             74%            0/200         1.10 条
    4 条以内      ◀         154/200             77%            0/200         1.87 条
    6 条以内                176/200             88%            0/200         2.93 条
    8 条以内                169/200             84%            0/200         4.04 条

    结论
      · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
      · 执行第 3 步的陈旧读恒为 0，代价是平均多等几条 apply。
      · 生产上这条曲线对应的现象就是：写接口返回成功，立刻查却查不到。
        滞后被撑大的典型原因是状态机执行慢（大事务 / 写放大 / compaction 抢 IO）。
      · 再说一遍：这一整张表里，ReadIndex 的第 1、2 步【全部通过】。
        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-2 · no-op 条目的必要性
  新 Leader 不发空日志，ReadIndex 三步走完照样读到旧值
═══════════════════════════════════════
…（截断，原长 3395 字符）
````

</details>

▸ **Bash** Patch and rerun Go lab

````
python3 - <<'PY'
import re
# --- rand.go: better Intn ---
p='rand.go'; s=open(p,encoding='utf-8').read()
s=s.replace('func (r *Rand) Intn(n int) int    { return int(r.Next() % uint32(n)) }',
            'func (r *Rand) Intn(n int) int    { return int(r.Float() * float64(n)) } // 用高位，避免低位周期性')
open(p,'w',encoding='utf-8').write(s)

# --- model.go: fix RunNoop classification ---
p='model.go'; s=open(p,encoding='utf-8').read()
old='''const (
	OutcomeCorrect  NoopOutcome = iota // 读到了正确的值
	OutcomeStale                       // 读到旧值，而客户端收到过「写入成功」⇒ 线性一致性破裂
	OutcomeRejected                    // Leader 明确拒绝服务（no-op 还没提交）⇒ 安全
	OutcomeNoAck                       // 客户端本来就没收到过成功 ⇒ 返回旧值不算违规
)'''
new='''const (
	OutcomeFresh    NoopOutcome = iota // 返回了最新的已提交值
	OutcomeStale                       // 读到
…（截断，原长 5082 字符）
````

<details><summary>↳ Bash 结果</summary>

````
patched
main.go
model.go
rand.go
    ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。
    apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
    ────────────────────────────────────────────────────────────────────────────────
    0 条以内                  0/200              0%            0/200         0.00 条
    1 条以内                105/200             52%            0/200         0.53 条
    2 条以内                130/200             65%            0/200         0.96 条
    4 条以内      ◀         156/200             78%            0/200         2.06 条
    6 条以内                170/200             85%            0/200         2.84 条
    8 条以内                167/200             84%            0/200         4.03 条

    结论
      · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
      · 执行第 3 步的陈旧读恒为 0，代价是平均多等几条 apply。
      · 生产上这条曲线对应的现象就是：写接口返回成功，立刻查却查不到。
        滞后被撑大的典型原因是状态机执行慢（大事务 / 写放大 / compaction 抢 IO）。
      · 再说一遍：这一整张表里，ReadIndex 的第 1、2 步【全部通过】。
        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-2 · no-op 条目的必要性
  新 Leader 不发空日志，ReadIndex 三步走完照样读到旧值
══════════════════════════════════════════════════════════════════════════════

    场景（每个种子随机决定两个分支）：
      任期 2：Leader 把 3 条写复制到了过半节点，然后崩溃。
              分支 ①（约一半概率）它崩之前【已经推进 commitIndex 并回复客户端成功】；
              分支 ②                它还没来得及推进 commitIndex。
              两种情况下它都没来得及广播 leaderComm
…（截断，原长 1930 字符）
````

</details>

▸ **Bash** Format and view rest of output

````
gofmt -w . && sed -i 's/readsF    = flag.Int("reads", 200,/readsF    = flag.Int("reads", 500,/' main.go && go run . 2>&1 | sed -n '45,140p'
````

<details><summary>↳ Bash 结果</summary>

````
    不发 no-op                            16               0             0                    24
    发 no-op（正确）                       0              25            15                     0

    结论
      · 不发 no-op：16/40 个种子读到了旧值，而客户端【收到过写入成功】——
        这是货真价实的线性一致性破裂，且 ReadIndex 三步一步没少、全部通过。
        问题出在第 1 步拿到的 commitIndex 本身就是错的。
      · 发 no-op：违规 0 次。no-op 是【当前任期】的条目，可以直接提交；
        提交是前缀性的 ⇒ 任期 2 那几条被顺带提交 ⇒ commitIndex 到位。
      · 那 15 次「正确拒绝」也很重要：no-op 还没提交完时，
        真实系统（etcd / TiKV）会拒绝服务读，而不是拿一个没到位的 commitIndex 去凑合。
        【给你错误，而不是给你错的数据。】
      · 「旧值但未 ack」的那些不算违规：那次写从未收到成功响应，
        它的结果本来就是 Part 0 说的「第三态」——客户端必须自己去查证。

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-3 · Quorum 读写 ≠ Raft 日志复制
  同一组并发写灌进两个世界，数副本间的分歧
══════════════════════════════════════════════════════════════════════════════

    2 个客户端【同时】写同一个 key，N=5，丢包率 22%。
      ① Dynamo 式：每个写发往全部副本、收到 W 个 ack 即成功，副本不带版本信息。
      ② Raft    ：全部请求先到 Leader，Leader 按到达先后分配 index 再复制。
    写 W           Dynamo 平均分歧    └ 出现分歧的种子       Raft 平均分歧      └ 出现分歧
    ──────────────────────────────────────────────────────────────────────────────────────
    1                      1.88 种               35/40             1.00 种            0/40
    2                      1.88 种               35/40             1.00 种            0/40
    3 ◀                    1.88 种               35/40             1.00 种            0/40
    4                      1.88 种               35/40             1.0
…（截断，原长 4529 字符）
````

</details>

▸ **Bash** Refine Go lab output

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()

# --- 3D-3: show WriteOK so W's real effect is visible ---
s=s.replace('''	w := []int{8, 4, 18, 20, 4, 16, 16}
	al := "LLRRLRR"
	tui.TableHeadA([]string{"写 W", "", "Dynamo 平均分歧", "└ 出现分歧的种子", "", "Raft 平均分歧", "└ 出现分歧"}, w, al)''',
'''	w := []int{8, 2, 16, 16, 18, 3, 14, 14}
	al := "LLRRRLRR"
	tui.TableHeadA([]string{"写 W", "", "Dynamo 写成功", "└ 平均分歧", "└ 出现分歧的种子",
		"", "Raft 写成功", "└ 平均分歧"}, w, al)''')
s=s.replace('''		sumD, seedsWithD := 0, 0
		sumR, seedsWithR := 0, 0''','''		sumD, seedsWithD, okD := 0, 0, 0
		sumR, seedsWithR, okR := 0, 0, 0''')
s=s.replace('''			sumD += d.Distinct
			if d.Distinct > 1 {
				seedsWithD++
			}''','''			sumD += d.Distinct
			okD += d.WriteOK
			if d.Distinct > 1 {
				seedsWithD++
			}''')
s
…（截断，原长 3324 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
    ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。
    apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
    ────────────────────────────────────────────────────────────────────────────────
    0 条以内                  0/500              0%            0/500         0.00 条
    1 条以内                264/500             53%            0/500         0.53 条
    2 条以内                326/500             65%            0/500         0.95 条
    4 条以内      ◀         387/500             77%            0/500         1.92 条
    6 条以内                418/500             84%            0/500         2.82 条
    8 条以内                432/500             86%            0/500         3.93 条

    结论
      · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
    2 个客户端【同时】写同一个 key，N=5，丢包率 22%。
      ① Dynamo 式：每个写发往全部副本、收到 W 个 ack 即成功，副本不带版本信息。
      ② Raft    ：全部请求先到 Leader，Leader 按到达先后分配 index 再复制。
    写 W         Dynamo 写成功      └ 平均分歧  └ 出现分歧的种子      Raft 写成功    └ 平均分歧
    ───────────────────────────────────────────────────────────────────────────────────────────
    1                   2.00/2         1.88 种             35/40           2.00/2       1.00 种
    2                   1.93/2         1.88 种             35/40           2.00/2       1.00 种
    3 ◀                 1.85/2         1.88 种             35/40           2.00/2       1.00 种
    4                   1.50/2         1.88 种             35/40           2.00/2       1.00 种
    5
…（截断，原长 2299 字符）
````

</details>

▸ **Bash** Check 3D-4 summary table

````
go run . 2>&1 | sed -n '96,106p'
````

<details><summary>↳ Bash 结果</summary>

````
  初始配置只能从外部给——给错了会静默长出第二个集群
══════════════════════════════════════════════════════════════════════════════
    场景                              集群数   能服务的 Leader    有效副本    危险类型
    ────────────────────────────────────────────────────────────────────────────────────────────────
    A · 正确的静态引导                     1                 1         3/3    —
    B · 一个节点列表写错                   2                 2         3/3    脑裂：数据分叉
    C · token 不同                         2                 1         2/3    降级：容错能力归零
    D · 灾难恢复用了 state=new             2                 1         2/3    静默丢数据（新集群顶替）
    E · 单节点引导 + member add            1                 1         3/3    —

    A · 正确的静态引导
````

</details>

▸ **Bash** Inspect Java lab structure

````
cd 〔路径00136〕 && head -60 Lab03C.java && echo "=== helpers ===" && grep -n "static .*\(dispw\|padR\|padL\|headN\|tableHead\|tableRow\|row(\)" Lab03C.java | head -20
````

<details><summary>↳ Bash 结果</summary>

````
/*
 * Lab 3-C · 持久化、快照、成员变更与线性一致读（Java 版）
 *
 * 运行：  cd java/lab03c && java Lab03C.java
 * 调参：  java Lab03C.java -lose votedfor | term | log
 *
 * 需要 JDK 17+。配套课件：courseware/ch03c-production-raft.html
 */
import java.util.*;

public class Lab03C {

    static int N = 5;
    static long DELAY = 20, JITTER = 10, SEED = 42;
    static double LOSS = 0.0;
    static boolean VERBOSE = false;
    static String LOSE = "";
    static int TRIALS = 30;
    static final long TICK = 5, HEARTBEAT = 100;
    static final int NO_VOTE = -1;

    // ══════════════════════════════════════════════════════════════════════
    // 确定性网络
    // ══════════════════════════════════════════════════════════════════════
    enum MType { REQUEST_VOTE, VOTE_REPLY, APPEND, APPEND_REPLY, SNAPSHOT, SNAPSHOT_REPLY }

    record Entry(int index, int term, String cmd) {}

    static class Msg {
        int from, to; MType type; int term;
        int lastLogIndex, lastLogTerm; boolean granted;                 // RequestVote
        int prevLogIndex, prevLogTerm; List<Entry> entries = List.of(); // AppendEntries
        int leaderCommit;
        boolean success; int matchIndex, conflictTerm, conflictIndex;   // 回复
        int lastIncludedIndex, lastIncludedTerm;                        // InstallSnapshot
        Map<String,String> snapshot = Map.of();
        long deliverAt; int seq;
        Msg(int f, int t, MType ty, int tm) { from=f; to=t; type=ty; term=tm; }
    }

    static class Network {
        final int n;
…（截断，原长 2902 字符）
````

</details>

▸ **Bash** Read Java TUI helpers and main

````
sed -n '760,810p' Lab03C.java; echo "=== main/UTF8 ==="; grep -n "setOut\|public static void main" Lab03C.java
````

<details><summary>↳ Bash 结果</summary>

````
  Part 3 到此结束。下一站：Part 4 · 分布式事务。""");
        System.out.println("═".repeat(78) + "\n");
    }


    // ─── 输出工具 ─────────────────────────────────────────────────────────
    static int dispw(String s) {
        int w=0;
        for (int i=0;i<s.length();i++) {
            char c=s.charAt(i);
            w += (c>=0x1100 && (c<=0x115F || (c>=0x2E80&&c<=0xA4CF&&c!=0x303F)
                    || (c>=0xAC00&&c<=0xD7A3) || (c>=0xF900&&c<=0xFAFF)
                    || (c>=0xFE30&&c<=0xFE6F) || (c>=0xFF00&&c<=0xFF60)
                    || (c>=0xFFE0&&c<=0xFFE6))) ? 2 : 1;
        }
        return w;
    }
    static String padR(String s,int w){int d=w-dispw(s);return d>0?s+" ".repeat(d):s;}
    static String padL(String s,int w){int d=w-dispw(s);return d>0?" ".repeat(d)+s:s;}
    static String layout(String[] c,int[] w,String al){
        StringBuilder sb=new StringBuilder("    ");
        for(int i=0;i<c.length;i++) sb.append(i<al.length()&&al.charAt(i)=='L'?padR(c[i],w[i]):padL(c[i],w[i]));
        return sb.toString().stripTrailing();
    }
    static void head(String label,String title,String sub){
        System.out.println("\n"+"═".repeat(78));
        System.out.printf("  Lab %s · %s%n  %s%n", label, title, sub);
        System.out.println("═".repeat(78));
    }
    static void row(String k,Object v){ System.out.println("    "+padR(k,30)+" "+v); }
    static void tableHead(String[] cols,int[] w,String al){
        System.out.println(layout(cols,w,al));
        int t=0; for
…（截断，原长 2502 字符）
````

</details>

▸ **Bash** Read Java main preamble

````
sed -n '514,525p' Lab03C.java
````

<details><summary>↳ Bash 结果</summary>

````
    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 每组 %d 个种子 ｜ 种子基准 %d%n",
                N, DELAY, JITTER, TRIALS, SEED);
        lab1(); lab2(); lab3(); lab4(); epilogue();
    }

    // ─── Lab 3C-1 ─────────────────────────────────────────────────────────
    static void lab1() {
        head("3C-1", "持久化与崩溃重启", "分别丢掉三样状态中的一样，看哪条安全性倒下");
````

</details>

▸ **Bash** Write and run Java lab03d

````
mkdir -p 〔路径00138〕 && cat > 〔路径00139〕 <<'JEOF'
/*
 * Lab 3-D · Raft 附录：索引、读、配置与日志（Java 版）
 *
 * 运行：  cd java/lab03d && java Lab03D.java
 * 调参：  java Lab03D.java -applylag 8 -reads 800 -seeds 60 -writers 4
 *
 * 需要 JDK 17+。配套课件：courseware/ch03d-faq.html
 *
 * 这一章和 3-A/3-B/3-C 不同：那三章要跑完整的 Raft 状态机，
 * 因为它们研究的是【算法本身】会不会出错。
 * 3-D 研究的是【算法与系统之间的接缝】——
 * 读路径怎么和状态机对齐、新 Leader 的 commitIndex 从哪来、
 * quorum 和日志复制到底差在哪、配置从哪来。
 * 这些问题用一个聚焦的小模型讲得更清楚。
 *
 * 输出与 Go 版逐字对应（连随机数都用同一个 xorshift32），可以直接对照阅读。
 */
import java.util.*;

public class Lab03D {

    static int APPLYLAG = 4, READS = 500, SEEDS = 40, WRITERS = 2;
    static int N = 5, W = 3, R = 3, ENTRIES = 3;
    static double LOSS = 0.22;
    static long SEED = 42;

    public static void main(String
…（截断，原长 22304 字符）
````

<details><summary>↳ Bash 结果</summary>

````
Picked up JAVA_TOOL_OPTIONS: -Djavax.net.ssl.trustStore=/root/.ccr/java-truststore.p12 -Djavax.net.ssl.trustStorePassword=changeit -Djavax.net.ssl.trustStoreType=PKCS12 -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=33367 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|127.*|0.*|::|169.254.*|api.anthropic.com|api-staging.anthropic.com|api-pr-preview.anthropic.com|mcp-proxy.anthropic.com|mcp-proxy-staging.anthropic.com|registry.npmjs.org|jsr.io|npm.jsr.io|pypi.org|files.pythonhosted.org|index.crates.io|proxy.golang.org|host.docker.internal|10.*|172.16.*|172.17.*|172.18.*|172.19.*|172.20.*|172.21.*|172.22.*|172.23.*|172.24.*|172.25.*|172.26.*|172.27.*|172.28.*|172.29.*|172.30.*|172.31.*|192.168.*|100.64.0.0/10|*.svc.cluster.local|*.svc.cluster.local -Djdk.http.auth.tunneling.disabledSchemes= -Djdk.http.auth.proxying.disabledSchemes=
Lab03D.java:431: error: illegal text block open delimiter sequence, missing line terminator
        System.out.println("""    ??
                                  ^
Lab03D.java:432: error: illegal character: '\u00b7'
      ? ??????????????????????? --initial-cluster ??
      ^
Lab03D.java:432: error: illegal character: '\u300c'
      ? ??????????????????????? --initial-cluster ??
        ^
Lab03D.java:432: error: illegal character: '\u300d'
      ? ??????????????????????? --initial-cluster ??
              ^
Lab03D.java:432: error: not a statement
      ? ??????????????????????? --initial-cluster ??
         ^
Lab03D.java:432: error: illegal character:
…（截断，原长 2022 字符）
````

</details>

▸ **Bash** Fix text blocks and rerun

````
python3 - <<'PY'
p='Lab03D.java'; s=open(p,encoding='utf-8').read()
s=s.replace('System.out.println("""    结论\n','System.out.println("""\n    结论\n')
s=s.replace('System.out.println("""  Part 3 到此真正结束\n','System.out.println("""\n  Part 3 到此真正结束\n')
open(p,'w',encoding='utf-8').write(s)
print('ok')
PY
java Lab03D.java 2>&1 | grep -v JAVA_TOOL_OPTIONS | head -45
````

<details><summary>↳ Bash 结果</summary>

````
ok

配置：N=5 W=3 R=3 ｜ applylag=4 ｜ 种子基准 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-1 · ReadIndex 的第三步
  前两步全部通过，少了第三步会读到什么
══════════════════════════════════════════════════════════════════════════════

模型：一个没有分区的 Leader。它一直合法，心跳永远能收到过半响应
——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。
    apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
    ────────────────────────────────────────────────────────────────────────────────
    0 条以内                  0/500              0%            0/500         0.00 条
    1 条以内                264/500             53%            0/500         0.53 条
    2 条以内                326/500             65%            0/500         0.95 条
    4 条以内      ◀         387/500             77%            0/500         1.92 条
    6 条以内                418/500             84%            0/500         2.82 条
    8 条以内                432/500             86%            0/500         3.93 条

结论
  · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
  · 执行第 3 步的陈旧读恒为 0，代价是平均多等几条 apply。
  · 生产上这条曲线对应的现象就是：写接口返回成功，立刻查却查不到。
    滞后被撑大的典型原因是状态机执行慢（大事务 / 写放大 / compaction 抢 IO）。
  · 再说一遍：这一整张表里，ReadIndex 的第 1、2 步【全部通过】。
    确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-2 · no-op 条目的必要性
  新 Leader 不发空日志，ReadIndex 三步走完照样读到旧值
═══════════════════════════════════════════════════════════════════════════
…（截断，原长 2054 字符）
````

</details>

▸ **Bash** Fix Java indentation and diff against Go

````
python3 - <<'PY'
p='Lab03D.java'; s=open(p,encoding='utf-8').read()
subs=[
 ('        System.out.println("""\n\n    模型：','        System.out.print("""\n\n    模型：'),
 ('    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。""");',
  '    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。""".indent(4));'),
 ('        System.out.println("""\n\n    结论\n      · 跳过第 3 步','        System.out.print("""\n\n    结论\n      · 跳过第 3 步'),
 ('        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。""");',
  '        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。""".indent(4));'),
 ('        System.out.println("""\n\n    结论\n      · 看第一列和第三列的对比','        System.out.print("""\n\n    结论\n      · 看第一列和第三列的对比'),
 ('        是【我还是不是 Leader】这一个 bit。""");','        是【我还是不是 Leader】这一个 bit。""".indent(4));'),
 ('        System.out.println("
…（截断，原长 1664 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
42a43
>
59a61
>
67a70
>
--- diffs above ---
````

</details>

▸ **Bash** Locate extra blank lines

````
sed -n '40,46p;57,63p;66,72p' 〔路径00210〕 | cat -A | sed 's/\$$/[EOL]/' | head -30; echo "=== go ==="; sed -n '40,45p;57,62p;65,70p' 〔路径00206〕
````

<details><summary>↳ Bash 结果</summary>

````
      M-dM-;M-;M-fM-^\M-^_ 3M-oM-<M-^ZM-fM-^_M-^PM-dM-8M-*M-fM-^LM-^AM-fM-^\M-^IM-hM-?M-^YM-dM-:M-^[M-fM-^WM-%M-eM-?M-^WM-gM-^ZM-^DM-hM-^JM-^BM-gM-^BM-9M-eM-=M-^SM-iM-^@M-^IM-oM-<M-^LM-eM-.M-^CM-gM-^ZM-^D commitIndex = 0M-cM-^@M-^B[EOL]
              M-eM-^OM-^W Figure 8 M-iM-^YM-^PM-eM-^HM-6M-oM-<M-^LM-eM-.M-^CM-cM-^@M-^PM-dM-8M-^MM-hM-^CM-=M-cM-^@M-^QM-gM-^[M-4M-fM-^NM-%M-fM-^OM-^PM-dM-:M-$M-dM-;M-;M-fM-^\M-^_ 2 M-gM-^ZM-^DM-fM-^]M-!M-gM-^[M-.M-cM-^@M-^B[EOL]
      M-gM-^DM-6M-eM-^PM-^NM-eM-.M-"M-fM-^HM-7M-gM-+M-/M-gM-^TM-(M-eM-.M-^LM-fM-^UM-4M-gM-^ZM-^D ReadIndexM-oM-<M-^HM-dM-8M-^IM-fM--M-%M-dM-8M-^@M-fM--M-%M-dM-8M-^MM-eM-0M-^QM-oM-<M-^IM-fM-^]M-%M-hM-/M-;M-cM-^@M-^B[EOL]
[EOL]
    M-fM-^VM-0 Leader M-dM-8M-^JM-dM-;M-;M-eM-^PM-^N          M-hM-/M-;M-eM-^HM-0M-fM-^WM-'M-eM-^@M-<(M-hM-?M-^]M-hM-'M-^D)      M-hM-?M-^TM-eM-^[M-^^M-fM-^\M-^@M-fM-^VM-0M-eM-^@M-<      M-fM--M-#M-gM-!M-.M-fM-^KM-^RM-gM-;M-^]     M-fM-^WM-'M-eM-^@M-<M-dM-=M-^FM-fM-^\M-*ack(M-eM-^EM-^AM-hM-.M-8)[EOL]
    M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^TM-^@M-bM-^T
…（截断，原长 7047 字符）
````

</details>

▸ **Bash** Fix trailing newlines and diff

````
cd 〔路径00138〕 && grep -n '%n"""' Lab03D.java && python3 - <<'PY'
p='Lab03D.java'; s=open(p,encoding='utf-8').read()
n=s.count('\n%n"""')
s=s.replace('\n%n"""','\n"""')
open(p,'w',encoding='utf-8').write(s)
print("replaced",n)
PY
java Lab03D.java > 〔路径00210〕 2>/dev/null; diff 〔路径00206〕 〔路径00210〕 && echo "=== IDENTICAL OUTPUT (Go == Java) ===" || echo "--- still differs ---"
````

<details><summary>↳ Bash 结果</summary>

````
174:%n""", ENTRIES);
197:%n""", a[0], SEEDS, b[2]);
277:%n""", WRITERS, N, LOSS * 100);
replaced 3
8c8
<
---
>
20c20
<
---
>
79c79
<
---
>
145c145
<
---
>
148c148
<
---
>
152c152
<
---
>
--- still differs ---
````

</details>

▸ **Bash** Add blk helper and re-diff

````
python3 - <<'PY'
import re
p='Lab03D.java'; s=open(p,encoding='utf-8').read()
s=re.sub(r'"""\.indent\((\d+)\)\)', r'"""), \1)', s)
s=s.replace('System.out.print("""','System.out.print(blk("""')
# add helper
s=s.replace('    static String pct(int a, int b)',
 '''    /** 缩进一段文本；空行不补空格（否则会留下看不见的行尾空白）。 */
    static String blk(String s, int k) { return s.indent(k).replaceAll("(?m)^[ ]+$", ""); }
    static String pct(int a, int b)''')
open(p,'w',encoding='utf-8').write(s)
PY
grep -n 'blk("""' Lab03D.java | head; grep -n '"""), [0-9])' Lab03D.java | head
java Lab03D.java > 〔路径00210〕 2>&1; diff 〔路径00206〕 〔路径00210〕 && echo "=== IDENTICAL OUTPUT (Go == Java) ===" || echo "--- differs ---"
````

<details><summary>↳ Bash 结果</summary>

````
88:        System.out.print(blk("""
114:        System.out.print(blk("""
309:        System.out.print(blk("""
431:        System.out.print(blk("""
443:        System.out.print(blk("""
92:    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。"""), 4);
122:        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。"""), 4);
323:        是【我还是不是 Leader】这一个 bit。"""), 4);
438:        【改配置文件不会改变集群成员】——只能走 member add / remove / update。"""), 4);
455:  然后才进分布式：2PC/XA、TCC、Saga、事务消息、Percolator、Calvin、Spanner。"""), 2);
1,157c1,33
<
< 配置：N=5 W=3 R=3 ｜ applylag=4 ｜ 种子基准 42
<
< ══════════════════════════════════════════════════════════════════════════════
<   Lab 3D-1 · ReadIndex 的第三步
<   前两步全部通过，少了第三步会读到什么
< ══════════════════════════════════════════════════════════════════════════════
<
<     模型：一个没有分区的 Leader。它一直合法，心跳永远能收到过半响应
<     ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
<     唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。
<     apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
<     ────────────────────────────────────────────────────────────────────────────────
<     0 条以内                  0/500              0%            0/500         0.00 条
<     1 条以内                264/500             53%            0/500         0.53 条
<     2 条以内                326/500             65%            0/500         0.95 条
<     4 条以内      ◀         387/500             77%            0/500         1.92 条
<     6 条以内                418/500             84%            0/500         2.82 条
<     8 条以内
…（截断，原长 10195 字符）
````

</details>

▸ **Bash** Fix paren placement and diff

````
python3 - <<'PY'
import re
p='Lab03D.java'; s=open(p,encoding='utf-8').read()
s=re.sub(r'"""\), (\d+)\);', r'""", \1));', s)
open(p,'w',encoding='utf-8').write(s)
PY
grep -n '""", [0-9]));' Lab03D.java
java Lab03D.java > 〔路径00210〕 2>/dev/null; diff 〔路径00206〕 〔路径00210〕 && echo "=== IDENTICAL OUTPUT (Go == Java) ===" || echo "--- differs ---"
````

<details><summary>↳ Bash 结果</summary>

````
92:    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。""", 4));
122:        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。""", 4));
323:        是【我还是不是 Leader】这一个 bit。""", 4));
438:        【改配置文件不会改变集群成员】——只能走 member add / remove / update。""", 4));
455:  然后才进分布式：2PC/XA、TCC、Saga、事务消息、Percolator、Calvin、Spanner。""", 2));
=== IDENTICAL OUTPUT (Go == Java) ===
````

</details>

▸ **Bash** Test flags in both languages

````
cd 〔路径00111〕 && go run . -applylag 8 -reads 300 -seeds 20 -writers 4 -w 4 > 〔路径00203〕 2>&1 && cd 〔路径00138〕 && java Lab03D.java -applylag 8 -reads 300 -seeds 20 -writers 4 -w 4 > 〔路径00207〕 2>/dev/null; diff 〔路径00203〕 〔路径00207〕 && echo "FLAGS: IDENTICAL" || echo "FLAGS DIFFER"; echo "--- sample ---"; sed -n '12,20p;63,74p' 〔路径00203〕
````

<details><summary>↳ Bash 结果</summary>

````
FLAGS: IDENTICAL
--- sample ---
    apply 滞后            跳过第3步        └ 陈旧读        执行第3步      └ 平均等待
    ────────────────────────────────────────────────────────────────────────────────
    0 条以内                  0/300              0%            0/300         0.00 条
    1 条以内                158/300             53%            0/300         0.53 条
    2 条以内                195/300             65%            0/300         0.96 条
    4 条以内                228/300             76%            0/300         1.91 条
    6 条以内                249/300             83%            0/300         2.84 条
    8 条以内      ◀         255/300             85%            0/300         3.87 条
    10 条以内               276/300             92%            0/300         4.75 条

══════════════════════════════════════════════════════════════════════════════
  Lab 3D-3 · Quorum 读写 ≠ Raft 日志复制
  同一组并发写灌进两个世界，数副本间的分歧
══════════════════════════════════════════════════════════════════════════════

    4 个客户端【同时】写同一个 key，N=5，丢包率 22%。
      ① Dynamo 式：每个写发往全部副本、收到 W 个 ack 即成功，副本不带版本信息。
      ② Raft    ：全部请求先到 Leader，Leader 按到达先后分配 index 再复制。
    写 W         Dynamo 写成功      └ 平均分歧  └ 出现分歧的种子      Raft 写成功    └ 平均分歧
    ───────────────────────────────────────────────────────────────────────────────────────────
    1                   4.00/4         3.05 种             20/20           4.00/4       1.00 种
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #23 status
````

</details>

▸ **Bash** Find README sections to update

````
cd 〔路径00054〕 && grep -n "3-C\|lab03c\|ch03c\|Part 3\|Part 4\|路线\|目录" README.md | head -40
````

<details><summary>↳ Bash 结果</summary>

````
16:open courseware/ch03a-consensus-election.html  # Part 3-A：共识与 Raft 选举
17:open courseware/ch03b-log-replication.html    # Part 3-B：日志复制与安全性
18:open courseware/ch03c-production-raft.html   # Part 3-C：持久化 / 快照 / 成员变更 / 线性一致读
27:cd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读
35:cd java/lab03c && java Lab03C.java
76:## 学习路线
85:| **3-C** | **持久化 / 快照 / 成员变更 / 线性一致读** —— 三样必落盘的状态、InstallSnapshot、单节点变更的交集证明、ReadIndex 与 Lease Read | 交互课件 + Lab 3-C | ✅ 已完成 |
90:> Part 3 的 Raft 实现是整条路线的分水岭。建议在那里慢下来，真的把代码跑通。
94:## 目录结构
103:│   ├── ch03a-consensus-election.html # Part 3-A 课件（2PC 阻塞 + Raft 选举模拟器）
104:│   ├── ch03b-log-replication.html # Part 3-B 课件（日志复制 + Figure 8 交互重放）
105:│   └── ch03c-production-raft.html # Part 3-C 课件（崩溃重启 + 成员变更裂脑 + 三种读）
131:│   └── lab03c/
141:    └── lab03c/Lab03C.java
215:## Lab 3-C 会让你亲手复现什么
````

</details>

▸ **Bash** Read README sections in detail

````
sed -n '76,92p;125,145p;210,240p' README.md
````

<details><summary>↳ Bash 结果</summary>

````
## 学习路线

| Part | 主题 | 核心产出 | 状态 |
|---|---|---|---|
| **0** | **序言与心智模型** —— 三个免费假设的崩塌、三个不可靠、故障模型、两将军、幂等性 | 交互课件 + Lab 0 | ✅ 已完成 |
| **1** | **时间、顺序与因果** —— happens-before、Lamport 时钟、向量时钟、LWW 陷阱、HLC、TrueTime | 交互课件 + Lab 1 | ✅ 已完成 |
| **2** | **复制与一致性模型** —— 复制拓扑、三大延迟异常、一致性谱系、CAP/PACELC、Quorum、CRDT | 交互课件 + Lab 2 | ✅ 已完成 |
| **3-A** | **共识与 Raft 选举** —— 共识定义、FLP、2PC 为什么不是共识、任期、选举规则、随机化 | 交互课件 + Lab 3-A | ✅ 已完成 |
| **3-B** | **日志复制与安全性** —— 日志匹配性质、提交规则、Figure 8、五条安全性属性、快速回退 | 交互课件 + Lab 3-B | ✅ 已完成 |
| **3-C** | **持久化 / 快照 / 成员变更 / 线性一致读** —— 三样必落盘的状态、InstallSnapshot、单节点变更的交集证明、ReadIndex 与 Lease Read | 交互课件 + Lab 3-C | ✅ 已完成 |
| 4 | 分布式事务 —— 隔离级别、MVCC、2PC/XA、TCC、Saga、事务消息、Percolator | Lab 4：TCC 与 Saga 对照实现 | ⏳ |
| 5 | 经典系统精读 —— GFS / MapReduce / Bigtable / Dynamo / Spanner / ZooKeeper / Kafka / RocketMQ / Vitess | 架构拆解笔记 | ⏳ |
| 6 | 工程化与验证 —— 分布式锁陷阱、混沌工程、Jepsen、TLA+ | Lab 6：给自己的 Raft 跑线性一致性检查 | ⏳ |

> Part 3 的 Raft 实现是整条路线的分水岭。建议在那里慢下来，真的把代码跑通。

---
│   │   ├── raft.go          # Raft 选举状态机 + Election Safety 断言
│   │   └── main.go
│   ├── lab03b/
│   │   ├── net.go           # 同上，Msg 扩展了日志字段
│   │   ├── raft.go          # 日志复制 + 提交规则 + 三条安全性断言
│   │   └── main.go
│   └── lab03c/
│       ├── net.go           # 再扩展 InstallSnapshot 消息
│       ├── raft.go          # 持久化 / 快照 / 每节点独立 config / 三种读
│       └── main.go
└── java/
    ├── lab00/Lab00.java     # 单文件，零依赖
    ├── lab01/Lab01.java
    ├── lab02/Lab02.java
    ├── lab03a/Lab03A.java
    ├── lab03b/Lab03B.java
    └── lab03c
…（截断，原长 2511 字符）
````

</details>

▸ **Bash** Update README for Part 3-D

````
python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
subs=[
# quick start: courseware
('open courseware/ch03c-production-raft.html   # Part 3-C：持久化 / 快照 / 成员变更 / 线性一致读\n',
 'open courseware/ch03c-production-raft.html   # Part 3-C：持久化 / 快照 / 成员变更 / 线性一致读\n'
 'open courseware/ch03d-faq.html               # Part 3-D：附录（四个 index / Quorum vs Raft / 配置与冷启动）\n'),
# quick start: go
('cd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读\n',
 'cd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读\n'
 'cd go/lab03d && go run .     # Lab 3-D：ReadIndex 第三步 / no-op / Quorum vs Raft / 冷启动\n'),
# quick start: java
('cd java/lab03c && java Lab03C.java\n','cd java/lab03c && java Lab03C.java\ncd java/lab03d && java Lab03D.java\n'),
# roadmap
('| 4 | 分布式事务 ',
 '| **3-D** | **附录：算法与
…（截断，原长 3120 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README updated
19:open courseware/ch03d-faq.html               # Part 3-D：附录（四个 index / Quorum vs Raft / 配置与冷启动）
29:cd go/lab03d && go run .     # Lab 3-D：ReadIndex 第三步 / no-op / Quorum vs Raft / 冷启动
38:cd java/lab03d && java Lab03D.java
89:| **3-D** | **附录：算法与系统的接缝** —— 四个 index 与 ReadIndex 的第三步、no-op 的必要性、Quorum ≠ Raft、配置从哪来与冷启动、Raft log ≠ redo log | 交互课件 + Lab 3-D | ✅ 已完成 |
110:│   └── ch03d-faq.html           # Part 3-D 课件（index 流水线 + Quorum/Raft 对照 + 冷启动沙盒）
140:│   └── lab03d/
151:    └── lab03d/Lab03D.java
238:## Lab 3-D 会让你亲手复现什么
242:| **3D-1** ReadIndex 第三步 | ReadIndex 的第 1、2 步**全部通过**，只跳过第 3 步：apply 滞后 4 条时 **387/500 次读到旧值**，滞后 8 条时 **86%**。执行第 3 步则恒为 0，代价是平均多等约 2 条 |
243:| **3D-2** no-op 的必要性 | 新 Leader 不发 no-op：**16/40 个种子读到旧值，而客户端收到过「写入成功」**。发 no-op：违规 0 次，另有 15 次「正确拒绝」——*给你错误，而不是给你错的数据* |
244:| **3D-3** Quorum vs Raft | W 从 1 调到 5，**写成功数从 2.00 掉到 0.70，副本分歧数一动不动（1.88 种）**。Raft 侧恒为 1.00 种——不是概率，是结构 |
245:| **3D-4** 冷启动 | 一个节点的 `--initial-cluster` 写错 → **2 个集群 2 个 Leader，都能写**；灾难恢复误用 `state=new` → 1 个 Leader、健康检查全绿、**半年数据静默消失** |
247:3D-1 的那张表是这一章的眼睛：**「我确认自己还是 Leader」和「我确认提交点是 7」都成立，读出来仍然可以是旧数据。** 少的那一步是「等状态机追上来」。
249:这个 Lab 不复用 `lab03c/raft.go`——3-A 到 3-C 研究的是**算法本身**会不会出错，需要完整状态机；3-D 研究的是**算法与系统之间的接缝**，用聚焦的小模型讲得更清楚。
253:go run . -reads 500       # 3D-1 的读次数
254:go run . -seeds 40        # 3D-2 / 3D-3 的种子数
255:go run . -writers 4 -w 5  # 3D-3 的并发写客户端数与写 quorum
````

</details>

▸ **Bash** List new files

````
ls go/lab03d java/lab03d courseware/ch03d-faq.html && echo "--- sizes ---" && du -sh courseware/ch03d-faq.html go/lab03d java/lab03d
````

<details><summary>↳ Bash 结果</summary>

````
courseware/ch03d-faq.html

go/lab03d:
main.go
model.go
rand.go

java/lab03d:
Lab03D.java
--- sizes ---
96K	courseware/ch03d-faq.html
40K	go/lab03d
36K	java/lab03d
````

</details>

▸ **device:get_device_info**

````
{}
````

<details><summary>↳ mcp__remote-devices__get_device_info 结果</summary>

````
{
  "platform": "darwin",
  "arch": "arm64",
  "appVersion": "1.52386.6",
  "electronVersion": "44.2.0",
  "nodeVersion": "24.20.0",
  "deviceName": "〔设备名〕",
  "connectedFolders": [
    "〔路径00001〕"
  ],
  "homeDirectories": [
    {
      "name": ".android"
    },
    {
      "name": ".azure"
    },
    {
      "name": ".azure-bot"
    },
    {
      "name": ".cache"
    },
    {
      "name": ".chelper"
    },
    {
      "name": ".cmuxterm"
    },
    {
      "name": ".codegraph"
    },
    {
      "name": ".codex"
    },
    {
      "name": ".colima"
    },
    {
      "name": ".config"
    },
    {
      "name": ".copilot"
    },
    {
      "name": ".gradle"
    },
    {
      "name": ".homebrew"
    },
    {
      "name": ".local"
    },
    {
      "name": ".m2"
    },
    {
      "name": ".mcp"
    },
    {
      "name": ".mirasim"
    },
    {
      "name": ".npm"
    },
    {
      "name": ".oh-my-zsh"
    },
    {
      "name": ".pi-gui"
    },
    {
      "name": ".sogouinput"
    },
    {
      "name": ".swiftpm"
    },
    {
      "name": ".Trash"
    },
    {
      "name": ".UniVPN"
    },
    {
      "name": ".vscode"
    },
    {
      "name": ".vscode-shared"
    },
    {
      "name": ".work-memory"
    },
    {
      "name": ".zsh_sessions"
    },
    {
      "name": "Applications"
    },
    {
      "name": "bin"
    },
    {
      "name": "Claude"
    },
    {
      "name": "Code"
    },
    {
      "name": "Desktop",
      "requires
…（截断，原长 1995 字符）
````

</details>

## 🤖 Claude · 2026-09-14 17:49

📎 **发送文件**：Part 3-D 附录课件：四个 index 流水线 + no-op 重放 + Quorum/Raft 对照沙盒 + 冷启动沙盒，10 道自测题。
- `〔路径00077〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00077〕 → file_uuid: 〔标识03169〕
````

</details>

💾 **写入你的 Mac**：
- `〔路径00011〕`
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

▸ **Bash** Stage files for commit

````
mkdir -p 〔路径00167〕 && cp 〔路径00112〕 〔路径00171〕 && cp 〔路径00113〕 〔路径00172〕 && cp 〔路径00114〕 〔路径00173〕 && cp 〔路径00139〕 〔路径00169〕 && cp 〔路径00056〕 〔路径00170〕 && ls -la 〔路径00168〕
````

<details><summary>↳ mcp__remote-devices__device_commit_files 结果（错误）</summary>

````
{"written":["〔路径00011〕"],"rejected":[{"devicePath":"〔路径00036〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00037〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00038〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00050〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00004〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."}]}
````

</details>

<details><summary>↳ Bash 结果</summary>

````
total 92
drwxr-xr-x 2 root root  4096 Sep 14 09:49 .
drwxr-xr-x 4 root root  4096 Sep 14 09:49 ..
-rw-r--r-- 1 root root 30127 Sep 14 09:49 Lab03D.java
-rw-r--r-- 1 root root 16244 Sep 14 09:49 README.md
-rw-r--r-- 1 root root 17767 Sep 14 09:49 lab03d-main.go
-rw-r--r-- 1 root root 10870 Sep 14 09:49 lab03d-model.go
-rw-r--r-- 1 root root   966 Sep 14 09:49 lab03d-rand.go
````

</details>

💾 **写入你的 Mac**：
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果（错误）</summary>

````
{"written":[],"rejected":[{"devicePath":"〔路径00036〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00037〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00038〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00050〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."},{"devicePath":"〔路径00004〕","reason":"HTTP 404 fetching session file. No such file under 〔路径00168〕 in this session. Call SendUserFile and pass its fileUuid to device_commit_files instead."}]}
````

</details>

📎 **发送文件**：Lab 3-D 的 Go / Java 源码与更新后的 README。
- `〔路径00112〕`
- `〔路径00113〕`
- `〔路径00114〕`
- `〔路径00139〕`
- `〔路径00056〕`

<details><summary>↳ SendUserFile 结果</summary>

````
5 files delivered to user.
  〔路径00112〕 → file_uuid: 〔标识00091〕
  〔路径00113〕 → file_uuid: 〔标识02761〕
  〔路径00114〕 → file_uuid: 〔标识03511〕
  〔路径00139〕 → file_uuid: 〔标识02013〕
  〔路径00056〕 → file_uuid: 〔标识02832〕
````

</details>

💾 **写入你的 Mac**：
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00036〕","〔路径00037〕","〔路径00038〕","〔路径00050〕","〔路径00004〕"],"rejected":[]}
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00144〕\" && ls courseware/ && echo \"--- go ---\" && ls go/lab03d/ && echo \"--- java ---\" && ls java/lab03d/ && echo \"--- go run ---\" && (cd go/lab03d && go run . 2>&1 | head -20)"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ch00-intro.html
ch01-time-order.html
ch02-replication.html
ch03a-consensus-election.html
ch03b-log-replication.html
ch03c-production-raft.html
ch03d-faq.html
--- go ---
main.go
model.go
rand.go
--- java ---
Lab03D.java
--- go run ---
bash: line 1: go: command not found
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #24 status
````

</details>

▸ **TaskUpdate** in_progress

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #25 status
````

</details>

▸ **Bash** Write Part 4-A courseware chunk 1

````
cat > 〔路径00227〕 <<'HTMLEOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 4-A · 单机事务：隔离级别、MVCC 与写偏斜</title>
HTMLEOF
cat 〔路径00240〕 >> 〔路径00227〕
cat >> 〔路径00227〕 <<'HTMLEOF'
<style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  background:var(--surface-2);border:1px solid var(--border);color:var(--ink)}
.two{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.three{display:grid;grid-template-colum
…（截断，原长 8932 字符）
````

<details><summary>↳ Bash 结果</summary>

````
22197 〔路径00227〕
````

</details>

▸ **Bash** Write Part 4-A courseware chunk 2

````
cat > 〔路径00228〕 <<'HTMLEOF'

<section id="s2">
  <h2>4.2 隔离级别：一个按「禁止什么」定义的烂摊子</h2>
  <p>ANSI SQL-92 定义隔离级别的方式是<strong>列举它禁止哪些异常</strong>：</p>

  <div class="grid-wrap">
  <table>
    <thead><tr><th style="width:24%">级别</th><th style="width:19%">脏读</th><th style="width:19%">不可重复读</th><th style="width:19%">幻读</th><th>说明</th></tr></thead>
    <tbody>
      <tr><td><code>READ UNCOMMITTED</code></td><td style="color:var(--crit)">可能</td><td style="color:var(--crit)">可能</td><td style="color:var(--crit)">可能</td><td>几乎无人使用</td></tr>
      <tr><td><code>READ COMMITTED</code></td><td style="color:var(--ok)">禁止</td><td style="color:var(--crit)">可能</td><td style="color:var(--crit)">可能</td><td><strong>PostgreSQL / Oracle / SQL Server 的默认</strong></td></tr>
      <tr><td><code>REPEATABLE READ</code></
…（截断，原长 10319 字符）
````

<details><summary>↳ Bash 结果</summary>

````
13385 〔路径00228〕
````

</details>

▸ **Bash** Write Part 4-A courseware chunk 3

````
cat > 〔路径00229〕 <<'HTMLEOF'

<section id="s4">
  <h2>4.4 MVCC 与 ReadView：InnoDB 的快照是怎么做出来的</h2>

  <div class="note warn">
    <div class="note-h"><span class="dot" style="background:var(--warn)"></span>先纠一个极常见的混淆：版本信息在 undo log，不在 redo log</div>
    <p style="margin-bottom:0"><strong>redo log</strong> 是<em>物理/物理逻辑</em>日志：「把表空间 5 第 137 页 offset 92 处的 4 个字节改成 0x2A」。它只用于崩溃恢复时的 roll-forward，<u>和可见性判断毫无关系</u>。<br>
    <strong>undo log</strong> 才是 MVCC 的载体：它保存每一行的<em>历史版本</em>，既用于回滚，也用于快照读。<br>
    <span style="color:var(--ink-3)">搞混的代价：会想不通「redo log 都刷盘了，事务怎么还能回滚」。</span></p>
  </div>

  <h3>行上的两个隐藏列</h3>
  <p>InnoDB 的每一行除了你定义的字段，还偷偷带着：</p>
<pre><code>| id | name  | balance | <span class="cm">DB_TRX_ID</span> | <span class="cm">DB_ROLL_PTR</span> |
                            ↑
…（截断，原长 10256 字符）
````

<details><summary>↳ Bash 结果</summary>

````
14227 〔路径00229〕
````

</details>

▸ **Bash** Write Part 4-A courseware chunk 4

````
cat > 〔路径00230〕 <<'HTMLEOF'

<section id="s7">
  <h2>4.7 可串行化的三条路</h2>
  <p>如果打折版太危险，为什么不都买全价的？因为全价的三种实现各有各的贵法。</p>

  <div style="display:grid;gap:8px;margin:18px 0">
    <div class="pstep" style="border-color:var(--s3)"><div class="n" style="background:var(--s3)">1</div><div>
      <div class="ti" style="font-weight:660">实际串行执行 —— 干脆别并发</div>
      <div class="de" style="font-size:13.5px;color:var(--ink-2)">单线程按顺序跑完每个事务。<strong>可串行化的定义是「等价于某个串行顺序」，那最省事的实现就是真的串行。</strong><br>
      能用的前提：① 整个工作集放得进内存 ② 事务必须<u>短</u>（不能有交互式等待）③ 通常要求事务以<strong>存储过程</strong>形式提交，一次性发过去，不能来回聊天。<br>
      代表：Redis（单线程）、VoltDB / H-Store、Datomic。<br>
      <span style="color:var(--ink-3)">吞吐天花板 = 单核。想扩展只能分区，而<u>跨分区事务会把优势全吃掉</u>——这个矛盾会在 Part 4-C 的 Calvin 里再次出现。</span></div></div>
    <div class="pstep" style="b
…（截断，原长 13628 字符）
````

<details><summary>↳ Bash 结果</summary>

````
21559 〔路径00230〕
````

</details>

▸ **Bash** Write mini transaction engine JS

````
cat > 〔路径00231〕 <<'JSEOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHeight)*100)+"%";
  let cur=secs[0].id;for(const s of secs)if(s.getBoundingClientRect().top<=120)cur=s.id;
  links.forEach(a=>a.classList.toggle("on",a.getAttribute("href")==="#"+cur));}
document.addEventListener("scroll",onScroll,{passive:true});onScroll();
function wrapT(root){(root||document).querySel
…（截断，原长 8523 字符）
````

<details><summary>↳ Bash 结果</summary>

````
9754 〔路径00231〕
````

</details>

▸ **Bash** Write scenarios and demos 1-2

````
cat > 〔路径00232〕 <<'JSEOF'

/* ══════════════════════════════════════════════════════════════════
   实验 1 · 七种异常 × 五种隔离配置
   ══════════════════════════════════════════════════════════════════ */
const P_ONCALL={tag:"oncall",label:"on_call = true",test:(k,v)=>k.indexOf("doc:")===0&&v===1};
const P_ADULT ={tag:"adult", label:"age > 18",      test:(k,v)=>k.indexOf("p:")===0&&v>18};
const P_ROOM  ={tag:"roomA", label:"room='A' AND slot=10",test:(k,v)=>k.indexOf("bk:")===0&&v===1};

const SCEN=[
{id:"dw",name:"脏写",en:"dirty write",ansi:true,
 init:{x:0,y:0},
 steps:[{t:1,op:"begin"},{t:2,op:"begin"},
  {t:1,op:"write",key:"x",val:1},{t:2,op:"write",key:"x",val:2},
  {t:2,op:"write",key:"y",val:2},{t:1,op:"write",key:"y",val:1},
  {t:1,op:"commit"},{t:2,op:"commit"}],
 check:e=>({bad:e.finalVa
…（截断，原长 11493 字符）
````

<details><summary>↳ Bash 结果</summary>

````
14702 〔路径00232〕
````

</details>

▸ **Bash** Write demos 3-4 and assemble Part 4-A page

````
cat > 〔路径00233〕 <<'JSEOF'

/* ══════════════════════════════════════════════════════════════════
   实验 3 · 医院值班表：谁能挡住写偏斜
   ══════════════════════════════════════════════════════════════════ */
function skewSteps(D,C,rnd){
  // C 个医生同时发起请假：读一遍在岗人数 → 如果 >= 2 就把自己设成不在岗
  const per=[];
  for(let t=1;t<=C;t++) per.push([
    {t,op:"begin"},
    {t,op:"scan",pred:P_ONCALL,lockable:true},
    {t,op:"write",key:"doc:"+t,val:0,guardMin:2},
    {t,op:"commit"}]);
  // 保持每个事务内部顺序不变，随机交错
  const idx=per.map(()=>0), out=[];
  let left=per.reduce((a,p)=>a+p.length,0);
  while(left>0){
    const avail=[];
    for(let i=0;i<C;i++) if(idx[i]<per[i].length) avail.push(i);
    const pick=avail[Math.floor(rnd()*avail.length)];
    out.push(per[pick][idx[pick]++]); left--;
  }
  return out;
}
function run3
…（截断，原长 10213 字符）
````

<details><summary>↳ Bash 结果</summary>

````
suspects: 0
107809 ch04a-transactions.html
````

</details>

▸ **Bash** Verify Part 4-A courseware

````
cat > 〔路径00256〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const p=await b.newPage({viewport:{width:1400,height:1000}});
  const errs=[];
  p.on('console',m=>{if(m.type()==='error')errs.push('CONSOLE: '+m.text());});
  p.on('pageerror',e=>errs.push('PAGEERROR: '+e.message));
  await p.goto('file://〔路径00078〕');
  await p.waitForTimeout(600);

  // ---- Demo 1 matrix ----
  const rows=await p.locator('#mx1 tbody tr').count();
  console.log('matrix rows:', rows);
  const mx=await p.evaluate(()=>{
    const out=[];
    document.querySelectorAll('#mx1 tbody tr').forEach(tr=>{
      const name=tr.query
…（截断，原长 3609 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
matrix rows: 0
           RU RC RR LK SER
node:internal/process/promises:394
    triggerUncaughtException(err, true /* fromPromise */);
    ^

page.click: Timeout 30000ms exceeded.
Call log:
[2m  - waiting for locator('#mx1 tbody tr:nth-child(6) .cellbtn:nth-child(1)')[22m

    at 〔路径00257〕 {
  log: [
    "  - waiting for locator('#mx1 tbody tr:nth-child(6) .cellbtn:nth-child(1)')"
  ],
  name: 'TimeoutError'
}

Node.js v22.22.2

[agent-proxy] While this command ran, 8 connections through the agent proxy failed:
- android.clients.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×2
- accounts.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×4
- www.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×2
For details: curl -sS http://127.0.0.1:33367/__agentproxy/status
````

</details>

▸ **Bash** Fix reads tracking and check errors

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()
a='      x={t:s.t,id:this.nextId++,rv:null,last:undefined,done:false};'
b='      x={t:s.t,id:this.nextId++,rv:null,last:undefined,reads:[],done:false};'
assert s.count(a)==1; s=s.replace(a,b)
a2='        const v=this.readFor(x,s.key,s.lockable); x.last=v;'
b2='        const v=this.readFor(x,s.key,s.lockable); x.last=v; x.reads.push(v);'
assert s.count(a2)==1; s=s.replace(a2,b2)
a3='        x.last=keys.length;'
b3='        x.last=keys.length; x.reads.push(keys.length);'
assert s.count(a3)==1; s=s.replace(a3,b3)
# dedupe blocked trace entries
a4='    this.deadlock=false;\n  }'
b4='    this.deadlock=false;\n    this.blkSeen=new Set();\n  }'
assert s.count(a4)==1; s=s.replace(a4,b4)
a5='        waiting[s.t]=true;\n
…（截断，原长 1848 字符）
````

<details><summary>↳ Bash 结果</summary>

````
patched
matrix rows: 7
no errors
````

</details>

▸ **Bash** Re-run full verification

````
node 〔路径00256〕
````

<details><summary>↳ Bash 结果</summary>

````
matrix rows: 7
           RU RC RR LK SER
脏写           X . . . .
脏读           X . . . .
不可重复读        X X . . .
幻读           X X . . .
丢失更新 不在 ANSI 清单 X X X . .
写偏斜 不在 ANSI 清单 X X X . .
写偏斜（插入型） 不在 ANSI 清单 X X X . .

write skew @RR trace rows: 8
verdict: 最终在岗医生数 0（不变量要求 >= 1）

readview default: 400
after trx24 commits: 400

write-skew stats: 120 / 200 | 0 / 200 | 120 / 200 | 0 / 200 | 0 / 200 | 120 / 200 | 0 / 200 | 0 / 200
D=4 C=4: 163 / 200 | 0 / 200 | 163 / 200 | 0 / 200 | 0 / 200 | 198 / 200 | 0 / 200 | 0 / 200

2PL: 总耗时15 ticks中止次数0吞吐53.3 /100t
SSI: 总耗时12 ticks中止次数8吞吐66.7 /100t
cross: 35
C=16 O=6 -> 2PL: 总耗时20000 ticks中止次数19997吞吐0.1 /100t
C=16 O=6 -> SSI: 总耗时48 ticks中止次数48吞吐33.3 /100t
cross: none

quiz items: 10
sections: 11
mobile overflow: true
NO CONSOLE ERRORS
````

</details>

▸ **Bash** Fix ReadView default and rewrite 2PL/SSI simulators

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()

# ---- fix 1: ReadView demo default so the toggle actually changes the answer ----
a='let MIDS=new Set([24,32]);'
b='let MIDS=new Set([28,32]);'
assert s.count(a)==1; s=s.replace(a,b)
a='把 <code>trx 24</code> 从活跃集合里去掉（相当于它提交了），再看返回值怎么跳——'
b='把 <code>trx 28</code> 从活跃集合里去掉（相当于它提交了），再看返回值怎么跳——'
assert s.count(a)==1; s=s.replace(a,b)

# ---- fix 2: rewrite 2PL (wait + 环检测 + 冷却) and add wasted-work metric ----
old_start=s.index('// 2PL：悲观。拿不到锁就等；所有人都等 = 死锁，中止一个重来。')
old_end=s.index('function bar(v,max,color){')
new_sims='''// 2PL：悲观。拿不到锁就等；等待图成环 = 死锁，挑一个牺牲者中止重来。
// 关键性质：等待不浪费已做的功——等到了就接着做。
function sim2PL(txns){
  const st=txns.map(t=>({id:t.id,ops:t.ops,p:0,locks:[],done:false,cool:0,retries:0}));
  const owner={
…（截断，原长 5485 字符）
````

<details><summary>↳ Bash 结果</summary>

````
patched
matrix rows: 7
no errors
````

</details>

▸ **Bash** Verify readview and 2PL/SSI behavior

````
cat > 〔路径00253〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const p=await b.newPage({viewport:{width:1400,height:1000}});
  const errs=[]; p.on('pageerror',e=>errs.push(e.message));
  p.on('console',m=>{if(m.type()==='error')errs.push(m.text());});
  await p.goto('file://〔路径00078〕');
  await p.waitForTimeout(700);
  // readview
  let o2=(await p.textContent('#o2')).replace(/\s+/g,' ');
  console.log('ReadView default →', /balance = (\d+)/.exec(o2)?.[1]);
  await p.click('#mids button[data-id="28"]');
  o2=(await p.textContent('#o2')).replace(/\s+/g,' ');
  console.log('after trx28 commits →', /balance
…（截断，原长 1874 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ReadView default → 300
after trx28 commits → 400
C=8 O=4 hot=0%  2PL[t=14 ab=0 waste=0 eff=25.0]  SSI[t=12 ab=4 waste=16 eff=16.7]
C=8 O=4 hot=50%  2PL[t=15 ab=0 waste=0 eff=25.0]  SSI[t=16 ab=9 waste=36 eff=11.8]
C=8 O=4 hot=100%  2PL[t=37 ab=5 waste=5 eff=21.6]  SSI[t=24 ab=20 waste=80 eff=7.1]
  cross: none
C=16 O=6 hot=0%  2PL[t=81 ab=8 waste=19 eff=13.9]  SSI[t=30 ab=26 waste=156 eff=6.3]
C=16 O=6 hot=50%  2PL[t=179 ab=21 waste=42 eff=11.6]  SSI[t=66 ab=64 waste=384 eff=3.3]
C=16 O=6 hot=100%  2PL[t=131 ab=14 waste=23 eff=13.4]  SSI[t=78 ab=89 waste=534 eff=2.5]
  cross: none
C=24 O=10 hot=0%  2PL[t=6000 ab=544 waste=107 eff=14.4]  SSI[t=170 ab=149 waste=1490 eff=1.4]
C=24 O=10 hot=50%  2PL[t=6000 ab=543 waste=107 eff=11.1]  SSI[t=190 ab=198 waste=1980 eff=1.1]
C=24 O=10 hot=100%  2PL[t=6000 ab=545 waste=588 eff=3.7]  SSI[t=230 ab=257 waste=2570 eff=0.9]
  cross: none
no errors
````

</details>

▸ **Bash** Rewrite simulators with bounded parallelism

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()
a=s.index('// 2PL：悲观。拿不到锁就等；等待图成环 = 死锁，挑一个牺牲者中止重来。')
b=s.index('function bar(v,max,color){')
new='''// 两个模拟器共用同一个物理前提：系统每 tick 最多能执行 PAR 个操作。
// 这一条很重要——它让「白做的功」真的变成时间，而不是一个漂亮的免费数字。
const PAR=4;
const inter=(A,B)=>{for(const k of A) if(B.has(k)) return true; return false;};

// 2PL：悲观。拿不到锁就等（等待期间不占用 CPU）；等待图成环 = 死锁，挑牺牲者中止。
// 关键性质：等待【不浪费】已经做的功——等到了就接着往下做。
// 进度保证：编号最小的活跃事务永远不会被选为牺牲者 ⇒ 它一定能跑完，然后轮到下一个。
function sim2PL(txns){
  const st=txns.map(t=>({id:t.id,ops:t.ops,p:0,locks:[],done:false,cool:0}));
  const owner={};
  let tick=0,aborts=0,wasted=0,work=0;
  while(st.some(s=>!s.done)&&tick<8000){
    tick++;
    let budget=PAR, moved=false; const waitFor={};
    for(const s of st){
      if(s.done||budget<=0) co
…（截断，原长 4711 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
ReadView default → 300
after trx28 commits → 400
C=8 O=4 hot=0%  2PL[t=11 ab=0 waste=0 eff=undefined]  SSI[t=16 ab=3 waste=12 eff=undefined]
C=8 O=4 hot=50%  2PL[t=21 ab=2 waste=4 eff=undefined]  SSI[t=16 ab=4 waste=16 eff=undefined]
C=8 O=4 hot=100%  2PL[t=8000 ab=2664 waste=3 eff=undefined]  SSI[t=20 ab=7 waste=28 eff=undefined]
  cross: 10
C=16 O=6 hot=0%  2PL[t=8000 ab=2663 waste=0 eff=undefined]  SSI[t=48 ab=12 waste=72 eff=undefined]
C=16 O=6 hot=50%  2PL[t=8000 ab=2664 waste=0 eff=undefined]  SSI[t=78 ab=30 waste=180 eff=undefined]
C=16 O=6 hot=100%  2PL[t=8000 ab=2665 waste=0 eff=undefined]  SSI[t=90 ab=38 waste=228 eff=undefined]
  cross: none
C=24 O=10 hot=0%  2PL[t=8000 ab=2661 waste=0 eff=undefined]  SSI[t=200 ab=47 waste=470 eff=undefined]
C=24 O=10 hot=50%  2PL[t=8000 ab=2664 waste=0 eff=undefined]  SSI[t=230 ab=62 waste=620 eff=undefined]
C=24 O=10 hot=100%  2PL[t=8000 ab=2666 waste=0 eff=undefined]  SSI[t=240 ab=66 waste=660 eff=undefined]
  cross: none
no errors
````

</details>

▸ **Bash** Fix 2PL scheduler and scale keyspace

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()
a=s.index('function sim2PL(txns){')
b=s.index('// SSI：乐观。谁也不等，全速往前跑；跑完一整个事务才验证')
new='''function sim2PL(txns){
  const st=txns.map(t=>({id:t.id,ops:t.ops,p:0,locks:[],done:false,cool:0}));
  const owner={};
  let tick=0,aborts=0,wasted=0,work=0;
  while(st.some(s=>!s.done)&&tick<8000){
    tick++;
    const alive=st.filter(s=>!s.done);
    // 先算出这一 tick 谁【能】动（锁没被占住），再按并行度预算执行
    const runnable=[], waitFor={};
    for(const s of alive){
      if(s.cool>0){ s.cool--; runnable.push(null); continue; }   // 退避中，算作有进展
      const op=s.ops[s.p], o=owner[op.key];
      if(o&&o.trx!==s.id&&(o.x||op.write)){ waitFor[s.id]=o.trx; continue; }
      runnable.push(s);
    }
    if(runnable.length===0){
      // 没有任何事务能动 ⇒ 等
…（截断，原长 2229 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
ReadView default → 300
after trx28 commits → 400
C=8 O=4 hot=0%  2PL[t=10 ab=0 waste=0 eff=undefined]  SSI[t=12 ab=2 waste=8 eff=undefined]
C=8 O=4 hot=50%  2PL[t=13 ab=0 waste=0 eff=undefined]  SSI[t=20 ab=6 waste=24 eff=undefined]
C=8 O=4 hot=100%  2PL[t=8000 ab=2666 waste=0 eff=undefined]  SSI[t=28 ab=14 waste=56 eff=undefined]
  cross: 65
C=16 O=6 hot=0%  2PL[t=27 ab=0 waste=0 eff=undefined]  SSI[t=42 ab=8 waste=48 eff=undefined]
C=16 O=6 hot=50%  2PL[t=8000 ab=2665 waste=0 eff=undefined]  SSI[t=84 ab=34 waste=204 eff=undefined]
C=16 O=6 hot=100%  2PL[t=8000 ab=2663 waste=0 eff=undefined]  SSI[t=90 ab=38 waste=228 eff=undefined]
  cross: 20
C=24 O=10 hot=0%  2PL[t=8000 ab=1590 waste=3180 eff=undefined]  SSI[t=100 ab=14 waste=140 eff=undefined]
C=24 O=10 hot=50%  2PL[t=8000 ab=2661 waste=0 eff=undefined]  SSI[t=220 ab=58 waste=580 eff=undefined]
C=24 O=10 hot=100%  2PL[t=8000 ab=2666 waste=0 eff=undefined]  SSI[t=240 ab=66 waste=660 eff=undefined]
  cross: none
no errors
````

</details>

▸ **Bash** Add proper deadlock cycle detection

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()
a='''    if(runnable.length===0){
      // 没有任何事务能动 ⇒ 等待图里必然有环。
      // 编号最小的活跃事务豁免，保证它一定跑得完，然后轮到下一个 ⇒ 一定终止。
      const protectedId=alive[0].id;
      const pool=alive.filter(s=>s.id!==protectedId);
      if(!pool.length) break;
      pool.sort((x,y)=>x.p-y.p);                 // 挑已做的功最少的，浪费最小
      const v=pool[0];'''
b='''    if(runnable.length===0){
      // 没有任何事务能动 ⇒ 等待图里必然有环。把环找出来，只在【环里】挑牺牲者——
      // 中止一个环外的事务不会解开任何东西，那正是「死锁检测」和「随便杀一个」的区别。
      let cyc=null;
      for(const k in waitFor){
        const path=[], seen=new Set(); let cur=+k;
        while(cur!==undefined&&!seen.has(cur)){ seen.add(cur); path.push(cur); cur=waitFor[cur]; }
        if(cur!==undefined){ cyc=path.slice(path.indexOf(cur)); b
…（截断，原长 1386 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
ReadView default → 300
after trx28 commits → 400
C=8 O=4 hot=0%  2PL[t=10 ab=0 waste=0 eff=undefined]  SSI[t=12 ab=2 waste=8 eff=undefined]
C=8 O=4 hot=50%  2PL[t=13 ab=0 waste=0 eff=undefined]  SSI[t=20 ab=6 waste=24 eff=undefined]
C=8 O=4 hot=100%  2PL[t=28 ab=3 waste=6 eff=undefined]  SSI[t=28 ab=14 waste=56 eff=undefined]
  cross: 70
C=16 O=6 hot=0%  2PL[t=27 ab=0 waste=0 eff=undefined]  SSI[t=42 ab=8 waste=48 eff=undefined]
C=16 O=6 hot=50%  2PL[t=79 ab=7 waste=23 eff=undefined]  SSI[t=84 ab=34 waste=204 eff=undefined]
C=16 O=6 hot=100%  2PL[t=92 ab=8 waste=19 eff=undefined]  SSI[t=90 ab=38 waste=228 eff=undefined]
  cross: 25
C=24 O=10 hot=0%  2PL[t=80 ab=2 waste=7 eff=undefined]  SSI[t=100 ab=14 waste=140 eff=undefined]
C=24 O=10 hot=50%  2PL[t=256 ab=20 waste=117 eff=undefined]  SSI[t=220 ab=58 waste=580 eff=undefined]
C=24 O=10 hot=100%  2PL[t=276 ab=22 waste=79 eff=undefined]  SSI[t=240 ab=66 waste=660 eff=undefined]
  cross: 30
no errors
````

</details>

▸ **Bash** Rewrite 2PL/SSI conclusion prose honestly

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()

# ── §4.7 SSI bullet: 说实话，别编造「吞吐反超」 ──
a='''代价：<strong>低冲突时几乎白送，高冲突时中止率飙升</strong>——中止的事务要重试，实际吞吐可能比 2PL 还差。而且需要应用层<u>有重试逻辑</u>。<br>'''
b='''代价：<strong>低冲突时几乎白送，高冲突时中止率飙升</strong>——每一次中止都意味着整个事务白做一遍，而白做的功照样烧 CPU。而且需要应用层<u>有正确的重试逻辑</u>。<br>'''
assert s.count(a)==1; s=s.replace(a,b)
a='''<div class="s">同一批事务，悲观锁 vs 乐观检测，看吞吐与中止率</div>'''
b='''<div class="s">同一批事务，悲观锁 vs 乐观检测：两张不同的账单</div>'''
assert s.count(a)==1; s=s.replace(a,b)

# ── 实验 4 结论文案：换成诚实版本 ──
a=s.index("  const winner=b.tput>a.tput?\"SSI\":\"2PL\";")
bnd=s.index('  $("#o4").innerHTML=note;')
new='''  const wr=x=>x.work?(x.wasted/x.work*100):0;
  note+='<div class="note" style="margin:14px 0 0"><div class="note-h">'+
    '<span class="dot" style="back
…（截断，原长 3557 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
matrix rows: 7
no errors
````

</details>

▸ **Bash** Screenshot demos and diagnose mobile overflow

````
python3 〔路径00196〕 ch04a-transactions.html && cat > 〔路径00242〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const p=await b.newPage({viewport:{width:1380,height:1200},deviceScaleFactor:2});
  const errs=[]; p.on('pageerror',e=>errs.push(e.message));
  await p.goto('file://〔路径00078〕');
  await p.waitForTimeout(800);
  await p.click('#mx1 tbody tr:nth-child(6) td:nth-child(4) .cellbtn');   // write skew @ RR
  await p.waitForTimeout(100);
  await p.locator('#s3 .lab').screenshot({path:'〔路径00249〕'});
  await p.locator('#s4 .lab').screenshot({path:'〔路径00250〕'});
  await p.locator('#s6 .lab').screensh
…（截断，原长 1717 字符）
````

<details><summary>↳ Bash 结果</summary>

````
suspects: 0
{"sw":445,"cw":400,"out":["TABLE. w=563","THEAD. w=563","TR. w=563","TH. w=82","TH. w=135","TBODY. w=563","TD. w=82","TD. w=135"]}
no errors
````

</details>

▸ **Bash** Identify overflowing table

````
cat > 〔路径00200〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  const m=await b.newPage({viewport:{width:400,height:900}});
  await m.goto('file://〔路径00078〕');
  await m.waitForTimeout(600);
  const r=await m.evaluate(()=>{const W=document.documentElement.clientWidth,out=[];
    document.querySelectorAll('main table').forEach(e=>{const bb=e.getBoundingClientRect();
      if(bb.right>W+2){
        let anc=[],p=e; while(p&&p.tagName!=='MAIN'){anc.push(p.tagName+(p.id?'#'+p.id:'')+(typeof p.className==='string'&&p.className?'.'+p.className.split(' ')[0]:''));p=p.parentElement;}
        out.push({w:Math.round
…（截断，原长 1034 字符）
````

<details><summary>↳ Bash 结果</summary>

````
[
 {
  "w": 563,
  "right": 591,
  "head": " 你写的级别MySQL InnoDBPostgreSQLOracle后果 READ UNCOMMITTED真的脏读静默升",
  "anc": "TABLE < DIV.grid-wrap < SECTION#s2"
 },
 {
  "w": 657,
  "right": 706,
  "head": "READUNCOMMITTEDREADCOMMITTEDREPEATABLEREADRR +FOR UPDATESERI",
  "anc": "TABLE.mx < DIV.grid-wrap < DIV#mx1 < DIV.grid-wrap < DIV.lab-b < DIV.lab < SECTION#s3"
 }
]
````

</details>

▸ **Bash** Fix grid blowout across all courseware and verify

````
for f in *.html; do
python3 - "$f" <<'PY'
import sys
p=sys.argv[1]; s=open(p,encoding='utf-8').read()
if 'min-width:0' in s and '/* fix: grid blowout' in s: print(p,'already'); raise SystemExit
add='<style>/* fix: grid blowout —— grid 子项默认 min-width:auto 会撑破窄屏 */\nmain{min-width:0}\n.wrap{max-width:1240px}\n@media(max-width:900px){.wrap{grid-template-columns:minmax(0,1fr)}}\n.grid-wrap{max-width:100%}\n</style>\n</head>'
assert s.count('</head>')==1
s=s.replace('</head>',add)
open(p,'w',encoding='utf-8').write(s); print(p,'patched')
PY
done
cat > 〔路径00212〕 <<'JS'
const {chromium}=require('〔路径00142〕');
const fs=require('fs');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  for(const f of fs
…（截断，原长 1411 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch00-intro.html patched
ch01-time-order.html patched
ch02-replication.html patched
ch03a-consensus-election.html patched
ch03b-log-replication.html patched
ch03c-production-raft.html patched
ch03d-faq.html patched
ch04a-transactions.html patched
ch00-intro.html                scrollW=440 clientW=400 OVERFLOW
ch01-time-order.html           scrollW=400 clientW=400 ok
ch02-replication.html          scrollW=400 clientW=400 ok
ch03a-consensus-election.html  scrollW=400 clientW=400 ok
ch03b-log-replication.html     scrollW=400 clientW=400 ok
ch03c-production-raft.html     scrollW=400 clientW=400 ok
ch03d-faq.html                 scrollW=400 clientW=400 ok
ch04a-transactions.html        scrollW=445 clientW=400 OVERFLOW
````

</details>

▸ **Bash** Find remaining overflow sources

````
cat > 〔路径00201〕 <<'JS'
const {chromium}=require('〔路径00142〕');
(async()=>{
  const b=await chromium.launch({executablePath:'〔路径00178〕'});
  for(const f of ['ch04a-transactions','ch00-intro']){
    const m=await b.newPage({viewport:{width:400,height:900}});
    await m.goto('file://〔路径00070〕'+f+'.html');
    await m.waitForTimeout(600);
    const r=await m.evaluate(()=>{const W=document.documentElement.clientWidth,out=[];
      document.querySelectorAll('body *').forEach(e=>{const bb=e.getBoundingClientRect();
        if(bb.right>W+2){
          let ok=false,p=e.parentElement;
          while(p){const cs=getComputedStyle(p); if(cs.overflowX==='auto'||cs.overflowX==='scroll'||cs.overflowX==='hidden'){ok
…（截断，原长 1371 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch04a-transactions [
 {
  "w": 417,
  "right": 445,
  "t": "1 实际串行执行 —— 干脆别并发 单线程按顺序跑完每个事务。可串行化的定义是「",
  "anc": "DIV.pstep < DIV < SECTION#s7 < MAIN"
 },
 {
  "w": 185,
  "right": 428,
  "t": " 实际串行执行 —— 干脆别并发 单线程按顺序跑完每个事务。可串行化的定义是「等",
  "anc": "DIV < DIV.pstep < DIV < SECTION#s7"
 },
 {
  "w": 185,
  "right": 428,
  "t": "实际串行执行 —— 干脆别并发",
  "anc": "DIV.ti < DIV < DIV.pstep < DIV"
 },
 {
  "w": 185,
  "right": 428,
  "t": "单线程按顺序跑完每个事务。可串行化的定义是「等价于某个串行顺序」，那最省事的实现",
  "anc": "DIV.de < DIV < DIV.pstep < DIV"
 },
 {
  "w": 182,
  "right": 426,
  "t": "可串行化的定义是「等价于某个串行顺序」，那最省事的实现就是真的串行。",
  "anc": "STRONG < DIV.de < DIV < DIV.pstep"
 },
 {
  "w": 0,
  "right": 406,
  "t": "",
  "anc": "BR < DIV.de < DIV < DIV.pstep"
 }
]
ch00-intro [
 {
  "w": 412,
  "right": 440,
  "t": " 术语一句话解释常见误解 / 补充 节点 node参与系统的一个独立进程（通常一",
  "anc": "TABLE < SECTION#s9 < MAIN < DIV.wrap"
 },
 {
  "w": 412,
  "right": 440,
  "t": "术语一句话解释常见误解 / 补充",
  "anc": "THEAD < TABLE < SECTION#s9 < MAIN"
 },
 {
  "w": 412,
  "right": 440,
  "t": "术语一句话解释常见误解 / 补充",
  "anc": "TR < THEAD < TABLE < SECTION#s9"
 },
 {
  "w": 170,
  "right": 440,
  "t": "常见误解 / 补充",
  "anc": "TH < TR < THEAD < TABLE"
 },
 {
  "w": 412,
  "right": 440,
  "t": " 节点 node参与系统的一个独立进程（通常一台机器一个）。关注点是\"独立的失败",
  "anc": "TBODY < TABLE < SECTION#s9 < MAIN"
 },
 {
  "w": 412,
  "right": 440,
  "t": "节点 node参与系统的一个独立进程（通常一台机器一个）。关注点是\"独立的失败单",
  "anc": "TR < TBODY < TABLE < SECTION#s9"
 }
]
````

</details>

▸ **Bash** Add responsive stacking and auto table wrapping

````
for f in *.html; do
python3 - "$f" <<'PY'
import sys
p=sys.argv[1]; s=open(p,encoding='utf-8').read()
css='''main{min-width:0}
.wrap{max-width:1240px}
@media(max-width:900px){.wrap{grid-template-columns:minmax(0,1fr)}}
.grid-wrap{max-width:100%}
@media(max-width:700px){
  .pstep,.lrung,.axm,.idxrow,.kv{grid-template-columns:minmax(0,1fr)!important;gap:8px!important}
}
'''
old='''main{min-width:0}
.wrap{max-width:1240px}
@media(max-width:900px){.wrap{grid-template-columns:minmax(0,1fr)}}
.grid-wrap{max-width:100%}
'''
assert s.count(old)==1, p
s=s.replace(old,css)
tail='''<script>/* 把没包在滚动容器里的表格自动包起来，防止窄屏横向溢出 */
(function(){var f=function(){document.querySelectorAll("main table").forEach(function(t){
  var p=t.parentElement; if(p&&p.classList.contains("grid-wrap"))return;
  var d=document.c
…（截断，原长 1197 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ch00-intro.html ok
ch01-time-order.html ok
ch02-replication.html ok
ch03a-consensus-election.html ok
ch03b-log-replication.html ok
ch03c-production-raft.html ok
ch03d-faq.html ok
ch04a-transactions.html ok
ch00-intro.html                scrollW=440 clientW=400 OVERFLOW
ch01-time-order.html           scrollW=400 clientW=400 ok
ch02-replication.html          scrollW=400 clientW=400 ok
ch03a-consensus-election.html  scrollW=400 clientW=400 ok
ch03b-log-replication.html     scrollW=400 clientW=400 ok
ch03c-production-raft.html     scrollW=400 clientW=400 ok
ch03d-faq.html                 scrollW=400 clientW=400 ok
ch04a-transactions.html        scrollW=400 clientW=400 ok
````

</details>

▸ **Bash** Diagnose ch00 overflow

````
grep -c "<main>" ch00-intro.html; grep -o 'querySelector("main")' ch00-intro.html | head -1; node -e '
const {chromium}=require("〔路径00142〕");
(async()=>{const b=await chromium.launch({executablePath:"〔路径00178〕"});
const p=await b.newPage({viewport:{width:400,height:900}});
const errs=[];p.on("pageerror",e=>errs.push(e.message));
await p.goto("file://〔路径00071〕");await p.waitForTimeout(600);
console.log("errors:",errs);
const r=await p.evaluate(()=>{const W=document.documentElement.clientWidth,out=[];
 document.querySelectorAll("body *").forEach(e=>{const bb=e.getBoundingClientRect();
  if(bb.right>W+2){let ok=false,q=e.parentElement;
   while(q){const cs=getComputedStyle(q);if(["auto","scro
…（截断，原长 1286 字符）
````

<details><summary>↳ Bash 结果</summary>

````
1
querySelector("main")
errors: []
[
 {
  "w": 412,
  "t": " 术语一句话解释常见误解 / 补充 节点 node参与系统的一个独立进程（通常一",
  "anc": "TABLE < DIV.grid-wrap < SECTION#s9 < MAIN"
 },
 {
  "w": 412,
  "t": "术语一句话解释常见误解 / 补充",
  "anc": "THEAD < TABLE < DIV.grid-wrap < SECTION#s9"
 },
 {
  "w": 412,
  "t": "术语一句话解释常见误解 / 补充",
  "anc": "TR < THEAD < TABLE < DIV.grid-wrap"
 },
 {
  "w": 170,
  "t": "常见误解 / 补充",
  "anc": "TH < TR < THEAD < TABLE"
 }
]
````

</details>

▸ **Bash** Add overflow-x to injected grid-wrap rule

````
sed -i 's|^\.grid-wrap{max-width:100%}$|.grid-wrap{max-width:100%;overflow-x:auto}|' *.html && grep -c "grid-wrap{max-width:100%;overflow-x:auto}" *.html && node 〔路径00212〕
````

<details><summary>↳ Bash 结果</summary>

````
ch00-intro.html:1
ch01-time-order.html:1
ch02-replication.html:1
ch03a-consensus-election.html:1
ch03b-log-replication.html:1
ch03c-production-raft.html:1
ch03d-faq.html:1
ch04a-transactions.html:1
ch00-intro.html                scrollW=400 clientW=400 ok
ch01-time-order.html           scrollW=400 clientW=400 ok
ch02-replication.html          scrollW=400 clientW=400 ok
ch03a-consensus-election.html  scrollW=400 clientW=400 ok
ch03b-log-replication.html     scrollW=400 clientW=400 ok
ch03c-production-raft.html     scrollW=400 clientW=400 ok
ch03d-faq.html                 scrollW=400 clientW=400 ok
ch04a-transactions.html        scrollW=400 clientW=400 ok
````

</details>

▸ **Bash** Re-verify both pages and capture screenshots

````
node 〔路径00242〕 >/dev/null 2>&1; node 〔路径00254〕 2>&1 | tail -6; echo "=== 4A ==="; node 〔路径00253〕 2>&1 | head -6; ls -la 〔路径00248〕
````

<details><summary>↳ Bash 结果</summary>

````
D4 scenario C: clusters=2 leaders=1 effective=2
D4 scenario D: clusters=2 leaders=1 effective=2
D4 scenario E: clusters=1 leaders=1 effective=3
quiz items: 10
sections: 10
NO CONSOLE ERRORS
=== 4A ===
ReadView default → 300
after trx28 commits → 400
C=8 O=4 hot=0%  2PL[t=10 ab=0 waste=0 eff=undefined]  SSI[t=12 ab=2 waste=8 eff=undefined]
C=8 O=4 hot=50%  2PL[t=13 ab=0 waste=0 eff=undefined]  SSI[t=20 ab=6 waste=24 eff=undefined]
C=8 O=4 hot=100%  2PL[t=28 ab=3 waste=6 eff=undefined]  SSI[t=28 ab=14 waste=56 eff=undefined]
  cross: 70
-rw-r--r-- 1 root root 455936 Sep 14 10:13 〔路径00249〕
-rw-r--r-- 1 root root 289442 Sep 14 10:13 〔路径00250〕
-rw-r--r-- 1 root root 279529 Sep 14 10:13 〔路径00251〕
-rw-r--r-- 1 root root 443342 Sep 14 10:13 〔路径00252〕
````

</details>

▸ **Read** `〔路径00249〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

## 👤 你 · 2026-09-14 18:13

🖼️（附图，图片本身未缓存）

▸ **Read** `〔路径00252〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

## 👤 你 · 2026-09-14 18:13

🖼️（附图，图片本身未缓存）

▸ **Bash** Fix demo 4 title and legend

````
python3 - <<'PY'
p='ch04a-transactions.html'; s=open(p,encoding='utf-8').read()
a='<div class="t">实验 4 · 2PL vs SSI：冲突率一变，胜负就反转</div><div class="s">同一批事务，悲观锁 vs 乐观检测：两张不同的账单</div>'
b='<div class="t">实验 4 · 2PL vs SSI：两种并发控制，两张不同的账单</div><div class="s">一个付「等待」，一个付「白做的功」——把热点集中度拉满，看哪张账单涨得快</div>'
assert s.count(a)==1; s=s.replace(a,b)
a='<span><i style="width:11px;height:11px;border-radius:3px;background:var(--surface-2);border:1px solid var(--border);opacity:.4;display:inline-block"></i>不适用</span>\n'
assert s.count(a)==1; s=s.replace(a,'')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
node 〔路径00234〕
````

<details><summary>↳ Bash 结果</summary>

````
ok
matrix rows: 7
no errors
````

</details>

▸ **Bash** Write Go engine part 1

````
mkdir -p 〔路径00115〕 && cp 〔路径00114〕 〔路径00119〕 && cat > 〔路径00117〕 <<'GOEOF'
// 一个迷你事务引擎：MVCC 版本链 + ReadView + 行锁 + 范围锁 + 串行执行。
//
// 它存在的理由只有一个：让「哪个隔离级别挡得住哪种异常」这句话
// 从【背下来的表格】变成【跑出来的结果】。
// 加起来不到 300 行，但 InnoDB 在这件事上做的核心判断它都做了。
package main

import "sort"

// 隔离配置。RRLOCK 不是一个真的隔离级别，而是「RR + 显式 SELECT ... FOR UPDATE」，
// 因为它才是工程上真正用来堵写偏斜的那个办法。
const (
	RU     = "RU"
	RC     = "RC"
	RR     = "RR"
	RRLOCK = "RRLOCK"
	SER    = "SER"
)

var Levels = []string{RU, RC, RR, RRLOCK, SER}
var LevelName = map[string]string{
	RU: "READ UNCOMMITTED", RC: "READ COMMITTED", RR: "REPEATABLE READ",
	RRLOCK: "RR + FOR UPDATE", SER: "SERIALIZABLE",
}

type Version struct{ Trx, Val int }

// Pred 是一个范围条件，比如 on_call = tr
…（截断，原长 4722 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Go engine part 2

````
cat >> 〔路径00117〕 <<'GOEOF'

const locked = "blocked"

// step 执行一步，返回 "ok" 或 "blocked"（被锁挡住，需要排队重试）。
func (e *Engine) step(s *Op) string {
	x := e.tx[s.T]
	if s.Kind == "begin" {
		x = &Tx{T: s.T, ID: e.nextID}
		e.nextID++
		e.tx[s.T] = x
		e.active[x.ID] = true
		e.Trace = append(e.Trace, TraceEntry{s.T, "c", "BEGIN"})
		return "ok"
	}
	if x == nil {
		return "ok"
	}
	lockMode := e.Level == RRLOCK || e.Level == SER
	switch s.Kind {
	case "read":
		if s.Lockable && lockMode {
			if e.lockBusy(s.Key, x.ID) {
				return locked
			}
			e.xlock[s.Key] = x.ID
		}
		v, _ := e.readFor(x, s.Key, s.Lockable)
		x.Last = v
		x.Reads = append(x.Reads, v)
		q := "SELECT " + s.Key
		if s.Lockable && lockMode {
			q += " FOR UPDATE"
		}
		e.Trace = append(e.Trace, TraceEntry{s
…（截断，原长 5184 字符）
````

<details><summary>↳ Bash 结果</summary>

````
engine.go
(fmt ok)
````

</details>

▸ **Bash** Write Go scenarios

````
cat > scen.go <<'GOEOF'
package main

// 七种并发异常。前四种在 ANSI SQL-92 的清单里，后三种不在——
// 而后三种才是线上真正出事的那几种。
var (
	POnCall = &Pred{"oncall", "on_call = true", func(k string, v int) bool { return len(k) > 4 && k[:4] == "doc:" && v == 1 }}
	PAdult  = &Pred{"adult", "age > 18", func(k string, v int) bool { return len(k) > 2 && k[:2] == "p:" && v > 18 }}
	PRoom   = &Pred{"roomA", "room='A' AND slot=10", func(k string, v int) bool { return len(k) > 3 && k[:3] == "bk:" && v == 1 }}
)

type Scenario struct {
	ID, Name, En string
	ANSI         bool
	Init         map[string]int
	Steps        func() []*Op
	Check        func(e *Engine) (bad bool, detail string)
}

func op(t int, kind string) *Op { return &Op{T: t, Kind: kind} }

var Scenarios = []Scenario{
	{ID: "dw", Name: "脏写", En: "dirty write", ANSI: true
…（截断，原长 5335 字符）
````

<details><summary>↳ Bash 结果</summary>

````
VET OK
````

</details>

▸ **Bash** Write Go 2PL/SSI simulators

````
cat > cc.go <<'GOEOF'
package main

// 2PL 与 SSI 的对照模拟。
//
// 两个模拟器共用同一个物理前提：系统每 tick 最多能执行 PAR 个操作。
// 这一条很重要——它让「白做的功」真的变成时间，而不是一个漂亮的免费数字。
const PAR = 4

type CCOp struct {
	Key   int
	Write bool
}
type CCTxn struct {
	ID  int
	Ops []CCOp
}
type CCResult struct {
	Tick, Aborts, Wasted, Work int
	Tput                       float64
}

var nkey = 40

const hotKeys = 2

func MkTxns(c, o int, hot float64, rnd *Rand) []CCTxn {
	nkey = c * o // 键空间随负载伸缩，否则「低冲突」根本低不下来
	if nkey < 24 {
		nkey = 24
	}
	ts := make([]CCTxn, c)
	for i := 0; i < c; i++ {
		ops := make([]CCOp, o)
		for j := 0; j < o; j++ {
			k := rnd.Intn(nkey)
			if rnd.Float() < hot {
				k = rnd.Intn(hotKeys)
			}
			ops[j] = CCOp{Key: k, Write: rnd.Float() < 0.5}
		}
		ts[i] = CCTxn{ID: i, Ops: ops}
	}
	return ts
}

type lk struct
…（截断，原长 5734 字符）
````

<details><summary>↳ Bash 结果</summary>

````
VET OK
````

</details>

▸ **Bash** Write Go lab04a main and run

````
cat > main.go <<'GOEOF'
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

// ═════════════
…（截断，原长 9472 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：医生 2 人 ｜ 同时请假 2 人 ｜ 试验 200 次 ｜ 并发事务 12 × 4 操作 ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 4A-1 · 并发异常 × 隔离级别矩阵
  这张表不是抄来的，每一格都是迷你事务引擎真的跑了一遍
══════════════════════════════════════════════════════════════════════════════
    异常                                RU          RC          RR    RR+FOR UPD      SERIAL
    ────────────────────────────────────────────────────────────────────────────────────────
    脏写                          异常发生        挡住        挡住          挡住        挡住
    脏读                          异常发生        挡住        挡住          挡住        挡住
    不可重复读                    异常发生    异常发生        挡住          挡住        挡住
    幻读                          异常发生    异常发生        挡住          挡住        挡住
    丢失更新              *       异常发生    异常发生    异常发生          挡住        挡住
    写偏斜                *       异常发生    异常发生    异常发生          挡住        挡住
    写偏斜（插入型）      *       异常发生    异常发生    异常发生          挡住        挡住

    * = 不在 ANSI SQL-92 的异常清单里 —— 而它们恰恰是线上真正出事的那几种。

    ── 丢失更新 × REPEATABLE READ ──────────────────────────────
      T1 BEGIN
      T2                                             BEGIN
      T1 SELECT x → 100
      T2                                             SELECT x → 100
      T1 UPDATE x = 110
      T2                                             （被锁阻塞，等待前一个事务提交）
      T1 COMMIT
      T2                                             UPDATE x = 110
      T2                                             COMMIT

…（截断，原长 2439 字符）
````

</details>

▸ **Bash** View rest of Go lab04a output

````
go run . 2>&1 | sed -n '50,175p'
````

<details><summary>↳ Bash 结果</summary>

````
      T2                                             UPDATE doc:bob = 0
      T1 COMMIT
      T2                                             COMMIT
      ⇒ 最终在岗医生数 0（不变量要求 >= 1）　★ 不变量被破坏

    ── 写偏斜 × RR + FOR UPDATE ──────────────────────────────
      T1 BEGIN
      T2                                             BEGIN
      T1 SELECT count(*) WHERE on_call = true FOR UPDATE → 2
      T2                                             （被锁阻塞，等待前一个事务提交）
      T1 UPDATE doc:alice = 0
      T1 COMMIT
      T2                                             SELECT count(*) WHERE on_call = true FOR UPDATE → 1
      T2                                             if (上次读到 1 >= 2) 不成立 → 放弃写入
      T2                                             COMMIT
      ⇒ 最终在岗医生数 1（不变量要求 >= 1）　正确

    结论
      · 看「丢失更新」和「写偏斜」两行：RR 那一列是红的。
        把 MySQL 从 RC 调到 RR，你买到的是「不可重复读」和「幻读」，
        【没有】买到「丢失更新」和「写偏斜」——而后两者才是业务不变量被破坏的那一类。
      · RR 和 RR+FOR UPDATE 的全部差别，是把隐式的读变成了显式的锁。
        数据库看不见你代码里那个 if，加锁是唯一能让它看见的办法。

══════════════════════════════════════════════════════════════════════════════
  Lab 4A-2 · MVCC 的 ReadView 可见性判定
  RC 和 RR 的实现差异只有一行：ReadView 什么时候创建
══════════════════════════════════════════════════════════════════════════════

    同一条版本链、同一套四步判定算法，只改「ReadView 何时创建」这一件事。

    balance 这一行的版本链（新 → 旧）：  [trx 32 = 500]  →  [trx 28 = 400]  →  [trx 24 = 300]  →  [trx 18 = 200]  →  [trx 12 = 100]

    ① 事务 25 建 ReadView 时，trx 28 和 32 都还没提交
      ReadView{ m_ids=[28 32]  min=28  max=40  creator=25
…（截断，原长 5986 字符）
````

</details>

▸ **Bash** Fix visibility table alignment

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
a='''			fmt.Printf("      trx %-3d = %-4d%s   %s\\n", v.Trx, v.Val, mark, why)'''
b='''			fmt.Printf("      trx %-3d = %-4d %s  %s\\n", v.Trx, v.Val, tui.PadR(mark, 8), why)'''
assert s.count(a)==1; s=s.replace(a,b)
s=s.replace('mark := "  不可见"','mark := "不可见"')
s=s.replace('mark = "  ★可见"','mark = "★ 可见"')
s=s.replace('mark = "  可见"','mark = "可见"')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
gofmt -w . && go run . 2>&1 | sed -n '86,102p'
````

<details><summary>↳ Bash 结果</summary>

````
ok
      trx 28  = 400  不可见    trx 在 m_ids 里 ⇒ 我建 ReadView 时它还没提交
      trx 24  = 300  ★ 可见    trx < min(28) ⇒ 建 ReadView 前就已提交　← 第一个可见的，就是答案
      trx 18  = 200  可见      trx < min(28) ⇒ 建 ReadView 前就已提交
      trx 12  = 100  可见      trx < min(28) ⇒ 建 ReadView 前就已提交

    ② trx 28 提交了 —— RC 会在下一条语句重新采集活跃集合
      ReadView{ m_ids=[32]  min=32  max=40  creator=25 }
      trx 32  = 500  不可见    trx 在 m_ids 里 ⇒ 我建 ReadView 时它还没提交
      trx 28  = 400  ★ 可见    trx < min(32) ⇒ 建 ReadView 前就已提交　← 第一个可见的，就是答案
      trx 24  = 300  可见      trx < min(32) ⇒ 建 ReadView 前就已提交
      trx 18  = 200  可见      trx < min(32) ⇒ 建 ReadView 前就已提交
      trx 12  = 100  可见      trx < min(32) ⇒ 建 ReadView 前就已提交

    结论
      · 同一条版本链，仅仅因为「活跃事务集合」采集的时刻不同，
        返回值就从 300 跳到了 400。
      · RC：每条语句都新建 ReadView ⇒ 上面两种情况会在同一个事务里先后出现 ⇒ 不可重复读。
````

</details>

▸ **Bash** Write Java Lab04A engine

````
mkdir -p 〔路径00140〕 && cat > 〔路径00141〕 <<'JEOF'
/*
 * Lab 4-A · 单机事务：隔离级别、MVCC 与写偏斜（Java 版）
 *
 * 运行：  cd java/lab04a && java Lab04A.java
 * 调参：  java Lab04A.java -trials 500 -doctors 3 -txns 12 -ops 4
 *
 * 需要 JDK 17+。配套课件：courseware/ch04a-transactions.html
 *
 * 里面那个迷你事务引擎（MVCC 版本链 + ReadView + 行锁 + 范围锁 + 串行执行）
 * 存在的理由只有一个：让「哪个隔离级别挡得住哪种异常」这句话
 * 从【背下来的表格】变成【跑出来的结果】。
 *
 * 输出与 Go 版逐字对应（连随机数都用同一个 xorshift32），可以直接 diff 对照。
 */
import java.util.*;
import java.util.function.BiPredicate;

public class Lab04A {

    static int TRIALS=200, DOCTORS=2, LEAVERS=2, TXNS=12, OPS=4;
    static long SEED=42;

    public static void main(String[] args){
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out)
…（截断，原长 12618 字符）
````

<details><summary>↳ Bash 结果</summary>

````
part1 done
262 〔路径00141〕
````

</details>

▸ **Bash** Write Java scenarios and labs 1-2

````
cat >> 〔路径00141〕 <<'JEOF'

    // ══════════════════════════════════════════════════════════════════
    // 七种并发异常。前四种在 ANSI SQL-92 的清单里，后三种不在——
    // 而后三种才是线上真正出事的那几种。
    // ══════════════════════════════════════════════════════════════════
    static final Pred P_ONCALL=new Pred("oncall","on_call = true",(k,v)->k.startsWith("doc:")&&v==1);
    static final Pred P_ADULT =new Pred("adult","age > 18",       (k,v)->k.startsWith("p:")&&v>18);
    static final Pred P_ROOM  =new Pred("roomA","room='A' AND slot=10",(k,v)->k.startsWith("bk:")&&v==1);

    interface Check { Object[] run(Engine e); }   // [Boolean bad, String detail]
    record Scen(String id,String name,String en,boolean ansi,
                Map<String,Integer> init, java.util.function.Supplier<Li
…（截断，原长 9325 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Java lab04a and run

````
cat >> 〔路径00141〕 <<'JEOF'

    // ══════════════════════════════════════════════════════════════════
    // Lab 4A-3：写偏斜 —— 医院值班表
    // ══════════════════════════════════════════════════════════════════
    static List<Op> skewSteps(int c, Rnd rnd){
        List<List<Op>> per=new ArrayList<>();
        for(int t=1;t<=c;t++) per.add(List.of(
            o(t,"begin"),
            o(t,"scan").p(P_ONCALL).lock(),
            o(t,"write").k("doc:"+t).v(0).gmin(2),
            o(t,"commit")));
        int[] idx=new int[c]; List<Op> out=new ArrayList<>(); int left=4*c;
        while(left>0){
            List<Integer> avail=new ArrayList<>();
            for(int i=0;i<c;i++) if(idx[i]<per.get(i).size()) avail.add(i);
            int p=avail.get(rnd.intn(avail.size()
…（截断，原长 13069 字符）
````

<details><summary>↳ Bash 结果</summary>

````
Picked up JAVA_TOOL_OPTIONS: -Djavax.net.ssl.trustStore=/root/.ccr/java-truststore.p12 -Djavax.net.ssl.trustStorePassword=changeit -Djavax.net.ssl.trustStoreType=PKCS12 -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=33367 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|127.*|0.*|::|169.254.*|api.anthropic.com|api-staging.anthropic.com|api-pr-preview.anthropic.com|mcp-proxy.anthropic.com|mcp-proxy-staging.anthropic.com|registry.npmjs.org|jsr.io|npm.jsr.io|pypi.org|files.pythonhosted.org|index.crates.io|proxy.golang.org|host.docker.internal|10.*|172.16.*|172.17.*|172.18.*|172.19.*|172.20.*|172.21.*|172.22.*|172.23.*|172.24.*|172.25.*|172.26.*|172.27.*|172.28.*|172.29.*|172.30.*|172.31.*|192.168.*|100.64.0.0/10|*.svc.cluster.local|*.svc.cluster.local -Djdk.http.auth.tunneling.disabledSchemes= -Djdk.http.auth.proxying.disabledSchemes=

配置：医生 2 人 ｜ 同时请假 2 人 ｜ 试验 200 次 ｜ 并发事务 12 × 4 操作 ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 4A-1 · 并发异常 × 隔离级别矩阵
  这张表不是抄来的，每一格都是迷你事务引擎真的跑了一遍
══════════════════════════════════════════════════════════════════════════════
    异常                                RU          RC          RR    RR+FOR UPD      SERIAL
    ────────────────────────────────────────────────────────────────────────────────────────
    脏写                          异常发生        挡住        挡住          挡住        挡住
    脏读                          异常发生        挡住        挡住          挡住        挡住
    不可重复读                    异常发生    异常发生
…（截断，原长 2338 字符）
````

</details>

▸ **Bash** Diff Go vs Java output

````
java Lab04A.java > 〔路径00208〕 2>/dev/null; cd 〔路径00115〕 && go run . > 〔路径00204〕 2>&1; diff 〔路径00204〕 〔路径00208〕 && echo "=== IDENTICAL (Go == Java) ===" || echo "--- differs ---"
````

<details><summary>↳ Bash 结果</summary>

````
=== IDENTICAL (Go == Java) ===
````

</details>

▸ **Bash** Test flags in both languages

````
go run . -trials 300 -doctors 4 -leavers 4 -txns 16 -ops 6 > 〔路径00205〕 2>&1 && cd 〔路径00140〕 && java Lab04A.java -trials 300 -doctors 4 -leavers 4 -txns 16 -ops 6 > 〔路径00209〕 2>/dev/null; diff 〔路径00205〕 〔路径00209〕 && echo "FLAGS IDENTICAL" ; echo "--- sample ---"; sed -n '/4A-3/,/结论/p' 〔路径00205〕 | head -22; sed -n '/4A-4/,/结论/p' 〔路径00205〕 | tail -14
````

<details><summary>↳ Bash 结果</summary>

````
145c145
<     60%                      84          10            17         15%             78          27           162         63%
---
>     60%                      70           6            11         10%             78          27           162         63%
--- sample ---
  Lab 4A-3 · 写偏斜：医院值班表
  不变量「任何时刻至少 1 个医生在岗」，看谁守得住
══════════════════════════════════════════════════════════════════════════════

    4 个医生在岗，4 个人同时发起请假。每个人的事务都是：
      BEGIN; SELECT count(*) WHERE on_call=true;  if (>= 2) UPDATE 自己 = 不在岗; COMMIT
    每次试验随机交错这些步骤，跑 300 次。
    隔离配置                        不变量被破坏    平均最终在岗      出现过锁等待      主动放弃请假
    ────────────────────────────────────────────────────────────────────────────────────────────────
    READ COMMITTED                     238 / 300         0.21 人           0 / 300             62 次
    REPEATABLE READ                    238 / 300         0.21 人           0 / 300             62 次
    RR + FOR UPDATE                      0 / 300         1.00 人         298 / 300            300 次
    SERIALIZABLE                         0 / 300         1.00 人           0 / 300            300 次

    结论

    16 个并发事务，每个 6 个操作，系统每 tick 最多执行 4 个操作。
      2PL：拿不到锁就等（等待不占 CPU，做过的功不浪费），等待图成环 = 死锁中止。
      SSI：谁也不等，全速跑完再验证；验证不过就整个事务从头再来（白做的功照样烧 CPU）。
    热点               2PL 耗时        中止      白做操作    白做占比       SSI 耗时        中止      白做操作    白做占比
    ──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
    0%
…（截断，原长 2238 字符）
````

</details>

▸ **Bash** Make Go deterministic and check

````
cd 〔路径00115〕 && python3 - <<'PY'
p='cc.go'; s=open(p,encoding='utf-8').read()
a='''			var cyc []int
			for a := range waitFor {'''
b='''			var cyc []int
			// Go 的 map 遍历顺序是随机的，这里必须排序，
			// 否则同一个种子跑两次结果都可能不同（Java 版也就对不上了）
			wkeys := make([]int, 0, len(waitFor))
			for a := range waitFor {
				wkeys = append(wkeys, a)
			}
			sort.Ints(wkeys)
			for _, a := range wkeys {'''
assert s.count(a)==1; s=s.replace(a,b)
s=s.replace('package main\n','package main\n\nimport "sort"\n',1)
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
gofmt -w . && go vet ./... && for i in 1 2 3; do go run . -trials 300 -doctors 4 -leavers 4 -txns 16 -ops 6 | md5sum; done
````

<details><summary>↳ Bash 结果</summary>

````
ok
57e0f5df12fa9d42491ef4a6f3b23fbe  -
57e0f5df12fa9d42491ef4a6f3b23fbe  -
57e0f5df12fa9d42491ef4a6f3b23fbe  -
````

</details>

▸ **Bash** Verify Go/Java parity on default and flags

````
go run . -trials 300 -doctors 4 -leavers 4 -txns 16 -ops 6 > 〔路径00205〕 2>&1 && go run . > 〔路径00204〕 2>&1 && cd 〔路径00140〕 && java Lab04A.java -trials 300 -doctors 4 -leavers 4 -txns 16 -ops 6 > 〔路径00209〕 2>/dev/null && java Lab04A.java > 〔路径00208〕 2>/dev/null; diff 〔路径00204〕 〔路径00208〕 && diff 〔路径00205〕 〔路径00209〕 && echo "=== BOTH IDENTICAL (default + flags) ==="
````

<details><summary>↳ Bash 结果</summary>

````
=== BOTH IDENTICAL (default + flags) ===
````

</details>

▸ **Bash** Update README for Part 4-A

````
cd 〔路径00054〕 && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
subs=[
('open courseware/ch03d-faq.html               # Part 3-D：附录（四个 index / Quorum vs Raft / 配置与冷启动）\n',
 'open courseware/ch03d-faq.html               # Part 3-D：附录（四个 index / Quorum vs Raft / 配置与冷启动）\n'
 'open courseware/ch04a-transactions.html      # Part 4-A：单机事务（隔离级别 / MVCC / 写偏斜）\n'),
('cd go/lab03d && go run .     # Lab 3-D：ReadIndex 第三步 / no-op / Quorum vs Raft / 冷启动\n',
 'cd go/lab03d && go run .     # Lab 3-D：ReadIndex 第三步 / no-op / Quorum vs Raft / 冷启动\n'
 'cd go/lab04a && go run .     # Lab 4-A：异常矩阵 / ReadView / 写偏斜 / 2PL vs SSI\n'),
('cd java/lab03d && java Lab03D.java\n','cd java/lab03d && java Lab03D.java\ncd java/lab04a && java Lab04A.java\n'),
('| 4 | 分布式事务 —— 隔离级别、MVCC、2PC/XA、TCC、S
…（截断，原长 3275 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README updated
20:open courseware/ch04a-transactions.html      # Part 4-A：单机事务（隔离级别 / MVCC / 写偏斜）
31:cd go/lab04a && go run .     # Lab 4-A：异常矩阵 / ReadView / 写偏斜 / 2PL vs SSI
41:cd java/lab04a && java Lab04A.java
93:| **4-A** | **单机事务：隔离级别、MVCC 与写偏斜** —— ACID 里哪个字母是假的、七种并发异常、ReadView 的四步判定、Next-Key Lock、快照隔离的盲区、2PL vs SSI | 交互课件 + Lab 4-A | ✅ 已完成 |
116:│   └── ch04a-transactions.html  # Part 4-A 课件（异常矩阵 + ReadView 判定器 + 写偏斜 + 2PL/SSI）
150:│   └── lab04a/
163:    └── lab04a/Lab04A.java
274:## Lab 4-A 会让你亲手复现什么
278:| **4A-1** 异常 × 隔离级别矩阵 | 七种并发异常 × 五种隔离配置，**每一格都真的跑一遍**，不是抄表。结果是一个漂亮的阶梯——但**「丢失更新」和「写偏斜」两行在 RR 那一列仍然是红的** |
279:| **4A-2** ReadView 可见性 | 同一条版本链、同一套四步判定，只把活跃事务集合的采集时刻挪一下，返回值就从 **300 跳到 400**。*这就是 RC 和 RR 的全部实现差异* |
280:| **4A-3** 写偏斜 | 医院值班表跑 200 次：RC **120/200** 破坏不变量，RR **120/200**（一模一样），加 `FOR UPDATE` 后 **0/200**，串行化 **0/200** |
281:| **4A-4** 2PL vs SSI | 热点集中度从 0 拉到 100%：**2PL 的「白做占比」几乎不涨（0% → 15%），SSI 一路涨到 71%**，而两者墙上时间相差不大 |
283:4A-1 的矩阵是这一章的眼睛。它把一句话变成了可验证的事实：**把 MySQL 从 RC 调到 RR，你买到的是「不可重复读」和「幻读」，没有买到「丢失更新」和「写偏斜」**——而后两者才是业务不变量被破坏的那一类。
288:go run . -trials 500          # 4A-3 的并发试验次数
289:go run . -doctors 4 -leavers 4 # 4A-3 在岗医生数与同时请假人数
290:go run . -txns 16 -ops 6      # 4A-4 并发事务数与每事务操作数
````

</details>

▸ **Bash** Check default 4A-4 numbers

````
sed -n '/4A-4/,$p' 〔路径00204〕 | sed -n '/热点/,/结论/p' | head -12
````

<details><summary>↳ Bash 结果</summary>

````
    热点               2PL 耗时        中止      白做操作    白做占比       SSI 耗时        中止      白做操作    白做占比
    ──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
    0%                       15           0             0          0%             16           2             8         14%
    20%                      15           0             0          0%             24           8            32         40%
    40%                      30           2             4          8%             32          13            52         52%
    60%                      29           2             6         11%             36          18            72         60%
    80%                      22           0             0          0%             28          11            44         48%
    100%                     43           3             8         14%             44          26           104         68%

    结论
      · 看两列「白做占比」：热点集中度往右拉，2PL 的几乎不涨，SSI 的一路涨到有效功的几倍。
        而两者的【墙上时间】相差不大——这正是这个对比的重点：
````

</details>

▸ **Bash** Correct README numbers

````
python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
a='| **4A-4** 2PL vs SSI | 热点集中度从 0 拉到 100%：**2PL 的「白做占比」几乎不涨（0% → 15%），SSI 一路涨到 71%**，而两者墙上时间相差不大 |'
b='| **4A-4** 2PL vs SSI | 热点集中度从 0 拉到 100%：**2PL 的「白做占比」几乎不涨（0% → 14%），SSI 从 14% 一路涨到 68%**，而两者墙上时间相差不大（43 vs 44 ticks）——*SSI 省下来的等待，是用真金白银的 CPU 换的* |'
assert s.count(a)==1; s=s.replace(a,b)
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
ls courseware/ go/lab04a/ java/lab04a/
````

<details><summary>↳ Bash 结果</summary>

````
ok
courseware/:
ch00-intro.html
ch01-time-order.html
ch02-replication.html
ch03a-consensus-election.html
ch03b-log-replication.html
ch03c-production-raft.html
ch03d-faq.html
ch04a-transactions.html

go/lab04a/:
cc.go
engine.go
main.go
rand.go
scen.go

java/lab04a/:
Lab04A.java
````

</details>

## 🤖 Claude · 2026-09-14 18:22

📎 **发送文件**：Part 4-A 课件：异常 × 隔离级别矩阵（每格都由页内的迷你事务引擎实跑）、ReadView 判定器、写偏斜统计、2PL vs SSI 对照，10 道自测题。
- `〔路径00078〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00078〕 → file_uuid: 〔标识02465〕
````

</details>

📎 **发送文件**：Lab 4-A 的 Go / Java 源码与更新后的 README。
- `〔路径00117〕`
- `〔路径00120〕`
- `〔路径00116〕`
- `〔路径00118〕`
- `〔路径00119〕`
- `〔路径00141〕`
- `〔路径00056〕`

<details><summary>↳ SendUserFile 结果</summary>

````
7 files delivered to user.
  〔路径00117〕 → file_uuid: 〔标识01005〕
  〔路径00120〕 → file_uuid: 〔标识01495〕
  〔路径00116〕 → file_uuid: 〔标识01734〕
  〔路径00118〕 → file_uuid: 〔标识03224〕
  〔路径00119〕 → file_uuid: 〔标识00401〕
  〔路径00141〕 → file_uuid: 〔标识02581〕
  〔路径00056〕 → file_uuid: 〔标识02337〕
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00146〕\" && (which python3 || which perl) && md5sum *.html 2>/dev/null | head -3"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
/usr/bin/python3
6da4266d94ae5cd1c79312d5c99ce78d  ch00-intro.html
7a2e5d7cddc8ca0f6a69e424f4a6bea6  ch01-time-order.html
76c052f86d8589ea69e1d352a2cca1a3  ch02-replication.html
````

</details>

▸ **Bash** Compute pre-patch checksums to verify remote files match

````
cd 〔路径00069〕 && git init -q 2>/dev/null; md5sum ch00-intro.html ch01-time-order.html ch02-replication.html; echo "--- pre-patch md5 (reconstruct) ---"; for f in ch00-intro.html ch01-time-order.html ch02-replication.html ch03a-consensus-election.html ch03b-log-replication.html ch03c-production-raft.html; do
python3 - "$f" <<'PY'
import sys,hashlib
p=sys.argv[1]; s=open(p,encoding='utf-8').read()
css='''<style>/* fix: grid blowout —— grid 子项默认 min-width:auto 会撑破窄屏 */
main{min-width:0}
.wrap{max-width:1240px}
@media(max-width:900px){.wrap{grid-template-columns:minmax(0,1fr)}}
.grid-wrap{max-width:100%;overflow-x:auto}
@media(max-width:700px){
  .pstep,.lrung,.axm,.idxrow,.kv{grid-template-columns:minmax(0,1fr)!important;gap:8px!important}
}
</style>
'''
tail='''<script>/* 把没
…（截断，原长 1388 字符）
````

<details><summary>↳ Bash 结果</summary>

````
43476e1d1e1a47482dc5d97f2430e2ae  ch00-intro.html
059ee032bb701c4801b7a2c89bc3bb04  ch01-time-order.html
17f93f2f30045b9b2510f17d3fdc5671  ch02-replication.html
--- pre-patch md5 (reconstruct) ---
ch00-intro.html 6da4266d94ae5cd1c79312d5c99ce78d
ch01-time-order.html 7a2e5d7cddc8ca0f6a69e424f4a6bea6
ch02-replication.html 76c052f86d8589ea69e1d352a2cca1a3
ch03a-consensus-election.html d2f659edcd934b6cdea2ec5cc3ff256c
ch03b-log-replication.html 1be8995b3b80d0d0103faa21c4597ce4
ch03c-production-raft.html 4866c3da1941e26bc527d757279f0676
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00146〕\" && cat > 〔路径00226〕 <<'PY'\nimport sys,glob\ncss='''<style>/* fix: grid blowout —— grid 子项默认 min-width:auto 会撑破窄屏 */\nmain{min-width:0}\n.wrap{max-width:1240px}\n@media(max-width:900px){.wrap{grid-template-columns:minmax(0,1fr)}}\n.grid-wrap{max-width:100%;overflow-x:auto}\n@media(max-width:700px){\n  .pstep,.lrung,.axm,.idxrow,.kv{grid-template-columns:minmax(0,1fr)!important;gap:8px!important}\n}\n</style>\n</head>'''\ntail='''<script>/* 把没包在滚动容器里的表格自动包起来，防止窄屏横向溢出 */\n(function(){var f=function(){document.querySelectorAll(\"main table\").forEach(function(t){\n  var p=t.parentElement; if(p&&p.classList.contains(\"grid-wrap\"))return;\n  var d=document.createElement(\"div\"); d.className=\"grid-wrap\"; p.ins
…（截断，原长 1388 字符）
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ch00-intro.html patched
ch01-time-order.html patched
ch02-replication.html patched
ch03a-consensus-election.html patched
ch03b-log-replication.html patched
ch03c-production-raft.html patched
ch03d-faq.html patched
--- md5 ---
43476e1d1e1a47482dc5d97f2430e2ae  ch00-intro.html
059ee032bb701c4801b7a2c89bc3bb04  ch01-time-order.html
17f93f2f30045b9b2510f17d3fdc5671  ch02-replication.html
53281be4c3c24ca452dceb0ce8c479ec  ch03a-consensus-election.html
913e28caf2cf4a790c53628e033c2460  ch03b-log-replication.html
489bce20018c3b785258792cbb18cd44  ch03c-production-raft.html
0b7179a1d342b8131b438404014c61ee  ch03d-faq.html
````

</details>

▸ **Bash** Compare local checksums with device

````
md5sum ch0*.html
````

<details><summary>↳ Bash 结果</summary>

````
43476e1d1e1a47482dc5d97f2430e2ae  ch00-intro.html
059ee032bb701c4801b7a2c89bc3bb04  ch01-time-order.html
17f93f2f30045b9b2510f17d3fdc5671  ch02-replication.html
53281be4c3c24ca452dceb0ce8c479ec  ch03a-consensus-election.html
913e28caf2cf4a790c53628e033c2460  ch03b-log-replication.html
489bce20018c3b785258792cbb18cd44  ch03c-production-raft.html
0b7179a1d342b8131b438404014c61ee  ch03d-faq.html
8068bf1ff796e3a7578b89b70a52460b  ch04a-transactions.html
````

</details>

💾 **写入你的 Mac**：
- `〔路径00012〕`
- `〔路径00040〕`
- `〔路径00043〕`
- `〔路径00039〕`
- `〔路径00041〕`
- `〔路径00042〕`
- `〔路径00051〕`
- `〔路径00004〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00012〕","〔路径00040〕","〔路径00043〕","〔路径00039〕","〔路径00041〕","〔路径00042〕","〔路径00051〕","〔路径00004〕"],"rejected":[]}
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00144〕\" && md5sum courseware/*.html && echo \"--- tree ---\" && ls go/ java/ && echo \"--- lab04a ---\" && ls go/lab04a java/lab04a && echo \"--- README 路线 ---\" && grep -n \"已完成\\|⏳\" README.md | head -14"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
43476e1d1e1a47482dc5d97f2430e2ae  courseware/ch00-intro.html
059ee032bb701c4801b7a2c89bc3bb04  courseware/ch01-time-order.html
17f93f2f30045b9b2510f17d3fdc5671  courseware/ch02-replication.html
53281be4c3c24ca452dceb0ce8c479ec  courseware/ch03a-consensus-election.html
913e28caf2cf4a790c53628e033c2460  courseware/ch03b-log-replication.html
489bce20018c3b785258792cbb18cd44  courseware/ch03c-production-raft.html
0b7179a1d342b8131b438404014c61ee  courseware/ch03d-faq.html
8068bf1ff796e3a7578b89b70a52460b  courseware/ch04a-transactions.html
--- tree ---
go/:
go.mod
internal
lab00
lab01
lab02
lab03a
lab03b
lab03c
lab03d
lab04a

java/:
lab00
lab01
lab02
lab03a
lab03b
lab03c
lab03d
lab04a
--- lab04a ---
go/lab04a:
cc.go
engine.go
main.go
rand.go
scen.go

java/lab04a:
Lab04A.java
--- README 路线 ---
86:| **0** | **序言与心智模型** —— 三个免费假设的崩塌、三个不可靠、故障模型、两将军、幂等性 | 交互课件 + Lab 0 | ✅ 已完成 |
87:| **1** | **时间、顺序与因果** —— happens-before、Lamport 时钟、向量时钟、LWW 陷阱、HLC、TrueTime | 交互课件 + Lab 1 | ✅ 已完成 |
88:| **2** | **复制与一致性模型** —— 复制拓扑、三大延迟异常、一致性谱系、CAP/PACELC、Quorum、CRDT | 交互课件 + Lab 2 | ✅ 已完成 |
89:| **3-A** | **共识与 Raft 选举** —— 共识定义、FLP、2PC 为什么不是共识、任期、选举规则、随机化 | 交互课件 + Lab 3-A | ✅ 已完成 |
90:| **3-B** | **日志复制与安全性** —— 日志匹配性质、提交规则、Figure 8、五条安全性属性、快速回退 | 交互课件 + Lab 3-B | ✅ 已完成 |
91:| **3-C** | **持久化 / 快照 / 成员变更 / 线性一致读** —— 三样必落盘的状态、InstallSnapshot、单节点变更的交集证明、ReadIndex 与 Lease Read | 交互课件 + Lab 3-C | ✅ 已完成 |
92:| **3-D** | **附录：算法与系统的接缝** —— 四个 index 与 ReadIndex 的第三步、no-op 的必要性、Quorum ≠ Raft、配置从哪来与冷启动、Raft
…（截断，原长 2031 字符）
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #25 status
````

</details>

两章都做完了，都在 `~/Code/distributed-systems-course/`。

**Part 3-D 附录**（你那三个问题的沉淀）：四个 index 的流水线可视化——手动推进 commit 和 apply 把窗口拉开，然后关掉 ReadIndex 第三步读一次，它会直接告诉你丢了几条已提交的写；no-op 场景重放（不发 no-op 时 ReadIndex 三步全通过，照样读到旧值）；Quorum vs Raft 并发写对照沙盒；冷启动沙盒五种参数组合。Lab 3-D 跑出的关键数字：apply 滞后 4 条时跳过第三步 **387/500 次陈旧读**，不发 no-op **16/40 个种子违规**，Dynamo 的分歧数在 W 从 1 调到 5 的过程中**一动不动**（写成功数倒是从 2.00 掉到 0.70）。

**Part 4-A 单机事务**：核心是那张 7×5 的异常矩阵，**每一格都由页内的迷你事务引擎实跑**（MVCC 版本链、ReadView 四步判定、行锁、间隙锁、串行执行都实现了），点任意格子看完整时序。结果是个漂亮的阶梯，但**「丢失更新」和「写偏斜」两行在 RR 那一列是红的**——医院值班表 200 次试验，RC 违规 120 次，RR 也是 120 次，一模一样；加 `FOR UPDATE` 后 0 次。另外 2PL vs SSI 那个实验我没按教科书写「高冲突 2PL 反超」，因为模型跑出来不支持：真实的差别是 **2PL 的白做占比几乎不涨（0%→14%），SSI 从 14% 涨到 68%，而墙上时间只差 1 tick**。

Go 和 Java 两版输出**逐字节相同**（共用同一个 8 行 xorshift32，可以直接 `diff`）。顺手修了全部 8 个课件页在手机宽度下的横向溢出。

---

下一站 **Part 4-B**（2PC/XA、TCC、Saga、本地消息表、事务消息）之前，先给你两道热身：

1. 4-A 里说「能变成约束的不变量，一律变成约束」。**跨服务的分布式事务里，这条还成立吗？**比如订单服务和库存服务各有自己的库,「下单不能超卖」这个不变量能变成谁的约束？
2. 2PC 的参与者在 prepare 之后、收到 commit 之前宕机了，重启后它该怎么办——**它能自己决定吗**？如果不能，它要去问谁，问不到又该怎么办？

> ✓ 本轮结束 · 2026-09-14 18:23 · 135 步

> ⚠️ **缺失片段**：序号 3660–3892 的事件不在本地缓存中。
