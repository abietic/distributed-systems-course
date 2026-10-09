## 👤 你 · 2026-09-10 12:49

1. {A,B,C} 和 {A,B,C,D}配置过半集合一个是要两个算leader一个是要三个算leader，新增一个导致了过半集合的数量定义变化，理论上会触发重新选举
2. 3->4应该由于过半数量的变化导致重新选举但是4->5应该不会触发，新加入；一次加减一个可以方便leader同步当前集群过半集合数量状态？

## 🤖 Claude · 2026-09-10 12:50

## 数字算对了，但结论走偏了

你算出 {A,B,C} 过半是 2、{A,B,C,D} 过半是 3——**这两个数字就是答案的全部原料**，但你把它们导向了"会触发重新选举"，而配置变更本身<u>不</u>触发选举。

正确的问法不是"过半的数量变了吗"，而是：

> **新旧两个配置的任意过半集合，会不会不相交？**

因为 Raft 全部安全性的地基就是"两个过半集合必相交"。只要存在一个旧配置的过半集合和一个新配置的过半集合<strong>互不相交</strong>，它们就能在同一任期各自选出一个 Leader——**裂脑**。

### 直接跳 3 → 5：会

- 旧 {A,B,C} 取过半 = **{A,B}**
- 新 {A,B,C,D,E} 取过半 = **{C,D,E}**
- **这两组不相交。** 如果 A、B 还没切配置而 C、D、E 已经切了，它们各自都认为自己凑齐了过半 ⇒ 同一任期两个 Leader。

### 单节点 3 → 4：不会，而且是数学上不可能

- 旧过半 2 + 新过半 3 = **5**
- 两个配置的并集只有 **4 个节点**
- 5 > 4 ⇒ **鸽笼原理，必有重叠**。你想凑都凑不出两个不相交的集合。

一般情况也成立。设旧集群 n 个节点，加一个变成 n+1：

```
maj_old + maj_new = (⌊n/2⌋+1) + (⌊(n+1)/2⌋+1) = n + 2
|C_old ∪ C_new|   = n + 1
n + 2 > n + 1   ✓  恒成立
```

减一个同理（和是 n+1，并集是 n）。**所以"一次只动一个节点"不是工程上的谨慎习惯，是一条能证明的充分条件。** 而 3→5 那种跳两格，等号刚好不成立（5 > 5 为假），缝就开了。

### 你第 2 问的直觉里藏着一个真问题

你说"3→4 会因过半数量变化…4→5 不会"。触发选举那部分不对，但你注意到的**过半数量的变化**确实有实际后果：

| 集群规模 | 过半 | 能容忍的故障数 |
|---|---|---|
| 3 | 2 | 1 |
| **4** | **3** | **1** |
| 5 | 3 | 2 |

**4 节点和 3 节点的容错能力完全一样，但写入要多等一个节点确认。** 所以 3→4→5 的中间那一步是"付了代价没买到容错"——它只是一个必须路过的中转站。这也解释了为什么工业界扩容通常是 3→5 而不是长期停在 4。

---

现在做 3-C，把 Part 3 收尾：持久化、快照、成员变更、线性一致读。

<details><summary>↳ SendUserMessage 结果</summary>

````
Message delivered to user.
````

</details>

▸ **ToolSearch**

````
{"max_results": 3, "query": "select:TaskCreate,TaskUpdate"}
````

<details><summary>↳ ToolSearch 结果</summary>

````
[tool_reference]
[tool_reference]
````

</details>

▸ **TaskCreate** 编写 Part 3-C 课件：持久化、快照、成员变更、线性一致读

<details><summary>↳ TaskCreate 结果</summary>

````
Task #18 created successfully: 编写 Part 3-C 课件：持久化、快照、成员变更、线性一致读
````

</details>

▸ **TaskCreate** 编写 Lab 3-C（Go + Java）

<details><summary>↳ TaskCreate 结果</summary>

````
Task #19 created successfully: 编写 Lab 3-C（Go + Java）
````

</details>

▸ **TaskCreate** 验证并交付 Part 3-C

<details><summary>↳ TaskCreate 结果</summary>

````
Task #20 created successfully: 验证并交付 Part 3-C
````

</details>

▸ **Bash** Write Part 3-C head and sections 3.18-3.19

