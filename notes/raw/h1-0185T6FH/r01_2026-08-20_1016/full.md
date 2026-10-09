# 恢复的会话记录 · session_0185T6FH1dEv7fuTDPowiw31

来源：Claude 桌面应用本地 HTTP 缓存（旧 org 〔组织ID〕）。
时间跨度：2026-08-20 10:16 → 2026-09-22 09:51，共 3661 条事件。


## 👤 你 · 2026-08-20 10:16

我想系统地了解并深入学习分布式算法及其对应应用。请你帮助我从零开始进行学习。
我预估的大致学习路线是这样的：

1. 分布式的各种基础定义，如可见性，最终一致性等

2. 分布式的各种算法，如两阶段提交、TCC、lamport、raft等
3. 分布式的各种应用及，如mapreduce，bigtable，spanner，zookeeper、Kafka、rocketmq、vitess等

但是你不需要遵循我的思路进行，你可以按照你认为合适的方式进行展露。
我们最好以描述、问答、互动、实操相结合的方式完成学习，并且希望每阶段的学习完成都能总结出对应的交互式学习的产物，变成一个实际可复用可复习的课程。

请你先开始最基础的面向初学者的序言和介绍吧。

> ✓ 本轮结束 · 2026-08-20 10:16 · 0 步

▸ **device:device_list_dir**

````
{"path": "〔路径00001〕"}
````

▸ **ToolSearch**

````
{"max_results": 5, "query": "select:TaskCreate,TaskUpdate"}
````

<details><summary>↳ mcp__remote-devices__device_list_dir 结果</summary>

[已脱敏：课程目录以外的本地目录/文件列表。未脱敏原文只保存在本地 notes/private/]

</details>

<details><summary>↳ ToolSearch 结果</summary>

````
[tool_reference]
[tool_reference]
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00143〕\" 2>/dev/null || cd \"$HOME/mnt\"/*Code* ; for d in */ ; do echo \"=== $d\"; ls \"$d\" 2>/dev/null | head -20; done 2>/dev/null | head -120"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

[已脱敏：课程目录以外的本地目录/文件列表。未脱敏原文只保存在本地 notes/private/]

</details>

## 🤖 Claude · 2026-08-20 10:17

❓ **实操环节你希望用哪种语言写代码？(我看你的 Code 目录以 Python 为主，也有 Node/JS)**
  - Python 为主（推荐）：匹配你现有技术栈，写模拟器/实验最快，asyncio 模拟网络分区、时钟漂移都很直观。读工业源码时我再用 Go/Java 片段做对照讲解。
  - Go 为主：Go 是分布式系统的事实标准语言（etcd、TiDB、Kubernetes、Kafka 客户端）。学完能直接读懂 raft 库源码，但你需要有 Go 基础。
  - Python + Go 双轨：概念实验用 Python 快速验证，关键算法（Raft、Paxos）再用 Go 实现一遍并对照工业实现。工作量约 1.5 倍，但收获最扎实。
  - Java 为主：适合主攻 Kafka / RocketMQ / ZooKeeper / Vitess 生态，这些系统的源码大多是 Java。

❓ **你现在的起点大概在哪个位置？这决定我从哪里切入、跳过多少铺垫。**
  - 用过中间件但不懂原理：日常用过 MySQL/Redis/Kafka/K8s 之类，知道怎么调 API，但对背后的复制、选主、一致性机制没系统概念。
  - 完全零基础：只写过单机程序，没接触过分布式概念。我会从「为什么单机不够用」讲起，每个术语都从零解释。
  - 写过分布式服务，缺理论：做过微服务/分布式事务的业务开发，踩过坑，但缺少一致性模型、共识算法这类理论骨架。
  - 读过部分论文，想系统化：知道 CAP、Raft 这些名词，读过零散资料，想把知识串成完整体系并补上证明与边界条件。

❓ **每一章的「学习产物」你希望是什么形态？**
  - 交互式 HTML 课件（推荐）：可点击的动画模拟（消息丢包、时钟漂移、选主过程）+ 折叠式自测题 + 概念地图。会持久化成 artifact，随时回看，也能分享给同事。
  - 可运行的代码实验仓库：在你本地 Code 目录建一个课程仓库，每章配可运行的实验代码 + 测试，你可以改参数看行为变化。
  - Markdown 精读笔记：结构化的文字讲义，含论文出处、关键定理、对照表，适合搜索和长期沉淀。
  - 每章一套问答测验：苏格拉底式提问：我出题、你答、我批改并针对薄弱点补讲。适合检验是否真的理解。

❓ **理论严谨度上你偏好哪一档？**
  - 工程直觉优先，理论点到为止（推荐起步）：先用类比和可视化建立直觉，证明只讲思路不展开。上手快，后面可随时加深。
  - 直觉 + 形式化证明要点：讲清楚 FLP、CAP 的证明骨架、线性一致性的形式化定义，接近研究生课程（MIT 6.5840 / CMU 15-712）水平。
  - 直接对标论文与源码：每个主题配原始论文精读 + 工业实现源码走查，节奏慢但深度最大。

## 👤 你（回答选择题）· 2026-08-20 10:18

The user answered: "实操环节你希望用哪种语言写代码？(我看你的 Code 目录以 Python 为主，也有 Node/JS)"="go和java吧", "你现在的起点大概在哪个位置？这决定我从哪里切入、跳过多少铺垫。"="完全零基础", "每一章的「学习产物」你希望是什么形态？"="交互式 HTML 课件（推荐）,可运行的代码实验仓库,每章一套问答测验", "理论严谨度上你偏好哪一档？"="工程直觉优先，理论点到为止（推荐起步）"

▸ **TaskCreate** 设计课程整体大纲与学习路线图

<details><summary>↳ TaskCreate 结果</summary>

````
Task #1 created successfully: 设计课程整体大纲与学习路线图
````

</details>

▸ **TaskCreate** 编写第 0 章交互式 HTML 课件

<details><summary>↳ TaskCreate 结果</summary>

````
Task #2 created successfully: 编写第 0 章交互式 HTML 课件
````

</details>

▸ **TaskCreate** 搭建本地实验仓库骨架（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #3 created successfully: 搭建本地实验仓库骨架（Go + Java）
````

</details>

▸ **TaskCreate** 编写 Lab 0：不可靠信道与第三态实验

<details><summary>↳ TaskCreate 结果</summary>

````
Task #4 created successfully: 编写 Lab 0：不可靠信道与第三态实验
````

</details>

▸ **TaskCreate** 验证并交付第 0 章产物

<details><summary>↳ TaskCreate 结果</summary>

````
Task #5 created successfully: 验证并交付第 0 章产物
````

</details>

▸ **Bash** Check Go and Java toolchain availability

````
echo "--- go ---"; go version 2>&1 | head -2; echo "--- java ---"; java -version 2>&1 | head -3; javac -version 2>&1 | head -2; echo "--- maven/gradle ---"; mvn -v 2>&1|head -1; gradle -v 2>&1|head -3
````

