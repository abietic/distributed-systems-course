/*
 * Lab 2 · 复制与一致性模型（Java 版）
 *
 * 运行：  cd java/lab02 && java Lab02.java
 * 调参：  java Lab02.java -n 5 -w 3 -r 3 -ops 500 -lag 20
 *
 * 需要 JDK 17+（单文件源码启动模式）。
 * 配套课件：courseware/ch02-replication.html
 */
import java.util.*;

public class Lab02 {

    static int N = 5, W = 3, R = 3, OPS = 400, LAG = 3;
    static long SEED = 42;
    static final String INIT = "⊥";

    // ══════════════════════════════════════════════════════════════════════
    // Quorum KV：N 个副本，W/R 可配，带复制延迟
    //   核心命题：W + R > N ⟺ 读写集合必有交集 ⟹ 读一定能看到最新写
    // ══════════════════════════════════════════════════════════════════════
    record Versioned(int ver, String val) {}

    static class QuorumKV {
        final int n, w, r, lag;
        final Random rnd;
        final Versioned[] rep;
        final List<int[]> queue = new ArrayList<>();   // {node, ver, deliverAt}
        final Map<Integer,String> vals = new HashMap<>();
        int clock = 0, opCount = 0, staleReads = 0, totalReads = 0;

        QuorumKV(int n, int w, int r, int lag, long seed) {
            this.n = n; this.w = w; this.r = r; this.lag = lag;
            this.rnd = new Random(seed);
            rep = new Versioned[n];
            Arrays.fill(rep, new Versioned(0, INIT));
        }

        int[] sample(int k) {
            List<Integer> p = new ArrayList<>();
            for (int i = 0; i < n; i++) p.add(i);
            Collections.shuffle(p, rnd);
            int[] out = new int[k];
            for (int i = 0; i < k; i++) out[i] = p.get(i);
            return out;
        }

        void tick() {                                   // 推进时间，投递到期的复制消息
            opCount++;
            queue.removeIf(q -> {
                if (q[2] <= opCount) {
                    if (q[1] > rep[q[0]].ver()) rep[q[0]] = new Versioned(q[1], vals.get(q[1]));
                    return true;
                }
                return false;
            });
        }

        void write(String val) {
            tick();
            clock++;
            vals.put(clock, val);
            Versioned v = new Versioned(clock, val);
            Set<Integer> inW = new HashSet<>();
            for (int i : sample(w)) { inW.add(i); rep[i] = v; }
            for (int i = 0; i < n; i++)
                if (!inW.contains(i)) queue.add(new int[]{i, clock, opCount + lag});
        }

        Versioned read() {
            tick();
            Versioned best = new Versioned(0, INIT);
            for (int i : sample(r)) if (rep[i].ver() > best.ver()) best = rep[i];
            totalReads++;
            if (best.ver() < clock) staleReads++;       // 上帝视角：最新已提交版本是 clock
            return best;
        }

        boolean guaranteed() { return w + r > n; }
    }

    // ══════════════════════════════════════════════════════════════════════
    // CRDT
    // ══════════════════════════════════════════════════════════════════════
    static int[] gInc(int[] g, int node) { g[node]++; return g; }

    /** 逐位取最大 —— 交换律、结合律、幂等律三律全满足。 */
    static void gMerge(int[] a, int[] b) { for (int i = 0; i < a.length; i++) a[i] = Math.max(a[i], b[i]); }

    static int gValue(int[] g) { int s = 0; for (int x : g) s += x; return s; }

    static class LWW {
        int val = 0; long ts = 0;
        void set(int v, long t) { val = v; ts = t; }
        void merge(LWW o) { if (o.ts > ts) { val = o.val; ts = o.ts; } }
    }

    /** OR-Set：每个元素带唯一 tag，删除只能删掉自己已观察到的 tag ⇒ add-wins。 */
    static class ORSet {
        final Map<String, Set<Long>> adds = new HashMap<>(), rems = new HashMap<>();
        void add(String e, long tag) { adds.computeIfAbsent(e, k -> new HashSet<>()).add(tag); }
        void remove(String e) {
            rems.computeIfAbsent(e, k -> new HashSet<>()).addAll(adds.getOrDefault(e, Set.of()));
        }
        boolean has(String e) {
            for (long t : adds.getOrDefault(e, Set.of()))
                if (!rems.getOrDefault(e, Set.of()).contains(t)) return true;
            return false;
        }
        void merge(ORSet o) {
            o.adds.forEach((e, ts) -> adds.computeIfAbsent(e, k -> new HashSet<>()).addAll(ts));
            o.rems.forEach((e, ts) -> rems.computeIfAbsent(e, k -> new HashSet<>()).addAll(ts));
        }
        int size() { int n = 0; for (String e : adds.keySet()) if (has(e)) n++; return n; }
    }

