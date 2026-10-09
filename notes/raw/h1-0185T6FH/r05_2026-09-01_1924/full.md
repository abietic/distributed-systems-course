## 👤 你 · 2026-09-01 19:24

1. 先不用管我机器上的服务，先着重在分布式系统原理和流行程序和中间件；应该是单实例，在更新过程中确实会有由于提交时间导致的cronjob扫描不全的问题，这个通过增加一个时间跨度更长周期更长的cronjob做sweep去兜底
2. 拆开来做吧

▸ **ToolSearch**

````
{"max_results": 4, "query": "create_artifact list_artifacts update_artifact"}
````

<details><summary>↳ ToolSearch 结果</summary>

````
No matching deferred tools found
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00153〕\" \"$HOME〔路径00161〕\" && echo ready && ls \"$HOME〔路径00146〕\""}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ready
ch00-intro.html
ch01-time-order.html
````

</details>

## 🤖 Claude · 2026-09-01 19:26

💾 **写入你的 Mac**：
- `〔路径00007〕`
- `〔路径00004〕`
- `〔路径00026〕`
- `〔路径00023〕`
- `〔路径00024〕`
- `〔路径00025〕`
- `〔路径00046〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00007〕","〔路径00004〕","〔路径00026〕","〔路径00023〕","〔路径00024〕","〔路径00025〕","〔路径00046〕"],"rejected":[]}
````

</details>

机器连上了，Part 2 的全部文件已经写进 `~/Code/distributed-systems-course/`。

**先更正我之前的一个说法。** 我在前两轮说过"课件同时存成了桌面 artifact（`ds-course-ch00` / `ds-course-ch01`）"——**这个说法是错的，我没有真的发布过 artifact**。这个会话里根本没有 artifact 发布工具，我当时不该那么说。实际情况是：课件以 HTML 文件的形式存在两个地方——你本地的 `courseware/` 目录，和对话里的文件卡片。双击本地文件就能在浏览器打开，交互全都正常。抱歉给了你一个不存在的东西。

---

**你的答案 1，补一句然后就放下你们的系统。**

`ON UPDATE` 由主库打戳、单点——对。而你说的"再加一个周期更长、时间跨度更长的 sweep cronjob 兜底"，正是我列的三种解法里的第二种（**安全回溯窗口 + 幂等消费**）。这个方案是对的，而且在工程上通常是最划算的，因为它不用改 schema、不用上 CDC。

只提醒一个它<u>关不住</u>的口子：**sweep 的回溯窗口必须大于你最长事务的持续时间**。如果哪天有个批量任务开了个 20 分钟的事务，而 sweep 只回看 10 分钟，那条记录还是会永久漏掉。所以这类兜底最好配一条监控：定期做一次全量对账（`count(*)` 或校验和比对），把"兜底没兜住"这件事变成可观测的。**兜底方案本身也需要兜底验证，否则你不知道它有没有在工作。**

好，你们的系统到此为止，后面回到原理和工业系统。

---

**Part 3 拆成三次，我按这个切：**

| | 主题 | Lab |
|---|---|---|
| **3-A** | 共识问题的定义、FLP、为什么 2PC 不是共识、**Raft 选举** | 实现选举 + 安全性断言 |
| **3-B** | **日志复制**、五条安全性属性、日志匹配与冲突回退 | 实现日志复制 + 一致性检查 |
| **3-C** | 持久化、快照、成员变更、**读的线性一致性**（ReadIndex / Lease Read） | 补全并跑完整测试 |

现在做 3-A。

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #11 status
````

</details>

▸ **TaskCreate** 编写 Part 3-A 课件：共识与 Raft 选举

<details><summary>↳ TaskCreate 结果</summary>

````
Task #12 created successfully: 编写 Part 3-A 课件：共识与 Raft 选举
````

</details>

▸ **TaskCreate** 编写 Lab 3-A：Raft 选举实现（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #13 created successfully: 编写 Lab 3-A：Raft 选举实现（Go + Java）
````

</details>

▸ **TaskCreate** 验证并交付 Part 3-A

<details><summary>↳ TaskCreate 结果</summary>

````
Task #14 created successfully: 验证并交付 Part 3-A
````

</details>

▸ **Bash** Write Part 3-A head and sections 3.0-3.2

````
cd 〔路径00054〕 && mkdir -p ch03a && cd ch03a && cat > ch03a.html <<'HEAD_EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 3-A · 共识与 Raft 选举</title>
HEAD_EOF
cat 〔路径00240〕 >> ch03a.html
cat >> ch03a.html <<'HEAD2_EOF'
<style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.three{display:grid;grid-template-columns:repeat(3,1fr);gap:14px}
@media(max-width:900px){.three{grid-template-columns:1fr}}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  backgr
…（截断，原长 7846 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-A sections 3.3-3.8

````
cat >> ch03a.html <<'EOF'
<section id="s3">
  <h2>3.3 两阶段提交为什么不是共识</h2>
  <p>2PC（two-phase commit）是分布式事务的经典方案，很多人以为它就是"分布式共识"。<strong>它不是</strong>，而且区别很致命。</p>
  <p>流程只有两步：</p>
  <div class="axm"><div class="b">1</div><div><strong>准备阶段</strong>：协调者问所有参与者"能提交吗？"参与者做完全部工作、写好 redo/undo 日志、<strong>持有锁不放</strong>，然后回答 yes 或 no。回答 yes 之后，它就<em>失去了单方面中止的权利</em>——必须等协调者的最终指令。</div></div>
  <div class="axm"><div class="b">2</div><div><strong>提交阶段</strong>：协调者收齐全部 yes 就广播 commit，否则广播 abort。参与者照做，释放锁。</div></div>

  <div class="note crit">
    <div class="note-h"><span class="dot" style="background:var(--crit)"></span>致命缺陷：协调者在两个阶段之间崩溃</div>
    <p style="margin-bottom:0">参与者已经回答了 yes、锁已经拿着、事务既不能提交也不能回滚——因为它不知道其他参与者答了什么，也不知道协调者最终决定了什么。<br>
    它只能<strong>无限期等待协调者恢复</strong>。这段时间里，它持有的行锁会把所有访问这些行的事务全部
…（截断，原长 13125 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-A quiz and lab sections

````
cat >> ch03a.html <<'EOF'
<section id="s9">
  <h2>3.9 自测题</h2>
  <p>先自己想 30 秒再展开。</p>

  <details class="q"><summary>1. FLP 不可能性到底否定了什么？为什么 Raft 没有违反它？</summary>
    <div class="a"><p>FLP 否定的是：在<strong>完全异步</strong>、允许<strong>一个节点崩溃</strong>的系统里，存在一个<strong>确定性</strong>算法能<strong>同时</strong>保证安全性和一定终止。</p>
    <p style="margin-bottom:0">Raft 没违反它，因为它<strong>放弃了"一定终止"</strong>：网络持续异常时 Raft 可以永远选不出 leader（活性没了），但绝不会出现两个 leader（安全性还在）。同时它用<strong>随机化选举超时</strong>把分裂投票的活锁变成"以概率 1 终止"——随机化本身就在 FLP 的确定性假设之外。<em>FLP 不是死刑判决，是一份必须做出取舍的清单。</em></p></div></details>

  <details class="q"><summary>2. 为什么 2PC 不是共识算法？两者解决的问题有什么不同？</summary>
    <div class="a"><p><strong>2PC 要求全体一致，共识只要求过半。</strong>这一条差别导致：2PC 的可用性是所有参与者可用性的乘积（节点越多越脆弱），共识的可用性随节点增多而提升。而且 2PC 的协调者崩溃后没有内建接替机制，参与者持锁无限阻塞。</p>
    <p style="mar
…（截断，原长 7690 字符）
````

<details><summary>↳ Bash 结果</summary>

````
sections: 11 kb: 53
````

</details>

▸ **Bash** Write Part 3-A JS: 2PC blocking experiment

````
cat >> ch03a.html <<'EOF'
<script>
(function(){
"use strict";
const $=s=>document.querySelector(s), $$=s=>Array.from(document.querySelectorAll(s));
const NS="http://www.w3.org/2000/svg";
function el(t,a,x){const e=document.createElementNS(NS,t);for(const k in a)e.setAttribute(k,a[k]);if(x!=null)e.textContent=x;return e;}
function cv(n){return getComputedStyle(document.body).getPropertyValue(n).trim();}

$("#tbtn").onclick=()=>{const c=document.documentElement.getAttribute("data-theme");
  document.documentElement.setAttribute("data-theme",c==="dark"?"light":"dark");draw1();draw3();};
const secs=$$("main section"),links=$$("#toc a:not(.sub)");
function onScroll(){const h=document.documentElement;
  $("#prog").style.width=(h.scrollTop/(h.scrollHeight-h.clientHeight)*100)+"%";
  let cur=secs[
…（截断，原长 7787 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Raft election simulator and randomization experiment

````
cat >> ch03a.html <<'EOF'

/* ══════════════════════════════════════════════════════════════════
   实验 2 · Raft 选举模拟器（真的在跑状态机 + 消息传递）
   ══════════════════════════════════════════════════════════════════ */
const NN=5, MAJ=Math.floor(NN/2)+1, HB=150;
const R={nodes:[],msgs:[],run:false,last:0,part:false,
  elections:0,splits:0,maxLeadersPerTerm:1,violated:false,log:[]};
const P2X=[], P2Y=[];
(function layout(){
  const cx=360, cy=150, rad=112;
  for(let i=0;i<NN;i++){ const a=-Math.PI/2+i*2*Math.PI/NN;
    P2X.push(cx+rad*Math.cos(a)); P2Y.push(cy+rad*Math.sin(a)); }
})();
function cfg2(){return {delay:+$("#r2d").value, loss:+$("#r2l").value/100,
  base:+$("#r2t").value, rand:$("#c2r").checked};}
function rto(){ const c=cfg2(); return c.base + (c.rand? Math.random()*400 : 0); }
function lo
…（截断，原长 13464 字符）
````

<details><summary>↳ Bash 结果</summary>

````
script: 1 1 kb: 78
````

</details>

▸ **Bash** Verify Part 3-A interactives

````
cd /tmp && cat > c3.mjs <<'EOF'
import { chromium } from 'playwright';
const b=await chromium.launch({executablePath:'〔路径00178〕'});
const p=await b.newPage({viewport:{width:1300,height:1000}});
const errs=[];p.on('pageerror',e=>errs.push('PAGEERROR: '+e.message));
p.on('console',m=>{if(m.type()==='error')errs.push('CONSOLE: '+m.text());});
await p.goto('file://〔路径00065〕');
await p.addStyleTag({content:'html{scroll-behavior:auto !important}'});
await p.waitForTimeout(500);
if(errs.length){console.log('LOAD ERRORS:',errs);process.exit(1);}

// 实验 1：走完四个场景
for(let s=0;s<4;s++){
  await p.click(`#sc1 button[data-s="${s}"]`);
  for(let i=0;i<5;i++) await p.click('#b1step');
  const t=await p.evaluate(()=>document.querySelector('#o
…（截断，原长 2672 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
LOAD ERRORS: [ "PAGEERROR: Unexpected identifier '我俩都投了'" ]

[agent-proxy] While this command ran, 3 connections through the agent proxy failed:
- www.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×2
- accounts.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination)
For details: curl -sS http://127.0.0.1:38513/__agentproxy/status
````

</details>

▸ **Bash** Fix broken JS string and re-verify

````
python3 - <<'PY'
p='ch03a.html'; s=open(p,encoding='utf-8').read()
bad='它们只知道"我俩都投了 yes、都没收到指令"，这和"协调者决定 abort"的情形'
good='它们只知道「我俩都投了 yes、都没收到指令」，这和「协调者决定 abort」的情形'
assert bad in s, "pattern not found"
s=s.replace(bad,good)
open(p,'w',encoding='utf-8').write(s); print('fixed')
PY
# 扫描 JS 区段里所有可能破坏字符串的裸 ASCII 双引号
python3 - <<'PY'
s=open('ch03a.html',encoding='utf-8').read()
js=s[s.index('<script>'):]
import re
bad=[]
for i,line in enumerate(js.split('\n'),1):
    # 统计行内 ASCII 双引号数量为奇数且含中文 → 可疑
    if line.count('"')%2==1 and re.search(r'[一-鿿]',line):
        bad.append((i,line.strip()[:90]))
print("suspicious lines:",len(bad))
for x in bad[:10]: print(x)
PY
node 〔路径00184〕
````

<details><summary>↳ Bash 结果</summary>

````
fixed
suspicious lines: 0
2PC scen 0 => 事务完成
2PC scen 1 => 事务完成
2PC scen 2 => 永久阻塞
2PC scen 3 => 状态分歧 + 永久阻塞
after 2.5s: {"当前 Leader":"N2","最大 term":"1","发起过的选举":"2","分裂投票":"0","同一 term 的 Leader 数":"1"}
after kill leader: {"当前 Leader":"N3","最大 term":"3","发起过的选举":"6","分裂投票":"0","同一 term 的 Leader 数":"1"}
after partition: {"当前 Leader":"N3","最大 term":"8","发起过的选举":"11","分裂投票":"2","同一 term 的 Leader 数":"1"}
after heal: {"当前 Leader":"N5","最大 term":"9","发起过的选举":"12","分裂投票":"2","同一 term 的 Leader 数":"1"}
after 40% loss: {"当前 Leader":"N1","最大 term":"11","发起过的选举":"15","分裂投票":"3","同一 term 的 Leader 数":"1"}
no-random reset: {"当前 Leader":"无","最大 term":"6","发起过的选举":"30","分裂投票":"25","同一 term 的 Leader 数":"1"}
log tail: N4 选举超时仍未过半 → 分裂投票，重新发起 | N4 选举超时 → 成为 Candidate，term=6 | N5 选举超时仍未过半 → 分裂投票，重新发起 | N5 选举超时 → 成为 Candidate，term=6
exp3: 固定超时 · 平均轮数=61.00 | 随机超时 · 平均轮数=1.06 | 耗时改善=55.1× | 固定超时分裂率=100%
exp3 jitter=0: 61.00 | 1.05 | 55.9× | 100%
ERRORS: none
````

</details>

▸ **Bash** Fix experiment 3 messaging and re-verify

````
python3 - <<'PY'
p='ch03a.html'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)

