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
    static boolean NAIVE = false, SLOW = false, NOLOGCHECK = false, VERBOSE = false;
    static final long TICK = 5, HEARTBEAT = 100;
    static final int NO_VOTE = -1;

    // ══════════════════════════════════════════════════════════════════════
    // 确定性网络
    // ══════════════════════════════════════════════════════════════════════
    enum MType { REQUEST_VOTE, VOTE_REPLY, APPEND, APPEND_REPLY }

    record Entry(int index, int term, String cmd) {}

    static class Msg {
        int from, to; MType type; int term;
        int lastLogIndex, lastLogTerm; boolean granted;                 // RequestVote
        int prevLogIndex, prevLogTerm; List<Entry> entries = List.of(); // AppendEntries
        int leaderCommit;
        boolean success; int matchIndex, conflictTerm, conflictIndex;   // 回复
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

        List<Entry> log = new ArrayList<>(List.of(new Entry(0,0,"")));  // 0 号是哨兵
        int commitIndex = 0, lastApplied = 0;
        final Map<Integer,String> applied = new HashMap<>();
        int[] nextIndex, matchIndex;

        Node(int id, int n) { this.id=id; nextIndex=new int[n]; matchIndex=new int[n]; }
        Entry last() { return log.get(log.size()-1); }
        Entry at(int i) { return (i>=0 && i<log.size()) ? log.get(i) : null; }
    }

    static class Cluster {
        final List<Node> nodes = new ArrayList<>();
        final Network net; final Random rnd; final long base, window;
        boolean naiveCommit = NAIVE, slowBackoff = SLOW;
        boolean noLogCheck = NOLOGCHECK;   // ★ 关掉「日志至少一样新」⇒ 退化成 Lab 3-A 的模型

        /** 上帝视角：任何节点提交一条日志就登记在这里，用于安全性断言。 */
        final Map<Integer,Entry> globalCommitted = new HashMap<>();
        int elections=0, splits=0, rejects=0, appends=0, cmdSeq=0;
        final List<String> violations = new ArrayList<>();

        Cluster(int n, Network net, long base, long window, long seed) {
            this.net=net; this.base=base; this.window=window; rnd=new Random(seed+999);
            for (int i=0;i<n;i++) { Node nd=new Node(i,n); nd.timeout=newTimeout(); nodes.add(nd); }
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
        }
        void startElection(Node n) {
            n.term++; n.state=State.CANDIDATE; n.votedFor=n.id;
            n.votes=new HashSet<>(List.of(n.id));
            n.timer=0; n.timeout=newTimeout(); elections++;
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
            Entry prev = n.at(ni-1);
            Msg m = new Msg(n.id, to, MType.APPEND, n.term);
            m.prevLogIndex = prev.index(); m.prevLogTerm = prev.term();
            m.entries = ni <= n.last().index() ? new ArrayList<>(n.log.subList(ni, n.log.size())) : List.of();
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
                    while (i > 1 && n.log.get(i-1).term() == prev.term()) i--;
                    r.conflictIndex = i;
                }
                net.send(r); return;
            }
            for (int k=0;k<m.entries.size();k++) {
                int idx = m.prevLogIndex + 1 + k;
                Entry e = m.entries.get(k), cur = n.at(idx);
                if (cur != null) {
                    if (cur.term() == e.term()) continue;
                    while (n.log.size() > idx) n.log.remove(n.log.size()-1);  // ★ 冲突 ⇒ 截断
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
            for (int i=n.log.size()-1;i>=1;i--) if (n.log.get(i).term()==m.conflictTerm) { last=i; break; }
            n.nextIndex[m.from] = Math.max(1, last >= 0 ? last+1 : m.conflictIndex);
        }

        /** §3.12 的提交规则，两个条件缺一不可。 */
        void maybeCommit(Node n) {
            for (int idx = n.last().index(); idx > n.commitIndex; idx--) {
                int cnt = 0;
                for (int j=0;j<nodes.size();j++) if (n.matchIndex[j] >= idx) cnt++;
                if (cnt < majority()) continue;
                // ★★★ 这一行就是 Figure 8 的那条限制 ★★★
                if (!naiveCommit && n.log.get(idx).term() != n.term) continue;
                n.commitIndex = idx; apply(n); return;
            }
        }

        void apply(Node n) {
            while (n.lastApplied < n.commitIndex) {
                n.lastApplied++;
                Entry e = n.log.get(n.lastApplied);
                n.applied.put(e.index(), e.cmd());
                Entry prev = globalCommitted.get(e.index());
                if (prev != null) {
                    if (!prev.cmd().equals(e.cmd()) || prev.term() != e.term())
                        violate("【状态机安全性被破坏】index=%d 上出现两条不同的已提交日志：先前 (term=%d,%s)，现在 N%d 执行了 (term=%d,%s)",
                                e.index(), prev.term(), prev.cmd(), n.id+1, e.term(), e.cmd());
                } else globalCommitted.put(e.index(), e);
            }
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
                    if (grant) { n.votedFor=m.from; n.timer=0; n.timeout=newTimeout(); }
                    Msg r=new Msg(n.id,m.from,MType.VOTE_REPLY,n.term); r.granted=grant; net.send(r);
                }
                case VOTE_REPLY -> {
                    if (m.term > n.term) { stepDown(n, m.term); return; }
                    if (n.state==State.CANDIDATE && m.granted && m.term==n.term) {
                        n.votes.add(m.from);
                        if (n.votes.size() >= majority()) becomeLeader(n);
                    }
                }
                case APPEND -> handleAppend(n, m);
                case APPEND_REPLY -> handleAppendReply(n, m);
            }
        }

        boolean submit(String cmd) {
            Node l = leader();
            if (l == null) return false;
            cmdSeq++;
            l.log.add(new Entry(l.last().index()+1, l.term, cmd));
            l.matchIndex[l.id] = l.last().index();
            return true;
        }

        void step(long dt) {
            for (Node n : nodes) {
                if (net.isDown(n.id)) continue;
                if (n.state == State.LEADER) {
                    n.hb += dt;
                    if (n.hb >= HEARTBEAT) {
                        n.hb = 0; n.matchIndex[n.id] = n.last().index();
                        for (int j=0;j<nodes.size();j++) if (j!=n.id) sendAppend(n, j);
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
            byTerm.forEach((t,k) -> { if (k>1) violate("【选举安全性被破坏】term %d 同时有 %d 个 Leader", t, k); });

            for (int i=0;i<nodes.size();i++) for (int j=i+1;j<nodes.size();j++) {
                Node a=nodes.get(i), b=nodes.get(j);
                int mn = Math.min(a.last().index(), b.last().index());
                for (int k=mn;k>=1;k--) {
                    if (a.log.get(k).term() == b.log.get(k).term()) {
                        for (int x=1;x<=k;x++)
                            if (a.log.get(x).term()!=b.log.get(x).term()
                                    || !a.log.get(x).cmd().equals(b.log.get(x).cmd()))
                                violate("【日志匹配性质被破坏】N%d 与 N%d 在 index=%d 上 term 相同，但 index=%d 不同",
                                        i+1, j+1, k, x);
                        break;
                    }
                }
            }
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
                for (int i=1;i<=mn;i++)
                    if (!base.log.get(i).cmd().equals(n.log.get(i).cmd())
                            || base.log.get(i).term()!=n.log.get(i).term()) return false;
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
    static Cluster build(long seed) {
        Cluster c = new Cluster(N, new Network(N, DELAY, JITTER, LOSS, seed), 400, 300, seed);
        c.naiveCommit = NAIVE; c.slowBackoff = SLOW; c.noLogCheck = NOLOGCHECK;
        return c;
    }
    static void run(Cluster c, long ms) { for (long t=0;t<ms;t+=TICK) c.step(TICK); }

    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 丢包 %.0f%% ｜ 种子 %d%s%s%n",
                N, DELAY, JITTER, LOSS*100, SEED,
                NAIVE ? " ｜ ★ 已关闭提交限制" : "", SLOW ? " ｜ 朴素回退" : "");
        lab1(); lab2(); lab3(); lab4(); lab5(); epilogue();
    }

    // ─── Lab 3B-1 ─────────────────────────────────────────────────────────
    static void lab1() {
        head("3B-1", "日志复制", "提交 200 条命令，验证所有节点的日志逐条相同");
        Cluster c = build(SEED);
        run(c, 1500);
        int submitted = 0;
        for (int r=0;r<200;r++) { if (c.submit("SET k"+(r%10)+"="+r)) submitted++; run(c, 30); }
        run(c, 2000);
        boolean ok = c.logsIdentical();
        int[] w = {26,14,4,44}; String al = "LRLL";
        tableHead(new String[]{"指标","数值","","说明"}, w, al);
        tableRow(new String[]{"提交的命令数", ""+submitted, "", "客户端发起"}, w, al);
        tableRow(new String[]{"最高 commitIndex", ""+c.maxCommit(), "", "已达成共识的日志条数"}, w, al);
        tableRow(new String[]{"AppendEntries 成功", ""+c.appends, "", "次"}, w, al);
        tableRow(new String[]{"AppendEntries 被拒", ""+c.rejects, "", "一致性检查未通过"}, w, al);
        tableRow(new String[]{"所有节点日志一致", ok?"✓ 是":"✗ 否", "", ok?"逐条比对通过":"发现分歧"}, w, al);
        System.out.println();
        for (int i=0;i<Math.min(N,3);i++) row("N"+(i+1)+" 日志的 term 序列", truncate(c.logStr(i), 60));
        System.out.println("""

    ▸ 所有节点的日志逐条相同，且状态机在每个 index 上执行了相同的命令。
    ▸ 「被拒」是 0，因为这是个全新集群、没有任何故障，所有节点的日志从一开始就一致。
      只有当日志真的分叉过（换过 Leader、发生过分区），一致性检查才会拒绝 ——
      Lab 3B-4 会故意制造那种局面。""".indent(4));
    }
    static String truncate(String s, int n) { return s.length()<=n ? s : s.substring(0,n)+"…"; }

    // ─── Lab 3B-2 ─────────────────────────────────────────────────────────
    static void lab2() {
        head("3B-2", "三条安全性断言", "边跑边杀节点、边制造分区，每个 tick 都检查");
        int[] w = {30,14,12,12,16}; String al = "LRRRR";
        tableHead(new String[]{"场景","提交的命令","已提交","违反","安全性"}, w, al);

        java.util.function.BiConsumer<String,java.util.function.Consumer<Cluster>> scenario = (label, f) -> {
            Cluster c = build(SEED); run(c, 1200); f.accept(c);
            int nv = c.violations.size();
            tableRow(new String[]{label, ""+c.cmdSeq, ""+c.maxCommit(), ""+nv, nv>0?"✗ 被破坏":"✓ 成立"}, w, al);
            for (String v : c.violations) row("  ↳", v);
        };

        scenario.accept("① 无故障，持续写入", c -> {
            for (int i=0;i<120;i++) { c.submit("cmd"+i); run(c,25); }
            run(c,1500);
        });
        scenario.accept("② 边写边杀 Leader", c -> {
            for (int i=0;i<120;i++) {
                c.submit("cmd"+i); run(c,25);
                if (i%30==29) { Node l=c.leader(); if (l!=null) { c.net.kill(l.id); run(c,900); c.net.revive(l.id); } }
            }
            run(c,2500);
        });
        scenario.accept("③ 反复 3|2 分区", c -> {
            int[] grp = {0,0,0,1,1};
            for (int i=0;i<120;i++) {
                c.submit("cmd"+i); run(c,25);
                if (i%40==20) c.net.partition(grp);
                if (i%40==39) c.net.heal();
            }
            c.net.heal(); run(c,3000);
        });
        scenario.accept("④ 30% 丢包 + 随机杀节点", c -> {
            c.net.loss = 0.3;
            for (int i=0;i<150;i++) {
                c.submit("cmd"+i); run(c,25);
                if (i%25==24) { int k=c.rnd.nextInt(N); if (c.net.isDown(k)) c.net.revive(k); else c.net.kill(k); }
            }
            for (int i=0;i<N;i++) c.net.revive(i);
            c.net.loss = 0; run(c,4000);
        });

        System.out.println("""

    ▸ 断言的是三条：选举安全性、日志匹配性质、状态机安全性。
    ▸ 注意场景③④里「提交的命令」多于「已提交」—— 分区和丢包期间客户端发出的
      请求进了 Leader 的日志却没能提交，最后被截断丢弃。
      客户端看到的是超时，然后数据消失。这就是 §3.10 说的第三态。""".indent(4));
    }

    // ─── Lab 3B-3 ─────────────────────────────────────────────────────────
    record F8Res(boolean committed, boolean overwritten, List<String> viol) {}

    static F8Res figure8(boolean naive) {
        Network net = new Network(5, 20, 5, 0, 1);
        Cluster c = new Cluster(5, net, 400, 300, 1);
        c.naiveCommit = naive;
        // (a)(b) 之后的局面：S1/S2 有 term2 的 index2，S5 有 term3 的 index2
        List<List<Entry>> logs = List.of(
            List.of(new Entry(1,1,"A"), new Entry(2,2,"X")),
            List.of(new Entry(1,1,"A"), new Entry(2,2,"X")),
            List.of(new Entry(1,1,"A")),
            List.of(new Entry(1,1,"A")),
            List.of(new Entry(1,1,"A"), new Entry(2,3,"Y")));
        for (int i=0;i<5;i++) {
            Node n=c.nodes.get(i);
            n.log = new ArrayList<>(List.of(new Entry(0,0,"")));
            n.log.addAll(logs.get(i));
            n.term = 4;
        }
        c.nodes.get(4).term = 3;
        // (c)：S1 在任期 4 当选，已把 term2 那条复制给 S3，并在 index3 写下当前任期的日志
        Node s1 = c.nodes.get(0);
        s1.state = State.LEADER; s1.term = 4;
        c.nodes.get(2).log.add(new Entry(2,2,"X"));
        s1.log.add(new Entry(3,4,"Z"));
        for (int j=0;j<5;j++) { s1.nextIndex[j]=s1.last().index()+1; s1.matchIndex[j]=0; }
        s1.matchIndex[0]=3; s1.matchIndex[1]=2; s1.matchIndex[2]=2;

        c.maybeCommit(s1);
        boolean committed = s1.commitIndex >= 2;
        String saved = c.globalCommitted.containsKey(2) ? c.globalCommitted.get(2).cmd() : "";

        // (d)：S1 崩溃，S5 靠 term3 的日志当选任期 5 并覆盖 S2/S3/S4
        c.net.kill(0);
        Node s5 = c.nodes.get(4);
        s5.term = 5; s5.state = State.LEADER;
        for (int j=0;j<5;j++) { s5.nextIndex[j]=1; s5.matchIndex[j]=0; }
        s5.matchIndex[4]=s5.last().index();
        for (int i=1;i<=3;i++) {
            Node n=c.nodes.get(i);
            n.term=5; n.state=State.FOLLOWER;
            n.log = new ArrayList<>(List.of(new Entry(0,0,"")));
            n.log.addAll(s5.log.subList(1, s5.log.size()));
            s5.matchIndex[i]=n.last().index();
        }
        c.becomeLeader(s5);
        c.checkSafety();
        c.maybeCommit(s5);
        for (int i=1;i<=4;i++) c.apply(c.nodes.get(i));
        boolean overwritten = !saved.isEmpty() && !c.nodes.get(1).log.get(2).cmd().equals(saved);
        return new F8Res(committed, overwritten, c.violations);
    }

    static void lab3() {
        head("3B-3", "Figure 8 复现", "手工构造论文里那五个节点的状态，然后按规则推进");
        F8Res a = figure8(false), b = figure8(true);
        int[] w = {40,18,18}; String al = "LRR";
        tableHead(new String[]{"阶段 (c)→(d) 发生了什么","① 完整 Raft","② 关掉限制"}, w, al);
        tableRow(new String[]{"index2 (term2) 复制到过半了吗","是 (3/5)","是 (3/5)"}, w, al);
        tableRow(new String[]{"S1 把它标记为已提交了吗", a.committed()?"是":"否", b.committed()?"是":"否"}, w, al);
        tableRow(new String[]{"S5 当选后覆盖掉它了吗","是","是"}, w, al);
        tableRow(new String[]{"被覆盖的是「已提交」的日志吗", a.overwritten()?"是":"否", b.overwritten()?"是":"否"}, w, al);
        tableRow(new String[]{"安全性断言", a.viol().isEmpty()?"✓ 成立":"✗ 被破坏",
                b.viol().isEmpty()?"✓ 成立":"✗ 被破坏"}, w, al);
        System.out.println();
        if (!b.viol().isEmpty()) { row("② 报出的违反",""); for (String v : b.viol()) System.out.println("      "+v); }
        System.out.println("""

    ▸ 两边的物理事实完全一样：那条 term2 的日志都复制到了 3/5 个节点，
      也都被 S5 覆盖掉了。唯一的区别是——① 从没把它叫做「已提交」。
    ▸ Raft 没有去阻止覆盖（那需要改选举规则，代价大得多），
      而是确保「被覆盖的东西从来没被承诺过」。这是一个非常克制的修补。
    ▸ 对客户端的含义：① 里客户端收到的是超时（第三态），数据可能在也可能不在；
      ② 里客户端收到的是"成功"，然后数据消失了。后者才是真正的事故。""".indent(4));
    }

    // ─── Lab 3B-4 ─────────────────────────────────────────────────────────
    static void lab4() {
        head("3B-4", "冲突回退：朴素 vs 快速", "让一个 follower 落后很多，数它追平要几轮 RPC");
        int[] w = {26,16,16,18}; String al = "LRRR";
        tableHead(new String[]{"回退策略","被拒次数","追平耗时(ms)","最终是否追平"}, w, al);
        for (boolean slow : new boolean[]{true,false}) {
            Cluster c = new Cluster(N, new Network(N, DELAY, JITTER, 0, SEED), 400, 300, SEED);
            c.slowBackoff = slow;
            run(c, 1200);
            c.net.kill(4);
            for (int i=0;i<260;i++) {
                c.submit("cmd"+i); run(c,12);
                if (i==90 || i==180) { Node l=c.leader(); if (l!=null) { c.net.kill(l.id); run(c,900); c.net.revive(l.id); } }
            }
            int before = c.rejects; long t0 = c.net.now;
            c.net.revive(4);
            long caught = -1;
            for (long t=0;t<40000;t+=TICK) {
                c.step(TICK);
                if (c.nodes.get(4).last().index() >= c.maxCommit() && c.nodes.get(4).commitIndex >= c.maxCommit()) {
                    caught = c.net.now - t0; break;
                }
            }
            tableRow(new String[]{slow?"朴素回退（每次 −1）":"快速回退（按任期跳）",
                    ""+(c.rejects-before), caught>=0?""+caught:"未追平", caught>=0?"✓":"✗"}, w, al);
        }
        System.out.println("""

    ▸ 一个任期内的日志是同一个 Leader 连续写下的，要么整段一致要么整段不一致
      （日志匹配性质）。所以在冲突任期内部逐条回退是纯粹的浪费。
    ▸ 真实系统里日志常有几万条却只跨越几个任期，这个优化能把几万轮 RPC 降到个位数。
      落后节点追不上，等于集群实际少了一个副本 —— 容错能力悄悄下降了。""".indent(4));
    }

    // ─── Lab 3B-5 ─────────────────────────────────────────────────────────
    // 补上 Lab 3-A 的一个已知简化：3-A 的模型没有日志，投票时缺了
    // 「candidate 的日志至少和我一样新」这一条。下面用 A/B/C 对照测出它的后果。
    record Out(int minorityWon, int majorityWon, int noLeader, int viol, int maxMinTerm) {}

    static Out run40(boolean noLogCheck, boolean writeDuringPartition) {
        int min=0, maj=0, none=0, viol=0, maxT=0;
        for (long seed=1; seed<=40; seed++) {
            Network net = new Network(5, 30, 15, 0, seed);
            Cluster c = new Cluster(5, net, 400, 300, seed);
            c.noLogCheck = noLogCheck;
            java.util.function.LongConsumer adv = ms -> { for (long t=0;t<ms;t+=TICK) c.step(TICK); };
            adv.accept(2000);
            for (int i=0;i<20;i++) { c.submit("pre"+i); adv.accept(30); }   // 分区前先提交一批
            adv.accept(800);
            int[] grp = {0,0,0,1,1};                                        // 多数派{N1,N2,N3}｜少数派{N4,N5}
            net.partition(grp);
            if (writeDuringPartition)
                for (int i=0;i<25;i++) { c.submit("during"+i); adv.accept(40); }
            adv.accept(12000);                                              // 少数派任期疯涨，日志冻结
            maxT = Math.max(maxT, c.nodes.get(3).term);
            net.heal();
            adv.accept(9000);
            Node l = c.leader();
            if (l == null) none++;
            else if (grp[l.id] == 1) min++;
            else maj++;
            if (!c.violations.isEmpty()) viol++;
        }
        return new Out(min, maj, none, viol, maxT);
    }

    static void lab5() {
        head("3B-5", "高任期节点回归", "3|2 分区让少数派任期涨到几十，再恢复，看谁当选");
        Out a = run40(false, true), b = run40(true, true), d = run40(false, false);
        int[] w = {32,16,16,18}; String al = "LRRR";
        tableHead(new String[]{"恢复后由谁当选（各 40 次）","① 完整 Raft","② 无日志检查","③ 多数派没写入"}, w, al);
        tableRow(new String[]{"多数派节点（日志更新）", ""+a.majorityWon(), ""+b.majorityWon(), ""+d.majorityWon()}, w, al);
        tableRow(new String[]{"★ 少数派节点（日志陈旧）", ""+a.minorityWon(), ""+b.minorityWon(), ""+d.minorityWon()}, w, al);
        tableRow(new String[]{"没能选出 Leader", ""+a.noLeader(), ""+b.noLeader(), ""+d.noLeader()}, w, al);
        tableRow(new String[]{"安全性违反", ""+a.viol(), ""+b.viol(), ""+d.viol()}, w, al);
        System.out.println();
        row("分区期间少数派任期最高涨到", a.maxMinTerm()+"　← 而它们的日志一条都没长");
        System.out.println("""

    ▸ ① 少数派一次都赢不了。它们的任期涨到几十（回来时确实会逼现任 Leader 退位、
      触发一次不必要的选举），但日志停在分区前 —— 那条「至少和我一样新」把它们全挡住了。
      高任期节点回归是「有破坏力，没有危险」：损失的是可用性，不是安全性。
    ▸ ② 去掉日志检查后，陈旧节点频繁当选，而且【安全性真的被破坏了】——
      它当选后会用自己陈旧的日志覆盖别人已提交的条目。这正是 Lab 3-A 那个模型的行为。
    ▸ ③ 最微妙的一列：分区期间多数派没有任何写入，两边日志一模一样 ——
      少数派根本不"陈旧"，它当选完全合法也无害。
      【所以准确的说法是："日志真的落后的节点赢不了"，而不是"少数派节点赢不了"。】
      真实系统里 ③ 这种窗口几乎不存在，因为新 Leader 一当选就会追加 no-op，
      lastLogTerm 立刻变成当前任期 —— 这是 no-op 的第二重作用。
    ▸ 顺带纠正一个常见误解：选举【从不看 commitIndex】，只比 (lastLogTerm, lastLogIndex)。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  三个思考题（答案在课件 §3.12 / §3.13）：

    1. -naivecommit 下被覆盖的那条日志，在被覆盖之前有几个节点持有它？
       "过半"为什么救不了它？
    2. 如果同时打开 -naivecommit 并让新 Leader 上任就追加一条 no-op 日志，
       Figure 8 还会发生吗？先想清楚再改代码试。
    3. 把 -loss 调到 0.4，Lab 3B-2 的日志匹配断言会不会被破坏？为什么？
    4. Lab 3B-5 的 ① 里少数派一次都没赢。但如果少数派那边有一个【僵尸 Leader】
       （持续追加日志但提交不了），情况会变吗？什么条件下它能赢？（提示：no-op）

  下一节：Part 3-C · 持久化、快照、成员变更与线性一致读。""");
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
            if (k.equals("naivecommit")) { NAIVE=true; continue; }
            if (k.equals("slowbackoff")) { SLOW=true; continue; }
            if (k.equals("nologcheck")) { NOLOGCHECK=true; continue; }
            if (k.equals("v")) { VERBOSE=true; continue; }
            if (i+1>=a.length) break;
            String v=a[++i];
            switch (k) {
                case "n" -> N=Integer.parseInt(v);
                case "delay" -> DELAY=Long.parseLong(v);
                case "jitter" -> JITTER=Long.parseLong(v);
                case "loss" -> LOSS=Double.parseDouble(v);
                case "seed" -> SEED=Long.parseLong(v);
                default -> System.out.println("未知参数: "+k);
            }
        }
    }
}
