## 👤 你 · 2026-09-14 17:16

1. 都不用跳过，作为查缺补漏了，可重复读可以防住幻读因为MySQL会使用其MVCC进行事务对应版本快照的读，其通过redo log中记录对应事务所处版本来锁定对应版本事务应该读到的修改有哪些，超出这个版本的修改对于这个事务不可见。
2. 第3章主要讲了raft但是我怎么感觉raft好像读写都是leader负责的？虽然你讲了follower look但是好像这是特定产品的特性，是因为raft作为共识算法主要是保证一致性对于读性能没有特别优化？然后quorum读可以算是一种相对更一致的但是不止依赖leader的读方式？这样会不会对集群吞吐造成比较大的负担？去quorum写又是什么？就是raft的日志append吗？
3. 还有raft的配置是如何启用的？比如每次集群增减机器都会重新选举吗？配置是从leader下发的吗？如果是冷启动怎么办？怎么知道有几个机器，机器都怎么访问到？
4. raft做的都是日志的记录，这个是不是和redo log类似？但是对于想kv和DB这种方式是不是代表执行日志本身还有延迟，真的涉及到读对应内容时是怎么确定具体内容已经和版本对应上？

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

## 👤 你 · 2026-09-14 17:38

🖼️（附图，图片本身未缓存）

## 🤖 Claude · 2026-09-14 17:49

📎 **发送文件**：Part 3-D 附录课件：四个 index 流水线 + no-op 重放 + Quorum/Raft 对照沙盒 + 冷启动沙盒，10 道自测题。
- `〔路径00077〕`

💾 **写入你的 Mac**：
- `〔路径00011〕`
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

💾 **写入你的 Mac**：
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

📎 **发送文件**：Lab 3-D 的 Go / Java 源码与更新后的 README。
- `〔路径00112〕`
- `〔路径00113〕`
- `〔路径00114〕`
- `〔路径00139〕`
- `〔路径00056〕`

💾 **写入你的 Mac**：
- `〔路径00036〕`
- `〔路径00037〕`
- `〔路径00038〕`
- `〔路径00050〕`
- `〔路径00004〕`

## 👤 你 · 2026-09-14 18:13

🖼️（附图，图片本身未缓存）

## 👤 你 · 2026-09-14 18:13

🖼️（附图，图片本身未缓存）

## 🤖 Claude · 2026-09-14 18:22

📎 **发送文件**：Part 4-A 课件：异常 × 隔离级别矩阵（每格都由页内的迷你事务引擎实跑）、ReadView 判定器、写偏斜统计、2PL vs SSI 对照，10 道自测题。
- `〔路径00078〕`

📎 **发送文件**：Lab 4-A 的 Go / Java 源码与更新后的 README。
- `〔路径00117〕`
- `〔路径00120〕`
- `〔路径00116〕`
- `〔路径00118〕`
- `〔路径00119〕`
- `〔路径00141〕`
- `〔路径00056〕`

💾 **写入你的 Mac**：
- `〔路径00012〕`
- `〔路径00040〕`
- `〔路径00043〕`
- `〔路径00039〕`
- `〔路径00041〕`
- `〔路径00042〕`
- `〔路径00051〕`
- `〔路径00004〕`

两章都做完了，都在 `~/Code/distributed-systems-course/`。

**Part 3-D 附录**（你那三个问题的沉淀）：四个 index 的流水线可视化——手动推进 commit 和 apply 把窗口拉开，然后关掉 ReadIndex 第三步读一次，它会直接告诉你丢了几条已提交的写；no-op 场景重放（不发 no-op 时 ReadIndex 三步全通过，照样读到旧值）；Quorum vs Raft 并发写对照沙盒；冷启动沙盒五种参数组合。Lab 3-D 跑出的关键数字：apply 滞后 4 条时跳过第三步 **387/500 次陈旧读**，不发 no-op **16/40 个种子违规**，Dynamo 的分歧数在 W 从 1 调到 5 的过程中**一动不动**（写成功数倒是从 2.00 掉到 0.70）。

**Part 4-A 单机事务**：核心是那张 7×5 的异常矩阵，**每一格都由页内的迷你事务引擎实跑**（MVCC 版本链、ReadView 四步判定、行锁、间隙锁、串行执行都实现了），点任意格子看完整时序。结果是个漂亮的阶梯，但**「丢失更新」和「写偏斜」两行在 RR 那一列是红的**——医院值班表 200 次试验，RC 违规 120 次，RR 也是 120 次，一模一样；加 `FOR UPDATE` 后 0 次。另外 2PL vs SSI 那个实验我没按教科书写「高冲突 2PL 反超」，因为模型跑出来不支持：真实的差别是 **2PL 的白做占比几乎不涨（0%→14%），SSI 从 14% 涨到 68%，而墙上时间只差 1 tick**。

Go 和 Java 两版输出**逐字节相同**（共用同一个 8 行 xorshift32，可以直接 `diff`）。顺手修了全部 8 个课件页在手机宽度下的横向溢出。

---

下一站 **Part 4-B**（2PC/XA、TCC、Saga、本地消息表、事务消息）之前，先给你两道热身：

1. 4-A 里说「能变成约束的不变量，一律变成约束」。**跨服务的分布式事务里，这条还成立吗？**比如订单服务和库存服务各有自己的库,「下单不能超卖」这个不变量能变成谁的约束？
2. 2PC 的参与者在 prepare 之后、收到 commit 之前宕机了，重启后它该怎么办——**它能自己决定吗**？如果不能，它要去问谁，问不到又该怎么办？

> ⚠️ **缺失片段**：序号 3660–3892 的事件不在本地缓存中。
