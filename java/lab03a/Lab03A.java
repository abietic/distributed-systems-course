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
    // 确定性网络模拟器
    //   没有线程、没有真实时间：只有一个虚拟时钟和一个按送达时刻排序的队列。
    //   同一个种子必定复现同一次执行 —— 这是调共识算法的前提。
    // ══════════════════════════════════════════════════════════════════════
    enum MType { REQUEST_VOTE, VOTE_REPLY, HEARTBEAT_MSG }

    static class Msg {
        int from, to; MType type; int term; boolean granted;
        long deliverAt; int seq;
        Msg(int f, int t, MType ty, int tm, boolean g) { from=f; to=t; type=ty; term=tm; granted=g; }
    }

    static class Network {
        final int n; final Random rnd;
        final long delay, jitter; final double loss;
        final int[] group; final boolean[] down;
        final List<Msg> queue = new ArrayList<>();
        long now = 0; int seq = 0;
        int sent = 0, dropped = 0, delivered = 0;

        Network(int n, long delay, long jitter, double loss, long seed) {
            this.n=n; this.delay=delay; this.jitter=jitter; this.loss=loss;
            this.rnd=new Random(seed); group=new int[n]; down=new boolean[n];
        }
        long now() { return now; }
        boolean isDown(int i) { return down[i]; }
        void kill(int i) { down[i] = true; }
        void revive(int i) { down[i] = false; }
        void partition(int[] g) { System.arraycopy(g, 0, group, 0, n); }
        void heal() { Arrays.fill(group, 0); }

        /** 丢包、分区、宕机在这里统一裁决 —— 发送方永远分不清是哪一种（Part 0 的第三态）。 */
        void send(Msg m) {
            sent++;
            if (down[m.from] || down[m.to] || group[m.from] != group[m.to]) { dropped++; return; }
            if (rnd.nextDouble() < loss) { dropped++; return; }
            long jit = jitter > 0 ? rnd.nextLong(2 * jitter) - jitter : 0;
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
                if (!down[m.to] && group[m.from] == group[m.to]) { due.add(m); delivered++; }
                else dropped++;
                return true;
            });
            return due;
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // Raft 选举状态机 —— 这一节只做三件事：任期递增、投票、心跳压制
    // ══════════════════════════════════════════════════════════════════════
    enum State { FOLLOWER, CANDIDATE, LEADER;
        String cn() { return switch (this) { case FOLLOWER -> "Follower";
            case CANDIDATE -> "Candidate"; case LEADER -> "Leader"; }; } }

    static class Node {
        int id; State state = State.FOLLOWER; int term = 0; int votedFor = NO_VOTE;
        Set<Integer> votes = new HashSet<>();
        long timer = 0, timeout = 0, hb = 0;
        Node(int id) { this.id = id; }
    }

    static class Cluster {
        final List<Node> nodes = new ArrayList<>();
        final Network net; final Random rnd;
        final long base, window; final boolean unsafe;
        int elections = 0, splits = 0, maxLeadersPerTerm = 0;
        final List<String> violations = new ArrayList<>();
        final List<String> log = new ArrayList<>();

        Cluster(int n, Network net, long base, long window, long seed, boolean unsafe) {
            this.net=net; this.base=base; this.window=window; this.unsafe=unsafe;
            this.rnd=new Random(seed + 999);
            for (int i = 0; i < n; i++) { Node nd = new Node(i); nd.timeout = newTimeout(); nodes.add(nd); }
        }

        long newTimeout() { return window <= 0 ? base : base + rnd.nextLong(window); }
        int majority() { return nodes.size() / 2 + 1; }

        void logf(String f, Object... a) {
            String s = String.format("[%6dms] ", net.now()) + String.format(f, a);
            log.add(s);
            if (VERBOSE) System.out.println("    " + s);
        }

        /** Raft 最重要的一条规则：看到更大的任期号就立刻退回 Follower。 */
        void stepDown(Node n, int term) {
            if (n.state == State.LEADER) logf("N%d 看到 term=%d，从 Leader 退位", n.id + 1, term);
            n.term = term; n.state = State.FOLLOWER; n.votedFor = NO_VOTE;
            n.votes.clear(); n.timer = 0; n.timeout = newTimeout();
        }

        void startElection(Node n) {
            n.term++; n.state = State.CANDIDATE; n.votedFor = n.id;
            n.votes = new HashSet<>(List.of(n.id));       // 先投自己一票
            n.timer = 0; n.timeout = newTimeout(); elections++;
            logf("N%d 选举超时 → Candidate，term=%d", n.id + 1, n.term);
            for (int j = 0; j < nodes.size(); j++)
                if (j != n.id) net.send(new Msg(n.id, j, MType.REQUEST_VOTE, n.term, false));
        }

        void becomeLeader(Node n) {
            n.state = State.LEADER; n.hb = HEARTBEAT;
            logf("★ N%d 当选 Leader，term=%d，得票 %d/%d", n.id + 1, n.term, n.votes.size(), nodes.size());
        }

        void handle(Msg m) {
            Node n = nodes.get(m.to);
            switch (m.type) {
                case REQUEST_VOTE -> {
                    if (m.term > n.term) stepDown(n, m.term);
                    // 投票条件：任期不过时、本任期还没投过票。
                    // ★ 这里缺了真实 Raft 的第三条：「candidate 的日志至少和我一样新」——
                    //   本 Lab 的模型没有日志，无法实现。后果：分区恢复后陈旧的少数派节点
                    //   有相当高的概率当选。Lab 3B-5 用 A/B 对照测了这件事。
                    boolean grant = m.term == n.term && n.state != State.LEADER
                            && (n.votedFor == NO_VOTE || n.votedFor == m.from);
                    if (unsafe) {                        // ★ 故意去掉「每任期一票」
                        grant = m.term >= n.term && n.state != State.LEADER;
                        if (m.term > n.term) n.term = m.term;
                    }
                    if (grant) { n.votedFor = m.from; n.timer = 0; n.timeout = newTimeout(); }
                    net.send(new Msg(n.id, m.from, MType.VOTE_REPLY, n.term, grant));
                }
                case VOTE_REPLY -> {
                    if (m.term > n.term) { stepDown(n, m.term); return; }
                    if (n.state == State.CANDIDATE && m.granted && m.term == n.term) {
                        n.votes.add(m.from);
                        if (n.votes.size() >= majority()) becomeLeader(n);
                    }
                }
                case HEARTBEAT_MSG -> {
                    if (m.term >= n.term) {
                        if (n.state == State.LEADER && m.from != n.id)
                            logf("N%d 收到 term=%d 的心跳，从 Leader 退位", n.id + 1, m.term);
                        n.term = m.term; n.state = State.FOLLOWER; n.votedFor = m.from;
                        n.votes.clear(); n.timer = 0; n.timeout = newTimeout();
                    }
                    // m.term < n.term：发送者是过时的 Leader，忽略
                }
            }
        }

        void step(long dt) {
            for (Node n : nodes) {
                if (net.isDown(n.id)) continue;
                if (n.state == State.LEADER) {
                    n.hb += dt;
                    if (n.hb >= HEARTBEAT) {
                        n.hb = 0;
                        for (int j = 0; j < nodes.size(); j++)
                            if (j != n.id) net.send(new Msg(n.id, j, MType.HEARTBEAT_MSG, n.term, false));
                    }
                } else {
                    n.timer += dt;
                    if (n.timer >= n.timeout) {
                        if (n.state == State.CANDIDATE) { splits++; logf("N%d 超时仍未过半 → 分裂投票", n.id + 1); }
                        startElection(n);
                    }
                }
            }
            for (Msg m : net.advance(dt)) if (!net.isDown(m.to)) handle(m);
            checkSafety();
        }

        /**
         * Election Safety：任何一个任期内至多只有一个 Leader。
         * 证明两步：当选需过半票 + 每任期每人只投一票 ⇒ 两个过半集合必相交 ⇒ 矛盾。
         */
        void checkSafety() {
            Map<Integer,List<Integer>> byTerm = new HashMap<>();
            for (Node n : nodes)
                if (!net.isDown(n.id) && n.state == State.LEADER)
                    byTerm.computeIfAbsent(n.term, k -> new ArrayList<>()).add(n.id + 1);
            byTerm.forEach((term, ids) -> {
                if (ids.size() > maxLeadersPerTerm) maxLeadersPerTerm = ids.size();
                if (ids.size() > 1 && violations.size() < 5)
                    violations.add("term " + term + " 同时存在 " + ids.size() + " 个 Leader：" + ids);
            });
        }

        Node leader() {
            for (Node n : nodes) if (!net.isDown(n.id) && n.state == State.LEADER) return n;
            return null;
        }
        int leadersIn(int[] grp, int want) {
            int k = 0;
            for (Node n : nodes) if (!net.isDown(n.id) && n.state == State.LEADER && grp[n.id] == want) k++;
            return k;
        }
        int maxTerm() { return nodes.stream().mapToInt(n -> n.term).max().orElse(0); }
        String summary() {
            StringBuilder sb = new StringBuilder();
            for (Node n : nodes)
                sb.append(String.format("N%d:%s(t%d) ", n.id + 1,
                        net.isDown(n.id) ? "☠" : n.state.cn().substring(0, 1), n.term));
            return sb.toString();
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    static Cluster build() {
        long win = NORANDOM ? 0 : WINDOW;
        return new Cluster(N, new Network(N, DELAY, JITTER, LOSS, SEED), BASE, win, SEED, UNSAFE);
    }
    static void run(Cluster c, long ms) { for (long t = 0; t < ms; t += TICK) c.step(TICK); }
    static String leaderName(Cluster c) {
        Node l = c.leader();
        return l == null ? "无" : String.format("N%d(t%d)", l.id + 1, l.term);
    }

    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        long win = NORANDOM ? 0 : WINDOW;
        System.out.printf("%n配置：%d 节点 ｜ 延迟 %d±%dms ｜ 丢包 %.0f%% ｜ 选举超时 %d~%dms ｜ 种子 %d%s%n",
                N, DELAY, JITTER, LOSS * 100, BASE, BASE + win, SEED, UNSAFE ? " ｜ ★ UNSAFE 模式" : "");
        lab1(); lab2(); lab3(); lab4(); epilogue();
    }

    // ─── Lab 3A-1 ─────────────────────────────────────────────────────────
    static void lab1() {
        head("3A-1", "确定性网络模拟", "调共识算法的第一件事：让 bug 能复现");
        java.util.function.LongFunction<String> sig = seed -> {
            long save = SEED; SEED = seed;
            Cluster c = build(); run(c, 3000); SEED = save;
            return String.format("Leader=%s maxTerm=%d 选举=%d 分裂=%d",
                    leaderName(c), c.maxTerm(), c.elections, c.splits);
        };
        String a = sig.apply(42), b = sig.apply(42), d = sig.apply(7);
        row("种子 42 · 第一次", a);
        row("种子 42 · 第二次", b);
        row("种子 7", d);
        row("同种子结果一致", a.equals(b) ? "✓ 完全可复现" : "✗ 不可复现（有 bug）");
        System.out.println("""

    ▸ 没有线程、没有真实时间，只有一个虚拟时钟和一个按时间排序的消息队列。
      这样任何一次诡异的执行都能靠种子复现，而不是"跑一百遍偶尔挂一次"。
    ▸ 这是 MIT 6.5840 的 labrpc、以及 etcd 的 raft 测试框架采用的同一套思路：
      把并发与时间从被测逻辑里彻底剥离出去。""".indent(4));
    }

    // ─── Lab 3A-2 ─────────────────────────────────────────────────────────
    static void lab2() {
        head("3A-2", "五个场景", "每个 tick 都断言「任一任期至多一个 Leader」");
        int[] w = {28, 20, 10, 9, 8, 16}; String al = "LLRRRR";
        tableHead(new String[]{"场景","结果 Leader","最大term","选举数","分裂","选举安全性"}, w, al);
        java.util.function.Function<Cluster,String> safe =
                c -> c.violations.isEmpty() ? "✓ 成立" : "✗ 被破坏！";

        Cluster c1 = build(); run(c1, 3000);
        tableRow(new String[]{"① 冷启动", leaderName(c1), "" + c1.maxTerm(),
                "" + c1.elections, "" + c1.splits, safe.apply(c1)}, w, al);

        Cluster c2 = build(); run(c2, 2000);
        Node l2 = c2.leader();
        if (l2 != null) { c2.net.kill(l2.id); c2.logf("外力：杀死 Leader N%d", l2.id + 1); }
        run(c2, 4000);
        tableRow(new String[]{"② 杀死 Leader 后重选", leaderName(c2), "" + c2.maxTerm(),
                "" + c2.elections, "" + c2.splits, safe.apply(c2)}, w, al);

        Cluster c3 = build(); run(c3, 2000);
        int[] grp = {0,0,0,1,1};
        c3.net.partition(grp); c3.logf("外力：网络分区 {N1,N2,N3} | {N4,N5}");
        run(c3, 6000);
        int majL = c3.leadersIn(grp, 0), minL = c3.leadersIn(grp, 1);
        tableRow(new String[]{"③ 3|2 分区", "多数派" + majL + " 少数派" + minL, "" + c3.maxTerm(),
                "" + c3.elections, "" + c3.splits, safe.apply(c3)}, w, al);

        c3.net.heal(); c3.logf("外力：分区修复"); run(c3, 5000);
        Set<Integer> terms = new HashSet<>();
        for (Node n : c3.nodes) terms.add(n.term);
        tableRow(new String[]{"④ 分区恢复后收敛", leaderName(c3), "" + c3.maxTerm(),
                "" + c3.elections, "" + c3.splits, safe.apply(c3)}, w, al);

        Cluster c5 = build();
        for (int r = 0; r < 12; r++) {
            run(c5, 1500);
            int i = c5.rnd.nextInt(N);
            if (c5.net.isDown(i)) c5.net.revive(i);
            else if (aliveCount(c5) > N / 2 + 1) c5.net.kill(i);
        }
        run(c5, 4000);
        tableRow(new String[]{"⑤ 反复随机杀/救节点", leaderName(c5), "" + c5.maxTerm(),
                "" + c5.elections, "" + c5.splits, safe.apply(c5)}, w, al);

        System.out.println();
        row("③ 多数派侧的 Leader", majL + " 个（能提交日志）");
        row("③ 少数派侧的 Leader", minL + " 个 —— "
                + (minL > 0 ? "★ 僵尸 Leader：自认为是主，但一条日志都提交不了" : "已退位"));
        row("③ 少数派能联系到几个节点", reachable(c3, grp, 1) + " 个，过半需要 " + c3.majority() + " 个 ⇒ 提交不了任何东西");
        row("④ 恢复后集群的任期数", terms.size() + " 种　← 应当收敛到 1~2 种");
        row("④ 最终状态", c3.summary());
        System.out.println("""

    ▸ 场景③是 CAP 的具体样子，但它给出的答案比教科书更微妙：
      如果分区前的 Leader 恰好落在少数派一侧，基础 Raft 不会让它主动退位 ——
      它继续给同侧的 Follower 发心跳，自认为还是 Leader。这叫「僵尸 Leader」。
      它不违反安全性（拿不到过半确认，一条日志都提交不了），但客户端把写请求
      发给它会一直超时，读请求还可能读到陈旧数据。
    ▸ 工程上用 CheckQuorum 解决：Leader 定期确认自己还能联系到过半节点，
      否则主动退位。etcd 默认开启它。
    ▸ 场景④注意少数派带回来的高任期：它一接触集群就会逼现任 Leader 退位。
      这正是 Raft 论文用 PreVote 优化解决的问题。
    ▸ 五个场景、上万个 tick，选举安全性一次都没被破坏 —— 这不是运气，
      是「过半票 + 每任期一票 ⇒ 两个过半集合必相交」这条数学事实。""".indent(4));
    }

    static int aliveCount(Cluster c) {
        int k = 0; for (Node n : c.nodes) if (!c.net.isDown(n.id)) k++; return k;
    }
    static int reachable(Cluster c, int[] grp, int g) {
        int k = 0; for (Node n : c.nodes) if (!c.net.isDown(n.id) && grp[n.id] == g) k++; return k;
    }

    // ─── Lab 3A-3 ─────────────────────────────────────────────────────────
    static void lab3() {
        head("3A-3", "随机化选举超时", "关掉它，看分裂投票怎么把集群拖住");
        int[] w = {24, 18, 12, 12, 16}; String al = "LLRRR";
        tableHead(new String[]{"选举超时策略","选出 Leader","耗时(ms)","分裂投票","发起过的选举"}, w, al);
        for (long window : new long[]{0, WINDOW}) {
            Network net = new Network(N, DELAY, JITTER, LOSS, SEED);
            Cluster c = new Cluster(N, net, BASE, window, SEED, false);
            long elected = -1;
            for (long t = 0; t < 20000; t += TICK) {
                c.step(TICK);
                if (elected < 0 && c.leader() != null) elected = net.now();
            }
            tableRow(new String[]{window == 0 ? "固定超时（窗口 0）" : "随机超时（窗口 " + WINDOW + "）",
                    c.leader() != null ? "✓ " + leaderName(c) : "✗ 一直没有",
                    elected >= 0 ? "" + elected : "未能选出",
                    "" + c.splits, "" + c.elections}, w, al);
        }
        System.out.println("""

    ▸ 固定超时下所有节点同时醒来、同时给自己投票，谁也拿不到过半票，
      超时后再来一轮 —— 任期一路飙升却选不出 Leader。这就是活锁。
    ▸ 随机化只用一个随机数就打破了对称性：总有人先醒，赶在别人之前收齐票。
    ▸ 注意这不是"随机化让选举更快"，而是"随机化让选举能够终止"。
      FLP 说确定性算法在异步系统里不能保证终止，Raft 换到了随机化这条赛道上，
      代价是只保证「以概率 1 终止」而非「N 步内一定终止」。工程上这就够了。""".indent(4));
    }

    // ─── Lab 3A-4 ─────────────────────────────────────────────────────────
    static void lab4() {
        head("3A-4", "把安全性拆掉给你看", "去掉「每个任期只投一票」这一条，其他不变");
        int[] w = {28, 22, 18, 16, 14}; String al = "LRRRR";
        tableHead(new String[]{"规则","同任期最多 Leader 数","违反的种子","选出了 Leader","选举安全性"}, w, al);
        final int TRIALS = 60;
        for (boolean unsafe : new boolean[]{false, true}) {
            int worst = 0, violations = 0, elected = 0;
            for (long seed = 1; seed <= TRIALS; seed++) {
                Network net = new Network(N, DELAY, 40, 0.15, seed);
                Cluster c = new Cluster(N, net, 300, 90, seed, unsafe);
                run(c, 6000);
                worst = Math.max(worst, c.maxLeadersPerTerm);
                if (!c.violations.isEmpty()) violations++;
                if (c.leader() != null) elected++;
            }
            tableRow(new String[]{unsafe ? "② 去掉「每任期一票」" : "① 完整 Raft（每任期一票）",
                    worst + " 个", violations + "/" + TRIALS + " 个种子",
                    elected + "/" + TRIALS, violations > 0 ? "✗ 被破坏" : "✓ 成立"}, w, al);
        }
        System.out.println("""

    ▸ ① 60 个种子、上万个 tick，同一任期的 Leader 数最多就是 1 —— 因为两个过半
      集合必然相交，交集里那个节点在同一任期投了两票，这是不可能的。
    ▸ ② 一旦允许一个节点在同一任期投多票，两个 Candidate 就能各自凑齐过半，
      同一任期出现两个 Leader —— 这就是脑裂，两个 Leader 各写各的日志，
      数据从此分叉且无法自动合并。
    ▸ 所以「每任期只投一票」不是实现细节，它和「过半」一起构成了 Raft 全部
      安全性的地基。而且 votedFor 必须持久化 —— 节点重启后如果忘了自己投过票，
      同样会破坏它（Lab 3-C 会处理）。试试 java Lab03A.java -unsafe -v。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  三个思考题（答案在课件 §3.6 / §3.7）：

    1. 把 -loss 调到 0.5，集群还能选出 Leader 吗？需要多久？
       再调到 0.8 呢？在什么丢包率下它彻底选不出来？
    2. -norandom 时任期会一路飙升。如果这时候网络突然恢复正常，
       这个高任期会对集群造成什么影响？（提示：PreVote）
    3. 场景③里少数派的两个节点，任期涨到了多少？
       如果分区持续一小时，它们的任期会涨到什么量级？这有害吗？

  下一节：Part 3-B · 日志复制与五条安全性属性。""");
        System.out.println("═".repeat(78) + "\n");
    }

    // ─── 输出工具 ─────────────────────────────────────────────────────────
    static int dispw(String s) {
        int w = 0;
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            w += (c >= 0x1100 && (c <= 0x115F
                    || (c >= 0x2E80 && c <= 0xA4CF && c != 0x303F)
                    || (c >= 0xAC00 && c <= 0xD7A3) || (c >= 0xF900 && c <= 0xFAFF)
                    || (c >= 0xFE30 && c <= 0xFE6F) || (c >= 0xFF00 && c <= 0xFF60)
                    || (c >= 0xFFE0 && c <= 0xFFE6))) ? 2 : 1;
        }
        return w;
    }
    static String padR(String s, int w) { int d = w - dispw(s); return d > 0 ? s + " ".repeat(d) : s; }
    static String padL(String s, int w) { int d = w - dispw(s); return d > 0 ? " ".repeat(d) + s : s; }
    static String layout(String[] c, int[] w, String al) {
        StringBuilder sb = new StringBuilder("    ");
        for (int i = 0; i < c.length; i++)
            sb.append(i < al.length() && al.charAt(i) == 'L' ? padR(c[i], w[i]) : padL(c[i], w[i]));
        return sb.toString().stripTrailing();
    }
    static void head(String label, String title, String sub) {
        System.out.println("\n" + "═".repeat(78));
        System.out.printf("  Lab %s · %s%n  %s%n", label, title, sub);
        System.out.println("═".repeat(78));
    }
    static void row(String k, Object v) { System.out.println("    " + padR(k, 30) + " " + v); }
    static void tableHead(String[] cols, int[] w, String al) {
        System.out.println(layout(cols, w, al));
        int t = 0; for (int x : w) t += x;
        System.out.println("    " + "─".repeat(t));
    }
    static void tableRow(String[] c, int[] w, String al) { System.out.println(layout(c, w, al)); }

    static void parseArgs(String[] a) {
        for (int i = 0; i < a.length; i++) {
            String k = a[i].replaceFirst("^-+", "");
            if (k.equals("norandom")) { NORANDOM = true; continue; }
            if (k.equals("unsafe"))   { UNSAFE = true; continue; }
            if (k.equals("v"))        { VERBOSE = true; continue; }
            if (i + 1 >= a.length) break;
            String v = a[++i];
            switch (k) {
                case "n"      -> N = Integer.parseInt(v);
                case "delay"  -> DELAY = Long.parseLong(v);
                case "jitter" -> JITTER = Long.parseLong(v);
                case "loss"   -> LOSS = Double.parseDouble(v);
                case "base"   -> BASE = Long.parseLong(v);
                case "window" -> WINDOW = Long.parseLong(v);
                case "seed"   -> SEED = Long.parseLong(v);
                default       -> System.out.println("未知参数: " + k);
            }
        }
    }
}
