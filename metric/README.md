
# metrics

## 类型定义

```go
// tags 为标签, 和prometheus的label一样
type Tags map[string]string

// 数据上报接口, 是对其他包括prometheus在内的平台的抽象
type Reporter interface {
 Count(name string, value float64, tags Tags) error
 Gauge(name string, value float64, tags Tags) error
 Histogram(name string, value float64, tags Tags) error
 Summary(name string, value float64, tags Tags) error
 Start() error
 Stop() error
}

// Counter 描述一种递增的指标
type Counter interface {
 With(tags Tags) Counter
 Add(delta float64) error
}

// Gauge 表示一个可增可减的数字变量，代表当前的指标
type Gauge interface {
 With(tags Tags) Gauge
 Set(value float64) error
 Add(delta float64) error
}

// Histogram 通过桶的方式采样统计指标
type Histogram interface {
 With(tags Tags) Histogram
 Observe(value float64) error
}

// Summary 在 client 端聚合数据, 直接存储了分位数
type Summary interface {
 With(tags Tags) Summary
 Observe(value float64) error
}
```

## 初始化和运行

```go
// 初始化一个prometheus的Reporter
reporter, err := prometheus.NewReporter(
    metrics.Path("/metrics"),
    metrics.Address(":8089"),
)
checkErr(err)

// Stop 会关闭reporter的数据上报服务
defer reporter.Stop()
// Start 开启reporter的数据上报服务
_ = reporter.Start()

// 把reporter赋值给全局Reporter, 以便可以全局的调用指标上报函数
metrics.SetDefaultReporter(reporter)

// GetDefaultReporter获取全局的Reporter
_=metrics.GetDefaultReporter()

// 定义一个Tags
var tag = metrics.Tags{"test": "123456"}

// 定义一个Counter指标
c := metrics.NewCounter("count_req")
checkErr(c.With(tag).Add(1))

// 定义一个Gauge指标
g := metrics.NewGauge("gauge")
checkErr(g.Set(1))

// 定义一个Histogram指标
h := metrics.NewHistogram("Histogram")
checkErr(h.With(tag).Observe(1))

// 定义一个Summary指标
s := metrics.NewSummary("summary")
checkErr(s.With(tag).Observe(1))

// 调用全局的Counter进行指标上报
checkErr(metrics.Count("count1", 2, metrics.Tags{"t1": "1", "t2": "2"}))
// 调用全局的Summary进行指标上报
checkErr(metrics.Summary("summary", 1, tag))
```

## 监控类型介绍

### Counter(计数器)

> 单调递增的计数器，重启时重置为0，其余时候只能增加。

1. Counter 类型代表一种样本数据单调递增的指标，即只增不减，除非监控系统发生了重置

2. 例如，你可以使用 Counter 类型的指标来表示**服务的请求数**、**已完成的任务数**、**错误发生的次数**等。

3. Counter 类型数据可以让用户方便的了解事件产生的速率的变化

4. 不要将 Counter 类型应用于样本数据非单调递增的指标，例如：**当前运行的进程数量**（应该用 Guage 类型）。

#### 计数器常测量对象

1. 请求的数量

2. 任务完成的数量

3. 函数调用次数

4. 错误发生次数

### Gauge(仪表盘)

> 表示一个可增可减的数字变量，初值为0

1. Guage 类型代表一种样本数据可以任意变化的指标，即可增可减

2. Guage 通常用于像**温度**或者**内存使用率**这种指标数据，也可以表示能随时增加或减少的“总数”，例如：**当前并发请求的数量**。

#### 仪表盘常测量对象

1. 温度

2. 内存用量

3. 并发请求数

### Histogram(直方图)

> Histogram 会对观测数据取样，然后将观测数据放入有数值上界的桶中，并记录各桶中数据的个数，所有数据的个数和数据数值总和。

1. 量化指标的平均值, 例如 **CPU 的平均使用率**、**页面的平均响应时间**

2. 以系统 API 调用的**平均响应时间**为例：如果大多数 API 请求都维持在 100ms 的响应时间范围内，而个别请求的响应时间需要 5s，那么就会导致某些 WEB 页面的响应时间落到中位数的情况，而这种现象被称为**长尾问题** 。

