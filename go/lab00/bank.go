package main

import "sync"

// Bank 是服务器端。它有三种「扣款」实现，对应三种工程水平。
type Bank struct {
	mu sync.Mutex

	Balance int // 账户余额
	Deducts int // 真正执行成功的扣款次数（上帝视角统计，客户端看不到）

	dedup  map[string]bool // 幂等去重表：幂等键 -> 是否处理过
	Lost   int             // Lab 0-4：因非原子写入而永久丢失的业务数
	Double int             // Lab 0-4：因非原子写入而重复扣款的次数
}

func NewBank(balance int) *Bank {
	return &Bank{Balance: balance, dedup: map[string]bool{}}
}

// ---------------------------------------------------------------------------
// ① 天真实现：没有幂等键。每收到一次请求就扣一次钱。
//
//	配合「超时重试」使用时，会重复扣款。
//
// ---------------------------------------------------------------------------
func (b *Bank) DeductNaive(amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Balance -= amount
	b.Deducts++
}

// ---------------------------------------------------------------------------
// ② 正确实现：幂等键 + 去重表，且「查重 + 扣款」在同一把锁内完成。
//
//	这把锁在真实系统里对应「同一个数据库事务」。
//
// ---------------------------------------------------------------------------
func (b *Bank) DeductIdempotent(key string, amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.dedup[key] { // 这个业务已经做过了，直接返回上次的结果
		return
	}
	b.dedup[key] = true // 记录
	b.Balance -= amount // 副作用
	b.Deducts++
	// ↑ 记录与副作用同处一个临界区 ⇒ 要么都发生，要么都不发生
}

// ---------------------------------------------------------------------------
// ③ 错误实现：去重表和业务操作被拆成两个非原子步骤，中间可能崩溃。
//
//	这是 Lab 0-4 要复现的两种真实事故。
//
//	markFirst = true  → 先写去重表，再扣款：崩在中间 ⇒ 永久丢单
//	markFirst = false → 先扣款，再写去重表：崩在中间 ⇒ 重复扣款
//
// ---------------------------------------------------------------------------
func (b *Bank) DeductSplitDedup(key string, amount int, markFirst, crash bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.dedup[key] {
		return
	}

	if markFirst {
		b.dedup[key] = true
		if crash {
			b.Lost++ // ★ 崩溃：去重表说"做过了"，但钱没扣。重试会被拦截 ⇒ 这笔业务永久丢失
			return
		}
		b.Balance -= amount
		b.Deducts++
	} else {
		b.Balance -= amount
		b.Deducts++
		if crash {
			b.Double++ // ★ 崩溃：钱扣了但没记录。重试会再扣一次 ⇒ 重复扣款
			return
		}
		b.dedup[key] = true
	}
}

func (b *Bank) Snapshot() (balance, deducts int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Balance, b.Deducts
}
