package openclaw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigModelsCore(t *testing.T) {
	data := []byte(`{
		"models": {
			"mode": "merge",
			"providers": {
				"openai": {
					"baseUrl": "https://api.openai.com/v1",
					"api": "openai-responses",
					"models": [
						{
							"id": "gpt-5",
							"name": "GPT-5",
							"api": "openai-responses",
							"reasoning": true,
							"input": ["text", "image"],
							"cost": {
								"input": 1.25,
								"output": 10.0
							},
							"contextWindow": 200000,
							"maxTokens": 100000
						}
					]
				}
			}
		}
	}`)

	cfg, err := LoadConfig(data)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "merge", cfg.Models.Mode)
	provider, ok := cfg.Models.Providers["openai"]
	require.True(t, ok)
	assert.Equal(t, "https://api.openai.com/v1", provider.BaseURL)
	assert.Equal(t, "openai-responses", provider.API)
	require.Len(t, provider.Models, 1)

	model := provider.Models[0]
	assert.Equal(t, "gpt-5", model.ID)
	assert.Equal(t, "GPT-5", model.Name)
	assert.Equal(t, "openai-responses", model.API)
	assert.True(t, model.Reasoning)
	assert.Equal(t, []string{"text", "image"}, model.Input)
	require.NotNil(t, model.Cost)
	assert.Equal(t, 1.25, model.Cost.Input)
	assert.Equal(t, 10.0, model.Cost.Output)
	assert.Equal(t, 200000.0, model.ContextWindow)
	assert.Equal(t, 100000.0, model.MaxTokens)
}

func TestLoadConfigAgentsCore(t *testing.T) {
	data := []byte(`{
		"agents": {
			"defaults": {
				"model": {
					"primary": "gpt-5",
					"fallbacks": ["gpt-4.1"]
				}
			},
			"list": [
				{
					"id": "default",
					"default": true,
					"name": "Default Agent",
					"workspace": "/workspace",
					"agentDir": "/agents/default",
					"model": "gpt-5"
				}
			]
		}
	}`)

	cfg, err := LoadConfig(data)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Agents)
	require.NotNil(t, cfg.Agents.Defaults)
	require.NotNil(t, cfg.Agents.Defaults.Model)

	assert.Equal(t, "gpt-5", cfg.Agents.Defaults.Model.Primary)
	assert.Equal(t, []string{"gpt-4.1"}, cfg.Agents.Defaults.Model.Fallbacks)
	require.Len(t, cfg.Agents.List, 1)

	agent := cfg.Agents.List[0]
	assert.Equal(t, "default", agent.ID)
	assert.True(t, agent.Default)
	assert.Equal(t, "Default Agent", agent.Name)
	assert.Equal(t, "/workspace", agent.Workspace)
	assert.Equal(t, "/agents/default", agent.AgentDir)
	assert.Equal(t, "gpt-5", agent.Model.Value)
}
