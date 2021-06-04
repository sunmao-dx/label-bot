package test

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"testing"
)

func TestCronRun(t *testing.T){
	DoByFixTime()
}

func DoByFixTime() {
	c := cron.New(cron.WithSeconds())
	//定时任务
	spec := "0 2 12 * * *" //cron表达式，每秒一次
	c.AddFunc(spec, func() {
		fmt.Println("11111")
	})
	c.Start()
}
