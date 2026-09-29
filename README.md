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
open courseware/ch03d-faq.html               # Part 3-D：附录（四个 index / Quorum vs Raft / 配置与冷启动）
open courseware/ch04a-transactions.html      # Part 4-A：单机事务（隔离级别 / MVCC / 写偏斜）
# Linux 用 xdg-open

# 2. 跑实验（Go 版，需要 Go 1.21+）
cd go/lab00 && go run .      # Lab 0：不可靠信道与第三态
cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC
cd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT
cd go/lab03a && go run .     # Lab 3-A：确定性网络 + Raft 选举 + 安全性断言
cd go/lab03b && go run .     # Lab 3-B：日志复制 + Figure 8 复现
cd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读
cd go/lab03d && go run .     # Lab 3-D：ReadIndex 第三步 / no-op / Quorum vs Raft / 冷启动
cd go/lab04a && go run .     # Lab 4-A：异常矩阵 / ReadView / 写偏斜 / 2PL vs SSI

# 3. 跑实验（Java 版，需要 JDK 17+，无需 Maven/Gradle）
cd java/lab00 && java Lab00.java
cd java/lab01 && java Lab01.java
cd java/lab02 && java Lab02.java
cd java/lab03a && java Lab03A.java
cd java/lab03b && java Lab03B.java
cd java/lab03c && java Lab03C.java
cd java/lab03d && java Lab03D.java
cd java/lab04a && java Lab04A.java
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
| `-retry` | 客户端最大重试次数（`0` = at-most-once） | `3` |
| `-n` | 转账笔数 | `200` |
| `-seed` | 随机种子，同种子结果完全可复现 | `42` |
| `-crash` | Lab 0-4 中两步之间注入崩溃的概率 | `0.30` |

**Lab 2 参数**

| 参数 | 含义 | 默认 |
|---|---|---|
| `-n` `-w` `-r` | 副本数 N、写 quorum W、读 quorum R | `5` `3` `3` |
| `-ops` | 每组实验的读写次数 | `400` |
| `-lag` | 复制延迟（多少次操作后异步副本才收到） | `3` |

**Lab 1 参数**

| 参数 | 含义 | 默认 |
|---|---|---|
| `-skew` | 副本 R1 的时钟偏移（毫秒） | `80` |
| `-jit` | 时钟偏移的随机抖动（毫秒） | `60` |
| `-gap` | 两次并发写的真实间隔（毫秒） | `60` |
| `-n` | 并发写轮数 | `200` |
| `-seed` | 随机种子 | `42` |

---

## 学习路线

| Part | 主题 | 核心产出 | 状态 |
|---|---|---|---|
| **0** | **序言与心智模型** —— 三个免费假设的崩塌、三个不可靠、故障模型、两将军、幂等性 | 交互课件 + Lab 0 | ✅ 已完成 |
| **1** | **时间、顺序与因果** —— happens-before、Lamport 时钟、向量时钟、LWW 陷阱、HLC、TrueTime | 交互课件 + Lab 1 | ✅ 已完成 |
| **2** | **复制与一致性模型** —— 复制拓扑、三大延迟异常、一致性谱系、CAP/PACELC、Quorum、CRDT | 交互课件 + Lab 2 | ✅ 已完成 |
| **3-A** | **共识与 Raft 选举** —— 共识定义、FLP、2PC 为什么不是共识、任期、选举规则、随机化 | 交互课件 + Lab 3-A | ✅ 已完成 |
| **3-B** | **日志复制与安全性** —— 日志匹配性质、提交规则、Figure 8、五条安全性属性、快速回退 | 交互课件 + Lab 3-B | ✅ 已完成 |
| **3-C** | **持久化 / 快照 / 成员变更 / 线性一致读** —— 三样必落盘的状态、InstallSnapshot、单节点变更的交集证明、ReadIndex 与 Lease Read | 交互课件 + Lab 3-C | ✅ 已完成 |
| **3-D** | **附录：算法与系统的接缝** —— 四个 index 与 ReadIndex 的第三步、no-op 的必要性、Quorum ≠ Raft、配置从哪来与冷启动、Raft log ≠ redo log | 交互课件 + Lab 3-D | ✅ 已完成 |
| **4-A** | **单机事务：隔离级别、MVCC 与写偏斜** —— ACID 里哪个字母是假的、七种并发异常、ReadView 的四步判定、Next-Key Lock、快照隔离的盲区、2PL vs SSI | 交互课件 + Lab 4-A | ✅ 已完成 |
| 4-B | 分布式事务的提交协议 —— 2PC/XA 与阻塞、TCC、Saga、本地消息表、事务消息 | Lab 4-B | ⏳ |
| 4-C | 分布式事务的版本层 —— Percolator、Calvin、Spanner 外部一致性、2PC + Raft | Lab 4-C | ⏳ |
| 5 | 经典系统精读 —— GFS / MapReduce / Bigtable / Dynamo / Spanner / ZooKeeper / Kafka / RocketMQ / Vitess | 架构拆解笔记 | ⏳ |
| 6 | 工程化与验证 —— 分布式锁陷阱、混沌工程、Jepsen、TLA+ | Lab 6：给自己的 Raft 跑线性一致性检查 | ⏳ |

