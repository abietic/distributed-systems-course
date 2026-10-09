## 👤 你 · 2026-09-05 20:14

1. 因为不知道是否有其他的参与者
2. 应该是旧数据，应该只影响不能再做写了
3. 我觉得可以了

## 🤖 Claude · 2026-09-05 20:15

## 第 1 题 — 方向对了，但还差一层

"不知道是否有其他参与者"——参与者名单其实 P2/P3 是知道的（协调者在 prepare 里就告诉它们了，或者本来就是固定集合）。所以它们**可以**去问 P1。这叫 cooperative termination protocol，确实能救回一部分情况。

但救不回全部。真正的原因往下再挖一层：

> **「决定」这条信息，在整个集群里只有一份副本——协调者的日志。**

P2/P3 互相通信得到的状态是「我俩都投了 yes、都没收到指令」。这个状态在两个世界里<u>完全相同</u>：
- 世界 A：协调者决定了 commit，发给了 P1，然后崩溃 → 它们必须 commit
- 世界 B：协调者决定了 abort（或还没来得及决定）→ 它们必须 abort

要区分这两个世界，唯一的信息源要么是协调者本身，要么是 P1。而 P1 不可达时，"P1 已提交但联系不上"和"P1 根本没收到"从 P2/P3 的角度**不可分辨**——又回到 Part 0 §0.5 的原罪。

**所以 2PC 的病根不在"投票"，在"决定没有被复制"。** 而共识做的恰恰就是这件事：把决定复制到过半节点，于是任何一个过半集合都能把它恢复出来。这就是为什么把协调者本身做成一个 Raft 组就能解决阻塞——不是加了个备份，而是**让决定本身获得了多副本**。

## 第 2 题 — 写的部分需要修正，读的部分取决于实现

**读**：你说的对，但要看读是怎么实现的，三种做法结果完全不同：

| 读的实现 | 僵尸 Leader 上会发生什么 |
|---|---|
| **本地读**（朴素做法） | **静默返回陈旧数据**。最危险的一种，因为它不报错 |
| **ReadIndex** | Leader 读前必须先跟过半节点确认自己还是主 → **超时/报错**，不会给你旧数据 |
| **Lease Read** | 租约期内本地读（安全，因为新主要等旧租约过期才能选出）；租约到期后报错 |

**写**：不是"不能写"。僵尸 Leader **会接受你的写请求**，把它追加进自己的日志，然后等过半确认——永远等不到。客户端看到的是**超时**，不是明确的失败。而且更糟：那条未提交的日志在节点重新加入集群后会被**截断丢弃**。

所以客户端体验是：**超时了，然后数据消失了**。第三态，又一次。

## 第 3 题 — 你答错了，而且这正是我想要的

> **不能。**

一条日志复制到过半节点，**不等于**它可以被提交。这是 Raft 论文 Figure 8 讲的事，也是整个 Raft 里最反直觉的一处——几乎所有人第一次都答"可以了"，包括 Raft 作者自己在早期草稿里也这么写过，后来发现有反例才补上了限制。

精确的规则是：

> **Leader 只能通过"计数副本"来提交<u>自己当前任期</u>的日志条目。**
> 前任任期留下的条目，即使已经复制到过半节点，也不能直接提交——只能等当前任期的某条日志被提交之后，靠日志匹配性质<u>间接</u>提交。

为什么？因为存在一个执行序列，让一条"已经复制到过半"的前任日志<strong>之后被覆盖掉</strong>。如果你在它复制到过半时就宣布提交并回复客户端"成功"，那客户端看到的成功就是假的——**已提交的数据被推翻，安全性直接破产**。

这个反例有五个阶段，只用五个节点就能构造出来。3-B 的核心实验就是让你一格一格把它走一遍，并且给你一个开关：打开"按副本数提交前任日志"，断言当场炸给你看。

现在做 3-B。

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **TaskCreate** 编写 Part 3-B 课件：日志复制与安全性

<details><summary>↳ TaskCreate 结果</summary>

````
Task #15 created successfully: 编写 Part 3-B 课件：日志复制与安全性
````

</details>

▸ **TaskCreate** 编写 Lab 3-B：日志复制实现（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #16 created successfully: 编写 Lab 3-B：日志复制实现（Go + Java）
````

</details>

▸ **TaskCreate** 验证并交付 Part 3-B

<details><summary>↳ TaskCreate 结果</summary>

