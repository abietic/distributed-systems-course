package main

import "sort"

// 2PL 与 SSI 的对照模拟。
//
// 两个模拟器共用同一个物理前提：系统每 tick 最多能执行 PAR 个操作。
// 这一条很重要——它让「白做的功」真的变成时间，而不是一个漂亮的免费数字。
const PAR = 4

type CCOp struct {
	Key   int
	Write bool
}
type CCTxn struct {
	ID  int
	Ops []CCOp
}
type CCResult struct {
	Tick, Aborts, Wasted, Work int
	Tput                       float64
}

var nkey = 40

const hotKeys = 2

func MkTxns(c, o int, hot float64, rnd *Rand) []CCTxn {
	nkey = c * o // 键空间随负载伸缩，否则「低冲突」根本低不下来
	if nkey < 24 {
		nkey = 24
	}
	ts := make([]CCTxn, c)
	for i := 0; i < c; i++ {
		ops := make([]CCOp, o)
		for j := 0; j < o; j++ {
			k := rnd.Intn(nkey)
			if rnd.Float() < hot {
				k = rnd.Intn(hotKeys)
			}
			ops[j] = CCOp{Key: k, Write: rnd.Float() < 0.5}
		}
		ts[i] = CCTxn{ID: i, Ops: ops}
	}
	return ts
}

type lk struct {
	trx int
	x   bool
}
type st2 struct {
	id, p, cool int
	ops         []CCOp
	locks       []int
	done        bool
}

// Sim2PL：悲观。拿不到锁就等（等待期间不占用 CPU）；等待图成环 = 死锁。
// 关键性质：等待【不浪费】已经做的功——等到了就接着往下做。
// 进度保证：编号最小的活跃事务永远不会被选为牺牲者 ⇒ 它一定跑得完，然后轮到下一个。
func Sim2PL(txns []CCTxn) CCResult {
	st := make([]*st2, len(txns))
	for i, t := range txns {
		st[i] = &st2{id: t.ID, ops: t.Ops}
	}
	owner := map[int]lk{}
	tick, aborts, wasted, work := 0, 0, 0, 0
	anyLeft := func() bool {
		for _, s := range st {
			if !s.done {
				return true
			}
		}
		return false
	}
	for anyLeft() && tick < 8000 {
		tick++
		alive := []*st2{}
		for _, s := range st {
			if !s.done {
				alive = append(alive, s)
			}
		}
		// 先算出这一 tick 谁【能】动（锁没被占住），再按并行度预算执行
		runnable := []*st2{}
		waitFor := map[int]int{}
		cooling := false
		for _, s := range alive {
			if s.cool > 0 {
				s.cool--
				cooling = true
				continue
			}
			op := s.ops[s.p]
			if o, ok := owner[op.Key]; ok && o.trx != s.id && (o.x || op.Write) {
				waitFor[s.id] = o.trx
				continue
			}
			runnable = append(runnable, s)
		}
		if len(runnable) == 0 && !cooling {
			// 没有任何事务能动 ⇒ 等待图里必然有环。把环找出来，只在【环里】挑牺牲者——
			// 中止一个环外的事务不会解开任何东西，这正是「死锁检测」和「随便杀一个」的区别。
			var cyc []int
			// Go 的 map 遍历顺序是随机的，这里必须排序，
			// 否则同一个种子跑两次结果都可能不同（Java 版也就对不上了）
			wkeys := make([]int, 0, len(waitFor))
			for a := range waitFor {
				wkeys = append(wkeys, a)
			}
			sort.Ints(wkeys)
			for _, a := range wkeys {
				path, seen := []int{}, map[int]bool{}
				cur, ok := a, true
				for ok && !seen[cur] {
					seen[cur] = true
					path = append(path, cur)
					cur, ok = waitFor[cur]
				}
				if ok {
					for i, v := range path {
						if v == cur {
							cyc = path[i:]
							break
						}
					}
					break
				}
			}
			protectedID := alive[0].id
			pool := []*st2{}
			for _, id := range cyc {
				for _, s := range alive {
					if s.id == id && s.id != protectedID {
						pool = append(pool, s)
					}
				}
			}
			if len(pool) == 0 {
				for _, s := range alive {
					if s.id != protectedID && len(s.locks) > 0 {
						pool = append(pool, s)
					}
				}
			}
			if len(pool) == 0 {
				for _, s := range alive {
					if s.id != protectedID {
						pool = append(pool, s)
					}
				}
			}
			if len(pool) == 0 {
				break
			}
			v := pool[0]
			for _, s := range pool { // 挑已做的功最少的，浪费最小
				if s.p < v.p {
					v = s
				}
			}
			aborts++
			wasted += v.p
			for _, k := range v.locks {
				if o, ok := owner[k]; ok && o.trx == v.id {
					delete(owner, k)
				}
			}
			v.locks = nil
			v.p = 0
			v.cool = 2
			continue
		}
		budget := PAR
		for _, s := range runnable {
			if budget <= 0 {
				break
			}
			op := s.ops[s.p]
			if o, ok := owner[op.Key]; ok && o.trx != s.id && (o.x || op.Write) {
				continue // 同一 tick 内被别人抢先拿走了
			}
			if o, ok := owner[op.Key]; !ok || o.trx != s.id {
				owner[op.Key] = lk{s.id, op.Write}
				s.locks = append(s.locks, op.Key)
			} else if op.Write {
				owner[op.Key] = lk{s.id, true}
			}
			s.p++
			work++
			budget--
			if s.p >= len(s.ops) {
				s.done = true
				for _, k := range s.locks {
					if o, ok := owner[k]; ok && o.trx == s.id {
						delete(owner, k)
					}
				}
				s.locks = nil
			}
		}
	}
	return CCResult{tick, aborts, wasted, work, float64(len(txns)) / float64(max(tick, 1)) * 100}
}

