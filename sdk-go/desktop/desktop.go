package desktop

import (
	"bytes"
	"context"
	"errors"
	"path"
	"strings"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/mholt/archiver/v4"
	oc "github.com/sst/opencode-sdk-go"
)

type Sandbox struct {
	*sandbox.Sandbox

	ocHandle *commands.CommandHandle
	ocClient *oc.Client
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts, sandbox.WithTemplate(sandbox.TemplateDesktop))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}

func checkZipRootDir(ctx context.Context, name string, data []byte) (string, bool, error) {
	format, stream, err := archiver.Identify(name, bytes.NewReader(data))
	if err != nil {
		return "", false, err
	}

	extractor, ok := format.(archiver.Extractor)
	if !ok {
		return "", false, errors.New("invalid extractor")
	}

	skillName := archiveBaseName(name)
	firstLevel := make(map[string]struct{})
	hasRootFile := false

	if err = extractor.Extract(ctx, stream, nil, func(ctx context.Context, info archiver.File) error {
		entry := strings.TrimPrefix(path.Clean("/"+info.NameInArchive), "/")
		if entry == "." || entry == "" {
			return nil
		}

		parts := strings.Split(entry, "/")
		if len(parts) == 0 || parts[0] == "" {
			return nil
		}

		firstLevel[parts[0]] = struct{}{}
		if len(parts) == 1 && !info.IsDir() {
			hasRootFile = true
		}
		return nil
	}); err != nil {
		return skillName, false, err
	}

	if hasRootFile || len(firstLevel) != 1 {
		return skillName, false, nil
	}

	_, ok = firstLevel[skillName]
	return skillName, ok, nil
}

func archiveBaseName(name string) string {
	base := path.Base(name)
	for {
		ext := path.Ext(base)
		if ext == "" {
			return base
		}
		base = strings.TrimSuffix(base, ext)
	}
}
