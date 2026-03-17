# cron

基于 [github.com/go-co-op/gocron](https://github.com/go-co-op/gocron) 的定时任务库，提供简洁的封装 API。

## 快速开始

```go
s, err := cron.New()
if err != nil {
    log.Fatal(err)
}
defer s.Shutdown()

// 每 5 秒执行
s.AddEvery(5*time.Second, func() {
    fmt.Println("tick")
})

s.Start()
// 阻塞或配合优雅关闭
```

## 任务类型

| 方法 | 说明 | 示例 |
|------|------|------|
| `AddEvery` | 固定间隔 | `AddEvery(5*time.Second, fn)` |
| `AddCron` | Cron 表达式 | `AddCron("0 0 * * *", fn)` 每天 0 点 |
| `AddDaily` | 每日指定时间 | `AddDaily(9, 0, 0, fn)` 每天 9:00 |
| `AddWeekly` | 每周指定时间 | `AddWeekly(time.Sunday, 9, 0, 0, fn)` |
| `AddOneTime` | 一次性任务 | `AddOneTime(t, fn)` |
| `Add` | 自定义（gocron 原生） | 见下方 |

## Add 用法

`Add` 直接使用 gocron 的 `JobDefinition` 和 `Task`，适用于封装方法未覆盖的场景：

```go
import "github.com/go-co-op/gocron/v2"

// 随机间隔任务：每 5–10 秒随机执行
s.Add(
    gocron.DurationRandomJob(5*time.Second, 10*time.Second),
    gocron.NewTask(func() { fmt.Println("random") }),
    cron.WithName("random-job"),
)

// 每月任务：每月 1 号 9:00 执行
s.Add(
    gocron.MonthlyJob(1, gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
    gocron.NewTask(func() { fmt.Println("monthly") }),
)

// 带参数的任务
s.Add(
    gocron.DurationJob(time.Minute),
    gocron.NewTask(func(name string, count int) {
        fmt.Printf("%s: %d\n", name, count)
    }, "job", 42),
)

// 任务完成后间隔：从任务结束开始算下次执行
s.Add(
    gocron.DurationJob(10*time.Second),
    gocron.NewTask(longRunningTask),
    cron.WithIntervalFromCompletion(),
)
```

## 任务选项

```go
// 任务名称
s.AddEvery(interval, fn, cron.WithName("my-job"))

// 单例模式：重叠时跳过
s.AddEvery(interval, fn, cron.WithSingletonReschedule())

// 单例模式：重叠时排队等待
s.AddEvery(interval, fn, cron.WithSingletonQueue())

// 限制执行次数
s.AddEvery(interval, fn, cron.WithLimitedRuns(10))

// 从任务完成时开始计算下次执行
s.AddEvery(interval, fn, cron.WithIntervalFromCompletion())
```

## 调度器选项

```go
import "github.com/go-co-op/gocron/v2"

// 启用日志
s, _ := cron.New(cron.WithLogger(gocron.NewLogger(log.Default())))
```

## 高级用法

### 获取底层 Scheduler

通过 `Scheduler()` 获取底层 `gocron.Scheduler`，可使用 gocron 全部能力（分布式锁、选举、事件监听、监控等）：

```go
raw := s.Scheduler()
// raw 为 gocron.Scheduler 接口，可调用 gocron 原生方法
```

### 动态移除任务

```go
job, _ := s.AddEvery(time.Minute, fn)
// 稍后移除
s.RemoveJob(job)
```

### 扩展调度器选项

需要分布式锁、选举、并发限制等时，可扩展 `cron.New` 或直接使用 gocron：

```go
import "github.com/go-co-op/gocron/v2"

// 方式一：直接使用 gocron
s, _ := gocron.NewScheduler(
    gocron.WithDistributedLocker(myLocker),      // 分布式锁
    gocron.WithDistributedElector(myElector),   // 主从选举
    gocron.WithLimitConcurrentJobs(5, gocron.LimitModeWait), // 全局并发限制
    gocron.WithLogger(gocron.NewLogger(log.Default())),
)
// 然后用 s.NewJob(...) 添加任务

// 方式二：用 cron 封装后，通过 Scheduler() 获取的为同一实例，但创建时的选项需在 New 阶段传入
```

### gocron 能力速览

| 能力 | gocron 选项 | 说明 |
|------|-------------|------|
| 分布式锁 | `WithDistributedLocker` | 多实例时任务仅在一台执行 |
| 主从选举 | `WithDistributedElector` | 多实例时仅主节点跑任务 |
| 全局并发限制 | `WithLimitConcurrentJobs` | 限制调度器同时运行任务数 |
| 任务事件 | `WithEventListeners` | 监听 JobStarted/JobCompleted/JobFailed |
| 监控指标 | `WithSchedulerMonitor` | 对接 Prometheus 等 |
| 时间 Mock | `WithClock` | 单元测试用 FakeClock |

详见 [gocron 文档](https://pkg.go.dev/github.com/go-co-op/gocron/v2)。
