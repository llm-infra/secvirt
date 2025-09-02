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

func TestRunCode(t *testing.T) {
	wg := sync.WaitGroup{}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				sbx, err := NewSandbox(
					context.TODO(),
					sandbox.WithHost("192.168.134.142"),
				)
				if !assert.NoError(t, err) {
					if strings.Contains(err.Error(), "container waitting") {
						fmt.Println("waitting:", err)
						time.Sleep(time.Second)
					} else {
						fmt.Println("create fail:", err)
					}
					continue
				}

				_, err = sbx.RunCode(context.TODO(), "python", `import os

# 打印当前工作路径
cwd = os.getcwd()
print("当前工作路径:", cwd)

# 创建文件 test
file_path = os.path.join(cwd, "test")
with open(file_path, "w", encoding="utf-8") as f:
    f.write("这是一个测试文件\n")

print(f"文件已创建: {file_path}")`, nil)
				if assert.NoError(t, err) {
					fmt.Println("success")
				} else {
					fmt.Println("error:", err)
				}
			}
		}()
	}
	wg.Wait()
}
