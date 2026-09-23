// 一个迷你事务引擎：MVCC 版本链 + ReadView + 行锁 + 范围锁 + 串行执行。
//
// 它存在的理由只有一个：让「哪个隔离级别挡得住哪种异常」这句话
// 从【背下来的表格】变成【跑出来的结果】。
// 加起来不到 300 行，但 InnoDB 在这件事上做的核心判断它都做了。
package main

import "sort"

// 隔离配置。RRLOCK 不是一个真的隔离级别，而是「RR + 显式 SELECT ... FOR UPDATE」，
// 因为它才是工程上真正用来堵写偏斜的那个办法。
const (
	RU     = "RU"
	RC     = "RC"
	RR     = "RR"
	RRLOCK = "RRLOCK"
	SER    = "SER"
)

var Levels = []string{RU, RC, RR, RRLOCK, SER}
var LevelName = map[string]string{
	RU: "READ UNCOMMITTED", RC: "READ COMMITTED", RR: "REPEATABLE READ",
	RRLOCK: "RR + FOR UPDATE", SER: "SERIALIZABLE",
}

type Version struct{ Trx, Val int }

// Pred 是一个范围条件，比如 on_call = true。Tag 同时充当范围锁的锁名。
type Pred struct {
	Tag, Label string
	Test       func(k string, v int) bool
}

// ReadView 的四个字段，和 InnoDB 一模一样。
type ReadView struct {
	Mids              []int // 创建时仍活跃（未提交）的事务 id
	Min, Max, Creator int
}

type Tx struct {
	T, ID int
	RV    *ReadView
	Last  int
	Reads []int
	Done  bool
}

type Op struct {
	T              int
	Kind           string // begin read scan write insert commit abort
	Key            string
	Val            int
	Delta          int
	HasDelta       bool
	Pred           *Pred
	Lockable       bool
	ExclusiveRange bool
	GuardMin       int
	HasGuardMin    bool
	GuardMax       int
	HasGuardMax    bool
	Tag            string
}

type TraceEntry struct {
	T    int
	Kind string // r w c blk
	Txt  string
}

type lockInfo struct {
	trx int
	x   bool
}

type Engine struct {
	Level     string
	rows      map[string][]Version // 版本链，新 → 旧
	keys      []string             // 保持稳定的遍历顺序
	committed map[int]bool
	active    map[int]bool
	xlock     map[string]int
	rlock     []struct {
		trx int
		tag string
	}
	tx       map[int]*Tx
	nextID   int
	Trace    []TraceEntry
	Deadlock bool
	blkSeen  map[int]bool
}

func NewEngine(level string, init map[string]int) *Engine {
	e := &Engine{Level: level, rows: map[string][]Version{},
		committed: map[int]bool{0: true}, active: map[int]bool{},
		xlock: map[string]int{}, tx: map[int]*Tx{}, nextID: 101,
		blkSeen: map[int]bool{}}
	ks := make([]string, 0, len(init))
	for k := range init {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		e.rows[k] = []Version{{Trx: 0, Val: init[k]}}
		e.keys = append(e.keys, k)
	}
	return e
}

func (e *Engine) mkRV(x *Tx) *ReadView {
	mids := []int{}
	for id := range e.active {
		if id != x.ID {
			mids = append(mids, id)
		}
	}
	sort.Ints(mids)
	min := e.nextID
	if len(mids) > 0 {
		min = mids[0]
	}
	return &ReadView{Mids: mids, Min: min, Max: e.nextID, Creator: x.ID}
}

// 四步可见性判定 —— 和 InnoDB 一模一样，顺序也一样。
func (e *Engine) visible(trx int, rv *ReadView) bool {
	if trx == rv.Creator {
		return true // 我自己改的，当然看得见
	}
	if trx < rv.Min {
		return true // 建 ReadView 之前就已提交
	}
	if trx >= rv.Max {
		return false // 在我之后才开始的事务
	}
	for _, m := range rv.Mids {
		if m == trx {
			return false // 我建 ReadView 时它还没提交
		}
	}
	return true
}

// 快照读：沿版本链从新到旧，第一个可见的就是答案。
func (e *Engine) snapRead(key string, rv *ReadView) (int, bool) {
	for _, v := range e.rows[key] {
		if e.visible(v.Trx, rv) {
			return v.Val, true
		}
	}
	return 0, false
}

// 当前读：最新的【已提交】版本。MVCC 在这里帮不上忙，所以才需要锁。
func (e *Engine) curRead(key string) (int, bool) {
	for _, v := range e.rows[key] {
		if e.committed[v.Trx] {
			return v.Val, true
		}
	}
	return 0, false
}

