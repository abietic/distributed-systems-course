## 👤 你 · 2026-08-31 13:45

1. 用的是updated_at，这个时间戳是数据库主库打的
2. 由于lamport时钟只能用来证伪即当a happens before b的时候一定不存在L(a)>=L(b)，因此想要一定有a happens before b需要进一步增加配合进程ID，形成（PID， Lamport）对通过排序确定，即使用lamport的另一个特性在单个进程下全序的能力来实现。
3. 有，主要是〔内部库C〕和〔内部服务A〕，其中后者负责统一对用户录入、提取的指标进行指标的标准化，〔内部表Y〕到〔内部表X〕；这个表只由〔内部服务A〕来写从而达到〔内部服务A〕统一盖戳的目的。同时现在这些服务中大部分冲突通过on conflict语义使用数据库来解决，这应该算是一种“数据库唯一约束”+“LWW”？

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00143〕\" && echo \"=== 〔内部表X〕 DDL ===\" && grep -rn --include=*.sql --include=*.py --include=*.yaml --include=*.yml -il \"〔内部表X〕\" 〔内部库C〕 〔内部服务A〕 2>/dev/null | head -20"}
````

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00143〕\" && echo \"=== ON CONFLICT 用法统计 ===\" && grep -rn -io \"on conflict[^;]\\{0,90\\}\" --include=*.py --include=*.sql 〔内部服务A〕 〔内部库C〕 2>/dev/null | head -40"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果（错误）</summary>

````
The device this session is bound to is not connected to the bridge.
The connection to the user's computer is often restored within a few seconds. If you have not already retried this tool call, retry it once now. If a retry also can't reach it, stop retrying and tell the user you can't reach their computer right now.
````

</details>

<details><summary>↳ mcp__remote-devices__device_bash 结果（错误）</summary>

````
The device this session is bound to is not connected to the bridge.
The connection to the user's computer is often restored within a few seconds. If you have not already retried this tool call, retry it once now. If a retry also can't reach it, stop retrying and tell the user you can't reach their computer right now.
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00143〕\" && grep -rn -il \"〔内部表X〕\" 〔内部库C〕 〔内部服务A〕 2>/dev/null | head -20"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果（错误）</summary>

````
The device this session is bound to is not connected to the bridge.
The connection to the user's computer is often restored within a few seconds. If you have not already retried this tool call, retry it once now. If a retry also can't reach it, stop retrying and tell the user you can't reach their computer right now.
````

</details>

## 🤖 Claude · 2026-08-31 13:48

先说一句：我想去看你 `〔内部服务A〕` 里 `ON CONFLICT` 的实际写法，但**你的电脑现在连不上**（桌面 App 可能关了）。等它回来我再看。下面基于你的描述来。

## 第 1 题 — 做法是对的，但有三个坑要查

"时间戳由数据库主库打"正是 §1.5 检查表的第三条，你已经做对了：单一时钟源，跨机器偏移问题直接消失。三个边界：

**① PostgreSQL 的 `now()` 是事务开始时刻，不是语句执行时刻。** 一个跑了 3 秒的事务里，所有行的 `updated_at` 完全相同。要语句时刻用 `statement_timestamp()`，要真实墙钟用 `clock_timestamp()`。

**② 主库 failover 之后，新主库的时钟可能比旧主库慢。** "单一时钟源"这个前提在切换那一刻断掉了，`updated_at` 会出现跨主库的倒退。

**③ 这个最要命 —— `updated_at` 当增量同步游标的静默丢数据。** 和时钟准不准无关，纯粹是"打戳时刻 ≠ 可见时刻"：

```
10:00:00.000  T1 开始事务，now() 定格在 .000（这行要写 3 秒）
10:00:00.500  T2 开始事务，now() = .500
10:00:00.510  T2 提交            ← 消费者能看见了
10:00:00.600  消费者拉取 updated_at > 上次游标，读到 T2，游标推进到 .510
10:00:03.000  T1 提交，但它的 updated_at 还是 .000
              → .000 < .510，下一次拉取永远拿不到它
```

**T1 的数据永久丢失，没有任何报错。** 如果 `〔内部表Y〕 → 〔内部表X〕` 那条链路是靠 `updated_at` 做增量的，值得去查一下。三种解法：改用单调递增的 sequence 列或逻辑复制 LSN 做游标；留一个安全回溯窗口（游标往回退 N 秒）配合幂等消费；或者干脆走 CDC，按提交顺序读而不是按时间戳读。

## 第 2 题 — 你答的是另一个问题，而且答对了

你说的 `(L, pid)` 全序完全正确，那是 §1.3 讲的东西。但它回答的是"怎么让所有节点得到一致的排序"，而我问的是"怎么让 Lamport 不再误导"。这两件事方向相反：

> **`(L, pid)` 没有消除误导，它把误导标准化了。** 对并发事件，它任意指定一个顺序，然后保证所有节点都以*同样的方式*误导。

而这在很多场景恰恰是你想要的——分布式互斥、状态机复制的输入定序，你不需要知道真相，你只需要所有人同意。**一致比正确更容易达到，也常常已经够用。**

