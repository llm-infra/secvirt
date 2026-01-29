package gemini

import (
	"context"
	"fmt"

	"github.com/mel2oo/go-dkit/ext"
)

type Config map[string]string

func NewConfig(ctx context.Context, model, baseURL string) Config {
	extv := ext.FromContextValue(ctx)

	return Config{
		"GOOGLE_GEMINI_BASE_URL":    baseURL,
		"GEMINI_MODEL":              model,
		"GEMINI_API_KEY":            "skip",
		"GEMINI_CLI_CUSTOM_HEADERS": fmt.Sprintf("EXT: %s", extv.ToString()),
	}
}
