/*
 * Lab 3-D · Raft 附录：索引、读、配置与日志（Java 版）
 *
 * 运行：  cd java/lab03d && java Lab03D.java
 * 调参：  java Lab03D.java -applylag 8 -reads 800 -seeds 60 -writers 4
 *
 * 需要 JDK 17+。配套课件：courseware/ch03d-faq.html
 *
 * 这一章和 3-A/3-B/3-C 不同：那三章要跑完整的 Raft 状态机，
 * 因为它们研究的是【算法本身】会不会出错。
 * 3-D 研究的是【算法与系统之间的接缝】——
 * 读路径怎么和状态机对齐、新 Leader 的 commitIndex 从哪来、
 * quorum 和日志复制到底差在哪、配置从哪来。
 * 这些问题用一个聚焦的小模型讲得更清楚。
 *
 * 输出与 Go 版逐字对应（连随机数都用同一个 xorshift32），可以直接对照阅读。
 */
import java.util.*;

public class Lab03D {

    static int APPLYLAG = 4, READS = 500, SEEDS = 40, WRITERS = 2;
    static int N = 5, W = 3, R = 3, ENTRIES = 3;
    static double LOSS = 0.22;
    static long SEED = 42;

    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：N=%d W=%d R=%d ｜ applylag=%d ｜ 种子基准 %d%n", N, W, R, APPLYLAG, SEED);
        lab1(); lab2(); lab3(); lab4(); epilogue();
    }

    // ══════════════════════════════════════════════════════════════════════
    // 随机数：最简单的 xorshift32
    //
    // 为什么不用 java.util.Random：Go 版要产出【逐字节相同】的输出，
    // 两边各自的标准库随机数是对不上的。自己写 8 行，两边一模一样。
    // ══════════════════════════════════════════════════════════════════════
    static final class Rnd {
        int s;
        Rnd(long seed) { s = (int) seed; if (s == 0) s = 1; }
        int next() { s ^= s << 13; s ^= s >>> 17; s ^= s << 5; return s; }
        double flt() { return (next() & 0xFFFFFFFFL) / 4294967296.0; }
        int intn(int n) { return (int) (flt() * n); }   // 用高位，避免低位周期性
        <T> void shuffle(List<T> xs) {
            for (int i = xs.size() - 1; i > 0; i--) Collections.swap(xs, i, intn(i + 1));
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // Lab 3D-1：ReadIndex 的第三步
    // ══════════════════════════════════════════════════════════════════════

    /** 一个被极度简化的 Leader：只保留这条读路径真正依赖的两个量。 */
    static final class LeaderState {
        int commitIndex, lastApplied;
        LeaderState(int c, int a) { commitIndex = c; lastApplied = a; }
    }
    record ReadResult(int got, int want, int waited, boolean stale) {}

    /**
     * 执行一次 ReadIndex 读。三步全在这里，一步不少：
     *   step1  readIndex := commitIndex
     *   step2  发一轮心跳确认自己还是 Leader          （本模型里恒成功——我们要隔离出第三步）
     *   step3  等 lastApplied >= readIndex 再读状态机  ← waitApply 控制开不开
     *
     * 之所以让 step2 恒成功，是为了证明一件事：
     * 【前两步全部正确通过，读出来依然可能是旧数据。】
     */
    static ReadResult doRead(LeaderState l, boolean waitApply) {
        int readIndex = l.commitIndex;              // step 1
        // step 2：心跳确认（本模型中无分区，恒通过）
        int got, waited = 0;
        if (waitApply) {
            waited = Math.max(0, readIndex - l.lastApplied);   // step 3：阻塞等 apply 追上
            l.lastApplied = readIndex;
            got = readIndex;
        } else {
            got = l.lastApplied;                    // 跳过 step 3，直接读状态机
        }
        return new ReadResult(got, readIndex, waited, got != readIndex);
    }

    static void lab1() {
        head("3D-1", "ReadIndex 的第三步", "前两步全部通过，少了第三步会读到什么");
        System.out.print(blk("""

    模型：一个没有分区的 Leader。它一直合法，心跳永远能收到过半响应
    ——也就是说 ReadIndex 的【第 1 步和第 2 步始终完美通过】。
    唯一的变量是 apply 线程滞后多少条（commitIndex - lastApplied）。""", 4));

        int[] w = {14, 3, 14, 16, 3, 14, 16};
        String al = "LLRRLRR";
        tableHead(new String[]{"apply 滞后", "", "跳过第3步", "└ 陈旧读", "", "执行第3步", "└ 平均等待"}, w, al);

        for (int lag = 0; lag <= APPLYLAG * 2; lag++) {
            if (lag > 0 && lag != APPLYLAG && lag % 2 == 1 && lag != 1) continue;
            int staleNo = 0, staleYes = 0, waited = 0;
            Rnd rnd = new Rnd(SEED * 1000 + lag * 17L + 3);
            for (int i = 0; i < READS; i++) {
                int commit = 10 + i;
                int applied = commit - rnd.intn(lag + 1);
                if (doRead(new LeaderState(commit, applied), false).stale()) staleNo++;
                ReadResult r2 = doRead(new LeaderState(commit, applied), true);
                if (r2.stale()) staleYes++;
                waited += r2.waited();
            }
            tableRow(new String[]{lag + " 条以内", lag == APPLYLAG ? "◀" : "",
                    staleNo + "/" + READS, pct(staleNo, READS), "",
                    staleYes + "/" + READS, String.format("%.2f 条", (double) waited / READS)}, w, al);
        }
        System.out.print(blk("""

    结论
      · 跳过第 3 步的陈旧读比例 ≈ lag/(lag+1) —— 滞后越大越糟，而且【没有任何报错】。
      · 执行第 3 步的陈旧读恒为 0，代价是平均多等几条 apply。
      · 生产上这条曲线对应的现象就是：写接口返回成功，立刻查却查不到。
        滞后被撑大的典型原因是状态机执行慢（大事务 / 写放大 / compaction 抢 IO）。
      · 再说一遍：这一整张表里，ReadIndex 的第 1、2 步【全部通过】。
        确认自己是 Leader、拿到正确的提交点，都不能替代「等状态机追上来」。""", 4));
    }

    // ══════════════════════════════════════════════════════════════════════
    // Lab 3D-2：no-op —— 新 Leader 的 commitIndex 从哪来
    // ══════════════════════════════════════════════════════════════════════
    enum Outcome { FRESH, STALE, QUEUED, ALLOWED }

    /**
     * 复现这样一个场景：
     *   任期 2：Leader L 把 k 条写复制到了过半节点。
     *           ① 有时它来得及推进自己的 commitIndex 并【回复客户端成功】，然后才崩；
     *           ② 有时它还没来得及推进 commitIndex 就崩了。
     *           无论哪种，它都【没来得及广播 leaderCommit】——
     *           所以「已提交」这个事实随它一起消失。
     *   任期 3：某个持有这些日志的节点当选。它的 commitIndex = 0。
     *           受 Figure 8 限制，它【不能】直接提交任期 2 的条目。
     *   然后客户端来读。
     */
    static Outcome runNoop(int k, boolean ackedBeforeCrash, boolean noop, boolean readEarly) {
        int acked = ackedBeforeCrash ? k : 0;
        int newLeaderCommit = 0;   // 易失，不会从旧 Leader 传过来
        if (noop) {
            if (readEarly) {
                // no-op 还在复制中。etcd raft 此时把读请求【挂起排队】：
                //   if (!r.committedEntryInCurrentTerm()) r.pendingReadIndexMessages.add(m)
                // 等本任期第一条提交后再放行 —— 那时 commitIndex 已经到位，返回的就是最新值。
                // （也有实现选择直接拒绝、让客户端重试；两种都安全，区别只在延迟。）
                return Outcome.QUEUED;
            }
            // no-op 是【当前任期】的条目 ⇒ 可以直接提交；
            // 提交是前缀性的 ⇒ 任期 2 的 k 条被顺带提交。
            newLeaderCommit = k + 1;
        }
        // 不发 no-op：没有新写入进来，commitIndex 就永远停在 0。
        int got = Math.min(newLeaderCommit, k);
        if (got == k) return Outcome.FRESH;
        if (acked > got) return Outcome.STALE;   // 客户端收到过成功却读不到 ⇒ 线性一致性破裂
        return Outcome.ALLOWED;                  // 那次写从未 ack，返回旧值是允许的
    }

    static void lab2() {
        head("3D-2", "no-op 条目的必要性", "新 Leader 不发空日志，ReadIndex 三步走完照样读到旧值");
        System.out.printf("""

    场景（每个种子随机决定两个分支）：
      任期 2：Leader 把 %d 条写复制到了过半节点，然后崩溃。
              分支 ①（约一半概率）它崩之前【已经推进 commitIndex 并回复客户端成功】；
              分支 ②                它还没来得及推进 commitIndex。
              两种情况下它都没来得及广播 leaderCommit ——
              「已提交」这个事实随它一起消失了。
      任期 3：某个持有这些日志的节点当选，它的 commitIndex = 0。
              受 Figure 8 限制，它【不能】直接提交任期 2 的条目。
      然后客户端用完整的 ReadIndex（三步一步不少）来读。
""", ENTRIES);

        int[] w = {22, 18, 16, 18, 22};
        String al = "LRRRR";
        tableHead(new String[]{"新 Leader 上任后", "读到旧值(违规)", "返回最新值", "排队后返回最新", "旧值但未ack(允许)"}, w, al);

        int[] a = runSeeds(false), b = runSeeds(true);
        tableRow(new String[]{"不发 no-op", "" + a[0], "" + a[1], "" + a[2], "" + a[3]}, w, al);
        tableRow(new String[]{"发 no-op（正确）", "" + b[0], "" + b[1], "" + b[2], "" + b[3]}, w, al);

        System.out.printf("""

    结论
      · 不发 no-op：%d/%d 个种子读到了旧值，而客户端【收到过写入成功】——
        这是货真价实的线性一致性破裂，且 ReadIndex 三步一步没少、全部通过。
        问题出在第 1 步拿到的 commitIndex 本身就是错的。
      · 发 no-op：违规 0 次。no-op 是【当前任期】的条目，可以直接提交；
        提交是前缀性的 ⇒ 任期 2 那几条被顺带提交 ⇒ commitIndex 到位。
      · 那 %d 次「排队」也很重要：no-op 还没提交时，etcd 把读请求挂起
        （pendingReadIndexMessages），等本任期第一条提交后再处理——那时 commitIndex 已经到位。
        也有实现选择直接拒绝、让客户端重试。两种都安全，
        绝不能做的是拿一个还没到位的 commitIndex 去凑合。
      · 「旧值但未 ack」的那些不算违规：那次写从未收到成功响应，
        它的结果本来就是 Part 0 说的「第三态」——客户端必须自己去查证。
""", a[0], SEEDS, b[2]);
    }

    /** 返回 [stale, fresh, queued, allowed] */
    static int[] runSeeds(boolean noop) {
        int[] r = new int[4];
        Rnd rnd = new Rnd(SEED * 31);
        for (int s = 0; s < SEEDS; s++) {
            boolean acked = rnd.flt() < 0.5;    // 分支 ① / ②
            boolean early = rnd.flt() < 0.35;   // 读请求到得早不早
            switch (runNoop(ENTRIES, acked, noop, early)) {
                case STALE -> r[0]++;
                case FRESH -> r[1]++;
                case QUEUED -> r[2]++;
                case ALLOWED -> r[3]++;
            }
        }
        return r;
    }

    // ══════════════════════════════════════════════════════════════════════
    // Lab 3D-3：Dynamo 式 Quorum vs Raft 日志复制
    // ══════════════════════════════════════════════════════════════════════
    record QOut(int[] replicas, int distinct, int readVers, int writeOK) {}

    /**
     * 模拟 c 个客户端【同时】写同一个 key。
     * 每个写都发往全部 n 个副本（Dynamo 的做法），收到 w 个 ack 就返回成功。
     * 关键在于：各副本收到这些写的【顺序是不同的】，而副本本身不带版本信息，
     * 于是每个副本保存的是「最后到达的那个」。
     * 这就是为什么并发写会在副本间留下分歧——和 W 设多大毫无关系。
     */
    static QOut runDynamo(int n, int w, int r, int c, double loss, Rnd rnd) {
        boolean[][] deliv = new boolean[c][n];
        for (int i = 0; i < c; i++) for (int j = 0; j < n; j++) deliv[i][j] = rnd.flt() > loss;
        int writeOK = 0;
        for (int i = 0; i < c; i++) {
            int acks = 0;
            for (int j = 0; j < n; j++) if (deliv[i][j]) acks++;
            if (acks >= w) writeOK++;
        }
        int[] rep = new int[n];
        for (int j = 0; j < n; j++) {
            List<Integer> got = new ArrayList<>();
            for (int i = 0; i < c; i++) if (deliv[i][j]) got.add(i);
            rnd.shuffle(got);
            rep[j] = got.isEmpty() ? -1 : got.get(got.size() - 1);   // 最后到达的胜出
        }
        Set<Integer> seen = new HashSet<>();
        for (int v : rep) if (v >= 0) seen.add(v);
        List<Integer> idx = new ArrayList<>();
        for (int i = 0; i < n; i++) idx.add(i);
        rnd.shuffle(idx);
        Set<Integer> rs = new HashSet<>();
        for (int i = 0; i < r && i < n; i++) { int v = rep[idx.get(i)]; if (v >= 0) rs.add(v); }
        return new QOut(rep, seen.size(), rs.size(), writeOK);
    }

    /**
     * 把同样的 c 个并发写灌进 Raft。
     * 它们全都先到 Leader，Leader 按到达先后分配 index —— 顺序在写入之前就定死了。
     * 所以这个函数不需要随机：分歧数恒为 1，这不是概率，是结构。
     */
    static QOut runRaft(int n, int c, Rnd rnd) {
        List<Integer> order = new ArrayList<>();
        for (int i = 0; i < c; i++) order.add(i);
        rnd.shuffle(order);
        int winner = order.get(c - 1);
        int[] rep = new int[n];
        Arrays.fill(rep, winner);
        return new QOut(rep, 1, 1, c);
    }

    static void lab3() {
        head("3D-3", "Quorum 读写 ≠ Raft 日志复制", "同一组并发写灌进两个世界，数副本间的分歧");
        System.out.printf("""

    %d 个客户端【同时】写同一个 key，N=%d，丢包率 %.0f%%。
      ① Dynamo 式：每个写发往全部副本、收到 W 个 ack 即成功，副本不带版本信息。
      ② Raft    ：全部请求先到 Leader，Leader 按到达先后分配 index 再复制。
""", WRITERS, N, LOSS * 100);

        int[] w = {8, 2, 16, 16, 18, 3, 14, 14};
        String al = "LLRRRLRR";
        tableHead(new String[]{"写 W", "", "Dynamo 写成功", "└ 平均分歧", "└ 出现分歧的种子",
                "", "Raft 写成功", "└ 平均分歧"}, w, al);

        for (int ww = 1; ww <= N; ww++) {
            int sumD = 0, seedsWithD = 0, okD = 0, sumR = 0, okR = 0;
            for (int s = 1; s <= SEEDS; s++) {
                QOut d = runDynamo(N, ww, R, WRITERS, LOSS, new Rnd(SEED + s * 2654435761L));
                sumD += d.distinct(); okD += d.writeOK();
                if (d.distinct() > 1) seedsWithD++;
                QOut r = runRaft(N, WRITERS, new Rnd(SEED + s * 2654435761L));
                sumR += r.distinct(); okR += r.writeOK();
            }
            tableRow(new String[]{ww + (ww == W ? " ◀" : ""), "",
                    String.format("%.2f/%d", (double) okD / SEEDS, WRITERS),
                    String.format("%.2f 种", (double) sumD / SEEDS),
                    seedsWithD + "/" + SEEDS, "",
                    String.format("%.2f/%d", (double) okR / SEEDS, WRITERS),
                    String.format("%.2f 种", (double) sumR / SEEDS)}, w, al);
        }

        QOut d = runDynamo(N, W, R, WRITERS, LOSS, new Rnd(SEED + 7 * 2654435761L));
        QOut r = runRaft(N, WRITERS, new Rnd(SEED + 7 * 2654435761L));
        System.out.printf("%n    随便挑一个种子看细节（W=%d）：%n", W);
        System.out.printf("      Dynamo 各副本最终保存： %s   ⇒ %d 种不同的值，读 R=%d 拿回 %d 个版本%n",
                vals(d.replicas()), d.distinct(), R, d.readVers());
        System.out.printf("      Raft   各副本最终保存： %s   ⇒ %d 种，读 Leader 一个拿回 %d 个版本%n",
                vals(r.replicas()), r.distinct(), r.readVers());

        System.out.print(blk("""

    结论
      · 看第一列和第三列的对比：W 从 1 调到 N，【写成功数在掉】（要求的 ack 更多了），
        但【分歧数一点没变】。W 决定的只是写端何时收到 ack，
        和副本之间是否一致毫无关系——把 W 调到 N 也救不了。
        分歧的成因是【各副本看到的到达顺序不同】，而副本没有版本信息去裁决。
      · Raft 的分歧恒为 1。不是概率上恰好，是结构上不可能：
        顺序在请求到达 Leader 的那一刻就定死了，复制的是【同一个 index 上的同一条 entry】。
      · 所以「过半」在两边的含义完全不同：
          Dynamo：读集合 ∩ 写集合 ≠ ∅  ⇒ 至少碰到一个持有最新值的副本（但不知道哪个是最新）
          Raft  ：任意两次成功操作的参与集合必相交 ⇒ 新 Leader 必持有全部已提交日志
        底层数学是同一条，用法完全相反。
      · 顺便：Raft 里没有「quorum 读」。ReadIndex 那轮心跳拿回的不是数据，
        是【我还是不是 Leader】这一个 bit。""", 4));
    }

    // ══════════════════════════════════════════════════════════════════════
    // Lab 3D-4：冷启动 —— 配置从哪来
    // ══════════════════════════════════════════════════════════════════════

    /** 一个节点的启动参数，以及它 data dir 的状态。 */
    record NodeCfg(String id, List<String> cluster, String state, String token,
                   boolean hasData, boolean wiped, int term) {
        static NodeCfg of(String id, List<String> cl, String st, String tk) {
            return new NodeCfg(id, cl, st, tk, false, false, 0);
        }
    }
    /** 运维按顺序做的一个动作：start 启动节点 | add 执行 member add | write 集群确认了 N 条写 */
    record Step(String kind, String node, int n) {
        static Step start(String id) { return new Step("start", id, 0); }
        static Step add(String id) { return new Step("add", id, 0); }
        static Step write(int n) { return new Step("write", "", n); }
    }

    /**
     * 实际形成的一个 Raft 集群。
     * key 是集群身份。etcd 里它是【确定性】算出来的：
     *   成员 ID = hash(排序后的 peer URL + --initial-cluster-token)   （静态引导时不带时间戳）
     *   集群 ID = hash(所有成员 ID)
     * 所以「同一个 token + 同一份成员列表」⇒ 同一个集群 ID。member add 改变成员配置，但不改变集群 ID。
     */
    static final class Cluster {
        String key; List<String> members; Map<String, Boolean> started = new HashMap<>();
        int term; boolean reset; int writes; String winner = ""; int idx;
        int started() { return started.size(); }
        int need() { return members.size() / 2 + 1; }
        boolean leader() { return started() >= need(); }
        String label() { return "集群#" + idx + "{" + String.join(",", members) + "}"; }
    }
    static final class BootResult {
        List<Cluster> clusters = new ArrayList<>();
        List<String> refused = new ArrayList<>(), rejected = new ArrayList<>(), log = new ArrayList<>();
        int lost;
        void logf(String f, Object... a) { log.add(String.format(f, a)); }
    }
    static String clusterKey(String token, List<String> list) {
        List<String> cl = new ArrayList<>(list); Collections.sort(cl);
        return token + "/" + String.join(",", cl);
    }
    static boolean sameSet(List<String> a, List<String> b) {
        if (a.size() != b.size()) return false;
        List<String> x = new ArrayList<>(a), y = new ArrayList<>(b);
        Collections.sort(x); Collections.sort(y);
        return x.equals(y);
    }
    /** 和 Go 的 %v 打印 []string 一模一样：[n1 n2 n3] */
    static String goList(List<String> xs) { return "[" + String.join(" ", xs) + "]"; }

    /**
     * 按顺序执行一串运维动作，模拟 etcd 实际会做的判断。三条规则（都来自 etcd 源码，见课件 §3.28）：
     *  1. data dir 里有数据 ⇒ --initial-* 全部忽略，按 data dir 回到原来的集群。
     *  2. state=new 且 data dir 为空 ⇒ 先问列表里的其他成员「我是不是已经被引导过」
     *     （isMemberBootstrapped）。有知情者说「是」⇒ 拒绝启动；问不到人 ⇒ 放行，按参数引导。
     *  3. state=existing 且 data dir 为空 ⇒ 不引导任何东西，去 peer 那里拉当前成员配置，
     *     和自己的 --initial-cluster 比对（ValidateClusterAndAssignIDs），对得上才加入。
     *     Raft 启动时【不带初始成员】（RestartNode），配置等 Leader 发过来。
     */
    static BootResult bootstrap(List<NodeCfg> nodes, List<Step> steps) {
        Map<String, NodeCfg> cfg = new HashMap<>();
        for (NodeCfg n : nodes) cfg.put(n.id(), n);
        Map<String, Cluster> byKey = new HashMap<>();
        BootResult r = new BootResult();
        java.util.function.BiFunction<String, List<String>, Cluster> getOrCreate = (key, members) -> {
            Cluster c = byKey.get(key);
            if (c != null) return c;
            c = new Cluster(); c.key = key; c.members = new ArrayList<>(members); c.idx = r.clusters.size() + 1;
            byKey.put(key, c); r.clusters.add(c);
            return c;
        };
        java.util.function.Function<String, Cluster> clusterOf = id -> {
            for (Cluster c : r.clusters) if (c.started.containsKey(id)) return c;
            return null;
        };
        for (Step st : steps) {
            NodeCfg n = cfg.get(st.node());
            switch (st.kind()) {
                case "start" -> {
                    if (n.hasData()) {
                        Cluster c = getOrCreate.apply(clusterKey(n.token(), n.cluster()), n.cluster());
                        boolean before = c.leader();
                        c.started.put(n.id(), true);
                        r.logf("启动 %s：data dir 有数据（任期 %d）⇒ 忽略 --initial-*，回到原集群", n.id(), n.term());
                        if (c.reset && c.leader() && n.term() > c.term) {
                            // 旧节点回归：任期更高 ⇒ 现任 Leader 下台；日志也更新 ⇒ 它当选，覆盖任期倒退期间写下的东西
                            r.logf("        ⇒ %s 的任期 %d > 现任 Leader 的 %d ⇒ 现任 Leader 下台", n.id(), n.term(), c.term);
                            r.logf("        ⇒ %s 的日志更新（最后一条任期 %d > %d）⇒ 当选，覆盖其他节点的日志", n.id(), n.term(), c.term);
                            if (c.writes > 0) {
                                r.logf("        ⇒ 任期 %d 里确认过的 %d 条写【全部被覆盖】", c.term, c.writes);
                                r.lost += c.writes;
                            }
                            c.term = n.term() + 1; c.writes = 0; c.winner = n.id();
                        } else if (!before && c.leader()) {
                            if (c.term == 0) c.term = n.term() + 1;
                            r.logf("        ⇒ %s 凑齐过半（%d/%d），选出 Leader", c.label(), c.started(), c.members.size());
                        } else if (!c.leader()) {
                            r.logf("        ⇒ %s 只有 %d/%d 在线，凑不齐过半，无法选主", c.label(), c.started(), c.members.size());
                        }
                    } else if (n.state().equals("new")) {
                        String key = clusterKey(n.token(), n.cluster());
                        Cluster ex = byKey.get(key);
                        if (ex != null) {
                            String knows = "";
                            for (String m : ex.members)
                                if (!m.equals(n.id()) && ex.started.containsKey(m) && cfg.get(m).hasData()) { knows = m; break; }
                            if (!knows.isEmpty()) {
                                r.logf("启动 %s（state=new）：先问 %s「我是不是已经被引导过」⇒ %s 说是", n.id(), knows, knows);
                                r.logf("        ⇒ 拒绝启动：member %s has already been bootstrapped", n.id());
                                r.refused.add(n.id());
                                continue;
                            }
                        }
                        boolean existed = ex != null;
                        int others = r.clusters.size();
                        Cluster c = getOrCreate.apply(key, n.cluster());
                        if (n.wiped()) c.reset = true;
                        boolean before = c.leader();
                        c.started.put(n.id(), true);
                        if (n.cluster().size() == 1)
                            r.logf("启动 %s（state=new，列表只有自己）⇒ 引导一个单成员集群 %s", n.id(), c.label());
                        else
                            r.logf("启动 %s（state=new）：列表里没有能证明它被引导过的在线成员 ⇒ 按参数引导 %s", n.id(), c.label());
                        if (!existed && others > 0)
                            r.logf("        ⇒ token 或成员列表和已有的集群不同 ⇒ 成员 ID、集群 ID 都不同 ⇒ 这是另一个集群，两边互不承认");
                        if (!before && c.leader() && c.term == 0) {
                            c.term = 2;
                            r.logf("        ⇒ %s 凑齐过半（%d/%d），选出 Leader，任期 %d", c.label(), c.started(), c.members.size(), c.term);
                        }
                        if (before && c.leader())
                            r.logf("        ⇒ 加入，%d/%d 在线", c.started(), c.members.size());
                        if (!c.leader())
                            r.logf("        ⇒ %s 只有 %d/%d 在线，凑不齐过半，无法选主", c.label(), c.started(), c.members.size());
                    } else if (n.state().equals("existing")) {
                        Cluster c = null;
                        for (String p : n.cluster()) if (!p.equals(n.id())) { c = clusterOf.apply(p); if (c != null) break; }
                        if (c == null) {
                            r.logf("启动 %s（state=existing）：列表里的 peer 都连不上 ⇒ 拒绝启动：cannot fetch cluster info", n.id());
                            r.refused.add(n.id());
                            continue;
                        }
                        if (!sameSet(n.cluster(), c.members)) {
                            r.logf("启动 %s（state=existing）：自己的列表 %s ≠ 集群当前成员 %s", n.id(), goList(n.cluster()), goList(c.members));
                            if (n.cluster().size() != c.members.size()) r.logf("        ⇒ 拒绝启动：member count is unequal");
                            else r.logf("        ⇒ 拒绝启动：PeerURLs: no match found for existing member");
                            r.refused.add(n.id());
                            continue;
                        }
                        boolean before = c.leader();
                        c.started.put(n.id(), true);
                        r.logf("启动 %s（state=existing）：从 peer 拉到成员配置 %s，与自己的参数一致 ⇒ 加入", n.id(), goList(c.members));
                        r.logf("        ⇒ Raft 启动时不带初始成员，配置随 Leader 的快照/日志到达（%d/%d 在线）", c.started(), c.members.size());
                        if (!before && c.leader()) r.logf("        ⇒ 重新凑齐过半，恢复服务");
                    }
                }
                case "add" -> {
                    Cluster c = null;
                    for (Cluster cc : r.clusters) if (cc.leader()) { c = cc; break; }
                    if (c == null) {
                        r.logf("member add %s ⇒ 没有能服务的 Leader，配置变更提交不了", st.node());
                        r.rejected.add(st.node());
                        continue;
                    }
                    // IsReadyToAddVotingMember：加完之后，已启动的成员必须还能凑齐过半
                    int nmembers = c.members.size() + 1, nstarted = c.started();
                    if (!(nstarted == 1 && nmembers == 2) && nstarted < nmembers / 2 + 1) {
                        r.logf("member add %s ⇒ 拒绝：加完后成员 %d、已启动 %d < 过半 %d（ErrNotEnoughStartedMembers）",
                                st.node(), nmembers, nstarted, nmembers / 2 + 1);
                        r.rejected.add(st.node());
                        continue;
                    }
                    c.members.add(st.node());
                    r.logf("member add %s ⇒ 一条配置变更日志，提交后成员变为 %s（集群 ID 不变）", st.node(), goList(c.members));
                    if (!c.leader())
                        r.logf("        ⇒ 现在过半要 %d 个、在线只有 %d 个：在 %s 启动之前集群停写（etcd 对 1→2 特许这种情况）",
                                c.need(), c.started(), st.node());
                }
                case "write" -> {
                    for (Cluster c : r.clusters) if (c.leader()) {
                        c.writes += st.n();
                        r.logf("客户端写入 %d 条 ⇒ %s 在任期 %d 提交并【确认】", st.n(), c.label(), c.term);
                        break;
                    }
                }
                default -> { }
            }
        }
        return r;
    }

    record Scen(String name, List<NodeCfg> nodes, List<Step> steps, String note) {}

    static void lab4() {
        head("3D-4", "冷启动沙盒", "初始配置只能从外部给——给错了会长出第二个集群，或者更糟");
        List<String> full = List.of("n1", "n2", "n3");
        List<Step> abc = List.of(Step.start("n1"), Step.start("n2"), Step.start("n3"));
        NodeCfg n1data = new NodeCfg("n1", full, "existing", "T1", true, false, 12);
        NodeCfg n2wiped = new NodeCfg("n2", full, "new", "T1", false, true, 0);
        NodeCfg n3wiped = new NodeCfg("n3", full, "new", "T1", false, true, 0);
        List<Scen> scens = List.of(
            new Scen("A · 正确的静态引导", List.of(
                NodeCfg.of("n1", full, "new", "T1"), NodeCfg.of("n2", full, "new", "T1"), NodeCfg.of("n3", full, "new", "T1")),
                abc, "期望结果：1 个集群、1 个 Leader、容错 1 台。"),
            new Scen("B · 一个节点的列表写错", List.of(
                NodeCfg.of("n1", full, "new", "T1"), NodeCfg.of("n2", full, "new", "T1"), NodeCfg.of("n3", List.of("n3"), "new", "T1")),
                abc, "最危险的一种：两个集群、两个 Leader，【都能写】，数据分叉且永远无法自动合并。\n"
                   + "      n3 的成员列表不同 ⇒ 集群 ID 不同 ⇒ 两边互不承认。从 n3 的视角看，它的集群完美无瑕。"),
            new Scen("C · token 不同", List.of(
                NodeCfg.of("n1", full, "new", "T1"), NodeCfg.of("n2", full, "new", "T1"), NodeCfg.of("n3", full, "new", "T2")),
                abc, "比 B 好，因为它失败得很响：token 是成员 ID 的一部分 ⇒ n3 眼里的集群 ID 和 n1/n2 的不同 ⇒ 被拒之门外；\n"
                   + "      它自己认为集群有 3 个成员，凑不齐过半 ⇒ 卡在启动中。n1/n2 能工作，但容错能力是 0。"),
            new Scen("D1 · 清空两台后用 state=new 重启（n1 在线）", List.of(n1data, n2wiped, n3wiped),
                abc, "同 token、同成员列表 ⇒ 集群 ID【完全相同】，n2/n3 不是在建「另一个集群」，而是在冒充已经存在的成员。\n"
                   + "      etcd 的 isMemberBootstrapped 检查拦住了它们：n1 记得它们发布过客户端地址 ⇒ 拒绝启动。\n"
                   + "      结果是停写，但数据还在。过半已经丢了，member remove 也提交不了，\n"
                   + "      正确的恢复路径是在 n1 上用 --force-new-cluster 重建单成员集群，再逐个 member add。"),
            new Scen("D2 · 同上，但 n1 当时不在线", List.of(n1data, n2wiped, n3wiped),
                List.of(Step.start("n2"), Step.start("n3"), Step.write(5), Step.start("n1")),
                "最阴险的一种。检查问不到任何知情者就放行了（源码里联系不上就返回 false）。\n"
                   + "      n2/n3 用同一个集群 ID 引导，任期从 0 重来，在任期 2 里确认了 5 条写。n1 回来时任期 12：\n"
                   + "      现任 Leader 收到它带着任期 12 的回复就下台，n1 的日志更新 ⇒ 当选 ⇒ 那 5 条写被覆盖。\n"
                   + "      根因回到 3-C：清空 data dir = 同时丢掉 currentTerm、votedFor、log。任期倒退，\n"
                   + "      才会出现「任期 2 的 Leader」这种本不该存在的东西，它提交的内容不受 Raft 保护。\n"
                   + "      如果 n1 永远不回来：集群健康、能读能写，半年的数据没了。"),
            new Scen("E · 单节点引导 + member add", List.of(
                NodeCfg.of("n1", List.of("n1"), "new", "T1"), NodeCfg.of("n2", List.of("n1", "n2"), "existing", "T1"),
                NodeCfg.of("n3", full, "existing", "T1")),
                List.of(Step.start("n1"), Step.add("n2"), Step.start("n2"), Step.add("n3"), Step.start("n3")),
                "同样正确，而且不需要提前知道最终拓扑。每个 existing 节点的 --initial-cluster\n"
                   + "      = 当时的成员 ∪ 自己（正是 member add 打印出来的内容），它不组建任何东西，只是加入。\n"
                   + "      n1 的 --initial-cluster 一直是「只有自己」，这没关系：集群跑起来后成员信息在 data dir 里。"),
            new Scen("F · E 的最后一步，n3 的列表漏了 n2", List.of(
                NodeCfg.of("n1", List.of("n1"), "new", "T1"), NodeCfg.of("n2", List.of("n1", "n2"), "existing", "T1"),
                NodeCfg.of("n3", List.of("n1", "n3"), "existing", "T1"),
                NodeCfg.of("n4", List.of("n1", "n2", "n3", "n4"), "existing", "T1")),
                List.of(Step.start("n1"), Step.add("n2"), Step.start("n2"), Step.add("n3"), Step.start("n3"), Step.add("n4")),
                "n3 的参数和集群当前成员对不上 ⇒ 被拒绝启动。但 member add n3 已经提交了：\n"
                   + "      集群现在是 3 个成员、2 个在线——还能服务，容错能力却是 0。\n"
                   + "      这时想再加 n4，etcd 会拒绝（加完后已启动的成员凑不齐过半）——这就是「一次只加一个、\n"
                   + "      确认正常了再加下一个」的原因，也是先用 --learner 加入、追平后再 promote 的原因。")
        );

        int[] w = {46, 8, 8, 14, 10, 30};
        String al = "LRRRRL";
        tableHead(new String[]{"场景", "集群数", "Leader", "在线/成员", "被拒启动", "  结果"}, w, al);
        List<BootResult> results = new ArrayList<>();
        for (Scen sc : scens) {
            BootResult r = bootstrap(sc.nodes(), sc.steps());
            results.add(r);
            int leaders = 0; List<String> online = new ArrayList<>(); boolean degraded = false;
            for (Cluster c : r.clusters) {
                if (c.leader()) leaders++;
                online.add(c.started() + "/" + c.members.size());
                if (c.leader() && c.started() < c.members.size()) degraded = true;
            }
            String v = "正常";
            if (leaders > 1) v = "脑裂：两个集群都能写";
            else if (r.lost > 0) v = r.lost + " 条已确认的写被覆盖";
            else if (leaders == 0) v = "停写（但数据还在）";
            else if (!r.rejected.isEmpty() || degraded) v = "降级：容错能力归零";
            String refused = r.refused.isEmpty() ? "—" : String.join(",", r.refused);
            tableRow(new String[]{sc.name(), "" + r.clusters.size(), "" + leaders,
                    String.join(" + ", online), refused, "  " + v}, w, al);
        }
        System.out.println();
        for (int i = 0; i < scens.size(); i++) {
            Scen sc = scens.get(i);
            System.out.printf("    %s%n", sc.name());
            for (String line : results.get(i).log) System.out.printf("      %s%n", line);
            System.out.printf("      ── %s%n%n", sc.note().replace("\n      ", "\n         "));
        }
        System.out.print(blk("""
    结论
      · 「有几台机器」和「怎么访问到对方」全部来自那份 --initial-cluster ——
        Raft 自己不做服务发现，DNS SRV / K8s Service 只是帮你【生成】这份列表。
      · 集群 ID 是由 token 和成员列表【确定性】算出来的：同 token 同列表 ⇒ 同一个集群。
        所以清空 data dir 后用 state=new 重启，不是「建了个新集群」，而是「冒充旧成员」。
      · 只有 state=new 的节点会按自己的参数引导集群；existing 节点只加入、从不组建，
        它的 --initial-cluster 必须和集群当前成员一致，否则拒绝启动。
      · --initial-* 只在 data dir 为空的首次启动生效。集群跑起来后成员信息就在本地状态里了，
        【改配置文件不会改变集群成员】——只能走 member add / remove / update。
        成员数据丢了的节点，正确做法是 member remove 再 member add 一个新成员（新 ID 带时间戳，不会撞）。""", 4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.print(blk("""
  Part 3 到此真正结束

  3-A 选谁当 Leader，3-B Leader 怎么把日志复制对，
  3-C 怎么活到生产，3-D 算法与系统之间的接缝。

  贯穿四章的还是那一条：【两个过半集合必然相交】。
  它在 Raft 里一共出现了四次——选举安全性、Leader 完整性、
  quorum 读写、成员变更——你已经全部见过了。

  下一站 Part 4 · 事务。先补单机的功课（ACID、隔离级别、
  四类并发异常、MVCC 与快照隔离），那正是 3D 里说的「第二层版本」。
  然后才进分布式：2PC/XA、TCC、Saga、事务消息、Percolator、Calvin、Spanner。""", 2));
        System.out.println("═".repeat(78) + "\n");
    }

    // ─── 输出工具 ─────────────────────────────────────────────────────────
    /** 缩进一段文本；空行不补空格（否则会留下看不见的行尾空白）。 */
    static String blk(String s, int k) { return s.indent(k).replaceAll("(?m)^[ ]+$", ""); }
    static String pct(int a, int b) { return b == 0 ? "—" : String.format("%.0f%%", 100.0 * a / b); }
    static String vals(int[] rs) {
        String[] names = {"A", "B", "C", "D"};
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < rs.length; i++) {
            if (i > 0) sb.append(' ');
            sb.append(rs[i] < 0 ? "∅" : names[rs[i] % 4]);
        }
        return sb.toString();
    }
    static int dispw(String s) {
        int w = 0;
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            w += (c >= 0x1100 && (c <= 0x115F || (c >= 0x2E80 && c <= 0xA4CF && c != 0x303F)
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
    static void tableHead(String[] cols, int[] w, String al) {
        System.out.println(layout(cols, w, al));
        int t = 0; for (int x : w) t += x;
        System.out.println("    " + "─".repeat(t));
    }
    static void tableRow(String[] c, int[] w, String al) { System.out.println(layout(c, w, al)); }

    static void parseArgs(String[] a) {
        for (int i = 0; i < a.length; i++) {
            String k = a[i].replaceFirst("^-+", "");
            if (i + 1 >= a.length) break;
            String v = a[++i];
            switch (k) {
                case "applylag" -> APPLYLAG = Integer.parseInt(v);
                case "reads" -> READS = Integer.parseInt(v);
                case "seeds" -> SEEDS = Integer.parseInt(v);
                case "writers" -> WRITERS = Integer.parseInt(v);
                case "n" -> N = Integer.parseInt(v);
                case "w" -> W = Integer.parseInt(v);
                case "r" -> R = Integer.parseInt(v);
                case "loss" -> LOSS = Double.parseDouble(v);
                case "seed" -> SEED = Long.parseLong(v);
                case "entries" -> ENTRIES = Integer.parseInt(v);
                default -> { }
            }
        }
    }
}