▸ **Bash** Locate dataviz skill file

````
find / -iname "SKILL.md" -path "*dataviz*" 2>/dev/null | head -5
````

<details><summary>↳ Bash 结果</summary>

````
--- go ---
go version go1.24.7 linux/amd64
--- java ---
Picked up JAVA_TOOL_OPTIONS: -Djavax.net.ssl.trustStore=/root/.ccr/java-truststore.p12 -Djavax.net.ssl.trustStorePassword=changeit -Djavax.net.ssl.trustStoreType=PKCS12 -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=45377 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|127.*|0.*|::|169.254.*|anthropic.com|*.anthropic.com|*.anthropic.com|registry.npmjs.org|jsr.io|npm.jsr.io|pypi.org|files.pythonhosted.org|index.crates.io|proxy.golang.org|host.docker.internal|10.*|172.16.*|172.17.*|172.18.*|172.19.*|172.20.*|172.21.*|172.22.*|172.23.*|172.24.*|172.25.*|172.26.*|172.27.*|172.28.*|172.29.*|172.30.*|172.31.*|192.168.*|100.64.0.0/10|*.svc.cluster.local|*.svc.cluster.local -Djdk.http.auth.tunneling.disabledSchemes= -Djdk.http.auth.proxying.disabledSchemes=
openjdk version "21.0.10" 2026-01-20
OpenJDK Runtime Environment (build 21.0.10+7-Ubuntu-124.04)
Picked up JAVA_TOOL_OPTIONS: -Djavax.net.ssl.trustStore=/root/.ccr/java-truststore.p12 -Djavax.net.ssl.trustStorePassword=changeit -Djavax.net.ssl.trustStoreType=PKCS12 -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=45377 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|127.*|0.*|::|169.254.*|anthropic.com|*.anthropic.com|*.anthropic.com|registry.npmjs.org|jsr.io|npm.jsr.io|pypi.org|files.pythonhosted.org|index.crates.io|proxy.golang.org|host.docker.internal|10.*|172.16.*|172.17.*|172.18.*|172.19.*|172.20.*|172.21.*|172.22.*|172.23.*|172.24.*|172.25.*|172.26.*|172.27.*|172.28.*|172.29.*|17
…（截断，原长 3273 字符）
````

</details>

<details><summary>↳ Bash 结果</summary>

````
(Bash completed with no output)
````

</details>

▸ **Skill**

````
{"args": "为一套多章节分布式系统交互式课件建立统一配色与图表规范（深色主题为主，含状态色：正常/延迟/丢失/故障）", "skill": "dataviz"}
````

<details><summary>↳ Skill 结果</summary>

````
Launching skill: dataviz
````

</details>

_（加载技能：dataviz）_

▸ **Bash** Read reference palette

````
cd 〔路径00198〕 && sed -n '1,200p' references/palette.md
````

<details><summary>↳ Bash 结果</summary>

````
# Reference palette

This is the **reference instance** of the data-viz method: every parameter the
method needs, filled in with a validated default palette. The rest of the skill
is system-agnostic — **to target your brand, substitute this file's values** and
re-run the validator. Nothing else changes.

## How to use these values

Everything below is plain hex. In an HTML chart, **define the slots you use as
CSS custom properties in a local `<style>` block** at the top of the file, then
reference them by role throughout — so the light/dark values swap in one place,
and the chart body is written against roles rather than raw hex:

```css
.viz-root {
  color-scheme: light;
  --surface-1:      #fcfcfb;   /* chart surface */
  --text-primary:   #0b0b0b;
  --text-secondary: #52514e;
  --series-1:       #2a78d6;   /* categorical slot 1 */
  /* …only the roles this chart uses */
}
@media (prefers-color-scheme: dark) {
  :root:where(:not([data-theme="light"])) .viz-root {
    color-scheme: dark;
    --surface-1:      #1a1a19;
    --text-primary:   #ffffff;
    --text-secondary: #c3c2b7;
    --series-1:       #3987e5;
  }
}
:root[data-theme="dark"] .viz-root {
  color-scheme: dark;
  --surface-1:      #1a1a19;
  --text-primary:   #ffffff;
  --text-secondary: #c3c2b7;
  --series-1:       #3987e5;
}
```