    // ══════════════════════════════════════════════════════════════════════
    // 一致性判定器：穷举所有串行化顺序（Jepsen 的极简版）
    // ══════════════════════════════════════════════════════════════════════
    record Op(int c, String kind, String v, int s, int e) {}
    record Hist(String name, String desc, List<Op> ops) {}

    interface Placeable { boolean ok(int j, Set<Integer> placed); }

    static List<Integer> dfsOrder(List<Op> ops, List<Integer> idx, Placeable canPlace) {
        Set<Integer> placed = new HashSet<>();
        List<Integer> out = new ArrayList<>();
        int[] steps = {0};
        if (go(ops, idx, canPlace, placed, out, INIT, steps)) return new ArrayList<>(out);
        return null;
    }

    static boolean go(List<Op> ops, List<Integer> idx, Placeable canPlace,
                      Set<Integer> placed, List<Integer> out, String cur, int[] steps) {
        if (++steps[0] > 500_000) return false;
        if (out.size() == idx.size()) return true;
        for (int j : idx) {
            if (placed.contains(j) || !canPlace.ok(j, placed)) continue;
            Op o = ops.get(j);
            if (o.kind().equals("r") && !o.v().equals(cur)) continue;   // 读必须返回当前值
            placed.add(j); out.add(j);
            if (go(ops, idx, canPlace, placed, out, o.kind().equals("w") ? o.v() : cur, steps)) return true;
            out.remove(out.size() - 1); placed.remove(j);
        }
        return false;
    }

    static List<Integer> allIdx(List<Op> ops) {
        List<Integer> l = new ArrayList<>();
        for (int i = 0; i < ops.size(); i++) l.add(i);
        return l;
    }

    /** 线性一致：必须尊重真实时间序。 */
    static List<Integer> checkLin(List<Op> ops) {
        return dfsOrder(ops, allIdx(ops), (j, placed) -> {
            for (int i = 0; i < ops.size(); i++)
                if (!placed.contains(i) && i != j && ops.get(i).e() < ops.get(j).s()) return false;
            return true;
        });
    }

    /** 顺序一致：去掉真实时间，只保留每个客户端的程序序。差别就这一处。 */
    static List<Integer> checkSeq(List<Op> ops) {
        return dfsOrder(ops, allIdx(ops), (j, placed) -> {
            for (int i = 0; i < ops.size(); i++)
                if (!placed.contains(i) && i != j
                        && ops.get(i).c() == ops.get(j).c() && ops.get(i).s() < ops.get(j).s()) return false;
            return true;
        });
    }

    /** 因果序传递闭包：程序序 + 写→读依赖。 */
    static boolean[][] causalReach(List<Op> ops) {
        int n = ops.size();
        List<List<Integer>> adj = new ArrayList<>();
        for (int i = 0; i < n; i++) adj.add(new ArrayList<>());
        for (int i = 0; i < n; i++) for (int j = 0; j < n; j++) {
            if (i == j) continue;
            Op a = ops.get(i), b = ops.get(j);
            if (a.c() == b.c() && a.s() < b.s()) adj.get(i).add(j);
            if (a.kind().equals("w") && b.kind().equals("r") && b.v().equals(a.v())) adj.get(i).add(j);
        }
        boolean[][] Rm = new boolean[n][n];
        for (int s = 0; s < n; s++) {
            Deque<Integer> st = new ArrayDeque<>(List.of(s));
            Set<Integer> seen = new HashSet<>(List.of(s));
            while (!st.isEmpty()) {
                int x = st.pop();
                for (int y : adj.get(x)) if (seen.add(y)) { Rm[s][y] = true; st.push(y); }
            }
        }
        return Rm;
    }

