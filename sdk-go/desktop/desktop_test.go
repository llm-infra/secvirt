package desktop

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestSandboxRun(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	res, err := sbx.PackageInstall(t.Context(),
		sandbox.PackageInstallRequest{
			PackageType: sandbox.PackageTypeArchive,
			PackageName: "file:PublicCMS-5.zip<c747811a77ed03a6678f1e1fef6cc058>",
			Destination: "",
		})
	assert.NoError(t, err)

	_, err = sbx.PackageInstall(t.Context(),
		sandbox.PackageInstallRequest{
			PackageType: sandbox.PackageTypeFile,
			PackageName: "file:agent-config.yaml<9c43b6b07536d851b597c15c0a1d85ed>",
			Destination: filepath.Join(res.RelativePath, ".leo"),
		})
	assert.NoError(t, err)

	_, err = sbx.PackageInstall(t.Context(),
		sandbox.PackageInstallRequest{
			PackageType: sandbox.PackageTypeFile,
			PackageName: "file:settings.json<4ee7bdf37f7d097141276a051eff2212>",
			Destination: filepath.Join(res.RelativePath, ".leo"),
		})
	assert.NoError(t, err)

	cli, err := sbx.NewLeo(t.Context(), 8003)
	assert.NoError(t, err)

	card, err := cli.GetAgentCard(t.Context(), "")
	assert.NoError(t, err)

	fmt.Println("Agent Name:", card.Name)
	fmt.Println("Agent Description:", card.Description)
}