3. 为了区分是**平均的慢**还是**长尾的慢**，最简单的方式就是**按照请求延迟的范围进行分组**。例如，**统计延迟在 0~10ms 之间的请求数**有多少而 **10~20ms 之间的请求数**又有多少。通过这种方式可以快速分析系统慢的原因。

4. Histogram 在**一段时间范围内对数据进行采样**（通常是**请求持续时间**或**响应大小**等），并将其计入**可配置的存储桶**（bucket）中，后续可**通过指定区间筛选样本**，也可以统计样本总数，最后一般将数据展示为直方图。

5. Histogram 类型的样本会提供三种指标（假设指标名称basename）
 
    1. 样本的值分布在 bucket 中的数量，命名为 _bucket{le="上边界"}

    2. 解释得更通俗易懂一点，这个值表示指标值小于等于上边界的所有样本数量

    3. 对每个采样点进行统计（并不是一段时间的统计），打到各个桶(bucket)中

    4. 对每个采样点值累计和(sum)

    5. 对采样点的次数累计和(count)

    6. 度量指标名称: [basename]的柱状图, 上面三类的作用度量指标名称

    7. [basename]_bucket{le=“上边界”}, 这个值为小于等于上边界的所有采样点数

    8. [basename]_sum

    9. [basename]_count

    10. histogram并不会保存数据采样点值，每个bucket只有个记录样本数的counter（float64），即histogram存储的是区间的样本数统计值，因此客户端性能开销相比 Counter 和 Gauge 而言没有明显改变，适合高并发的数据收集。

    11. 具体实现：Histogram 会根据观测的样本生成如下数据：

    12. inf 表无穷值，a1，a2，……是单调递增的数值序列。

    13. [basename]_count：数据的个数，类型为 counter

    14. [basename]_sum：数据的加和，类型为 counter

    15. [basename]_bucket{le=a1}： 处于 [-inf,a1] 的数值个数

    16. [basename]_bucket{le=a2}：处于 [-inf,a2] 的数值个数


6. ……


7. [basename]_bucket{le=<+inf>}：处于 [-inf,+inf] 的数值个数，Prometheus 默认额外生成，无需用户定义

8. Histogram 可以计算样本数据的百分位数，其计算原理为：通过找特定的百分位数值在哪个桶中，然后再通过插值得到结果。比如目前有两个桶，分别存储了 [-inf, 1] 和 [-inf, 2] 的数据。然后现在有 20% 的数据在 [-inf, 1] 的桶，100% 的数据在 [-inf, 2] 的桶。那么，50% 分位数就应该在 [1, 2] 的区间中，且处于 (50%-20%) / (100%-20%) = 30% / 80% = 37.5% 的位置处。Prometheus 计算时假设区间中数据是均匀分布，因此直接通过线性插值可以得到 (2-1)*3/8+1 = 1.375。

#### 直方图常测量对象

1. 请求时延


2. 回复长度


3. ……各种有样本数据



### Summary(摘要)

> Summary 与 Histogram 类似，会对观测数据进行取样，得到数据的个数和总和。此外，还会取一个滑动窗口，计算窗口内样本数据的分位数。

1. 具体实现：Summary 完全是**在 client 端聚合数据**，每次调用 obeserve 会计算出如下数据

2. [basename]_count：数据的个数，类型为 counter

3. [basename]_sum：数据的加和，类型为 counter

4. [basename]{quantile=0.5}：滑动窗口内 50% 分位数值

5. [basename]{quantile=0.9}： 滑动窗口内 90% 分位数值

6. [basename]{quantile=0.99}：滑动窗口内 99% 分位数值

7. ……

8. 实际分位数值可根据需求制定，且是对每一个 Label 组合做聚合。

9. 与 Histogram 类型类似，用于表示一段时间内的数据采样结果（通常是请求持续时间或响应大小等），但它**直接存储了分位数**（通过客户端计算，然后展示出来），而**不是通过区间来计算**。

10. Summary 类型的样本也会提供三种指标（假设指标名称为 <basename>）：

11. 样本值的分位数分布情况，命名为 <basename>{quantile="<φ>"}。