    /** 因果一致：允许每个客户端有各自的串行化，但都要尊重因果序。 */
    static Map<Integer,List<Integer>> checkCausal(List<Op> ops) {
        boolean[][] Rm = causalReach(ops);
        Map<Integer,List<Integer>> out = new TreeMap<>();
        for (int c : clients(ops)) {
            List<Integer> idx = new ArrayList<>();
            for (int i = 0; i < ops.size(); i++)
                if (ops.get(i).kind().equals("w") || ops.get(i).c() == c) idx.add(i);
            List<Integer> res = dfsOrder(ops, idx, (j, placed) -> {
                for (int i : idx) if (!placed.contains(i) && i != j && Rm[i][j]) return false;
                return true;
            });
            if (res == null) return null;
            out.put(c, res);
        }
        return out;
    }

    static List<Integer> clients(List<Op> ops) {
        return ops.stream().map(Op::c).distinct().sorted().toList();
    }

    /** 版本序：写按真实开始时间排序，初始值版本 0。 */
    static Map<String,Integer> versions(List<Op> ops) {
        Map<String,Integer> m = new HashMap<>(Map.of(INIT, 0));
        List<Op> ws = new ArrayList<>(ops.stream().filter(o -> o.kind().equals("w")).toList());
        ws.sort(Comparator.comparingInt(Op::s));
        for (int i = 0; i < ws.size(); i++) m.put(ws.get(i).v(), i + 1);
        return m;
    }

    static List<String> checkRYW(List<Op> ops) {
        Map<String,Integer> V = versions(ops);
        List<String> bad = new ArrayList<>();
        for (int c : clients(ops)) {
            int lastW = -1;
            for (Op o : clientOps(ops, c)) {
                if (o.kind().equals("w")) lastW = V.get(o.v());
                else if (lastW >= 0 && V.get(o.v()) < lastW)
                    bad.add("C" + (c + 1) + " 写入后读到了更旧的 " + o.v());
            }
        }
        return bad;
    }

    static List<String> checkMono(List<Op> ops) {
        Map<String,Integer> V = versions(ops);
        List<String> bad = new ArrayList<>();
        for (int c : clients(ops)) {
            List<Op> rs = clientOps(ops, c).stream().filter(o -> o.kind().equals("r")).toList();
            for (int i = 1; i < rs.size(); i++)
                if (V.get(rs.get(i).v()) < V.get(rs.get(i - 1).v()))
                    bad.add("C" + (c + 1) + " 先读到 " + rs.get(i - 1).v() + " 又读回了更旧的 " + rs.get(i).v());
        }
        return bad;
    }

    static List<Op> clientOps(List<Op> ops, int c) {
        List<Op> r = new ArrayList<>(ops.stream().filter(o -> o.c() == c).toList());
        r.sort(Comparator.comparingInt(Op::s));
        return r;
    }

    /** 与课件「实验 2」完全相同的六个场景。 */
    static final List<Hist> HISTORIES = List.of(
        new Hist("① 教科书式的线性一致", "C1 写完并返回后 C2 才开始读，必须读到新值",
            List.of(new Op(0,"w","A",0,3), new Op(1,"r","A",5,8))),
        new Hist("② 重叠区间里读到旧值", "读与写在时间上重叠，线性一致允许读到旧值",
            List.of(new Op(0,"w","A",2,8), new Op(1,"r",INIT,3,6), new Op(2,"r","A",10,12))),
        new Hist("③ 陈旧读", "写早已返回，之后开始的读却看到初始值",
            List.of(new Op(0,"w","A",0,3), new Op(1,"r",INIT,5,8))),
        new Hist("④ 两个观察者，相反的顺序", "两个并发写，C3 看到 A→B，C4 看到 B→A",
            List.of(new Op(0,"w","A",0,2), new Op(1,"w","B",1,3),
                    new Op(2,"r","A",5,6), new Op(2,"r","B",8,9),
                    new Op(3,"r","B",5,6), new Op(3,"r","A",8,9))),
        new Hist("⑤ 因果倒置", "C2 读到 A 后才写 B（A 因果先于 B），C3 却先看到 B 又看回 A",
            List.of(new Op(0,"w","A",0,2), new Op(1,"r","A",3,4), new Op(1,"w","B",5,7),
                    new Op(2,"r","B",9,10), new Op(2,"r","A",12,13))),
        new Hist("⑥ 读不到自己的写", "C1 自己写完紧接着自己读，却读到初始值",
            List.of(new Op(0,"w","A",0,2), new Op(0,"r",INIT,4,6), new Op(1,"r","A",8,10)))
    );

