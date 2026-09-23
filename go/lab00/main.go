// Lab 0 · 不可靠信道与第三态
//
// 运行：  cd go/lab00 && go run .
// 调参：  go run . -loss 0.4 -retry 5 -n 500
//
// 配套课件：courseware/ch00-intro.html 第 0.3 / 0.5 / 0.6 节
package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"dsc/internal/tui"
)

const (
	amount  = 100 // 每笔转账金额
	balance = 1_000_000
)

var (
	nFlag     = flag.Int("n", 200, "转账笔数")
	lossFlag  = flag.Float64("loss", 0.30, "单向丢包率 0~1")
	retryFlag = flag.Int("retry", 3, "最大重试次数")
	seedFlag  = flag.Int64("seed", 42, "随机种子（同一种子结果可复现）")
	crashFlag = flag.Float64("crash", 0.30, "Lab 0-4 中两步之间崩溃的概率")
)

func cfg() NetConfig {
	return NetConfig{
		Loss:    *lossFlag,
		Latency: 1200 * time.Microsecond,
		Jitter:  800 * time.Microsecond,
		DupRate: 0.25,
	}
}

const timeout = 5 * time.Millisecond

func main() {
	flag.Parse()
	fmt.Printf("\n配置：转账 %d 笔 × %d 元 ｜ 单向丢包率 %.0f%% ｜ 最大重试 %d 次 ｜ 超时 %v ｜ 种子 %d\n",
		*nFlag, amount, *lossFlag*100, *retryFlag, timeout, *seedFlag)

	lab01()
	lab02()
	lab03()
	lab04()
	epilogue()
}

