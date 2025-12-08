package desktop

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
	"github.com/stretchr/testify/assert"
)

func TestLeoChat(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("10.50.10.18"),
		sandbox.WithUser("mel2oo"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	// res, err := sbx.PackageInstall(
	// 	t.Context(),
	// 	sandbox.PackageInstallRequest{
	// 		PackageType: sandbox.PackageTypeArchive,
	// 		PackageName: "file:PublicCMS-5.zip<333>",
	// 		Destination: "PublicCMS-5",
	// 	},
	// )
	// assert.NoError(t, err)

	// _, err = sbx.Filesystem().Mkdir(
	// 	t.Context(),
	// 	filepath.Join(res.UserPath, res.RelativePath, ".leo"))
	// assert.NoError(t, err)

	// config, _ := os.ReadFile("data/agent-config.yaml")
	// err = sbx.Filesystem().Write(t.Context(),
	// 	filepath.Join(res.UserPath, res.RelativePath, ".leo", "agent-config.yaml"),
	// 	[]byte(config))
	// assert.NoError(t, err)

	// settings, _ := os.ReadFile("data/settings.json")
	// err = sbx.Filesystem().Write(t.Context(),
	// 	filepath.Join(res.UserPath, res.RelativePath, ".leo", "settings.json"),
	// 	[]byte(settings))
	// assert.NoError(t, err)

	stream, err := sbx.LeoChat(t.Context(), "你好")
	if !assert.NoError(t, err) {
		return
	}
	defer stream.Close()

	for {
		res, err := stream.Recv()
		if err != nil {
			break
		}

		fmt.Println(json.MarshalString(res))
	}
}
