/*
 * Lab 4-A · 单机事务：隔离级别、MVCC 与写偏斜（Java 版）
 *
 * 运行：  cd java/lab04a && java Lab04A.java
 * 调参：  java Lab04A.java -trials 500 -doctors 3 -txns 12 -ops 4
 *
 * 需要 JDK 17+。配套课件：courseware/ch04a-transactions.html
 *
 * 里面那个迷你事务引擎（MVCC 版本链 + ReadView + 行锁 + 范围锁 + 串行执行）
 * 存在的理由只有一个：让「哪个隔离级别挡得住哪种异常」这句话
 * 从【背下来的表格】变成【跑出来的结果】。
 *
 * 输出与 Go 版逐字对应（连随机数都用同一个 xorshift32），可以直接 diff 对照。
 */
import java.util.*;
import java.util.function.BiPredicate;

public class Lab04A {

    static int TRIALS=200, DOCTORS=2, LEAVERS=2, TXNS=12, OPS=4;
    static long SEED=42;

    public static void main(String[] args){
        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),
                true, java.nio.charset.StandardCharsets.UTF_8));
        parseArgs(args);
        System.out.printf("%n配置：医生 %d 人 ｜ 同时请假 %d 人 ｜ 试验 %d 次 ｜ 并发事务 %d × %d 操作 ｜ 种子 %d%n",
                DOCTORS, LEAVERS, TRIALS, TXNS, OPS, SEED);
        lab1(); lab2(); lab3(); lab4(); epilogue();
    }

    // ══════════════════════════════════════════════════════════════════
    // 随机数：和 Go 版同一个 xorshift32，保证两边输出逐字节一致
    // ══════════════════════════════════════════════════════════════════
    static final class Rnd {
        int s;
        Rnd(long seed){ s=(int)seed; if(s==0) s=1; }
        int next(){ s^=s<<13; s^=s>>>17; s^=s<<5; return s; }
        double flt(){ return (next()&0xFFFFFFFFL)/4294967296.0; }
        int intn(int n){ return (int)(flt()*n); }
    }

    // ══════════════════════════════════════════════════════════════════
    // 迷你事务引擎
    // ══════════════════════════════════════════════════════════════════
    static final String RU="RU", RC="RC", RR="RR", RRLOCK="RRLOCK", SER="SER";
    static final String[] LEVELS={RU,RC,RR,RRLOCK,SER};
    static final Map<String,String> LVNAME=Map.of(
        RU,"READ UNCOMMITTED", RC,"READ COMMITTED", RR,"REPEATABLE READ",
        RRLOCK,"RR + FOR UPDATE", SER,"SERIALIZABLE");

    record Version(int trx,int val){}
    record Pred(String tag,String label,BiPredicate<String,Integer> test){}
    /** ReadView 的四个字段，和 InnoDB 一模一样。 */
    static final class ReadView { List<Integer> mids; int min,max,creator; }
    static final class Tx { int t,id; ReadView rv; int last; List<Integer> reads=new ArrayList<>(); boolean done; }
    static final class Op {
        int t; String kind, key="", tag=""; int val, delta;
        boolean hasDelta, lockable, exclusiveRange, hasGuardMin, hasGuardMax;
        int guardMin, guardMax; Pred pred;
        Op(int t,String kind){ this.t=t; this.kind=kind; }
        Op k(String x){ key=x; return this; }
        Op v(int x){ val=x; return this; }
        Op d(int x){ delta=x; hasDelta=true; return this; }
        Op p(Pred x){ pred=x; return this; }
        Op lock(){ lockable=true; return this; }
        Op xr(){ exclusiveRange=true; return this; }
        Op gmin(int x){ guardMin=x; hasGuardMin=true; return this; }
        Op gmax(int x){ guardMax=x; hasGuardMax=true; return this; }
        Op tg(String x){ tag=x; return this; }
    }
    record Trace(int t,String kind,String txt){}
    record Lock(int trx,boolean x){}
    record RangeLock(int trx,String tag){}

    static final class Engine {
        final String level;
        final Map<String,List<Version>> rows=new HashMap<>();
        final List<String> keys=new ArrayList<>();
        final Set<Integer> committed=new HashSet<>(Set.of(0));
        final Set<Integer> active=new HashSet<>();
        final Map<String,Integer> xlock=new HashMap<>();
        final List<RangeLock> rlock=new ArrayList<>();
        final Map<Integer,Tx> tx=new HashMap<>();
        int nextId=101;
        final List<Trace> trace=new ArrayList<>();
        boolean deadlock=false;
        final Set<Integer> blkSeen=new HashSet<>();

        Engine(String level, Map<String,Integer> init){
            this.level=level;
            List<String> ks=new ArrayList<>(init.keySet()); Collections.sort(ks);
            for(String k:ks){ rows.put(k,new ArrayList<>(List.of(new Version(0,init.get(k))))); keys.add(k); }
        }
        ReadView mkRV(Tx x){
            List<Integer> mids=new ArrayList<>();
            for(int id:active) if(id!=x.id) mids.add(id);
            Collections.sort(mids);
            ReadView rv=new ReadView(); rv.mids=mids; rv.max=nextId; rv.creator=x.id;
            rv.min = mids.isEmpty()? nextId : mids.get(0);
            return rv;
        }
        /** 四步可见性判定 —— 和 InnoDB 一模一样，顺序也一样。 */
        boolean visible(int trx, ReadView rv){
            if(trx==rv.creator) return true;    // 我自己改的，当然看得见
            if(trx<rv.min) return true;         // 建 ReadView 之前就已提交
            if(trx>=rv.max) return false;       // 在我之后才开始的事务
            if(rv.mids.contains(trx)) return false; // 我建 ReadView 时它还没提交
            return true;
        }
        /** 快照读：沿版本链从新到旧，第一个可见的就是答案。 */
        Integer snapRead(String key, ReadView rv){
            for(Version v: rows.getOrDefault(key,List.of())) if(visible(v.trx(),rv)) return v.val();
            return null;
        }
        /** 当前读：最新的【已提交】版本。MVCC 在这里帮不上忙，所以才需要锁。 */
        Integer curRead(String key){
            for(Version v: rows.getOrDefault(key,List.of())) if(committed.contains(v.trx())) return v.val();
            return null;
        }
        Integer dirtyRead(String key){
            List<Version> ch=rows.getOrDefault(key,List.of());
            return ch.isEmpty()? null : ch.get(0).val();
        }
        Integer readFor(Tx x,String key,boolean lockable){
            if(level.equals(RU)) return dirtyRead(key);
            if(lockable&&(level.equals(RRLOCK)||level.equals(SER))) return curRead(key);
            if(level.equals(RC)) return snapRead(key,mkRV(x));   // RC：每条语句新建一个 ReadView
            if(x.rv==null) x.rv=mkRV(x);   // RR：整个事务只建一次，之后复用 —— 和 RC 的唯一区别
            return snapRead(key,x.rv);
        }
        List<String> matching(Tx x,Pred p,boolean lockable){
            List<String> out=new ArrayList<>();
            for(String k:keys){ Integer v=readFor(x,k,lockable); if(v!=null&&p.test().test(k,v)) out.add(k); }
            return out;
        }
        boolean lockBusy(String key,int id){
            Integer o=xlock.get(key); return o!=null&&o!=id&&active.contains(o);
        }
        boolean rangeBusy(String tag,int id){
            for(RangeLock r:rlock) if(r.tag().equals(tag)&&r.trx()!=id&&active.contains(r.trx())) return true;
            return false;
        }
        void put(String key,Version v){
            if(!rows.containsKey(key)){ keys.add(key); rows.put(key,new ArrayList<>()); }
            rows.get(key).add(0,v);
        }
        void dropVersions(int id){
            for(List<Version> ch: rows.values()) ch.removeIf(v->v.trx()==id);
        }

        static final String LOCKED="blocked";
        String step(Op s,int qi){
            Tx x=tx.get(s.t);
            if(s.kind.equals("begin")){
                x=new Tx(); x.t=s.t; x.id=nextId++;
                tx.put(s.t,x); active.add(x.id);
                trace.add(new Trace(s.t,"c","BEGIN")); return "ok";
            }
            if(x==null) return "ok";
            boolean lockMode = level.equals(RRLOCK)||level.equals(SER);
            switch(s.kind){
                case "read": {
                    if(s.lockable&&lockMode){
                        if(lockBusy(s.key,x.id)) return LOCKED;
                        xlock.put(s.key,x.id);
                    }
                    Integer v=readFor(x,s.key,s.lockable); int vv=(v==null?0:v);
                    x.last=vv; x.reads.add(vv);
                    String q="SELECT "+s.key+(s.lockable&&lockMode?" FOR UPDATE":"");
                    trace.add(new Trace(s.t,"r",q+" → "+vv)); return "ok";
                }
                case "scan": {
                    List<String> ks=matching(x,s.pred,s.lockable);
                    if(s.lockable&&lockMode){
                        for(String k:ks) if(lockBusy(k,x.id)) return LOCKED;
                        if(s.exclusiveRange&&rangeBusy(s.pred.tag(),x.id)) return LOCKED;
                        for(String k:ks) xlock.put(k,x.id);
                        rlock.add(new RangeLock(x.id,s.pred.tag()));
                    }
                    x.last=ks.size(); x.reads.add(ks.size());
                    String q="SELECT count(*) WHERE "+s.pred.label()+(s.lockable&&lockMode?" FOR UPDATE":"");
                    trace.add(new Trace(s.t,"r",q+" → "+ks.size())); return "ok";
                }
                case "write": {
                    // guardMin 模拟应用代码里那个 if：读到的结果是决策的前提
                    if(s.hasGuardMin&&!(x.last>=s.guardMin)){
                        trace.add(new Trace(s.t,"c","if (上次读到 "+x.last+" >= "+s.guardMin+") 不成立 → 放弃写入"));
                        return "ok";
                    }
                    if(!level.equals(RU)&&lockBusy(s.key,x.id)) return LOCKED;
                    xlock.put(s.key,x.id);
                    int val = s.hasDelta ? x.last+s.delta : s.val;
                    put(s.key,new Version(x.id,val));
                    trace.add(new Trace(s.t,"w","UPDATE "+s.key+" = "+val)); return "ok";
                }
                case "insert": {
                    if(s.hasGuardMax&&!(x.last<=s.guardMax)){
                        trace.add(new Trace(s.t,"c","if (上次读到 "+x.last+" <= "+s.guardMax+") 不成立 → 放弃插入"));
                        return "ok";
                    }
                    // 间隙锁：锁的正是「还不存在的行」
                    if(lockMode&&!s.tag.isEmpty()&&rangeBusy(s.tag,x.id)) return LOCKED;
                    xlock.put(s.key,x.id);
                    put(s.key,new Version(x.id,s.val));
                    trace.add(new Trace(s.t,"w","INSERT "+s.key+" = "+s.val)); return "ok";
                }
                case "commit":
                    committed.add(x.id); active.remove(x.id); x.done=true;
                    trace.add(new Trace(s.t,"c","COMMIT")); return "ok";
                case "abort":
                    active.remove(x.id); x.done=true; dropVersions(x.id);
                    trace.add(new Trace(s.t,"c","ROLLBACK")); return "ok";
            }
            return "ok";
        }
        /** 按给定时序执行。被锁挡住就排队，等持锁方提交后再放行；全卡住 = 死锁 ⇒ 中止一个。 */
        void run(List<Op> steps){
            List<Op> q=new ArrayList<>(steps);
            Map<Integer,Boolean> waiting=new HashMap<>();
            for(int guard=0; !q.isEmpty()&&guard<4000; guard++){
                boolean moved=false;
                for(int i=0;i<q.size();i++){
                    Op s=q.get(i);
                    if(Boolean.TRUE.equals(waiting.get(s.t))) continue;
                    if(step(s,i).equals("ok")){ q.remove(i); moved=true; break; }
                    waiting.put(s.t,true);
                    if(blkSeen.add(s.t*1000+i))
                        trace.add(new Trace(s.t,"blk","（被锁阻塞，等待前一个事务提交）"));
                }
                if(!moved){
                    List<Integer> stuck=new ArrayList<>();
                    for(var e: waiting.entrySet()) if(e.getValue()) stuck.add(e.getKey());
                    if(stuck.isEmpty()) break;
                    Collections.sort(stuck);
                    deadlock=true;
                    int victim=stuck.get(stuck.size()-1);
                    Tx x=tx.get(victim);
                    if(x!=null){
                        active.remove(x.id); dropVersions(x.id);
                        rlock.removeIf(r->r.trx()==x.id);
                        trace.add(new Trace(victim,"blk","⚡ 死锁，本事务被选为牺牲者并回滚"));
                    }
                    final int vv=victim; q.removeIf(s->s.t==vv);
                }
                waiting.replaceAll((k,v)->false);
            }
        }
        Integer finalVal(String key){ return curRead(key); }
        int countWhere(Pred p){
            int n=0;
            for(String k:keys){ Integer v=curRead(k); if(v!=null&&p.test().test(k,v)) n++; }
            return n;
        }
        boolean blocked(){ for(Trace t:trace) if(t.kind().equals("blk")) return true; return false; }
    }
    /** 可串行化的定义就是「等价于某个串行顺序」，那最直接的实现就是真的串行。 */
    static List<Op> serialize(List<Op> steps){
        List<Integer> order=new ArrayList<>(); Map<Integer,List<Op>> b=new LinkedHashMap<>();
        for(Op s:steps){ if(!b.containsKey(s.t)){ b.put(s.t,new ArrayList<>()); order.add(s.t);} b.get(s.t).add(s); }
        List<Op> out=new ArrayList<>(); for(int t:order) out.addAll(b.get(t)); return out;
    }

    // ══════════════════════════════════════════════════════════════════
    // 七种并发异常。前四种在 ANSI SQL-92 的清单里，后三种不在——
    // 而后三种才是线上真正出事的那几种。
    // ══════════════════════════════════════════════════════════════════
    static final Pred P_ONCALL=new Pred("oncall","on_call = true",(k,v)->k.startsWith("doc:")&&v==1);
    static final Pred P_ADULT =new Pred("adult","age > 18",       (k,v)->k.startsWith("p:")&&v>18);
    static final Pred P_ROOM  =new Pred("roomA","room='A' AND slot=10",(k,v)->k.startsWith("bk:")&&v==1);

    interface Check { Object[] run(Engine e); }   // [Boolean bad, String detail]
    record Scen(String id,String name,String en,boolean ansi,
                Map<String,Integer> init, java.util.function.Supplier<List<Op>> steps, Check check){}

    static Op o(int t,String kind){ return new Op(t,kind); }

    static final List<Scen> SCEN=List.of(
      new Scen("dw","脏写","dirty write",true, Map.of("x",0,"y",0),
        ()->List.of(o(1,"begin"),o(2,"begin"),
          o(1,"write").k("x").v(1), o(2,"write").k("x").v(2),
          o(2,"write").k("y").v(2), o(1,"write").k("y").v(1),
          o(1,"commit"),o(2,"commit")),
        e->{ Integer x=e.finalVal("x"), y=e.finalVal("y");
             return new Object[]{!Objects.equals(x,y), "最终 x="+x+"，y="+y}; }),

      new Scen("dr","脏读","dirty read",true, Map.of("x",1),
        ()->List.of(o(1,"begin"),o(2,"begin"),
          o(1,"write").k("x").v(2), o(2,"read").k("x"),
          o(1,"abort"),o(2,"commit")),
        e->{ int got=e.tx.containsKey(2)?e.tx.get(2).last:-1;
             return new Object[]{got==2,"T2 读到 x="+got+"，而 T1 最终回滚了"}; }),

      new Scen("nr","不可重复读","non-repeatable read",true, Map.of("x",1),
        ()->List.of(o(1,"begin"), o(1,"read").k("x"),
          o(2,"begin"), o(2,"write").k("x").v(2), o(2,"commit"),
          o(1,"read").k("x"), o(1,"commit")),
        e->{ List<Integer> r=reads(e,1);
             return new Object[]{r.size()>1&&!r.get(0).equals(r.get(1)),
               "T1 两次读到 "+join(r," 和 ")}; }),

      new Scen("ph","幻读","phantom read",true, new LinkedHashMap<>(Map.of("p:1",20,"p:2",30,"p:3",10)),
        ()->List.of(o(1,"begin"), o(1,"scan").p(P_ADULT).lock().xr(),
          o(2,"begin"), o(2,"insert").k("p:4").v(25).tg("adult"), o(2,"commit"),
          o(1,"scan").p(P_ADULT).lock().xr(), o(1,"commit")),
        e->{ List<Integer> r=reads(e,1);
             return new Object[]{r.size()>1&&!r.get(0).equals(r.get(1)),
               "T1 两次范围查询的行数："+join(r," 和 ")}; }),

      new Scen("lu","丢失更新","lost update",false, Map.of("x",100),
        ()->List.of(o(1,"begin"),o(2,"begin"),
          o(1,"read").k("x").lock(), o(2,"read").k("x").lock(),
          o(1,"write").k("x").d(10), o(2,"write").k("x").d(10),
          o(1,"commit"),o(2,"commit")),
        e->{ Integer v=e.finalVal("x");
             return new Object[]{v==null||v!=120,"两次各加 10，期望 120，实际 "+v}; }),

      new Scen("ws","写偏斜","write skew",false, Map.of("doc:alice",1,"doc:bob",1),
        ()->List.of(o(1,"begin"),o(2,"begin"),
          o(1,"scan").p(P_ONCALL).lock(), o(2,"scan").p(P_ONCALL).lock(),
          o(1,"write").k("doc:alice").v(0).gmin(2),
          o(2,"write").k("doc:bob").v(0).gmin(2),
          o(1,"commit"),o(2,"commit")),
        e->{ int n=e.countWhere(P_ONCALL);
             return new Object[]{n<1,"最终在岗医生数 "+n+"（不变量要求 >= 1）"}; }),

      new Scen("pws","写偏斜（插入型）","phantom write skew",false, Map.of("bk:0",0),
        ()->List.of(o(1,"begin"),o(2,"begin"),
          o(1,"scan").p(P_ROOM).lock().xr(), o(2,"scan").p(P_ROOM).lock().xr(),
          o(1,"insert").k("bk:1").v(1).gmax(0).tg("roomA"),
          o(2,"insert").k("bk:2").v(1).gmax(0).tg("roomA"),
          o(1,"commit"),o(2,"commit")),
        e->{ int n=e.countWhere(P_ROOM);
             return new Object[]{n>1,"最终这个时段的预订数 "+n+"（不变量要求 <= 1）"}; })
    );

    static List<Integer> reads(Engine e,int t){ return e.tx.containsKey(t)?e.tx.get(t).reads:List.of(); }
    static String join(List<Integer> xs,String sep){
        StringBuilder sb=new StringBuilder();
        for(int i=0;i<xs.size();i++){ if(i>0) sb.append(sep); sb.append(xs.get(i)); }
        return sb.toString();
    }
    /** 在给定隔离配置下跑一遍某个异常场景。 */
    static Object[] runScenario(Scen sc,String level){
        Engine e=new Engine(level,new HashMap<>(sc.init()));
        List<Op> steps=sc.steps().get();
        e.run(level.equals(SER)?serialize(steps):steps);
        Object[] r=sc.check().run(e);
        return new Object[]{r[0],r[1],e};
    }

    // ══════════════════════════════════════════════════════════════════
    // Lab 4A-1
    // ══════════════════════════════════════════════════════════════════
    static void lab1(){
        head("4A-1","并发异常 × 隔离级别矩阵","这张表不是抄来的，每一格都是迷你事务引擎真的跑了一遍");
        int[] w={22,4,12,12,12,14,12}; String al="LLRRRRR";
        List<String> hdr=new ArrayList<>(List.of("异常",""));
        for(String L:LEVELS) hdr.add(shortName(L));
        tableHead(hdr.toArray(new String[0]),w,al);
        for(Scen sc:SCEN){
            List<String> row=new ArrayList<>(List.of(sc.name(), sc.ansi()?"":"*"));
            for(String L:LEVELS){
                Object[] r=runScenario(sc,L);
                String cell=((Boolean)r[0])?"异常发生":"挡住";
                if(((Engine)r[2]).deadlock) cell+="(死锁)";
                row.add(cell);
            }
            tableRow(row.toArray(new String[0]),w,al);
        }
        System.out.println("\n    * = 不在 ANSI SQL-92 的异常清单里 —— 而它们恰恰是线上真正出事的那几种。");
        for(String id: new String[]{"lu","ws"})
          for(Scen sc:SCEN){
            if(!sc.id().equals(id)) continue;
            for(String L: new String[]{RR,RRLOCK}){
                Object[] r=runScenario(sc,L); Engine e=(Engine)r[2];
                System.out.printf("%n    ── %s × %s ──────────────────────────────%n", sc.name(), LVNAME.get(L));
                for(Trace t:e.trace){
                    String pad = t.t()==2 ? " ".repeat(44) : "";
                    System.out.printf("      T%d %s%s%n", t.t(), pad, t.txt());
                }
                System.out.printf("      ⇒ %s　%s%n", r[1], ((Boolean)r[0])?"★ 不变量被破坏":"正确");
            }
          }
        System.out.print(blk("""

    结论
      · 看「丢失更新」和「写偏斜」两行：RR 那一列是红的。
        把 MySQL 从 RC 调到 RR，你买到的是「不可重复读」和「幻读」，
        【没有】买到「丢失更新」和「写偏斜」——而后两者才是业务不变量被破坏的那一类。
      · RR 和 RR+FOR UPDATE 的全部差别，是把隐式的读变成了显式的锁。
        数据库看不见你代码里那个 if，加锁是唯一能让它看见的办法。""",4));
    }
    static String shortName(String l){
        return switch(l){ case RU->"RU"; case RC->"RC"; case RR->"RR";
                          case RRLOCK->"RR+FOR UPD"; default->"SERIAL"; };
    }

    // ══════════════════════════════════════════════════════════════════
    // Lab 4A-2
    // ══════════════════════════════════════════════════════════════════
    static void lab2(){
        head("4A-2","MVCC 的 ReadView 可见性判定","RC 和 RR 的实现差异只有一行：ReadView 什么时候创建");
        System.out.print(blk("""

    同一条版本链、同一套四步判定算法，只改「ReadView 何时创建」这一件事。""",4));
        Version[] chain={new Version(32,500),new Version(28,400),new Version(24,300),
                         new Version(18,200),new Version(12,100)};
        System.out.print("\n    balance 这一行的版本链（新 → 旧）：");
        for(int i=0;i<chain.length;i++){
            if(i>0) System.out.print("  →");
            System.out.printf("  [trx %d = %d]", chain[i].trx(), chain[i].val());
        }
        System.out.println();
        Engine e=new Engine(RR,Map.of());
        for(int c=0;c<2;c++){
            List<Integer> mids = c==0? List.of(28,32) : List.of(32);
            String label = c==0 ? "① 事务 25 建 ReadView 时，trx 28 和 32 都还没提交"
                                : "② trx 28 提交了 —— RC 会在下一条语句重新采集活跃集合";
            ReadView rv=new ReadView(); rv.mids=mids; rv.max=40; rv.creator=25;
            rv.min=mids.isEmpty()?40:mids.get(0);
            System.out.printf("%n    %s%n", label);
            System.out.printf("      ReadView{ m_ids=%s  min=%d  max=%d  creator=%d }%n",
                mids.toString().replace(",",""), rv.min, rv.max, rv.creator);
            boolean picked=false;
            for(Version v:chain){
                boolean ok=e.visible(v.trx(),rv);
                String why;
                if(v.trx()==rv.creator) why="trx == creator ⇒ 我自己改的";
                else if(v.trx()<rv.min) why="trx < min("+rv.min+") ⇒ 建 ReadView 前就已提交";
                else if(v.trx()>=rv.max) why="trx >= max("+rv.max+") ⇒ 在我之后才开始";
                else if(mids.contains(v.trx())) why="trx 在 m_ids 里 ⇒ 我建 ReadView 时它还没提交";
                else why="不在 m_ids 且 < max ⇒ 已提交";
                String mark="不可见";
                if(ok&&!picked){ mark="★ 可见"; picked=true; why+="　← 第一个可见的，就是答案"; }
                else if(ok) mark="可见";
                System.out.printf("      trx %-3d = %-4d %s  %s%n", v.trx(), v.val(), padR(mark,8), why);
            }
        }
        System.out.print(blk("""

    结论
      · 同一条版本链，仅仅因为「活跃事务集合」采集的时刻不同，
        返回值就从 300 跳到了 400。
      · RC：每条语句都新建 ReadView ⇒ 上面两种情况会在同一个事务里先后出现 ⇒ 不可重复读。
        RR：整个事务只建一次并复用     ⇒ 永远停留在情况 ①     ⇒ 可重复读。
      · 【这就是 RC 和 RR 的全部实现差异。】版本链、判定算法、undo log 结构，两者完全一样。
      · 顺带纠一个极常见的错误：版本链存在 undo log 里，不是 redo log。
        redo log 是物理日志，只用于崩溃恢复的 roll-forward，和可见性判断毫无关系。""",4));
    }

    // ══════════════════════════════════════════════════════════════════
    // Lab 4A-3：写偏斜 —— 医院值班表
    // ══════════════════════════════════════════════════════════════════
    static List<Op> skewSteps(int c, Rnd rnd){
        List<List<Op>> per=new ArrayList<>();
        for(int t=1;t<=c;t++) per.add(List.of(
            o(t,"begin"),
            o(t,"scan").p(P_ONCALL).lock(),
            o(t,"write").k("doc:"+t).v(0).gmin(2),
            o(t,"commit")));
        int[] idx=new int[c]; List<Op> out=new ArrayList<>(); int left=4*c;
        while(left>0){
            List<Integer> avail=new ArrayList<>();
            for(int i=0;i<c;i++) if(idx[i]<per.get(i).size()) avail.add(i);
            int p=avail.get(rnd.intn(avail.size()));
            out.add(per.get(p).get(idx[p]++)); left--;
        }
        return out;
    }
    static void lab3(){
        head("4A-3","写偏斜：医院值班表","不变量「任何时刻至少 1 个医生在岗」，看谁守得住");
        int D=DOCTORS, C=Math.min(LEAVERS,DOCTORS), N=TRIALS;
        System.out.printf("""

    %d 个医生在岗，%d 个人同时发起请假。每个人的事务都是：
      BEGIN; SELECT count(*) WHERE on_call=true;  if (>= 2) UPDATE 自己 = 不在岗; COMMIT
    每次试验随机交错这些步骤，跑 %d 次。
""", D, C, N);
        int[] w={26,18,16,18,18}; String al="LRRRR";
        tableHead(new String[]{"隔离配置","不变量被破坏","平均最终在岗","出现过锁等待","主动放弃请假"},w,al);
        for(String lv: new String[]{RC,RR,RRLOCK,SER}){
            int bad=0,sumOn=0,blocked=0,gaveUp=0;
            for(int s=1;s<=N;s++){
                Rnd rnd=new Rnd(s*2654435761L);
                Map<String,Integer> init=new LinkedHashMap<>();
                for(int i=1;i<=D;i++) init.put("doc:"+i,1);
                Engine e=new Engine(lv,init);
                List<Op> steps=skewSteps(C,rnd);
                e.run(lv.equals(SER)?serialize(steps):steps);
                int on=e.countWhere(P_ONCALL);
                sumOn+=on; if(on<1) bad++;
                if(e.blocked()) blocked++;
                for(Trace t:e.trace) if(t.txt().contains("放弃写入")) gaveUp++;
            }
            tableRow(new String[]{LVNAME.get(lv), bad+" / "+N,
                String.format("%.2f 人",(double)sumOn/N), blocked+" / "+N, gaveUp+" 次"},w,al);
        }
        System.out.print(blk("""

    结论
      · RC 和 RR 的违规次数【几乎一样】。把隔离级别从 RC 调到 RR，对写偏斜没有任何帮助——
        RR 加强的是「我读到的不变」，而这里的问题是【你读到的确实没变，只是它已经过时了】。
      · 加上 FOR UPDATE 后违规降到 0，代价是大量锁等待，以及若干次
        「有人查到人不够，主动放弃请假」——后者正是我们想要的行为：
        不是数据库报错，而是应用自己发现前提不成立。
      · 加锁把「读」这个动作物化成了一件别人看得见的事，
        于是数据库终于知道你的读和写之间有因果关系。""",4));
    }

    // ══════════════════════════════════════════════════════════════════
    // Lab 4A-4：2PL vs SSI
    //
    // 两个模拟器共用同一个物理前提：系统每 tick 最多能执行 PAR 个操作。
    // 这一条很重要——它让「白做的功」真的变成时间，而不是一个漂亮的免费数字。
    // ══════════════════════════════════════════════════════════════════
    static final int PAR=4, HOTKEYS=2;
    record CCOp(int key,boolean write){}
    record CCTxn(int id,List<CCOp> ops){}
    record CCResult(int tick,int aborts,int wasted,int work){}

    static List<CCTxn> mkTxns(int c,int o,double hot,Rnd rnd){
        int nkey=Math.max(24,c*o);   // 键空间随负载伸缩，否则「低冲突」根本低不下来
        List<CCTxn> ts=new ArrayList<>();
        for(int i=0;i<c;i++){
            List<CCOp> ops=new ArrayList<>();
            for(int j=0;j<o;j++){
                int k=rnd.intn(nkey);
                if(rnd.flt()<hot) k=rnd.intn(HOTKEYS);
                ops.add(new CCOp(k, rnd.flt()<0.5));
            }
            ts.add(new CCTxn(i,ops));
        }
        return ts;
    }
    static final class St2 { int id,p,cool; List<CCOp> ops; List<Integer> locks=new ArrayList<>(); boolean done; }
    record Lk(int trx,boolean x){}

    /**
     * 2PL：悲观。拿不到锁就等（等待期间不占用 CPU）；等待图成环 = 死锁。
     * 关键性质：等待【不浪费】已经做的功——等到了就接着往下做。
     * 进度保证：编号最小的活跃事务永远不会被选为牺牲者 ⇒ 它一定跑得完，然后轮到下一个。
     */
    static CCResult sim2PL(List<CCTxn> txns){
        List<St2> st=new ArrayList<>();
        for(CCTxn t:txns){ St2 s=new St2(); s.id=t.id(); s.ops=t.ops(); st.add(s); }
        Map<Integer,Lk> owner=new HashMap<>();
        int tick=0,aborts=0,wasted=0,work=0;
        while(st.stream().anyMatch(s->!s.done)&&tick<8000){
            tick++;
            List<St2> alive=st.stream().filter(s->!s.done).toList();
            List<St2> runnable=new ArrayList<>(); Map<Integer,Integer> waitFor=new LinkedHashMap<>();
            boolean cooling=false;
            for(St2 s:alive){
                if(s.cool>0){ s.cool--; cooling=true; continue; }
                CCOp op=s.ops.get(s.p); Lk o=owner.get(op.key());
                if(o!=null&&o.trx()!=s.id&&(o.x()||op.write())){ waitFor.put(s.id,o.trx()); continue; }
                runnable.add(s);
            }
            if(runnable.isEmpty()&&!cooling){
                // 没有任何事务能动 ⇒ 等待图里必然有环。把环找出来，只在【环里】挑牺牲者——
                // 中止一个环外的事务不会解开任何东西，这正是「死锁检测」和「随便杀一个」的区别。
                List<Integer> cyc=new ArrayList<>();
                for(int a: waitFor.keySet()){
                    List<Integer> path=new ArrayList<>(); Set<Integer> seen=new HashSet<>();
                    Integer cur=a;
                    while(cur!=null&&!seen.contains(cur)){ seen.add(cur); path.add(cur); cur=waitFor.get(cur); }
                    if(cur!=null){ cyc=path.subList(path.indexOf(cur),path.size()); break; }
                }
                int protectedId=alive.get(0).id;
                List<St2> pool=new ArrayList<>();
                for(int id:cyc) for(St2 s:alive) if(s.id==id&&s.id!=protectedId) pool.add(s);
                if(pool.isEmpty()) for(St2 s:alive) if(s.id!=protectedId&&!s.locks.isEmpty()) pool.add(s);
                if(pool.isEmpty()) for(St2 s:alive) if(s.id!=protectedId) pool.add(s);
                if(pool.isEmpty()) break;
                St2 v=pool.get(0);
                for(St2 s:pool) if(s.p<v.p) v=s;   // 挑已做的功最少的，浪费最小
                aborts++; wasted+=v.p;
                for(int k:v.locks){ Lk o=owner.get(k); if(o!=null&&o.trx()==v.id) owner.remove(k); }
                v.locks=new ArrayList<>(); v.p=0; v.cool=2;
                continue;
            }
            int budget=PAR;
            for(St2 s:runnable){
                if(budget<=0) break;
                CCOp op=s.ops.get(s.p); Lk o=owner.get(op.key());
                if(o!=null&&o.trx()!=s.id&&(o.x()||op.write())) continue; // 同 tick 内被别人抢先拿走
                if(o==null||o.trx()!=s.id){ owner.put(op.key(),new Lk(s.id,op.write())); s.locks.add(op.key()); }
                else if(op.write()) owner.put(op.key(),new Lk(s.id,true));
                s.p++; work++; budget--;
                if(s.p>=s.ops.size()){ s.done=true;
                    for(int k:s.locks){ Lk oo=owner.get(k); if(oo!=null&&oo.trx()==s.id) owner.remove(k); }
                    s.locks=new ArrayList<>(); }
            }
        }
        return new CCResult(tick,aborts,wasted,work);
    }
    static final class StS { int id,p,start; List<CCOp> ops; Set<Integer> rs,ws; boolean done; }
    /**
     * SSI：乐观。谁也不等，全速往前跑；跑完一整个事务才验证
     *   「我读到的前提有没有被别人推翻」——被推翻就整个重做。
     * 验证条件用的是 SSI 的危险结构近似：
     *   ① 写-写冲突（快照隔离本来就要求的 first-committer-wins），或
     *   ② 双向的读-写反依赖（我读了它写的，同时它读了我写的）—— 这才是真正会成环的那种。
     * 关键性质：中止 = 这一轮做的功【全部白费】，而且白费的功照样占 CPU。
     */
    static CCResult simSSI(List<CCTxn> txns){
        List<StS> st=new ArrayList<>();
        for(CCTxn t:txns){
            StS s=new StS(); s.id=t.id(); s.ops=t.ops();
            s.rs=new HashSet<>(); s.ws=new HashSet<>();
            for(CCOp o:t.ops()){ if(o.write()) s.ws.add(o.key()); else s.rs.add(o.key()); }
            st.add(s);
        }
        record H(Set<Integer> rs,Set<Integer> ws){}
        List<H> hist=new ArrayList<>();
        int tick=0,aborts=0,wasted=0,work=0;
        while(st.stream().anyMatch(s->!s.done)&&tick<8000){
            tick++; int budget=PAR;
            for(StS s:st){
                if(s.done||budget<=0) continue;
                s.p++; work++; budget--;
                if(s.p<s.ops.size()) continue;
                boolean conflict=false;
                for(int i=s.start;i<hist.size()&&!conflict;i++){
                    H u=hist.get(i);
                    if(inter(s.ws,u.ws())) conflict=true;                    // ① 写-写
                    else if(inter(s.rs,u.ws())&&inter(s.ws,u.rs())) conflict=true; // ② 双向 rw 反依赖
                }
                if(conflict){ aborts++; wasted+=s.ops.size(); s.p=0; s.start=hist.size(); }
                else { s.done=true; hist.add(new H(s.rs,s.ws)); }
            }
        }
        return new CCResult(tick,aborts,wasted,work);
    }
    static boolean inter(Set<Integer> a,Set<Integer> b){ for(int k:a) if(b.contains(k)) return true; return false; }

    static void lab4(){
        head("4A-4","2PL vs SSI：两张不同的账单","一个付「等待」，一个付「白做的功」");
        System.out.printf("""

    %d 个并发事务，每个 %d 个操作，系统每 tick 最多执行 %d 个操作。
      2PL：拿不到锁就等（等待不占 CPU，做过的功不浪费），等待图成环 = 死锁中止。
      SSI：谁也不等，全速跑完再验证；验证不过就整个事务从头再来（白做的功照样烧 CPU）。
""", TXNS, OPS, PAR);
        int[] w={12,3,12,12,14,12,3,12,12,14,12}; String al="LLRRRRLRRRR";
        tableHead(new String[]{"热点","","2PL 耗时","中止","白做操作","白做占比",
            "","SSI 耗时","中止","白做操作","白做占比"},w,al);
        for(int hh=0;hh<=100;hh+=20){
            CCResult a=sim2PL(mkTxns(TXNS,OPS,hh/100.0,new Rnd(SEED*2654435761L)));
            CCResult b=simSSI(mkTxns(TXNS,OPS,hh/100.0,new Rnd(SEED*2654435761L)));
            tableRow(new String[]{hh+"%","",
                ""+a.tick(),""+a.aborts(),""+a.wasted(),pctOf(a),"",
                ""+b.tick(),""+b.aborts(),""+b.wasted(),pctOf(b)},w,al);
        }
        System.out.print(blk("""

    结论
      · 看两列「白做占比」：热点集中度往右拉，2PL 的几乎不涨，SSI 的一路涨到有效功的几倍。
        而两者的【墙上时间】相差不大——这正是这个对比的重点：
        SSI 省下来的等待时间，是用真金白银的 CPU 换的。
      · 所以没有哪一个「总是更好」：
          冲突低、事务短 ⇒ SSI 更合适（谁也不等，读事务完全无锁，延迟分布好看得多）
          冲突高、事务长 ⇒ 2PL 更稳  （等待难看，但做过的功不会被扔掉，行为可预测）
      · 还有一个数字上看不见的代价：SSI 要求应用层有【正确的】重试逻辑。
        重试时如果复用了上一次读到的值，那比不用 SSI 还危险——
        你把一个本来会被数据库拦下的错误，变成了一个静默的错误。""",4));
    }
    static String pctOf(CCResult r){
        return r.work()==0? "—" : String.format("%.0f%%", (double)r.wasted()/r.work()*100);
    }

    static void epilogue(){
        System.out.println("\n"+"═".repeat(78));
        System.out.print(blk("""
Part 4-A 到此结束

这一章的一句话：【隔离级别是打折出售的「世界是静止的」这个假设，
而折扣的具体内容，各家数据库同名不同物。】

最该带走的三条：
  1. 版本链在 undo log，不在 redo log；RC 与 RR 的实现差异只有
     「ReadView 何时创建」这一行。
  2. InnoDB 的 RR 防幻读靠两套机制：快照读靠 ReadView，当前读靠 Next-Key Lock。
     但它【防不住写偏斜】，Oracle 的 SERIALIZABLE 也防不住。
  3. 能变成约束的不变量，一律变成约束；剩下的，把读物化成锁。

下一站 Part 4-B · 分布式事务的提交协议：
2PC 与 XA、它的阻塞问题、TCC、Saga、本地消息表、事务消息——
以及一个贯穿始终的问题：当「提交」这个动作本身要跨机器时，
Part 0 的那个「第三态」又回来了。""",2));
        System.out.println("═".repeat(78)+"\n");
    }

    // ─── 输出工具 ─────────────────────────────────────────────────────
    /** 缩进一段文本；空行不补空格（否则会留下看不见的行尾空白）。 */
    static String blk(String s,int k){ return s.indent(k).replaceAll("(?m)^[ ]+$",""); }
    static int dispw(String s){
        int w=0;
        for(int i=0;i<s.length();i++){
            char c=s.charAt(i);
            w += (c>=0x1100 && (c<=0x115F || (c>=0x2E80&&c<=0xA4CF&&c!=0x303F)
                    || (c>=0xAC00&&c<=0xD7A3) || (c>=0xF900&&c<=0xFAFF)
                    || (c>=0xFE30&&c<=0xFE6F) || (c>=0xFF00&&c<=0xFF60)
                    || (c>=0xFFE0&&c<=0xFFE6))) ? 2 : 1;
        }
        return w;
    }
    static String padR(String s,int w){ int d=w-dispw(s); return d>0? s+" ".repeat(d):s; }
    static String padL(String s,int w){ int d=w-dispw(s); return d>0? " ".repeat(d)+s:s; }
    static String layout(String[] c,int[] w,String al){
        StringBuilder sb=new StringBuilder("    ");
        for(int i=0;i<c.length;i++)
            sb.append(i<al.length()&&al.charAt(i)=='L'? padR(c[i],w[i]) : padL(c[i],w[i]));
        return sb.toString().stripTrailing();
    }
    static void head(String label,String title,String sub){
        System.out.println("\n"+"═".repeat(78));
        System.out.printf("  Lab %s · %s%n  %s%n", label, title, sub);
        System.out.println("═".repeat(78));
    }
    static void tableHead(String[] cols,int[] w,String al){
        System.out.println(layout(cols,w,al));
        int t=0; for(int x:w) t+=x;
        System.out.println("    "+"─".repeat(t));
    }
    static void tableRow(String[] c,int[] w,String al){ System.out.println(layout(c,w,al)); }

    static void parseArgs(String[] a){
        for(int i=0;i<a.length;i++){
            String k=a[i].replaceFirst("^-+","");
            if(i+1>=a.length) break;
            String v=a[++i];
            switch(k){
                case "trials" -> TRIALS=Integer.parseInt(v);
                case "doctors" -> DOCTORS=Integer.parseInt(v);
                case "leavers" -> LEAVERS=Integer.parseInt(v);
                case "txns" -> TXNS=Integer.parseInt(v);
                case "ops" -> OPS=Integer.parseInt(v);
                case "seed" -> SEED=Long.parseLong(v);
                default -> { }
            }
        }
    }
}