func (e *Engine) dirtyRead(key string) (int, bool) {
	if ch := e.rows[key]; len(ch) > 0 {
		return ch[0].Val, true
	}
	return 0, false
}

func (e *Engine) readFor(x *Tx, key string, lockable bool) (int, bool) {
	switch {
	case e.Level == RU:
		return e.dirtyRead(key)
	case lockable && (e.Level == RRLOCK || e.Level == SER):
		return e.curRead(key)
	case e.Level == RC:
		return e.snapRead(key, e.mkRV(x)) // RC：每条语句新建一个 ReadView
	}
	if x.RV == nil {
		x.RV = e.mkRV(x) // RR：整个事务只建一次，之后复用 —— 和 RC 的唯一区别
	}
	return e.snapRead(key, x.RV)
}

func (e *Engine) matching(x *Tx, p *Pred, lockable bool) []string {
	out := []string{}
	for _, k := range e.keys {
		if v, ok := e.readFor(x, k, lockable); ok && p.Test(k, v) {
			out = append(out, k)
		}
	}
	return out
}

func (e *Engine) lockBusy(key string, id int) bool {
	o, ok := e.xlock[key]
	return ok && o != id && e.active[o]
}
func (e *Engine) rangeBusy(tag string, id int) bool {
	for _, r := range e.rlock {
		if r.tag == tag && r.trx != id && e.active[r.trx] {
			return true
		}
	}
	return false
}
func (e *Engine) put(key string, v Version) {
	if _, ok := e.rows[key]; !ok {
		e.keys = append(e.keys, key)
	}
	e.rows[key] = append([]Version{v}, e.rows[key]...)
}

const locked = "blocked"

// step 执行一步，返回 "ok" 或 "blocked"（被锁挡住，需要排队重试）。
func (e *Engine) step(s *Op) string {
	x := e.tx[s.T]
	if s.Kind == "begin" {
		x = &Tx{T: s.T, ID: e.nextID}
		e.nextID++
		e.tx[s.T] = x
		e.active[x.ID] = true
		e.Trace = append(e.Trace, TraceEntry{s.T, "c", "BEGIN"})
		return "ok"
	}
	if x == nil {
		return "ok"
	}
	lockMode := e.Level == RRLOCK || e.Level == SER
	switch s.Kind {
	case "read":
		if s.Lockable && lockMode {
			if e.lockBusy(s.Key, x.ID) {
				return locked
			}
			e.xlock[s.Key] = x.ID
		}
		v, _ := e.readFor(x, s.Key, s.Lockable)
		x.Last = v
		x.Reads = append(x.Reads, v)
		q := "SELECT " + s.Key
		if s.Lockable && lockMode {
			q += " FOR UPDATE"
		}
		e.Trace = append(e.Trace, TraceEntry{s.T, "r", q + " → " + itoa(v)})
		return "ok"

	case "scan":
		keys := e.matching(x, s.Pred, s.Lockable)
		if s.Lockable && lockMode {
			for _, k := range keys {
				if e.lockBusy(k, x.ID) {
					return locked
				}
			}
			if s.ExclusiveRange && e.rangeBusy(s.Pred.Tag, x.ID) {
				return locked
			}
			for _, k := range keys {
				e.xlock[k] = x.ID
			}
			e.rlock = append(e.rlock, struct {
				trx int
				tag string
			}{x.ID, s.Pred.Tag})
		}
		x.Last = len(keys)
		x.Reads = append(x.Reads, len(keys))
		q := "SELECT count(*) WHERE " + s.Pred.Label
		if s.Lockable && lockMode {
			q += " FOR UPDATE"
		}
		e.Trace = append(e.Trace, TraceEntry{s.T, "r", q + " → " + itoa(len(keys))})
		return "ok"

	case "write":
		// guardMin 模拟应用代码里那个 if：读到的结果是决策的前提
		if s.HasGuardMin && !(x.Last >= s.GuardMin) {
			e.Trace = append(e.Trace, TraceEntry{s.T, "c",
				"if (上次读到 " + itoa(x.Last) + " >= " + itoa(s.GuardMin) + ") 不成立 → 放弃写入"})
			return "ok"
		}
		if e.Level != RU && e.lockBusy(s.Key, x.ID) {
			return locked
		}
		e.xlock[s.Key] = x.ID
		val := s.Val
		if s.HasDelta {
			val = x.Last + s.Delta
		}
		e.put(s.Key, Version{Trx: x.ID, Val: val})
		e.Trace = append(e.Trace, TraceEntry{s.T, "w", "UPDATE " + s.Key + " = " + itoa(val)})
		return "ok"

	case "insert":
		if s.HasGuardMax && !(x.Last <= s.GuardMax) {
			e.Trace = append(e.Trace, TraceEntry{s.T, "c",
				"if (上次读到 " + itoa(x.Last) + " <= " + itoa(s.GuardMax) + ") 不成立 → 放弃插入"})
			return "ok"
		}
		if lockMode && s.Tag != "" && e.rangeBusy(s.Tag, x.ID) {
			return locked // 间隙锁：锁的正是「还不存在的行」
		}
		e.xlock[s.Key] = x.ID
		e.put(s.Key, Version{Trx: x.ID, Val: s.Val})
		e.Trace = append(e.Trace, TraceEntry{s.T, "w", "INSERT " + s.Key + " = " + itoa(s.Val)})
		return "ok"

	case "commit":
		e.committed[x.ID] = true
		delete(e.active, x.ID)
		x.Done = true
		e.Trace = append(e.Trace, TraceEntry{s.T, "c", "COMMIT"})
		return "ok"

	case "abort":
		delete(e.active, x.ID)
		x.Done = true
		e.dropVersions(x.ID)
		e.Trace = append(e.Trace, TraceEntry{s.T, "c", "ROLLBACK"})
		return "ok"
	}
	return "ok"
}