我那个问题的答案是：**让并发对本身消失**。Lamport 的盲区不是精度问题，是<u>信息缺失</u>——一个整数编码不了 N 维的偏序。所以只有两条出路：

- **补足信息** → 向量时钟，代价 O(N) 空间
- **消灭偏序** → 让所有操作串行经过一个共同的点（单 leader / 全序广播），并发根本不产生

第二条就是共识算法。代价是吞吐和延迟——你在 Lab 1-2 里把示例改成"每个事件前后都全局同步一轮"，误导率会降到 0，那个系统的吞吐也没了。**这是一致性和性能的第一次正面交锋**，正好是 Part 2 的主题。

## 第 3 题 — "唯一约束 + LWW"，一半对，得拆开看

`ON CONFLICT` 有两种完全不同的语义，混在一起说会掩盖真正的风险：

| 写法 | 真实语义 | 会丢更新吗 |
|---|---|---|
| `ON CONFLICT DO NOTHING` | **first-write-wins**，不是 LWW | 不会。这是幂等插入，重试安全 |
| `DO UPDATE SET c = EXCLUDED.c` | **last-arriving-wins** | **会丢**。整行覆盖，先到那次改的字段被抹掉 |
| `DO UPDATE SET c = t.c + EXCLUDED.c` | 基于当前值合并 | 不会丢 —— **这其实就是一个 CRDT**（G-Counter 的数据库版） |
| `DO UPDATE ... WHERE t.version < EXCLUDED.version` | 条件写 / 乐观锁 | 不会丢，冲突方会失败并重试 |

关键在第二行的"last"是什么：**它是"到达数据库的顺序"，不是时间戳大小。** 所以它比时间戳 LWW **安全得多**——顺序来自单一的、真实的串行化点（数据库的行锁和 MVCC），完全不受任何机器时钟影响。你等于把检查表第三条做到了极致：干脆不用时间戳。

但它仍然会丢并发更新（经典的 lost update），只是丢法不同：不是"时钟歪了丢错人"，而是"后到的整行覆盖先到的"。

**接下来是我最想问的一个问题**：`〔内部服务A〕` 是单实例，还是 K8s 上跑多副本？

"只由 〔内部服务A〕 写"是很好的设计——**单写者是最省事的冲突消除手段**，比引入向量时钟划算一个数量级。但如果它跑了多个 pod，"只由 〔内部服务A〕 写"就不等于"单写者"了，两个 pod 照样并发写同一行。真正的单写者需要三者之一：单实例、按 key 分片路由（同一个 indicator 永远落到同一个 pod）、或者依赖数据库层的串行化。

---

现在去做 Part 2。我会把你这个 `ON CONFLICT` 场景直接放进引子，一致性谱系挂在它上面讲。

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **TaskCreate** 编写 Part 2 交互式课件：复制与一致性模型

<details><summary>↳ TaskCreate 结果</summary>

````
Task #9 created successfully: 编写 Part 2 交互式课件：复制与一致性模型
````

</details>

▸ **TaskCreate** 编写 Lab 2：Quorum KV 与 CRDT（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #10 created successfully: 编写 Lab 2：Quorum KV 与 CRDT（Go + Java）
````

</details>

▸ **TaskCreate** 验证并交付 Part 2

<details><summary>↳ TaskCreate 结果</summary>

````
Task #11 created successfully: 验证并交付 Part 2
````

</details>

▸ **Bash** Write Part 2 head and sections 2.0-2.1

````
cd 〔路径00054〕 && mkdir -p ch02 && cd ch02 && cat > ch02.html <<'HEAD_EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 2 · 复制与一致性模型</title>
HEAD_EOF
cat 〔路径00240〕 >> ch02.html
cat >> ch02.html <<'HEAD2_EOF'
<style>
/* Part 2 专用 */
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.verd{display:grid;grid-template-columns:repeat(auto-fit,minmax(146px,1fr));gap:1px;background:var(--border);
  border:1px solid var(--border);border-radius:10px;overflow:hidden;margin-top:16px}
