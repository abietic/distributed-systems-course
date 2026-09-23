package main

import "time"

// Outcome 是客户端「以为」的结果——注意它和服务器的真实状态可能完全不一致。
type Outcome int

const (
	ClientSuccess Outcome = iota // 收到了响应
	ClientGaveUp                 // 超时放弃（服务器可能已经执行过！）
)

// DoWithRetry 是客户端的重试策略。
//
//	maxRetry = 0 ⇒ at-most-once  语义：不重试，可能漏做
//	maxRetry > 0 ⇒ at-least-once 语义：重试，可能重做
//
// 没有第三种选择。exactly-once 不是靠重试策略得到的，
// 而是靠「at-least-once + 服务端幂等」得到的。
func DoWithRetry(net *UnreliableNet, maxRetry int, timeout time.Duration, exec func()) (Outcome, int) {
	for attempt := 0; attempt <= maxRetry; attempt++ {
		if err := net.Call(timeout, exec); err == nil {
			return ClientSuccess, attempt + 1
		}
		// err == ErrTimeout：我们不知道 exec 有没有跑过，只能选择再试一次。
	}
	return ClientGaveUp, maxRetry + 1
}
