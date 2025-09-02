package codeide

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestPackages(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	for range 100 {
		time.Sleep(time.Second)

		_, err = sbx.Packages(context.Background(), "python")
		if assert.NoError(t, err) {
			fmt.Println("success")
		} else {
			fmt.Println("error", err)
		}
	}
}

func TestSandboxConcurrency(t *testing.T) {
	const (
		concurrency = 1   // 并发数
		iterations  = 200 // 每个 goroutine 执行次数
	)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		success   int
		fail      int
		durations []time.Duration
	)

	start := time.Now()
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				begin := time.Now()

				sbx, err := NewSandbox(
					context.TODO(),
					sandbox.WithHost("192.168.134.142"),
				)
				if err != nil {
					if strings.Contains(err.Error(), "container waitting") {
						time.Sleep(time.Second) // 等待再试
						continue
					}
					mu.Lock()
					fail++
					mu.Unlock()
					continue
				}

				_, err = sbx.RunCode(context.TODO(), "python", `
import os
cwd = os.getcwd()
print("当前工作路径:", cwd)
file_path = os.path.join(cwd, "test")
with open(file_path, "w", encoding="utf-8") as f:
    f.write("这是一个测试文件\n")
print(f"文件已创建: {file_path}")
`, nil)

				mu.Lock()
				if err == nil {
					success++
					durations = append(durations, time.Since(begin))
				} else {
					fail++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	total := time.Since(start)

	// 统计结果
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("并发数: %d, 每个协程执行: %d 次\n", concurrency, iterations)
	fmt.Printf("成功次数: %d, 失败次数: %d\n", success, fail)
	fmt.Printf("总耗时: %v\n", total)

	if len(durations) > 0 {
		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(len(durations))
		fmt.Printf("平均单次执行耗时: %v\n", avg)
		fmt.Printf("QPS: %.2f\n", float64(success)/total.Seconds())
	}
}