rep('''      Math.round(row[1].time)+" ms　·　平均 "+row[1].rounds.toFixed(2)+" 轮"));''',
    '''      Math.round(row[1].time)+" ms　·　平均 "+(row[1].rounds>=60?"60+ 轮（不收敛）":row[1].rounds.toFixed(2)+" 轮")));''')
rep('''    "选出 Leader 的平均耗时（越短越好）· 每组 2000 次模拟"));''',
    '''    "选出 Leader 的平均耗时（越短越好）· 每组 2000 次模拟 · 单次 RequestVote 飞行时间取 30ms"));''')
rep('''    '<div class="stat"><div class="k">固定超时 · 平均轮数</div><div class="v" style="color:var(--crit)">'+fixed.rounds.toFixed(2)+'</div><div class="u">轮才选出 Leader</div></div>'+''',
    '''    '<div class="stat"><div class="k">固定超时 · 平均轮数</div><div class="v" style="color:var(--crit)">'+(fixed.rounds>=60?"60+":fixed.round
…（截断，原长 3296 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
jitter 0 => 60+ | 1.07 | 54.9× | 100%
jitter 8 => 60+ | 1.06 | 55.2× | 100%
jitter 60 => 1.12 | 1.04 | 1.0× | 11%
jitter 100 => 1.03 | 1.02 | 0.9× | 3%
errors: none
Shell cwd was reset to 〔路径00052〕
````

</details>

▸ **Read** `〔路径00220〕`

<details><summary>↳ Read 结果</summary>

````
[image]
````

</details>

▸ **Bash** Fix legend position and write Go network simulator

````
cd 〔路径00064〕 && python3 - <<'PY'
p='ch03a.html'; s=open(p,encoding='utf-8').read()
s=s.replace('<svg id="sv2" viewBox="0 0 720 300"','<svg id="sv2" viewBox="0 0 720 320"')
s=s.replace('''  svg.appendChild(el("text",{x:14,y:18,fill:cv("--ink-3"),"font-size":10.5,"font-family":cv("--sans")},
    "蓝点 = RequestVote　金点 = 投票回复　绿点 = 心跳　外环 = 选举计时器"));''',
'''  svg.appendChild(el("text",{x:14,y:312,fill:cv("--ink-3"),"font-size":10.5,"font-family":cv("--sans")},
    "蓝点 = RequestVote　金点 = 投票回复　绿点 = 心跳　节点外环 = 选举计时器进度"));''')
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
mkdir -p 〔路径00100〕 && cat > 〔路径00102〕 <<'EOF'
package main

import (
	"math/rand"
	"sort"
)

// ════════════════════════════════════════════════════════════════════════
//
…（截断，原长 3566 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
net ok
````

</details>

▸ **Bash** Write Go Raft election state machine

````
cat > 〔路径00103〕 <<'EOF'
package main

import (
	"fmt"
	"math/rand"
)

// ════════════════════════════════════════════════════════════════════════
// Raft 选举状态机
//
//	这一节只做三件事：任期递增、投票、心跳压制。
//	日志复制与安全性属性留给 Lab 3-B。
//	核心逻辑不到 150 行 —— Raft 之所以出名，正是因为它能这么短。
//
// ════════════════════════════════════════════════════════════════════════

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

func (s State) String() string { return [...]string{"Follower", "Candidate", "Leader"}[s] }

const (
	HeartbeatMs = 100 // Leader 发心跳的间隔
	NoVote      = -1
)

type Node struct {
	id       int
	state    State
	term     int
	votedFor int          // 本任期把票投给了谁；NoVote 表示还没投
	votes    map[int]bool // Candidate 收到的选票
	timer    int64        // 选举计时器已走过的毫秒
	timeout  int64
…（截断，原长 6724 字符）
````

<details><summary>↳ Bash 结果</summary>

````
raft ok
````

</details>

▸ **Bash** Write and run Go Lab 3-A

````
cat > 〔路径00101〕 <<'EOF'
// Lab 3-A · 共识与 Raft 选举
//
// 运行：  cd go/lab03a && go run .
// 调参：  go run . -loss 0.3 -delay 80 -norandom
//         go run . -unsafe        ← 去掉「每任期一票」，看安全性怎么炸
//
// 配套课件：courseware/ch03a-consensus-election.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"dsc/internal/tui"
)

var (
	nF      = flag.Int("n", 5, "节点数（取奇数）")
	delayF  = flag.Int64("delay", 30, "单向网络延迟（毫秒）")
	jitterF = flag.Int64("jitter", 15, "延迟抖动（毫秒）")
	lossF   = flag.Float64("loss", 0.0, "丢包率 0~1")
	baseF   = flag.Int64("base", 400, "选举超时下限（毫秒）")
	winF    = flag.Int64("window", 300, "选举超时随机区间宽度（毫秒）")
	seedF   = flag.Int64("seed", 42, "随机种子，同种子完全可复现")
	noRandF = flag.Bool("norandom", false, "关闭随机化选举超时（制造分裂投票）")
	unsafeF = flag.Bool("unsafe", false, "★ 去掉「每个任期只投一票」规则，
…（截断，原长 8515 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：5 节点 ｜ 延迟 30±15ms ｜ 丢包 0% ｜ 选举超时 400~700ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3-1 · 确定性网络模拟
  调共识算法的第一件事：让 bug 能复现
══════════════════════════════════════════════════════════════════════════════
    种子 42 · 第一次               Leader=N4(t1) maxTerm=1 选举=1 分裂=0
    种子 42 · 第二次               Leader=N4(t1) maxTerm=1 选举=1 分裂=0
    种子 7                         Leader=N1(t1) maxTerm=1 选举=1 分裂=0
    同种子结果一致                 ✓ 完全可复现

    ▸ 没有 goroutine、没有真实时间，只有一个虚拟时钟和一个按时间排序的消息队列。
      这样任何一次诡异的执行都能靠种子复现，而不是"跑一百遍偶尔挂一次"。
    ▸ 这是 MIT 6.5840 的 labrpc、以及 etcd 的 raft 测试框架采用的同一套思路：
      把并发与时间从被测逻辑里彻底剥离出去。

══════════════════════════════════════════════════════════════════════════════
  Lab 3-2 · 五个场景
  每个 tick 都断言「任一任期至多一个 Leader」
══════════════════════════════════════════════════════════════════════════════
    场景                        结果 Leader       最大term    选举数      分裂选举安全性
    ──────────────────────────────────────────────────────────────────────────────────────────────
    ① 冷启动                    N4(t1)                   1         1         0✓ 成立
    ② 杀死 Leader 后重选        N1(t2)                   2         3         0✓ 成立
    ③ 3|2 分区                  多数派1 少数派1          2         2         0✓ 成立
    ④ 分区恢复后收敛            N1(t2)                   2         2         0✓ 成立
    ⑤ 反复随机杀/救节点         N4(t1)                   1         1         0✓ 成立

    ③ 多数派侧的 Leader            1 个　← 应当是 1
    ③ 少数派侧的 Leader            1
…（截断，原长 2161 字符）
````

</details>

▸ **Bash** Fix table alignment and zombie-leader explanation

````
cd 〔路径00079〕 && cat >> internal/tui/tui.go <<'EOF'

// HeadN 与 Head 相同，但实验编号由调用方自己给（比如 "3A-1"）。
func HeadN(label, title, subtitle string) {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Printf("  Lab %s · %s\n  %s\n", label, title, subtitle)
	fmt.Printf("%s\n", strings.Repeat("═", 78))
}
EOF
python3 - <<'PY'
p='lab03a/main.go'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)
rep('tui.Head(3, 1, "确定性网络模拟"','tui.HeadN("3A-1", "确定性网络模拟"')
rep('tui.Head(3, 2, "五个场景"','tui.HeadN("3A-2", "五个场景"')
rep('tui.Head(3, 3, "随机化选举超时"','tui.HeadN("3A-3", "随机化选举超时"')
rep('tui.Head(3, 4, "把安全性拆掉给你看"','tui.HeadN("3A-4", "把安全性拆掉给你看"')
# 表格宽度与对齐
rep('''	w := []int{28, 16, 10, 10, 10, 20}
	al := "LLRRRL"
	tui.TableHeadA([]string{"场景", "结果 Le
…（截断，原长 2999 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
  Lab 3A-2 · 五个场景
  每个 tick 都断言「任一任期至多一个 Leader」
══════════════════════════════════════════════════════════════════════════════
    场景                        结果 Leader           最大term   选举数    分裂      选举安全性
    ───────────────────────────────────────────────────────────────────────────────────────────
    ① 冷启动                    N4(t1)                       1        1       0          ✓ 成立
    ② 杀死 Leader 后重选        N1(t2)                       2        3       0          ✓ 成立
    ③ 3|2 分区                  多数派1 少数派1              2        2       0          ✓ 成立
    ④ 分区恢复后收敛            N1(t2)                       2        2       0          ✓ 成立
    ⑤ 反复随机杀/救节点         N4(t1)                       1        1       0          ✓ 成立

    ③ 多数派侧的 Leader            1 个（能提交日志）
    ③ 少数派侧的 Leader            1 个 —— ★ 僵尸 Leader：自认为是主，但一条日志都提交不了
    ③ 少数派能联系到几个节点       2 个，过半需要 3 个 ⇒ 提交不了任何东西
    ④ 恢复后集群的任期数           1 种　← 应当收敛到 1~2 种
    ④ 最终状态                     N1:L(t2) N2:F(t2) N3:F(t2) N4:F(t2) N5:F(t2)

    ▸ 场景③是 CAP 的具体样子，但它给出的答案比教科书更微妙：
      如果分区前的 Leader 恰好落在少数派一侧，基础 Raft 不会让它主动退位 ——
      它继续给同侧的 Follower 发心跳，自认为还是 Leader。这叫「僵尸 Leader」。
      它不违反安全性（拿不到过半确认，一条日志都提交不了），但客户端把写请求
      发给它会一直超时，读请求还可能读到陈旧数据。
    ▸ 工程上用 CheckQuorum 解决：Leader 定期确认自己还能联系到过半节点，
      否则主动退位。etcd 默认开启它。这也是「用了 Raft 就一定线性一致」
    ▸ 场景④注意少数派带回来的高任期：它一接触集群就会逼现任 Leader 退位，
      引发一次不必要的选举。这正是 Raft 论文用 PreVote 优化解决的问
…（截断，原长 2442 字符）
````

</details>

▸ **Bash** Check Lab 3A-4 safety demolition output

````
go run . 2>&1 | sed -n '66,88p'; echo "=== unsafe -v 片段 ==="; go run . -unsafe 2>&1 | sed -n '70,80p'
````

<details><summary>↳ Bash 结果</summary>

````
  Lab 3A-4 · 把安全性拆掉给你看
  去掉「每个任期只投一票」这一条，其他不变
══════════════════════════════════════════════════════════════════════════════
    规则                            同任期最多 Leader 数    违反次数      选举安全性
    ────────────────────────────────────────────────────────────────────────────────
    ① 完整 Raft（每任期一票）                          0           0          ✓ 成立
    ② 去掉「每任期一票」                               4          55        ✗ 被破坏

    ▸ ① 无论怎么折腾都是 1 —— 因为两个过半集合必然相交，
      交集里那个节点在同一任期投了两票，这是不可能的。
    ▸ ② 一旦允许一个节点在同一任期投多票，两个 Candidate 就能各自凑齐过半，
      同一任期出现两个 Leader —— 这就是脑裂，两个 Leader 会各写各的日志，
      数据从此分叉且无法自动合并。
    ▸ 所以「每任期只投一票」不是实现细节，它和「过半」一起构成了
      Raft 全部安全性的地基。而且 votedFor 必须持久化 ——
      节点重启后如果忘了自己投过票，同样会破坏它（Lab 3-C 会处理）。
      试试 go run . -unsafe -v 看具体是怎么炸的。

══════════════════════════════════════════════════════════════════════════════
  三个思考题（答案在课件 §3.6 / §3.7）：

    1. 把 -loss 调到 0.5，集群还能选出 Leader 吗？需要多久？
       再调到 0.8 呢？在什么丢包率下它彻底选不出来？
=== unsafe -v 片段 ===
    ────────────────────────────────────────────────────────────────────────────────
    ① 完整 Raft（每任期一票）                          0           0          ✓ 成立
    ② 去掉「每任期一票」                               4          55        ✗ 被破坏

    ▸ ① 无论怎么折腾都是 1 —— 因为两个过半集合必然相交，
      交集里那个节点在同一任期投了两票，这是不可能的。
    ▸ ② 一旦允许一个节点在同一任期投多票，两个 Candidate 就能各自凑齐过半，
      同一任期出现两个 Leader —— 这就是脑裂，两个 Leader 会各写各的日志，
      数据从此分叉且无法自动合并。
    ▸ 所以「每任期只投一票」不是实现细节，它和「过半」一起构成了
      Raft 全部安全性的地基。而且 votedFor 必须持久化 ——
````

</details>

▸ **Bash** Tune safety demolition experiment

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
old='''		worst, violations := 0, 0
		for seed := int64(1); seed <= 60; seed++ { // 多跑几个种子，撞上并发选举的概率才够高
			net := NewNetwork(*nF, *delayF, 40, 0.15, seed)
			c := New(*nF, net, 300, 0, seed, unsafe) // 固定超时 ⇒ 制造同时选举
			run(c, 6000)
			if c.MaxLeadersPerTerm > worst {
				worst = c.MaxLeadersPerTerm
			}
			if len(c.Violations) > 0 {
				violations++
			}
		}
		tui.TableRowA([]string{label, fmt.Sprint(worst), fmt.Sprint(violations),
			map[bool]string{true: "✗ 被破坏", false: "✓ 成立"}[violations > 0]}, w, al)
		return'''
new='''		worst, violations, elected := 0, 0, 0
		const trials = 60
		for seed := int64(1); seed <= trials; seed++ { // 多跑几个种子，撞上并发选举的概率才够高
			// 窗口很窄 + 有丢包 ⇒ 经常出现两个 Candidate 同时索票
			net := NewNetwork(*nF, *delayF,
…（截断，原长 1851 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok

══════════════════════════════════════════════════════════════════════════════
  Lab 3A-4 · 把安全性拆掉给你看
  去掉「每个任期只投一票」这一条，其他不变
══════════════════════════════════════════════════════════════════════════════
    规则                          同任期最多 Leader 数        违反的种子   选出了 Leader    选举安全性
    ──────────────────────────────────────────────────────────────────────────────────────────────────
    ① 完整 Raft（每任期一票）                     1 个       0/60 个种子           56/60        ✓ 成立
    ② 去掉「每任期一票」                          3 个      37/60 个种子           58/60      ✗ 被破坏

    ▸ ① 60 个种子、上万个 tick，同一任期的 Leader 数最多就是 1 —— 因为两个过半集合必然相交，
````

</details>

▸ **Bash** Write and run Java Lab 3-A

````
mkdir -p 〔路径00132〕 && cat > 〔路径00133〕 <<'EOF'
/*
 * Lab 3-A · 共识与 Raft 选举（Java 版）
 *
 * 运行：  cd java/lab03a && java Lab03A.java
 * 调参：  java Lab03A.java -loss 0.3 -delay 80 -norandom
 *         java Lab03A.java -unsafe      ← 去掉「每任期一票」，看安全性怎么炸
 *
 * 需要 JDK 17+（单文件源码启动模式）。
 * 配套课件：courseware/ch03a-consensus-election.html
 */
import java.util.*;

public class Lab03A {

    static int    N = 5;
    static long   DELAY = 30, JITTER = 15, BASE = 400, WINDOW = 300, SEED = 42;
    static double LOSS = 0.0;
    static boolean NORANDOM = false, UNSAFE = false, VERBOSE = false;
    static final long TICK = 5, HEARTBEAT = 100;
    static final int NO_VOTE = -1;

    // ══════════════════════════════════════════════════════════════════════
    // 确定性网络
…（截断，原长 21300 字符）
````

<details><summary>↳ Bash 结果</summary>

````
配置：5 节点 ｜ 延迟 30±15ms ｜ 丢包 0% ｜ 选举超时 400~700ms ｜ 种子 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3A-1 · 确定性网络模拟
  调共识算法的第一件事：让 bug 能复现
══════════════════════════════════════════════════════════════════════════════
    种子 42 · 第一次               Leader=N3(t1) maxTerm=1 选举=1 分裂=0
    种子 42 · 第二次               Leader=N3(t1) maxTerm=1 选举=1 分裂=0
    种子 7                         Leader=N2(t1) maxTerm=1 选举=1 分裂=0
    同种子结果一致                 ✓ 完全可复现

▸ 没有线程、没有真实时间，只有一个虚拟时钟和一个按时间排序的消息队列。
  这样任何一次诡异的执行都能靠种子复现，而不是"跑一百遍偶尔挂一次"。
  Lab 3A-2 · 五个场景
  每个 tick 都断言「任一任期至多一个 Leader」
══════════════════════════════════════════════════════════════════════════════
    场景                        结果 Leader           最大term   选举数    分裂      选举安全性
    ───────────────────────────────────────────────────────────────────────────────────────────
    ① 冷启动                    N3(t1)                       1        1       0          ✓ 成立
    ② 杀死 Leader 后重选        N5(t3)                       3        4       0          ✓ 成立
    ③ 3|2 分区                  多数派1 少数派0             12       13       8          ✓ 成立
    ④ 分区恢复后收敛            N5(t13)                     13       14       9          ✓ 成立
    ⑤ 反复随机杀/救节点         N2(t2)                       2        3       0          ✓ 成立

    ③ 多数派侧的 Leader            1 个（能提交日志）
    ③ 少数派侧的 Leader            0 个 —— 已退位
    ③ 少数派能联系到几个节点       2 个，过半需要 3 个 ⇒ 提交不了任何东西
    ④ 恢复后集群的任期数           1 种　← 应当收敛到 1~2 种
  超时后再来一轮 —— 任期一路飙升
…（截断，原长 2063 字符）
````

</details>

▸ **Bash** Restore text-block indentation in Java labs

````
cd 〔路径00121〕 && sed -i 's/"""\.indent(0));/""".indent(4));/g' lab02/Lab02.java lab03a/Lab03A.java && (cd lab03a && java Lab03A.java 2>&1 | grep -v "Picked up" | sed -n '11,14p') && (cd lab02 && java Lab02.java 2>&1 | grep -v "Picked up" | sed -n '17,19p')
````

<details><summary>↳ Bash 结果</summary>

````
    同种子结果一致                 ✓ 完全可复现

    ▸ 没有线程、没有真实时间，只有一个虚拟时钟和一个按时间排序的消息队列。
      这样任何一次诡异的执行都能靠种子复现，而不是"跑一百遍偶尔挂一次"。
    W+R≤N 的陈旧读                 34.0%　← 交集可能为空，读不到最新

    ▸ 前三行的 W+R 都 > N，陈旧读恒为 0 —— 交集性质是一条数学保证，不是概率。
````

</details>

▸ **Bash** Update README and final courseware verification

````
cd 〔路径00054〕 && cp ch03a/ch03a.html courseware/ch03a-consensus-election.html && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('''open courseware/ch02-replication.html   # Part 2：复制与一致性模型
# Linux 用 xdg-open''','''open courseware/ch02-replication.html   # Part 2：复制与一致性模型
open courseware/ch03a-consensus-election.html  # Part 3-A：共识与 Raft 选举
# Linux 用 xdg-open''')
rep('cd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT',
    'cd go/lab02 && go run .      # Lab 2：Quorum / CAP / 一致性判定器 / CRDT\ncd go/lab03a && go run .     # Lab 3-A：确定性网络 + Raft 选举 + 安全性断言')
rep('cd java/lab02 && java Lab02.java','cd java/lab02 && java Lab02.java\ncd java/lab03a && java Lab03A.java')
rep('| 3 | 共识算法 —
…（截断，原长 3103 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README ok
after partition: {"当前 Leader":"N3","最大 term":"6","发起过的选举":"7","分裂投票":"3","同一 term 的 Leader 数":"1"}
after heal: {"当前 Leader":"N5","最大 term":"7","发起过的选举":"8","分裂投票":"4","同一 term 的 Leader 数":"1"}
after 40% loss: {"当前 Leader":"无","最大 term":"8","发起过的选举":"9","分裂投票":"4","同一 term 的 Leader 数":"1"}
no-random reset: {"当前 Leader":"无","最大 term":"6","发起过的选举":"30","分裂投票":"25","同一 term 的 Leader 数":"1"}
log tail: N4 选举超时仍未过半 → 分裂投票，重新发起 | N4 选举超时 → 成为 Candidate，term=6 | N5 选举超时仍未过半 → 分裂投票，重新发起 | N5 选举超时 → 成为 Candidate，term=6
exp3: 固定超时 · 平均轮数=60+ | 随机超时 · 平均轮数=1.06 | 耗时改善=55.0× | 固定超时分裂率=100%
exp3 jitter=0: 60+ | 1.07 | 55.1× | 100%
ERRORS: none
````

</details>

📎 **发送文件**：Part 3-A 交互式课件。核心是「实验 2 · Raft 选举模拟器」——真的在跑状态机和消息传递，可以点节点杀死它、制造分区、关掉随机化看活锁。统计面板里那个「同一 term 的 Leader 数」无论你怎么折腾都是 1。
- `〔路径00074〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00074〕 → file_uuid: 〔标识03281〕
````

</details>

📎 **发送文件**：Lab 3-A 代码。raft.go 的选举状态机核心不到 150 行；net.go 是确定性网络（同种子必复现）。`-unsafe` 开关专门用来把 Election Safety 拆给你看。
- `〔路径00056〕`
- `〔路径00083〕`
- `〔路径00102〕`
- `〔路径00103〕`
- `〔路径00101〕`
- `〔路径00133〕`
- `〔路径00131〕`

<details><summary>↳ SendUserFile 结果</summary>

````
7 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识02819〕
  〔路径00083〕 → file_uuid: 〔标识02490〕
  〔路径00102〕 → file_uuid: 〔标识03830〕
  〔路径00103〕 → file_uuid: 〔标识03981〕
  〔路径00101〕 → file_uuid: 〔标识00285〕
  〔路径00133〕 → file_uuid: 〔标识03259〕
  〔路径00131〕 → file_uuid: 〔标识03493〕
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00154〕\" \"$HOME〔路径00162〕\" && echo ready"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ready
````

</details>

💾 **写入你的 Mac**：
- `〔路径00008〕`
- `〔路径00004〕`
- `〔路径00014〕`
- `〔路径00028〕`
- `〔路径00029〕`
- `〔路径00027〕`
- `〔路径00047〕`
- `〔路径00046〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00008〕","〔路径00004〕","〔路径00014〕","〔路径00028〕","〔路径00029〕","〔路径00027〕","〔路径00047〕","〔路径00046〕"],"rejected":[]}
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #12 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #13 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #14 status
````

</details>

Part 3-A 交付完成，全部文件已写进 `~/Code/distributed-systems-course/`。

**先补一句关于你那个 sweep 兜底的提醒**：它必须满足「回溯窗口 > 最长事务时长」。哪天有个批量任务开了 20 分钟事务而 sweep 只回看 10 分钟，那条记录还是永久漏掉。所以最好再配一条定期全量对账，把"兜底有没有兜住"变成可观测的——**兜底方案本身也需要验证**。

---

**这一节最值得花时间的三个地方：**

1. **实验 2 是一个真的在跑的 Raft 选举实现**。可以点节点杀死、制造 3|2 分区、看少数派的 term 一路飙升却永远选不出主。盯住那个「同一 term 的 Leader 数」统计——我用各种方式折腾（40% 丢包、反复分区、连续杀节点），它始终是 1。这不是运气，是「过半票 + 每任期一票 ⇒ 两个过半集合必相交」，和 Part 2 的 Quorum 是同一个数学事实。

2. **Lab 3A-4 专门把安全性拆给你看**。加了个 `-unsafe` 开关，只去掉「每个任期只投一票」这一条规则，其他全不动：

   | | 同任期最多 Leader | 违反的种子 |
   |---|---|---|
   | 完整 Raft | 1 个 | 0/60 |
   | 去掉「每任期一票」 | **3 个** | **37/60** |

   把一条看起来像实现细节的规则删掉，脑裂立刻出现。这比读十遍论文里的证明管用。

3. **有个我没预料到但值得讲的发现**。场景③（3|2 分区）里，如果分区前的 Leader 恰好落在**少数派**一侧，基础 Raft **不会让它主动退位**——它继续给同侧 follower 发心跳，自认为还是 Leader。我把它叫做「僵尸 Leader」。它不违反安全性（拿不到过半确认，一条日志都提交不了），但客户端把写请求发给它会一直超时，读还可能读到陈旧数据。

   工程解法是 **CheckQuorum**（Leader 定期确认自己还能联系过半节点，否则主动退位），etcd 默认开启。**这也是"用了 Raft 就一定线性一致"的又一个反例**——线性一致读还需要 ReadIndex 或 Lease Read，那是 3-C 的内容。

Go 和 Java 两版的随机数生成器不同，所以具体谁当选不一样，但所有不变量在两边都成立。

---

**开始 3-B 之前，三个问题：**

1. **2PC 那个实验的场景④**（commit 只发出去一半就杀协调者），P2 和 P3 互相通信也解决不了问题。**为什么？**它俩明明都知道"我投了 yes、我没收到指令"。
2. **僵尸 Leader 那件事**：如果不加 CheckQuorum，客户端从这个僵尸 Leader 读数据会读到什么？是旧数据、报错、还是超时？取决于什么？
3. **提前想一想 3-B 的核心问题**：Leader 收到客户端请求，写进自己的日志，复制给 follower。**过半 follower 确认之后，这条日志就能提交了吗？**

第 3 题是 Raft 论文 Figure 8 的入口，也是整个 Raft 里最反直觉的一处。你先给一个直觉答案，不用怕答错——答错了正好，那说明这个坑对你是真实存在的。

> ✓ 本轮结束 · 2026-09-01 19:44 · 34 步