> Part 3 的 Raft 实现是整条路线的分水岭。建议在那里慢下来，真的把代码跑通。

---

## 目录结构

```
distributed-systems-course/
├── README.md
├── CLAUDE.md                # 与 Claude 续学的约定（开场读什么、每轮怎么记录）
├── notes/                   # 可公开的学习记录（已脱敏）
│   ├── progress-review.md   # 学习进度、作答记录、待处理勘误
│   ├── questions.md         # 追问索引（原话 + 结论 + 答案位置）
│   ├── sessions/            # 会话记录（按天，已脱敏）
│   └── tools/               # 导入恢复会话、脱敏检查的脚本
├── courseware/
│   ├── ch00-intro.html      # 第 0 章课件（6 个模拟实验 + 12 道自测题）
│   ├── ch01-time-order.html # Part 1 课件（时空图编辑器 + LWW 复现 + HLC，10 道自测题）
│   ├── ch02-replication.html # Part 2 课件（一致性判定器 + Quorum 配置器 + CRDT）
│   ├── ch03a-consensus-election.html # Part 3-A 课件（2PC 阻塞 + Raft 选举模拟器）
│   ├── ch03b-log-replication.html # Part 3-B 课件（日志复制 + Figure 8 交互重放）
│   ├── ch03c-production-raft.html # Part 3-C 课件（崩溃重启 + 成员变更裂脑 + 三种读）
│   ├── ch03d-faq.html           # Part 3-D 课件（index 流水线 + Quorum/Raft 对照 + 冷启动沙盒）
│   └── ch04a-transactions.html  # Part 4-A 课件（异常矩阵 + ReadView 判定器 + 写偏斜 + 2PL/SSI）
├── go/
│   ├── go.mod
│   ├── internal/tui/        # 中英混排的终端表格对齐工具（各 Lab 共用）
│   ├── lab00/
│   │   ├── net.go           # 不可靠信道：丢包 / 延迟 / 乱序 / 重复 + 第三态
│   │   ├── bank.go          # 服务端：三种扣款实现（天真 / 幂等 / 非原子陷阱）
│   │   ├── client.go        # 客户端重试策略：at-most-once vs at-least-once
│   │   └── main.go          # 四个实验的编排与输出
│   ├── lab01/
│   │   ├── clocks.go        # Lamport 时钟 / 向量时钟 / HLC 三种实现
│   │   ├── trace.go         # 事件轨迹引擎 + happens-before 可达性校验
│   │   └── main.go
│   ├── lab02/
│   │   ├── quorum.go        # N 副本 Quorum KV，W/R 可配，带复制延迟
│   │   ├── crdt.go          # G-Counter / PN-Counter / OR-Set / LWW-Register
│   │   ├── history.go       # 一致性判定器：线性 / 顺序 / 因果 / 会话保证
│   │   └── main.go
│   ├── lab03a/
│   │   ├── net.go           # 确定性网络：虚拟时钟 + 可注入延迟/丢包/分区/宕机
│   │   ├── raft.go          # Raft 选举状态机 + Election Safety 断言
│   │   └── main.go
│   ├── lab03b/
│   │   ├── net.go           # 同上，Msg 扩展了日志字段
│   │   ├── raft.go          # 日志复制 + 提交规则 + 三条安全性断言
│   │   └── main.go
│   ├── lab03c/
│   │   ├── net.go           # 再扩展 InstallSnapshot 消息
│   │   ├── raft.go          # 持久化 / 快照 / 每节点独立 config / 三种读
│   │   └── main.go
│   ├── lab03d/
│   │   ├── model.go         # 四个聚焦模型：读路径 / no-op / quorum 对照 / bootstrap
│   │   ├── rand.go          # 8 行 xorshift32——为了让 Java 版输出逐字节一致
│   │   └── main.go
│   └── lab04a/
│       ├── engine.go        # 迷你事务引擎：MVCC 版本链 + ReadView + 行锁 + 范围锁
│       ├── scen.go          # 七种并发异常的时序定义与判定
│       ├── cc.go            # 2PL 与 SSI 对照模拟（共用同一个并行度预算）
│       └── main.go
└── java/
    ├── lab00/Lab00.java     # 单文件，零依赖
    ├── lab01/Lab01.java
    ├── lab02/Lab02.java
    ├── lab03a/Lab03A.java
    ├── lab03b/Lab03B.java
    ├── lab03c/Lab03C.java
    ├── lab03d/Lab03D.java
    └── lab04a/Lab04A.java
```

