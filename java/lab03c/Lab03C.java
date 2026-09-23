/*
 * Lab 3-C · 持久化、快照、成员变更与线性一致读（Java 版）
 *
 * 运行：  cd java/lab03c && java Lab03C.java
 * 调参：  java Lab03C.java -lose votedfor | term | log
 *
 * 需要 JDK 17+。配套课件：courseware/ch03c-production-raft.html
 */
import java.util.*;

public class Lab03C {

    static int N = 5;
    static long DELAY = 20, JITTER = 10, SEED = 42;
    static double LOSS = 0.0;
    static boolean VERBOSE = false;
    static String LOSE = "";
    static int TRIALS = 30;
    static final long TICK = 5, HEARTBEAT = 100;
    static final int NO_VOTE = -1;

    // ══════════════════════════════════════════════════════════════════════
    // 确定性网络
    // ══════════════════════════════════════════════════════════════════════
    enum MType { REQUEST_VOTE, VOTE_REPLY, APPEND, APPEND_REPLY, SNAPSHOT, SNAPSHOT_REPLY }

    record Entry(int index, int term, String cmd) {}

    static class Msg {
        int from, to; MType type; int term;
        int lastLogIndex, lastLogTerm; boolean granted;                 // RequestVote
        int prevLogIndex, prevLogTerm; List<Entry> entries = List.of(); // AppendEntries
        int leaderCommit;
        boolean success; int matchIndex, conflictTerm, conflictIndex;   // 回复
        int lastIncludedIndex, lastIncludedTerm;                        // InstallSnapshot
        Map<String,String> snapshot = Map.of();
        long deliverAt; int seq;
        Msg(int f, int t, MType ty, int tm) { from=f; to=t; type=ty; term=tm; }
    }

    static class Network {
        final int n; final Random rnd; final long delay, jitter; double loss;
        final int[] group; final boolean[] down;
        final List<Msg> queue = new ArrayList<>();
        long now = 0; int seq = 0;

        Network(int n, long delay, long jitter, double loss, long seed) {
            this.n=n; this.delay=delay; this.jitter=jitter; this.loss=loss;
            rnd=new Random(seed); group=new int[n]; down=new boolean[n];
        }
        boolean isDown(int i) { return down[i]; }
        void kill(int i) { down[i]=true; }
        void revive(int i) { down[i]=false; }
        void partition(int[] g) { System.arraycopy(g,0,group,0,n); }
        void heal() { Arrays.fill(group,0); }
        boolean sameGroup(int a, int b) { return group[a] == group[b]; }

