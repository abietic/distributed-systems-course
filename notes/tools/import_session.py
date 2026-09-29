#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
导入一份从 Claude 桌面应用本地缓存恢复出来的会话（events.jsonl / full.md / conversation.md）。

它做四件事：
  1. 脱敏：按本地的 notes/private/redaction-map.json 替换内部名称；课程目录以外的本地目录列表整段抹掉；
     环境元数据（可用工具 / MCP 服务器 / 插件 / 命令列表）移除；thinking 的加密签名移除。
     映射文件只在本地，不进 git。
  2. 按「对话轮」切分原始文件，写到 notes/raw/<code>-<sid>/rNN_<日期>_<时分>/
  3. 生成按天整理的逐字会话记录 notes/sessions/<日期>-<code>-rNN[-rMM].md
  4. 自检：脱敏后不得残留 forbidden 词；每一条用户消息、Claude 对你说的每一段话、
     每一道选择题和你的选择，都必须原样出现在逐字记录里。任何一项不通过就报错退出。

用法（在仓库根目录执行）：
  python3 notes/tools/import_session.py \
      --src notes/private/recovered-session-0185T6FH --code h1 --sid 0185T6FH \
      --gap 3660-3892
"""
import argparse, copy, datetime, json, os, re, sys, collections

CST = datetime.timezone(datetime.timedelta(hours=8))
ENV_REDACTED = "[已移除：环境元数据]"
LISTING_REDACTED = "[已脱敏：课程目录以外的本地目录/文件列表。未脱敏原文只保存在本地 notes/private/]"
REMINDER_RE = re.compile(r"<system-reminder>.*?</system-reminder>", re.S)


# ───────────────────────────── 脱敏 ─────────────────────────────
class Redactor:
    def __init__(self, path):
        m = json.load(open(path, encoding="utf-8"))
        self.pairs = m["replace"]
        self.markers = m.get("redact_tool_results_containing", [])
        self.forbidden = m.get("forbidden", [])

    def s(self, text):
        for a, b in self.pairs:
            text = text.replace(a, b)
        return text

    def deep(self, o):
        if isinstance(o, dict):
            return {k: ("[omitted]" if k == "signature" else self.deep(v)) for k, v in o.items()}
        if isinstance(o, list):
            return [self.deep(v) for v in o]
        if isinstance(o, str):
            return self.s(o)
        return o

    def hits(self, text):
        # 图片等 base64 数据里会偶然拼出 forbidden 词，检查前先去掉长 base64 串
        text = re.sub(r"[A-Za-z0-9+/=]{200,}", "", text)
        low = text.lower()
        return [w for w in self.forbidden if w.lower() in low]


def sanitize_events(events, R):
    """返回脱敏后的事件列表（顺序、序号、时间都不变）。"""
    out = []
    redacted_tool_ids = set()
    # 先找出需要整段抹掉的工具结果
    for e in events:
        if e["event_type"] == "user":
            c = e["payload"].get("message", {}).get("content")
            if isinstance(c, list):
                for b in c:
                    if b.get("type") == "tool_result":
                        raw = json.dumps(b.get("content"), ensure_ascii=False)
                        if any(mk in raw for mk in R.markers):
                            redacted_tool_ids.add(b.get("tool_use_id"))
    for e in events:
        e = copy.deepcopy(e)
        p = e["payload"]
        et = e["event_type"]
        if et == "system" and p.get("subtype") in ("init", "commands_changed"):
            for k in ("tools", "mcp_servers", "plugins", "skills", "slash_commands",
                      "terminal_slash_commands", "agents", "commands", "messaging_socket_path"):
                if k in p:
                    p[k] = ENV_REDACTED
        if et == "control_request":
            req = p.get("request", {})
            if req.get("subtype") in ("mcp_set_servers", "mcp_toggle"):
                for k in list(req.keys()):
                    if k not in ("subtype",):
                        req[k] = ENV_REDACTED
        if et == "control_response":
            resp = p.get("response", {})
            inner = resp.get("response")
            if isinstance(inner, dict) and ({"added", "removed", "errors"} & set(inner) or "account" in inner or "commands" in inner):
                resp["response"] = ENV_REDACTED
        if et == "user":
            c = p.get("message", {}).get("content")
            hit = False
            if isinstance(c, list):
                for b in c:
                    if b.get("type") == "tool_result" and b.get("tool_use_id") in redacted_tool_ids:
                        b["content"] = LISTING_REDACTED
                        hit = True
            if hit and "tool_use_result" in p:
                p["tool_use_result"] = LISTING_REDACTED
        out.append(R.deep(e))
    return out, redacted_tool_ids


def sanitize_md(text, R):
    # 课程目录以外的目录列表：整段 <details> 抹掉
    def repl(m):
        body = m.group(0)
        if any(mk in body for mk in R.markers):
            head = body.split("\n", 1)[0]
            return head + "\n\n" + LISTING_REDACTED + "\n\n</details>"
        return body
    text = re.sub(r"<details><summary>↳.*?</details>", repl, text, flags=re.S)
    return R.s(text)


PUBLIC_PATH_RE = re.compile(r'/(?:Users|home|tmp|var|Applications|Volumes|workspace|mnt|opt|private)/[^\s`"\'<>\[\](),;]+')
PUBLIC_UUID_RE = re.compile(r'\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b', re.I)
PUBLIC_IP_RE = re.compile(r'(?<!\d)(?:10\.(?:\d{1,3}\.){2}\d{1,3}|192\.168\.(?:\d{1,3}\.)\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.(?:\d{1,3}\.)\d{1,3})(?!\d)')


def public_sanitize(text):
    """Remove local metadata from the public session record, including tool summaries."""
    text = PUBLIC_PATH_RE.sub("〔本地路径〕", text)
    text = PUBLIC_UUID_RE.sub("〔会话标识〕", text)
    text = PUBLIC_IP_RE.sub("〔内网地址〕", text)
    return "\n".join(line.rstrip() for line in text.split("\n"))


# ───────────────────────────── 解析 ─────────────────────────────
def cst(ts):
    return datetime.datetime.fromisoformat(ts.replace("Z", "+00:00")).astimezone(CST)


def user_text(e):
    """真实的用户输入（去掉 system-reminder）；不是则返回 None。"""
    if e["event_type"] != "user" or e["source"] != "client":
        return None
    c = e["payload"].get("message", {}).get("content")
    if isinstance(c, list):
        c = "\n".join(b.get("text", "") for b in c if b.get("type") == "text")
    if not isinstance(c, str):
        return None
    t = REMINDER_RE.sub("", c).strip()
    return t or None


def demote(md, by=2):
    """把 Claude 回答里的 Markdown 标题降两级，避免打乱记录的目录；代码块内不动。"""
    out, fence = [], False
    for line in md.split("\n"):
        if line.lstrip().startswith("```") or line.lstrip().startswith("~~~"):
            fence = not fence
        if not fence and re.match(r"^#{1,4} ", line):
            line = "#" * by + line
        out.append(line)
    return "\n".join(out)


def quote(text):
    return "\n".join(("> " + l) if l.strip() else ">" for l in text.split("\n"))


def one_line(s, n=160):
    s = (s or "").strip().split("\n", 1)[0]
    return s if len(s) <= n else s[: n - 1] + "…"


def fmt_ask(inp):
    lines = []
    for q in inp.get("questions", []):
        multi = "（可多选）" if q.get("multiSelect") else ""
        lines.append(f"**❓ {q.get('question','')}**{multi}")
        for o in q.get("options", []):
            d = o.get("description", "")
            lines.append(f"  - {o.get('label','')}：{d}" if d else f"  - {o.get('label','')}")
        lines.append("")
    return "\n".join(lines).rstrip()


def fmt_answer(tur, content):
    if isinstance(tur, dict) and isinstance(tur.get("answers"), dict):
        lines = []
        for q, a in tur["answers"].items():
            if isinstance(a, list):
                a = "、".join(a)
            lines.append(f"- {q}\n  → **{a}**")
        ann = tur.get("annotations")
        if ann:
            lines.append("- 备注：" + json.dumps(ann, ensure_ascii=False))
        return "\n".join(lines)
    if isinstance(content, list):
        content = "\n".join(b.get("text", "") for b in content if isinstance(b, dict))
    return quote(str(content))


VISIBLE = {"SendUserFile", "mcp__remote-devices__device_commit_files", "mcp__remote-devices__create_artifact"}


def fmt_op(name, inp, err):
    short = name.replace("mcp__remote-devices__", "device:")
    if name == "SendUserFile":
        cap = inp.get("caption", "")
        files = "、".join(f"`{f}`" for f in inp.get("files", []))
        s = f"📎 **发送文件**：{cap}　{files}"
    elif name == "mcp__remote-devices__device_commit_files":
        files = "、".join(f"`{f.get('devicePath','')}`" for f in inp.get("files", []))
        s = f"💾 **写入你的电脑**：{files}"
    elif name == "mcp__remote-devices__create_artifact":
        s = f"🗂 **保存为桌面 artifact** `{inp.get('id','')}`：{inp.get('description','')}"
    elif name in ("Bash", "mcp__remote-devices__device_bash"):
        d = inp.get("description")
        s = f"`{short}` " + (f"{d}：" if d else "") + f"`{one_line(inp.get('command',''), 140)}`"
    elif name == "Read":
        s = f"`Read` `{inp.get('file_path','')}`"
    elif name == "Write" or name == "Edit":
        s = f"`{name}` `{inp.get('file_path','')}`"
    elif name == "WebFetch":
        s = f"`WebFetch` {inp.get('url','')} —— {one_line(inp.get('prompt',''), 120)}"
    elif name == "WebSearch":
        s = f"`WebSearch` {inp.get('query','')}"
    elif name == "Skill":
        s = f"`Skill` {inp.get('skill','')}：{one_line(inp.get('args',''), 120)}"
    elif name == "TaskCreate":
        s = f"`TaskCreate` {inp.get('subject','')}"
    elif name == "TaskUpdate":
        s = f"`TaskUpdate` #{inp.get('taskId','')} → {inp.get('status', '') or '（更新）'}"
    elif name == "ToolSearch":
        s = f"`ToolSearch` {inp.get('query','')}"
    elif name == "mcp__remote-devices__device_list_dir":
        s = f"`{short}` `{inp.get('path','')}`"
    else:
        s = f"`{short}` {one_line(json.dumps(inp, ensure_ascii=False), 140)}"
    if err:
        s += f"　**（失败：{one_line(err, 100)}）**"
    return s


def build_rounds(events):
    starts = [int(e["sequence_num"]) for e in events if user_text(e)]
    rounds = []
    for i, s in enumerate(starts):
        lo = 1 if i == 0 else s
        hi = starts[i + 1] - 1 if i + 1 < len(starts) else None
        rounds.append({"no": i + 1, "start_seq": s, "lo": lo, "hi": hi})
    return rounds


def render_round(rd, evs, tool_uses, tool_results):
    """返回 (markdown, 应出现的原文片段列表)。"""
    md, must = [], []
    ops = []  # 当前累计的操作
    claude_texts = []

    def flush_ops():
        if not ops:
            return
        vis = [o for o in ops if o[1] in VISIBLE]
        rest = [o for o in ops if o[1] not in VISIBLE]
        for t, name, inp, err in vis:
            md.append(f"- {t:%H:%M} {fmt_op(name, inp, err)}")
        if vis:
            md.append("")
        if rest:
            t0, t1 = rest[0][0], rest[-1][0]
            md.append(f"<details><summary>操作明细 · {t0:%H:%M}–{t1:%H:%M} · {len(rest)} 项</summary>\n")
            for t, name, inp, err in rest:
                md.append(f"- {t:%H:%M} {fmt_op(name, inp, err)}")
            md.append("\n</details>\n")
        ops.clear()

    for e in evs:
        t = cst(e["created_at"])
        et, p = e["event_type"], e["payload"]
        ut = user_text(e)
        if ut:
            flush_ops()
            md.append(f"### 你 · {t:%m-%d %H:%M} · seq {e['sequence_num']}\n")
            md.append(quote(ut) + "\n")
            must.extend(l.strip() for l in ut.split("\n") if l.strip())
            continue
        if et == "assistant":
            for b in p["message"]["content"]:
                bt = b.get("type")
                if bt == "text" and b.get("text", "").strip():
                    flush_ops()
                    txt = b["text"].strip()
                    md.append(f"### Claude · {t:%m-%d %H:%M}\n")
                    md.append(demote(txt) + "\n")
                    must.append(demote(txt)); claude_texts.append(txt)
                elif bt == "tool_use":
                    name, inp = b["name"], b.get("input", {})
                    if name == "AskUserQuestion":
                        flush_ops()
                        md.append(f"### Claude · {t:%m-%d %H:%M} · 选择题\n")
                        a = fmt_ask(inp); md.append(a + "\n"); must.append(a)
                        res = tool_results.get(b["id"])
                        if res:
                            rt = cst(res["created_at"])
                            md.append(f"### 你的选择 · {rt:%m-%d %H:%M}\n")
                            ans = fmt_answer(res.get("tur"), res.get("content"))
                            md.append(ans + "\n"); must.append(ans)
                    elif name == "SendUserMessage":
                        flush_ops()
                        msg = (inp.get("message") or "").strip()
                        md.append(f"### Claude · {t:%m-%d %H:%M} · 消息\n")
                        md.append(demote(msg) + "\n")
                        must.append(demote(msg)); claude_texts.append(msg)
                        for att in inp.get("attachments", []) or []:
                            md.append(f"- 附件：`{att}`")
                    else:
                        res = tool_results.get(b["id"], {})
                        err = None
                        if res.get("is_error"):
                            c = res.get("content")
                            if isinstance(c, list):
                                c = " ".join(x.get("text", "") for x in c if isinstance(x, dict))
                            err = str(c)
                        ops.append((t, name, inp, err))
        elif et == "user" and e["source"] == "worker":
            c = p.get("message", {}).get("content")
            if isinstance(c, list):
                for b in c:
                    if b.get("type") != "text":
                        continue
                    tx = b.get("text", "")
                    if tx.startswith("[Request interrupted"):
                        flush_ops()
                        md.append(f"### （你中断了这一轮 · {t:%m-%d %H:%M}）\n")
                    elif tx.startswith("This session is being continued"):
                        flush_ops()
                        md.append(f"### （上下文自动压缩 · {t:%m-%d %H:%M}）\n")
                        md.append("对话太长，系统把此前的内容压缩成下面这份摘要，之后的回答基于它继续。摘要原文：\n")
                        md.append("<details><summary>压缩摘要原文（英文）</summary>\n\n" + tx.strip() + "\n\n</details>\n")
                        must.append(tx.strip())
                    elif tx.startswith("[Image:"):
                        ops.append((t, "（查看截图）", {"说明": one_line(tx, 80)}, None))
                    elif tx.startswith("Base directory for this skill"):
                        ops.append((t, "（载入技能说明）", {"说明": one_line(tx, 80)}, None))
                    else:
                        flush_ops()
                        md.append(f"### （系统插入的文本 · {t:%m-%d %H:%M}）\n")
                        md.append(quote(tx) + "\n"); must.append(tx.strip())
        elif et == "result":
            r = (p.get("result") or "").strip()
            if r and r not in claude_texts:
                flush_ops()
                md.append(f"### Claude · {t:%m-%d %H:%M} · 本轮结果\n")
                md.append(demote(r) + "\n"); must.append(demote(r))
        elif et == "prompt_suggestion":
            flush_ops()
            sg = p.get("suggestion", "")
            md.append(f"> 💡 界面给出的建议回复：「{sg}」\n")
            must.append(sg)
        elif et == "system" and p.get("subtype") == "compact_boundary":
            ops.append((t, "（上下文压缩边界）", p.get("compact_metadata", {}), None))
    flush_ops()
    return "\n".join(md), must


# ───────────────────────────── 主流程 ─────────────────────────────
def split_md_by_rounds(text, round_starts):
    """按每轮第一条用户消息的标题切 full.md / conversation.md。"""
    lines = text.split("\n")
    idx, pos = [], 0
    for t in round_starts:
        pat = f"## 👤 你 · {t:%Y-%m-%d %H:%M}"
        found = None
        for i in range(pos, len(lines)):
            if lines[i].strip() == pat:
                found = i; break
        if found is None:
            raise SystemExit(f"在 md 里找不到轮次开头：{pat}")
        idx.append(found); pos = found + 1
    parts = []
    for k, i in enumerate(idx):
        lo = 0 if k == 0 else i
        hi = idx[k + 1] if k + 1 < len(idx) else len(lines)
        parts.append("\n".join(lines[lo:hi]).rstrip() + "\n")
    return parts


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--src", required=True)
    ap.add_argument("--code", required=True, help="会话代号，例如 h1")
    ap.add_argument("--sid", required=True, help="会话 ID 缩写，例如 0185T6FH")
    ap.add_argument("--map", default="notes/private/redaction-map.json")
    ap.add_argument("--repo", default=".")
    ap.add_argument("--gap", default="", help="已知缺失的序号段，例如 3660-3892")
    a = ap.parse_args()

    R = Redactor(a.map)
    raw_events = [json.loads(l) for l in open(os.path.join(a.src, "events.jsonl"), encoding="utf-8")]
    raw_events.sort(key=lambda e: int(e["sequence_num"]))
    events, red_ids = sanitize_events(raw_events, R)

    tool_uses, tool_results = {}, {}
    for e in events:
        if e["event_type"] == "assistant":
            for b in e["payload"]["message"]["content"]:
                if b.get("type") == "tool_use":
                    tool_uses[b["id"]] = b
        if e["event_type"] == "user":
            c = e["payload"].get("message", {}).get("content")
            if isinstance(c, list):
                for b in c:
                    if b.get("type") == "tool_result":
                        tool_results[b["tool_use_id"]] = {"content": b.get("content"), "is_error": b.get("is_error"),
                                                          "tur": e["payload"].get("tool_use_result"), "created_at": e["created_at"]}

    gap_lo = gap_hi = None
    if a.gap:
        gap_lo, gap_hi = map(int, a.gap.split("-"))

    rounds = build_rounds(events)
    by_seq = {int(e["sequence_num"]): e for e in events}
    seqs = sorted(by_seq)
    # 每轮包含的事件；缺口之后的尾巴单独放
    tail = [e for s, e in by_seq.items() if gap_hi and s > gap_hi]
    for i, rd in enumerate(rounds):
        hi = rd["hi"] if rd["hi"] is not None else (gap_lo - 1 if gap_lo else seqs[-1])
        rd["hi"] = hi
        rd["events"] = [by_seq[s] for s in seqs if rd["lo"] <= s <= hi]
        rd["t0"] = cst(by_seq[rd["start_seq"]]["created_at"])
        content_evs = [e for e in rd["events"] if e["event_type"] in ("assistant", "user", "result", "prompt_suggestion")]
        rd["t_last"] = cst((content_evs or rd["events"])[-1]["created_at"])
        rd["tag"] = f"r{rd['no']:02d}_{rd['t0']:%Y-%m-%d_%H%M}"

    raw_dir = os.path.join(a.repo, "notes", "raw", f"{a.code}-{a.sid}")
    ses_dir = os.path.join(a.repo, "notes", "sessions")
    os.makedirs(raw_dir, exist_ok=True)

    # 1) 原始底稿：按轮切分
    full = sanitize_md(open(os.path.join(a.src, "full.md"), encoding="utf-8").read(), R)
    conv = sanitize_md(open(os.path.join(a.src, "conversation.md"), encoding="utf-8").read(), R)
    starts_t = [rd["t0"] for rd in rounds]
    full_parts = split_md_by_rounds(full, starts_t)
    conv_parts = split_md_by_rounds(conv, starts_t)
    written = []
    for rd, fp, cp in zip(rounds, full_parts, conv_parts):
        d = os.path.join(raw_dir, rd["tag"]); os.makedirs(d, exist_ok=True)
        with open(os.path.join(d, "events.jsonl"), "w", encoding="utf-8") as f:
            for e in rd["events"]:
                f.write(json.dumps(e, ensure_ascii=False) + "\n")
        open(os.path.join(d, "full.md"), "w", encoding="utf-8").write(fp)
        open(os.path.join(d, "conversation.md"), "w", encoding="utf-8").write(cp)
        written += [os.path.join(d, x) for x in ("events.jsonl", "full.md", "conversation.md")]
    if tail:
        t0 = cst(tail[0]["created_at"])
        d = os.path.join(raw_dir, f"after-gap_{t0:%Y-%m-%d_%H%M}"); os.makedirs(d, exist_ok=True)
        with open(os.path.join(d, "events.jsonl"), "w", encoding="utf-8") as f:
            for e in tail:
                f.write(json.dumps(e, ensure_ascii=False) + "\n")
        written.append(os.path.join(d, "events.jsonl"))

    # 2) 逐字记录：按天
    by_day = collections.OrderedDict()
    for rd in rounds:
        by_day.setdefault(rd["t0"].date(), []).append(rd)
    session_files, must_all = [], []
    for day, rds in by_day.items():
        rr = f"r{rds[0]['no']:02d}" + (f"-r{rds[-1]['no']:02d}" if len(rds) > 1 else "")
        name = f"{day:%Y-%m-%d}-{a.code}-{rr}.md"
        body = [f"# 历史会话 {a.code}（{a.sid}）· {day:%Y-%m-%d} · {rr.upper()}\n",
                "> 这份记录由 `notes/tools/import_session.py` 从恢复出的原始事件自动生成，**不要手改**；要补充说明请写在 `notes/README.md` 或 `notes/progress-review.md`。",
                "> - 公开记录保留问答内容，内部名称、本地路径、私有 IP 和会话标识已替换。",
                "> - Claude 回答里的 Markdown 标题降了两级，以免打乱这份记录的目录。",
                "> - 工具调用压缩成「操作明细」；完整输入输出只在本地原始底稿里。\n",
                "| 轮次 | 时间（北京时间） | 事件序号 | 原始底稿 |", "|---|---|---|---|"]
        for rd in rds:
            body.append(f"| R{rd['no']:02d} | {rd['t0']:%m-%d %H:%M} → {rd['t_last']:%m-%d %H:%M} | {rd['lo']}–{rd['hi']} | 本地保存 |")
        body.append("")
        for rd in rds:
            text, must = render_round(rd, rd["events"], tool_uses, tool_results)
            body.append("---\n")
            body.append(f"## R{rd['no']:02d} · {rd['t0']:%m-%d %H:%M} · seq {rd['lo']}–{rd['hi']}\n")
            body.append(public_sanitize(text))
            must_all.append((name, [public_sanitize(m) for m in must]))
        path = os.path.join(ses_dir, name)
        open(path, "w", encoding="utf-8").write("\n".join(body).rstrip() + "\n")
        session_files.append(path)

    # 2b) 缺口占位 + 底稿索引
    if gap_lo:
        before = by_seq.get(gap_lo - 1); after = by_seq.get(gap_hi + 1)
        tb = cst(before["created_at"]) if before else None
        ta = cst(after["created_at"]) if after else None
        gname = f"{tb:%Y-%m-%d}-{a.code}-gap-seq{gap_lo}-{gap_hi}.md"
        g = [f"# 历史会话 {a.code}（{a.sid}）· 缺失片段 · seq {gap_lo}–{gap_hi}\n",
             "> 这一段事件不在恢复出的本地缓存里，**内容暂缺**。以后如果恢复出这一段，用 `notes/tools/import_session.py` 重新导入即可补上。\n",
             f"- 缺口之前的最后一条事件：seq {gap_lo - 1}，{tb:%Y-%m-%d %H:%M}（北京时间）" if tb else "- 缺口之前：无",
             f"- 缺口之后的第一条事件：seq {gap_hi + 1}，{ta:%Y-%m-%d %H:%M}（北京时间）" if ta else "- 缺口之后：无",
             "- 缺口之后只剩 2 条客户端控制消息，没有对话内容。\n",
             "## 从课件和仓库推断，这段时间里大概发生了什么\n",
             "- 学完 3-D 之后你提了 15 个追问，其中几个戳中了课件原版的错误；课件 3-D 为此新增了「§3.31 追问篇（含勘误表）」。",
             "- 对应的勘误改动了 3-C / 3-D / 4-A 的课件和 Lab 3-B / 3-C / 3-D（仓库里这些文件的修改时间是 09-23 20:40）。",
             "- 09-23 21:21 仓库以「Publish sanitized distributed systems course」首次提交并推到 GitHub。",
             "- 注意：09-22 09:51 之后到 09-28 之间的事件也不在恢复数据里。09-23 的勘误和发布可能发生在这一段，也可能在缺口里。\n",
             "目前能还原出的追问主题见 `notes/questions.md` 的 Q004–Q018。"]
        open(os.path.join(ses_dir, gname), "w", encoding="utf-8").write("\n".join(g) + "\n")
        session_files.append(os.path.join(ses_dir, gname))
    idx = [f"# 原始底稿 · 历史会话 {a.code}（{a.sid}）\n",
           "从 Claude 桌面应用本地缓存恢复出的原始数据，已脱敏，按「对话轮」切分。由 `notes/tools/import_session.py` 生成，不要手改。\n",
           "每个文件夹里有三份文件：",
           "- `events.jsonl`：原始事件流（一行一个事件，含全部工具调用的输入和输出）",
           "- `full.md`：恢复工具生成的完整可读版（含工具调用与结果）",
           "- `conversation.md`：恢复工具生成的对话版（只有双方的话）\n",
           "脱敏做了这些事：内部系统名 / 表名 / 邮箱 / 用户名 / 组织 ID 换成〔…〕占位符；课程目录以外的本地目录列表整段抹掉；",
           "环境元数据（可用工具、MCP 服务器、插件、命令列表）移除；thinking 的加密签名移除。未脱敏的原文只在本地 `notes/private/`（不进 git）。\n",
           "| 文件夹 | 事件序号 | 事件数 | 时间（北京时间） |", "|---|---|---|---|"]
    for rd in rounds:
        idx.append(f"| [{rd['tag']}/]({rd['tag']}/) | {rd['lo']}–{rd['hi']} | {len(rd['events'])} | {rd['t0']:%m-%d %H:%M} → {rd['t_last']:%m-%d %H:%M} |")
    if gap_lo:
        idx.append(f"| **缺失** | {gap_lo}–{gap_hi} | — | {tb:%m-%d %H:%M} → {ta:%m-%d %H:%M} |")
    if tail:
        t0 = cst(tail[0]["created_at"])
        idx.append(f"| [after-gap_{t0:%Y-%m-%d_%H%M}/](after-gap_{t0:%Y-%m-%d_%H%M}/) | {tail[0]['sequence_num']}–{tail[-1]['sequence_num']} | {len(tail)} | {t0:%m-%d %H:%M}（只有客户端控制消息） |")
    open(os.path.join(raw_dir, "README.md"), "w", encoding="utf-8").write("\n".join(idx) + "\n")
    written.append(os.path.join(raw_dir, "README.md"))

    # 3) 自检
    problems = []
    for path in session_files + written:
        txt = open(path, encoding="utf-8").read()
        h = R.hits(txt)
        if h:
            problems.append(f"脱敏残留 {path}: {sorted(set(h))}")
    contents = {os.path.basename(p): open(p, encoding="utf-8").read() for p in session_files}
    n_checked = 0
    for name, must in must_all:
        for m in must:
            n_checked += 1
            if m not in contents[name]:
                problems.append(f"逐字记录缺内容 {name}: {one_line(m, 80)}")
    # 原始事件条数守恒
    n_out = sum(len(rd["events"]) for rd in rounds) + len(tail)
    if n_out != len(events):
        problems.append(f"事件条数不守恒：输入 {len(events)}，输出 {n_out}")

    report = {
        "events_in": len(events), "events_out": n_out,
        "seq_range": [seqs[0], seqs[-1]], "gap": a.gap,
        "rounds": [{"no": rd["no"], "tag": rd["tag"], "seq": [rd["lo"], rd["hi"]], "events": len(rd["events"])} for rd in rounds],
        "tail_events": len(tail), "redacted_tool_results": len(red_ids),
        "session_files": [os.path.basename(p) for p in session_files],
        "fragments_checked": n_checked, "problems": problems,
    }
    print(json.dumps(report, ensure_ascii=False, indent=1))
    if problems:
        sys.exit(1)


if __name__ == "__main__":
    main()