---

## Lab 0 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **0-1** 不可靠信道 | 按 1..10 顺序发出的消息，接收端看到的是丢失、乱序和重复 |
| **0-2** 两种重试语义 | 不重试 → **漏扣**（幽灵成功）；重试 → **重复扣款**。两条路都错，方向相反 |
| **0-3** 幂等键 | 同一份网络故障，只改服务端加去重表，重复扣款归零 |
| **0-4** 非原子去重 | 把去重表和业务操作拆成两步 + 注入崩溃 → 复现「永久丢单」和「重复扣款」两种真实事故 |

关键代码在 `go/lab00/net.go` 的 `Call()` 与 `java/lab00/Lab00.java` 的 `call()`——
只有二十来行，但它们把「第三态」这件事讲清楚了。请务必读一遍那四条注释路径。

---

## Lab 1 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **1-1** 两种逻辑时钟 | 跑出课件时空图里的每一个数字：Lamport 值与向量时钟 |
| **1-2** Lamport 的盲区 | 枚举全部事件对，找出「Lamport 给了顺序、但其实并发」的那些。判定结果用 happens-before 定义做独立可达性校验，不一致直接 panic |
| **1-3** LWW 丢写 | LWW **每一轮都丢掉一条写入**；时钟歪不歪只改变「丢哪一条」。版本向量检出全部冲突、丢弃数为 0 |
| **1-4** HLC | 注入 NTP 回拨，物理时钟倒退，而 HLC 保持严格单调且与物理时间偏差有界 |

关键代码在 `clocks.go` 的 `Compare()`（向量时钟的全部精髓，五行）和 `HLC.Local()`。

---

## Lab 2 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **2-1** Quorum | W+R>N 的配置陈旧读**恒为 0**（不是"概率低"，是数学保证）；W+R≤N 立刻出现陈旧读 |
| **2-2** CAP | 同一份负载跑 CP 与 AP，把取舍变成两个具体数字：CP 的写入失败率 vs AP 的待合并冲突数 |
| **2-3** 一致性判定器 | 穷举所有串行化顺序，判定 6 条执行历史属于谱系哪一档。内置自检：谱系必须满足 线性 ⟹ 顺序 ⟹ 因果，违反直接 panic |
| **2-4** CRDT | 1000 次随机分区 + 随机合并顺序，G-Counter 结果 100% 精确；LWW 每次也"收敛"但丢掉约 2/3 的写入 |

关键代码在 `history.go` 的 `CheckLinearizable` / `CheckSequential`——两者的唯一差别就在那一个 `canPlace` 闭包里，而那就是"是否尊重真实时间"的全部含义。

---

## Lab 3-A 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **3A-1** 确定性网络 | 同一个种子必定复现同一次执行。**这是调共识算法的前提**，比 print 有用一百倍 |
| **3A-2** 五个场景 | 冷启动 / 杀 Leader / 3\|2 分区 / 分区恢复 / 反复杀节点。每个 tick 断言 Election Safety，上万个 tick 一次都没破 |
| **3A-3** 随机化超时 | 关掉它：245 次分裂投票、任期飙到几百、**始终选不出 Leader**（活锁）。开启：一轮搞定 |
| **3A-4** 拆掉安全性 | 去掉「每任期一票」，60 个种子里 37 个出现**同任期多 Leader**，最多 3 个 —— 脑裂现场 |

关键代码在 `raft.go` 的 `handle()`（投票规则）和 `checkSafety()`（Election Safety 断言）。
`-unsafe` 开关专门用来把安全性拆给你看。