    // ══════════════════════════════════════════════════════════════════════
    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：N=%d ｜ W=%d ｜ R=%d ｜ 操作数 %d ｜ 复制延迟 %d 次操作 ｜ 种子 %d%n",
                N, W, R, OPS, LAG, SEED);
        lab21(); lab22(); lab23(); lab24(); epilogue();
    }

    // ─── Lab 2-1 ──────────────────────────────────────────────────────────
    static void lab21() {
        head(1, "Quorum：W + R > N", "同一份工作负载，换不同的 W/R，数陈旧读");
        int[] w = {22, 7, 7, 7, 13, 16}; String al = "LRRRRR";
        tableHead(new String[]{"配置","N","W","R","W+R>N","陈旧读比例"}, w, al);

        java.util.function.BiFunction<String,int[],Double> run = (label, cfg) -> {
            QuorumKV q = new QuorumKV(cfg[0], cfg[1], cfg[2], LAG, SEED);
            for (int i = 0; i < OPS; i++) { q.write("v" + i); q.read(); }
            double ratio = q.staleReads * 100.0 / q.totalReads;
            tableRow(new String[]{label, "" + cfg[0], "" + cfg[1], "" + cfg[2],
                    q.guaranteed() ? "是" : "否", String.format("%.1f%%", ratio)}, w, al);
            return ratio;
        };
        int maj = N / 2 + 1;
        double strict = run.apply("多数派 W=R=⌊N/2⌋+1", new int[]{N, maj, maj});
        run.apply("W=N, R=1", new int[]{N, N, 1});
        run.apply("W=1, R=N", new int[]{N, 1, N});
        double loose = run.apply("W=R=⌊N/2⌋（故意不足）", new int[]{N, Math.max(1, N / 2), Math.max(1, N / 2)});
        run.apply("W=1, R=1（最快最弱）", new int[]{N, 1, 1});
        System.out.println();
        row("W+R>N 的三行陈旧读", String.format("%.1f%%　← 应当恒为 0", strict));
        row("W+R≤N 的陈旧读", String.format("%.1f%%　← 交集可能为空，读不到最新", loose));
        System.out.println("""

    ▸ 前三行的 W+R 都 > N，陈旧读恒为 0 —— 交集性质是一条数学保证，不是概率。
    ▸ 后两行故意让 W+R ≤ N，陈旧读立刻出现。它未必是错误配置：
      Cassandra 的 ONE/ONE 就是这么设的，用一致性换最低延迟和最高可用性。
    ▸ 试试 -lag 20，看后两行的比例怎么变、前三行会不会变。""".indent(4));
    }

    // ─── Lab 2-2 ──────────────────────────────────────────────────────────
    static void lab22() {
        head(2, "CAP：把取舍变成两个数字",
                String.format("%d 个副本被切成 %d|%d 两侧，同一份写入负载跑两种策略", N, (N + 1) / 2, N / 2));
        int majSize = (N + 1) / 2, quorum = N / 2 + 1;

        int[][] out = new int[2][3];        // [cp/ap][ok, fail, conflict]
        for (int mode = 0; mode < 2; mode++) {
            boolean cp = mode == 0;
            Random rnd = new Random(SEED);
            Set<Integer> majW = new HashSet<>(), minW = new HashSet<>();
            int ok = 0, fail = 0;
            for (int i = 0; i < OPS; i++) {
                int key = rnd.nextInt(Math.max(1, OPS / 4));
                boolean toMajority = rnd.nextInt(N) < majSize;
                if (cp) {
                    if (toMajority && majSize >= quorum) { ok++; majW.add(key); }
                    else fail++;
                } else {
                    ok++;
                    if (toMajority) majW.add(key); else minW.add(key);
                }
            }
            int conflict = 0;
            for (int k : majW) if (minW.contains(k)) conflict++;
            out[mode] = new int[]{ok, fail, conflict};
        }
        int[] w = {16, 14, 14, 16, 18}; String al = "LRRRR";
        tableHead(new String[]{"策略","写入成功","写入失败","成功率","恢复后的冲突 key"}, w, al);
        tableRow(new String[]{"CP（少数派拒绝）", "" + out[0][0], "" + out[0][1],
                String.format("%.1f%%", out[0][0] * 100.0 / OPS), out[0][2] + "（不可能有）"}, w, al);
        tableRow(new String[]{"AP（两侧都收）", "" + out[1][0], "" + out[1][1],
                String.format("%.1f%%", out[1][0] * 100.0 / OPS), out[1][2] + " 个待合并"}, w, al);
        System.out.println("""

    ▸ 这就是 CAP 的全部内容：分区期间，你要么损失一部分写入的可用性（CP），
      要么接下一堆需要合并的冲突（AP）。没有第三个选项，因为 P 不可选。
    ▸ 注意 CP 那一行的失败率 ≈ 少数派副本占比 —— 少数派侧的请求全部被拒。
      这不是 bug，是设计：宁可停下，不可出错。
    ▸ AP 那一行的冲突数就是你欠下的债，分区恢复后必须用 LWW / siblings / CRDT 还上。""".indent(4));
    }

    // ─── Lab 2-3 ──────────────────────────────────────────────────────────
    static void lab23() {
        head(3, "一致性判定器", "穷举所有串行化顺序，判定每条历史属于谱系的哪一档");
        int[] w = {26, 10, 10, 10, 10, 10}; String al = "LRRRRR";
        tableHead(new String[]{"执行历史","线性一致","顺序一致","因果一致","读己之写","单调读"}, w, al);

        for (Hist h : HISTORIES) {
            boolean lin = checkLin(h.ops()) != null;
            boolean seq = checkSeq(h.ops()) != null;
            boolean cau = checkCausal(h.ops()) != null;
            boolean ryw = checkRYW(h.ops()).isEmpty();
            boolean mono = checkMono(h.ops()).isEmpty();

            // 谱系必须嵌套：线性 ⟹ 顺序 ⟹ 因果。不成立说明判定器有 bug。
            if (lin && !seq) throw new AssertionError("判定器自相矛盾：线性却不顺序 — " + h.name());
            if (seq && !cau) throw new AssertionError("判定器自相矛盾：顺序却不因果 — " + h.name());

            tableRow(new String[]{h.name(), yn(lin), yn(seq), yn(cau), yn(ryw), yn(mono)}, w, al);
        }
        System.out.println();
        Hist h = HISTORIES.get(3);
        row("细看 " + h.name(), h.desc());
        Map<Integer,List<Integer>> cau = checkCausal(h.ops());
        if (cau != null) cau.forEach((c, order) -> {
            StringBuilder sb = new StringBuilder();
            for (int j : order) {
                Op o = h.ops().get(j);
                sb.append(sb.isEmpty() ? "" : " → ").append(o.kind()).append("(").append(o.v()).append(")@C").append(o.c() + 1);
            }
            row("  C" + (c + 1) + " 眼中的顺序", sb);
        });
        System.out.println("""

    ▸ 判定器内置了自检：谱系必须满足 线性 ⟹ 顺序 ⟹ 因果，违反直接抛 AssertionError。
    ▸ ④ 那一行是最值得盯的：不存在任何单一全序能同时解释 C3 和 C4，
      但允许每人有自己的顺序之后就都说得通了 —— 因为那两个写是并发的。
    ▸ ⑥ 违反读己之写，同时也违反了因果一致 —— 因果一致蕴含读己之写。""".indent(4));
    }

    // ─── Lab 2-4 ──────────────────────────────────────────────────────────
    static void lab24() {
        head(4, "CRDT：收敛不等于正确", "随机分区、随机合并顺序，跑一千遍");
        final int REP = 3;
        Random rnd = new Random(SEED);
        int trials = 0, gExact = 0, lwwLost = 0, totalClicks = 0, diverged = 0;

        for (int t = 0; t < 1000; t++) {
            int[][] gs = new int[REP][REP];
            LWW[] lw = new LWW[REP];
            for (int i = 0; i < REP; i++) lw[i] = new LWW();
            long ts = 0; int total = 0;
            for (int i = 0; i < REP; i++) {
                int k = 1 + rnd.nextInt(5);
                for (int j = 0; j < k; j++) {
                    ts++;
                    gInc(gs[i], i);                    // G-Counter：只增自己那一槽
                    lw[i].set(lw[i].val + 1, ts);      // LWW：本地读改写
                    total++;
                }
            }
            for (int round = 0; round < 6; round++) {  // 随机顺序两两合并
                int a = rnd.nextInt(REP), b = rnd.nextInt(REP);
                if (a != b) { gMerge(gs[a], gs[b]); lw[a].merge(lw[b]); }
            }
            for (int i = 0; i < REP; i++) for (int j = 0; j < REP; j++)
                if (i != j) { gMerge(gs[i], gs[j]); lw[i].merge(lw[j]); }

            trials++; totalClicks += total;
            if (gValue(gs[0]) == total) gExact++; else diverged++;
            lwwLost += total - lw[0].val;
        }

        int[] w = {28, 14, 20, 24}; String al = "LRRR";
        tableHead(new String[]{"数据类型","实验次数","结果精确","累计丢失"}, w, al);
        tableRow(new String[]{"G-Counter（逐位取最大）", "" + trials,
                String.format("%d 次（%.0f%%）", gExact, gExact * 100.0 / trials), "0"}, w, al);
        tableRow(new String[]{"LWW-Register（取时间戳大）", "" + trials, "0 次",
                lwwLost + " / " + totalClicks + " 次点击"}, w, al);
        System.out.println();
        row("G-Counter 与真值不符", diverged + " 次　← 应当恒为 0");
        row("LWW 丢失率", String.format("%.1f%%", lwwLost * 100.0 / totalClicks));

        // 三律验证
        int[] a = new int[3], b = new int[3], c = new int[3];
        gInc(a, 0); gInc(a, 0); gInc(b, 1); gInc(c, 2); gInc(c, 2); gInc(c, 2);
        int[] m1 = a.clone(); gMerge(m1, b); gMerge(m1, c);      // (a∪b)∪c
        int[] m2 = c.clone(); gMerge(m2, a); gMerge(m2, b);      // (c∪a)∪b
        int[] m3 = m1.clone(); gMerge(m3, m1); gMerge(m3, b); gMerge(m3, b);
        System.out.println();
        row("交换律 + 结合律", Arrays.equals(m1, m2) ? "✓ 两种合并顺序结果相同" : "✗");
        row("幂等律", Arrays.equals(m1, m3) ? "✓ 重复合并结果不变" : "✗");

        ORSet s1 = new ORSet(), s2 = new ORSet();
        s1.add("牛奶", 1); s2.merge(s1); s2.remove("牛奶"); s1.add("牛奶", 2);
        s1.merge(s2); s2.merge(s1);
        row("OR-Set 并发加删",
                String.format("两副本收敛=%s，元素存在=%s（add-wins）", s1.size() == s2.size(), s1.has("牛奶")));
        System.out.println("""

    ▸ G-Counter 在 1000 次随机分区 + 随机合并顺序下，结果 100% 精确 ——
      因为逐位取最大满足交换律、结合律、幂等律，合并顺序根本不影响结果。
    ▸ LWW 每次也都"收敛"了（三个副本值相同），但值是错的。
      收敛是关于「最终一致」的性质，不丢数据是关于「语义」的性质，两者正交。
    ▸ OR-Set 的 add-wins：删除只能删掉自己已经观察到的 tag，删不掉没见过的新 tag。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  三个思考题（答案在课件 §2.5 / §2.6）：

    1. 把 -w 2 -r 2 -n 5（W+R=4 ≤ 5）跑一遍，陈旧读比例是多少？
       再把 -lag 从 3 调到 30，这个比例怎么变？为什么 W+R>N 那几行不受影响？
    2. Lab 2-2 里，如果把分区改成 4|1 而不是 3|2，CP 的成功率会怎么变？
    3. LWW-Register 是一个合法的 CRDT（满足三律、必然收敛），但它丢数据。
       那"CRDT"这个保证到底值多少钱？

  下一站：Part 3 · 共识算法 —— 从 FLP 到 Raft，Lab 3 用 Go 从零写一个能跑的 Raft。""");
        System.out.println("═".repeat(78) + "\n");
    }

    // ─── 输出工具 ─────────────────────────────────────────────────────────
    static String yn(boolean b) { return b ? "✓" : "✗"; }

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
    static void head(int n, String title, String sub) {
        System.out.println("\n" + "═".repeat(78));
        System.out.printf("  Lab 2-%d · %s%n  %s%n", n, title, sub);
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
        for (int i = 0; i + 1 < a.length; i += 2) {
            String k = a[i].replaceFirst("^-+", ""), v = a[i + 1];
            switch (k) {
                case "n"    -> N = Integer.parseInt(v);
                case "w"    -> W = Integer.parseInt(v);
                case "r"    -> R = Integer.parseInt(v);
                case "ops"  -> OPS = Integer.parseInt(v);
                case "lag"  -> LAG = Integer.parseInt(v);
                case "seed" -> SEED = Long.parseLong(v);
                default     -> System.out.println("未知参数: " + k);
            }
        }
    }
}