12. 所有样本值的大小总和，命名为 <basename>_sum。

13. 样本总数，命名为 <basename>_count。

14. 在客户端对于一段时间内（默认是10分钟）的每个采样点进行统计，并形成分位图。（如：正态分布一样，统计低于60分不及格的同学比例，统计低于80分的同学比例，统计低于95分的同学比例）

15. 统计班上所有同学的总成绩(sum)

16. 统计班上同学的考试总人数(count)

17. 带有度量指标的[basename]的summary 在抓取时间序列数据展示。

18. 观察时间的φ-quantiles (0 ≤ φ ≤ 1), 显示为[basename]{分位数="[φ]"}

19. [basename]_sum， 是指所有观察值的总和

20. [basename]_count, 是指已观察到的事件计数值

#### 摘要常测量对象

1. 请求时延

2. 回复长度

3. ……各种有样本数据

### Histogram与Summary的比较

1. 它们都包含了 <basename>_sum 和 <basename>_count 指标

2. Histogram 需要通过 <basename>_bucket 来计算分位数，而 Summary 则直接存储了分位数的值。

3. [https://prometheus.io/docs/practices/histograms/](https://prometheus.io/docs/practices/histograms/)

4. Summary 结构有频繁的全局锁操作，对高并发程序性能存在一定影响。

5. histogram仅仅是给每个桶做一个原子变量的计数就可以了，而summary要每次执行算法计算出最新的X分位value是多少，算法需要并发保护。会占用客户端的cpu和内存。

6. 不能对Summary产生的quantile值进行aggregation运算（例如sum, avg等）。例如有两个实例同时运行，都对外提供服务，分别统计各自的响应时间。最后分别计算出的0.5-quantile的值为60和80，这时如果简单的求平均(60+80)/2，认为是总体的0.5-quantile值，那么就错了。

7. summary的百分位是提前在客户端里指定的，在服务端观测指标数据时不能获取未指定的分为数。而histogram则可以通过promql随便指定，虽然计算的不如summary准确，但带来了灵活性。

8. histogram不能得到精确的分为数，设置的bucket不合理的话，误差会非常大。会消耗服务端的计算资源。

#### Summary：

1. 优点

    1. 能够非常准确的计算百分位数

    2. 不需要提前知道数据的分布

2. 缺点：

    1. 灵活性不足，实时性需要通过 maxAge 来保证，写死了后灵活性就不太够（比如想知道更长维度的百分位数）

    2. 在 client 端已经做了聚合，即在各个用户集群的 ipamD 中已经聚合了，我们如果需要观察全部 user 下的百分位数数据是不行的（只能看均值）

    3. 用户集群的 ipamD 的调用频率可能很低（如小集群或者稳定集群），这种情况下 client 端聚合计算百分位数值失去意义（数据太少不稳定），如果把 maxAge 增大则失去实时性

#### Histogram：

1. 优点

    1. 兼具灵活性和实时性

    2. 可以灵活的聚合数据，观察各个尺度和维度下的数据

2. 缺点

    1. 需要提前知道数据的大致分布，并以此设计出合适而准确的桶序列

    2. 难以通过 Label 串联多种 Metrics，因为各个 Metrics 的数据分布可能差异较大，如果都只用一种桶序列的话会导致百分位数计算差异较大

### 总结

1. 如果需要聚合（aggregate），选择histograms。

2. 如果比较清楚要观测的指标的范围和分布情况，选择histograms。

3. 如果需要精确的分位数选择summary。

4. Summary 的缺点过于致命，难以回避。

5. Histogram 的缺点可以通过增加工作量（即通过测试环境中的实验来确定各 Metrics 的大致分布）和增加 Metrics（不用 Label 区分）来较好解决。

6. Histogram 计算误差大，但灵活性较强，适用客户端监控、或组件在系统中较多、或不太关心精确的百分位数值的场景

7. Summary 计算精确，但灵活性较差，适用服务端监控、或组件在系统中唯一或只有个位数、或需要知道较准确的百分位数值（如性能优化场景）的场景

![](https://api2.mubu.com/v3/document_image/7b1bc08a-5dc4-414a-87db-02f0727ebd9d-40263.jpg)

## 一般性监控指标

### 从需求出发

1. 延迟：服务请求的时间。

2. 通讯量：监控当前系统的流量，用于衡量服务的容量需求。

3. 错误：监控当前系统所有发生的错误请求，衡量当前系统错误发生的速率。

4. 饱和度： 衡量当前服务的饱和度。 主要强调最能影响服务状态的受限制的资源。 例如，如果系统主要受内存影响，那就主要关注系统的内存状态。

以上四种指标，其实是为了满足四个监控需求：

1. 反映用户体验，衡量系统核心性能。如：在线系统的时延，作业计算系统的作业完成时间等。

2. 反映系统的服务量。如：请求数，发出和接收的网络包大小等。

3. 帮助发现和定位故障和问题。如：错误计数、调用失败率等。

4. 反映系统的饱和度和负载。 如： 系统占用的内存、作业队列的长度等。

### 从需监控的系统出发

另一方面，为了满足相应的需求，不同系统需要观测的测量对象也是不同的。在 官方文档 的最佳实践中，将需要监控的应用分为了三类：

1. 线上服务系统（Online-serving systems）：需对请求做即时的响应，请求发起者会等待响应。如 web 服务器。

2. 线下计算系统（Offline processing）：请求发起者不会等待响应，请求的作业通常会耗时较长。如批处理计算框架 Spark 等。

3. 批处理作业（Batch jobs）：这类应用通常为一次性的，不会一直运行，运行完成后便会结束运行。如数据分析的 MapReduce 作业。

对于每一类应用其通常情况下测量的对象是不太一样的。其总结如下：

1. 线上服务系统：主要有请求、出错的数量，请求的时延等。

2. 线下计算系统：最后开始处理作业的时间，目前正在处理作业的数量，发出了多少 items， 作业队列的长度等。

3. 批处理作业：最后成功执行的时刻，每个主要 stage 的执行时间，总的耗时，处理的记录数量等。
除了系统本身，有时还需监控子系统：

4. 使用的库（Libraries）: 调用次数，成功数，出错数，调用的时延。

5. 日志（Logging）：计数每一条写入的日志，从而可找到每条日志发生的频率和时间。

6. Failures: 错误计数。

7. 线程池： 排队的请求数，正在使用的线程数，总线程数，耗时，正在处理的任务数等。

8. 缓存：请求数，命中数，总时延等。

9. ……

## 如何选用 Vector

选用 Vec 的原则：

1. 数据类型类似但资源类型、收集地点等不同

2. Vec 内数据单位统一

3. 例子：

1. 不同资源对象的请求延迟

2. 不同地域服务器的请求延迟

3. 不同 http 请求错误的计数

4. ……

5. 此外，官方文档 中建议，对于一个资源对象的不同操作，如 Read/Write、Send/Receive， 应采用不同的 Metric 去记录，而不要放在一个 Metric 里。

6. 原因是监控时一般不会对这两者做聚合，而是分别去观测。

7. 不过对于 request 的测量，通常是以 Label 做区分不同的 action。

## 如何确定 Label

根据上文，常见 Label 的选择有：

1. resource

2. region

3. type

4. ……

5. 确定 Label 的一个重要原则是：**同一维度 Label 的数据是可平均和可加和的**，也即**单位要统一**。如风扇的风速和电压就不能放在一个 Label 里。

6. 此外，不建议下列做法：

1. my_metric { label = a } 1

2. my_metric { label = b } 6

3. my_metric { label = total } 7

4. 即在 Label 中同时统计了分和总的数据，建议在服务器端聚合得到总和的结果。

5. 或者用另外的 Metric 去测量总的数据。

## 如何命名 Metrics 和 Label

### Metric 的命名

1. 需要符合 pattern: [a-zA-Z:][a-zA-Z0-9:]*

2. 应该包含一个单词作为前缀，表明这个 Metric 所属的域。如：
prometheus_notifications_total
process_cpu_seconds_total

3. 应该包含一个单位的单位作为后缀，表明这个 Metric 的单位。 如：
http_request_duration_seconds
node_memory_usage_bytes
http_requests_total (for a unit-less accumulating count)

4. 逻辑上与被测量的变量含义相同。

5. 尽量使用基本单位，如 seconds，bytes。而不是 Milliseconds, megabytes。

### Label 的命名

依据选择的维度命名，如：

1. region：shenzhen/guangzhou/beijing

2. owner：user1/user2/user3

3. stage：extract/transform/load

## 进程级别监控指标

```md
pid
threads         线程数 <threads>[type:gauge]
memory usage    内存使用% <process_memory_precent>[type:gauge]
cpu usage       cpu使用% <process_cpu_precent>[type:gauge]
username        用户名
cmd+args        命令+参数
进程内栈监控指标
goroutine       协程数量{状态:[运行,休眠,系统调用,管道接收]} <goroutines>[type:gauge]
进程内内存监控指标
Alloc           进程堆空间分配的字节数 <memstats_alloc_bytes>[type:gauge]
TotalAlloc      从开始运行至今分配器为分配的堆空间总和，只增不减 <memstats_alloc_bytes_total>[type:counter]
Sys             Number of bytes obtained from system. <memstats_sys_bytes>[type:gauge]
Lookups         被runtime监视的指针数 <memstats_lookups_total>[type:counter]
Mallocs         进程malloc heap objects的次数 <memstats_mallocs_total>[type:counter]
Frees           进程回收的heap objects的次数 <memstats_frees_total>[type:counter]
HeapAlloc       当前进程分配的堆内存字节数 <memstats_heap_alloc_bytes>[type:gauge]
HeapSys         系统分配的作为运行堆的字节数 <memstats_heap_sys_bytes>[type:gauge]
HeapIdle        申请但是未分配的堆内存或者回收了的堆内存（空闲）字节数 <memstats_heap_idle_bytes>[type:gauge]
HeapInuse       正在使用的堆内存字节数 <memstats_heap_inuse_bytes>[type:gauge]
HeapReleased    返回给系统的堆内存 <memstats_heap_released_bytes>[type:gauge]
HeapObjects     堆内存块申请的量 <memstats_heap_objects>[type:gauge]
StackInuse      正在使用的栈字节数 <memstats_stack_inuse_bytes>[type:gauge]
StackSys        系统分配的作为运行栈的内存 <memstats_stack_sys_bytes>[type:gauge]
MSpanInuse      分配给用于调试的mspan结构体字节数 <memstats_mspan_inuse_bytes>[type:gauge]
MSpanSys        系统为mspn结构体分配的字节数 <memstats_mspan_sys_bytes>[type:gauge]
MCacheInuse     mcache结构体申请的字节数(不会被视为垃圾回收) <memstats_mcache_inuse_byte>[type:gauge]
MCacheSys       堆空间用于mcache的字节数 <memstats_mcache_sys_bytes>[type:gauge]
BuckHashSys     用于剖析桶散列表的堆空间 <memstats_buck_hash_sys_bytes>[type:gauge]
GCSys           垃圾回收标记元信息使用的内存 <memstats_gc_sys_bytes>[type:gauge]
OtherSys        golang系统架构占用的额外空间 <memstats_other_sys_bytes>[type:gauge]
NextGC          垃圾回收器下次检视的内存大小 <memstats_next_gc_bytes>[type:gauge]
LastGC          垃圾回收器最后一次执行时间 <memstats_last_gc_time_seconds>[type:gauge]
PauseTotalNs    垃圾回收或者其他信息收集导致服务暂停的次数 <memstats_pause_total_seconds>[type:gauge]
PauseNs         一个循环队列，记录最近垃圾回收系统中断的时间(不记录)
PauseEnd        一个循环队列，记录最近垃圾回收系统中断的时间开始点(不记录)
NumGC           gc完成周期次数 <memstats_number_gc>[type:gauge]
NumForcedGC     调用runtime.GC()强制使用垃圾回收的次数 <memstats_num_forced_gc>[type:gauge]
GCCPUFraction   垃圾回收占用服务CPU工作的时间总和。goroutine*垃圾回收的时间 <memstats_gc_cpu_fraction>[type:gauge]
EnableGC        是否开启gc(不记录)
```