// ───────────────────────────────────────────────────────────────────────────
// ───────────────────────────────────────────────────────────────────────────
// Lab 0-1：网络的四种恶行
// ───────────────────────────────────────────────────────────────────────────
func lab01() {
	tui.Head(0, 1, "不可靠信道：丢包 / 延迟 / 乱序 / 重复",
		"按 1..10 的顺序发出 10 条消息，看接收端到底收到了什么")

	c := cfg()
	c.Loss = 0.15                      // 这一个演示刻意降低丢包，让存活消息多一些
	c.Jitter = 1500 * time.Microsecond // 抖动 > 发送间隔 ⇒ 乱序必然出现
	net := NewNet(c, *seedFlag)
	var mu sync.Mutex
	var arrivals []struct {
		Seq, Copy int
		At        time.Time
	}
	var wg sync.WaitGroup

	start := time.Now()
	for i := 1; i <= 10; i++ {
		net.DeliverOnce(i, func(seq, copyNo int) {
			mu.Lock()
			arrivals = append(arrivals, struct {
				Seq, Copy int
				At        time.Time
			}{seq, copyNo, time.Now()})
			mu.Unlock()
		}, &wg)
		time.Sleep(150 * time.Microsecond) // 发送方严格按序、匀速发出
	}
	wg.Wait()

	sort.Slice(arrivals, func(i, j int) bool { return arrivals[i].At.Before(arrivals[j].At) })

	got := make(map[int]int)
	var order []string
	for _, a := range arrivals {
		got[a.Seq]++
		tag := fmt.Sprint(a.Seq)
		if a.Copy > 1 {
			tag += "*" // 重复到达
		}
		order = append(order, tag)
	}

	var lost, dup, inversions []string
	for i := 1; i <= 10; i++ {
		switch {
		case got[i] == 0:
			lost = append(lost, fmt.Sprint(i))
		case got[i] > 1:
			dup = append(dup, fmt.Sprint(i))
		}
	}
	prev := 0
	for _, a := range arrivals {
		if a.Copy == 1 {
			if a.Seq < prev {
				inversions = append(inversions, fmt.Sprintf("%d 排在 %d 之后", a.Seq, prev))
			}
			prev = a.Seq
		}
	}

	tui.Row("发送顺序", "1 2 3 4 5 6 7 8 9 10")
	tui.Row("实际到达顺序", strings.Join(order, " "), "        （* = 重复投递）")
	tui.Row("丢失的消息", or(strings.Join(lost, " "), "无"))
	tui.Row("重复到达的消息", or(strings.Join(dup, " "), "无"))
	tui.Row("乱序", or(strings.Join(inversions, "；"), "无"))
	fmt.Printf("    用时 %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println(`
    ▸ 结论：发送方眼中"我按 1..10 发出去了"，接收方眼中却是另一个故事。
      TCP 能在单条连接内修复丢包/乱序/重复，但修不了连接本身断掉——
      而分布式系统的麻烦恰恰发生在连接断掉之后。`)
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// ───────────────────────────────────────────────────────────────────────────
// Lab 0-2：at-most-once vs at-least-once —— 两条路都是错的
// ───────────────────────────────────────────────────────────────────────────
func lab02() {
	tui.Head(0, 2, "at-most-once vs at-least-once",
		"同样的网络故障，两种重试策略，两种错法")

	w := []int{26, 11, 11, 11, 17}
	tui.TableHead([]string{"策略", "应扣(元)", "实扣(元)", "差额(元)", "客户端认为成功"}, w)

	runNaive := func(label string, maxRetry int) {
		net := NewNet(cfg(), *seedFlag)
		bank := NewBank(balance)
		ok := 0
		for i := 0; i < *nFlag; i++ {
			out, _ := DoWithRetry(net, maxRetry, timeout, func() { bank.DeductNaive(amount) })
			if out == ClientSuccess {
				ok++
			}
		}
		bal, _ := bank.Snapshot()
		expect, actual := *nFlag*amount, balance-bal
		tui.TableRow([]string{label, fmt.Sprint(expect), fmt.Sprint(actual),
			fmt.Sprintf("%+d", actual-expect), fmt.Sprintf("%d/%d", ok, *nFlag)}, w)
	}

	runNaive("A · 不重试", 0)
	runNaive(fmt.Sprintf("B · 重试至多 %d 次", *retryFlag), *retryFlag)

	fmt.Println(`
    ▸ A（at-most-once）：差额为负 ⇒ 漏扣。用户看到"转账失败"，但有一部分其实
      已经在服务器上执行了（幽灵成功），只是响应包丢了。
    ▸ B（at-least-once）：差额为正 ⇒ 重复扣款。每一次超时重试，都可能撞上
      "服务器其实已经做过了"的情况。
    ▸ 两条路都错，而且错得方向相反。这不是代码 bug，是第三态的必然后果。`)
}

// ───────────────────────────────────────────────────────────────────────────
// Lab 0-3：幂等键把 at-least-once 变成 effectively-once
// ───────────────────────────────────────────────────────────────────────────
func lab03() {
	tui.Head(0, 3, "幂等键：让「重试」变得安全",
		"同一份网络故障 + 同样的重试次数，只改服务端")

	w := []int{26, 11, 11, 11, 17}
	tui.TableHead([]string{"服务端实现", "应扣(元)", "实扣(元)", "差额(元)", "重复扣款次数"}, w)

	run := func(label string, idem bool) {
		net := NewNet(cfg(), *seedFlag)
		bank := NewBank(balance)
		for i := 0; i < *nFlag; i++ {
			key := fmt.Sprintf("txn-%04d", i) // 幂等键：由客户端生成，重试时保持不变
			DoWithRetry(net, *retryFlag, timeout, func() {
				if idem {
					bank.DeductIdempotent(key, amount)
				} else {
					bank.DeductNaive(amount)
				}
			})
		}
		bal, deducts := bank.Snapshot()
		expect, actual := *nFlag*amount, balance-bal
		tui.TableRow([]string{label, fmt.Sprint(expect), fmt.Sprint(actual),
			fmt.Sprintf("%+d", actual-expect), fmt.Sprint(max(0, deducts-*nFlag))}, w)
	}

	run("无幂等键", false)
	run("有幂等键（同事务）", true)

	fmt.Println(`
    ▸ 有幂等键的那一行，重复扣款恒为 0。差额若仍为负，说明有几笔业务的请求包
      被连续丢了 ` + fmt.Sprint(*retryFlag+1) + ` 次，服务器压根没收到过——
      这部分只能靠加大重试次数压低，幂等性解决不了。
    ▸ 记住这个组合拳：at-least-once 传输 + 服务端幂等 = 工程上的 exactly-once 效果。
      Kafka / Flink 所谓的 exactly-once 也是这么做的，没有魔法。`)
}

// ───────────────────────────────────────────────────────────────────────────
// Lab 0-4：去重表和业务操作必须原子
// ───────────────────────────────────────────────────────────────────────────
func lab04() {
	tui.Head(0, 4, "陷阱：去重表与业务操作被拆成两步",
		fmt.Sprintf("在两步之间以 %.0f%% 的概率注入崩溃", *crashFlag*100))

	w := []int{38, 11, 11, 11}
	tui.TableHead([]string{"写入顺序", "应扣(元)", "实扣(元)", "差额(元)"}, w)

	run := func(label string, markFirst bool) (lost, double int) {
		net := NewNet(cfg(), *seedFlag)
		bank := NewBank(balance)
		crashRnd := NewNet(NetConfig{}, *seedFlag+7) // 只借用它的随机源
		for i := 0; i < *nFlag; i++ {
			key := fmt.Sprintf("txn-%04d", i)
			DoWithRetry(net, *retryFlag, timeout, func() {
				bank.DeductSplitDedup(key, amount, markFirst, crashRnd.roll() < *crashFlag)
			})
		}
		bal, _ := bank.Snapshot()
		expect, actual := *nFlag*amount, balance-bal
		tui.TableRow([]string{label, fmt.Sprint(expect), fmt.Sprint(actual),
			fmt.Sprintf("%+d", actual-expect)}, w)
		return bank.Lost, bank.Double
	}

	l1, _ := run("① 先写去重表 → 崩溃 → 再扣款", true)
	_, d2 := run("② 先扣款 → 崩溃 → 再写去重表", false)

	tui.Row("① 造成的永久丢单", fmt.Sprintf("%d 笔（去重表说做过了，钱其实没扣，重试全被拦截）", l1))
	tui.Row("② 造成的重复扣款", fmt.Sprintf("%d 次（钱扣了但没记录，重试又扣一次）", d2))

	fmt.Println(`
    ▸ 两种拆分顺序，两种事故，没有哪种更安全。
    ▸ 唯一正确的做法：把「写去重表」和「业务副作用」放进同一个原子操作
      （同一个数据库事务）。这正是 Part 4 里本地消息表 / Outbox / 事务消息
      要解决的同一个问题——只不过那时候两个副作用分别在数据库和消息队列里，
      没有一个共同的事务能包住它们，于是才需要 2PC、TCC、Saga 这些方案。`)
}

func epilogue() {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Println(`  做完了，回答这三个问题（答案在课件 §0.5 / §0.6）：

    1. 把 -loss 调成 0 再跑一遍。Lab 0-2 的两种策略还有差异吗？为什么？
    2. -loss 0.5 -retry 10 时，Lab 0-2 的 B 组最多会重复扣几次？先猜再跑。
    3. Lab 0-3 里，如果幂等键改成"每次重试都生成一个新的 UUID"，
       结果会变成什么样？（提示：幂等键必须由业务唯一性决定，不能随机生成）

  下一站：Part 1 · 时间、顺序与因果 —— 亲手实现 Lamport 时钟与向量时钟。`)
	fmt.Printf("%s\n\n", strings.Repeat("═", 78))
}
