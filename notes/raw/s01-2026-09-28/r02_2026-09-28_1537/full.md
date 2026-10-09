# s01 · r02 · 工具调用记录（手工整理）

| # | 工具 | 目的 / 关键参数 | 结果摘要 |
|---|---|---|---|
| 1 | TaskCreate ×4 / TaskUpdate | 任务清单：重读原文、核实事实、作答、记录并提交 | — |
| 2 | device_bash | 提取 ch04a §4.5、§4.6、§4.7 全文 | 确认原文讲的是锁的代价，以及 FOR UPDATE 能挡住写偏斜 |
| 3 | device_bash | 提取 ch03d §3.28 全文 | 得到冷启动路径 A/B、沙盒 D1/D2 的引用 |
| 4 | device_bash | `grep` ch03d 里的 D1/D2/undo/applied | 定位到 §3.24 的「截断永远不需要 undo」卡片和 JS 里的场景定义 |
| 5 | device_bash | 打印第 285–300 行与第 1170–1215 行 | 读到 D1/D2 的结论原文 |
| 6 | ToolSearch | 加载 WebSearch、WebFetch | — |
| 7 | WebFetch | etcd.io/docs/v3.5/faq/ | 确认「先 remove 再 add」和 strict-reconfig-check |
| 8 | WebFetch | etcd release-3.5 `server/etcdserver/server.go` | 确认 `Stop()` 先调用 `TransferLeadership()`；ErrNotEnoughStartedMembers；成员被移除后自行停机 |
| 9 | WebFetch | etcd-io/raft `log.go` | 确认 `maybeAppend` 冲突落在已提交范围会 panic；`commitTo(min(committed, lastnewi))` |
| 10 | WebFetch | etcd release-3.5 `membership/member.go` | 确认成员 ID = sha1(排序后的 peer URL + 集群名 + 可选时间) |
| 11 | WebFetch | etcd-io/raft `raft.go` | 确认快照恢复的检查、转移期间丢弃提案、MsgTimeoutNow、低任期消息回 MsgAppResp、取不到条目就发快照 |
| 12 | WebSearch ×3 | PD max-store-down-time、CockroachDB time_until_store_dead、PG FOR UPDATE 与聚合 | **失败**：组织没有开通网页搜索（403） |
| 13 | WebFetch | docs.pingcap.com PD 配置文档 | 确认 `max-store-down-time` 默认 30m |
| 14 | WebFetch | cockroachlabs.com 集群设置文档 | **失败**：robots.txt 超时 |
| 15 | WebFetch | postgresql.org SELECT 文档 | 确认锁定子句不能和聚合、GROUP BY、DISTINCT 一起用 |
| 16 | WebFetch | dev.mysql.com 隔离级别文档 | 确认关掉 autocommit 时，SERIALIZABLE 会把普通 SELECT 隐式改成 FOR SHARE |
| 17 | WebFetch | CockroachDB 仓库 `settings-for-tenants.txt` | 确认 `server.time_until_store_dead` 默认 5m0s |
| 18 | WebFetch | Students' Guide to Raft | 确认原文「The *if* here is crucial」 |
| 19 | WebFetch | etcd.io 恢复文档 | 确认 snapshot restore 会改写成员 ID 和集群 ID |
| 20 | WebFetch | etcd release-3.5 `etcdserver/raft.go` | 确认 maxSizePerMsg = 1MB、maxInflightMsgs = 512，以及 Leader 并行落盘的注释 |
| 21 | device_bash | 把第 2 轮逐字追加到 session 记录 | 15:37 |
| 22 | device_bash | questions.md 追加 Q020–Q026；更新 progress-review、notes/README 和 session 顶部小结 | — |
| 23 | device_bash | `git add notes` + `git commit` | 提交 `8a84731` |
