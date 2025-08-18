package codeide

import (
	"context"
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestPackages(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.Packages(context.Background(), "python")
	assert.NoError(t, err)

	fmt.Println(res)

	b, err := sbx.Filesystem().Mkdir(t.Context(), "testpath")
	assert.Equal(t, b, true)
	assert.NoError(t, err)
}

func TestRunCode(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("10.20.152.105"),
	)
	assert.NoError(t, err)

	res, err := sbx.RunCode(context.TODO(), "python", `import os

# 打印当前工作路径
cwd = os.getcwd()
print("当前工作路径:", cwd)

# 创建文件 test
file_path = os.path.join(cwd, "test")
with open(file_path, "w", encoding="utf-8") as f:
    f.write("这是一个测试文件\n")

print(f"文件已创建: {file_path}")`, nil)
	assert.NoError(t, err)

	fmt.Println(res)
}