**场景③有一个值得注意的现象**：如果分区前的 Leader 恰好落在少数派一侧，基础 Raft 不会让它主动退位 —— 它成为「僵尸 Leader」，自认为是主但一条日志都提交不了。工程上用 CheckQuorum 解决，etcd 默认开启。

---

## Lab 3-B 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **3B-1** 日志复制 | 200 条命令，所有节点日志逐条相同，状态机在每个 index 上执行相同命令 |
| **3B-2** 三条安全性断言 | 边写边杀 Leader / 反复分区 / 30% 丢包，每个 tick 检查 Log Matching、Leader Completeness、State Machine Safety，全部成立 |
| **3B-3** Figure 8 复现 | 精确构造论文那五个节点的状态。**`-naivecommit` 一开，断言立刻报出「Leader 完整性被破坏」** |
| **3B-4** 回退策略对照 | 朴素回退 163 次被拒 / 16.3 秒；快速回退 **1 次 / 0.13 秒**，121× 加速 |

关键代码在 `raft.go` 的 `maybeCommit()`——那一行 `if !c.NaiveCommit && n.log[N].Term != n.term { continue }` 就是 Figure 8 的全部修补。

---

## Lab 3-C 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **3C-1** 持久化 | 分别丢掉 `currentTerm` / `votedFor` / `log`。**构造场景下丢失 votedFor 必然产生同任期两个 Leader**——票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次 |
| **3C-2** 快照 | 220 条日志压缩到只剩 6 条；隔离的节点落后到快照点之前，靠 InstallSnapshot 在 200ms 内追平，日志匹配性质不变 |
| **3C-3** 成员变更 | 直接跳 3→5：**30/30 裂脑**；单节点 3→4：**0/30，一次都造不出来** |
| **3C-4** 三种读 | 僵尸 Leader 上：本地读 **20/20 静默返回陈旧值**；ReadIndex 与 Lease Read **20/20 正确报错** |

3C-1 里有个值得注意的结果：**随机故障注入跑几十个种子都撞不出 votedFor 的问题，但构造场景一次就命中。** 有些 bug 需要非常特定的交错才现形——这正是 TLA+ 和 Jepsen 存在的理由（Part 6）。

---

## Lab 3-D 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **3D-1** ReadIndex 第三步 | ReadIndex 的第 1、2 步**全部通过**，只跳过第 3 步：apply 滞后 4 条时 **387/500 次读到旧值**，滞后 8 条时 **86%**。执行第 3 步则恒为 0，代价是平均多等约 2 条 |
| **3D-2** no-op 的必要性 | 新 Leader 不发 no-op：**16/40 个种子读到旧值，而客户端收到过「写入成功」**。发 no-op：违规 0 次，其中 15 次读到得早、**排队**等 no-op 提交后才返回（etcd 的 `pendingReadIndexMessages`） |
| **3D-3** Quorum vs Raft | W 从 1 调到 5，**写成功数从 2.00 掉到 0.70，副本分歧数一动不动（1.88 种）**。Raft 侧恒为 1.00 种——不是概率，是结构 |
| **3D-4** 冷启动 | 按 etcd 源码的真实判断逐步执行 7 个场景。列表写错 → **2 个集群 2 个 Leader，都能写**；清空两台后用 `state=new` 重启：n1 在线 → **etcd 拒绝启动**（`already bootstrapped`），n1 不在线 → 放行，n1 回来后 **5 条已确认的写被覆盖**；`existing` 节点列表写错 → 拒绝启动，集群容错归零、再加节点被拒 |

3D-1 的那张表是这一章的眼睛：**「我确认自己还是 Leader」和「我确认提交点是 7」都成立，读出来仍然可以是旧数据。** 少的那一步是「等状态机追上来」。

这个 Lab 不复用 `lab03c/raft.go`——3-A 到 3-C 研究的是**算法本身**会不会出错，需要完整状态机；3-D 研究的是**算法与系统之间的接缝**，用聚焦的小模型讲得更清楚。

```bash
go run . -applylag 8      # apply 线程滞后多少条
go run . -reads 500       # 3D-1 的读次数
go run . -seeds 40        # 3D-2 / 3D-3 的种子数
go run . -writers 4 -w 5  # 3D-3 的并发写客户端数与写 quorum
```

> Go 版和 Java 版的输出**逐字节相同**（连随机数都用同一个 8 行 xorshift32），可以 `diff` 对照。

### 勘误（2026-09 追问后修正）

课件 3-D 新增了「§3.31 追问篇」，其中的勘误表列出了这些修正：