        void send(Msg m) {
            if (down[m.from] || down[m.to] || group[m.from] != group[m.to]) return;
            if (rnd.nextDouble() < loss) return;
            long jit = jitter > 0 ? rnd.nextLong(2*jitter) - jitter : 0;
            m.deliverAt = now + Math.max(1, delay + jit);
            m.seq = seq++;
            queue.add(m);
        }
        List<Msg> advance(long dt) {
            now += dt;
            queue.sort(Comparator.<Msg>comparingLong(m -> m.deliverAt).thenComparingInt(m -> m.seq));
            List<Msg> due = new ArrayList<>();
            queue.removeIf(m -> {
                if (m.deliverAt > now) return false;
                if (!down[m.to] && group[m.from] == group[m.to]) due.add(m);
                return true;
            });
            return due;
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // Raft 节点
    // ══════════════════════════════════════════════════════════════════════
    enum State { FOLLOWER, CANDIDATE, LEADER }

    static class Node {
        final int id;
        State state = State.FOLLOWER;
        int term = 0, votedFor = NO_VOTE;
        Set<Integer> votes = new HashSet<>();
        long timer = 0, timeout = 0, hb = 0;

        List<Entry> log = new ArrayList<>(List.of(new Entry(0,0,"")));  // 0 号是哨兵/快照代表
        int commitIndex = 0, lastApplied = 0;
        final Map<Integer,String> applied = new HashMap<>();
        int[] nextIndex, matchIndex;

        // —— Part 3-C ——
        Persisted disk = new Persisted();          // 模拟磁盘：只有这三样能挺过重启
        int snapIdx = 0, snapTerm = 0;
        Map<String,String> snapshot = new HashMap<>();
        final Map<String,String> kv = new HashMap<>();   // 状态机
        Set<Integer> config = new HashSet<>();           // 本节点认为的集群成员
        long leaseTill = 0; int ackTicks = 0;

        Node(int id, int n) {
            this.id=id; nextIndex=new int[n]; matchIndex=new int[n];
            for (int j=0;j<n;j++) config.add(j);
        }
        Entry last() { return log.get(log.size()-1); }

        /** 按【逻辑 index】取条目：有快照之后数组下标不再等于日志 index。 */
        Entry entryAt(int idx) {
            int base = log.get(0).index(), off = idx - base;
            if (off < 0 || off >= log.size()) return new Entry(-1,-1,"");
            return log.get(off);
        }
        Entry at(int i) { Entry e = entryAt(i); return e.index()==i ? e : null; }
        int majorityOf() { return config.size()/2 + 1; }

        /** persist 必须在响应任何 RPC 之前调用 —— 论文的要求就是这么严格。 */
        void persist() { disk = new Persisted(term, votedFor, new ArrayList<>(log)); }
    }

    /** Raft 论文 Figure 2 的「Persistent state on all servers」——就这三样。 */
    static class Persisted {
        int currentTerm = 0, votedFor = NO_VOTE;
        List<Entry> log = new ArrayList<>(List.of(new Entry(0,0,"")));
        Persisted() {}
        Persisted(int t, int v, List<Entry> l) { currentTerm=t; votedFor=v; log=l; }
    }

    static class Cluster {
        final List<Node> nodes = new ArrayList<>();
        final Network net; final Random rnd; final long base, window;
        boolean naiveCommit = false, slowBackoff = false, noLogCheck = false;
        // —— Part 3-C：故意"丢失"某一样持久化状态 ——
        boolean loseTerm = false, loseVote = false, loseLog = false;
        int snapThreshold = 0;
        long leaseMs = 0;
        boolean twoLeadersEver = false;
        int snapshots = 0, installSnaps = 0;

        /** 上帝视角：任何节点提交一条日志就登记在这里，用于安全性断言。 */
        final Map<Integer,Entry> globalCommitted = new HashMap<>();
        int elections=0, splits=0, rejects=0, appends=0, cmdSeq=0;
        final List<String> violations = new ArrayList<>();

        Cluster(int n, Network net, long base, long window, long seed) {
            this.net=net; this.base=base; this.window=window; rnd=new Random(seed+999);
            for (int i=0;i<n;i++) { Node nd=new Node(i,n); nd.timeout=newTimeout(); nd.persist(); nodes.add(nd); }
        }
        long newTimeout() { return window<=0 ? base : base + rnd.nextLong(window); }
        int majority() { return nodes.size()/2 + 1; }

        void violate(String f, Object... a) {
            String v = String.format(f, a);
            if (!violations.contains(v) && violations.size() < 8) violations.add(v);
        }

        // ─── 选举 ───────────────────────────────────────────────────────
        void stepDown(Node n, int term) {
            n.term=term; n.state=State.FOLLOWER; n.votedFor=NO_VOTE;
            n.votes.clear(); n.timer=0; n.timeout=newTimeout();
            n.persist();   // ★ 响应任何 RPC 之前必须落盘
        }
        void startElection(Node n) {
            n.term++; n.state=State.CANDIDATE; n.votedFor=n.id;
            n.votes=new HashSet<>(List.of(n.id));
            n.timer=0; n.timeout=newTimeout(); elections++;
            n.persist();   // ★ 先落盘，再发 RequestVote
            Entry last=n.last();
            for (int j=0;j<nodes.size();j++) if (j!=n.id) {
                Msg m=new Msg(n.id,j,MType.REQUEST_VOTE,n.term);
                m.lastLogIndex=last.index(); m.lastLogTerm=last.term();
                net.send(m);
            }
        }
        /** 「candidate 的日志至少和我一样新」：先比 term，再比长度。 */
        static boolean upToDate(int candTerm, int candIdx, Entry mine) {
            return candTerm != mine.term() ? candTerm > mine.term() : candIdx >= mine.index();
        }
        void becomeLeader(Node n) {
            n.state=State.LEADER; n.hb=HEARTBEAT;
            for (int j=0;j<nodes.size();j++) { n.nextIndex[j]=n.last().index()+1; n.matchIndex[j]=0; }
            n.matchIndex[n.id]=n.last().index();
            // Leader 完整性断言：新 Leader 必须包含所有已提交的日志
            globalCommitted.forEach((idx,e) -> {
                Entry got=n.at(idx);
                if (got==null || got.term()!=e.term() || !got.cmd().equals(e.cmd()))
                    violate("【Leader 完整性被破坏】N%d 在 term=%d 当选，但它缺少已提交的 index=%d (term=%d, cmd=%s)",
                            n.id+1, n.term, idx, e.term(), e.cmd());
            });
        }

        // ─── 日志复制 ────────────────────────────────────────────────────
        void sendAppend(Node n, int to) {
            int ni = Math.max(1, n.nextIndex[to]);
            if (ni <= n.snapIdx) { sendSnapshotOf(this, n, to); return; }   // ★ 日志已压缩 ⇒ 改发快照
            Entry prev = n.at(ni-1);
            Msg m = new Msg(n.id, to, MType.APPEND, n.term);
            m.prevLogIndex = prev.index(); m.prevLogTerm = prev.term();
            int base = n.log.get(0).index();
            m.entries = ni <= n.last().index()
                    ? new ArrayList<>(n.log.subList(ni-base, n.log.size())) : List.of();
            m.leaderCommit = n.commitIndex;
            net.send(m);
        }

        void handleAppend(Node n, Msg m) {
            if (m.term < n.term) {
                Msg r=new Msg(n.id,m.from,MType.APPEND_REPLY,n.term); r.success=false; net.send(r); return;
            }
            n.term=m.term; n.state=State.FOLLOWER; n.votedFor=m.from;
            n.timer=0; n.timeout=newTimeout();

            // ★ 一致性检查
            Entry prev = n.at(m.prevLogIndex);
            if (prev == null || prev.term() != m.prevLogTerm) {
                rejects++;
                Msg r = new Msg(n.id, m.from, MType.APPEND_REPLY, n.term);
                r.success = false;
                if (prev == null) { r.conflictTerm = -1; r.conflictIndex = n.last().index()+1; }
                else {
                    r.conflictTerm = prev.term();
                    int i = m.prevLogIndex;
                    while (i > n.snapIdx+1 && n.entryAt(i-1).term() == prev.term()) i--;
                    r.conflictIndex = i;
                }
                net.send(r); return;
            }
            for (int k=0;k<m.entries.size();k++) {
                int idx = m.prevLogIndex + 1 + k;
                Entry e = m.entries.get(k), cur = n.at(idx);
                if (cur != null) {
                    if (cur.term() == e.term()) continue;
                    int base = n.log.get(0).index();
                    while (n.log.size() > idx-base) n.log.remove(n.log.size()-1);  // ★ 冲突 ⇒ 截断
                }
                n.log.add(e);
            }
            appends++;
            // ★ 上限是「这次 RPC 里最后一条新条目」，不是「我自己的最后一条」——
            //   后者可能是还没被这次 RPC 验证过的旧条目。（Figure 2：index of last new entry）
            int lastNew = m.prevLogIndex + m.entries.size();
            if (m.leaderCommit > n.commitIndex) {
                n.commitIndex = Math.min(m.leaderCommit, lastNew);
                apply(n);
            }
            n.persist();
            Msg r=new Msg(n.id,m.from,MType.APPEND_REPLY,n.term);
            r.success=true; r.matchIndex=m.prevLogIndex+m.entries.size();
            net.send(r);
        }

        void handleAppendReply(Node n, Msg m) {
            if (m.term > n.term) { stepDown(n, m.term); return; }
            if (n.state != State.LEADER || m.term != n.term) return;
            if (m.success) {
                n.matchIndex[m.from] = Math.max(n.matchIndex[m.from], m.matchIndex);
                n.nextIndex[m.from] = n.matchIndex[m.from] + 1;
                maybeCommit(n);
                return;
            }
            if (slowBackoff) { if (n.nextIndex[m.from] > 1) n.nextIndex[m.from]--; return; }
            if (m.conflictTerm == -1) { n.nextIndex[m.from] = m.conflictIndex; return; }
            int last = -1;
            for (int i=n.log.size()-1;i>=1;i--) if (n.log.get(i).term()==m.conflictTerm) { last=n.log.get(i).index(); break; }
            n.nextIndex[m.from] = Math.max(1, last >= 0 ? last+1 : m.conflictIndex);
        }

        /** §3.12 的提交规则，两个条件缺一不可。 */
        void maybeCommit(Node n) {
            for (int idx = n.last().index(); idx > n.commitIndex; idx--) {
                int cnt = 0;
                for (int j=0;j<nodes.size();j++) if (n.matchIndex[j] >= idx) cnt++;
                if (cnt < n.majorityOf()) continue;      // ★ 同上
                // ★★★ 这一行就是 Figure 8 的那条限制 ★★★
                if (!naiveCommit && n.entryAt(idx).term() != n.term) continue;
                n.commitIndex = idx; apply(n); return;
            }
        }

        void apply(Node n) {
            while (n.lastApplied < n.commitIndex) {
                n.lastApplied++;
                Entry e = n.entryAt(n.lastApplied);
                if (e.index() != n.lastApplied) { n.lastApplied--; break; }
                n.applied.put(e.index(), e.cmd());
                applyKV(n, e.cmd());
                Entry prev = globalCommitted.get(e.index());
                if (prev != null) {
                    if (!prev.cmd().equals(e.cmd()) || prev.term() != e.term())
                        violate("【状态机安全性被破坏】index=%d 上出现两条不同的已提交日志：先前 (term=%d,%s)，现在 N%d 执行了 (term=%d,%s)",
                                e.index(), prev.term(), prev.cmd(), n.id+1, e.term(), e.cmd());
                } else globalCommitted.put(e.index(), e);
            }
            maybeSnapshotOf(this, n);
        }

        void applyKV(Node n, String cmd) {
            if (!cmd.startsWith("SET ")) return;
            String[] kv = cmd.substring(4).split("=", 2);
            if (kv.length == 2) n.kv.put(kv[0], kv[1]);
        }

        // ─── 主循环 ──────────────────────────────────────────────────────
        void handle(Msg m) {
            Node n = nodes.get(m.to);
            switch (m.type) {
                case REQUEST_VOTE -> {
                    if (m.term > n.term) stepDown(n, m.term);
                    boolean grant = m.term == n.term && n.state != State.LEADER
                            && (n.votedFor == NO_VOTE || n.votedFor == m.from)
                            && (noLogCheck || upToDate(m.lastLogTerm, m.lastLogIndex, n.last())); // ★ 日志至少一样新
                    if (grant) { n.votedFor=m.from; n.timer=0; n.timeout=newTimeout(); n.persist(); }
                    // ★ 投票必须先落盘再回复 —— 否则重启后会忘记，同任期投两次
                    Msg r=new Msg(n.id,m.from,MType.VOTE_REPLY,n.term); r.granted=grant; net.send(r);
                }
                case VOTE_REPLY -> {
                    if (m.term > n.term) { stepDown(n, m.term); return; }
                    if (n.state==State.CANDIDATE && m.granted && m.term==n.term) {
                        n.votes.add(m.from);
                        if (n.votes.size() >= n.majorityOf()) becomeLeader(n);  // ★ 用本节点自己的配置
                    }
                }
                case APPEND -> handleAppend(n, m);
                case APPEND_REPLY -> { n.ackTicks++; handleAppendReply(n, m); }
                case SNAPSHOT -> handleSnapshot(this, n, m);
                case SNAPSHOT_REPLY -> {
                    if (n.state==State.LEADER && m.success) {
                        n.ackTicks++;
                        n.matchIndex[m.from] = Math.max(n.matchIndex[m.from], m.matchIndex);
                        n.nextIndex[m.from] = n.matchIndex[m.from] + 1;
                    }
                }
            }
        }

        boolean submit(String cmd) {
            Node l = leader();
            if (l == null) return false;
            cmdSeq++;
            l.log.add(new Entry(l.last().index()+1, l.term, cmd));
            l.matchIndex[l.id] = l.last().index();
            l.persist();
            return true;
        }

        void step(long dt) {
            for (Node n : nodes) {
                if (net.isDown(n.id)) continue;
                if (n.state == State.LEADER) {
                    n.hb += dt;
                    if (n.hb >= HEARTBEAT) {
                        // 上一轮心跳收到过半响应 ⇒ 续租约（Lease Read 的依据）
                        if (n.ackTicks+1 >= n.majorityOf() && leaseMs > 0) n.leaseTill = net.now + leaseMs;
                        n.ackTicks = 0;
                        n.hb = 0; n.matchIndex[n.id] = n.last().index();
                        for (int j=0;j<nodes.size();j++) if (j!=n.id && n.config.contains(j)) sendAppend(n, j);
                        maybeCommit(n);
                    }
                } else {
                    n.timer += dt;
                    if (n.timer >= n.timeout) { if (n.state==State.CANDIDATE) splits++; startElection(n); }
                }
            }
            for (Msg m : net.advance(dt)) if (!net.isDown(m.to)) handle(m);
            checkSafety();
        }

        void checkSafety() {
            Map<Integer,Integer> byTerm = new HashMap<>();
            for (Node n : nodes) if (!net.isDown(n.id) && n.state==State.LEADER)
                byTerm.merge(n.term, 1, Integer::sum);
            byTerm.forEach((t,k) -> { if (k>1) { twoLeadersEver = true;
                violate("【选举安全性被破坏】term %d 同时有 %d 个 Leader", t, k); } });
            if (activeLeaders() > 1) twoLeadersEver = true;

            for (int i=0;i<nodes.size();i++) for (int j=i+1;j<nodes.size();j++) {
                Node a=nodes.get(i), b=nodes.get(j);
                int lo = Math.max(a.snapIdx, b.snapIdx) + 1;
                int mn = Math.min(a.last().index(), b.last().index());
                for (int k=mn;k>=lo;k--) {
                    if (a.entryAt(k).term() == b.entryAt(k).term()) {
                        for (int x=lo;x<=k;x++)
                            if (a.entryAt(x).term()!=b.entryAt(x).term()
                                    || !a.entryAt(x).cmd().equals(b.entryAt(x).cmd()))
                                violate("【日志匹配性质被破坏】N%d 与 N%d 在 index=%d 上 term 相同，但 index=%d 不同",
                                        i+1, j+1, k, x);
                        break;
                    }
                }
            }
        }

        /** 有几个 Leader【能在自己的配置下凑齐过半】——这才是裂脑的准确定义。
         *  一个被隔离到少数派的僵尸 Leader 虽然也自称 Leader，但它一条日志都提交不了，不算裂脑。 */
        int activeLeaders() {
            int cnt = 0;
            for (Node n : nodes) {
                if (net.isDown(n.id) || n.state != State.LEADER) continue;
                int reach = 0;
                for (int j=0;j<nodes.size();j++)
                    if (n.config.contains(j) && !net.isDown(j) && net.sameGroup(n.id, j)) reach++;
                if (reach >= n.majorityOf()) cnt++;
            }
            return cnt;
        }

        Node leader() {
            for (Node n : nodes) if (!net.isDown(n.id) && n.state==State.LEADER) return n;
            return null;
        }
        int maxCommit() { return nodes.stream().mapToInt(n -> n.commitIndex).max().orElse(0); }
        boolean logsIdentical() {
            Node base = nodes.get(0);
            for (Node n : nodes) {
                int mn = Math.min(base.commitIndex, n.commitIndex);
                for (int i=Math.max(base.snapIdx, n.snapIdx)+1;i<=mn;i++)
                    if (!base.entryAt(i).cmd().equals(n.entryAt(i).cmd())
                            || base.entryAt(i).term()!=n.entryAt(i).term()) return false;
            }
            return true;
        }
        String logStr(int i) {
            StringBuilder sb=new StringBuilder();
            for (Entry e : nodes.get(i).log.subList(1, nodes.get(i).log.size())) sb.append(e.term()).append(' ');
            return sb.isEmpty() ? "(空)" : sb.toString();
        }
    }

    // ══════════════════════════════════════════════════════════════════════
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
        n.timer = 0; n.timeout = c.newTimeout(); n.hb = 0; n.leaseTill = 0;
        for (int j=0;j<c.nodes.size();j++) { n.nextIndex[j]=1; n.matchIndex[j]=0; }
        n.persist();
        c.net.revive(i);
    }

