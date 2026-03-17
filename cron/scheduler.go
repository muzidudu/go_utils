// Package cron 基于 gocron 的定时任务库
package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	s gocron.Scheduler
}

// Option 调度器选项
type Option func(*schedulerOpts)

type schedulerOpts struct {
	logger gocron.Logger
}

// WithLogger 设置日志
func WithLogger(logger gocron.Logger) Option {
	return func(o *schedulerOpts) {
		o.logger = logger
	}
}

// New 创建调度器
func New(opts ...Option) (*Scheduler, error) {
	var cfg schedulerOpts
	for _, opt := range opts {
		opt(&cfg)
	}

	schedulerOpts := []gocron.SchedulerOption{}
	if cfg.logger != nil {
		schedulerOpts = append(schedulerOpts, gocron.WithLogger(cfg.logger))
	}

	s, err := gocron.NewScheduler(schedulerOpts...)
	if err != nil {
		return nil, err
	}
	return &Scheduler{s: s}, nil
}

// JobOption 任务选项
type JobOption func([]gocron.JobOption) []gocron.JobOption

// WithName 设置任务名称
func WithName(name string) JobOption {
	return func(opts []gocron.JobOption) []gocron.JobOption {
		return append(opts, gocron.WithName(name))
	}
}

// WithSingletonMode 单例模式，防止任务重叠执行
func WithSingletonMode(mode gocron.LimitMode) JobOption {
	return func(opts []gocron.JobOption) []gocron.JobOption {
		return append(opts, gocron.WithSingletonMode(mode))
	}
}

// WithSingletonReschedule 单例模式，重叠时跳过本次
func WithSingletonReschedule() JobOption {
	return WithSingletonMode(gocron.LimitModeReschedule)
}

// WithSingletonQueue 单例模式，重叠时排队等待
func WithSingletonQueue() JobOption {
	return WithSingletonMode(gocron.LimitModeWait)
}

// WithLimitedRuns 限制执行次数
func WithLimitedRuns(n int) JobOption {
	return func(opts []gocron.JobOption) []gocron.JobOption {
		return append(opts, gocron.WithLimitedRuns(uint(n)))
	}
}

// WithIntervalFromCompletion 从任务完成时开始计算下次执行时间
func WithIntervalFromCompletion() JobOption {
	return func(opts []gocron.JobOption) []gocron.JobOption {
		return append(opts, gocron.WithIntervalFromCompletion())
	}
}

// AddEvery 添加固定间隔任务
// 示例: AddEvery(5*time.Second, fn)
func (s *Scheduler) AddEvery(d time.Duration, fn func(), jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(
		gocron.DurationJob(d),
		gocron.NewTask(fn),
		opts...,
	)
}

// AddCron 添加 Cron 表达式任务
// 示例: AddCron("0 0 * * *", fn) 每天 0 点
func (s *Scheduler) AddCron(expr string, fn func(), jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(
		gocron.CronJob(expr, true),
		gocron.NewTask(fn),
		opts...,
	)
}

// AddDaily 添加每日任务
// 示例: AddDaily(9, 0, 0, fn) 每天 9:00:00
func (s *Scheduler) AddDaily(hour, min, sec uint, fn func(), jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(hour, min, sec))),
		gocron.NewTask(fn),
		opts...,
	)
}

// AddWeekly 添加每周任务
// 示例: AddWeekly(time.Sunday, 9, 0, 0, fn) 每周日 9:00:00
func (s *Scheduler) AddWeekly(weekday time.Weekday, hour, min, sec uint, fn func(), jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(weekday), gocron.NewAtTimes(gocron.NewAtTime(hour, min, sec))),
		gocron.NewTask(fn),
		opts...,
	)
}

// AddOneTime 添加一次性任务
// 示例: AddOneTime(time.Now().Add(time.Hour), fn)
func (s *Scheduler) AddOneTime(t time.Time, fn func(), jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
		gocron.NewTask(fn),
		opts...,
	)
}

// Add 添加自定义任务（直接使用 gocron 定义）
func (s *Scheduler) Add(jobDef gocron.JobDefinition, task gocron.Task, jobOpts ...JobOption) (gocron.Job, error) {
	opts := applyJobOpts(jobOpts)
	return s.s.NewJob(jobDef, task, opts...)
}

// RemoveJob 移除任务
func (s *Scheduler) RemoveJob(job gocron.Job) error {
	return s.s.RemoveJob(job.ID())
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.s.Start()
}

// Shutdown 关闭调度器
func (s *Scheduler) Shutdown() error {
	return s.s.Shutdown()
}

// Scheduler 返回底层 gocron.Scheduler，用于高级用法
func (s *Scheduler) Scheduler() gocron.Scheduler {
	return s.s
}

func applyJobOpts(jobOpts []JobOption) []gocron.JobOption {
	var opts []gocron.JobOption
	for _, opt := range jobOpts {
		opts = opt(opts)
	}
	return opts
}
