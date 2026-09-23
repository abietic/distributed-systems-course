/*
 * Lab 1 · 时间、顺序与因果（Java 版）
 *
 * 运行：  cd java/lab01 && java Lab01.java
 * 调参：  java Lab01.java -skew 200 -gap 50 -n 500
 *
 * 需要 JDK 17+（单文件源码启动模式，不需要 Maven / Gradle）。
 * 配套课件：courseware/ch01-time-order.html
 */
import java.util.*;

public class Lab01 {

    // ─── 可调参数 ──────────────────────────────────────────────────────────
    static int  N    = 200;  // Lab 1-3 的并发写轮数
    static long SKEW = 80;   // 副本 R1 的时钟偏移（毫秒）
    static long GAP  = 60;   // 两次写入的真实间隔（毫秒）
    static long JIT  = 60;   // 时钟偏移的随机抖动（毫秒）
    static long SEED = 42;

    static final int NP = 3;
    static final String[] PN = {"P1", "P2", "P3"};

    // ══════════════════════════════════════════════════════════════════════
    // Lamport 逻辑时钟
    //   保证：  a → b  ⟹  L(a) < L(b)
    //   不保证：L(a) < L(b) ⇒ a → b     ← 这就是它的根本局限
    // ══════════════════════════════════════════════════════════════════════
    static class Lamport {
        long t = 0;
        long local()           { return ++t; }                          // 本地事件 / 发送
        long recv(long remote) { if (remote > t) t = remote; return ++t; }
    }

    // ══════════════════════════════════════════════════════════════════════
    // 向量时钟
    //   保证：a → b  ⟺  V(a) < V(b)     ← 充要条件，所以能精确检测并发
    //   代价：O(N) 空间
    // ══════════════════════════════════════════════════════════════════════
    enum Ord {
        EQUAL("相同"), BEFORE("a → b"), AFTER("b → a"), CONCURRENT("a ∥ b（并发）");
        final String cn; Ord(String cn) { this.cn = cn; }
    }

    static long[] vLocal(long[] v, int i) { long[] c = v.clone(); c[i]++; return c; }

    /** 收到消息：逐位取最大，再自增自己那一位。 */
    static long[] vMerge(long[] v, long[] o, int i) {
        long[] c = v.clone();
        for (int k = 0; k < c.length; k++) c[k] = Math.max(c[k], o[k]);
        c[i]++;
        return c;
    }

    static boolean le(long[] a, long[] b) {
        for (int k = 0; k < a.length; k++) if (a[k] > b[k]) return false;
        return true;
    }

    /** 向量时钟的全部精髓，只有五行。 */
    static Ord compare(long[] a, long[] b) {
        boolean ab = le(a, b), ba = le(b, a);
        if (ab && ba) return Ord.EQUAL;
        if (ab) return Ord.BEFORE;
        if (ba) return Ord.AFTER;
        return Ord.CONCURRENT;   // 各自都知道对方不知道的事 ⇒ 谁都没影响谁 ⇒ 冲突
    }

