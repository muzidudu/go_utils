package cron_test

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/muzidudu/go_utils/cron"
)

func ExampleNew() {
	s, err := cron.New()
	if err != nil {
		panic(err)
	}
	defer s.Shutdown()

	// 每 1 秒执行
	_, _ = s.AddEvery(time.Second, func() {
		fmt.Println("tick")
	})

	s.Start()
	time.Sleep(2500 * time.Millisecond) // 触发约 2 次
	// Output:
	// tick
	// tick
}

func ExampleScheduler_AddCron() {
	s, _ := cron.New()
	defer s.Shutdown()

	// 每天 0 点执行（cron 表达式）
	_, _ = s.AddCron("0 0 * * *", func() {
		fmt.Println("midnight")
	}, cron.WithName("daily-cleanup"))

	s.Start()
}

func ExampleScheduler_AddDaily() {
	s, _ := cron.New()
	defer s.Shutdown()

	// 每天 9:00:00 执行
	_, _ = s.AddDaily(9, 0, 0, func() {
		fmt.Println("morning job")
	})

	s.Start()
}

func ExampleScheduler_AddWeekly() {
	s, _ := cron.New()
	defer s.Shutdown()

	// 每周日 9:00 执行
	_, _ = s.AddWeekly(time.Sunday, 9, 0, 0, func() {
		fmt.Println("weekly report")
	})

	s.Start()
}

func ExampleScheduler_AddEvery() {
	s, _ := cron.New()
	defer s.Shutdown()

	// 单例模式：防止任务重叠执行
	_, _ = s.AddEvery(time.Minute, func() {
		// 长时间任务
	}, cron.WithSingletonReschedule(), cron.WithName("sync-job"))

	s.Start()
}

func ExampleScheduler_Add() {
	s, _ := cron.New()
	defer s.Shutdown()

	// 使用 gocron 原生 JobDefinition：随机间隔 2–5 秒
	_, _ = s.Add(
		gocron.DurationRandomJob(2*time.Second, 5*time.Second),
		gocron.NewTask(func() { fmt.Println("random tick") }),
		cron.WithName("random-job"),
	)

	s.Start()
	time.Sleep(6 * time.Second) // 随机间隔内至少触发 1 次
}