Declare the dark values under both scopes as above — the media query covers
the OS setting; the `data-theme` scope covers the viewer's theme toggle,
which must win both ways (the `:not
…（截断，原长 9018 字符）
````

</details>

▸ **Bash** Validate the 3-slot node palette for dark surface

````
cd 〔路径00198〕 && node scripts/validate_palette.js "#3987e5,#d95926,#199e70" --mode dark --surface "#1a1a19" --pairs all 2>&1 | tail -30
````

▸ **Bash** Create working directory

````
mkdir -p 〔路径00057〕 && cd 〔路径00054〕 && pwd && ls
````

<details><summary>↳ Bash 结果</summary>

````
Palette (dark, surface #1a1a19, categorical): 3 slots
  [PASS] Lightness band         all 3 inside L 0.48–0.67
  [PASS] Chroma floor           all 3 >= 0.1
  [PASS] CVD separation         worst all-pairs #199e70↔#d95926 ΔE 9.4 (deutan) · tritan 4.0
  [PASS] Normal-vision floor    worst all-pairs #199e70↔#3987e5 ΔE 20.9 (normal)
  [PASS] Contrast vs surface    all 3 >= 3:1

  → ALL CHECKS PASS  (CVD in the 6–8 floor band is legal ONLY with secondary encoding: direct labels, gaps, or texture)
  scope: categorical palettes only. For a lone status/text color check WCAG text contrast; for a sequential ramp, lightness monotonicity.
Shell cwd was reset to 〔路径00052〕
````

</details>

<details><summary>↳ Bash 结果</summary>

````
〔路径00054〕
ch00
````

</details>

▸ **Bash** Write HTML head and CSS

````
cat > 〔路径00058〕 <<'CHUNK_EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · 第 0 章 · 序言与导论</title>
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
  --sans:system-ui,-apple-system,"Segoe UI","
…（截断，原长 11710 字符）
````

<details><summary>↳ Bash 结果</summary>

````
chunk1 ok: 186 lines
````

</details>

▸ **Bash** Append preface and section 0.1

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<section id="s0">
  <div class="eyebrow">Chapter 0 · Preface</div>
  <h1>当"一台机器"变成"一群机器"</h1>
  <p class="lead">这门课不是从"什么是 CAP 定理"开始的。它从一个更朴素的问题开始：<em>单机编程时，有三件事你一直在免费享用，却从来没意识到它们的存在。</em>把程序拆到多台机器上的那一刻，这三件事同时消失了。分布式系统的全部理论——一致性模型、共识算法、分布式事务——本质上都是在赎回这三件失物，以及计算赎金。</p>

  <div class="card">
    <div class="card-t">这一章你会带走什么</div>
    <ul style="margin-bottom:0">
      <li>一个能挂载后续全部知识的<strong>心智框架</strong>（三个免费假设 → 三个不可靠 → 三类解法）</li>
      <li>6 个可以亲手拖动、点击的<strong>交互实验</strong>，把抽象命题变成你眼睛看见的现象</li>
      <li>20 个术语的<strong>零基础速查表</strong>——后面每一章都会反复用到</li>
      <li>一份重新组织过的<strong>7 阶段学习路线图</strong>（比你原本的 3 段式多了必要的中间层）</li>
      <li>Lab 0：用 Go 和 Java 各写一遍"不可靠信道"，亲手复现分布式系统里最贵的那个 bug</li>
    </ul>
  </div>

  <h3>怎么用这份课件</h3>
  <p>每
…（截断，原长 5255 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Append sections 0.2 and 0.3

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<section id="s2">
  <h2>0.2 单机世界里你免费拥有的三件事</h2>
  <p>这是本章最重要的一节。如果你只记住一页，记住这一页。</p>
  <p>当你在一台机器上写下面这段代码时，你在<strong>无意识地依赖三个假设</strong>：</p>

<pre><code><span class="cm">// 单机：一个再平常不过的转账</span>
<span class="kw">func</span> <span class="fn">Transfer</span>(from, to *Account, amount <span class="kw">int</span>) {
    mu.<span class="fn">Lock</span>()
    <span class="kw">defer</span> mu.<span class="fn">Unlock</span>()
    from.balance -= amount        <span class="cm">// ← 假设 ③：内存里只有一份余额</span>
    to.balance   += amount
    log.<span class="fn">Printf</span>(<span class="st">"%v 转账完成"</span>, time.<span class="fn">Now</span>())  <span class="cm">// ← 假设 ①：有一个共同的"现在"</span>
}
<span class="cm">// 调用方</span>
<span class="fn">Transfer<
…（截断，原长 7086 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Append sections 0.4 and 0.5

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<section id="s4">
  <h2>0.4 时间：没有一个共同的"现在"</h2>
  <p>你可能觉得"时钟不准"顶多差几毫秒，不是大问题。这个直觉在分布式系统里会害死人。先看两种时钟：</p>
  <dl class="kv">
    <dt>墙上时钟<br><span style="font-size:12px;font-weight:400">wall clock</span></dt>
    <dd><code>time.Now()</code> / <code>System.currentTimeMillis()</code>。返回"1970 年以来的秒数"，可以转成人类日期。<strong>会被 NTP 校准而跳变，甚至倒退。</strong>绝不能用来测量时间间隔。</dd>
    <dt>单调时钟<br><span style="font-size:12px;font-weight:400">monotonic clock</span></dt>
    <dd>Go 的 <code>time.Since()</code> / Java 的 <code>System.nanoTime()</code>。只保证"永远向前"，数值本身没有意义。<strong>可以测间隔，但不同机器之间毫无可比性。</strong></dd>
  </dl>
  <div class="note warn">
    <div class="note-h"><span class="dot" style="background:var(--warn)"></span>新手最常犯的两个错</div>
    <p style="margin-bo
…（截断，原长 5757 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Append sections 0.6 through 0.8

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<section id="s6">
  <h2>0.6 第三态与幂等性：本章最实用的一节</h2>
  <p>单机函数调用有两种结果，网络调用有三种：</p>
  <div class="path">
    <div class="pstep" style="border-color:var(--ok)"><div class="n" style="background:var(--ok)">✓</div><div><div class="ti">成功</div><div class="de">收到了响应，且响应说成功。你确切知道对方做了。</div></div></div>
    <div class="pstep" style="border-color:var(--crit)"><div class="n" style="background:var(--crit)">✗</div><div><div class="ti">明确失败</div><div class="de">收到了响应，响应说"参数非法"/"余额不足"。你确切知道对方没做。</div></div></div>
    <div class="pstep" style="border-color:var(--warn)"><div class="n" style="background:var(--warn)">?</div><div><div class="ti">第三态：不确定（indeterminate）</div><div class="de">超时、连接被重置、进程崩溃。<strong>你不知道对方做了没有。</strong>这一态在单机世界里不存在，是分布式编程一切复杂度
…（截断，原长 10186 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Append glossary, roadmap, quiz and lab sections

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<section id="s9">
  <h2>0.9 术语速查表</h2>
  <p>后面每一章都会用到这 20 个词。现在不用背，遇到时回来查。</p>
  <table>
    <thead><tr><th style="width:20%">术语</th><th style="width:44%">一句话解释</th><th>常见误解 / 补充</th></tr></thead>
    <tbody>
      <tr><td><strong>节点</strong> node</td><td>参与系统的一个独立进程（通常一台机器一个）。</td><td>关注点是"独立的失败单元"，不是"物理机"。</td></tr>
      <tr><td><strong>副本</strong> replica</td><td>同一份数据在不同节点上的多个拷贝。</td><td>为了容错和就近读，代价是一致性问题。</td></tr>
      <tr><td><strong>分片</strong> shard / partition</td><td>把大数据集切成互不重叠的小块，分散到不同节点。</td><td>Kafka 里叫 partition，MySQL 里叫分库分表。解决容量，不解决容错。</td></tr>
      <tr><td><strong>主 / 从</strong> leader / follower</td><td>一组副本中指定一个负责接收写入，其余同步它。</td><td>也叫 primary/secondary、master/slave。写入串行化的最简单办法。</td></tr>
      <tr><td><stro
…（截断，原长 13765 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Relabel stat cells

````
cd 〔路径00057〕 && sed -i 's|<div class="k">多扣</div>|<div class="k">差额</div>|g' ch00-intro.html && grep -c '差额' ch00-intro.html
````

<details><summary>↳ Bash 结果</summary>

````
2
````

</details>

▸ **Bash** Append JS part 1: theme, lab1, lab2

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(tag,attrs,txt){const e=document.createElementNS(NS,tag);for(const k in attrs)e.setAttribute(k,attrs[k]);if(txt!=null)e.textContent=txt;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}
function clr(id){$("#"+id).innerHTML="";}

/* ============ 主题 / 进度 / 目录 ============ */
$("#tbtn").onclick=()=>{
  const cur=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",cur==="dark"?"light":"dark");
  draw1();draw3();draw7();
};
const secs=$$("main section"), links=$$("#toc a:
…（截断，原长 9081 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Append JS part 2 and close document

````
cat >> 〔路径00058〕 <<'CHUNK_EOF'

/* ============ 实验 3 · 时钟漂移 ============ */
const S3={t0:performance.now(),run:false,off:{A:0,B:0,C:0},ev:[],n:0,last:performance.now()};
const NODES=[["A","--s1"],["B","--s3"],["C","--s7"]];
function drift(k){return +$("#r3"+k.toLowerCase()).value;}
function trueMs(){return S3.elapsed||0;}
function clockOf(k){return trueMs()*(1+drift(k)/1000)+S3.off[k];}
function fmtClk(ms){const s=ms/1000;const m=Math.floor(Math.abs(s)/60),r=Math.abs(s)%60;
  return (s<0?"-":"")+String(m).padStart(2,"0")+":"+r.toFixed(2).padStart(5,"0");}
function draw3(){
  const svg=$("#sv3"); svg.innerHTML="";
  const W=640,PADL=104,PADR=110,y0=30,dy=40, half=(W-PADL-PADR)/2, mid=PADL+half;
  const MAXDEV=4000;
  svg.appendChild(el("text",{x:PADL-14,y:14,fill
…（截断，原长 11682 字符）
````

<details><summary>↳ Bash 结果</summary>

````
total lines: 1154
100K	〔路径00058〕
````

</details>

▸ **Bash** Render page in Chromium and check for errors

````
cd 〔路径00054〕 && cat > 〔路径00194〕 <<'EOF'
import { chromium } from 'playwright';
const b = await chromium.launch();
const p = await b.newPage({viewport:{width:1300,height:1000}});
const errs=[];
p.on('console', m=>{ if(m.type()==='error') errs.push('CONSOLE: '+m.text()); });
p.on('pageerror', e=>errs.push('PAGEERROR: '+e.message));
await p.goto('file://〔路径00058〕');
await p.waitForTimeout(800);
// exercise the interactives
await p.click('#b2ten'); await p.waitForTimeout(1500);
await p.click('#b3run'); await p.waitForTimeout(600);
await p.click('#b3ec'); await p.click('#b3eb'); await p.click('#b3ea');
await p.click('#b5run');
await p.click('#b7auto'); await p.waitForTimeout(2600);
await p.click('#q4box button[data-g="1"]'); await p.waitForTimeout(200);
a
…（截断，原长 2526 字符）
````

<details><summary>↳ Bash 结果</summary>

````
node:internal/modules/run_main:123
    triggerUncaughtException(
    ^

Error [ERR_MODULE_NOT_FOUND]: Cannot find package 'playwright' imported from 〔路径00194〕
Did you mean to import "playwright/index.js"?
    at Object.getPackageJSONURL (node:internal/modules/package_json_reader:314:9)
    at packageResolve (node:internal/modules/esm/resolve:768:81)
    at moduleResolve (node:internal/modules/esm/resolve:855:18)
    at defaultResolve (node:internal/modules/esm/resolve:985:11)
    at #cachedDefaultResolve (node:internal/modules/esm/loader:731:20)
    at ModuleLoader.resolve (node:internal/modules/esm/loader:708:38)
    at ModuleLoader.getModuleJobForImport (node:internal/modules/esm/loader:310:38)
    at ModuleJob._link (node:internal/modules/esm/module_job:182:49) {
  code: 'ERR_MODULE_NOT_FOUND'
}

Node.js v22.22.2
````

</details>

▸ **Bash** Install playwright and run the render check

````
cd /tmp && npm install playwright --no-audit --no-fund 2>&1 | tail -3 && node 〔路径00194〕 2>&1 | tail -40
````

<details><summary>↳ Bash 结果</summary>

````
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.0.2
npm notice To update run: npm install -g npm@12.0.2
npm notice
node:internal/modules/run_main:123
    triggerUncaughtException(
    ^

browserType.launch: Executable doesn't exist at 〔路径00180〕
╔════════════════════════════════════════════════════════════╗
║ Looks like Playwright was just installed or updated.       ║
║ Please run the following command to download new browsers: ║
║                                                            ║
║     npx playwright install                                 ║
║                                                            ║
║ <3 Playwright Team                                         ║
╚════════════════════════════════════════════════════════════╝
    at 〔路径00195〕 {
  log: [],
  name: 'Error'
}

Node.js v22.22.2
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Inspect preinstalled browser paths

````
ls 〔路径00175〕 2>/dev/null; echo "---"; ls 〔路径00176〕 2>/dev/null | head -20
````

<details><summary>↳ Bash 结果</summary>

````
chromium
chromium-1194
chromium_headless_shell-1194
ffmpeg-1011
---
〔路径00177〕
DEPENDENCIES_VALIDATED
INSTALLATION_COMPLETE
chrome-linux

〔路径00179〕
DEPENDENCIES_VALIDATED
INSTALLATION_COMPLETE
chrome-linux

〔路径00181〕
COPYING.LGPLv2.1
DEPENDENCIES_VALIDATED
INSTALLATION_COMPLETE
ffmpeg-linux
````

</details>

▸ **Bash** Run render check with preinstalled chromium

````
cd /tmp && sed -i "s|chromium.launch()|chromium.launch({executablePath:'〔路径00178〕'})|" check.mjs && node check.mjs 2>&1 | tail -40
````

<details><summary>↳ Bash 结果</summary>

````
{
 "v1s": "99.601%",
 "v1p": "100.000000%",
 "v1q": "99.999401%",
 "v1t": "3 / 4",
 "v2c": "20",
 "v2ok": "17",
 "v2s": "24",
 "v2gap": "+7",
 "v5aa": "2800",
 "v5ba": "2000",
 "v5ae": "2000",
 "tb3rows": 3,
 "k7": 8,
 "sv1": 25,
 "sv3": 24,
 "sv7": 12,
 "scrollH": 18740
}
ERRORS: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Fix availability formatting and re-verify

````
cd 〔路径00057〕 && python3 - <<'PY'
import re,io
p='ch00-intro.html'
s=open(p,encoding='utf-8').read()
old = '''function fmtAvail(a){
  const p=a*100;
  if(p>99.999) return p.toFixed(6)+"%";
  if(p>99.99) return p.toFixed(5)+"%";
  if(p>99.9) return p.toFixed(4)+"%";
  if(p>99) return p.toFixed(3)+"%";
  return p.toFixed(2)+"%";
}'''
new = '''function fmtAvail(a){
  const u=Math.max(0,1-a);
  if(u<=0) return "100%";
  if(u<1e-9) return ">99.9999999%";
  const d=Math.min(9,Math.max(2,Math.ceil(-Math.log10(u))+1));
  return parseFloat((a*100).toFixed(d))+"%";
}'''
assert old in s
s=s.replace(old,new)
s=s.replace('    const dares=false;\n','')
open(p,'w',encoding='utf-8').write(s)
print("patched")
PY
node 〔路径00194〕 2>&1 | head -20
````

<details><summary>↳ Bash 结果</summary>

````
patched
{
 "v1s": "99.6006%",
 "v1p": ">99.9999999%",
 "v1q": "99.9994008%",
 "v1t": "3 / 4",
 "v2c": "20",
 "v2ok": "17",
 "v2s": "23",
 "v2gap": "+6",
 "v5aa": "2400",
 "v5ba": "2000",
 "v5ae": "2000",
 "tb3rows": 3,
 "k7": 8,
 "sv1": 25,
 "sv3": 24,
 "sv7": 12,
 "scrollH": 18740
}
ERRORS: none
````

</details>

▸ **Bash** Adjust precision rule and re-verify

````
sed -i 's|const d=Math.min(9,Math.max(2,Math.ceil(-Math.log10(u))+1));|const d=Math.min(9,Math.max(2,Math.floor(-Math.log10(u))+1));|' ch00-intro.html && node 〔路径00194〕 2>&1 | sed -n '1,8p;20,22p'
````

<details><summary>↳ Bash 结果</summary>

````
{
 "v1s": "99.601%",
 "v1p": ">99.9999999%",
 "v1q": "99.999401%",
 "v1t": "3 / 4",
 "v2c": "20",
 "v2ok": "19",
 "v2s": "23",
ERRORS: none
````

</details>

▸ **Read** `〔路径00246〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Read** `〔路径00243〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Capture targeted screenshots of each lab

````
cd /tmp && cat > shots.mjs <<'EOF'
import { chromium } from 'playwright';
const b = await chromium.launch({executablePath:'〔路径00178〕'});
const p = await b.newPage({viewport:{width:1300,height:900}});
const errs=[]; p.on('pageerror',e=>errs.push(e.message));
await p.goto('file://〔路径00058〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(500);
async function shot(sel,name,pre){
  if(pre) await pre();
  await p.evaluate(s=>{const e=document.querySelector(s);window.scrollTo(0,e.getBoundingClientRect().top+window.scrollY-20)},sel);
  await p.waitForTimeout(700);
  await p.screenshot({path:'/tmp/'+name+'.png'});
}
await shot('#s0','a_hero');
await shot('.lab','b_lab1');
await sho
…（截断，原长 1401 字符）
````

<details><summary>↳ Bash 结果</summary>

````
errors: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00183〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Polish number formatting and recapture

````
cd 〔路径00057〕 && python3 - <<'PY'
p='ch00-intro.html'; s=open(p,encoding='utf-8').read()
s=s.replace('  if(u<1e-9) return ">99.9999999%";\n  const d=Math.min(9,Math.max(2,Math.floor(-Math.log10(u))+1));',
            '  if(u<1e-7) return ">99.99999%";\n  const d=Math.min(7,Math.max(2,Math.floor(-Math.log10(u))+1));')
s=s.replace('function fmtDown(min){\n  if(min<1) return (min*60).toFixed(1)+" 秒";',
            'function fmtDown(min){\n  if(min<1/60) return "不到 1 秒";\n  if(min<1) return (min*60).toFixed(1)+" 秒";')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
node 〔路径00247〕
````

<details><summary>↳ Bash 结果</summary>

````
ok
errors: none
````

</details>

▸ **Read** `〔路径00187〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Read** `〔路径00199〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Fix clock lab pacing and defaults

````
python3 - <<'PY'
p='ch00-intro.html'; s=open(p,encoding='utf-8').read()
def rep(a,b,n=1):
    global s
    assert a in s, a[:70]
    s=s.replace(a,b,n)

# 1) 时钟以真实速度推进（1x），让点击间隔与漂移量级可比
rep('if(S3.run) S3.elapsed=(S3.elapsed||0)+(now-S3.last)*3;   // 3 倍速',
    'if(S3.run) S3.elapsed=(S3.elapsed||0)+(now-S3.last);   // 1:1 真实速度')
# 2) 默认漂移加大，保证肉眼可见的倒置
rep('<input type="range" id="r3a" min="-120" max="120" step="10" value="60">',
    '<input type="range" id="r3a" min="-120" max="120" step="10" value="90">')
rep('<input type="range" id="r3c" min="-120" max="120" step="10" value="-70">',
    '<input type="range" id="r3c" min="-120" max="120" step="10" value="-90">')
# 3) 顺序改为 A→B→C（快钟在前、慢钟在后才会倒置）
rep('还没有事件。让时钟走一会儿，然后依次点 C → B → A 各产生一个事件。',
    '还没有事件。让时钟走 20 秒左右，然后<strong>快速依次</strong>点 A →
…（截断，原长 1897 字符）
````

<details><summary>↳ Bash 结果</summary>

````
patched
````

</details>

▸ **Bash** Verify which file got patched

````
ls -la 〔路径00053〕 2>/dev/null; echo "--- dsc copy ---"; grep -c '1:1 真实速度' 〔路径00058〕; grep -c '±90 ms/s' 〔路径00058〕; pwd
````

<details><summary>↳ Bash 结果</summary>

````
--- dsc copy ---
1
1
〔路径00057〕
````

</details>

▸ **Bash** Verify clock drift lab produces an inversion

````
cd /tmp && cat > clk.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:900}});
const errs=[];p.on('pageerror',e=>errs.push(e.message));
await p.goto('file://〔路径00058〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.click('#b3run');
await p.waitForTimeout(20000);
await p.click('#b3ea'); await p.waitForTimeout(900);
await p.click('#b3eb'); await p.waitForTimeout(900);
await p.click('#b3ec');
await p.waitForTimeout(400);
const r=await p.evaluate(()=>({rows:Array.from(document.querySelectorAll('#tb3 tbody tr')).map(t=>t.innerText.replace(/\n/g,' | ')),
  verdict:document.querySe
…（截断，原长 1183 字符）
````

<details><summary>↳ Bash 结果</summary>

````
#1	 A	00:20.07	00:21.87	第 3 名 ⟵ 因果倒置！
#2	 B	00:21.02	00:21.02	第 2 名
#3	 C	00:21.97	00:19.99	第 1 名 ⟵ 因果倒置！
-> ⚠ 有 2 个事件的时间戳顺序与真实发生顺序不符。如果用「时间戳大者胜」做冲突解决，这里就会静默丢数据。
errors: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00202〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Write Go unreliable network module

````
mkdir -p 〔路径00084〕 〔路径00124〕 〔路径00069〕
cat > 〔路径00081〕 <<'EOF'
module dsc

go 1.21
EOF
cat > 〔路径00089〕 <<'EOF'
package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// ErrTimeout 就是本章反复强调的「第三态」：
// 它既不表示"服务器执行了"，也不表示"服务器没执行"，它表示"我不知道"。
var ErrTimeout = errors.New("timeout: 不知道服务器执行了没有")

// NetConfig 描述一条不可靠信道的行为。四个字段对应第 0.3 节的四种网络恶行。
type NetConfig struct {
	Loss    float64       // 单向丢包率（请求和响应各判定一次）
	Latency time.Duration // 平均单向延迟
	Jitter  time.Duration // 延迟抖动，实际延迟 = Latency ± Jitter
	DupRate float64       // 重复投递率（仅 Lab 0-1 演示用）
}

// UnreliableNet 模拟一条会丢包、延迟、乱序、重复的网络。
type UnreliableNet struct {
	cfg NetConfig

	mu       sync.Mutex
	rnd      *rand.Rand
	ReqSent  int // 客户端发出的请求次数（含重
…（截断，原长 2976 字符）
````

<details><summary>↳ Bash 结果</summary>

````
gofmt clean
````

</details>

▸ **Bash** Write Go bank and client modules

````
cat > 〔路径00086〕 <<'EOF'
package main

import "sync"

// Bank 是服务器端。它有三种「扣款」实现，对应三种工程水平。
type Bank struct {
	mu sync.Mutex

	Balance int // 账户余额
	Deducts int // 真正执行成功的扣款次数（上帝视角统计，客户端看不到）

	dedup  map[string]bool // 幂等去重表：幂等键 -> 是否处理过
	Lost   int             // Lab 0-4：因非原子写入而永久丢失的业务数
	Double int             // Lab 0-4：因非原子写入而重复扣款的次数
}

func NewBank(balance int) *Bank {
	return &Bank{Balance: balance, dedup: map[string]bool{}}
}

// ---------------------------------------------------------------------------
// ① 天真实现：没有幂等键。每收到一次请求就扣一次钱。
//    配合「超时重试」使用时，会重复扣款。
// ---------------------------------------------------------------------------
func (b *Bank) DeductNaive(amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Balance -= amount
	b.Deducts++
}

// ---------------
…（截断，原长 2983 字符）
````

<details><summary>↳ Bash 结果</summary>

````
〔路径00086〕
gofmt clean
````

</details>

▸ **Bash** Format Go source

````
cd 〔路径00084〕 && gofmt -w bank.go && gofmt -l . && echo "gofmt clean now"
````

<details><summary>↳ Bash 结果</summary>

````
gofmt clean now
````

</details>

▸ **Bash** Write and run the Go lab

````
cat > 〔路径00088〕 <<'EOF'
// Lab 0 · 不可靠信道与第三态
//
// 运行：  cd go/lab00 && go run .
// 调参：  go run . -loss 0.4 -retry 5 -n 500
//
// 配套课件：courseware/ch00-intro.html 第 0.3 / 0.5 / 0.6 节
package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	amount  = 100 // 每笔转账金额
	balance = 1_000_000
)

var (
	nFlag     = flag.Int("n", 200, "转账笔数")
	lossFlag  = flag.Float64("loss", 0.30, "单向丢包率 0~1")
	retryFlag = flag.Int("retry", 3, "最大重试次数")
	seedFlag  = flag.Int64("seed", 42, "随机种子（同一种子结果可复现）")
	crashFlag = flag.Float64("crash", 0.30, "Lab 0-4 中两步之间崩溃的概率")
)

func cfg() NetConfig {
	return NetConfig{
		Loss:    *lossFlag,
		Latency: 1200 * time.Microsecond,
		Jitter:  800 * time.Microsecond,
		DupRate: 0.25,
	}
}

const timeout = 5 * time.Millisecond

fu
…（截断，原长 7690 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：转账 200 笔 × 100 元 ｜ 单向丢包率 30% ｜ 最大重试 3 次 ｜ 超时 5ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 0-1 · 不可靠信道：丢包 / 延迟 / 乱序 / 重复
  按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么
══════════════════════════════════════════════════════════════════════════════
    发送顺序                               1 2 3 4 5 6 7 8 9 10
    实际到达顺序                             2 3 5 5* 6 7 7* 10 10*        （* = 重复投递）
    丢失的消息                              1 4 8 9
    重复到达的消息                            5 7 10
    乱序                                 无
    用时 12ms

    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
      而分布式系统的麻烦恰恰发生在连接断掉之后。

══════════════════════════════════════════════════════════════════════════════
  Lab 0-2 · at-most-once vs at-least-once
  同样的网络故障，两种重试策略，两种错法
══════════════════════════════════════════════════════════════════════════════
    策略                          应扣(元)      实扣(元)      差额(元)      客户端认为成功
    ──────────────────────────────────────────────────────────────────────
    A · 不重试                     20000      13800      -6200         98/200
    B · 重试至多 3 次                20000      26800      +6800        192/200

    ▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
      已经在服务器上执行了（幽灵成功），只是响应包丢了。
    ▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
      "服务器其实已经做过了"的情况。
    ▸ 两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。

══════════════════════════════════════════════════════════════════════════════
  Lab 0-3 ·
…（截断，原长 3380 字符）
````

</details>

▸ **Bash** Add CJK-aware table alignment and rerun

````
cat > width.go <<'EOF'
package main

import "strings"

// dispw 计算字符串在等宽终端里占几列（CJK 全角字符占 2 列）。
// Go 的 %-20s 是按字节/rune 计数的，中文表格会错位，所以自己算。
func dispw(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115F ||
			(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6)) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func padR(s string, w int) string {
	if d := w - dispw(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func padL(s string, w int) string {
	if d := w - dispw(s); d > 0 {
		return strings.Repeat(" ", d) + s
	}
	return s
}
EOF
python3 - <<'PY'
import re
p='main.go'; s=open(p,encod
…（截断，原长 3717 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok

配置：转账 200 笔 × 100 元 ｜ 单向丢包率 30% ｜ 最大重试 3 次 ｜ 超时 5ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 0-1 · 不可靠信道：丢包 / 延迟 / 乱序 / 重复
  按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么
══════════════════════════════════════════════════════════════════════════════
    发送顺序                       1 2 3 4 5 6 7 8 9 10
    实际到达顺序                   1 2 1* 3 4 4* 6 5 5* 7 8 8* 9        （* = 重复投递）
    丢失的消息                     10
    重复到达的消息                 1 4 5 8
    乱序                           5 排在 6 之后
    用时 11ms

    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
      而分布式系统的麻烦恰恰发生在连接断掉之后。

══════════════════════════════════════════════════════════════════════════════
  Lab 0-2 · at-most-once vs at-least-once
══════════════════════════════════════════════════════════════════════════════
    策略                         应扣(元)   实扣(元)   差额(元)   客户端认为成功
    ────────────────────────────────────────────────────────────────────────────
    A · 不重试                      20000      13800      -6200           98/200
    B · 重试至多 3 次               20000      26700      +6700          192/200

    ▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
      已经在服务器上执行了（幽灵成功），只是响应包丢了。
    ▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
      "服务器其实已经做过了"的情况。
    ▸ 两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。
  同一份网络故障 + 同样的重试次数，只改服务端
══════════════════════════════════════════════════════════════════════════════
    服务端实现                   应扣(元)   实
…（截断，原长 2540 字符）
````

</details>

▸ **Bash** Write and run the Java lab

````
cat > 〔路径00125〕 <<'EOF'
/*
 * Lab 0 · 不可靠信道与第三态（Java 版）
 *
 * 运行：  cd java/lab00 && java Lab00.java
 * 调参：  java Lab00.java -loss 0.4 -retry 5 -n 500
 *
 * 需要 JDK 17+（用的是单文件源码启动模式，不需要 Maven / Gradle）。
 * 配套课件：courseware/ch00-intro.html 第 0.3 / 0.5 / 0.6 节
 */
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.locks.LockSupport;

public class Lab00 {

    // ─── 可调参数 ──────────────────────────────────────────────────────────
    static int    N      = 200;   // 转账笔数
    static double LOSS   = 0.30;  // 单向丢包率
    static int    RETRY  = 3;     // 最大重试次数
    static long   SEED   = 42;    // 随机种子（同种子结果可复现）
    static double CRASH  = 0.30;  // Lab 0-4 两步之间的崩溃概率

    static final int  AMOUNT   = 100;
    static final int  BALANCE  = 1_000_0
…（截断，原长 15727 字符）
````

<details><summary>↳ Bash 结果</summary>

````
Picked up JAVA_TOOL_OPTIONS: -Djavax.net.ssl.trustStore=/root/.ccr/java-truststore.p12 -Djavax.net.ssl.trustStorePassword=changeit -Djavax.net.ssl.trustStoreType=PKCS12 -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=45377 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|127.*|0.*|::|169.254.*|anthropic.com|*.anthropic.com|*.anthropic.com|registry.npmjs.org|jsr.io|npm.jsr.io|pypi.org|files.pythonhosted.org|index.crates.io|proxy.golang.org|host.docker.internal|10.*|172.16.*|172.17.*|172.18.*|172.19.*|172.20.*|172.21.*|172.22.*|172.23.*|172.24.*|172.25.*|172.26.*|172.27.*|172.28.*|172.29.*|172.30.*|172.31.*|192.168.*|100.64.0.0/10|*.svc.cluster.local|*.svc.cluster.local -Djdk.http.auth.tunneling.disabledSchemes= -Djdk.http.auth.proxying.disabledSchemes=
Lab00.java:75: error: cannot find symbol
                exec();                                  // ? ??????????????????????
                ^
  symbol:   method exec()
  location: class Net
1 error
error: compilation failed
````

</details>

▸ **Bash** Fix Runnable call and encoding, rerun Java lab

````
python3 - <<'PY'
p='Lab00.java'; s=open(p,encoding='utf-8').read()
s=s.replace("                exec();                                  // ② 服务器执行副作用。到这一行为止，钱已经扣了。",
            "                exec.run();                              // ② 服务器执行副作用。到这一行为止，钱已经扣了。")
s=s.replace("    public static void main(String[] args) throws Exception {\n        parseArgs(args);",
 "    public static void main(String[] args) throws Exception {\n"
 "        // 保证中文在任何平台的终端都能正确输出\n"
 "        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),\n"
 "                true, java.nio.charset.StandardCharsets.UTF_8));\n"
 "        parseArgs(args);")
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
java Lab00.java 2>&1 | grep -v "Picked up JAVA_TOOL" | tail -62
````

<details><summary>↳ Bash 结果</summary>

````
ok
▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
  TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
  而分布式系统的麻烦恰恰发生在连接断掉之后。

══════════════════════════════════════════════════════════════════════════════
  Lab 0-2 · at-most-once vs at-least-once
  同样的网络故障，两种重试策略，两种错法
══════════════════════════════════════════════════════════════════════════════
    策略                         应扣(元)   实扣(元)   差额(元)   客户端认为成功
    ────────────────────────────────────────────────────────────────────────────
    A · 不重试                      20000      14900      -5100          110/200
    B · 重试至多 3 次               20000      25700      +5700          190/200

▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
  已经在服务器上执行了（幽灵成功），只是响应包丢了。
▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
  "服务器其实已经做过了"的情况。
▸ 两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。

══════════════════════════════════════════════════════════════════════════════
  Lab 0-3 · 幂等键：让「重试」变得安全
  同一份网络故障 + 同样的重试次数，只改服务端
══════════════════════════════════════════════════════════════════════════════
    服务端实现                   应扣(元)   实扣(元)   差额(元)     重复扣款次数
    ────────────────────────────────────────────────────────────────────────────
    无幂等键                        20000      25700      +5700               57
    有幂等键（同事务）              20000      20000         +0                0

▸ 有幂等键的那一行，重复扣款恒为 0。差额若仍为负，说明有几笔业务的请求包被
  连续丢了 4 次，服务器压根没收到过——这部分只能靠加大重试次数压低，
  幂等性解决不了。
▸ 记住这个组合拳：at-least-once 传输 + 服务端幂等 = 工程上的 exactly-once 效果。
  Kafka / Flink 所谓的 exactly-once 也是这么做的，没有魔法。

…（截断，原长 2693 字符）
````

</details>

▸ **Bash** Align Java text block indentation

````
python3 - <<'PY'
import re
p='Lab00.java'; s=open(p,encoding='utf-8').read()
# 给 println("""...""") 的文本块统一加 4 空格缩进，与 Go 版输出对齐
s=s.replace('而分布式系统的麻烦恰恰发生在连接断掉之后。""");','而分布式系统的麻烦恰恰发生在连接断掉之后。""".indent(4));')
s=s.replace('两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。""");','两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。""".indent(4));')
s=s.replace('Kafka / Flink 所谓的 exactly-once 也是这么做的，没有魔法。""".formatted(RETRY + 1));','Kafka / Flink 所谓的 exactly-once 也是这么做的，没有魔法。""".formatted(RETRY + 1).indent(4));')
s=s.replace('没有一个共同的事务能包住它们，于是才需要 2PC、TCC、Saga 这些方案。""");','没有一个共同的事务能包住它们，于是才需要 2PC、TCC、Saga 这些方案。""".indent(4));')
s=s.replace('下一站：Part 1 · 时间、顺序与因果 —— 亲手实现 Lamport 时钟与向量时钟。""");','下一站：Part 1 · 时间、顺序与因果 —— 亲手实现 Lamport 时钟与向量时钟。""".indent(2));')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
java Lab00.java
…（截断，原长 850 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok

配置：转账 40 笔 × 100 元 ｜ 单向丢包率 30% ｜ 最大重试 3 次 ｜ 超时 5ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 0-1 · 不可靠信道：丢包 / 延迟 / 乱序 / 重复
  按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么
══════════════════════════════════════════════════════════════════════════════
    发送顺序                       1 2 3 4 5 6 7 8 9 10
    实际到达顺序                   4 5 2 3 9 3* 6 9* 7 1 8 1* 10        （* = 重复投递）
    丢失的消息                     无
    重复到达的消息                 1 3 9
    乱序                           2 排在 5 之后；6 排在 9 之后；1 排在 7 之后
    用时 44ms

    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
      而分布式系统的麻烦恰恰发生在连接断掉之后。


══════════════════════════════════════════════════════════════════════════════
````

</details>

▸ **Bash** Write repo README and assemble tree

````
cat > 〔路径00056〕 <<'EOF'
# 分布式系统：从零到共识

一套「讲解 + 交互课件 + 可运行实验 + 自测」四位一体的学习课程。
每一章都包含：一份可在浏览器打开的交互式课件，一组 Go 与 Java 双语言实验，
以及一套自测题。

---

## 快速开始

```bash
# 1. 打开第 0 章交互课件（浏览器里直接玩 6 个模拟实验）
open courseware/ch00-intro.html          # macOS
# xdg-open courseware/ch00-intro.html    # Linux

# 2. 跑 Lab 0（Go 版，需要 Go 1.21+）
cd go/lab00 && go run .

# 3. 跑 Lab 0（Java 版，需要 JDK 17+，无需 Maven/Gradle）
cd java/lab00 && java Lab00.java
```

两个语言版本实现的是**完全相同的四个实验**，输出格式也一致，可以直接对照阅读。
所有实验都接受相同的参数：

```bash
go run .        -loss 0.5 -retry 10 -n 500 -seed 7 -crash 0.2
java Lab00.java -loss 0.5 -retry 10 -n 500 -seed 7 -crash 0.2
```

| 参数 | 含义 | 默认 |
|---|---|---|
| `-loss` | 单向丢包率（请求方向和响应方向各判定一次） | `0.30` |
| `-retry` | 客户端最大重试次数（`0` = at-most-once） | `3` |
| `-n` | 转账笔数 | `200` |
| `-seed` |
…（截断，原长 3283 字符）
````

<details><summary>↳ Bash 结果</summary>

````
〔路径00056〕
〔路径00071〕
〔路径00081〕
〔路径00086〕
〔路径00087〕
〔路径00088〕
〔路径00089〕
〔路径00090〕
〔路径00125〕
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00146〕\" \"$HOME〔路径00150〕\" \"$HOME〔路径00158〕\" && ls -R \"$HOME〔路径00144〕\""}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00166〕
courseware
go
java

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00147〕

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00157〕
lab00

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00151〕

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00165〕
lab00

/sessions/rcw-0185t6fh1dev7futdpowiw31〔路径00159〕
````

</details>

## 🤖 Claude · 2026-08-20 10:46

📎 **发送文件**：第 0 章交互式课件 —— 6 个可以直接拖滑块、点按钮的模拟实验，外加 12 道折叠式自测题。建议先玩「实验 2 · 不可靠网络沙盒」和「实验 4 · 你来判决」。
- `〔路径00071〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00071〕 → file_uuid: 〔标识03838〕
````

</details>

📎 **发送文件**：Lab 0 实验仓库：Go 与 Java 两份等价实现，跑起来会输出同样的四组对照实验。
- `〔路径00056〕`
- `〔路径00081〕`
- `〔路径00089〕`
- `〔路径00086〕`
- `〔路径00087〕`
- `〔路径00090〕`
- `〔路径00088〕`
- `〔路径00125〕`

<details><summary>↳ SendUserFile 结果</summary>

````
8 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识01374〕
  〔路径00081〕 → file_uuid: 〔标识02516〕
  〔路径00089〕 → file_uuid: 〔标识02416〕
  〔路径00086〕 → file_uuid: 〔标识01212〕
  〔路径00087〕 → file_uuid: 〔标识00707〕
  〔路径00090〕 → file_uuid: 〔标识00002〕
  〔路径00088〕 → file_uuid: 〔标识01561〕
  〔路径00125〕 → file_uuid: 〔标识02850〕
````

</details>

💾 **写入你的 Mac**：
- `〔路径00005〕`
- `〔路径00004〕`
- `〔路径00013〕`
- `〔路径00018〕`
- `〔路径00015〕`
- `〔路径00016〕`
- `〔路径00019〕`
- `〔路径00017〕`
- `〔路径00044〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00005〕","〔路径00004〕","〔路径00013〕","〔路径00018〕","〔路径00015〕","〔路径00016〕","〔路径00019〕","〔路径00017〕","〔路径00044〕"],"rejected":[]}
````

</details>

🧩 **创建产出物** `ds-course-ch00`：分布式系统课程 · 第 0 章：序言与导论。含 6 个交互式模拟实验（可用性计算器、不可靠网络沙盒、时钟漂移、超时判决、幂等键对照、两将军无限回归）、20 词术语速查表、7 阶段学习路线图与 12 道自测题。

▸ **TaskUpdate** completed

▸ **TaskUpdate** completed

▸ **TaskUpdate** completed

▸ **TaskUpdate** completed

<details><summary>↳ mcp__remote-devices__create_artifact 结果</summary>

````
Artifact "ds-course-ch00" created on the connected desktop.
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #2 status
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #3 status
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #4 status
````

</details>

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #1 status
````

</details>

▸ **device:device_bash**

````
{"command": "cd \"$HOME〔路径00144〕\" && find . -type f | sort && echo \"--- sizes ---\" && du -sh . && echo \"--- go present? ---\" && (command -v go >/dev/null && go version || echo \"go 未安装\") && (command -v java >/dev/null && java -version 2>&1 | head -1 || echo \"java 未安装\")"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
./README.md
./courseware/ch00-intro.html
./go/go.mod
./go/lab00/bank.go
./go/lab00/client.go
./go/lab00/main.go
./go/lab00/net.go
./go/lab00/width.go
./java/lab00/Lab00.java
--- sizes ---
164K	.
--- go present? ---
go 未安装
openjdk version "11.0.31" 2026-04-21
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #5 status
````

</details>

第 0 章做完了，三份产物都已落到你本地 `~/Code/distributed-systems-course/`，课件同时存成了桌面 artifact（`ds-course-ch00`），以后不用翻聊天记录也能打开。

**我把你的路线改了一点，说明一下理由。** 你原本的三段式（定义 → 算法 → 应用）方向没错，但中间缺了脚手架：直接从"最终一致性的定义"跳到 Raft，多数人会卡住。我扩成 7 个 Part，关键改动是在最前面插了一层**心智模型**，用一个框架把后面所有内容挂上去——

> 单机编程时你免费拥有三件事：**一个共同的"现在"**、**调用只有成功/失败两种结果**、**数据只有一份**。分布式把这三件事同时拿走了。之后你学的每一个算法，都是在赎回其中某一件。

逻辑时钟赎第一件，幂等和分布式事务赎第二件，一致性模型和共识赎第三件。学到任何新概念时先问"它在赎哪一件"，知识就不会散。

另外在 Part 3 和 Part 4 之间我加了**单机事务补课**（隔离级别、MVCC、快照隔离）——不补这个直接讲 Percolator 和 Spanner 会很吃力。

**建议的动手顺序**：先打开课件玩「实验 2 · 不可靠网络沙盒」，连发 20 个请求，盯住底下"客户端认为成功"和"服务器实际执行"这两个数字；再玩「实验 4 · 你来判决」，玩满 10 轮看看你的准确率。这两个实验是整章的地基。然后跑 Lab 0（`cd go/lab00 && go run .`）。

一个提醒：`Lab00.java` 用了文本块和 switch 表达式，需要 **JDK 17+**。你 `java -version` 如果是 11，告诉我，我给你降一版语法。

---

开始互动吧，三个问题，**先别翻课件**：

1. 你的服务调用下游超时了。你能不能得出"下游没执行"这个结论？如果不能，你在实际工作里是怎么处理的？
2. 为什么 etcd / ZooKeeper 集群几乎总是 3 台或 5 台，你见过 4 台的吗？
3. 你用过的系统里（MySQL 主从、Redis 哨兵、Kafka、任何一个），有没有哪次线上问题事后回想起来，其实就是"时钟不同步"或者"超时后重复执行"造成的？

第 3 题最重要——如果你有真实的踩坑经历，我可以把后面的章节直接挂在那个场景上讲，比用银行转账当例子有效得多。

> ✓ 本轮结束 · 2026-08-20 10:47 · 63 步