.vc{background:var(--surface);padding:12px 14px}
.vc .k{font-size:11px;color:var(--ink-3);le
…（截断，原长 7583 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 2 sections 2.2-2.3

````
cat >> ch02.html <<'EOF'
<section id="s2">
  <h2>2.2 复制延迟的三个经典异常</h2>
  <p>单主复制里没有写冲突，但有另一类问题：<strong>从库总是落后主库一点</strong>。这个"一点"在正常时候是几毫秒，在主库压力大或网络抖动时可以变成几十秒。三种异常都由它引起，而且各有一个对应的修复手段。</p>

  <div class="ladder">
    <div class="lrung"><div><div class="nm">① 读不到自己的写</div><div class="en">read-your-writes</div></div>
      <div class="de">用户改完昵称立刻刷新，看到的还是旧昵称——因为写走了主库，读走了从库，而复制还没到。<br>
        <strong>修复</strong>：用户自己修改过的数据，一段时间内的读走主库；或按会话粘性把该用户的读固定到某个足够新的从库；或客户端记住自己最后一次写的位点，只读位点不低于它的副本。</div></div>
    <div class="lrung"><div><div class="nm">② 时光倒流</div><div class="en">monotonic reads</div></div>
      <div class="de">用户刷新两次，第一次看到了新评论，第二次评论消失了——因为两次读随机落到了两个复制进度不同的从库。<br>
        <strong>修复</strong>：让同一个用户的读始终落到<em>同一个</em>副本（按 user_id 哈希路由，而不是随机负载均衡）。</div></div>
    <div class="lrung"><div><d
…（截断，原长 6066 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 2 sections 2.4-2.7

````
cat >> ch02.html <<'EOF'
<section id="s4">
  <h2>2.4 CAP 的正确读法</h2>
  <p>CAP 大概是分布式领域被误读最多的一个定理。先把常见的错误说法列出来：</p>
  <div class="note crit">
    <div class="note-h"><span class="dot" style="background:var(--crit)"></span>三个流行的错误</div>
    <p style="margin-bottom:0"><strong>✗ "三选二"</strong> —— <strong>P 不是可选项。</strong>只要你的系统跨网络，分区就一定会发生（网线、交换机、机房专线、云厂商的网络抖动）。你不能"选择不要分区容错"，你只能选择<em>分区发生时</em>怎么办。所以真正的选择只有两个：CP 还是 AP。<br>
    <strong>✗ "MongoDB 是 CP，Cassandra 是 AP"</strong> —— 这是系统<em>配置</em>的属性，不是系统的属性。Cassandra 把 W 和 R 都设成 QUORUM 就偏 CP，MongoDB 用 <code>w:1</code> 就偏 AP。<br>
    <strong>✗ "选了 AP 就没有一致性"</strong> —— C 在 CAP 里特指<strong>线性一致性</strong>。放弃它不等于什么都不保证，因果一致和各种会话保证仍然拿得到，而且它们在分区期间依然可用。</p>
  </div>
  <p>精确的表述是：<em>当网络分区发生时，你必须在"继续响应但可能返回陈旧/冲突数据"和"拒绝服务直到分区恢复"之间二选一。</em></p>
  <table>

…（截断，原长 10074 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 2 quiz and lab sections

````
cat >> ch02.html <<'EOF'
<section id="s8">
  <h2>2.8 自测题</h2>
  <p>先自己想 30 秒再展开。</p>

  <details class="q"><summary>1. 线性一致和顺序一致，唯一的区别是什么？</summary>
    <div class="a"><p><strong>是否尊重真实时间。</strong></p>
    <p>两者都要求存在一个全序，且每个客户端自己的操作在全序里保持程序序。线性一致<em>额外</em>要求：如果操作 A 在操作 B 开始之前就已经返回，那么 A 在全序里必须排在 B 前面。</p>
    <p style="margin-bottom:0">实际后果：你写完并收到成功响应，3 秒后我去读，顺序一致<strong>允许</strong>我读到旧值（只要能找到一个自洽的全序把我的读排在你的写之前）；线性一致<strong>禁止</strong>。这就是为什么线性一致必须走主库或走共识——它需要一个真实时间上的同步点。</p></div></details>

  <details class="q"><summary>2. 为什么说 CAP 的"三选二"是错的？</summary>
    <div class="a"><p>因为 <strong>P 不是可选项</strong>。只要系统跨网络，分区就一定会发生——你不能选择"不要分区容错"，你只能选择<em>分区发生时</em>的行为。</p>
    <p style="margin-bottom:0">所以真正的选择只有二选一：分区期间是继续服务（AP，可能返回陈旧或冲突数据），还是拒绝服务（CP，少数派停摆）。而且这是<strong>系统配置</strong>的属性而非系统本身的属性：Cass
…（截断，原长 7009 字符）
````

<details><summary>↳ Bash 结果</summary>

````
sections: 10 kb: 55
````

</details>

▸ **Bash** Write Part 2 JS: replication lag and consistency checker

````
cat >> ch02.html <<'EOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(t,a,x){const e=document.createElementNS(NS,t);for(const k in a)e.setAttribute(k,a[k]);if(x!=null)e.textContent=x;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}

$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");draw1();draw2();draw3();draw4();};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHeight)*100)+"%";

…（截断，原长 13981 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 2 JS: quorum and CRDT experiments

````
cat >> ch02.html <<'EOF'

/* ══════════════════════════════════════════════════════════════════
   实验 3 · Quorum 配置器
   ══════════════════════════════════════════════════════════════════ */
function draw3(){
  let N=+$("#r3n").value;
  const rw=$("#r3w"), rr=$("#r3r");
  rw.max=N; rr.max=N;
  let W=Math.min(+rw.value,N), R=Math.min(+rr.value,N);
  rw.value=W; rr.value=R;
  $("#l3n").textContent=N; $("#l3w").textContent=W; $("#l3r").textContent=R;

  const svg=$("#sv3"); svg.innerHTML="";
  const CXp=250, CY=100, RAD=72, CX2=520;
  // 写集合取前 W 个，读集合取后 R 个 —— 最坏情况的重叠
  const wSet=new Set(), rSet=new Set();
  for(let i=0;i<W;i++) wSet.add(i);
  for(let i=0;i<R;i++) rSet.add((N-1-i+N)%N);
  const inter=[...wSet].filter(i=>rSet.has(i));
  const guaranteed = W+R>N;

  const ring=(cx,title,set,col
…（截断，原长 9898 字符）
````

<details><summary>↳ Bash 结果</summary>

````
script: 1 1 kb: 84
````

</details>

▸ **Bash** Verify Part 2 interactives

````
cd /tmp && cat > c2.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:950}});
const errs=[];p.on('pageerror',e=>errs.push('PAGEERROR: '+e.message));
p.on('console',m=>{if(m.type()==='error')errs.push('CONSOLE: '+m.text());});
await p.goto('file://〔路径00063〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(500);

// 一致性判定器：遍历全部 6 个场景
const n = await p.evaluate(()=>document.querySelectorAll('#sc2 button').length);
const res=[];
for(let i=0;i<n;i++){
  await p.click(`#sc2 button[data-i="${i}"]`);
  await p.waitForTimeout(150);
  const r = await p.evaluate(()=>({
    name
…（截断，原长 2686 字符）
````

<details><summary>↳ Bash 结果</summary>

````
① 教科书式的线性一致              线性一致:✓ 顺序一致:✓ 因果一致:✓ 读己之写:✓ 单调读:✓
② 重叠区间里读到旧值              线性一致:✓ 顺序一致:✓ 因果一致:✓ 读己之写:✓ 单调读:✓
③ 陈旧读                    线性一致:✗ 顺序一致:✓ 因果一致:✓ 读己之写:✓ 单调读:✓
④ 两个观察者，相反的顺序            线性一致:✗ 顺序一致:✗ 因果一致:✓ 读己之写:✓ 单调读:✗
⑤ 因果倒置                   线性一致:✗ 顺序一致:✗ 因果一致:✗ 读己之写:✓ 单调读:✗
⑥ 读不到自己的写                线性一致:✗ 顺序一致:✗ 因果一致:✗ 读己之写:✗ 单调读:✓
lag scen 0 => 违反了 read-your-writes（读不到自己的写）
lag scen 1 => 违反了 monotonic reads（时光倒流）
lag scen 2 => 违反了 consistent prefix（因果错乱）
quorum majority: W + R=6 | 读到最新=保证 | 写可容忍故障=2 | 读可容忍故障=2 | 读写都可用=2
quorum W+R<=N: 交集可能为空
CRDT: 真实点击总数=6 | G-Counter 读数=6 / 6 / 6 | LWW 读数=1 / 1 / 1 | LWW 丢失=5
after 2nd merge: 6 | 6 / 6 / 6 | 1 / 1 / 1 | 5
ERRORS: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Add causal/monotonic note and screenshot

````
cd 〔路径00062〕 && python3 - <<'PY'
p='ch02.html'; s=open(p,encoding='utf-8').read()
old='''  if(ryw.length) h+='<p style="margin:10px 0 0;color:var(--crit);font-size:13.5px"><strong>读己之写被破坏：</strong>'+ryw.join("；")+'</p>';
  if(mono.length) h+='<p style="margin:6px 0 0;color:var(--crit);font-size:13.5px"><strong>单调读被破坏：</strong>'+mono.join("；")+'</p>';'''
new='''  if(ryw.length) h+='<p style="margin:10px 0 0;color:var(--crit);font-size:13.5px"><strong>读己之写被破坏：</strong>'+ryw.join("；")+'</p>';
  if(mono.length) h+='<p style="margin:6px 0 0;color:var(--crit);font-size:13.5px"><strong>单调读被破坏：</strong>'+mono.join("；")+'</p>';
  if(cau&&mono.length) h+='<div class="note warn" style="margin:12px 0 0"><div class="note-h"><span class="dot" style="background:var(--warn)"></span>'+
    '为什么
…（截断，原长 2487 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
errors: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00217〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Write Go Quorum KV and CRDT modules

````
mkdir -p 〔路径00095〕 && cat > 〔路径00099〕 <<'EOF'
package main

import "math/rand"

// ════════════════════════════════════════════════════════════════════════
// Quorum KV：N 个副本的键值存储，W/R 可配，带复制延迟
//
//	核心命题：W + R > N  ⟺  任意读集合与任意写集合必有交集  ⟹  读一定能看到最新写
//	Lab 2-1 会用实测数据验证它。
//
// ════════════════════════════════════════════════════════════════════════

type Versioned struct {
	Ver int
	Val string
}

type pending struct {
	node     int
	v        Versioned
	deliverAt int // 第几次操作之后送达（模拟复制延迟）
}

type QuorumKV struct {
	N, W, R int
	Lag     int // 未被写 quorum 覆盖的副本，要延迟这么多次操作才收到
	rnd     *rand.Rand

	replicas []Versioned
	queue    []pending
	clock    int // 全局版本号（上帝视角）
	opCount  int

	StaleReads, TotalReads, FailedWrites int
}

func NewQuorumKV(n, w, r, lag i
…（截断，原长 4816 字符）
````

<details><summary>↳ Bash 结果</summary>

````
quorum+crdt ok
````

</details>

▸ **Bash** Write Go consistency checker

````
cat > 〔路径00097〕 <<'EOF'
package main

import "sort"

// ════════════════════════════════════════════════════════════════════════
// 一致性判定器：和课件「实验 2」跑的是同一套算法。
// 这是 Jepsen 做的事情的极简版 —— 穷举所有可能的串行化顺序，
// 看是否存在一个满足对应约束的顺序。
// ════════════════════════════════════════════════════════════════════════

const Init = "⊥"

type Op struct {
	C    int    // 客户端编号
	Kind string // "w" 写 / "r" 读
	V    string // 值
	S, E int    // 真实时间上的 [发出, 返回] 区间
}

type History struct {
	Name string
	Desc string
	Ops  []Op
}

// dfsOrder 带剪枝的穷举：在 idx 这些操作里找一个合法串行化顺序。
//
//	canPlace 决定"此刻能不能把 j 放在下一个位置"——不同一致性模型的差别全在这里。
//	读的合法性检查是共同的：读必须返回当前寄存器的值。
func dfsOrder(ops []Op, idx []int, canPlace func(j int, placed map[int]bool) bool) []int {
	placed := map[int]bool{}
	var out []int
	steps := 0

	var
…（截断，原长 6094 字符）
````

<details><summary>↳ Bash 结果</summary>

````
history ok
````

</details>

▸ **Bash** Write and run Go Lab 2

````
cd 〔路径00095〕 && cat > main.go <<'EOF'
// Lab 2 · 复制与一致性模型
//
// 运行：  cd go/lab02 && go run .
// 调参：  go run . -n 5 -w 3 -r 3 -ops 500 -lag 200
//
// 配套课件：courseware/ch02-replication.html
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"dsc/internal/tui"
)

var (
	nF    = flag.Int("n", 5, "副本数 N")
	wF    = flag.Int("w", 3, "写 quorum W")
	rF    = flag.Int("r", 3, "读 quorum R")
	opsF  = flag.Int("ops", 400, "每组实验的读写次数")
	lagF  = flag.Int("lag", 3, "复制延迟（多少次操作之后异步副本才收到）")
	seedF = flag.Int64("seed", 42, "随机种子")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：N=%d ｜ W=%d ｜ R=%d ｜ 操作数 %d ｜ 复制延迟 %d 次操作 ｜ 种子 %d\n",
		*nF, *wF, *rF, *opsF, *lagF, *seedF)
	lab21()
	lab22()
	lab23()
	lab24()
	epilogue()
}

// ═══════════════════════════════════════════════════════════
…（截断，原长 9407 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 2
stat lab02: no such file or directory
````

</details>

▸ **Bash** Format, vet and run Go Lab 2

````
gofmt -w lab02 && go vet ./lab02/ && cd lab02 && go run . 2>&1 | head -34
````

<details><summary>↳ Bash 结果</summary>

````
配置：N=5 ｜ W=3 ｜ R=3 ｜ 操作数 400 ｜ 复制延迟 3 次操作 ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 2-1 · Quorum：W + R > N
  同一份工作负载，换不同的 W/R，数陈旧读
══════════════════════════════════════════════════════════════════════════════
    配置                        N      W      R        W+R>N      陈旧读比例
    ────────────────────────────────────────────────────────────────────────
    多数派 W=R=⌊N/2⌋+1          5      3      3           是            0.0%
    W=N, R=1                    5      5      1           是            0.0%
    W=1, R=N                    5      1      5           是            0.0%
    W=R=⌊N/2⌋（故意不足）       5      2      2           否           29.5%
    W=1, R=1（最快最弱）        5      1      1           否           81.2%

    W+R>N 的三行陈旧读             0.0%　← 应当恒为 0
    W+R≤N 的陈旧读                 29.5%　← 交集可能为空，读不到最新

    ▸ 前三行的 W+R 都 > N，陈旧读恒为 0 —— 交集性质是一条数学保证，不是概率。
    ▸ 后两行故意让 W+R ≤ N，陈旧读立刻出现。它未必是错误配置：
      Cassandra 的 ONE/ONE 就是这么设的，用一致性换最低延迟和最高可用性。
    ▸ 试试 -lag 20，看后两行的比例怎么变、前三行会不会变。

══════════════════════════════════════════════════════════════════════════════
  Lab 2-2 · CAP：把取舍变成两个数字
  5 个副本被切成 3|2 两侧，同一份写入负载跑两种策略
══════════════════════════════════════════════════════════════════════════════
    策略                  写入成功      写入失败          成功率  恢复后的冲突 key
    ──────────────────────────────────────────────────────────────────────────────
    CP（少数派拒绝）           246           154           61.5%     0（不可能有）
    AP（两侧都收）
…（截断，原长 1640 字符）
````

</details>

▸ **Bash** View Lab 2-3 and 2-4 output

````
go run . 2>&1 | sed -n '36,90p'
````

<details><summary>↳ Bash 结果</summary>

````
      这不是 bug，是设计：宁可停下，不可出错。
    ▸ AP 那一行的冲突数就是你欠下的债，分区恢复后必须用 LWW / siblings / CRDT 还上。
      Lab 2-4 会告诉你这三种还法的差别有多大。

══════════════════════════════════════════════════════════════════════════════
  Lab 2-3 · 一致性判定器
  穷举所有串行化顺序，判定每条历史属于谱系的哪一档
══════════════════════════════════════════════════════════════════════════════
    执行历史                    线性一致  顺序一致  因果一致  读己之写    单调读
    ────────────────────────────────────────────────────────────────────────────
    ① 教科书式的线性一致               ✓         ✓         ✓         ✓         ✓
    ② 重叠区间里读到旧值               ✓         ✓         ✓         ✓         ✓
    ③ 陈旧读                           ✗         ✓         ✓         ✓         ✓
    ④ 两个观察者，相反的顺序           ✗         ✗         ✓         ✓         ✗
    ⑤ 因果倒置                         ✗         ✗         ✗         ✓         ✗
    ⑥ 读不到自己的写                   ✗         ✗         ✗         ✗         ✓

    细看 ④ 两个观察者，相反的顺序  两个并发写，C3 看到 A→B，C4 看到 B→A
      C1 眼中的顺序                w(A)@C1 → w(B)@C2
      C2 眼中的顺序                w(A)@C1 → w(B)@C2
      C3 眼中的顺序                w(A)@C1 → r(A)@C3 → w(B)@C2 → r(B)@C3
      C4 眼中的顺序                w(B)@C2 → r(B)@C4 → w(A)@C1 → r(A)@C4

    ▸ 判定器内置了自检：谱系必须满足 线性 ⟹ 顺序 ⟹ 因果，违反直接 panic。
    ▸ ④ 那一行是最值得盯的：不存在任何单一全序能同时解释 C3 和 C4，
      但允许每人有自己的顺序之后就都说得通了 —— 因为那两个写是并发的。
      这正是因果一致比顺序一致可用性高的全部原因。
    ▸ ⑥ 违反读己之写，同时也违反了因果一致 —— 因果一致蕴含读己之写。

══════════════════════════════════════════════════════════════════════════════
  Lab 2-4 · CRDT：收敛不等于正确
  随机分区、
…（截断，原长 2459 字符）
````

</details>

▸ **Bash** Widen CRDT table columns

````
sed -i 's|	w := \[\]int{26, 14, 16, 16}\n	al := "LRRR"|X|' main.go
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
s=s.replace('	w := []int{26, 14, 16, 16}\n	al := "LRRR"\n	tui.TableHeadA([]string{"数据类型", "1000 次实验", "结果精确", "累计丢失"}, w, al)',
            '	w := []int{28, 14, 20, 24}\n	al := "LRRR"\n	tui.TableHeadA([]string{"数据类型", "实验次数", "结果精确", "累计丢失"}, w, al)')
s=s.replace('fmt.Sprintf("%d 次", 0), fmt.Sprintf("%d / %d 次点击", lwwLost, totalClicks)}, w, al)',
            '"0 次", fmt.Sprintf("%d / %d 次点击", lwwLost, totalClicks)}, w, al)')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
gofmt -w . && go vet . && go run . 2>&1 | sed -n '70,80p'
````

<details><summary>↳ Bash 结果</summary>

````
ok
    ──────────────────────────────────────────────────────────────────────────────────────
    G-Counter（逐位取最大）               1000     1000 次（100%）                       0
    LWW-Register（取时间戳大）            1000                0 次      6013 / 9048 次点击

    G-Counter 与真值不符           0 次　← 应当恒为 0
    LWW 丢失率                     66.5%

    交换律 + 结合律                ✓ 两种合并顺序结果相同
    幂等律                         ✓ 重复合并结果不变
    OR-Set 并发加删                两副本收敛=true，元素存在=true（add-wins）
````

</details>

▸ **Bash** Write and run Java Lab 2

````
mkdir -p 〔路径00130〕 && cat > 〔路径00131〕 <<'EOF'
/*
 * Lab 2 · 复制与一致性模型（Java 版）
 *
 * 运行：  cd java/lab02 && java Lab02.java
 * 调参：  java Lab02.java -n 5 -w 3 -r 3 -ops 500 -lag 20
 *
 * 需要 JDK 17+（单文件源码启动模式）。
 * 配套课件：courseware/ch02-replication.html
 */
import java.util.*;

public class Lab02 {

    static int N = 5, W = 3, R = 3, OPS = 400, LAG = 3;
    static long SEED = 42;
    static final String INIT = "⊥";

    // ══════════════════════════════════════════════════════════════════════
    // Quorum KV：N 个副本，W/R 可配，带复制延迟
    //   核心命题：W + R > N ⟺ 读写集合必有交集 ⟹ 读一定能看到最新写
    // ══════════════════════════════════════════════════════════════════════
    record Versioned(int ver, String val) {}

    static class QuorumKV {
        final int n, w, r,
…（截断，原长 23296 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：N=5 ｜ W=3 ｜ R=3 ｜ 操作数 400 ｜ 复制延迟 3 次操作 ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 2-1 · Quorum：W + R > N
  同一份工作负载，换不同的 W/R，数陈旧读
══════════════════════════════════════════════════════════════════════════════
    配置                        N      W      R        W+R>N      陈旧读比例
    ────────────────────────────────────────────────────────────────────────
    多数派 W=R=⌊N/2⌋+1          5      3      3           是            0.0%
    W=N, R=1                    5      5      1           是            0.0%
    W=1, R=N                    5      1      5           是            0.0%
    W=R=⌊N/2⌋（故意不足）       5      2      2           否           34.0%
    W=1, R=1（最快最弱）        5      1      1           否           80.8%

    W+R>N 的三行陈旧读             0.0%　← 应当恒为 0

══════════════════════════════════════════════════════════════════════════════
  Lab 2-3 · 一致性判定器
  穷举所有串行化顺序，判定每条历史属于谱系的哪一档
══════════════════════════════════════════════════════════════════════════════
    执行历史                    线性一致  顺序一致  因果一致  读己之写    单调读
    ────────────────────────────────────────────────────────────────────────────
    ① 教科书式的线性一致               ✓         ✓         ✓         ✓         ✓
    ② 重叠区间里读到旧值               ✓         ✓         ✓         ✓         ✓
    ③ 陈旧读                           ✗         ✓         ✓         ✓         ✓
    ④ 两个观察者，相反的顺序           ✗         ✗         ✓         ✓         ✗
    ⑤ 因果倒置                         ✗         ✗
…（截断，原长 1749 字符）
````

</details>

▸ **Bash** Update README and assemble Part 2

````
cd 〔路径00054〕 && cp ch02/ch02.html courseware/ch02-replication.html && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)

rep('''open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
# Linux 用 xdg-open''','''open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
open courseware/ch02-replication.html    # Part 2：复制与一致性模型
# Linux 用 xdg-open''')
rep('''cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC''',
    '''cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC
cd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT''')
rep('''cd java/lab01 && java Lab01.java''','''cd java/lab01 && java Lab01.java
cd java/lab02 && java Lab02.java''')
rep('''| 2 | 复制与一致性模型 —— 一致性谱系、CAP/P
…（截断，原长 2903 字符）
````

<details><summary>↳ Bash 结果</summary>

````
Traceback (most recent call last):
  File "<stdin>", line 26, in <module>
  File "<stdin>", line 3, in rep
AssertionError: │   ├── ch01-time-order.html # Part 1 课件（时空图编辑器 + LWW 复现 + H
ch00-intro.html
ch01-time-order.html
ch02-replication.html
````

</details>

▸ **Bash** Inspect README tree section

````
grep -n 'ch01-time-order.html' README.md | head; echo "---"; sed -n '/^├── courseware/,/^└── java/p' README.md
````

<details><summary>↳ Bash 结果</summary>

````
14:open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
79:│   └── ch01-time-order.html # Part 1 课件（时空图编辑器 + LWW 复现 + HLC，10 道自测题）
---
├── courseware/
│   ├── ch00-intro.html      # 第 0 章课件（6 个模拟实验 + 12 道自测题）
│   └── ch01-time-order.html # Part 1 课件（时空图编辑器 + LWW 复现 + HLC，10 道自测题）
├── go/
│   ├── go.mod
│   ├── internal/tui/        # 中英混排的终端表格对齐工具（各 Lab 共用）
│   ├── lab00/
│   │   ├── net.go           # 不可靠信道：丢包 / 延迟 / 乱序 / 重复 + 第三态
│   │   ├── bank.go          # 服务端：三种扣款实现（天真 / 幂等 / 非原子陷阱）
│   │   ├── client.go        # 客户端重试策略：at-most-once vs at-least-once
│   │   └── main.go          # 四个实验的编排与输出
│   └── lab01/
│       ├── clocks.go        # Lamport 时钟 / 向量时钟 / HLC 三种实现
│       ├── trace.go         # 事件轨迹引擎 + happens-before 可达性校验
│       └── main.go          # 四个实验的编排与输出
└── java/
````

</details>

▸ **Bash** Update README for Part 2

````
python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)

rep('''open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
# Linux 用 xdg-open''','''open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
open courseware/ch02-replication.html   # Part 2：复制与一致性模型
# Linux 用 xdg-open''')
rep('cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC',
    'cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC\ncd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT')
rep('cd java/lab01 && java Lab01.java','cd java/lab01 && java Lab01.java\ncd java/lab02 && java Lab02.java')
rep('| 2 | 复制与一致性模型 —— 一致性谱系、CAP/PACELC、Quorum、读修复、CRDT | Lab 2：Quorum KV + CRDT 计数器 | ⏳ |',
    '| **2** | **复制与一致性模型** —— 复制拓扑
…（截断，原长 2800 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README ok
〔路径00056〕
〔路径00071〕
〔路径00072〕
〔路径00073〕
〔路径00081〕
〔路径00083〕
〔路径00086〕
〔路径00087〕
〔路径00088〕
〔路径00089〕
〔路径00092〕
〔路径00093〕
〔路径00094〕
〔路径00096〕
〔路径00097〕
〔路径00098〕
〔路径00099〕
〔路径00125〕
〔路径00127〕
〔路径00131〕
````

</details>

📎 **发送文件**：Part 2 交互式课件。核心是「实验 2 · 一致性判定器」——它真的在页面里跑穷举算法，给一条执行历史就告诉你它属于谱系哪一档，并把找到的合法串行化顺序列出来。引子直接从你的 ON CONFLICT 讲起。
- `〔路径00073〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00073〕 → file_uuid: 〔标识02224〕
````

</details>

📎 **发送文件**：Lab 2 代码（Go + Java）。`history.go` 里 CheckLinearizable 和 CheckSequential 的唯一差别就在那一个 canPlace 闭包——那就是"是否尊重真实时间"的全部含义。
- `〔路径00056〕`
- `〔路径00099〕`
- `〔路径00096〕`
- `〔路径00097〕`
- `〔路径00098〕`
- `〔路径00131〕`

<details><summary>↳ SendUserFile 结果</summary>

````
6 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识02960〕
  〔路径00099〕 → file_uuid: 〔标识03459〕
  〔路径00096〕 → file_uuid: 〔标识02805〕
  〔路径00097〕 → file_uuid: 〔标识01857〕
  〔路径00098〕 → file_uuid: 〔标识01294〕
  〔路径00131〕 → file_uuid: 〔标识00692〕
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00153〕\" \"$HOME〔路径00161〕\" && echo ready"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果（错误）</summary>

````
The device this session is bound to is not connected to the bridge.
The connection to the user's computer is often restored within a few seconds. If you have not already retried this tool call, retry it once now. If a retry also can't reach it, stop retrying and tell the user you can't reach their computer right now.
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00153〕\" \"$HOME〔路径00161〕\" && echo ready"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果（错误）</summary>

````
The device this session is bound to is not connected to the bridge.
The connection to the user's computer is often restored within a few seconds. If you have not already retried this tool call, retry it once now. If a retry also can't reach it, stop retrying and tell the user you can't reach their computer right now.
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #9 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #10 status
````

</details>

▸ **TaskUpdate** 课件与代码已交付到对话；等你的电脑重新连上后，我再把 Part 2 的文件写进本地仓库并建 artifact

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #11 description, status
````

</details>

Part 2 做完了。**但你的电脑还是连不上**，所以这次的文件只到了对话里（卡片可以直接打开和下载），没能写进 `~/Code/distributed-systems-course/`，artifact 也还没建。等你把桌面 App 打开，跟我说一声，我一次性补上。

**这一章我做了一件之前没做过的事**：课件里的「实验 2 · 一致性判定器」不是查表，是真的在页面里跑穷举——给一条执行历史，它枚举所有可能的串行化顺序，判定属于谱系哪一档，并把找到的合法顺序列出来。Go 和 Java 的 Lab 2-3 是同一套算法的另一份实现，跑出来的判定矩阵和浏览器里逐格一致。

**三个我认为最值得看的点：**

1. **`CheckLinearizable` 和 `CheckSequential` 只差一个闭包。** 两个函数结构完全相同，唯一的区别是那个 `canPlace`：前者检查"有没有未放置的操作必须排在你前面（按真实时间）"，后者检查"有没有同一客户端的更早操作还没放"。**线性一致和顺序一致的全部差别，就是这几行代码。** 我一直想找一个方式把这句话变成可验证的，这是最好的形式。

2. **实验 2 的场景④会同时给出「因果一致 ✓」和「单调读 ✗」**，看起来矛盾。我在页面上专门写了一段解释：两个判定用的**版本序不是同一个**——单调读按真实时间给写排版本，因果一致根本不承认真实时间。这不是 bug，而是"单调读是相对于某个版本序定义的会话保证"这句话的具体样子。真实系统里那个版本序通常由存储层给（LSN、事务号），不是墙上时钟。

3. **Lab 2-4 的数字很扎眼**：1000 次随机分区 + 随机合并顺序，G-Counter 结果 100% 精确，LWW **丢掉 66.5% 的写入**——而且 LWW 每一次也都"收敛"了（三副本值相同）。收敛和正确是两回事，这个数字把它钉死了。

§2.7 我写了一份针对 `〔内部表X〕` 的四问清单，第一问就是"真的是单写者吗"。等你的电脑连上，我想去看一眼 〔内部服务A〕 的实际写法和部署副本数——那会决定后面三问怎么答。

---

两个问题，一个技术的一个安排上的：

1. **〔内部服务A〕 是单实例还是多副本？** 以及那条 `〔内部表Y〕 → 〔内部表X〕` 的链路是怎么触发增量的——定时轮询 `updated_at`、消息队列、还是全量重算？
2. **Part 3 你想怎么走？** 那是整条路线的分水岭，Lab 3 是用 Go 从零写一个能跑的 Raft（对标 MIT 6.5840 Lab 2）。我可以：
   - **一次性给完**（课件 + 完整 Raft 骨架 + 测试），你自己填空
   - **拆成三次**（选举 → 日志复制 → 持久化与快照），每次先讲透再写，中间穿插问答

第二种慢，但 Raft 是那种"跳着看就会漏掉安全性论证"的东西。我建议拆开，你定。

> ✓ 本轮结束 · 2026-08-31 14:07 · 34 步