- **冷启动沙盒 D**：原版把「清空后用 `state=new` 重启」建模成生成新的集群 ID。实际上成员 ID = hash(peer URL + token)、集群 ID = hash(成员 ID)，**同 token 同列表就是同一个集群**。现拆成 D1（n1 在线，etcd 拒绝启动）和 D2（n1 不在线，之后覆盖已确认的写）。
- **冷启动沙盒 E**：n1 的 `--initial-cluster` 应只有自己；模型改为 `existing` 节点只加入、不组建集群。
- **配置生效时机**：论文是「追加即生效」，etcd 是「旧配置下提交、apply 时生效，一次只允许一个未提交的变更」。
- **§3.28 例外 2**：会捣乱的是被移除的节点，不是新节点。
- **no-op 提交前的读**：etcd 是排队，不是拒绝。
- **3-C lastApplied**：持久化状态机必须把 applied index 与数据同事务落盘（etcd 的 `consistent_index`）。
- **Lab 3-B / 3-C**：follower 的提交上限改为 `min(leaderCommit, 本次 RPC 最后一条新条目)`（Figure 2 原文）。在这两个 Lab 里修复前后输出逐字节不变——原写法是潜伏的，一旦限制单条消息大小才会触发。
- **课件 4-A §4.5**：补充「InnoDB 锁的是扫过的行，不是匹配的行」，无合适索引时锁全表。

---

## Lab 4-A 会让你亲手复现什么

| 实验 | 现象 |
|---|---|
| **4A-1** 异常 × 隔离级别矩阵 | 七种并发异常 × 五种隔离配置，**每一格都真的跑一遍**，不是抄表。结果是一个漂亮的阶梯——但**「丢失更新」和「写偏斜」两行在 RR 那一列仍然是红的** |
| **4A-2** ReadView 可见性 | 同一条版本链、同一套四步判定，只把活跃事务集合的采集时刻挪一下，返回值就从 **300 跳到 400**。*这就是 RC 和 RR 的全部实现差异* |
| **4A-3** 写偏斜 | 医院值班表跑 200 次：RC **120/200** 破坏不变量，RR **120/200**（一模一样），加 `FOR UPDATE` 后 **0/200**，串行化 **0/200** |
| **4A-4** 2PL vs SSI | 热点集中度从 0 拉到 100%：**2PL 的「白做占比」几乎不涨（0% → 14%），SSI 从 14% 一路涨到 68%**，而两者墙上时间相差不大（43 vs 44 ticks）——*SSI 省下来的等待，是用真金白银的 CPU 换的* |

4A-1 的矩阵是这一章的眼睛。它把一句话变成了可验证的事实：**把 MySQL 从 RC 调到 RR，你买到的是「不可重复读」和「幻读」，没有买到「丢失更新」和「写偏斜」**——而后两者才是业务不变量被破坏的那一类。

这个 Lab 里的迷你事务引擎（`engine.go`，不到 300 行）实现了 MVCC 版本链、ReadView 四步判定、行锁、间隙锁和串行执行。它存在的理由只有一个：让「哪个隔离级别挡得住哪种异常」从**背下来的表格**变成**跑出来的结果**。

```bash
go run . -trials 500          # 4A-3 的并发试验次数
go run . -doctors 4 -leavers 4 # 4A-3 在岗医生数与同时请假人数
go run . -txns 16 -ops 6      # 4A-4 并发事务数与每事务操作数
```

> Go 版和 Java 版的输出**逐字节相同**，可以 `diff` 对照。

---

## 阅读顺序建议

1. 先打开 `courseware/ch00-intro.html`，**动手玩完 6 个实验**（尤其是实验 2 和实验 4）
2. 读完 §0.1 到 §0.8，遇到不认识的词回 §0.9 术语速查表
3. 做 §0.11 的 12 道自测题，**先自己想 30 秒再展开答案**
4. 跑 Lab 0（Go 或 Java 任选，两个都跑更好），改参数再跑几遍
5. 回答 Lab 结尾的三个思考题

---

## 参考资料

- Lamport, *Time, Clocks, and the Ordering of Events in a Distributed System* (1978)
- Fischer, Lynch, Paterson, *Impossibility of Distributed Consensus with One Faulty Process* (1985)
- Ongaro & Ousterhout, *In Search of an Understandable Consensus Algorithm (Raft)* (2014)
- Kleppmann, *Designing Data-Intensive Applications*（中译《数据密集型应用系统设计》）
- MIT 6.5840 (原 6.824) Distributed Systems 课程与 Lab
