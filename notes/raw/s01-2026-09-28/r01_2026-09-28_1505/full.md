# s01 · r01 · 工具调用记录（手工整理）

> 工具说明：`device_bash` 在你电脑上的工作区运行，课程仓库挂载在 `$HOME〔路径00001〕`。本轮所有读写都在你电脑上完成，没有上传文件。

| # | 工具 | 目的 / 关键参数 | 结果摘要 |
|---|---|---|---|
| 1 | ToolSearch | 加载 device_bash、device_list_dir、get_device_info、SendUserMessage、TaskCreate、TaskUpdate | 已加载 |
| 2 | device_bash | 列出仓库全部文件、`du`、`git log` | 48 个文件，2.2M；只有一个提交 `c7f563c Publish sanitized distributed systems course` |
| 3 | device_bash | 读 README、.gitignore、目录、文件大小、`git status`、`git log --stat` | 读到学习路线表和 09-23 的勘误段；工作区干净 |
| 4 | device_bash | 用 python 提取 8 份课件的 h1–h3 标题 | 得到每章的小节结构 |
| 5 | device_bash | 统计「追问/勘误」等词，尝试提取 §3.31 | 首次提取为空（标题里含标签） |
| 6 | device_bash | `grep -n` 定位 ch03d 里的追问/勘误行 | 定位到 678–756 行 |
| 7 | device_bash | 提取 ch03d 第 676–756 行（§3.31 全文） | 得到勘误表 7 条和 7 个问答 |
| 8 | device_bash | `git remote`/`reflog`/分支；提取 3-D 引言；尝试提取 §3.15b | remote 是 GitHub `abietic/distributed-systems-course`；§3.15b 提取失败 |
| 9 | device_bash | 按 h2 行号提取 §3.15b 全文 | 成功 |
| 10 | device_bash | 搜「好问题/坦白/常见误解」等；提取 4-A 的「下一站」 | 找到 3.15b「这一节回答一个非常好的问题」 |
| 11 | device_bash | 统计各课件折叠题数量和开篇段落 | ch01、ch03b 开篇都引用了你的作答 |
| 12 | device_bash | 搜「你答/你说/你已经…」等提到学习者的句子 | 找到 10 余处互动痕迹 |
| 13 | device_bash | 提取 4-A 开篇问题、各章预告、0.10 路线图 | — |
| 14 | device_bash | 按修改时间列出所有文件（北京时间） | 得到 08-20 → 09-23 的时间线 |
| 15 | AskUserQuestion | 对话记录要不要纳入 git | 你选：纳入 git 一起提交 |
| 16 | TaskCreate ×4 / TaskUpdate | 建立并更新任务清单 | — |
| 17 | device_bash ×2 | 提取 ch00 §0.2 与 ch04a §4.1、§4.3 的正文 | 用于写各章回顾 |
| 18 | device_bash | 写 `notes/progress-review.md` | 124 行 |
| 19 | device_bash | 写 `notes/questions.md` | Q001–Q019 |
| 20 | device_bash | 写 `CLAUDE.md`、`notes/README.md`，在根目录 README 的目录树里补上 notes | README 增加 5 行 |
| 21 | device_bash | 写 `notes/sessions/2026-09-28-s01-history-recovery.md` 第 1 轮 | — |
| 22 | device_bash | `git add` + `git commit` | **失败**：git 删不掉自己的锁文件（`.git/index.lock`、`tmp_obj_*`），因为默认不允许删除文件 |
| 23 | device_request_delete_permission | 为课程文件夹申请删除权限（理由：清理 git 临时文件） | 你批准了 |
| 24 | device_bash | 删 `.git/index.lock` 和 6 个 `tmp_obj_*`，重新提交 | 提交 `2c704db` |
| 25 | device_bash | `git status`、`git show --stat` | 领先 origin 1 个提交，6 个文件 +303 行 |
| 26 | device_bash | 在 session 记录里补一条备注（删除权限），`git commit --amend` | 提交改为 `64948a1` |