    static String vs(long[] v) {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < v.length; i++) sb.append(i > 0 ? "," : "").append(v[i]);
        return sb.append("]").toString();
    }

    // ══════════════════════════════════════════════════════════════════════
    // 混合逻辑时钟 HLC（Kulkarni et al., 2014）
    //   时间戳是二元组 (l, c)，按字典序比较。
    //   l 尽量贴着物理时钟走，c 只在物理时钟"不够用"时递增。
    // ══════════════════════════════════════════════════════════════════════
    static class HLC {
        long l = 0, c = 0;

        long[] local(long pt) {                 // 本地事件 / 发送
            long prev = l;
            l = Math.max(l, pt);
            if (l == prev) c++;                 // 物理时钟没前进（含回拨）→ 靠逻辑位撑住单调
            else c = 0;                         // 物理时钟前进了 → 逻辑位归零
            return new long[]{l, c};
        }

        long[] recv(long pt, long lm, long cm) {
            long prev = l;
            l = Math.max(Math.max(prev, lm), pt);
            if (l == prev && l == lm) c = Math.max(c, cm) + 1;
            else if (l == prev)       c++;
            else if (l == lm)         c = cm + 1;
            else                      c = 0;
            return new long[]{l, c};
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // 事件轨迹：和课件「实验 1 · 时空图编辑器」的「载入经典示例」完全一致
    // ══════════════════════════════════════════════════════════════════════
    enum Kind {
        LOCAL("本地事件"), SEND("发送消息"), RECV("接收消息");
        final String cn; Kind(String cn) { this.cn = cn; }
    }

    static class Ev {
        int id, p, t, peer; Kind kind = Kind.LOCAL;
        long L; long[] V;
        String name() { return PN[p] + "@t" + t; }
    }

    static final List<Ev> trace = new ArrayList<>();
    static final Map<Integer, Ev> byId = new HashMap<>();

    static Ev local(int p, int t) {
        Ev e = new Ev(); e.id = trace.size() + 1; e.p = p; e.t = t;
        trace.add(e); byId.put(e.id, e); return e;
    }

    static void msg(int sp, int st, int rp, int rt) {
        Ev s = local(sp, st), r = local(rp, rt);
        s.kind = Kind.SEND; s.peer = r.id;
        r.kind = Kind.RECV; r.peer = s.id;
    }

    static void buildTrace() {
        trace.clear(); byId.clear();
        local(0, 1);           // P1 第一个事件
        local(1, 0);           // P2 第一个事件（与上面并发）
        local(2, 2);           // P3 第一个事件
        msg(0, 3, 1, 5);       // P1 → P2
        msg(2, 6, 0, 8);       // P3 → P1
        msg(1, 7, 2, 9);       // P2 → P3
        local(0, 11);
        local(2, 12);
        computeClocks();
    }

    /** 按 (t, p) 排序即为合法拓扑序：消息一定满足「接收列 > 发送列」。 */
    static void computeClocks() {
        List<Ev> order = new ArrayList<>(trace);
        order.sort(Comparator.<Ev>comparingInt(e -> e.t).thenComparingInt(e -> e.p));
        Lamport[] lam = new Lamport[NP];
        long[][] vec = new long[NP][NP];
        for (int i = 0; i < NP; i++) lam[i] = new Lamport();

        for (Ev e : order) {
            if (e.kind == Kind.RECV) {
                Ev s = byId.get(e.peer);
                e.L = lam[e.p].recv(s.L);
                vec[e.p] = vMerge(vec[e.p], s.V, e.p);
            } else {
                e.L = lam[e.p].local();
                vec[e.p] = vLocal(vec[e.p], e.p);
            }
            e.V = vec[e.p].clone();
        }
    }

    /** 用 happens-before 的三条规则做可达性搜索，独立校验向量时钟的判定。 */
    static boolean causalPath(Ev a, Ev b) { return dfs(a, b, new HashSet<>()); }

    static boolean dfs(Ev x, Ev b, Set<Integer> seen) {
        if (x.id == b.id) return true;
        if (!seen.add(x.id)) return false;
        for (Ev n : trace)                                      // 规则①：同进程内的后继事件
            if (n.p == x.p && n.t > x.t && dfs(n, b, seen)) return true;
        return x.kind == Kind.SEND && dfs(byId.get(x.peer), b, seen);  // 规则②：发送 → 接收
    }

    // ══════════════════════════════════════════════════════════════════════
    public static void main(String[] args) {
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：并发写 %d 轮 ｜ R1 时钟偏移 %+dms（抖动 ±%dms）｜ 写入真实间隔 %dms ｜ 种子 %d%n",
                N, SKEW, JIT, GAP, SEED);
        lab11(); lab12(); lab13(); lab14(); epilogue();
    }

    // ─── Lab 1-1 ──────────────────────────────────────────────────────────
    static void lab11() {
        head(1, "Lamport 时钟与向量时钟",
                "和课件「实验 1」的经典示例完全一致 —— 浏览器里看到的数字，这里能跑出来");
        buildTrace();
        int[] w = {12, 12, 12, 12, 16}; String al = "LRRRR";
        tableHead(new String[]{"事件", "类型", "对端", "Lamport", "向量时钟"}, w, al);
        for (Ev e : trace) {
            String peer = e.peer == 0 ? "—" : byId.get(e.peer).name();
            tableRow(new String[]{e.name(), e.kind.cn, peer, "" + e.L, vs(e.V)}, w, al);
        }
        System.out.println("""

    ▸ 注意 P1@t8 的向量是 [3,0,2] —— 第 2 位是 2，说明 P1 通过那条来自 P3 的消息
      "得知"了 P3 已经走了 2 步。向量时钟的每一位就是"我所知道的它走到哪了"。""".indent(4));
    }

    // ─── Lab 1-2 ──────────────────────────────────────────────────────────
    static void lab12() {
        head(2, "Lamport 的盲区", "枚举所有事件对，看两种时钟在哪些对上给出不同答案");
        buildTrace();
        int total = 0, causal = 0, conc = 0, misled = 0;
        List<String[]> examples = new ArrayList<>();

        for (int i = 0; i < trace.size(); i++)
            for (int j = i + 1; j < trace.size(); j++) {
                Ev a = trace.get(i), b = trace.get(j);
                total++;
                Ord ord = compare(a.V, b.V);

                // 独立校验：不依赖向量时钟，直接按 happens-before 的定义搜索
                Ord want = causalPath(a, b) ? Ord.BEFORE
                         : causalPath(b, a) ? Ord.AFTER : Ord.CONCURRENT;
                if (ord != want)
                    throw new AssertionError("向量时钟判定与 happens-before 定义不符: "
                            + a.name() + " vs " + b.name() + " → " + ord + " / " + want);

                if (ord == Ord.CONCURRENT) {
                    conc++;
                    if (a.L != b.L) {
                        misled++;
                        if (examples.size() < 6) {
                            Ev first  = a.L < b.L ? a : b;
                            Ev second = a.L < b.L ? b : a;
                            examples.add(new String[]{a.name() + " ∥ " + b.name(),
                                    "L=" + a.L, "L=" + b.L, "",
                                    "Lamport 误以为 " + first.name() + " 在 " + second.name() + " 之前"});
                        }
                    }
                } else causal++;
            }

        row("事件对总数", total);
        row("有因果关系（a→b 或 b→a）", causal + " 对　← 这些 Lamport 判断正确");
        row("并发（互不影响）", conc + " 对");
        row("其中被 Lamport 误导的", String.format("%d 对　← 占并发对的 %.0f%%", misled, misled * 100.0 / conc));
        System.out.println();

        int[] w = {24, 8, 8, 2, 46}; String al = "LRRLL";
        tableHead(new String[]{"并发的事件对", "L(a)", "L(b)", "", "Lamport 会得出的错误结论"}, w, al);
        for (String[] ex : examples) tableRow(ex, w, al);
        System.out.println("""

    ▸ 向量时钟的判定通过了独立校验：程序另外用 happens-before 的三条规则做了
      一次可达性搜索，两者结论完全一致（不一致会直接抛 AssertionError）。
    ▸ Lamport 在所有"有因果关系"的对上都判断正确 —— 它从不把因果判反。
      它只是在"并发"的对上会编造出一个不存在的顺序。这就是"只能证伪，不能证实"。""".indent(4));
    }

    // ─── Lab 1-3 ──────────────────────────────────────────────────────────
    static void lab13() {
        head(3, "last-write-wins 到底丢了多少",
                "两个副本各收到一次写，互不知情 —— 这是一次真正的并发写");
        Random rnd = new Random(SEED);
        int lwwDiscard = 0, lwwDiscardLater = 0, vvConflict = 0, vvDiscard = 0;

        for (int i = 0; i < N; i++) {
            long trueA = (long) i * 1000;
            long trueB = trueA + GAP;                       // B 在真实时间上确确实实晚于 A
            long jit = JIT > 0 ? rnd.nextInt((int) (2 * JIT)) - JIT : 0;
            long stampA = trueA + SKEW + jit;               // R1 用自己（歪的）时钟打戳
            long stampB = trueB;                            // R2 时钟准确

            lwwDiscard++;                                   // LWW 必须二选一 ⇒ 每轮必丢一条
            if (stampA > stampB) lwwDiscardLater++;         // 丢掉的是真实更晚的 B —— 肉眼可见的错误

            // 版本向量：A 写在 R1 上是 [1,0]，B 写在 R2 上是 [0,1]，互不 ≤ ⇒ 并发
            if (compare(new long[]{1, 0}, new long[]{0, 1}) == Ord.CONCURRENT) vvConflict++;
        }

        int[] w = {34, 12, 12}; String al = "LRR";
        tableHead(new String[]{"", "last-write-wins", "版本向量"}, w, al);
        tableRow(new String[]{"并发写轮数", "" + N, "" + N}, w, al);
        tableRow(new String[]{"检测出冲突", "0", "" + vvConflict}, w, al);
        tableRow(new String[]{"被丢弃的写入", "" + lwwDiscard, "" + vvDiscard}, w, al);
        tableRow(new String[]{"其中丢的是真实更晚那条", "" + lwwDiscardLater, "—"}, w, al);
        System.out.println();
        row("LWW 的\"明显错误率\"",
                String.format("%.1f%%（丢掉真实更晚写入的比例）", lwwDiscardLater * 100.0 / N));
        System.out.println("""

    ▸ 最容易被忽略的一行是「被丢弃的写入」：LWW 每一轮都丢掉一条，
      因为它必须二选一。把 -skew 设成 0，"明显错误率"会降到 0，
      但丢弃数依然是 100% —— 时钟准不准，改变的只是"丢哪一条"。
    ▸ 版本向量不假装知道答案：它把两条都留下来（siblings），
      把"该怎么合并"这个只有业务代码知道的问题交还给业务代码。
    ▸ 试试 -skew 0 和 -skew 300，看两个数字怎么变。""".indent(4));
    }

    // ─── Lab 1-4 ──────────────────────────────────────────────────────────
    static void lab14() {
        head(4, "混合逻辑时钟 HLC", "注入一次 NTP 回拨，看三种时钟谁活了下来");
        Random rnd = new Random(SEED + 1);
        long pt = 12000;
        Lamport lam = new Lamport();
        HLC hlc = new HLC();

        int[] w = {12, 12, 11, 14, 16, 4}; String al = "LRRRRL";
        tableHead(new String[]{"事件", "物理时钟", "Lamport", "HLC (l, c)", "l − 物理时钟", ""}, w, al);

        long prevPT = -1, prevL = -1, prevC = -1, maxDrift = 0;
        int n = 0, backCnt = 0;
        boolean monotonic = true;
        long[] jumps = {0, 0, 0, -800, 0, 0, 0, 0};   // 第 4 个事件处 NTP 把时钟往回拨 800ms

        for (long jump : jumps) {
            pt += jump != 0 ? jump : 20 + rnd.nextInt(70);
            long[] hc = hlc.local(pt);
            boolean back = prevPT >= 0 && pt < prevPT;
            if (back) backCnt++;
            if (hc[0] < prevL || (hc[0] == prevL && hc[1] <= prevC)) monotonic = false;
            prevL = hc[0]; prevC = hc[1];
            maxDrift = Math.max(maxDrift, hc[0] - pt);
            n++;
            tableRow(new String[]{"#" + n + (jump != 0 ? " *回拨*" : ""), "" + pt, "" + lam.local(),
                    "(" + hc[0] + ", " + hc[1] + ")", "+" + (hc[0] - pt) + " ms",
                    back ? "  ⟵ 物理时钟倒退了" : ""}, w, al);
            prevPT = pt;
        }
        System.out.println();
        row("物理时钟倒退次数", backCnt + " 次　← 任何基于它的排序 / TTL / 快照都会出错");
        row("HLC 是否严格单调", monotonic ? "是 ✓" : "否 ✗");
        row("HLC 与物理时钟的最大偏差", maxDrift + " ms　← 有界，且回拨结束后会自动收敛回 0");
        System.out.println("""

    ▸ 回拨发生后，HLC 的 l 冻结在回拨前的值不动，靠 c 递增维持严格单调；
      等物理时钟重新追上来，c 自动归零，l 继续贴着物理时间走。
    ▸ Lamport 也单调，但它的数值和现实时间毫无关系 —— 你没法用它做
      "给我 10:00 那一刻的快照"这类时间范围查询。HLC 两样都要到了，
      代价只是每个时间戳多一个整数。CockroachDB / MongoDB 用的就是它。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  三个思考题（答案在课件 §1.3 / §1.5）：

    1. -skew 设成 0，LWW 还会丢数据吗？先猜再跑。
    2. Lab 1-2 里"被 Lamport 误导的对数"能不能降到 0？
       如果把所有进程都两两互发一次消息会怎样？（提示：还剩多少并发对）
    3. 向量时钟能判定并发，那它能告诉你"该保留哪一条"吗？
       如果不能，谁能？

  下一站：Part 2 · 复制与一致性模型 —— 把"一致性"拆成一个精确的谱系。""");
        System.out.println("═".repeat(78) + "\n");
    }

    // ─── 输出工具（按显示宽度对齐，中英混排不错位）────────────────────────
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
        System.out.printf("  Lab 1-%d · %s%n  %s%n", n, title, sub);
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
                case "skew" -> SKEW = Long.parseLong(v);
                case "gap"  -> GAP = Long.parseLong(v);
                case "jit"  -> JIT = Long.parseLong(v);
                case "seed" -> SEED = Long.parseLong(v);
                default     -> System.out.println("未知参数: " + k);
            }
        }
    }
}