    // ══════════════════════════════════════════════════════════════════════
    // Part 3-C ②：快照
    // ══════════════════════════════════════════════════════════════════════
    static void maybeSnapshotOf(Cluster c, Node n) {
        if (c.snapThreshold <= 0 || n.lastApplied - n.snapIdx < c.snapThreshold) return;
        int cut = n.lastApplied;
        Entry e = n.entryAt(cut);
        n.snapIdx = cut; n.snapTerm = e.term();
        n.snapshot = new HashMap<>(n.kv);
        List<Entry> kept = new ArrayList<>(List.of(new Entry(n.snapIdx, n.snapTerm, "")));
        for (Entry x : n.log) if (x.index() > cut) kept.add(x);
        n.log = kept; n.persist(); c.snapshots++;
    }

    static void sendSnapshotOf(Cluster c, Node n, int to) {
        Msg m = new Msg(n.id, to, MType.SNAPSHOT, n.term);
        m.lastIncludedIndex = n.snapIdx; m.lastIncludedTerm = n.snapTerm;
        m.snapshot = new HashMap<>(n.snapshot);
        c.net.send(m); c.installSnaps++;
    }

    static void handleSnapshot(Cluster c, Node n, Msg m) {
        if (m.term < n.term) return;
        n.term=m.term; n.state=State.FOLLOWER; n.votedFor=m.from;
        n.timer=0; n.timeout=c.newTimeout();
        if (m.lastIncludedIndex <= n.snapIdx) return;
        n.snapIdx=m.lastIncludedIndex; n.snapTerm=m.lastIncludedTerm;
        n.snapshot=new HashMap<>(m.snapshot);
        n.kv.clear(); n.kv.putAll(m.snapshot);
        n.log = new ArrayList<>(List.of(new Entry(n.snapIdx, n.snapTerm, "")));
        n.commitIndex=n.snapIdx; n.lastApplied=n.snapIdx;
        n.persist();
        Msg r=new Msg(n.id,m.from,MType.SNAPSHOT_REPLY,n.term);
        r.success=true; r.matchIndex=n.snapIdx; c.net.send(r);
    }

