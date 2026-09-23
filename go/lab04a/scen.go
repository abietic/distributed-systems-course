package main

// 七种并发异常。前四种在 ANSI SQL-92 的清单里，后三种不在——
// 而后三种才是线上真正出事的那几种。
var (
	POnCall = &Pred{"oncall", "on_call = true", func(k string, v int) bool { return len(k) > 4 && k[:4] == "doc:" && v == 1 }}
	PAdult  = &Pred{"adult", "age > 18", func(k string, v int) bool { return len(k) > 2 && k[:2] == "p:" && v > 18 }}
	PRoom   = &Pred{"roomA", "room='A' AND slot=10", func(k string, v int) bool { return len(k) > 3 && k[:3] == "bk:" && v == 1 }}
)

type Scenario struct {
	ID, Name, En string
	ANSI         bool
	Init         map[string]int
	Steps        func() []*Op
	Check        func(e *Engine) (bad bool, detail string)
}

func op(t int, kind string) *Op { return &Op{T: t, Kind: kind} }

var Scenarios = []Scenario{
	{ID: "dw", Name: "脏写", En: "dirty write", ANSI: true,
		Init: map[string]int{"x": 0, "y": 0},
		Steps: func() []*Op {
			return []*Op{op(1, "begin"), op(2, "begin"),
				{T: 1, Kind: "write", Key: "x", Val: 1}, {T: 2, Kind: "write", Key: "x", Val: 2},
				{T: 2, Kind: "write", Key: "y", Val: 2}, {T: 1, Kind: "write", Key: "y", Val: 1},
				op(1, "commit"), op(2, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			xv, _ := e.FinalVal("x")
			yv, _ := e.FinalVal("y")
			return xv != yv, "最终 x=" + itoa(xv) + "，y=" + itoa(yv)
		}},

	{ID: "dr", Name: "脏读", En: "dirty read", ANSI: true,
		Init: map[string]int{"x": 1},
		Steps: func() []*Op {
			return []*Op{op(1, "begin"), op(2, "begin"),
				{T: 1, Kind: "write", Key: "x", Val: 2}, {T: 2, Kind: "read", Key: "x"},
				op(1, "abort"), op(2, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			got := -1
			if x := e.tx[2]; x != nil {
				got = x.Last
			}
			return got == 2, "T2 读到 x=" + itoa(got) + "，而 T1 最终回滚了"
		}},

	{ID: "nr", Name: "不可重复读", En: "non-repeatable read", ANSI: true,
		Init: map[string]int{"x": 1},
		Steps: func() []*Op {
			return []*Op{op(1, "begin"), {T: 1, Kind: "read", Key: "x"},
				op(2, "begin"), {T: 2, Kind: "write", Key: "x", Val: 2}, op(2, "commit"),
				{T: 1, Kind: "read", Key: "x"}, op(1, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			r := reads(e, 1)
			return len(r) > 1 && r[0] != r[1], "T1 两次读到 " + joinInts(r, " 和 ")
		}},

	{ID: "ph", Name: "幻读", En: "phantom read", ANSI: true,
		Init: map[string]int{"p:1": 20, "p:2": 30, "p:3": 10},
		Steps: func() []*Op {
			sc := func(t int) *Op {
				return &Op{T: t, Kind: "scan", Pred: PAdult, Lockable: true, ExclusiveRange: true}
			}
			return []*Op{op(1, "begin"), sc(1),
				op(2, "begin"), {T: 2, Kind: "insert", Key: "p:4", Val: 25, Tag: "adult"}, op(2, "commit"),
				sc(1), op(1, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			r := reads(e, 1)
			return len(r) > 1 && r[0] != r[1], "T1 两次范围查询的行数：" + joinInts(r, " 和 ")
		}},

	{ID: "lu", Name: "丢失更新", En: "lost update", ANSI: false,
		Init: map[string]int{"x": 100},
		Steps: func() []*Op {
			return []*Op{op(1, "begin"), op(2, "begin"),
				{T: 1, Kind: "read", Key: "x", Lockable: true}, {T: 2, Kind: "read", Key: "x", Lockable: true},
				{T: 1, Kind: "write", Key: "x", Delta: 10, HasDelta: true},
				{T: 2, Kind: "write", Key: "x", Delta: 10, HasDelta: true},
				op(1, "commit"), op(2, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			v, _ := e.FinalVal("x")
			return v != 120, "两次各加 10，期望 120，实际 " + itoa(v)
		}},

	{ID: "ws", Name: "写偏斜", En: "write skew", ANSI: false,
		Init: map[string]int{"doc:alice": 1, "doc:bob": 1},
		Steps: func() []*Op {
			return []*Op{op(1, "begin"), op(2, "begin"),
				{T: 1, Kind: "scan", Pred: POnCall, Lockable: true},
				{T: 2, Kind: "scan", Pred: POnCall, Lockable: true},
				{T: 1, Kind: "write", Key: "doc:alice", Val: 0, GuardMin: 2, HasGuardMin: true},
				{T: 2, Kind: "write", Key: "doc:bob", Val: 0, GuardMin: 2, HasGuardMin: true},
				op(1, "commit"), op(2, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			n := e.CountWhere(POnCall)
			return n < 1, "最终在岗医生数 " + itoa(n) + "（不变量要求 >= 1）"
		}},

	{ID: "pws", Name: "写偏斜（插入型）", En: "phantom write skew", ANSI: false,
		Init: map[string]int{"bk:0": 0},
		Steps: func() []*Op {
			sc := func(t int) *Op {
				return &Op{T: t, Kind: "scan", Pred: PRoom, Lockable: true, ExclusiveRange: true}
			}
			return []*Op{op(1, "begin"), op(2, "begin"), sc(1), sc(2),
				{T: 1, Kind: "insert", Key: "bk:1", Val: 1, GuardMax: 0, HasGuardMax: true, Tag: "roomA"},
				{T: 2, Kind: "insert", Key: "bk:2", Val: 1, GuardMax: 0, HasGuardMax: true, Tag: "roomA"},
				op(1, "commit"), op(2, "commit")}
		},
		Check: func(e *Engine) (bool, string) {
			n := e.CountWhere(PRoom)
			return n > 1, "最终这个时段的预订数 " + itoa(n) + "（不变量要求 <= 1）"
		}},
}

func reads(e *Engine, t int) []int {
	if x := e.tx[t]; x != nil {
		return x.Reads
	}
	return nil
}

// RunScenario 在给定隔离配置下跑一遍某个异常场景。
func RunScenario(sc Scenario, level string) (bad bool, detail string, e *Engine) {
	init := map[string]int{}
	for k, v := range sc.Init {
		init[k] = v
	}
	e = NewEngine(level, init)
	steps := sc.Steps()
	if level == SER {
		steps = Serialize(steps)
	}
	e.Run(steps)
	bad, detail = sc.Check(e)
	return
}