````
Task #17 created successfully: 验证并交付 Part 3-B
````

</details>

▸ **Bash** Write Part 3-B head and sections 3.10-3.11

````
cd 〔路径00054〕 && mkdir -p ch03b && cd ch03b && cat > ch03b.html <<'HEAD_EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 3-B · 日志复制与安全性</title>
HEAD_EOF
cat 〔路径00240〕 >> ch03b.html
cat >> ch03b.html <<'HEAD2_EOF'
<style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  background:var(--surface-2);border:1px solid var(--border);color:var(--ink)}
.axm{display:grid;grid-template-columns:26px 1fr;gap:10px
…（截断，原长 8401 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-B sections 3.12-3.16

````
cat >> ch03b.html <<'EOF'
<section id="s3">
  <h2>3.12 提交规则：什么时候才能对客户端说"成功"</h2>
  <p>Leader 为每个 follower 维护两个指针：</p>
  <dl class="kv" style="display:grid;grid-template-columns:auto 1fr;gap:8px 16px;font-size:14px;margin:14px 0">
    <dt style="color:var(--ink-3);font-weight:600"><code>nextIndex[i]</code></dt>
    <dd style="margin:0;color:var(--ink-2)"><strong>乐观猜测</strong>：下一条要发给节点 i 的日志位置。当选时初始化为「自己日志长度 + 1」，被拒绝就回退。</dd>
    <dt style="color:var(--ink-3);font-weight:600"><code>matchIndex[i]</code></dt>
    <dd style="margin:0;color:var(--ink-2)"><strong>保守事实</strong>：已确认节点 i 与自己一致的最高位置。只在 AppendEntries 成功时前进。</dd>
  </dl>
  <p>然后提交规则是（Raft 论文 §5.4.2，逐字翻译）：</p>
  <div class="note crit">
    <p style="margin-bottom:0">存在一个 <code>N &gt; commitIndex</code>，使得<br>
    ① <strong>过半</strong>的
…（截断，原长 10043 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-B quiz and lab sections

````
cat >> ch03b.html <<'EOF'
<section id="s8">
  <h2>3.17 自测题</h2>
  <p>先自己想 30 秒再展开。</p>

  <details class="q"><summary>1. 用一句话说清日志匹配性质，以及它为什么能从一个"只看一个位置"的检查里得到。</summary>
    <div class="a"><p><strong>性质</strong>：若两条日志在同一 index 上的条目 term 相同，则该 index 及之前的<em>全部</em>条目都相同。</p>
    <p style="margin-bottom:0"><strong>怎么来的</strong>：AppendEntries 的一致性检查就是一次归纳步骤。空日志平凡匹配（基础情形）；每次 AppendEntries 成功，都证明 follower 在 <code>prevLogIndex</code> 处与 Leader 一致，由归纳假设从头到那里全一致，追加后到新位置也一致（归纳步骤）。<em>局部检查维持了全局不变量</em>——这是整个 Raft 里最优雅的一处设计。</p></div></details>

  <details class="q"><summary>2. 一条日志复制到过半节点，能提交吗？给出完整的规则。</summary>
    <div class="a"><p><strong>不一定。</strong>提交需要同时满足两个条件：<br>
    ① 过半的 <code>matchIndex[i] ≥ N</code>；<br>
    ② <code>log[N].term == currentTerm</code>——只能提交<u>当前任期</u>创建的条目。</p>
    <p sty
…（截断，原长 7413 字符）
````

<details><summary>↳ Bash 结果</summary>

````
sections: 10 kb: 48
````

</details>

▸ **Bash** Write Part 3-B JS: log replication experiment

````
cat >> ch03b.html <<'EOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(t,a,x){const e=document.createElementNS(NS,t);for(const k in a)e.setAttribute(k,a[k]);if(x!=null)e.textContent=x;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}
const TC=["--ink-3","--s1","--s3","--s7","--s2","--s4","--ok","--warn","--crit"];
function termColor(t){return cv(TC[t%TC.length]);}

$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");draw1();draw2();draw3();};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function
…（截断，原长 6038 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Figure 8 replay and verify Part 3-B

````
cat >> ch03b.html <<'EOF'

/* ══════════════════════════════════════════════════════════════════
   实验 2 · Figure 8 交互重放
   ══════════════════════════════════════════════════════════════════ */
const F8=[
 {t:"(a) 任期 2：S1 是 Leader，把 index2 部分复制出去",
  leader:0, term:2,
  logs:[[1,2],[1,2],[1],[1],[1]],
  note:"S1 在任期 2 当选，接收客户端请求写下 <span class='chip'>index2 term2</span>，只复制到了 S2。<br>"+
       "2/5 个节点持有它，<strong>没有过半</strong>，当然不能提交。到这里一切正常。",
  hi:{idx:2,nodes:[0,1],kind:"warn"}},
 {t:"(b) 任期 3：S1 崩溃，S5 当选，在同一位置写下不同内容",
  leader:4, term:3, dead:[0],
  logs:[[1,2],[1,2],[1],[1],[1,3]],
  note:"S1 崩溃。S5 拿到 S3、S4 和自己的票当选任期 3（三者日志都是 <code>[1]</code>，S5 不比谁旧）。<br>"+
       "S5 在<strong>同一个 index2</strong> 上写下了 <span class='chip'>term3</span> 的<em>不同内容</em>，还没来得及复制出去就崩了。<br>"+
       "<span styl
…（截断，原长 10787 字符）
````

<details><summary>↳ Bash 结果</summary>

````
scen 0 => 全部追平 | F1  (7 轮)
scen 1 => 全部追平 | F1  (1 轮)
scen 2 => 全部追平 | F1  (6 轮)
scen 3 => 全部追平 | (a)  (2 轮) (b)  (7 轮) (c)  (1 轮) (d)  (1 轮) (e)  (6 轮) (f)  (8 轮)
naive=false => — || — || Raft 在这里踩了刹车 || 覆盖发生了，但没有造成任何伤害 || 安全的那条路
naive=true => — || — || 你关掉了限制 —— 这里 Leader 会宣布 index2 已提交 || ★ 安全性破产：已经告诉客户端"成功"的数据凭空消失了 || 安全的那条路
exp3: 朴素回退=41 | 快速回退=4 | 加速比=10.3× | 节省时间=2960
exp3 gap=200: 201 | 4 | 50.3× | 15760
ERRORS: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Bash** Screenshot Figure 8 and adapt network for logs

````
cd /tmp && cat > s3b.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:1050}});
await p.goto('file://〔路径00067〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(400);
await p.evaluate(()=>{const c=document.querySelector('#c2naive');c.checked=true;c.dispatchEvent(new Event('change'))});
for(let i=0;i<3;i++) await p.click('#b2next');
await p.evaluate(()=>{const e=document.querySelector('#s4 .lab');window.scrollTo(0,e.getBoundingClientRect().top+window.scrollY-14)});
await p.waitForTimeout(500); await p.screenshot({path:'〔路径00221〕'});
await b.close();
EOF
node
…（截断，原长 2035 字符）
````

<details><summary>↳ Bash 结果</summary>

````
net.go adapted
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00221〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Write Go Raft log replication for Lab 3-B

````
cd 〔路径00066〕 && python3 - <<'PY'
p='ch03b.html'; s=open(p,encoding='utf-8').read()
old='''    svg.appendChild(el("rect",{x:x+1,y:y,width:cw-3,height:24,rx:4,
      fill:hi==="del"?"none":termColor(t),
      stroke:hi?cv(hi==="del"?"--crit":hi==="new"?"--ok":hi==="chk"?"--warn":"--border"):cv("--border"),
      "stroke-width":hi?2.5:1, opacity:hi==="del"?.5:1,
      "stroke-dasharray":hi==="del"?"3 3":""}));
    svg.appendChild(el("text",{x:x+cw/2-1,y:y+17,fill:hi==="del"?cv("--crit"):"#fff","font-size":11.5,
      "font-weight":700,"text-anchor":"middle","font-family":cv("--mono")},t));'''
new='''    const stroke = hi==="del"?"--crit" : hi==="ow"?"--crit" : hi==="new"?"--ok" : hi==="chk"?"--warn" : "--border";
    svg.appendChild(el("rect",{x:x+1,y:y,width:cw-3,height:24,rx:4,
…（截断，原长 12838 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
raft3b written
````

</details>

▸ **Bash** Write Go Lab 3-B main

````
cat > 〔路径00105〕 <<'GOEOF'
// Lab 3-B · 日志复制与安全性
//
// 运行：  cd go/lab03b && go run .
// ★关键： go run . -naivecommit    ← 关掉「只提交当前任期」，看 Figure 8 怎么炸
//         go run . -slowbackoff    ← 关掉快速回退，对比 RPC 轮数
//
// 配套课件：courseware/ch03b-log-replication.html
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
	lossF   = flag.Float64("loss", 0.0, "丢包率")
	seedF   = flag.Int64("seed", 42, "随机种子")
	naiveF  = flag.Bool("naivecommit", false, "★ 去掉「只提交当前任期」的限制")
	slowF   = flag.Bool("slowbackoff", false, "关掉快速回退，每次只退 1")
	vF      = flag.Bool("v", false, "打印细节")
)

const tick = 5

func build(seed int64) *Cluster {
	net := NewNetw
…（截断，原长 11338 字符）
````

<details><summary>↳ Bash 结果</summary>

````
(Bash completed with no output)
````

</details>

▸ **Bash** Run Go Lab 3-B

````
cd lab03b && go run . 2>&1 | head -50
````

<details><summary>↳ Bash 结果</summary>

````
配置：5 节点 ｜ 延迟 20±10ms ｜ 丢包 0% ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3B-1 · 日志复制
  提交 200 条命令，验证所有节点的日志逐条相同
══════════════════════════════════════════════════════════════════════════════
    指标                                  数值说明
    ──────────────────────────────────────────────────────────────────────────────────────
    提交的命令数                           200客户端发起
    最高 commitIndex                       200已达成共识的日志条数
    AppendEntries 成功                     364次
    AppendEntries 被拒                       0一致性检查未通过
    所有节点日志一致                      ✓ 是逐条比对通过

    N1 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …
    N2 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …
    N3 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …

    ▸ 所有节点的日志逐条相同，且状态机在每个 index 上执行了相同的命令。
    ▸ 注意「被拒」那一行：即使没有故障，Leader 刚当选时也会因为 nextIndex 猜得太乐观
      而被拒几次 —— 那正是它在二分查找每个 follower 的分歧点。

══════════════════════════════════════════════════════════════════════════════
  Lab 3B-2 · 三条安全性断言
  边跑边杀节点、边制造分区，每个 tick 都检查
══════════════════════════════════════════════════════════════════════════════
    场景                              提交的命令      已提交        违反          安全性
    ────────────────────────────────────────────────────────────────────────────────────
    ① 无故障，持续写入                       120         120           0          ✓ 成立
    ② 边写边杀 L
…（截断，原长 2550 字符）
````

</details>

▸ **Bash** View Lab 3B-3 and 3B-4 output

````
go run . 2>&1 | sed -n '50,90p'
````

<details><summary>↳ Bash 结果</summary>

````
    被覆盖的是「已提交」的日志吗                            否                是
    安全性断言                                          ✓ 成立          ✗ 被破坏

    ② 报出的违反
      【Leader 完整性被破坏】N5 在 term=5 当选，但它缺少已提交的 index=2 (term=2, cmd=X)

    ▸ 两边的物理事实完全一样：那条 term2 的日志都复制到了 3/5 个节点，
      也都被 S5 覆盖掉了。唯一的区别是——① 从没把它叫做「已提交」。
    ▸ Raft 没有去阻止覆盖（那需要改选举规则，代价大得多），
      而是确保「被覆盖的东西从来没被承诺过」。这是一个非常克制的修补。
    ▸ 对客户端的含义：① 里客户端收到的是超时（第三态），数据可能在也可能不在；
      ② 里客户端收到的是"成功"，然后数据消失了。后者才是真正的事故。

══════════════════════════════════════════════════════════════════════════════
  Lab 3B-4 · 冲突回退：朴素 vs 快速
  让一个 follower 落后很多，数它追平要几轮 RPC
══════════════════════════════════════════════════════════════════════════════
    回退策略                          被拒次数    追平耗时(ms)      最终是否追平
    ────────────────────────────────────────────────────────────────────────────
    朴素回退（每次 −1）                    163           16345                 ✓
    快速回退（按任期跳）                     1             135                 ✓

    ▸ 一个任期内的日志是同一个 Leader 连续写下的，要么整段一致要么整段不一致
      （日志匹配性质）。所以在冲突任期内部逐条回退是纯粹的浪费。
    ▸ 真实系统里日志常有几万条却只跨越几个任期，这个优化能把几万轮 RPC 降到个位数。
      落后节点追不上，等于集群实际少了一个副本 —— 容错能力悄悄下降了。

══════════════════════════════════════════════════════════════════════════════
  三个思考题（答案在课件 §3.12 / §3.13）：

    1. -naivecommit 下被覆盖的那条日志，在被覆盖之前有几个节点持有它？
       "过半"为什么救不了它？
    2. 如果同时打开 -naivecommit 并让新 Leader 上任就追加一条 no-op 日志，
       Figure 8 还会发生吗？先想清楚再改代码试。
    3. 把 -loss 调到 0.4，Lab 3B-2 的日志匹配断言会不会被破坏？为什么？

  下一节：Part 3-C · 持久化、快
…（截断，原长 1592 字符）
````

</details>

▸ **Bash** Clean dead code and fix table alignment

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
# 1) 删掉死代码
start=s.index('\t\treport := func(label string, naive bool) {')
end=s.index('\t// 为了并排展示，把两次结果收集起来')
s=s[:start]+s[end:]
s=s.replace('''	report := func(label string, naive bool) {''','')
s=s.replace('	_ = report\n','')
s=s.replace('\nfunc return2(c *Cluster, a, b, n bool, w []int, al string) {}\n','')
# 2) 表格加分隔列
s=s.replace('''	w := []int{26, 16, 44}
	al := "LRL"
	tui.TableHeadA([]string{"指标", "数值", "说明"}, w, al)''',
'''	w := []int{26, 14, 4, 44}
	al := "LRLL"
	tui.TableHeadA([]string{"指标", "数值", "", "说明"}, w, al)''')
for a,b in [('{"提交的命令数", fmt.Sprint(submitted), "客户端发起"}','{"提交的命令数", fmt.Sprint(submitted), "", "客户端发起"}'),
            ('{"最高 commitIndex", fmt.Sprint(c.MaxCommit()), "已达成共识的日志条数"}','{"最高 commitInde
…（截断，原长 1826 字符）
````

<details><summary>↳ Bash 结果</summary>

````
Traceback (most recent call last):
  File "<stdin>", line 3, in <module>
ValueError: substring not found
══════════════════════════════════════════════════════════════════════════════
    指标                                  数值说明
    ──────────────────────────────────────────────────────────────────────────────────────
    提交的命令数                           200客户端发起
    最高 commitIndex                       200已达成共识的日志条数
    AppendEntries 成功                     364次
    AppendEntries 被拒                       0一致性检查未通过
    所有节点日志一致                      ✓ 是逐条比对通过

    N1 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …
````

</details>

▸ **Bash** Locate dead code lines

````
grep -n 'report := func' main.go; grep -n '_ = report' main.go; grep -n 'func return2' main.go
````

<details><summary>↳ Bash 结果</summary>

````
235:	report := func(label string, naive bool) {
314:	_ = report
348:func return2(c *Cluster, a, b, n bool, w []int, al string) {}
````

</details>

▸ **Bash** Remove dead code and fix table

````
python3 - <<'PY'
p='main.go'; lines=open(p,encoding='utf-8').read().split('\n')
# 删除 report 闭包（235..(collect 之前)）、_ = report、return2
out=[]; i=0
while i < len(lines):
    ln=lines[i]
    if ln.strip().startswith('report := func(label string, naive bool) {'):
        # 跳到 "// 为了并排展示" 之前
        while i < len(lines) and '为了并排展示' not in lines[i]:
            i+=1
        continue
    if ln.strip()=='_ = report':
        i+=1; continue
    if ln.startswith('func return2('):
        i+=1; continue
    out.append(ln); i+=1
s='\n'.join(out)
def rep(a,b):
    global s; assert a in s, a[:50]; s=s.replace(a,b)
rep('''	w := []int{26, 16, 44}
	al := "LRL"
	tui.TableHeadA([]string{"指标", "数值", "说明"}, w, al)''',
'''	w := []int{26, 14, 4, 44}
	al := "LRLL"
	tui.TableHeadA([]string{"指标", "数值", "", "说明"}, w
…（截断，原长 1964 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
    指标                                数值    说明
    ────────────────────────────────────────────────────────────────────────────────────────
    提交的命令数                         200    客户端发起
    最高 commitIndex                     200    已达成共识的日志条数
    AppendEntries 成功                   364    次
    AppendEntries 被拒                     0    一致性检查未通过
    所有节点日志一致                    ✓ 是    逐条比对通过

    N1 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …
    N2 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …
    N3 日志前 20 条的 term         1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 …

    ▸ 所有节点的日志逐条相同，且状态机在每个 index 上执行了相同的命令。
````

</details>

▸ **Bash** Write and run Java Lab 3-B

````
mkdir -p 〔路径00134〕 && cat > 〔路径00135〕 <<'JEOF'
/*
 * Lab 3-B · 日志复制与安全性（Java 版）
 *
 * 运行：  cd java/lab03b && java Lab03B.java
 * ★关键： java Lab03B.java -naivecommit   ← 关掉「只提交当前任期」，看 Figure 8 怎么炸
 *         java Lab03B.java -slowbackoff   ← 关掉快速回退，对比 RPC 轮数
 *
 * 需要 JDK 17+。配套课件：courseware/ch03b-log-replication.html
 */
import java.util.*;

public class Lab03B {

    static int N = 5;
    static long DELAY = 20, JITTER = 10, SEED = 42;
    static double LOSS = 0.0;
    static boolean NAIVE = false, SLOW = false, VERBOSE = false;
    static final long TICK = 5, HEARTBEAT = 100;
    static final int NO_VOTE = -1;

    // ══════════════════════════════════════════════════════════════════════
    // 确定性网络
    // ═════════════════════════════════
…（截断，原长 27618 字符）
````

<details><summary>↳ Bash 结果</summary>

````
    ▸ 所有节点的日志逐条相同，且状态机在每个 index 上执行了相同的命令。
    ▸ 「被拒」是 0，因为这是个全新集群、没有任何故障，所有节点的日志从一开始就一致。
      只有当日志真的分叉过（换过 Leader、发生过分区），一致性检查才会拒绝 ——
      Lab 3B-4 会故意制造那种局面。


══════════════════════════════════════════════════════════════════════════════
  Lab 3B-2 · 三条安全性断言
  边跑边杀节点、边制造分区，每个 tick 都检查
══════════════════════════════════════════════════════════════════════════════
    场景                              提交的命令      已提交        违反          安全性
    ────────────────────────────────────────────────────────────────────────────────────
    ① 无故障，持续写入                       120         120           0          ✓ 成立
    ② 边写边杀 Leader                        120         108           0          ✓ 成立
    ③ 反复 3|2 分区                           84          80           0          ✓ 成立
    ④ 30% 丢包 + 随机杀节点                  150         150           0          ✓ 成立

      客户端看到的是超时，然后数据消失。这就是 §3.10 说的第三态。


══════════════════════════════════════════════════════════════════════════════
  Lab 3B-3 · Figure 8 复现
  手工构造论文里那五个节点的状态，然后按规则推进
══════════════════════════════════════════════════════════════════════════════
    阶段 (c)→(d) 发生了什么                        ① 完整 Raft        ② 关掉限制
    ────────────────────────────────────────────────────────────────────────────
    index2 (term2) 复制到过半了吗                     是 (3/5)          是 (3/5)
    S1 把它标记为已提交了吗                                 否                是
    S5 当选后覆盖掉它了吗                                   是                是
    被覆盖的是「已提交」的日志吗
…（截断，原长 1761 字符）
````

</details>

▸ **Bash** Update README and final verification

````
cd 〔路径00054〕 && cp ch03b/ch03b.html courseware/ch03b-log-replication.html && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('''open courseware/ch03a-consensus-election.html  # Part 3-A：共识与 Raft 选举
# Linux 用 xdg-open''','''open courseware/ch03a-consensus-election.html  # Part 3-A：共识与 Raft 选举
open courseware/ch03b-log-replication.html    # Part 3-B：日志复制与安全性
# Linux 用 xdg-open''')
rep('cd go/lab03a && go run .     # Lab 3-A：确定性网络 + Raft 选举 + 安全性断言',
    'cd go/lab03a && go run .     # Lab 3-A：确定性网络 + Raft 选举 + 安全性断言\ncd go/lab03b && go run .     # Lab 3-B：日志复制 + Figure 8 复现')
rep('cd java/lab03a && java Lab03A.java','cd java/lab03a && java Lab03A.java\ncd java/lab03b && java Lab03B.java')
rep('|
…（截断，原长 2797 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README ok
naive=true => — || — || 你关掉了限制 —— 这里 Leader 会宣布 index2 已提交 || ★ 安全性破产：已经告诉客户端"成功"的数据凭空消失了 || 安全的那条路
exp3: 朴素回退=41 | 快速回退=4 | 加速比=10.3× | 节省时间=2960
exp3 gap=200: 201 | 4 | 50.3× | 15760
ERRORS: none
````

</details>

📎 **发送文件**：Part 3-B 交互式课件。核心是「实验 2 · Figure 8 交互重放」——五个阶段一格一格走，上面有个开关「关掉限制」。开关关着时你看到覆盖发生但无害；打开后同样这一步就是安全性破产。
- `〔路径00075〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00075〕 → file_uuid: 〔标识03097〕
````

</details>

📎 **发送文件**：Lab 3-B 代码。raft.go 的 maybeCommit() 里那一行 `if !c.NaiveCommit && n.log[N].Term != n.term { continue }` 就是 Figure 8 的全部修补。
- `〔路径00056〕`
- `〔路径00106〕`
- `〔路径00107〕`
- `〔路径00105〕`
- `〔路径00135〕`

<details><summary>↳ SendUserFile 结果</summary>

````
5 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识00006〕
  〔路径00106〕 → file_uuid: 〔标识02654〕
  〔路径00107〕 → file_uuid: 〔标识02022〕
  〔路径00105〕 → file_uuid: 〔标识02833〕
  〔路径00135〕 → file_uuid: 〔标识00284〕
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00155〕\" \"$HOME〔路径00163〕\" && echo ready"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ready
````

</details>

💾 **写入你的 Mac**：
- `〔路径00009〕`
- `〔路径00004〕`
- `〔路径00031〕`
- `〔路径00032〕`
- `〔路径00030〕`
- `〔路径00048〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00009〕","〔路径00004〕","〔路径00031〕","〔路径00032〕","〔路径00030〕","〔路径00048〕"],"rejected":[]}
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #15 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #16 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #17 status
````

</details>

Part 3-B 交付完成，文件已写进 `~/Code/distributed-systems-course/`。

**这一节我最满意的是它把"抽象的安全性论证"变成了两个能跑的数字。**

课件里的 Figure 8 重放有一个开关。关着的时候，你看到 S5 覆盖掉 S2/S3 上那条 term2 日志——但没造成任何伤害，因为 Raft 从没把它叫做"已提交"。打开开关，**同样这一步就变成了事故**：客户端收到过"成功"的数据凭空消失。

Lab 3B-3 用代码精确复现了同一件事，两边的物理事实完全相同：

| | 完整 Raft | 关掉限制 |
|---|---|---|
| term2 那条复制到过半了吗 | 是 (3/5) | 是 (3/5) |
| 被标记为已提交了吗 | **否** | **是** |
| 被 S5 覆盖了吗 | 是 | 是 |
| 安全性断言 | ✓ 成立 | **✗ 被破坏** |

`-naivecommit` 一开，断言当场报出：

```
【Leader 完整性被破坏】N5 在 term=5 当选，但它缺少已提交的 index=2 (term=2, cmd=X)
```

**修补只有一行**——`raft.go` 的 `maybeCommit()` 里：

```go
if !c.NaiveCommit && n.log[N].Term != n.term { continue }
```

Raft 的高明之处在于它<u>没有</u>去阻止覆盖（那要改选举规则，代价大得多），而是**确保被覆盖的东西从来没被承诺过**。这是一个非常克制的设计。

另外 Lab 3B-4 的数字挺震撼：朴素回退（每次 −1）需要 163 次被拒、16.3 秒才追平；快速回退 **1 次、0.13 秒**。121 倍。落后节点追不上就等于集群实际少了一个副本——**容错能力会悄悄下降而没人察觉**。

---

**3-C 之前，三个问题：**

1. **回到你上一节答对的那件事**：僵尸 Leader 用本地读会返回陈旧数据。现在你知道了 commitIndex 的存在——**为什么僵尸 Leader 的本地读会陈旧？它的 commitIndex 明明是它自己维护的、看起来一直有效。**

2. **Raft 论文说"日志一旦持久化就不会丢"。那到底哪些状态必须落盘？** 试着列出来，并对每一条说明"如果它没落盘，重启后会破坏哪条安全性属性"。

3. **想一想 3-C 的核心难题**：集群要从 3 个节点扩到 5 个节点，不能停服。**如果直接把配置从 {A,B,C} 改成 {A,B,C,D,E}，但各节点是在不同时刻切换配置的**，会出什么事？

第 3 题是成员变更的入口。提示：想想在切换的那个瞬间，"过半"这个词对不同节点意味着什么。

> ✓ 本轮结束 · 2026-09-05 20:30 · 28 步