    // ══════════════════════════════════════════════════════════════════════
    // Part 3-C ④：三种读
    // ══════════════════════════════════════════════════════════════════════
    enum ReadMode {
        LOCAL("本地读"), READ_INDEX("ReadIndex"), LEASE("Lease Read");
        final String cn; ReadMode(String cn){this.cn=cn;}
    }

    /** 返回 {值, 是否成功}。失败表示节点正确地拒绝了服务 —— 这是好事。 */
    static String[] read(Cluster c, int i, String key, ReadMode mode) {
        Node n = c.nodes.get(i);
        if (c.net.isDown(n.id) || n.state != State.LEADER) return new String[]{null,"0"};
        return switch (mode) {
            case LOCAL -> new String[]{n.kv.get(key), "1"};                 // 无法证明自己此刻还是主
            case READ_INDEX -> n.ackTicks+1 < n.majorityOf()
                    ? new String[]{null,"0"} : new String[]{n.kv.get(key),"1"};
            case LEASE -> c.net.now > n.leaseTill
                    ? new String[]{null,"0"} : new String[]{n.kv.get(key),"1"};
        };
    }

    // ══════════════════════════════════════════════════════════════════════
    static Cluster mkw(long seed, long window, double loss) {
        Cluster c = new Cluster(N, new Network(N, DELAY, JITTER, loss, seed), 400, window, seed);
        c.leaseMs = 250;
        return c;
    }
    static Cluster mk(long seed) { return mkw(seed, 300, 0); }
    static void run(Cluster c, long ms) { for (long t=0;t<ms;t+=TICK) c.step(TICK); }

    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 每组 %d 个种子 ｜ 种子基准 %d%n",
                N, DELAY, JITTER, TRIALS, SEED);
        lab1(); lab2(); lab3(); lab4(); epilogue();
    }

    // ─── Lab 3C-1 ─────────────────────────────────────────────────────────
    static void lab1() {
        head("3C-1", "持久化与崩溃重启", "分别丢掉三样状态中的一样，看哪条安全性倒下");
        int[] w = {28,14,14,14,18}; String al = "LRRRR";
        tableHead(new String[]{"磁盘上丢了什么","违规的种子","选举安全性","状态机安全性","结论"}, w, al);

        java.util.function.BiConsumer<String,boolean[]> tryIt = (label, flags) -> {
            int bad=0, elec=0, sm=0;
            for (long seed=1; seed<=TRIALS; seed++) {
                // 固定超时 + 丢包 ⇒ 经常出现两个 candidate 抢同一个任期
                Cluster c = mkw(SEED+seed, 0, 0.12);
                c.loseTerm=flags[0]; c.loseVote=flags[1]; c.loseLog=flags[2];
                run(c, 1200);
                for (int i=0;i<80;i++) {
                    c.submit("SET k"+(i%5)+"="+i); run(c, 30);
                    if (i%8==7) {
                        Node l=c.leader(); if (l!=null) c.net.kill(l.id);
                        run(c, 420);
                        int k=c.rnd.nextInt(N);
                        if (!c.net.isDown(k)) { c.net.kill(k); run(c,30); restart(c,k); }
                        run(c, 250);
                        for (int j=0;j<N;j++) if (c.net.isDown(j)) restart(c,j);
                        run(c, 350);
                    }
                }
                run(c, 3000);
                if (!c.violations.isEmpty()) {
                    bad++;
                    for (String v : c.violations) if (v.contains("选举安全性")) { elec++; break; }
                    for (String v : c.violations) if (v.contains("状态机")||v.contains("Leader 完整性")) { sm++; break; }
                }
            }
            tableRow(new String[]{label, bad+"/"+TRIALS, ""+elec, ""+sm, bad>0?"✗ 被破坏":"✓ 安全"}, w, al);
        };

        tryIt.accept("（什么都不丢，正确）", new boolean[]{false,false,false});
        if (LOSE.isEmpty() || LOSE.equals("votedfor")) tryIt.accept("★ 丢失 votedFor", new boolean[]{false,true,false});
        if (LOSE.isEmpty() || LOSE.equals("term"))     tryIt.accept("丢失 currentTerm", new boolean[]{true,false,false});
        if (LOSE.isEmpty() || LOSE.equals("log"))      tryIt.accept("丢失 log[]", new boolean[]{false,false,true});

        System.out.println();
        row("构造场景 · votedFor 已落盘", constructed(false)
                ? "✗ 出现两个 Leader" : "✓ N5 记得投过票，拒绝第二次索票 ⇒ 只有一个 Leader");
        row("构造场景 · votedFor 丢失", constructed(true)
                ? "✗ 同一任期出现两个 Leader —— 票集合 {N1,N3,N5} 与 {N2,N4,N5} 的交集 N5 投了两次" : "✓ 安全");
        System.out.println("""

    ▸ 上面两行是【构造】出来的场景：直接把"两个 candidate 抢同一任期"摆出来，
      不依赖随机调度去撞。这也说明随机故障注入为什么不够 ——
      有些 bug 需要非常特定的交错才会现形，这正是 TLA+ 和 Jepsen 存在的理由（Part 6）。
    ▸ 表格里 votedFor 那一行的随机试验可能是 0，不代表它安全，只代表没撞上。
    ▸ log 丢失破坏的是 Leader 完整性："持有已提交日志的节点构成过半"变成了假的。
    ▸ 判断什么必须落盘的通用直觉：【可推导的结论不用存，对外许下的承诺必须存】。
      commitIndex 是结论（Leader 会告诉你），votedFor 是承诺（只有自己知道）。""".indent(4));
    }

    /** 构造：N1、N2 在任期 5 同时索票；N5 投给 N1 后崩溃重启。 */
    static boolean constructed(boolean loseVote) {
        Cluster c = new Cluster(5, new Network(5,20,0,0,1), 400, 0, 1);
        c.loseVote = loseVote;
        for (Node n : c.nodes) { n.term = 5; n.persist(); }
        Node n1=c.nodes.get(0), n2=c.nodes.get(1), n5=c.nodes.get(4);
        for (Node cd : List.of(n1,n2)) {
            cd.state=State.CANDIDATE; cd.votedFor=cd.id;
            cd.votes=new HashSet<>(List.of(cd.id)); cd.persist();
        }
        c.nodes.get(2).votedFor=0; c.nodes.get(2).persist(); n1.votes.add(2);   // N3 投 N1
        c.nodes.get(3).votedFor=1; c.nodes.get(3).persist(); n2.votes.add(3);   // N4 投 N2
        n5.votedFor=0; n5.persist(); n1.votes.add(4);                            // N5 投 N1
        c.becomeLeader(n1);

        c.net.kill(4); restart(c, 4);                                            // ★ 崩溃重启

        Msg rv = new Msg(1, 4, MType.REQUEST_VOTE, 5);
        rv.lastLogIndex = n2.last().index(); rv.lastLogTerm = n2.last().term();
        c.handle(rv);
        if (n5.votedFor == 1) {
            n2.votes.add(4);
            if (n2.votes.size() >= n2.majorityOf()) c.becomeLeader(n2);
        }
        c.checkSafety();
        return c.twoLeadersEver;
    }

    // ─── Lab 3C-2 ─────────────────────────────────────────────────────────
    static void lab2() {
        head("3C-2", "日志压缩与 InstallSnapshot", "把一个节点隔离到落后超过快照点，再放回来");
        Cluster c = mk(SEED); c.snapThreshold = 30;
        run(c, 1500);
        c.net.kill(4);
        for (int i=0;i<220;i++) { c.submit("SET k"+(i%8)+"="+i); run(c, 20); }
        run(c, 1500);
        Node l = c.leader();
        int before = c.installSnaps;
        int logLen = l==null?0:l.last().index()-l.snapIdx, snapAt = l==null?0:l.snapIdx;
        c.net.revive(4);
        long caught=-1, t0=c.net.now;
        for (long t=0;t<20000;t+=TICK) {
            c.step(TICK);
            if (l!=null && c.nodes.get(4).lastApplied >= l.commitIndex && l.commitIndex>0) { caught=c.net.now-t0; break; }
        }
        int[] w = {30,16,4,32}; String al = "LRLL";
        tableHead(new String[]{"指标","数值","","说明"}, w, al);
        tableRow(new String[]{"提交的命令数", ""+c.cmdSeq, "", "客户端发起"}, w, al);
        tableRow(new String[]{"Leader 做过的快照", ""+c.snapshots, "", "每 30 条压缩一次"}, w, al);
        tableRow(new String[]{"快照点 lastIncludedIndex", ""+snapAt, "", "这之前的日志已丢弃"}, w, al);
        tableRow(new String[]{"Leader 保留的日志条数", ""+logLen, "", "而不是全部 "+c.cmdSeq+" 条"}, w, al);
        tableRow(new String[]{"发出的 InstallSnapshot", ""+(c.installSnaps-before), "", "N5 落后太多，走快照追赶"}, w, al);
        tableRow(new String[]{"N5 追平耗时", caught>=0?caught+" ms":"未追平", "", ""}, w, al);
        boolean ok = c.logsIdentical();
        tableRow(new String[]{"日志仍然一致", ok?"✓ 是":"✗ 否", "", ok?"快照没有破坏日志匹配性质":"发现分歧"}, w, al);
        row("安全性违反", c.violations.size()+" 条");
        System.out.println("""

    ▸ Leader 只保留了最近几十条日志，其余压缩进了快照 —— 磁盘和重放时间都是常数级。
    ▸ N5 落后到快照点之前，Leader 发现 nextIndex ≤ lastIncludedIndex，
      于是改发 InstallSnapshot 而不是逐条补日志。
    ▸ 快照里必须带 (lastIncludedIndex, lastIncludedTerm)：日志被截断后，
      AppendEntries 的一致性检查还要拿它们当"被截断部分的代表"来比较。""".indent(4));
    }

    // ─── Lab 3C-3 ─────────────────────────────────────────────────────────
    static void setConfig(Cluster c, int i, int... members) {
        Set<Integer> cfg = new HashSet<>();
        for (int m : members) cfg.add(m);
        c.nodes.get(i).config = cfg;
    }

    static void lab3() {
        head("3C-3", "成员变更：直接跳 vs 单节点", "让不同节点在不同时刻切换配置，看会不会选出两个 Leader");
        int[] w = {34,18,20}; String al = "LRR";
        tableHead(new String[]{"变更方式","出现两个 Leader","结论"}, w, al);

        java.util.function.LongPredicate direct = seed -> {
            Cluster c = mk(seed);
            for (int i=0;i<N;i++) setConfig(c, i, 0,1,2);
            c.net.kill(3); c.net.kill(4);
            run(c, 2500);
            c.net.revive(3); c.net.revive(4);
            for (int i : new int[]{2,3,4}) setConfig(c, i, 0,1,2,3,4);   // 只有一部分切了
            c.net.partition(new int[]{0,0,1,1,1});
            run(c, 9000);
            return !c.violations.isEmpty() || c.twoLeadersEver;
        };
        java.util.function.LongPredicate single = seed -> {
            Cluster c = mk(seed);
            for (int i=0;i<N;i++) setConfig(c, i, 0,1,2);
            c.net.kill(3); c.net.kill(4);                                 // N5 不属于任何配置
            run(c, 2500);
            c.net.revive(3);
            for (int i : new int[]{2,3}) setConfig(c, i, 0,1,2,3);
            c.net.partition(new int[]{0,0,1,1,1});                        // 用同样的方式去"造"裂脑
            run(c, 9000);
            return !c.violations.isEmpty() || c.twoLeadersEver;
        };
        java.util.function.ToIntFunction<java.util.function.LongPredicate> cnt = f -> {
            int k=0; for (long s=1;s<=TRIALS;s++) if (f.test(SEED+s*13)) k++; return k;
        };
        int d = cnt.applyAsInt(direct), sg = cnt.applyAsInt(single);
        tableRow(new String[]{"① 直接跳 {A,B,C} → {A..E}", d+"/"+TRIALS, d>0?"✗ 会裂脑":"✓"}, w, al);
        tableRow(new String[]{"② 单节点 {A,B,C} → {A,B,C,D}", sg+"/"+TRIALS, sg>0?"✗ 会裂脑":"✓ 一次都造不出来"}, w, al);
        System.out.println();
        row("① 的两个过半集合", "旧 {A,B}（2/3 够）与 新 {C,D,E}（3/5 够）—— 不相交");
        row("① 的算式", "maj_old + maj_new = 2+3 = 5，|并集| = 5，5 > 5 ✗ 不成立");
        row("② 的算式", "maj_old + maj_new = 2+3 = 5，|并集| = 4，5 > 4 ✓ 成立 ⇒ 必相交");
        System.out.println("""

    ▸ ① 里 {A,B} 用旧配置算已经过半（2/3），{C,D,E} 用新配置算也已经过半（3/5），
      两组没有任何共同节点 ⇒ 可以在同一任期各选一个 Leader ⇒ 脑裂。
    ▸ ② 里要凑出两个不相交的过半集合需要 5 个不同节点，而并集只有 4 个 —— 鸽笼原理。
      所以「一次只动一个节点」不是谨慎习惯，是可以证明的充分条件。
    ▸ 一般式：加一个节点时 maj_old + maj_new = n+2 > n+1 = |并集|，恒成立。""".indent(4));
    }

    // ─── Lab 3C-4 ─────────────────────────────────────────────────────────
    static void lab4() {
        head("3C-4", "三种读的对照", "制造僵尸 Leader，看哪种读会静默返回陈旧数据");
        Map<ReadMode,int[]> res = new LinkedHashMap<>();   // {stale, errs, good}
        for (ReadMode m : ReadMode.values()) res.put(m, new int[3]);

        for (long seed=1; seed<=TRIALS; seed++) {
            Cluster c = mk(SEED + seed*7);
            run(c, 1500);
            Node l = c.leader();
            if (l == null) continue;
            c.submit("SET x=1"); run(c, 600);

            int other = (l.id+1)%N;
            int[] grp = new int[N]; grp[l.id]=1; grp[other]=1;
            c.net.partition(grp);
            run(c, 6000);

            Node newL = null;
            for (Node n : c.nodes) if (n.state==State.LEADER && grp[n.id]==0) newL = n;
            if (newL == null) continue;
            newL.log.add(new Entry(newL.last().index()+1, newL.term, "SET x=99"));
            newL.matchIndex[newL.id] = newL.last().index();
            run(c, 2500);
            if (!"99".equals(newL.kv.get("x"))) continue;

            Node zombie = c.nodes.get(l.id);
            if (zombie.state != State.LEADER) continue;
            for (ReadMode m : ReadMode.values()) {
                String[] r = read(c, zombie.id, "x", m);
                int[] a = res.get(m);
                if (r[1].equals("0")) a[1]++;
                else if (!"99".equals(r[0])) a[0]++;
                else a[2]++;
            }
        }
        int[] w = {20,20,20,20}; String al = "LRRR";
        tableHead(new String[]{"读的实现","★ 静默返回陈旧值","正确报错","返回正确值"}, w, al);
        for (ReadMode m : ReadMode.values()) {
            int[] a = res.get(m);
            tableRow(new String[]{m.cn, ""+a[0], ""+a[1], ""+a[2]}, w, al);
        }
        System.out.println("""

    ▸ 本地读会静默返回旧值 —— 不报错、不留日志。客户端刚从新 Leader 拿到"写入成功"，
      转头从僵尸 Leader 读到旧值，线性一致性当场破裂，而监控上什么都看不出来。
    ▸ ReadIndex 读前先确认过半节点仍认自己是主，僵尸凑不齐 ⇒ 报错。
      【给你错误，而不是给你错的数据】。
    ▸ Lease Read 的租约续不上就过期，同样拒绝。它更快，但安全性建立在【时钟漂移有界】上。
      Part 1 那个坑，在这里变成了一致性问题。
    ▸ 所以"Raft 集群是线性一致的"准确说法是：写入是线性一致的；
      读是否线性一致，取决于你选了哪一种。etcd 默认 ReadIndex，TiKV 默认 Lease Read。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  三个思考题（答案在课件 §3.18 / §3.20 / §3.21）：

    1. 三样状态里，丢掉哪一样造成的违规最多？为什么是它？
    2. Lab 3C-3 的 ① 里两个 Leader 分别靠哪些节点的票当选？验证它们确实不相交。
    3. Lab 3C-4 里 Lease Read 一次陈旧读都没有 —— 这是否意味着它和 ReadIndex 一样安全？

  Part 3 到此结束。下一站：Part 4 · 分布式事务。""");
        System.out.println("═".repeat(78) + "\n");
    }


    // ─── 输出工具 ─────────────────────────────────────────────────────────
    static int dispw(String s) {
        int w=0;
        for (int i=0;i<s.length();i++) {
            char c=s.charAt(i);
            w += (c>=0x1100 && (c<=0x115F || (c>=0x2E80&&c<=0xA4CF&&c!=0x303F)
                    || (c>=0xAC00&&c<=0xD7A3) || (c>=0xF900&&c<=0xFAFF)
                    || (c>=0xFE30&&c<=0xFE6F) || (c>=0xFF00&&c<=0xFF60)
                    || (c>=0xFFE0&&c<=0xFFE6))) ? 2 : 1;
        }
        return w;
    }
    static String padR(String s,int w){int d=w-dispw(s);return d>0?s+" ".repeat(d):s;}
    static String padL(String s,int w){int d=w-dispw(s);return d>0?" ".repeat(d)+s:s;}
    static String layout(String[] c,int[] w,String al){
        StringBuilder sb=new StringBuilder("    ");
        for(int i=0;i<c.length;i++) sb.append(i<al.length()&&al.charAt(i)=='L'?padR(c[i],w[i]):padL(c[i],w[i]));
        return sb.toString().stripTrailing();
    }
    static void head(String label,String title,String sub){
        System.out.println("\n"+"═".repeat(78));
        System.out.printf("  Lab %s · %s%n  %s%n", label, title, sub);
        System.out.println("═".repeat(78));
    }
    static void row(String k,Object v){ System.out.println("    "+padR(k,30)+" "+v); }
    static void tableHead(String[] cols,int[] w,String al){
        System.out.println(layout(cols,w,al));
        int t=0; for(int x:w) t+=x;
        System.out.println("    "+"─".repeat(t));
    }
    static void tableRow(String[] c,int[] w,String al){ System.out.println(layout(c,w,al)); }

    static void parseArgs(String[] a) {
        for (int i=0;i<a.length;i++) {
            String k=a[i].replaceFirst("^-+","");
            if (k.equals("v")) { VERBOSE=true; continue; }
            if (i+1>=a.length) break;
            String v=a[++i];
            switch (k) {
                case "n" -> N=Integer.parseInt(v);
                case "delay" -> DELAY=Long.parseLong(v);
                case "jitter" -> JITTER=Long.parseLong(v);
                case "loss" -> LOSS=Double.parseDouble(v);
                case "seed" -> SEED=Long.parseLong(v);
                case "lose" -> LOSE=v;
                case "trials" -> TRIALS=Integer.parseInt(v);
                default -> System.out.println("未知参数: "+k);
            }
        }
    }
}
