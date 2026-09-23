/*
 * Lab 0 · 不可靠信道与第三态（Java 版）
 *
 * 运行：  cd java/lab00 && java Lab00.java
 * 调参：  java Lab00.java -loss 0.4 -retry 5 -n 500
 *
 * 需要 JDK 17+（用的是单文件源码启动模式，不需要 Maven / Gradle）。
 * 配套课件：courseware/ch00-intro.html 第 0.3 / 0.5 / 0.6 节
 */
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.locks.LockSupport;

public class Lab00 {

    // ─── 可调参数 ──────────────────────────────────────────────────────────
    static int    N      = 200;   // 转账笔数
    static double LOSS   = 0.30;  // 单向丢包率
    static int    RETRY  = 3;     // 最大重试次数
    static long   SEED   = 42;    // 随机种子（同种子结果可复现）
    static double CRASH  = 0.30;  // Lab 0-4 两步之间的崩溃概率

    static final int  AMOUNT   = 100;
    static final int  BALANCE  = 1_000_000;
    static final long LAT_US   = 1200;   // 平均单向延迟（微秒）
    static final long JIT_US   = 800;    // 抖动
    static final long TIMEOUT_MS = 5;    // 客户端超时

    static final ExecutorService POOL = Executors.newCachedThreadPool(r -> {
        Thread t = new Thread(r); t.setDaemon(true); return t;
    });

    // ══════════════════════════════════════════════════════════════════════
    // 不可靠信道
    // ══════════════════════════════════════════════════════════════════════
    static class Net {
        final double loss, dupRate;
        final long latUs, jitUs;
        final Random rnd;
        int reqSent, reqLost, respLost;

        Net(double loss, long latUs, long jitUs, double dupRate, long seed) {
            this.loss = loss; this.latUs = latUs; this.jitUs = jitUs;
            this.dupRate = dupRate; this.rnd = new Random(seed);
        }

        synchronized double roll() { return rnd.nextDouble(); }

        void fly() {                       // 模拟一次单向飞行
            long j = jitUs == 0 ? 0 : (long) ((roll() * 2 - 1) * jitUs);
            long us = Math.max(0, latUs + j);
            LockSupport.parkNanos(us * 1_000L);
        }

