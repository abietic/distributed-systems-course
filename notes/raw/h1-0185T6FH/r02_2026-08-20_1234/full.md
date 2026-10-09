## 👤 你 · 2026-08-20 12:34

1. 不能，也可能是本身服务就没执行完（如执行过程中遇到stop the world）或者是由于网络原因下游的返回没有送到，甚至是执行中下游服务直接宕机。这个时候如果下游可以保证接口幂等或者是有其它验证下游是否触发的方式（如共享的数据库表状态），可以等待一段时间后重试或者返回执行任务失败。
2. 因为这些分布式系统一般涉及到投票选主和数据同步多数完成这些动作，如果4台那么就相当于要有3台达成一致，总体来说不太经济。所以一般是3台或5台这样多数就是2台和3台。
3. 这个我一时没有印象，只记得之前看redis和nginx好像都提到过时钟偏移和单调时钟防止时间明明向后但是实际值却回拨了的情况。我想这个在实际场景会影响到可见顺序或者超时剔除？

▸ **ToolSearch**

````
{"max_results": 3, "query": "select:WebSearch,WebFetch"}
````

<details><summary>↳ ToolSearch 结果</summary>

````
[tool_reference]
[tool_reference]
````

</details>

▸ **WebSearch**

````
{"query": "Redis monotonic clock monotonic.c introduced version 5.0 getMonotonicUs"}
````

▸ **WebSearch**

````
{"query": "nginx ngx_monotonic_time clock_gettime CLOCK_MONOTONIC timers version 1.11"}
````

<details><summary>↳ WebSearch 结果</summary>

````
Web search results for query: "Redis monotonic clock monotonic.c introduced version 5.0 getMonotonicUs"

Links: [{"title":"Monotonic clock and updates to AE by JimB123 · Pull Request #7644 · redis/redis","url":"https://github.com/redis/redis/pull/7644"},{"title":"Use H/W Monotonic clock and updates to AE (#7644) · redis/redis@c01e94a","url":"https://github.com/redis/redis/commit/c01e94a4319c416c4c231ffbea9e778d52424e30?branch=c01e94a4319c416c4c231ffbea9e778d52424e30"},{"title":"Monotonic Clock Going Backwards · Issue #11759 · redis/redis","url":"https://github.com/redis/redis/issues/11759"},{"title":"redis: src/monotonic.c Source File - doxygen documentation | Fossies Dox","url":"https://fossies.org/dox/redis-7.0.4/monotonic_8c_source.html"},{"title":"redis/monotonic.c at unstable · redis/redis","url":"https://github.com/redis/redis/blob/unstable/src/monotonic.c"},{"title":"redis-learning/monotonic.h at master · zhl3105645/redis-learning","url":"https://github.com/zhl3105645/redis-learning/blob/master/monotonic.h"},{"title":"Add USE_PROCESSOR_CLOCK for RISC-V to improve monotonic clock support by huangzhengx · Pull Request #14251 · redis/redis","url":"https://github.com/redis/redis/pull/14251"},{"title":"Added a monotonic clock read. by Dvanderploeg · Pull Request #4694 · redis/redis","url":"https://github.com/redis/redis/pull/4694"},{"title":"[patch 13/34] introduce get_clock_monotonic","url":"https://lkml.iu.edu/hypermail/linux/kernel/0908.1/02481.html"}]


REMINDER: You MU
…（截断，原长 1584 字符）
````

</details>

<details><summary>↳ WebSearch 结果</summary>

````
Web search results for query: "nginx ngx_monotonic_time clock_gettime CLOCK_MONOTONIC timers version 1.11"

Links: [{"title":"#189 (timeouts break when time changes) – nginx","url":"https://trac.nginx.org/nginx/ticket/189"},{"title":"unit/src/nxt_time.c at master · nginx/unit","url":"https://github.com/nginx/unit/blob/master/src/nxt_time.c"},{"title":"Changes with nginx 1.14.2 04 Dec 2018","url":"https://nginx.org/en/CHANGES-1.14"},{"title":"Changes with nginx 1.22.1 19 Oct 2022","url":"http://nginx.org/en/CHANGES-1.22"},{"title":"nginx/src/core/ngx_times.c at master · nginx/nginx","url":"https://github.com/nginx/nginx/blob/master/src/core/ngx_times.c"},{"title":"#989 (Use monotonic time for event handlers) – nginx","url":"https://trac.nginx.org/nginx/ticket/989"},{"title":"#1308 (nginx timer don't work if change system date.) – nginx","url":"https://trac.nginx.org/nginx/ticket/1308"},{"title":"ngx.timer.at will not keep correct if system time is changed · Issue #707 · openresty/lua-nginx-module","url":"https://github.com/openresty/lua-nginx-module/issues/707"},{"title":"clock_gettime(MONOTONIC) failed · Issue #894 nginx-proxy/acme-companion","url":"https://github.com/nginx-proxy/acme-companion/issues/894"}]


REMINDER: You MUST include the sources above in your response to the user using markdown hyperlinks.
````

</details>

▸ **WebFetch**

````
{"prompt": "Does this file use CLOCK_MONOTONIC? Show the ngx_monotonic_time function and what ngx_current_msec is set from. Quote the relevant code lines and any #if NGX_HAVE_CLOCK_MONOTONIC guards.", "url": "https://github.com/nginx/nginx/blob/master/src/core/ngx_times.c"}
````

▸ **WebFetch**

````
{"prompt": "What Redis version / release did this monotonic clock PR land in? What is the monotonic clock used for in Redis according to this PR (event loop, activeExpireCycle, etc.)? Does it change key expiration to monotonic time?", "url": "https://github.com/redis/redis/pull/7644"}
````

<details><summary>↳ WebFetch 结果</summary>

````
# CLOCK_MONOTONIC Usage in ngx_times.c

Yes, this file uses `CLOCK_MONOTONIC` conditionally. Here are the relevant sections:

## The ngx_monotonic_time Function

```c
static ngx_msec_t
ngx_monotonic_time(time_t sec, ngx_uint_t msec)
{
#if (NGX_HAVE_CLOCK_MONOTONIC)
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    sec = ts.tv_sec;
    msec = ts.tv_nsec / 1000000;
#endif
    return (ngx_msec_t) sec * 1000 + msec;
}
```

## How ngx_current_msec is Set

In the `ngx_time_update()` function:

```c
ngx_current_msec = ngx_monotonic_time(sec, msec);
```

This call occurs after obtaining the current time via `ngx_gettimeofday()`. When the `NGX_HAVE_CLOCK_MONOTONIC` preprocessor guard is active, the function retrieves monotonic time using `clock_gettime()`. Otherwise, it falls back to using the passed-in `sec` and `msec` values derived from the system's regular time.
````

</details>

<details><summary>↳ WebFetch 结果</summary>

````
# Redis Monotonic Clock PR Analysis

## Redis Version

