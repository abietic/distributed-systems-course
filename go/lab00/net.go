package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// ErrTimeout 就是本章反复强调的「第三态」：
// 它既不表示"服务器执行了"，也不表示"服务器没执行"，它表示"我不知道"。
var ErrTimeout = errors.New("timeout: 不知道服务器执行了没有")

// NetConfig 描述一条不可靠信道的行为。四个字段对应第 0.3 节的四种网络恶行。
type NetConfig struct {
	Loss    float64       // 单向丢包率（请求和响应各判定一次）
	Latency time.Duration // 平均单向延迟
	Jitter  time.Duration // 延迟抖动，实际延迟 = Latency ± Jitter
	DupRate float64       // 重复投递率（仅 Lab 0-1 演示用）
}

// UnreliableNet 模拟一条会丢包、延迟、乱序、重复的网络。
type UnreliableNet struct {
	cfg NetConfig

	mu       sync.Mutex
	rnd      *rand.Rand
	ReqSent  int // 客户端发出的请求次数（含重试）
	ReqLost  int // 请求方向丢失的包数
	RespLost int // 响应方向丢失的包数
}

func NewNet(cfg NetConfig, seed int64) *UnreliableNet {
	return &UnreliableNet{cfg: cfg, rnd: rand.New(rand.NewSource(seed))}
}

func (n *UnreliableNet) roll() float64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.rnd.Float64()
}

func (n *UnreliableNet) oneWay() time.Duration {
	n.mu.Lock()
	j := time.Duration(0)
	if n.cfg.Jitter > 0 {
		j = time.Duration(n.rnd.Int63n(int64(2*n.cfg.Jitter))) - n.cfg.Jitter
	}
	n.mu.Unlock()
	d := n.cfg.Latency + j
	if d < 0 {
		d = 0
	}
	return d
}

func (n *UnreliableNet) bump(p *int) { n.mu.Lock(); *p++; n.mu.Unlock() }

// ============================================================================
// Call 是整个 Lab 最重要的一个函数。请逐行读它的四条路径。
//
//	exec 是「服务器端的副作用」——一旦它被调用，钱就已经扣了，无法撤销。
//	返回 nil    → 客户端确切知道成功
//	返回 ErrTimeout → 客户端什么都不知道（第三态）
//
// 注意：路径 ② 和路径 ③ 在客户端看来完全一样，都是超时。
// 这就是第 0.5 节说的「你无法区分慢和死」。
// ============================================================================
func (n *UnreliableNet) Call(timeout time.Duration, exec func()) error {
	n.bump(&n.ReqSent)
	ack := make(chan struct{}, 1)

	go func() {
		time.Sleep(n.oneWay()) // 请求在网络中飞行

		if n.roll() < n.cfg.Loss { // ① 请求丢了 —— 服务器永远不会执行
			n.bump(&n.ReqLost)
			return
		}

		exec() // ② 服务器执行副作用。到这一行为止，钱已经扣了。

		time.Sleep(n.oneWay()) // 响应在网络中飞行

		if n.roll() < n.cfg.Loss { // ③ 响应丢了 —— 服务器做了，但客户端不知道
			n.bump(&n.RespLost)
			return
		}
		ack <- struct{}{} // ④ 圆满完成
	}()

	select {
	case <-ack:
		return nil
	case <-time.After(timeout):
		// 注意：这里返回后，上面那个 goroutine 可能仍在运行，
		// 甚至可能在几毫秒后才调用 exec()。真实世界里也是如此——
		// 你的超时并不会让对方停下来。
		return ErrTimeout
	}
}

// DeliverOnce 用于 Lab 0-1：单向投递一条消息，演示丢包 / 延迟 / 乱序 / 重复。
// 到达时调用 onArrive(seq, copyNo)；丢失则什么都不发生。
func (n *UnreliableNet) DeliverOnce(seq int, onArrive func(seq, copyNo int), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(n.oneWay())
		if n.roll() < n.cfg.Loss {
			n.bump(&n.ReqLost)
			return
		}
		onArrive(seq, 1)
		// 底层重传 + 原包其实没丢 ⇒ 同一条消息到达两次
		if n.roll() < n.cfg.DupRate {
			time.Sleep(n.oneWay() / 2)
			onArrive(seq, 2)
		}
	}()
}