type stS struct {
	id, p, start int
	ops          []CCOp
	rs, ws       map[int]bool
	done         bool
}

// SimSSI：乐观。谁也不等，全速往前跑；跑完一整个事务才验证
//
//	「我读到的前提有没有被别人推翻」——被推翻就整个重做。
//
// 验证条件用的是 SSI 的危险结构近似：
//   - 写-写冲突（快照隔离本来就要求的 first-committer-wins），或
//   - 双向的读-写反依赖（我读了它写的，同时它读了我写的）—— 这才是真正会成环的那种。
//
// 关键性质：中止 = 这一轮做的功【全部白费】，而且白费的功照样占 CPU。
func SimSSI(txns []CCTxn) CCResult {
	st := make([]*stS, len(txns))
	for i, t := range txns {
		rs, ws := map[int]bool{}, map[int]bool{}
		for _, o := range t.Ops {
			if o.Write {
				ws[o.Key] = true
			} else {
				rs[o.Key] = true
			}
		}
		st[i] = &stS{id: t.ID, ops: t.Ops, rs: rs, ws: ws}
	}
	type hEntry struct{ rs, ws map[int]bool }
	hist := []hEntry{}
	tick, aborts, wasted, work := 0, 0, 0, 0
	inter := func(a, b map[int]bool) bool {
		for k := range a {
			if b[k] {
				return true
			}
		}
		return false
	}
	anyLeft := func() bool {
		for _, s := range st {
			if !s.done {
				return true
			}
		}
		return false
	}
	for anyLeft() && tick < 8000 {
		tick++
		budget := PAR
		for _, s := range st {
			if s.done || budget <= 0 {
				continue
			}
			s.p++
			work++
			budget--
			if s.p < len(s.ops) {
				continue
			}
			conflict := false
			for i := s.start; i < len(hist) && !conflict; i++ {
				u := hist[i]
				if inter(s.ws, u.ws) {
					conflict = true // ① 写-写
				} else if inter(s.rs, u.ws) && inter(s.ws, u.rs) {
					conflict = true // ② 双向 rw 反依赖
				}
			}
			if conflict {
				aborts++
				wasted += len(s.ops)
				s.p = 0
				s.start = len(hist)
			} else {
				s.done = true
				hist = append(hist, hEntry{s.rs, s.ws})
			}
		}
	}
	return CCResult{tick, aborts, wasted, work, float64(len(txns)) / float64(max(tick, 1)) * 100}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