        /* ------------------------------------------------------------------
         * call 是整个 Lab 最重要的一个方法。请逐行读它的四条路径。
         *
         *   exec 是「服务器端的副作用」——一旦它被调用，钱就已经扣了，无法撤销。
         *   返回 true  → 客户端确切知道成功
         *   返回 false → 超时。客户端什么都不知道（第三态）
         *
         * 注意：路径 ② 和路径 ③ 在客户端看来完全一样，都是超时。
         * 这就是课件 §0.5 说的「你无法区分慢和死」。
         * ---------------------------------------------------------------- */
        boolean call(long timeoutMs, Runnable exec) {
            synchronized (this) { reqSent++; }
            CompletableFuture<Boolean> ack = new CompletableFuture<>();

            POOL.execute(() -> {
                fly();                                   // 请求在网络中飞行
                if (roll() < loss) {                     // ① 请求丢了 → 服务器永远不会执行
                    synchronized (this) { reqLost++; }
                    return;
                }
                exec.run();                              // ② 服务器执行副作用。到这一行为止，钱已经扣了。
                fly();                                   // 响应在网络中飞行
                if (roll() < loss) {                     // ③ 响应丢了 → 服务器做了，但客户端不知道
                    synchronized (this) { respLost++; }
                    return;
                }
                ack.complete(true);                      // ④ 圆满完成
            });

            try {
                return ack.get(timeoutMs, TimeUnit.MILLISECONDS);
            } catch (TimeoutException e) {
                // 注意：这里返回后，上面那个任务可能仍在运行，甚至可能在几毫秒后
                // 才调用 exec()。真实世界里也是如此——你的超时不会让对方停下来。
                return false;
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }

        /** 用于 Lab 0-1：单向投递一条消息，演示丢包 / 延迟 / 乱序 / 重复。 */
        void deliverOnce(int seq, java.util.function.BiConsumer<Integer,Integer> onArrive,
                         CountDownLatch done) {
            POOL.execute(() -> {
                try {
                    fly();
                    if (roll() < loss) { synchronized (this) { reqLost++; } return; }
                    onArrive.accept(seq, 1);
                    if (roll() < dupRate) {            // 底层重传 + 原包其实没丢 ⇒ 到达两次
                        LockSupport.parkNanos(latUs * 500L);
                        onArrive.accept(seq, 2);
                    }
                } finally { done.countDown(); }
            });
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // 服务器端：三种「扣款」实现，对应三种工程水平
    // ══════════════════════════════════════════════════════════════════════
    static class Bank {
        int balance, deducts, lost, doubled;
        final Set<String> dedup = new HashSet<>();

        Bank(int balance) { this.balance = balance; }

        /** ① 天真实现：没有幂等键，收到几次请求就扣几次钱。 */
        synchronized void deductNaive(int amount) {
            balance -= amount; deducts++;
        }

        /** ② 正确实现：幂等键 + 去重表，且「查重 + 扣款」在同一个临界区（= 同一个事务）内。 */
        synchronized void deductIdempotent(String key, int amount) {
            if (dedup.contains(key)) return;   // 这个业务已经做过了，直接返回上次结果
            dedup.add(key);                    // 记录
            balance -= amount; deducts++;      // 副作用
            // ↑ 记录与副作用同处一个临界区 ⇒ 要么都发生，要么都不发生
        }

        /** ③ 错误实现：去重表与业务操作被拆成两个非原子步骤，中间可能崩溃。 */
        synchronized void deductSplit(String key, int amount, boolean markFirst, boolean crash) {
            if (dedup.contains(key)) return;
            if (markFirst) {
                dedup.add(key);
                if (crash) { lost++; return; }     // ★ 去重表说"做过了"，钱却没扣，重试全被拦截 ⇒ 永久丢单
                balance -= amount; deducts++;
            } else {
                balance -= amount; deducts++;
                if (crash) { doubled++; return; }  // ★ 钱扣了但没记录，重试会再扣一次 ⇒ 重复扣款
                dedup.add(key);
            }
        }
    }

    // ══════════════════════════════════════════════════════════════════════
    // 客户端重试策略
    //   maxRetry = 0 ⇒ at-most-once  语义：不重试，可能漏做
    //   maxRetry > 0 ⇒ at-least-once 语义：重试，可能重做
    //   没有第三种选择。exactly-once 不是靠重试策略得到的，
    //   而是靠「at-least-once + 服务端幂等」得到的。
    // ══════════════════════════════════════════════════════════════════════
    static boolean doWithRetry(Net net, int maxRetry, Runnable exec) {
        for (int attempt = 0; attempt <= maxRetry; attempt++) {
            if (net.call(TIMEOUT_MS, exec)) return true;
            // 超时：我们不知道 exec 有没有跑过，只能选择再试一次。
        }
        return false;
    }

    // ══════════════════════════════════════════════════════════════════════
    public static void main(String[] args) throws Exception {
        // 保证中文在任何平台的终端都能正确输出
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：转账 %d 笔 × %d 元 ｜ 单向丢包率 %.0f%% ｜ 最大重试 %d 次 ｜ 超时 %dms ｜ 种子 %d%n",
                N, AMOUNT, LOSS * 100, RETRY, TIMEOUT_MS, SEED);
        lab01(); lab02(); lab03(); lab04(); epilogue();
        POOL.shutdownNow();
    }

    // ─── Lab 0-1 ──────────────────────────────────────────────────────────
    static void lab01() throws Exception {
        head(1, "不可靠信道：丢包 / 延迟 / 乱序 / 重复",
                "按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么");

        // 这一个演示刻意降低丢包、加大抖动（抖动 > 发送间隔 ⇒ 乱序必然出现）
        Net net = new Net(0.15, LAT_US, 1500, 0.25, SEED);
        List<int[]> arrivals = Collections.synchronizedList(new ArrayList<>());
        CountDownLatch done = new CountDownLatch(10);

        long t0 = System.nanoTime();
        for (int i = 1; i <= 10; i++) {
            net.deliverOnce(i, (seq, copyNo) -> arrivals.add(new int[]{seq, copyNo}), done);
            LockSupport.parkNanos(150_000L);   // 发送方严格按序、匀速发出
        }
        done.await(5, TimeUnit.SECONDS);
        Thread.sleep(20);

        Map<Integer,Integer> got = new HashMap<>();
        StringBuilder order = new StringBuilder();
        int prev = 0; List<String> inv = new ArrayList<>();
        for (int[] a : arrivals) {
            got.merge(a[0], 1, Integer::sum);
            order.append(a[0]).append(a[1] > 1 ? "* " : " ");
            if (a[1] == 1) { if (a[0] < prev) inv.add(a[0] + " 排在 " + prev + " 之后"); prev = a[0]; }
        }
        List<String> lost = new ArrayList<>(), dup = new ArrayList<>();
        for (int i = 1; i <= 10; i++) {
            int c = got.getOrDefault(i, 0);
            if (c == 0) lost.add("" + i); else if (c > 1) dup.add("" + i);
        }
        row("发送顺序", "1 2 3 4 5 6 7 8 9 10");
        row("实际到达顺序", order.toString().trim() + "        （* = 重复投递）");
        row("丢失的消息", or(String.join(" ", lost)));
        row("重复到达的消息", or(String.join(" ", dup)));
        row("乱序", or(String.join("；", inv)));
        System.out.printf("    用时 %dms%n", (System.nanoTime() - t0) / 1_000_000);
        System.out.println("""

    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
      而分布式系统的麻烦恰恰发生在连接断掉之后。""".indent(4));
    }

    // ─── Lab 0-2 ──────────────────────────────────────────────────────────
    static void lab02() {
        head(2, "at-most-once vs at-least-once", "同样的网络故障，两种重试策略，两种错法");
        int[] w = {26, 11, 11, 11, 17};
        tableHead(new String[]{"策略","应扣(元)","实扣(元)","差额(元)","客户端认为成功"}, w);

        for (int mode = 0; mode < 2; mode++) {
            int maxRetry = mode == 0 ? 0 : RETRY;
            String label = mode == 0 ? "A · 不重试" : "B · 重试至多 " + RETRY + " 次";
            Net net = new Net(LOSS, LAT_US, JIT_US, 0, SEED);
            Bank bank = new Bank(BALANCE);
            int ok = 0;
            for (int i = 0; i < N; i++)
                if (doWithRetry(net, maxRetry, () -> bank.deductNaive(AMOUNT))) ok++;
            int expect = N * AMOUNT, actual = BALANCE - bank.balance;
            tableRow(new String[]{label, "" + expect, "" + actual,
                    String.format("%+d", actual - expect), ok + "/" + N}, w);
        }
        System.out.println("""

    ▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
      已经在服务器上执行了（幽灵成功），只是响应包丢了。
    ▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
      "服务器其实已经做过了"的情况。
    ▸ 两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。""".indent(4));
    }

    // ─── Lab 0-3 ──────────────────────────────────────────────────────────
    static void lab03() {
        head(3, "幂等键：让「重试」变得安全", "同一份网络故障 + 同样的重试次数，只改服务端");
        int[] w = {26, 11, 11, 11, 17};
        tableHead(new String[]{"服务端实现","应扣(元)","实扣(元)","差额(元)","重复扣款次数"}, w);

        for (int mode = 0; mode < 2; mode++) {
            boolean idem = mode == 1;
            Net net = new Net(LOSS, LAT_US, JIT_US, 0, SEED);
            Bank bank = new Bank(BALANCE);
            for (int i = 0; i < N; i++) {
                String key = String.format("txn-%04d", i);   // 幂等键由客户端生成，重试时保持不变
                doWithRetry(net, RETRY, () -> {
                    if (idem) bank.deductIdempotent(key, AMOUNT); else bank.deductNaive(AMOUNT);
                });
            }
            int expect = N * AMOUNT, actual = BALANCE - bank.balance;
            tableRow(new String[]{idem ? "有幂等键（同事务）" : "无幂等键", "" + expect, "" + actual,
                    String.format("%+d", actual - expect), "" + Math.max(0, bank.deducts - N)}, w);
        }
        System.out.println("""

    ▸ 有幂等键的那一行，重复扣款恒为 0。差额若仍为负，说明有几笔业务的请求包被
      连续丢了 %d 次，服务器压根没收到过——这部分只能靠加大重试次数压低，
      幂等性解决不了。
    ▸ 记住这个组合拳：at-least-once 传输 + 服务端幂等 = 工程上的 exactly-once 效果。
      Kafka / Flink 所谓的 exactly-once 也是这么做的，没有魔法。""".formatted(RETRY + 1).indent(4));
    }

    // ─── Lab 0-4 ──────────────────────────────────────────────────────────
    static void lab04() {
        head(4, "陷阱：去重表与业务操作被拆成两步",
                String.format("在两步之间以 %.0f%% 的概率注入崩溃", CRASH * 100));
        int[] w = {38, 11, 11, 11};
        tableHead(new String[]{"写入顺序","应扣(元)","实扣(元)","差额(元)"}, w);

        int lostCnt = 0, doubleCnt = 0;
        for (int mode = 0; mode < 2; mode++) {
            boolean markFirst = mode == 0;
            Net net = new Net(LOSS, LAT_US, JIT_US, 0, SEED);
            Random crashRnd = new Random(SEED + 7);
            Bank bank = new Bank(BALANCE);
            for (int i = 0; i < N; i++) {
                String key = String.format("txn-%04d", i);
                doWithRetry(net, RETRY, () ->
                        bank.deductSplit(key, AMOUNT, markFirst, crashRnd.nextDouble() < CRASH));
            }
            int expect = N * AMOUNT, actual = BALANCE - bank.balance;
            tableRow(new String[]{markFirst ? "① 先写去重表 → 崩溃 → 再扣款" : "② 先扣款 → 崩溃 → 再写去重表",
                    "" + expect, "" + actual, String.format("%+d", actual - expect)}, w);
            if (markFirst) lostCnt = bank.lost; else doubleCnt = bank.doubled;
        }
        row("① 造成的永久丢单", lostCnt + " 笔（去重表说做过了，钱其实没扣，重试全被拦截）");
        row("② 造成的重复扣款", doubleCnt + " 次（钱扣了但没记录，重试又扣一次）");
        System.out.println("""

    ▸ 两种拆分顺序，两种事故，没有哪种更安全。
    ▸ 唯一正确的做法：把「写去重表」和「业务副作用」放进同一个原子操作
      （同一个数据库事务）。这正是 Part 4 里本地消息表 / Outbox / 事务消息
      要解决的同一个问题——只不过那时候两个副作用分别在数据库和消息队列里，
      没有一个共同的事务能包住它们，于是才需要 2PC、TCC、Saga 这些方案。""".indent(4));
    }

    static void epilogue() {
        System.out.println("\n" + "═".repeat(78));
        System.out.println("""
  做完了，回答这三个问题（答案在课件 §0.5 / §0.6）：

    1. 把 -loss 调成 0 再跑一遍。Lab 0-2 的两种策略还有差异吗？为什么？
    2. -loss 0.5 -retry 10 时，Lab 0-2 的 B 组最多会重复扣几次？先猜再跑。
    3. Lab 0-3 里，如果幂等键改成"每次重试都生成一个新的 UUID"，
       结果会变成什么样？（提示：幂等键必须由业务唯一性决定，不能随机生成）

  下一站：Part 1 · 时间、顺序与因果 —— 亲手实现 Lamport 时钟与向量时钟。""".indent(2));
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

    static void head(int n, String title, String sub) {
        System.out.println("\n" + "═".repeat(78));
        System.out.printf("  Lab 0-%d · %s%n  %s%n", n, title, sub);
        System.out.println("═".repeat(78));
    }
    static void row(String k, Object v) { System.out.println("    " + padR(k, 30) + " " + v); }
    static void tableHead(String[] cols, int[] w) {
        StringBuilder sb = new StringBuilder("    ");
        for (int i = 0; i < cols.length; i++) sb.append(i == 0 ? padR(cols[i], w[i]) : padL(cols[i], w[i]));
        System.out.println(sb);
        int t = 0; for (int x : w) t += x;
        System.out.println("    " + "─".repeat(t));
    }
    static void tableRow(String[] c, int[] w) {
        StringBuilder sb = new StringBuilder("    ");
        for (int i = 0; i < c.length; i++) sb.append(i == 0 ? padR(c[i], w[i]) : padL(c[i], w[i]));
        System.out.println(sb);
    }
    static String or(String s) { return s.isEmpty() ? "无" : s; }

    static void parseArgs(String[] a) {
        for (int i = 0; i + 1 < a.length; i += 2) {
            String k = a[i].replaceFirst("^-+", ""), v = a[i + 1];
            switch (k) {
                case "n"      -> N = Integer.parseInt(v);
                case "loss"   -> LOSS = Double.parseDouble(v);
                case "retry"  -> RETRY = Integer.parseInt(v);
                case "seed"   -> SEED = Long.parseLong(v);
                case "crash"  -> CRASH = Double.parseDouble(v);
                default       -> System.out.println("未知参数: " + k);
            }
        }
    }
}
