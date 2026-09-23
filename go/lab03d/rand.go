package main

import "strconv"

// Rand 是一个最简单的 xorshift32。
//
// 为什么不用 math/rand：Java 版要产出【逐字节相同】的输出，
// 两边各自的标准库随机数是对不上的。自己写 8 行，两边一模一样。
type Rand struct{ s uint32 }

func NewRand(seed int64) *Rand {
	v := uint32(seed)
	if v == 0 {
		v = 1
	}
	return &Rand{s: v}
}

func (r *Rand) Next() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}

func (r *Rand) Float() float64 { return float64(r.Next()) / 4294967296.0 }
func (r *Rand) Intn(n int) int { return int(r.Float() * float64(n)) } // 用高位，避免低位周期性
func (r *Rand) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		swap(i, r.Intn(i+1))
	}
}

func join(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x
	}
	return out
}
func itoa(n int) string { return strconv.Itoa(n) }