func (e *Engine) dropVersions(id int) {
	for k, ch := range e.rows {
		out := ch[:0]
		for _, v := range ch {
			if v.Trx != id {
				out = append(out, v)
			}
		}
		e.rows[k] = out
	}
}

// Run 按给定时序执行。某一步被锁挡住就让它排队，等持锁方提交后再放行；
// 所有事务都卡住 = 死锁 ⇒ 中止一个（真实引擎也是这么干的）。
func (e *Engine) Run(steps []*Op) {
	q := append([]*Op{}, steps...)
	waiting := map[int]bool{}
	for guard := 0; len(q) > 0 && guard < 4000; guard++ {
		moved := false
		for i := 0; i < len(q); i++ {
			s := q[i]
			if waiting[s.T] {
				continue
			}
			if e.step(s) == "ok" {
				q = append(q[:i], q[i+1:]...)
				moved = true
				break
			}
			waiting[s.T] = true
			if !e.blkSeen[s.T*1000+i] {
				e.blkSeen[s.T*1000+i] = true
				e.Trace = append(e.Trace, TraceEntry{s.T, "blk", "（被锁阻塞，等待前一个事务提交）"})
			}
		}
		if !moved {
			stuck := []int{}
			for t, w := range waiting {
				if w {
					stuck = append(stuck, t)
				}
			}
			if len(stuck) == 0 {
				break
			}
			sort.Ints(stuck)
			e.Deadlock = true
			victim := stuck[len(stuck)-1]
			if x := e.tx[victim]; x != nil {
				delete(e.active, x.ID)
				e.dropVersions(x.ID)
				rl := e.rlock[:0]
				for _, r := range e.rlock {
					if r.trx != x.ID {
						rl = append(rl, r)
					}
				}
				e.rlock = rl
				e.Trace = append(e.Trace, TraceEntry{victim, "blk", "⚡ 死锁，本事务被选为牺牲者并回滚"})
			}
			out := q[:0]
			for _, s := range q {
				if s.T != victim {
					out = append(out, s)
				}
			}
			q = out
		}
		for t := range waiting {
			waiting[t] = false
		}
	}
}

func (e *Engine) FinalVal(key string) (int, bool) { return e.curRead(key) }
func (e *Engine) CountWhere(p *Pred) int {
	n := 0
	for _, k := range e.keys {
		if v, ok := e.curRead(k); ok && p.Test(k, v) {
			n++
		}
	}
	return n
}
func (e *Engine) Blocked() bool {
	for _, t := range e.Trace {
		if t.Kind == "blk" {
			return true
		}
	}
	return false
}

// Serialize 把同一个事务的步骤聚到一起：一个跑完再跑下一个。
// 可串行化的定义就是「等价于某个串行顺序」，那最直接的实现就是真的串行。
func Serialize(steps []*Op) []*Op {
	order := []int{}
	buckets := map[int][]*Op{}
	for _, s := range steps {
		if _, ok := buckets[s.T]; !ok {
			order = append(order, s.T)
		}
		buckets[s.T] = append(buckets[s.T], s)
	}
	out := []*Op{}
	for _, t := range order {
		out = append(out, buckets[t]...)
	}
	return out
}