This PR (#7644) was merged on **August 28, 2020** into the `unstable` branch and was included in **Redis 6.2**.

## Primary Use Cases

According to the PR description, the monotonic clock was primarily implemented for:

1. **Event loop timer processing** - The main application mentioned is refactoring `ae.c` (the event loop) to use monotonic timing for improved performance and reliability.

2. **Timer measurements** - The API provides `getMonotonicUs()` for measuring elapsed time intervals without concerns about system clock adjustments.

The PR states: "The timer processing in `ae.c` has been refactored to use the new monotonic clock interface."

## Key Expiration - NOT Changed

**The monotonic clock does NOT change key expiration mechanisms.** This was explicitly clarified in discussions within the PR:

One reviewer noted that monotonic time "is only useful for determining elapsed time (aka deltas)" and cannot be used for wall-clock purposes like key expiration timestamps.

The expiration mechanism continues using wall-clock time (`mstime`). As one maintainer confirmed: "the expiration mechanism always uses wall-clock (not monotonic)."

This design choice remains a known limitation regarding Redis TTLs and distributed locking scenarios.
````

</details>

▸ **WebFetch**

````
{"prompt": "What is the resolution of this ticket about timeouts breaking when system time changes? Which nginx version or changeset introduced ngx_monotonic_time / CLOCK_MONOTONIC? Quote the closing comment and any version number mentioned.", "url": "https://trac.nginx.org/nginx/ticket/189"}
````

<details><summary>↳ WebFetch 结果</summary>

````
# Resolution of Ticket #189

**Resolution:** The ticket was resolved by implementing monotonic time support for nginx's timer handling.

**Closing Comment:**
> "Core: ngx_current_msec now uses monotonic time if available. When clock_gettime(CLOCK_MONOTONIC) (or faster variants, _FAST on FreeBSD, and _COARSE on Linux) is available, we now use it for ngx_current_msec."

**Version/Changeset:** The fix was introduced in changeset `7222:81fae70d6cb8` and the ticket was closed on March 1, 2018. No specific nginx version number is mentioned in the closing comment, but this appears to have been committed to the master branch.

**Key Point:** Rather than creating a separate monotonic timer variable, the solution directly modified `ngx_current_msec` to leverage monotonic clock sources when available, with platform-specific optimizations (CLOCK_MONOTONIC_FAST for FreeBSD and CLOCK_MONOTONIC_COARSE for Linux).
````

</details>

## 🤖 Claude · 2026-08-20 12:37

先批改，再往下走。

## 第 1 题 — 答得比"零基础"该有的水平好

四种可能性你都列全了（没到 / 执行中 STW / 返回丢 / 执行中宕机）。三点补强：

**① "反查下游状态"这个手段本身也有第三态。** 查询也是一次 RPC，也会超时。你没有消除不确定性，只是降低了概率并推迟了它。更要命的是竞态：你查的那一瞬间是"未执行"，但那笔请求可能正卡在网络队列里，你查完之后才到达服务端 —— 这是典型的 check-then-act。所以优先级是：**幂等键（唯一索引，判定在存储层原子完成）> 反查 > 猜**。反查是对账/定时补偿的手段，不该做主路径。

**② "等待一段时间后重试"要加指数退避 + 随机抖动。** 固定间隔重试在下游抖动时会形成重试风暴，把一次 200ms 的 GC 停顿放大成雪崩。再加重试预算和熔断，超过阈值就停手。

**③ "或者返回执行任务失败"这句话把第三态向上游传播了。** 你的调用方收到"失败"，它同样不知道你到底做没做，于是它也会重试。第三态在调用链上会累积。终结它只有两条路：每一跳都幂等，或者用补偿事务（Saga）把已做的撤销。这是 Part 4 的主线。

## 第 2 题 — 结论对，但"不经济"背后有个更硬的理由

先把"不经济"量化：N=3 → quorum 2 → 容忍 1 台；**N=4 → quorum 3 → 还是只容忍 1 台**；N=5 → quorum 3 → 容忍 2 台。通式 N=2f+1。偶数节点纯亏：容错能力不变，却多了一台会坏的机器，**整体可用性反而下降**。

但"为什么必须是过半"有个比经济性更根本的答案：**任意两个过半集合必然有交集。** 这个交集保证了新 leader 至少有一个节点见过旧 leader 已提交的最新数据 —— 数据不会因为换主而丢。quorum 的本质是*交集性质*，不是"少数服从多数"。想通这点，Raft 的选举限制（candidate 日志不能落后于多数派）就是自然推论，而不是一条要背的规则。

顺带两句：过半只是特例，一般化是 Dynamo 的 **W + R > N**；而很少超过 7 台是因为写要等过半确认，延迟由第 ⌈N/2⌉ 慢的节点决定，扩读能力应该用不投票的 learner / observer。

## 第 3 题 — 你的直觉对，而且两个例子都精确命中

我去核实了源码，因为要给你版本号：

**nginx** 确实如此。trac 上 "timeouts break when time changes" 这张票（#189）挂了很多年，2018 年 3 月（changeset 7222）才修：`ngx_current_msec` 改用 `clock_gettime(CLOCK_MONOTONIC)`，Linux 上取更快的 `_COARSE` 变体。在那之前改系统时间会让 nginx 定时器全乱 —— 正是你猜的"超时剔除"。

**Redis** 也如此，但有个漂亮的分界。monotonic clock 是 PR #7644（2020 年 8 月合入，随 **6.2** 发布），`getMonotonicUs()` 用于事件循环 `ae.c` 的定时器和耗时测量。**但 key 的过期判断至今仍用墙上时钟 `mstime()`** —— 这不是偷懒，是必须的：

| 用途 | 用哪种时钟 | 为什么 |
|---|---|---|
| 测耗时、超时、心跳、限流窗口 | **单调时钟** | 只关心"过了多久"，绝对值无意义，必须免疫 NTP 跳变 |
| TTL / 过期时间 / 定时任务 | **墙上时钟** | 语义就是"到某个绝对时刻作废"，要跨重启、跨主从一致 |
| 判断两台机器上事件的先后 | **两种都不行** | 需要逻辑时钟 —— 这就是 Part 1 |

第三行就是你说的"可见顺序"，而且比你想的更严重。Redis 主从复制里，**从库不会自己删过期 key，必须等主库发 DEL 过来**。为什么？因为主从时钟哪怕差几十毫秒，从库自作主张删了就会出现"主库读得到、从库读不到"的分歧。Redis 宁可让从库暂时留着一个逻辑上已过期的 key，也不允许两边各自根据本地时钟做决定 —— 这就是**可见性**问题的一个具体实例。

同一个坑的更贵版本是 Cassandra 的 last-write-wins：两个节点时钟差 100ms，后写的那条被判成旧值静默丢弃，不留任何日志。就是你在实验 3 里拖出来的那个现象。

---

三题合起来看，你的短板不在直觉，在**词汇和边界条件**——知道会出事，但说不出"为什么必然出事"和"什么条件下不会出事"。所以 Part 1 我会把重点放在给你一套能精确表达的工具。现在去做 Part 1：时间、顺序与因果。

Sources: [nginx ticket #189](https://trac.nginx.org/nginx/ticket/189) · [nginx ngx_times.c](https://github.com/nginx/nginx/blob/master/src/core/ngx_times.c) · [Redis PR #7644](https://github.com/redis/redis/pull/7644)

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **TaskCreate** 编写 Part 1 交互式课件：时间、顺序与因果

<details><summary>↳ TaskCreate 结果</summary>

````
Task #6 created successfully: 编写 Part 1 交互式课件：时间、顺序与因果
````

</details>

▸ **TaskCreate** 编写 Lab 1：逻辑时钟实现（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #7 created successfully: 编写 Lab 1：逻辑时钟实现（Go + Java）
````

</details>

▸ **TaskCreate** 验证并交付 Part 1

<details><summary>↳ TaskCreate 结果</summary>

````
Task #8 created successfully: 验证并交付 Part 1
````

</details>

▸ **Bash** Extract shared CSS from chapter 0

````
cd 〔路径00054〕 && mkdir -p ch01 && python3 - <<'PY'
src=open('courseware/ch00-intro.html',encoding='utf-8').read()
css=src[src.index('<style>'):src.index('</style>')+8]
open('〔路径00240〕','w',encoding='utf-8').write(css)
print(len(css),"bytes of CSS extracted")
PY
````

<details><summary>↳ Bash 结果</summary>

````
10484 bytes of CSS extracted
````

</details>

▸ **Bash** Build Part 1 head with shared CSS

````
cd 〔路径00059〕 && cat > head.tmp <<'EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 1 · 时间、顺序与因果</title>
EOF
cat 〔路径00240〕 >> head.tmp
cat >> head.tmp <<'EOF'
<style>
/* Part 1 专用 */
.grid-wrap{overflow-x:auto}
.modes{display:flex;gap:6px;flex-wrap:wrap}
.modes button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.rel{display:grid;grid-template-columns:auto 1fr;gap:8px 16px;font-size:14px;align-items:baseline}
.rel dt{color:var(--ink-3);font-weight:600;white-space:nowrap}
.rel dd{margin:0;color:var(--ink-2)}
.vt{font-family:var(--mono);font-size:12.5px;color:var(--ink)}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0
…（截断，原长 2222 字符）
````

<details><summary>↳ Bash 结果</summary>

````
12790 ch01.html
````

</details>

▸ **Bash** Write Part 1 sections 1.0-1.2

````
cat >> ch01.html <<'EOF'
<section id="s0">
  <div class="eyebrow">Part 1 · Time, Order &amp; Causality</div>
  <h1>没有全局时钟，怎么说"这件事在那件事之前"</h1>
  <p class="lead">你在第 0 章的自测里说对了一句关键的话：单调时钟能防止"时间明明向后却回拨"，而这会影响<em>可见顺序</em>和<em>超时剔除</em>。这一章把那句话拆成两半——超时剔除属于单调时钟的领域，已经解决了；<strong>可见顺序不属于任何一种物理时钟</strong>，它需要一个全新的东西。</p>

  <table>
    <thead><tr><th style="width:32%">用途</th><th style="width:20%">用哪种时钟</th><th>为什么</th></tr></thead>
    <tbody>
      <tr><td>测耗时、超时、心跳、限流窗口</td><td><strong>单调时钟</strong></td><td>只关心"过了多久"，绝对值无意义，必须免疫 NTP 跳变。<span style="color:var(--ink-3)">nginx 2018 年才把 <code>ngx_current_msec</code> 改成 CLOCK_MONOTONIC；Redis 6.2 才引入 <code>getMonotonicUs()</code>。</span></td></tr>
      <tr><td>TTL / 过期时间 / 定时任务</td><td><strong>墙上时钟</strong></td><td>语义就是"到某个绝对时刻作废"，必须跨进程重启、跨主从一致。<span
…（截断，原长 6567 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 1 sections 1.3-1.5

````
cat >> ch01.html <<'EOF'
<section id="s3">
  <h2>1.3 Lamport 逻辑时钟</h2>
  <p>happens-before 是一个关系，不是一个数。Lamport 接着问：能不能给每个事件贴一个整数 <code>L(e)</code>，让这个数<em>尊重</em>因果关系？算法只有三行：</p>
  <div class="axm"><div class="b">1</div><div>每个进程维护一个计数器 <code>L</code>，初始为 0。</div></div>
  <div class="axm"><div class="b">2</div><div><strong>任何本地事件（含发送）前</strong>：<code>L = L + 1</code>，把 <code>L</code> 作为该事件的时间戳。发送消息时把 <code>L</code> 一起带上。</div></div>
  <div class="axm"><div class="b">3</div><div><strong>收到携带 <code>Lm</code> 的消息时</strong>：<code>L = max(L, Lm) + 1</code>，作为接收事件的时间戳。</div></div>

<pre><code><span class="cm">// Go：完整实现只有这么多</span>
<span class="kw">type</span> Lamport <span class="kw">struct</span>{ mu sync.Mutex; t <span class="kw">uint64</span> }

<span class="kw">func</span> (l *Lamport) <spa
…（截断，原长 7511 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 1 sections 1.6-1.10

````
cat >> ch01.html <<'EOF'
<section id="s6">
  <h2>1.6 混合逻辑时钟 HLC：把物理时间赎回来</h2>
  <p>逻辑时钟解决了因果，但丢掉了一样东西：<strong>可读性和可比性</strong>。Lamport 值 <code>84213</code> 对应现实中的什么时刻？不知道。你没法问"给我 10:00 那一刻的数据快照"，也没法给日志排序时和监控图表对齐。</p>
  <p>HLC（Hybrid Logical Clock，2014）的想法很聪明：<strong>时间戳是二元组 <code>(l, c)</code></strong>，<code>l</code> 尽量贴着物理时钟走，<code>c</code> 是一个小计数器，只在物理时钟"不够用"时才递增。</p>
  <div class="axm"><div class="b">1</div><div><strong>本地事件 / 发送</strong>：<code>l' = max(l, pt)</code>（pt 是当前物理时钟）。若 <code>l' == l</code> 则 <code>c++</code>，否则 <code>c = 0</code>。</div></div>
  <div class="axm"><div class="b">2</div><div><strong>收到 <code>(lm, cm)</code></strong>：<code>l' = max(l, lm, pt)</code>。<br>若 <code>l' == l == lm</code> 则 <code>c = max(c, cm) + 1</code>；若 <code>l' == l</code> 则 <code>c++</code>；若 <cod
…（截断，原长 10950 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 1 JS: spacetime diagram editor

````
cat >> ch01.html <<'EOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(t,a,x){const e=document.createElementNS(NS,t);for(const k in a)e.setAttribute(k,a[k]);if(x!=null)e.textContent=x;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}

/* ===== 主题 / 进度 / 目录 ===== */
$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");draw1();draw2();};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHei
…（截断，原长 9720 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Finish Part 1 JS and close document

````
cat >> ch01.html <<'EOF'

/* ══════════════════════════════════════════════════════════════════════
   实验 2 · LWW 静默丢写 vs 版本向量
   ══════════════════════════════════════════════════════════════════════ */
const S2={played:false};
function p2(){ return {skew:+$("#r2o").value, gap:+$("#r2g").value}; }
function draw2(){
  const {skew,gap}=p2();
  $("#l2o").textContent=(skew>=0?"+":"")+skew+" ms";
  $("#l2g").textContent=gap+" ms";
  const svg=$("#sv2"); svg.innerHTML="";
  const W=700, L=96, R=660, T0=140, TB=1000;              // 真实时间轴 0..TB ms 映射到 L..R
  const xt=ms=>L+(ms/TB)*(R-L);
  const yA=52, yB=116, yAx=162;
  // 轴
  svg.appendChild(el("line",{x1:L,y1:yAx,x2:R,y2:yAx,stroke:cv("--axis"),"stroke-width":1.5}));
  for(let ms=0;ms<=TB;ms+=200){
    svg.appendChild(el("line",{x1:xt(ms),y1:
…（截断，原长 7367 字符）
````

<details><summary>↳ Bash 结果</summary>

````
868 ch01.html
script tags: 1 1
sections: 11
size kb: 70
````

</details>

▸ **Bash** Verify Part 1 interactives in browser

````
cd /tmp && cat > c1.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:900}});
const errs=[];p.on('pageerror',e=>errs.push('PAGEERROR: '+e.message));
p.on('console',m=>{if(m.type()==='error')errs.push('CONSOLE: '+m.text());});
await p.goto('file://〔路径00060〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(500);

// 检查时空图引擎的时钟计算
const clocks = await p.evaluate(()=>{
  const out=[];
  document.querySelectorAll('#sv1 text').forEach(t=>out.push(t.textContent));
  return out;
});
console.log('sv1 labels:', clocks.filter(x=>x.startsWith('[')).join(' '));

// 比较模式：选 P3@t2 (H
…（截断，原长 3026 字符）
````

<details><summary>↳ Bash 结果</summary>

````
sv1 labels: [1,0,0] [0,1,0] [0,0,1] [2,0,0] [2,2,0] [0,0,2] [3,0,2] [2,3,0] [2,3,3] [4,0,2] [2,3,4]
REL(H,B): 关系判定 | 事件 a | P3@t2 Lamport 1　向量 [0,0,1] | 事件 b | P1@t3 Lamport 2　向量 [2,0,0] | 向量时钟判定 | a ∥ b（并发） —— 谁都没有影响谁　（充要，这是真相） | Lamport 判定 | L(a) < L(b) ⇒ Lamport 认为 a 在前 | 抓到一个 Lamport 时钟的盲区 | 这两个事件实际是并发的，但它们的 Lamport 值一大一小。如果你据此认为「P3@t2 发生在 P1@t3 之前」，你就错了——它们之间没有任何消息路径，谁都不可能影响谁。 | 这正是 §1.3 说的：Lamport 时钟只能证伪，不能证实。
REL(A,K): 关系判定 | 事件 a | P1@t1 Lamport 1　向量 [1,0,0] | 事件 b | P3@t12 Lamport 6　向量 [2,3,4] | 向量时钟判定 | a → b（a happens-before b）　（充要，这是真相） | Lamport 判定 | L(a) < L(b) ⇒ Lamport 认为 a 在前 | 存在一条从 P1@t1 到 P3@t12 的因果路径（沿着进程时间线和消息箭头走得通）。此时 Lamport 判定与向量时钟一致——有因果关系时两者永远不会矛盾。
circles after add+msg: 13
LWW: 策略一 · LAST-WRITE-WINS | A 的时间戳 | 260　= 真实 140 + 偏移 120 | B 的时间戳 | 200　= 真实 200 + 偏移 0 | 最终保留 | 值 A | 后写的 B 被静默丢弃了。C2 明明在 C1 之后 60ms 才写，但 R1 的时钟快了 120ms，于是 A 的时间戳更大。 | 没有报错，没有日志，用户看到"保存成功"然后数据消失。
LWW skew=0: 策略一 · LAST-WRITE-WINS | A 的时间戳 | 140　= 真实 140 + 偏移 0 | B 的时间戳 | 200　= 真实 200 + 偏移 0 | 最终保留 | 值 B | 这一次侥幸正确（时钟偏移 0ms 小于写入间隔 60ms）。但正确与否取决于一个和业务毫无关系的参数——把偏移拖大一点就翻转了。
HLC rows: #1 12067 1 (12067, 0) +0 ms || #2 12144 2 (12144, 0) +0 ms || #3 12192 3 (12192, 0) +0 ms || #4 回拨后 11692 ⟵ 倒退！ 4 (12192, 1) +500 ms || #5 11773 5 (12192, 2) +419 ms || #6 11798 6 (12192, 3) +394 ms || #7 11866 7 (12192, 4) +326 ms
HLC verdict: ⚠ 纯物理时钟出现了 1 次倒退。而 Lamport 与 HLC 两列始终单调递增——注意 HLC 的 l 在回拨期间冻结不动、靠 c 递增撑住顺序，等物理时钟追上来后 c 自动归零。
ERRORS: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Screenshot Part 1 experiments

````
cd /tmp && cat > s1.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:900}});
await p.goto('file://〔路径00060〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(400);
async function cell(pp,tt){const q=await p.evaluate(({pp,tt})=>{const s=document.querySelector('#sv1'),r=s.getBoundingClientRect(),v=s.viewBox.baseVal;
  const X0=78,X1=730,C=13,DX=(X1-X0)/(C-1),Y=[62,140,218];return {x:r.left+(X0+tt*DX)*(r.width/v.width),y:r.top+Y[pp]*(r.height/v.height)};},{pp,tt});
  await p.mouse.click(q.x,q.y);}
await p.click('#m1 button[data-m="cmp"]'); await cell(2,2); await cell(0,
…（截断，原长 1300 字符）
````

<details><summary>↳ Bash 结果</summary>

````
done
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00214〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Fix spacetime diagram viewBox height

````
cd 〔路径00059〕 && sed -i 's|viewBox="0 0 760 260" style="width:100%;min-width:660px|viewBox="0 0 760 278" style="width:100%;min-width:660px|' ch01.html && grep -o 'viewBox="0 0 760 [0-9]*"' ch01.html
````

<details><summary>↳ Bash 结果</summary>

````
viewBox="0 0 760 278"
````

</details>

▸ **Bash** Extract shared TUI package and refactor lab00

````
python3 - <<'PY'
p='ch01.html'; s=open(p,encoding='utf-8').read()
old='<td><strong>Lab 1-3</strong><br>LWW 丢写</td><td>两个副本 + 可配置时钟偏移，跑 N 次并发写，统计<strong>有多少次写入被静默丢弃</strong>；再用版本向量重跑，统计检测出的冲突数——两个数字应该吻合。</td>'
new='<td><strong>Lab 1-3</strong><br>LWW 丢写</td><td>两个副本 + 可配置时钟偏移，跑 N 轮并发写。你会发现一个更狠的事实：<strong>LWW 每一轮都丢掉一条写入</strong>（因为它必须二选一），其中一部分丢的还是真实更晚的那条。换成版本向量重跑，冲突全部被检出，<strong>丢弃数为 0</strong>。</td>'
assert old in s; s=s.replace(old,new); open(p,'w',encoding='utf-8').write(s); print('ok')
PY
mkdir -p 〔路径00082〕 && cat > 〔路径00083〕 <<'EOF'
// Package tui 提供中英混排的终端对齐工具，供各个 Lab 共用。
package tui

import (
	"fmt"
	"strings"
)

// Dispw 计算字符串在等宽终端里占几列（CJK 全角字符占 2 列）。
// Go 的 %-20s 是按 rune 计数的，中文表格会错位，所以自己算。
func Dispw(s string) int {
	w := 0
	for
…（截断，原长 3337 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
rewrote lab00/main.go
````

</details>

▸ **Bash** Check Go build after refactor

````
gofmt -l . ; go vet ./... 2>&1 | head -30; echo "---build---"; go build ./... 2>&1 | head -20
````

<details><summary>↳ Bash 结果</summary>

````
---build---
````

</details>

▸ **Bash** Verify lab00 still runs after refactor

````
cd lab00 && go run . -n 20 2>&1 | sed -n '1,16p;20,30p'
````

<details><summary>↳ Bash 结果</summary>

````
配置：转账 20 笔 × 100 元 ｜ 单向丢包率 30% ｜ 最大重试 3 次 ｜ 超时 5ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 0-1 · 不可靠信道：丢包 / 延迟 / 乱序 / 重复
  按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么
══════════════════════════════════════════════════════════════════════════════
    发送顺序                       1 2 3 4 5 6 7 8 9 10
    实际到达顺序                   1 2 1* 3 4 4* 6 5 5* 7 8 8* 9        （* = 重复投递）
    丢失的消息                     10
    重复到达的消息                 1 4 5 8
    乱序                           5 排在 6 之后
    用时 12ms

    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
  Lab 0-2 · at-most-once vs at-least-once
  同样的网络故障，两种重试策略，两种错法
══════════════════════════════════════════════════════════════════════════════
    策略                         应扣(元)   实扣(元)   差额(元)   客户端认为成功
    ────────────────────────────────────────────────────────────────────────────
    A · 不重试                       2000       1000      -1000             7/20
    B · 重试至多 3 次                2000       2400       +400            18/20

    ▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
      已经在服务器上执行了（幽灵成功），只是响应包丢了。
    ▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
````

</details>

▸ **Bash** Write Go clock implementations for Lab 1

````
cd 〔路径00054〕 && python3 - <<'PY'
p='ch01/ch01.html'; s=open(p,encoding='utf-8').read()
old='把 <code>-skew</code> 设成 0，LWW 还会丢数据吗？<em>先猜再跑。</em>（提示：想想两次写入的真实间隔和网络传输时间的关系——即使时钟完全同步，"打时间戳的时刻"和"到达副本的时刻"也不是一回事。）'
new='把 <code>-skew</code> 设成 0（时钟完全同步），LWW 还会丢数据吗？<em>先猜再跑。</em><br><span style="font-size:13px;color:var(--ink-3)">提示：想清楚 LWW 面对两条<strong>并发</strong>写入时，有没有可能把两条都保留下来。</span>'
assert old in s; open(p,'w',encoding='utf-8').write(s.replace(old,new)); print('ok')
PY
mkdir -p go/lab01 && cat > go/lab01/clocks.go <<'EOF'
package main

import (
	"fmt"
	"strings"
)

// ════════════════════════════════════════════════════════════════════════
// Lamport 逻辑时钟
//
//	保证：a → b  ⟹  L(a) < L(b)
//	不保证：L(a) < L(b) ⇒ a → b     ← 这就是它的根本局限
//
// ══════════════════════════════════════════════════
…（截断，原长 3612 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
clocks.go ok
````

</details>

▸ **Bash** Write event trace engine for Lab 1

````
cd 〔路径00091〕 && cat > trace.go <<'EOF'
package main

import (
	"fmt"
	"sort"
)

// ════════════════════════════════════════════════════════════════════════
// 事件轨迹：和课件「实验 1 · 时空图编辑器」里的「载入经典示例」完全一致。
// 你在浏览器里看到的每一个数字，这里都能跑出来。
// ════════════════════════════════════════════════════════════════════════

const NP = 3 // 三个进程

var PN = [NP]string{"P1", "P2", "P3"}

type Kind int

const (
	KLocal Kind = iota
	KSend
	KRecv
)

type Ev struct {
	ID   int
	P    int // 所在进程
	T    int // 时空图上的列（只用于排版和拓扑排序，不是物理时钟！）
	Kind Kind
	Peer int // 配对事件的 ID（发送 ↔ 接收）

	L uint64 // Lamport 时间戳
	V Vector // 向量时间戳
}

func (e *Ev) Name() string { return fmt.Sprintf("%s@t%d", PN[e.P], e.T) }

func (e *Ev) KindName() string {
	return [...]string{"本地事件", "发送消息", "接收消息"}[e.Kind]
}

// BuildTrace 构造经典示例场景并
…（截断，原长 2933 字符）
````

<details><summary>↳ Bash 结果</summary>

````
trace-ok
````

</details>

▸ **Bash** Write and run Go Lab 1

````
cat > main.go <<'EOF'
// Lab 1 · 时间、顺序与因果
//
// 运行：  cd go/lab01 && go run .
// 调参：  go run . -skew 200 -gap 50 -n 500
//
// 配套课件：courseware/ch01-time-order.html
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"dsc/internal/tui"
)

var (
	nFlag    = flag.Int("n", 200, "Lab 1-3 的并发写轮数")
	skewFlag = flag.Int64("skew", 120, "副本 R1 的时钟偏移（毫秒）")
	gapFlag  = flag.Int64("gap", 60, "两次写入的真实间隔（毫秒）")
	jitFlag  = flag.Int64("jit", 40, "时钟偏移的随机抖动（毫秒）")
	seedFlag = flag.Int64("seed", 42, "随机种子")
)

func main() {
	flag.Parse()
	fmt.Printf("\n配置：并发写 %d 轮 ｜ R1 时钟偏移 %+dms（抖动 ±%dms）｜ 写入真实间隔 %dms ｜ 种子 %d\n",
		*nFlag, *skewFlag, *jitFlag, *gapFlag, *seedFlag)
	lab11()
	lab12()
	lab13()
	lab14()
	epilogue()
}

// ═══════════════════════════════════════════════════════════════════════
// Lab 1-1
…（截断，原长 7606 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：并发写 50 轮 ｜ R1 时钟偏移 +120ms（抖动 ±40ms）｜ 写入真实间隔 60ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 1-1 · Lamport 时钟与向量时钟
  和课件「实验 1」的经典示例完全一致 —— 浏览器里看到的数字，这里能跑出来
══════════════════════════════════════════════════════════════════════════════
    事件                类型        对端     Lamport        向量时钟
    ────────────────────────────────────────────────────────────────
    P1@t1           本地事件           —           1         [1,0,0]
    P2@t0           本地事件           —           1         [0,1,0]
    P3@t2           本地事件           —           1         [0,0,1]
    P1@t3           发送消息       P2@t5           2         [2,0,0]
    P2@t5           接收消息       P1@t3           3         [2,2,0]
    P3@t6           发送消息       P1@t8           2         [0,0,2]
    P1@t8           接收消息       P3@t6           3         [3,0,2]
    P2@t7           发送消息       P3@t9           4         [2,3,0]
    P3@t9           接收消息       P2@t7           5         [2,3,3]
    P1@t11          本地事件           —           4         [4,0,2]
    P3@t12          本地事件           —           6         [2,3,4]

    ▸ 注意 P1@t8 的向量是 [3,0,2] —— 第 2 位是 2，说明 P1 通过那条来自 P3 的消息
      "得知"了 P3 已经走了 2 步。向量时钟的每一位就是"我所知道的它走到哪了"。

══════════════════════════════════════════════════════════════════════════════
  Lab 1-2 · Lamport 的盲区
  枚举所有事件对，看两种时钟在哪些对上给出不同答案
══════════════════════════════════════════════════════════════════════════════
    事件对总数                     55
    有因果关系（a→b 或 b
…（截断，原长 2470 字符）
````

</details>

▸ **Bash** View remaining Lab 1 output

````
go run . -n 50 2>&1 | tail -42
````

<details><summary>↳ Bash 结果</summary>

````
      但丢弃数依然是 100% —— 时钟准不准，改变的只是"丢哪一条"。
    ▸ 版本向量不假装知道答案：它把两条都留下来（siblings），
      把"该怎么合并"这个只有业务代码知道的问题交还给业务代码。
    ▸ 试试 -skew 0 和 -skew 300，看两个数字怎么变。

══════════════════════════════════════════════════════════════════════════════
  Lab 1-4 · 混合逻辑时钟 HLC
  注入一次 NTP 回拨，看三种时钟谁活了下来
══════════════════════════════════════════════════════════════════════════════
    事件            物理时钟    Lamport    HLC (l, c)    l − 物理时钟
    ─────────────────────────────────────────────────────────────────
    #1                 12075          1    (12075, 0)           +0 ms
    #2                 12148          2    (12148, 0)           +0 ms
    #3                 12233          3    (12233, 0)           +0 ms
    #4 *回拨*   11433 ⟵ 倒退          4    (12233, 1)         +800 ms
    #5                 11465          5    (12233, 2)         +768 ms
    #6                 11539          6    (12233, 3)         +694 ms
    #7                 11605          7    (12233, 4)         +628 ms
    #8                 11681          8    (12233, 5)         +552 ms

    物理时钟倒退次数               1 次　← 任何基于它的排序 / TTL / 快照都会出错
    HLC 是否严格单调               是 ✓
    HLC 与物理时钟的最大偏差       800 ms　← 有界，且回拨结束后会自动收敛回 0

    ▸ 回拨发生后，HLC 的 l 冻结在回拨前的值不动，靠 c 递增维持严格单调；
      等物理时钟重新追上来，c 自动归零，l 继续贴着物理时间走。
    ▸ Lamport 也单调，但它的数值和现实时间毫无关系 —— 你没法用它做
      "给我 10:00 那一刻的快照"这类时间范围查询。HLC 两样都要到了，
      代价只是每个时间戳多一个整数。CockroachDB / MongoDB 用的就是它。

══════════════════════════════════════════════════════════════════════════════
  三个思考题（答案
…（截断，原长 1809 字符）
````

</details>

▸ **Bash** Add per-column alignment and fix tables

````
cd 〔路径00079〕 && python3 - <<'PY'
p='internal/tui/tui.go'; s=open(p,encoding='utf-8').read()
old='''// TableHead / TableRow 按显示宽度对齐：第一列左对齐，其余右对齐。
func TableHead(cols []string, w []int) {
	line := "    "
	for i, c := range cols {
		if i == 0 {
			line += PadR(c, w[i])
		} else {
			line += PadL(c, w[i])
		}
	}
	fmt.Println(line)
	t := 0
	for _, x := range w {
		t += x
	}
	fmt.Printf("    %s\\n", strings.Repeat("─", t))
}

func TableRow(cells []string, w []int) {
	line := "    "
	for i, c := range cells {
		if i == 0 {
			line += PadR(c, w[i])
		} else {
			line += PadL(c, w[i])
		}
	}
	fmt.Println(line)
}'''
new='''// TableHead / TableRow 按显示宽度对齐：第一列左对齐，其余右对齐。
func TableHead(cols []string, w []int) { TableHeadA(cols, w, defaultAlign(len(cols))) }
func TableRow(cells []string, w []i
…（截断，原长 3184 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 2
tui ok
main ok
stat ./...: no such file or directory
````

</details>

▸ **Bash** Format, vet and rerun Lab 1

````
cd 〔路径00079〕 && gofmt -w internal lab00 lab01 && go vet ./... && cd lab01 && go run . -n 50 2>&1 | sed -n '35,44p;52,63p'
````

<details><summary>↳ Bash 结果</summary>

````
    ────────────────────────────────────────────────────────────────────────────────────────
    P1@t1 ∥ P3@t6                L=1     L=2  Lamport 误以为 P1@t1 在 P3@t6 之前
    P2@t0 ∥ P1@t3                L=1     L=2  Lamport 误以为 P2@t0 在 P1@t3 之前
    P2@t0 ∥ P3@t6                L=1     L=2  Lamport 误以为 P2@t0 在 P3@t6 之前
    P2@t0 ∥ P1@t8                L=1     L=3  Lamport 误以为 P2@t0 在 P1@t8 之前
    P2@t0 ∥ P1@t11               L=1     L=4  Lamport 误以为 P2@t0 在 P1@t11 之前
    P3@t2 ∥ P1@t3                L=1     L=2  Lamport 误以为 P3@t2 在 P1@t3 之前

    ▸ 向量时钟的判定通过了独立校验：程序另外用 happens-before 的三条规则做了
      一次可达性搜索，两者结论完全一致（不一致会直接 panic）。
                                      last-write-wins    版本向量
    ──────────────────────────────────────────────────────────
    并发写轮数                                  50          50
    检测出冲突                                   0          50
    被丢弃的写入                                50           0
    其中丢的是真实更晚那条                      50           —

    LWW 的"明显错误率"             100.0%（丢掉真实更晚写入的比例）

    ▸ 最容易被忽略的一行是「被丢弃的写入」：LWW 每一轮都丢掉一条，
      因为它必须二选一。把 -skew 设成 0，"明显错误率"会降到 0，
      但丢弃数依然是 100% —— 时钟准不准，改变的只是"丢哪一条"。
````

</details>

▸ **Bash** Retune LWW defaults and verify both regimes

````
sed -i 's|skewFlag = flag.Int64("skew", 120, |skewFlag = flag.Int64("skew", 80, |; s|jitFlag  = flag.Int64("jit", 40, |jitFlag  = flag.Int64("jit", 60, |' main.go && go run . -n 200 2>&1 | sed -n '52,60p' && echo "=== skew 0 ===" && go run . -n 200 -skew 0 2>&1 | sed -n '52,60p' && echo "=== HLC ===" && go run . -n 20 2>&1 | sed -n '68,82p'
````

<details><summary>↳ Bash 结果</summary>

````
                                      last-write-wins    版本向量
    ──────────────────────────────────────────────────────────
    并发写轮数                                 200         200
    检测出冲突                                   0         200
    被丢弃的写入                               200           0
    其中丢的是真实更晚那条                     137           —

    LWW 的"明显错误率"             68.5%（丢掉真实更晚写入的比例）

=== skew 0 ===
                                      last-write-wins    版本向量
    ──────────────────────────────────────────────────────────
    并发写轮数                                 200         200
    检测出冲突                                   0         200
    被丢弃的写入                               200           0
    其中丢的是真实更晚那条                       0           —

    LWW 的"明显错误率"             0.0%（丢掉真实更晚写入的比例）

=== HLC ===
══════════════════════════════════════════════════════════════════════════════
  Lab 1-4 · 混合逻辑时钟 HLC
  注入一次 NTP 回拨，看三种时钟谁活了下来
══════════════════════════════════════════════════════════════════════════════
    事件            物理时钟    Lamport    HLC (l, c)    l − 物理时钟
    ───────────────────────────────────────────────────────────────────────────────────────
    #1                 12075          1    (12075, 0)           +0 ms
    #2                 12148          2    (12148, 0)           +0 ms
    #3                 12233          3    (12233, 0)           +0 ms
    #4 *回拨*          11433          4    (12233, 1)         +800 ms  ⟵ 物理时钟倒退了
    #5                 1146
…（截断，原长 1752 字符）
````

</details>

▸ **Bash** Write and run Java Lab 1

````
sed -i 's|w := \[\]int{12, 12, 11, 14, 16, 22}|w := []int{12, 12, 11, 14, 16, 4}|' main.go && gofmt -w . && go vet . && cat > 〔路径00127〕 <<'EOF'
/*
 * Lab 1 · 时间、顺序与因果（Java 版）
 *
 * 运行：  cd java/lab01 && java Lab01.java
 * 调参：  java Lab01.java -skew 200 -gap 50 -n 500
 *
 * 需要 JDK 17+（单文件源码启动模式，不需要 Maven / Gradle）。
 * 配套课件：courseware/ch01-time-order.html
 */
import java.util.*;

public class Lab01 {

    // ─── 可调参数 ──────────────────────────────────────────────────────────
    static int  N    = 200;  // Lab 1-3 的并发写轮数
    static long SKEW = 80;   // 副本 R1 的时钟偏移（毫秒）
    static long GAP  = 60;   // 两次写入的真实间隔（毫秒）
    static long JIT  = 60;   // 时钟偏移的随机抖动（毫秒）
    static long SEED = 42;

    static final int NP = 3;
    static final String[] PN = {"P1", "P2", "P3"}
…（截断，原长 16097 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
/bin/bash: line 393: 〔路径00128〕 No such file or directory
/bin/bash: line 785: cd: 〔路径00129〕 No such file or directory
````

</details>

▸ **Bash** Create java lab01 directory

````
mkdir -p 〔路径00126〕 && ls -d 〔路径00123〕
````

<details><summary>↳ Bash 结果</summary>

````
〔路径00124〕
〔路径00126〕
````

</details>

▸ **Bash** Confirm Go table width patch applied

````
grep -n 'w := \[\]int{12, 12, 11, 14, 16' 〔路径00093〕
````

<details><summary>↳ Bash 结果</summary>

````
242:	w := []int{12, 12, 11, 14, 16, 4}
````

</details>

▸ **Bash** Write and run Java Lab 1

````
cat > 〔路径00127〕 <<'EOF'
/*
 * Lab 1 · 时间、顺序与因果（Java 版）
 *
 * 运行：  cd java/lab01 && java Lab01.java
 * 调参：  java Lab01.java -skew 200 -gap 50 -n 500
 *
 * 需要 JDK 17+（单文件源码启动模式，不需要 Maven / Gradle）。
 * 配套课件：courseware/ch01-time-order.html
 */
import java.util.*;

public class Lab01 {

    // ─── 可调参数 ──────────────────────────────────────────────────────────
    static int  N    = 200;  // Lab 1-3 的并发写轮数
    static long SKEW = 80;   // 副本 R1 的时钟偏移（毫秒）
    static long GAP  = 60;   // 两次写入的真实间隔（毫秒）
    static long JIT  = 60;   // 时钟偏移的随机抖动（毫秒）
    static long SEED = 42;

    static final int NP = 3;
    static final String[] PN = {"P1", "P2", "P3"};

    // ══════════════════════════════════════════════════════════════════════
    // Lamport 逻辑时钟
    //   保证：  a → b
…（截断，原长 16216 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：并发写 200 轮 ｜ R1 时钟偏移 +80ms（抖动 ±60ms）｜ 写入真实间隔 60ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 1-1 · Lamport 时钟与向量时钟
  和课件「实验 1」的经典示例完全一致 —— 浏览器里看到的数字，这里能跑出来
══════════════════════════════════════════════════════════════════════════════
    事件                类型        对端     Lamport        向量时钟
    ────────────────────────────────────────────────────────────────
    P1@t1           本地事件           —           1         [1,0,0]
    P2@t0           本地事件           —           1         [0,1,0]
    P3@t2           本地事件           —           1         [0,0,1]
    P1@t3           发送消息       P2@t5           2         [2,0,0]
    P2@t5           接收消息       P1@t3           3         [2,2,0]
    P3@t6           发送消息       P1@t8           2         [0,0,2]
    P1@t8           接收消息       P3@t6           3         [3,0,2]
    P2@t7           发送消息       P3@t9           4         [2,3,0]
    P3@t9           接收消息       P2@t7           5         [2,3,3]
    P1@t11          本地事件           —           4         [4,0,2]
    P3@t12          本地事件           —           6         [2,3,4]

▸ 注意 P1@t8 的向量是 [3,0,2] —— 第 2 位是 2，说明 P1 通过那条来自 P3 的消息
    P1@t1 ∥ P3@t6                L=1     L=2  Lamport 误以为 P1@t1 在 P3@t6 之前
    P2@t0 ∥ P1@t3                L=1     L=2  Lamport 误以为 P2@t0 在 P1@t3 之前
    P2@t0 ∥ P3@t6                L=1     L=2  Lamport 误以为 P2@t0 在 P3@t6 之前
    P2@t0 ∥ P1@t8                L=1     L=3  Lamport 误以为 P2@t0 在 P1@t8 之前
    P2@t0 ∥
…（截断，原长 1960 字符）
````

</details>

▸ **Bash** Fix Java indentation and view remaining output

````
python3 - <<'PY'
p='Lab01.java'; s=open(p,encoding='utf-8').read()
ends=['"得知"了 P3 已经走了 2 步。向量时钟的每一位就是"我所知道的它走到哪了"。""");',
      '它只是在"并发"的对上会编造出一个不存在的顺序。这就是"只能证伪，不能证实"。""");',
      '▸ 试试 -skew 0 和 -skew 300，看两个数字怎么变。""");',
      '代价只是每个时间戳多一个整数。CockroachDB / MongoDB 用的就是它。""");']
for e in ends:
    assert e in s, e[:40]
    s=s.replace(e, e[:-3]+'.indent(4));')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
java Lab01.java 2>&1 | grep -v "Picked up" | tail -32
````

<details><summary>↳ Bash 结果</summary>

````
ok
Lab01.java:333: error: illegal character: '\u3002'
      "?? 10:00 ??????"?????????HLC ???????
                               ^
Lab01.java:333: error: not a statement
      "?? 10:00 ??????"?????????HLC ???????
                       ^
Lab01.java:333: error: illegal character: '\uff0c'
      "?? 10:00 ??????"?????????HLC ???????
                                          ^
Lab01.java:334: error: illegal character: '\u3002'
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                    ^
Lab01.java:334: error: not a statement
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
      ^
Lab01.java:334: error: not a statement
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                                 ^
Lab01.java:334: error: ';' expected
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                                          ^
Lab01.java:334: error: illegal character: '\u3002'
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                                                ^
Lab01.java:334: error: not a statement
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                                           ^
Lab01.java:334: error: ';' expected
      ???????????????CockroachDB / MongoDB ??????"".indent(4));
                                                             ^
55 errors
error: compilation failed
````

</details>

▸ **Bash** Repair text block terminators and rerun

````
python3 - <<'PY'
p='Lab01.java'; s=open(p,encoding='utf-8').read()
n=s.count('"".indent(4));')
s=s.replace('"".indent(4));','""".indent(4));')
open(p,'w',encoding='utf-8').write(s); print('fixed',n)
PY
java Lab01.java 2>&1 | grep -v "Picked up" | sed -n '18,24p;52,64p;70,84p'
````

<details><summary>↳ Bash 结果</summary>

````
fixed 4
    P3@t9           接收消息       P2@t7           5         [2,3,3]
    P1@t11          本地事件           —           4         [4,0,2]
    P3@t12          本地事件           —           6         [2,3,4]

    ▸ 注意 P1@t8 的向量是 [3,0,2] —— 第 2 位是 2，说明 P1 通过那条来自 P3 的消息
      "得知"了 P3 已经走了 2 步。向量时钟的每一位就是"我所知道的它走到哪了"。

  两个副本各收到一次写，互不知情 —— 这是一次真正的并发写
══════════════════════════════════════════════════════════════════════════════
                                      last-write-wins    版本向量
    ──────────────────────────────────────────────────────────
    并发写轮数                                 200         200
    检测出冲突                                   0         200
    被丢弃的写入                               200           0
    其中丢的是真实更晚那条                     124           —

    LWW 的"明显错误率"             62.0%（丢掉真实更晚写入的比例）

    ▸ 最容易被忽略的一行是「被丢弃的写入」：LWW 每一轮都丢掉一条，
      因为它必须二选一。把 -skew 设成 0，"明显错误率"会降到 0，

══════════════════════════════════════════════════════════════════════════════
  Lab 1-4 · 混合逻辑时钟 HLC
  注入一次 NTP 回拨，看三种时钟谁活了下来
══════════════════════════════════════════════════════════════════════════════
    事件            物理时钟    Lamport    HLC (l, c)    l − 物理时钟
    ─────────────────────────────────────────────────────────────────────
    #1                 12056          1    (12056, 0)           +0 ms
    #2                 12127          2    (12127, 0)           +0 ms
    #3                 12214          3    (12214, 0)           +0 ms
    #4 *回拨*          11414
…（截断，原长 1822 字符）
````

</details>

▸ **Bash** Update README and assemble Part 1 tree

````
cd 〔路径00054〕 && cp ch01/ch01.html courseware/ch01-time-order.html && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
s=s.replace('''# 1. 打开第 0 章交互课件（浏览器里直接玩 6 个模拟实验）
open courseware/ch00-intro.html          # macOS
# xdg-open courseware/ch00-intro.html    # Linux

# 2. 跑 Lab 0（Go 版，需要 Go 1.21+）
cd go/lab00 && go run .

# 3. 跑 Lab 0（Java 版，需要 JDK 17+，无需 Maven/Gradle）
cd java/lab00 && java Lab00.java''',
'''# 1. 打开交互课件（浏览器里直接玩模拟实验）
open courseware/ch00-intro.html          # 第 0 章：序言与导论
open courseware/ch01-time-order.html     # Part 1：时间、顺序与因果
# Linux 用 xdg-open

# 2. 跑实验（Go 版，需要 Go 1.21+）
cd go/lab00 && go run .      # Lab 0：不可靠信道与第三态
cd go/lab01 && go run .      # Lab 1：Lamport / 向量时钟 / HLC

# 3. 跑实验（Java 版，需要 JDK 17+，无需 Maven/Gradle）
cd java/lab00 && java Lab00.ja
…（截断，原长 3788 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README updated
〔路径00056〕
〔路径00071〕
〔路径00072〕
〔路径00081〕
〔路径00083〕
〔路径00086〕
〔路径00087〕
〔路径00088〕
〔路径00089〕
〔路径00092〕
〔路径00093〕
〔路径00094〕
〔路径00125〕
〔路径00127〕
````

</details>

▸ **Bash** Re-verify final Part 1 courseware and recapture

````
cd /tmp && sed -i 's|〔路径00061〕' c1.mjs s1.mjs && node c1.mjs 2>&1 | tail -6 && node s1.mjs && echo shots-done
````

<details><summary>↳ Bash 结果</summary>

````
circles after add+msg: 13
LWW: 策略一 · LAST-WRITE-WINS | A 的时间戳 | 260　= 真实 140 + 偏移 120 | B 的时间戳 | 200　= 真实 200 + 偏移 0 | 最终保留 | 值 A | 后写的 B 被静默丢弃了。C2 明明在 C1 之后 60ms 才写，但 R1 的时钟快了 120ms，于是 A 的时间戳更大。 | 没有报错，没有日志，用户看到"保存成功"然后数据消失。
LWW skew=0: 策略一 · LAST-WRITE-WINS | A 的时间戳 | 140　= 真实 140 + 偏移 0 | B 的时间戳 | 200　= 真实 200 + 偏移 0 | 最终保留 | 值 B | 这一次侥幸正确（时钟偏移 0ms 小于写入间隔 60ms）。但正确与否取决于一个和业务毫无关系的参数——把偏移拖大一点就翻转了。
HLC rows: #1 12055 1 (12055, 0) +0 ms || #2 12130 2 (12130, 0) +0 ms || #3 12196 3 (12196, 0) +0 ms || #4 回拨后 11696 ⟵ 倒退！ 4 (12196, 1) +500 ms || #5 11751 5 (12196, 2) +445 ms || #6 11776 6 (12196, 3) +420 ms || #7 11849 7 (12196, 4) +347 ms
HLC verdict: ⚠ 纯物理时钟出现了 1 次倒退。而 Lamport 与 HLC 两列始终单调递增——注意 HLC 的 l 在回拨期间冻结不动、靠 c 递增撑住顺序，等物理时钟追上来后 c 自动归零。
ERRORS: none
shots-done
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00215〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Fix LWW diagram viewBox clipping

````
cd 〔路径00054〕 && sed -i 's|<svg id="sv2" viewBox="0 0 700 190"|<svg id="sv2" viewBox="0 0 700 208"|' courseware/ch01-time-order.html ch01/ch01.html && grep -o 'id="sv2" viewBox="[^"]*"' courseware/ch01-time-order.html
````

<details><summary>↳ Bash 结果</summary>

````
id="sv2" viewBox="0 0 700 208"
````

</details>

📎 **发送文件**：Part 1 交互式课件。核心是「实验 1 · 时空图编辑器」——可以自己摆事件、拉消息箭头，实时算出 Lamport 与向量时钟，点两个事件就告诉你它们到底有没有因果关系，以及 Lamport 在哪里会骗你。
- `〔路径00072〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00072〕 → file_uuid: 〔标识00994〕
````

</details>

📎 **发送文件**：Lab 1 代码（Go + Java）。`clocks.go` 的 Compare() 是向量时钟的全部精髓，五行；trace.go 里另外用 happens-before 定义做了一次独立可达性校验，判定不一致会直接 panic。
- `〔路径00056〕`
- `〔路径00083〕`
- `〔路径00088〕`
- `〔路径00092〕`
- `〔路径00094〕`
- `〔路径00093〕`
- `〔路径00127〕`

<details><summary>↳ SendUserFile 结果</summary>

````
7 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识00434〕
  〔路径00083〕 → file_uuid: 〔标识01995〕
  〔路径00088〕 → file_uuid: 〔标识01459〕
  〔路径00092〕 → file_uuid: 〔标识01430〕
  〔路径00094〕 → file_uuid: 〔标识00428〕
  〔路径00093〕 → file_uuid: 〔标识00439〕
  〔路径00127〕 → file_uuid: 〔标识00638〕
````

</details>

▸ **device:device_bash**

````
{"command": "D=\"$HOME〔路径00144〕\"; mkdir -p \"$D/go/internal/tui\" \"$D/go/lab01\" \"$D/java/lab01\" && [ -f \"$D/go/lab00/width.go\" ] && mkdir -p \"$D/_to_delete\" && mv \"$D/go/lab00/width.go\" \"$D/_to_delete/width.go\" && echo \"width.go 已移至 _to_delete（已被 internal/tui 取代）\"; ls -R \"$D\" | head -40"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
width.go 已移至 _to_delete（已被 internal/tui 取代）
/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00166〕
README.md
_to_delete
courseware
go
java

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00145〕
width.go

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00147〕
ch00-intro.html

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00157〕
go.mod
internal
lab00
lab01

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00149〕
tui

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00148〕

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00151〕
bank.go
client.go
main.go
net.go

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00152〕

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00165〕
lab00
lab01

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00159〕
Lab00.java

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00160〕
````

</details>

💾 **写入你的 Mac**：
- `〔路径00006〕`
- `〔路径00004〕`
- `〔路径00014〕`
- `〔路径00017〕`
- `〔路径00020〕`
- `〔路径00022〕`
- `〔路径00021〕`
- `〔路径00045〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00006〕","〔路径00004〕","〔路径00014〕","〔路径00017〕","〔路径00020〕","〔路径00022〕","〔路径00021〕","〔路径00045〕"],"rejected":[]}
````

</details>

🧩 **创建产出物** `ds-course-ch01`：分布式系统课程 · Part 1：时间、顺序与因果。含 3 个交互实验（可编辑的时空图编辑器 + 实时 Lamport/向量时钟判定、LWW 静默丢写复现、HLC 时钟回拨对照）、happens-before 偏序、向量时钟比较规则、TrueTime 与 commit wait，以及 10 道自测题。

▸ **TaskUpdate** completed

▸ **TaskUpdate** completed

▸ **TaskUpdate** completed

<details><summary>↳ mcp__remote-devices__create_artifact 结果</summary>

````
Artifact "ds-course-ch01" created on the connected desktop.
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #6 status
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #8 status
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #7 status
````

</details>

Part 1 交付完成，课件和 Lab 1 都已落到 `~/Code/distributed-systems-course/`，artifact 是 `ds-course-ch01`。

**这一章的设计有个刻意安排**：课件里「实验 1 · 时空图编辑器」的经典示例，和 `go run .` 跑出来的 Lab 1-1 是<u>同一组数据</u>。你在浏览器里点出来的每一个向量时钟，命令行里都能复现。这样你能确认自己理解的和代码算的是一回事，而不是两套并行的说法。

**三件我认为最值得你花时间的事：**

1. **时空图编辑器的"比较"模式**。切到③，点 `P3@t2` 和 `P1@t3` —— Lamport 值是 1 和 2，看起来有先后，但向量时钟判定并发。程序会用红字把这个盲区标出来。这一个交互抵得上读三页论文。

2. **Lab 1-2 有一个自我校验**。它不是拿向量时钟自己验证自己：程序另外用 happens-before 的三条规则做了一遍图可达性搜索，两种方法的结论必须完全一致，否则直接 panic。跑出来的 55 对事件里 22 对并发，其中 **16 对被 Lamport 误导**（73%）。这个比例比大多数人的直觉高得多。

3. **Lab 1-3 有个反直觉结论**，比课件正文更狠：LWW **每一轮都丢掉一条写入**，因为面对两条并发写它必须二选一。时钟准不准，改变的只是"丢哪一条"——把 `-skew` 设成 0，"明显错误率"降到 0%，但丢弃率仍然是 100%。跑一下 `go run . -skew 0` 亲眼确认。

顺带说一句 Go 代码的变化：我把终端表格对齐工具抽成了 `go/internal/tui`，lab00 也改成引用它了。你本地那份 `go/lab00/width.go` 被我移到了 `_to_delete/`（device_bash 不允许删文件），确认没问题后自己删掉就行。

---

三道题，还是先别翻课件：

1. 你在生产里判断"两条数据谁新"，用的是什么？如果是 `updated_at` 时间戳，那这个时间戳是谁打的——应用服务器、数据库、还是客户端？
2. Lab 1-2 跑出 73% 的误导率。要让这个数字降到 0，需要什么条件？（提示：想想什么情况下"并发对"会消失）
3. **这题最想听你的答案**：你们 `〔内部库C〕` / `〔内部服务F〕` 这类系统里，有没有哪张表是多个服务/多个实例都会写的？如果有，冲突现在是怎么处理的——数据库唯一约束、乐观锁版本号、还是"最后写的赢"？

第 3 题不用答得完整，说个大概就行。Part 2 讲一致性谱系时，如果能挂在你真实的表结构上，比用抽象的 KV 例子有效得多。

> ✓ 本轮结束 · 2026-08-20 12:58 · 50 步