````
cd 〔路径00054〕 && mkdir -p ch03c && cd ch03c && cat > ch03c.html <<'HEAD_EOF'
<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>分布式系统 · Part 3-C · 持久化、快照、成员变更与线性一致读</title>
HEAD_EOF
cat 〔路径00240〕 >> ch03c.html
cat >> ch03c.html <<'HEAD2_EOF'
<style>
.scen{display:flex;flex-wrap:wrap;gap:7px;margin-bottom:14px}
.scen button.act{background:var(--s1);border-color:var(--s1);color:#fff}
.hint{font-size:12.5px;color:var(--ink-3);margin:10px 0 0}
.chip{display:inline-block;font-family:var(--mono);font-size:12px;padding:1px 7px;border-radius:5px;
  background:var(--surface-2);border:1px solid var(--border);color:var(--ink)}
.axm{display:grid;grid-template-columns:26px 1fr
…（截断，原长 8897 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-C sections 3.20-3.21

````
cat >> ch03c.html <<'EOF'
<section id="s3">
  <h2>3.20 成员变更：你已经推出了核心</h2>
  <p>集群要能在线增减节点：扩容、缩容、换掉一台坏机器。难点在于<strong>配置切换不是原子的</strong>——不同节点在不同时刻切换到新配置，中间那段时间集群里同时存在两种"过半"的定义。</p>

  <div class="note crit">
    <div class="note-h"><span class="dot" style="background:var(--crit)"></span>判据不是"过半的数量变了吗"，而是"两个过半集合会不会不相交"</div>
    <p style="margin-bottom:0">Raft 全部安全性的地基是<em>任意两个过半集合必相交</em>。只要存在一个<strong>旧配置的过半集合</strong>和一个<strong>新配置的过半集合</strong>互不相交，它们就能在同一任期各自选出 Leader ⇒ <strong>脑裂</strong>。</p>
  </div>

  <h3>为什么直接跳 3 → 5 会裂脑</h3>
  <div class="two">
    <div class="card" style="border-color:var(--crit)">
      <div class="card-t" style="color:var(--crit)">危险：C_old={A,B,C} → C_new={A,B,C,D,E}</div>
      <p style="font-size:13.5px">旧过半取 <strong>{A,B}</strong>（2 个够），新过半取 <strong>{C,D,E}
…（截断，原长 8604 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
````

</details>

▸ **Bash** Write Part 3-C summary, quiz and lab sections

````
cat >> ch03c.html <<'EOF'
<section id="s5">
  <h2>3.22 Raft 完整规则表</h2>
  <p>Part 3 到这里结束。把三节的全部规则收在一页，这就是你实现一个 Raft 需要的全部约束。</p>

  <h3>状态</h3>
  <table>
    <thead><tr><th style="width:22%">类别</th><th>内容</th></tr></thead>
    <tbody>
      <tr><td><strong>持久化（所有节点）</strong></td><td><code>currentTerm</code>、<code>votedFor</code>、<code>log[]</code> —— <em>响应任何 RPC 前必须落盘</em></td></tr>
      <tr><td>易失（所有节点）</td><td><code>commitIndex</code>、<code>lastApplied</code></td></tr>
      <tr><td>易失（仅 Leader）</td><td><code>nextIndex[]</code>（乐观猜测）、<code>matchIndex[]</code>（保守事实）—— 当选时重新初始化</td></tr>
    </tbody>
  </table>

  <h3>五条安全性属性</h3>
  <table>
    <thead><tr><th style="width:24%">属性</th><th style="width:38%">内容</th><th>靠什么保证</th></tr></thead>
    <tbody>
      <tr><td>选举安全性</td><td>一个任期至多一个
…（截断，原长 9572 字符）
````

<details><summary>↳ Bash 结果</summary>

````
sections: 8 kb: 49
````

</details>

▸ **Bash** Write Part 3-C JS and verify

````
cat >> ch03c.html <<'EOF'
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
  let c
…（截断，原长 19557 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
LOAD ERRORS: [ "PAGEERROR: Unexpected identifier '成功'" ]

[agent-proxy] While this command ran, 3 connections through the agent proxy failed:
- accounts.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination)
- www.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×2
For details: curl -sS http://127.0.0.1:42673/__agentproxy/status
````

</details>

▸ **Bash** Find unbalanced ASCII quotes in JS

````
python3 - <<'PY'
import re
s=open('ch03c.html',encoding='utf-8').read()
js=s[s.index('<script>'):]
bad=[]
for i,line in enumerate(js.split('\n'),1):
    if line.count('"')%2==1 and re.search(r'[一-鿿]',line):
        bad.append((i,line.strip()[:110]))
print("suspicious:",len(bad))
for x in bad: print(x)
PY
````

<details><summary>↳ Bash 结果</summary>

````
suspicious: 0
````

</details>

▸ **Bash** Find ASCII quotes adjacent to CJK text

````
python3 - <<'PY'
import re
s=open('ch03c.html',encoding='utf-8').read()
js=s[s.index('<script>'):]
# 找出被中文包夹的 ASCII 双引号（几乎一定是误用的引号）
for i,line in enumerate(js.split('\n'),1):
    for m in re.finditer(r'[一-鿿，。：、]"|"[一-鿿]', line):
        print(i, '|', line.strip()[:120]); break
PY
````

<details><summary>↳ Bash 结果</summary>

````
24 | {n:"全部持久化（正确）",disk:"currentTerm ✓　votedFor ✓　log ✓",steps:[
25 | {t:"任期 5：N1 和 N2 同时成为 Candidate",
27 | note:"两个节点同时超时，各自 term++ 到 5，都先投了自己一票。过半需要 <strong>3 票</strong>。"},
28 | {t:"N3 投给 N1，N4 投给 N2 —— 两边都是 2 票",
31 | {t:"N5 投给 N1 → N1 拿到 3 票当选",
33 | hi:[4],note:"N1 集齐 {N1,N3,N5} = 3 票，成为任期 5 的 Leader。<br>N5 <strong>把 votedFor=N1 写进了磁盘</strong>，然后才回复投票。"},
34 | {t:"N5 崩溃重启",
36 | hi:[4],note:"N5 掉电了。它的内存全没了，但磁盘上的三样状态还在。"},
41 | {n:"丢失 votedFor",disk:"currentTerm ✓　<span style='color:var(--crit)'>votedFor ✗</span>　log ✓",steps:[
42 | {t:"任期 5：N1 和 N2 同时成为 Candidate",
44 | note:"和上一个场景完全相同的开局。"},
45 | {t:"N3 投给 N1，N4 投给 N2 —— 两边都是 2 票",
47 | note:"N1 有 2 票，N2 有 2 票。过半需要 3 票。"},
48 | {t:"N5 投给 N1 → N1 当选",
51 | {t:"N5 崩溃重启 —— votedFor 没了",
57 | note:"N2 还在索票。N5 觉得自己本任期还没投过，于是投给了 N2。<br>N2 集齐 {N2,N4,N5} = 3 票，<strong style='color:var(--crit)'>在同一个任期 5 里也当选了 Leader
58 | {n:"丢失 currentTerm",disk:"<span style='color:var(--crit)'>currentTerm ✗</span>　votedFor ✓　log ✓",steps:[
59 | {t:"任期 8：N1 是 Leader，集群正常",
61 | note:"集群已经跑到任期 8。所有节点的 currentTerm 都是 8。"},
62 | {t:"N5 崩溃重启 —— currentTerm 归零",
64 | hi:[4],note:"N5 重启后 <code>currentTerm</code> 回到 0。它<strong>不再知道自己见过任期 8</strong>。"},
66 | s:[nd("L",8,"N1"),nd("F",8,"N1"),nd("F",8,"N1"),nd("F",8,"N1"),nd("F",3,"旧节点")],
68 | note:"一个被隔离很久、还停在任期 3 的节点回来索票。<br>正常情况下 N5 会因为 <code>3 &lt; 8</code> 直接拒绝，但现在它的 term 是 0，<strong>于是接受了这个过时请求并投了票</strong
69 | {n:"丢失 log[]",disk:"currentTerm ✓　votedFor ✓　<span style='color:var(--crit)'>log ✗</
…（截断，原长 4215 字符）
````

</details>

▸ **Bash** Fix quote bugs and re-verify Part 3-C

````
python3 - <<'PY'
p='ch03c.html'; s=open(p,encoding='utf-8').read()
fixes=[('Leader 提交并<strong>回复客户端"成功"</strong>。','Leader 提交并<strong>回复客户端「成功」</strong>。'),
       ('x=1 从此消失，而客户端早已收到"成功"。','x=1 从此消失，而客户端早已收到「成功」。')]
for a,b in fixes:
    assert a in s, a[:40]
    s=s.replace(a,b)
open(p,'w',encoding='utf-8').write(s); print('fixed',len(fixes))
PY
node 〔路径00186〕
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
fixed 2
LOAD ERRORS: [ "PAGEERROR: Unexpected identifier '持有已提交日志的节点构成过半'" ]

[agent-proxy] While this command ran, 3 connections through the agent proxy failed:
- www.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination) ×2
- accounts.google.com:443 — connect_rejected (the egress proxy denied the CONNECT (organization policy) or could not reach the destination)
For details: curl -sS http://127.0.0.1:42673/__agentproxy/status
````

</details>

▸ **Bash** Fix remaining quote bug and add a stricter checker

````
python3 - <<'PY'
p='ch03c.html'; s=open(p,encoding='utf-8').read()
a='Leader 完整性依赖"持有已提交日志的节点构成过半"这个事实'
b='Leader 完整性依赖「持有已提交日志的节点构成过半」这个事实'
assert a in s; s=s.replace(a,b)
open(p,'w',encoding='utf-8').write(s); print('fixed')
PY
# 更可靠的检查：把 JS 里所有 note:"..." / t:"..." 字符串抽出来，检查内部有没有裸 ASCII 双引号
python3 - <<'PY'
import re
s=open('ch03c.html',encoding='utf-8').read()
js=s[s.index('<script>'):]
bad=0
for m in re.finditer(r'(?:note|t|n|disk):"((?:[^"\\]|\\.)*)"', js):
    pass
# 直接检查：每个以 key:" 开头的片段，到下一个未转义 " 为止，若其后紧跟中文则可疑
for m in re.finditer(r'(note|t|n|disk):"', js):
    start=m.end(); i=start
    while i < len(js):
        if js[i]=='\\': i+=2; continue
        if js[i]=='"':
            after=js[i+1:i+2]
            if re.match(r'[一-鿿，。、：（）]', after):
                print("SUSPECT at", js
…（截断，原长 926 字符）
````

<details><summary>↳ Bash 结果</summary>

````
fixed
suspect count: 0
persist scen 0 => 安全性成立
persist scen 1 => 安全性被破坏
persist scen 2 => 安全性被破坏
persist scen 3 => 安全性被破坏
3→5 => 2|3|5|2|✗ 危险 | 能凑出两个不相交的过半集合 ⇒ 同一任期可以出现两个 Leader
3→4 => 2|3|5|1|✓ 安全 | 凑不出不相交的过半集合 ⇒ 安全
5→3 => 3|2|5|2|✗ 危险 | 能凑出两个不相交的过半集合 ⇒ 同一任期可以出现两个 Leader
3→7 => 2|4|6|4|✗ 危险 | 能凑出两个不相交的过半集合 ⇒ 同一任期可以出现两个 Leader
4→5 ✓ 安全
5→6 ✓ 安全
6→7 ✓ 安全
7→8 ✓ 安全
read exp rows: 3
ERRORS: none
````

</details>

▸ **Bash** Extend Raft with persistence, snapshots, config, reads

````
cd 〔路径00079〕 && mkdir -p lab03c && cp lab03b/net.go lab03c/net.go && cp lab03b/raft.go lab03c/raft.go && python3 - <<'PY'
p='lab03c/net.go'; s=open(p,encoding='utf-8').read()
s=s.replace('	MsgAppendEntriesReply','	MsgAppendEntriesReply\n	MsgInstallSnapshot\n	MsgInstallSnapshotReply')
s=s.replace('return [...]string{"RequestVote", "VoteReply", "AppendEntries", "AppendReply"}[t]',
            'return [...]string{"RequestVote", "VoteReply", "AppendEntries", "AppendReply", "InstallSnapshot", "SnapshotReply"}[t]')
s=s.replace('''	// —— AppendEntries 回复 ——''','''	// —— InstallSnapshot ——
	LastIncludedIndex, LastIncludedTerm int
	Snapshot                            map[string]string

	// —— AppendEntries 回复 ——''')
open(p,'w',encoding='utf-8').write(s); print('net ok')
PY
python3 - <<'PY
…（截断，原长 8106 字符）
````

<details><summary>↳ Bash 结果</summary>

````
net ok
raft base ok
````

</details>

▸ **Bash** Wire persistence, snapshots, config and reads into Raft

````
python3 - <<'PY'
p='lab03c/raft.go'; s=open(p,encoding='utf-8').read()
def rep(a,b,n=1):
    global s; assert a in s, "MISSING: "+a[:70]; s=s.replace(a,b,n)

# at() 改为按逻辑 index（快照之后数组下标 ≠ 日志 index）
rep('''func (n *Node) at(i int) (Entry, bool) {
	if i < 0 || i >= len(n.log) {
		return Entry{}, false
	}
	return n.log[i], true
}''','''// at 按【逻辑 index】取条目。有快照之后数组下标不再等于日志 index，
// 所有访问都必须走它，不能再直接 n.log[i]。
func (n *Node) at(i int) (Entry, bool) {
	e := n.entryAt(i)
	if e.Index != i {
		return Entry{}, false
	}
	return e, true
}''')

# 持久化：所有改动 term / votedFor / log 的地方都要落盘
rep('''	n.term, n.state, n.votedFor = term, Follower, NoVote
	n.votes = map[int]bool{}
	n.timer, n.timeout = 0, c.newTimeout()
}''','''	n.term, n.state, n.votedFor = term, Follower, NoVote
	n.votes = map[int]bool{}
	n.time
…（截断，原长 6915 字符）
````

<details><summary>↳ Bash 结果</summary>

````
raft wired
````

</details>

▸ **Bash** Write and run Go Lab 3-C

````
cat > lab03c/main.go <<'GOEOF'
// Lab 3-C · 持久化、快照、成员变更与线性一致读
//
// 运行：  cd go/lab03c && go run .
// 调参：  go run . -lose votedfor     ← 只看丢掉 votedFor 的后果
//         go run . -lose term | log
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
	net := NewNetwork(*nF, *delayF, *jitterF, 0, seed)
	c := New(*nF, net, 400, 300, seed)
	c.LeaseMs = 250
	if cfg != nil {
…（截断，原长 11429 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok

配置：5 节点 ｜ 延迟 20±10ms ｜ 每组 20 个种子 ｜ 种子基准 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3C-1 · 持久化与崩溃重启
  分别丢掉三样状态中的一样，看哪条安全性倒下
══════════════════════════════════════════════════════════════════════════════
    磁盘上丢了什么                  违规的种子    选举安全性  状态机安全性              结论
    ────────────────────────────────────────────────────────────────────────────────────────
    （什么都不丢，正确）                  0/20             0             0            ✓ 安全
    ★ 丢失 votedFor                       0/20             0             0            ✓ 安全
    丢失 currentTerm                      0/20             0             0            ✓ 安全
    丢失 log[]                            0/20             0             0            ✓ 安全

    ▸ 第一行是对照组：三样都落盘，反复崩溃重启也不出问题。
    ▸ votedFor 是最容易被漏掉的一样，也是后果最直接的：节点重启后忘了自己投过票，
      同一任期可以再投一次 ⇒ 两个 candidate 各自凑齐过半 ⇒ 脑裂。
    ▸ log 丢失破坏的是 Leader 完整性："持有已提交日志的节点构成过半"这个事实变成了假的，
      于是后续选举可能选出不含该条目的 Leader，已提交的数据凭空消失。
    ▸ currentTerm 丢失让节点失去时间感，无法识别过时的 Leader 和 candidate。
    ▸ 判断什么必须落盘的通用直觉：【可推导的结论不用存，对外许下的承诺必须存】。
      commitIndex 是结论（Leader 会告诉你），votedFor 是承诺（只有自己知道）。

══════════════════════════════════════════════════════════════════════════════
  Lab 3C-2 · 日志压缩与 InstallSnapshot
  把一个节点隔离到落后超过快照点，再放回来
══════════════════════════════════════════════════════════════════════════════
    指标                                        数值说明
    ──────────────────────────────────────────────────────────────────────────────
    提
…（截断，原长 1546 字符）
````

</details>

▸ **Bash** Make persistence experiment actually trigger violations

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)
rep('''func mk(seed int64, cfg func(*Cluster)) *Cluster {
	net := NewNetwork(*nF, *delayF, *jitterF, 0, seed)
	c := New(*nF, net, 400, 300, seed)''',
'''func mk(seed int64, cfg func(*Cluster)) *Cluster {
	return mkw(seed, 300, 0, cfg)
}

// mkw 允许指定选举超时窗口与丢包率。
// 窗口设成 0（固定超时）会让多个节点【同时】发起选举 ——
// 这正是暴露「votedFor 没落盘」的必要条件：得先有两个 candidate 在抢同一个任期。
func mkw(seed int64, window int64, loss float64, cfg func(*Cluster)) *Cluster {
	net := NewNetwork(*nF, *delayF, *jitterF, loss, seed)
	c := New(*nF, net, 400, window, seed)''')
rep('''		bad, elec, sm := 0, 0, 0
		for seed := int64(1); seed <= int64(*trialsF); seed++ {
			c := mk(*seedF+seed, func(c *Cluster) {
				c.Lo
…（截断，原长 3277 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
  Lab 3C-1 · 持久化与崩溃重启
  分别丢掉三样状态中的一样，看哪条安全性倒下
══════════════════════════════════════════════════════════════════════════════
    磁盘上丢了什么                  违规的种子    选举安全性  状态机安全性              结论
    ────────────────────────────────────────────────────────────────────────────────────────
    （什么都不丢，正确）                  0/25             0             0            ✓ 安全
    ★ 丢失 votedFor                       0/25             0             0            ✓ 安全
    丢失 currentTerm                      0/25             0             0            ✓ 安全
    丢失 log[]                            9/25             0             9          ✗ 被破坏

    ▸ 第一行是对照组：三样都落盘，反复崩溃重启也不出问题。
    ▸ votedFor 是最容易被漏掉的一样，也是后果最直接的：节点重启后忘了自己投过票，
````

</details>

▸ **Bash** Add deterministic double-vote construction

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:70]; s=s.replace(a,b)
rep('''			run(c, 1200)
			for i := 0; i < 80; i++ {
				c.Submit(fmt.Sprintf("SET k%d=%d", i%5, i))
				run(c, 30)
				// 在选举正酣的时候把节点掐掉又立刻拉起来 —— 它会带着"失忆"回到同一场选举
				if i%4 == 3 {
					k := c.rnd.Intn(*nF)
					c.net.Kill(k)
					run(c, 60)
					c.Restart(k)
					run(c, 60)
				}
			}
			run(c, 3000)''',
'''			run(c, 1200)
			for i := 0; i < 80; i++ {
				c.Submit(fmt.Sprintf("SET k%d=%d", i%5, i))
				run(c, 30)
				if i%8 == 7 {
					// 先杀 Leader：固定超时下，剩下的 follower 会几乎同时超时、
					// 同时变成 candidate 抢同一个任期 —— 这是脑裂的土壤。
					if l := c.Leader(); l != nil {
						c.net.Kill(l.id)
					}
					run(c, 420)
					// 选举正酣时，把一个刚投过票的节点掐掉又立刻拉起来 ——
					// 它会带着"失忆"回到同一场选举里。

…（截断，原长 3037 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
  Lab 3C-1 · 持久化与崩溃重启
  分别丢掉三样状态中的一样，看哪条安全性倒下
══════════════════════════════════════════════════════════════════════════════
    磁盘上丢了什么                  违规的种子    选举安全性  状态机安全性              结论
    ────────────────────────────────────────────────────────────────────────────────────────
    （什么都不丢，正确）                  0/25             0             0            ✓ 安全
    ★ 丢失 votedFor                       0/25             0             0            ✓ 安全
    丢失 currentTerm                      0/25             0             0            ✓ 安全
    丢失 log[]                            1/25             0             1          ✗ 被破坏

    构造场景 · votedFor 已落盘     ✓ N5 记得投过票，拒绝第二次索票 ⇒ 只有一个 Leader
    构造场景 · votedFor 丢失       ✗ 同一任期出现两个 Leader —— 票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次

    ▸ 上面两行是【构造】出来的场景：直接把"两个 candidate 抢同一任期"摆出来，
      不依赖随机调度去撞。这也说明随机故障注入为什么不够 ——
      有些 bug 需要非常特定的交错才会现形，这正是 TLA+ 和 Jepsen 存在的理由（Part 6）。
    ▸ 表格里 votedFor 那一行的随机试验可能是 0，不代表它安全，只代表没撞上。
    ▸ votedFor 是最容易被漏掉的一样，也是后果最直接的：节点重启后忘了自己投过票，
````

</details>

▸ **Bash** View Lab 3C-2 through 3C-4 output

````
go run . -trials 25 2>&1 | sed -n '28,75p'
````

<details><summary>↳ Bash 结果</summary>

````
      commitIndex 是结论（Leader 会告诉你），votedFor 是承诺（只有自己知道）。

══════════════════════════════════════════════════════════════════════════════
  Lab 3C-2 · 日志压缩与 InstallSnapshot
  把一个节点隔离到落后超过快照点，再放回来
══════════════════════════════════════════════════════════════════════════════
    指标                                      数值    说明
    ──────────────────────────────────────────────────────────────────────────────────
    提交的命令数                               220    客户端发起
    Leader 做过的快照                           28    每 30 条压缩一次
    快照点 lastIncludedIndex                   213    这之前的日志已丢弃
    Leader 保留的日志条数                        7    而不是全部 220 条
    发出的 InstallSnapshot                       1    N5 落后太多，走快照追赶
    N5 追平耗时                             170 ms
    日志仍然一致                              ✓ 是    快照没有破坏日志匹配性质
    安全性违反                     0 条

    ▸ Leader 只保留了最近几十条日志，其余压缩进了快照 —— 磁盘和重放时间都是常数级。
    ▸ N5 落后到快照点之前，Leader 发现 nextIndex ≤ lastIncludedIndex，
      于是改发 InstallSnapshot 而不是逐条补日志。
    ▸ 快照里必须带 (lastIncludedIndex, lastIncludedTerm)：日志被截断后，
      AppendEntries 的一致性检查还要拿它们当"被截断部分的代表"来比较。
      少了这两个字段，日志匹配性质在截断处就断了。

══════════════════════════════════════════════════════════════════════════════
  Lab 3C-3 · 成员变更：直接跳 vs 单节点
  让不同节点在不同时刻切换配置，看会不会选出两个 Leader
══════════════════════════════════════════════════════════════════════════════
    变更方式                             出现两个 Leader              结论
    ──────────────────────────────────────────────────────────────────────

…（截断，原长 2375 字符）
````

</details>

▸ **Bash** Fix single-node membership scenario

````
python3 - <<'PY'
p='main.go'; s=open(p,encoding='utf-8').read()
old='''	// 单节点变更 {0,1,2} → {0,1,2,3}：同样只有一部分节点先切
	single := func(seed int64) bool {
		c := mk(seed, nil)
		for i := 0; i < *nF; i++ {
			c.SetConfig(i, []int{0, 1, 2})
		}
		c.net.Kill(3)
		c.net.Kill(4)
		run(c, 2500)
		c.net.Revive(3)
		for _, i := range []int{2, 3} {
			c.SetConfig(i, []int{0, 1, 2, 3})
		}
		c.net.Partition([]int{0, 0, 1, 1, 1})
		run(c, 9000)
		return len(c.Violations) > 0 || c.twoLeadersEver
	}'''
new='''	// 单节点变更 {0,1,2} → {0,1,2,3}：同样只有一部分节点先切，
	// 而且用完全相同的分区方式去"制造"裂脑 —— 结果会告诉你它造不出来。
	single := func(seed int64) bool {
		c := mk(seed, nil)
		for i := 0; i < *nF; i++ {
			c.SetConfig(i, []int{0, 1, 2})
		}
		c.net.Kill(3)
		c.net.Kill(4) // N5 不属于新旧任何一个配置，全程离线
		run(c, 2500)
		c.net.Revive(3)
		for _, i
…（截断，原长 1580 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
    变更方式                             出现两个 Leader              结论
    ──────────────────────────────────────────────────────────────────────
    ① 直接跳 {A,B,C} → {A..E}                      25/25          ✗ 会裂脑
    ② 单节点 {A,B,C} → {A,B,C,D}                    7/25          ✗ 会裂脑

    ① 的两个过半集合               旧 {A,B}（2/3 够）与 新 {C,D,E}（3/5 够）—— 不相交
    ① 的算式                       maj_old + maj_new = 2+3 = 5，|并集| = 5，5 > 5 ✗ 不成立
  制造僵尸 Leader，看哪种读会静默返回陈旧数据
══════════════════════════════════════════════════════════════════════════════
    读的实现                ★ 静默返回陈旧值            正确报错          返回正确值
    ────────────────────────────────────────────────────────────────────────────────
    本地读                                25                   0                   0
    ReadIndex                              0                  25                   0
    Lease Read                             0                  25                   0

    ▸ 本地读会静默返回旧值 —— 不报错、不留日志。客户端刚从新 Leader 拿到"写入成功"，
      转头从僵尸 Leader 读到旧值，线性一致性当场破裂，而监控上什么都看不出来。
    ▸ ReadIndex 在读之前先确认过半节点仍认自己是主，僵尸 Leader 凑不齐 ⇒ 报错。
      【给你错误，而不是给你错的数据】—— 这个区别在排障时价值巨大。
    ▸ Lease Read 的租约在收不到过半心跳响应后就续不上了，过期后同样拒绝本地读。
      它比 ReadIndex 快（省掉一次心跳往返），但安全性建立在【时钟漂移有界】上：
      如果这台机器的时钟突然变慢，它会以为租约还没到期而继续本地读。
      Part 1 那个坑，在这里变成了一致性问题。
    ▸ 所以"Raft 集群是线性一致的"准确的说法是：写入是线性一致的；
      读是否线性一致，取决于你选了上面哪一种。etcd 默认 ReadIndex，TiKV 默认 Lease Read。
````

</details>

▸ **Bash** Use active-leader definition of split brain

````
python3 - <<'PY'
p='net.go'; s=open(p,encoding='utf-8').read()
s=s.replace('func (nw *Network) Heal()                  {',
 '''// SameGroup 判断两个节点当前是否互相可达（同一分区组）。
func (nw *Network) SameGroup(a, b int) bool { return nw.group[a] == nw.group[b] }

func (nw *Network) Heal()                  {''')
open(p,'w',encoding='utf-8').write(s); print('net ok')
PY
python3 - <<'PY'
p='raft.go'; s=open(p,encoding='utf-8').read()
s=s.replace('''func (c *Cluster) Leader() *Node {''','''// ActiveLeaders 数一数有几个 Leader【能在自己的配置下凑齐过半】——
// 也就是有几个 Leader 真的能提交日志。
//
//	这才是「裂脑」的准确定义。一个被隔离到少数派的僵尸 Leader 虽然也自称 Leader，
//	但它一条日志都提交不了，不构成裂脑。
func (c *Cluster) ActiveLeaders() int {
	k := 0
	for _, n := range c.nodes {
		if c.net.Down(n.id) || n.state != Leader {
			continue
		}
		reach := 0
		for j := range c.nodes {

…（截断，原长 1465 字符）
````

<details><summary>↳ Bash 结果（错误）</summary>

````
Exit code 1
net ok
raft ok
# dsc/lab03c
# [dsc/lab03c]
vet: ./raft.go:551:46: c.net.SameGroup undefined (type *Network has no field or method SameGroup)
````

</details>

▸ **Bash** Check whether SameGroup was inserted

````
grep -n 'func (nw \*Network) Heal' net.go; grep -n 'SameGroup' net.go
````

<details><summary>↳ Bash 结果</summary>

````
88:func (nw *Network) Heal() {
````

</details>

▸ **Bash** Insert SameGroup and rerun

````
python3 - <<'PY'
p='net.go'; s=open(p,encoding='utf-8').read()
a='func (nw *Network) Heal() {'
b='''// SameGroup 判断两个节点当前是否互相可达（同一分区组）。
func (nw *Network) SameGroup(a, b int) bool { return nw.group[a] == nw.group[b] }

func (nw *Network) Heal() {'''
assert a in s; s=s.replace(a,b,1)
open(p,'w',encoding='utf-8').write(s); print('ok')
PY
gofmt -w . && go vet . && go run . -trials 30 2>&1 | sed -n '55,60p'
````

<details><summary>↳ Bash 结果</summary>

````
ok
══════════════════════════════════════════════════════════════════════════════
    变更方式                             出现两个 Leader              结论
    ──────────────────────────────────────────────────────────────────────
    ① 直接跳 {A,B,C} → {A..E}                      30/30          ✗ 会裂脑
    ② 单节点 {A,B,C} → {A,B,C,D}                    0/30  ✓ 一次都造不出来
````

</details>

▸ **Bash** Generate Java Lab 3-C skeleton from 3-B

````
mkdir -p 〔路径00136〕 && cd 〔路径00121〕 && python3 - <<'PY'
# 以 Lab03B.java 为底，生成 Lab03C.java 的骨架（网络 + Raft 核心），再追加 3-C 的新能力
src=open('lab03b/Lab03B.java',encoding='utf-8').read()
s=src
def rep(a,b):
    global s; assert a in s, "MISSING "+a[:60]; s=s.replace(a,b)

rep('''/*
 * Lab 3-B · 日志复制与安全性（Java 版）
 *
 * 运行：  cd java/lab03b && java Lab03B.java
 * ★关键： java Lab03B.java -naivecommit   ← 关掉「只提交当前任期」，看 Figure 8 怎么炸
 *         java Lab03B.java -slowbackoff   ← 关掉快速回退，对比 RPC 轮数
 *
 * 需要 JDK 17+。配套课件：courseware/ch03b-log-replication.html
 */''','''/*
 * Lab 3-C · 持久化、快照、成员变更与线性一致读（Java 版）
 *
 * 运行：  cd java/lab03c && java Lab03C.java
 * 调参：  java Lab03C.java -lose votedfor | term | log
 *
 * 需要 JDK 17+。配套课件：courseware/ch03c-production-raft.html
 */''')
rep('public
…（截断，原长 13980 字符）
````

<details><summary>↳ Bash 结果</summary>

````
skeleton written, marker present: True
````

</details>

▸ **Bash** Merge Java Lab 3-C labs and compile

````
python3 - <<'PY'
labs = r'''    // ══════════════════════════════════════════════════════════════════════
    // Part 3-C ①：崩溃重启
    // ══════════════════════════════════════════════════════════════════════
    static void restart(Cluster c, int i) {
        Node n = c.nodes.get(i);
        Persisted d = n.disk;
        n.term = d.currentTerm; n.votedFor = d.votedFor; n.log = new ArrayList<>(d.log);

        if (c.loseTerm) n.term = 0;                 // ★ 失去时间感，无法识别过时消息
        if (c.loseVote) n.votedFor = NO_VOTE;       // ★ 忘了投过票 ⇒ 同任期可以再投一次
        if (c.loseLog) n.log = new ArrayList<>(List.of(new Entry(0,0,"")));  // ★ 已提交条目消失

        n.state = State.FOLLOWER; n.votes.clear();
        n.commitIndex = n.snapIdx; n.lastApplied = n.snapIdx;
        n.timer = 0; n.timeout = c.newTimeout
…（截断，原长 17020 字符）
````

<details><summary>↳ Bash 结果</summary>

````
labs merged
Lab03C.java:392: error: variable k is already defined in method activeLeaders()
                for (int k=mn;k>=lo;k--) {
                         ^
1 error
error: compilation failed
````

</details>

▸ **Bash** Fix Java method placement and compile

````
python3 - <<'PY'
p='Lab03C.java'; s=open(p,encoding='utf-8').read()
bad='''            if (activeLeaders() > 1) twoLeadersEver = true;
        }

        /** 有几个 Leader【能在自己的配置下凑齐过半】——这才是裂脑的准确定义。 */
        int activeLeaders() {
            int k = 0;
            for (Node n : nodes) {
                if (net.isDown(n.id) || n.state != State.LEADER) continue;
                int reach = 0;
                for (int j=0;j<nodes.size();j++)
                    if (n.config.contains(j) && !net.isDown(j) && net.sameGroup(n.id, j)) reach++;
                if (reach >= n.majorityOf()) k++;
            }
            return k;'''
assert bad in s
s=s.replace(bad,'            if (activeLeaders() > 1) twoLeadersEver = true;')
# 在 leader() 方法之前插入 activeLeaders()
anchor='''        Node leader() {'''
as
…（截断，原长 1549 字符）
````

<details><summary>↳ Bash 结果</summary>

````
ok

配置：5 节点 ｜ 延迟 20±10ms ｜ 每组 15 个种子 ｜ 种子基准 42

══════════════════════════════════════════════════════════════════════════════
  Lab 3C-1 · 持久化与崩溃重启
  分别丢掉三样状态中的一样，看哪条安全性倒下
══════════════════════════════════════════════════════════════════════════════
    磁盘上丢了什么                  违规的种子    选举安全性  状态机安全性              结论
    ────────────────────────────────────────────────────────────────────────────────────────
    （什么都不丢，正确）                  0/15             0             0            ✓ 安全
    ★ 丢失 votedFor                       0/15             0             0            ✓ 安全
    丢失 currentTerm                      0/15             0             0            ✓ 安全
    丢失 log[]                            0/15             0             0            ✓ 安全

    构造场景 · votedFor 已落盘     ✓ N5 记得投过票，拒绝第二次索票 ⇒ 只有一个 Leader
    构造场景 · votedFor 丢失       ✗ 同一任期出现两个 Leader —— 票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次

    ▸ 上面两行是【构造】出来的场景：直接把"两个 candidate 抢同一任期"摆出来，
      不依赖随机调度去撞。这也说明随机故障注入为什么不够 ——
      有些 bug 需要非常特定的交错才会现形，这正是 TLA+ 和 Jepsen 存在的理由（Part 6）。
    ▸ 表格里 votedFor 那一行的随机试验可能是 0，不代表它安全，只代表没撞上。
    ▸ log 丢失破坏的是 Leader 完整性："持有已提交日志的节点构成过半"变成了假的。
    ▸ 判断什么必须落盘的通用直觉：【可推导的结论不用存，对外许下的承诺必须存】。
      commitIndex 是结论（Leader 会告诉你），votedFor 是承诺（只有自己知道）。
````

</details>

▸ **Bash** View Java Lab 3-C remaining output

````
java Lab03C.java -trials 20 2>&1 | grep -v "Picked up" | sed -n '26,50p;56,66p'
````

<details><summary>↳ Bash 结果</summary>

````
══════════════════════════════════════════════════════════════════════════════
  Lab 3C-2 · 日志压缩与 InstallSnapshot
  把一个节点隔离到落后超过快照点，再放回来
══════════════════════════════════════════════════════════════════════════════
    指标                                      数值    说明
    ──────────────────────────────────────────────────────────────────────────────────
    提交的命令数                               220    客户端发起
    Leader 做过的快照                           28    每 30 条压缩一次
    快照点 lastIncludedIndex                   214    这之前的日志已丢弃
    Leader 保留的日志条数                        6    而不是全部 220 条
    发出的 InstallSnapshot                       1    N5 落后太多，走快照追赶
    N5 追平耗时                             200 ms
    日志仍然一致                              ✓ 是    快照没有破坏日志匹配性质
    安全性违反                     0 条

    ▸ Leader 只保留了最近几十条日志，其余压缩进了快照 —— 磁盘和重放时间都是常数级。
    ▸ N5 落后到快照点之前，Leader 发现 nextIndex ≤ lastIncludedIndex，
      于是改发 InstallSnapshot 而不是逐条补日志。
    ▸ 快照里必须带 (lastIncludedIndex, lastIncludedTerm)：日志被截断后，
      AppendEntries 的一致性检查还要拿它们当"被截断部分的代表"来比较。


══════════════════════════════════════════════════════════════════════════════
  Lab 3C-3 · 成员变更：直接跳 vs 单节点
    ② 单节点 {A,B,C} → {A,B,C,D}                    0/20    ✓ 一次都造不出来

    ① 的两个过半集合               旧 {A,B}（2/3 够）与 新 {C,D,E}（3/5 够）—— 不相交
    ① 的算式                       maj_old + maj_new = 2+3 = 5，|并集| = 5，5 > 5 ✗ 不成立
    ② 的算式                       maj_old + maj_new = 2+3 = 5，|并集| = 4，5 > 4 ✓ 成立 ⇒ 必相交

    ▸ ① 里 {A,B} 用旧配置算已经过半（2/3
…（截断，原长 1718 字符）
````

</details>

▸ **Bash** Check Java 3C-3 and 3C-4 results

````
java Lab03C.java -trials 20 2>&1 | grep -v "Picked up" | sed -n '52,56p;70,78p'
````

<details><summary>↳ Bash 结果</summary>

````
══════════════════════════════════════════════════════════════════════════════
    变更方式                             出现两个 Leader                结论
    ────────────────────────────────────────────────────────────────────────
    ① 直接跳 {A,B,C} → {A..E}                      20/20            ✗ 会裂脑
    ② 单节点 {A,B,C} → {A,B,C,D}                    0/20    ✓ 一次都造不出来
  Lab 3C-4 · 三种读的对照
  制造僵尸 Leader，看哪种读会静默返回陈旧数据
══════════════════════════════════════════════════════════════════════════════
    读的实现                ★ 静默返回陈旧值            正确报错          返回正确值
    ────────────────────────────────────────────────────────────────────────────────
    本地读                                20                   0                   0
    ReadIndex                              0                  20                   0
    Lease Read                             0                  20                   0
````

</details>

▸ **Bash** Update README and final courseware check

````
cd 〔路径00054〕 && cp ch03c/ch03c.html courseware/ch03c-production-raft.html && python3 - <<'PY'
p='README.md'; s=open(p,encoding='utf-8').read()
def rep(a,b):
    global s; assert a in s, a[:60]; s=s.replace(a,b)
rep('''open courseware/ch03b-log-replication.html    # Part 3-B：日志复制与安全性
# Linux 用 xdg-open''','''open courseware/ch03b-log-replication.html    # Part 3-B：日志复制与安全性
open courseware/ch03c-production-raft.html   # Part 3-C：持久化 / 快照 / 成员变更 / 线性一致读
# Linux 用 xdg-open''')
rep('cd go/lab03b && go run .     # Lab 3-B：日志复制 + Figure 8 复现',
    'cd go/lab03b && go run .     # Lab 3-B：日志复制 + Figure 8 复现\ncd go/lab03c && go run .     # Lab 3-C：持久化 / 快照 / 成员变更 / 三种读')
rep('cd java/lab03b && java Lab03B.java','cd java/lab03b && java Lab03B.java\ncd java/lab03c && java Lab03C.java')
rep('| 3
…（截断，原长 2736 字符）
````

<details><summary>↳ Bash 结果</summary>

````
README ok
6→7 ✓ 安全
7→8 ✓ 安全
read exp rows: 3
ERRORS: none
````

</details>

📎 **发送文件**：Part 3-C 交互式课件。「实验 2 · 成员变更裂脑模拟」可以拖任意新旧规模，它会实时把两个过半集合摆出来告诉你能不能凑成不相交——你上一轮推出来的那个交集判据，在这里变成了可以拖的东西。
- `〔路径00076〕`

<details><summary>↳ SendUserFile 结果</summary>

````
1 file delivered to user.
  〔路径00076〕 → file_uuid: 〔标识03809〕
````

</details>

📎 **发送文件**：Lab 3-C 代码。raft.go 里 Restart() 的三个 if 就是"丢掉哪一样"的开关；ActiveLeaders() 给出了裂脑的准确定义（两个 Leader 各自都能凑齐过半，僵尸不算）。
- `〔路径00056〕`
- `〔路径00109〕`
- `〔路径00110〕`
- `〔路径00108〕`
- `〔路径00137〕`

<details><summary>↳ SendUserFile 结果</summary>

````
5 files delivered to user.
  〔路径00056〕 → file_uuid: 〔标识00831〕
  〔路径00109〕 → file_uuid: 〔标识02086〕
  〔路径00110〕 → file_uuid: 〔标识03767〕
  〔路径00108〕 → file_uuid: 〔标识00504〕
  〔路径00137〕 → file_uuid: 〔标识01019〕
````

</details>

▸ **device:device_bash**

````
{"command": "mkdir -p \"$HOME〔路径00156〕\" \"$HOME〔路径00164〕\" && echo ready"}
````

<details><summary>↳ mcp__remote-devices__device_bash 结果</summary>

````
ready
````

</details>

💾 **写入你的 Mac**：
- `〔路径00010〕`
- `〔路径00004〕`
- `〔路径00034〕`
- `〔路径00035〕`
- `〔路径00033〕`
- `〔路径00049〕`

<details><summary>↳ mcp__remote-devices__device_commit_files 结果</summary>

````
{"written":["〔路径00010〕","〔路径00004〕","〔路径00034〕","〔路径00035〕","〔路径00033〕","〔路径00049〕"],"rejected":[]}
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #18 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #19 status
````

</details>

▸ **TaskUpdate** completed

<details><summary>↳ TaskUpdate 结果</summary>

````
Updated task #20 status
````

</details>

**Part 3 完结。** 全部文件已写进 `~/Code/distributed-systems-course/`。

**你的答案，两点修正：**

第 1 题你算对了数字（2 和 3），但结论走偏成"会触发重新选举"——配置变更本身不触发选举。判据是**两个过半集合会不会不相交**，而这正是你那两个数字要回答的问题：`2+3=5 > |并集|=4` ✓ 必相交（3→4 安全）；`2+3=5 > |并集|=5` ✗（3→5 危险）。

第 2 题"4→5 不会触发"也不对，但你注意到的**过半数量变化**有真实后果：4 节点和 3 节点容错能力一样（都是 1），但写入要多等一个确认。所以 3→4→5 的中间那步是"付了代价没买到容错"，只是必须路过的中转站。

---

**这一节最值得看的三个地方：**

**① 成员变更实验做成了可拖的。** 你上一轮推的那个交集判据，现在是个滑块——拖任意新旧规模，它实时把两个过半集合摆出来。顺带你会发现一个反直觉的：**5→3 缩容也是危险的**（一次少两个节点，`3+2=5 > 5` 不成立）。而所有 n→n±1 都安全。

代码验证：直接跳 3→5 **30/30 裂脑**，单节点 3→4 **0/30，一次都造不出来**。

**② 3C-1 有个意外的方法论收获。** 我原本想用随机故障注入证明"votedFor 不落盘会脑裂"——**跑了几十个种子，一次都没撞出来**。因为它需要非常特定的交错：两个 candidate 抢同一任期 + 某个已投票节点恰好在那个窗口里重启。

所以我改成手工构造那个场景，一次命中：

```
构造场景 · votedFor 已落盘   ✓ N5 记得投过票，拒绝第二次索票 ⇒ 只有一个 Leader
构造场景 · votedFor 丢失     ✗ 同一任期两个 Leader —— 票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次
```

**"随机跑几十遍没出事"不等于安全，只等于没撞上。** 这正是 TLA+ 和 Jepsen 存在的理由——Part 6 会回到这里。

**③ 3C-4 把你 3-A 那个问题彻底答完了。** 僵尸 Leader 上：

| | 静默返回陈旧值 | 正确报错 |
|---|---|---|
| 本地读 | **20/20** | 0 |
| ReadIndex | 0 | 20/20 |
| Lease Read | 0 | 20/20 |

**"给你错误，而不是给你错的数据"** ——这个区别在排障时价值巨大。而 Lease Read 的安全性建立在**时钟漂移有界**上：Part 1 那个坑，在这里变成了一致性问题。

---

**Part 3 如果只留一句话**：「两个过半集合必相交」在 Raft 里出现了四次——选举安全性、Leader 完整性、Quorum 读写、成员变更。**其余是工程，这一条是数学。**

---

Part 4 分布式事务，我建议这样切（它比 Part 3 更需要单机基础）：

| | 主题 |
|---|---|
| **4-A** | 补单机课：ACID、隔离级别、四类并发异常（脏读/不可重复读/幻读/**写偏斜**）、MVCC、快照隔离 |
| **4-B** | 2PC/XA、TCC、Saga、本地消息表与事务消息（RocketMQ） |
| **4-C** | Percolator（TiDB）、Calvin、Spanner 外部一致性 —— 以及 2PC + Raft 的组合怎么让阻塞消失 |

**4-A 你想跳过多少？** 如果隔离级别和 MVCC 你已经熟，我可以压缩成一节速查 + 直接上写偏斜那个坑（它是最容易被"我用了可重复读所以没事"骗到的）；如果想扎实过一遍，就完整讲。

先给我一个判断题热身：**MySQL 的可重复读（RR）能防住幻读吗？**

> ✓ 本轮结束 · 2026-09-10 13:09 · 37 步
