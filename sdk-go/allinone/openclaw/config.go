package openclaw

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"unsafe"

	"github.com/a3tai/openclaw-go/identity"
)

type Config struct {
	Gateway *GatewayConfig `json:"gateway,omitempty"`
	Models  *ModelsConfig  `json:"models,omitempty"`
	Agents  *AgentsConfig  `json:"agents,omitempty"`
}

type GatewayConfig struct {
	Port int `json:"port"`
	Auth struct {
		Token string `json:"token"`
	} `json:"auth"`
}

type ModelsConfig struct {
	Mode      string                   `json:"mode,omitempty"`
	Providers map[string]ModelProvider `json:"providers"`
}

const (
	APIOpenAICompletions = "openai-completions"
)

type ModelProvider struct {
	BaseURL string        `json:"baseUrl"`
	APIKey  string        `json:"apiKey"`
	API     string        `json:"api"`
	Models  []ModelConfig `json:"models"`
}

type ModelConfig struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	API           string     `json:"api"`
	Reasoning     bool       `json:"reasoning"`
	Input         []string   `json:"input,omitempty"`
	Cost          *ModelCost `json:"cost,omitempty"`
	ContextWindow float64    `json:"contextWindow,omitempty"`
	MaxTokens     float64    `json:"maxTokens,omitempty"`
}

type ModelCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

type AgentsConfig struct {
	Defaults *AgentDefaults  `json:"defaults,omitempty"`
	List     []AgentListItem `json:"list,omitempty"`
}

type AgentDefaults struct {
	Model *ModelRef `json:"model,omitempty"`
}

type AgentListItem struct {
	ID        string    `json:"id,omitempty"`
	Default   bool      `json:"default,omitempty"`
	Name      string    `json:"name,omitempty"`
	Workspace string    `json:"workspace,omitempty"`
	AgentDir  string    `json:"agentDir,omitempty"`
	Model     *ModelRef `json:"model,omitempty"`
}

type ModelRef struct {
	Value     string   `json:"-"`
	Primary   string   `json:"primary,omitempty"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

func (m *ModelRef) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		m.Value = value
		m.Primary = value
		m.Fallbacks = nil
		return nil
	}

	type alias ModelRef
	var model alias
	if err := json.Unmarshal(data, &model); err != nil {
		return err
	}
	*m = ModelRef(model)
	return nil
}

func LoadConfig(data []byte) (*Config, error) {
	var config Config
	return &config, json.Unmarshal(data, &config)
}

type IdentityDevice struct {
	DeviceID      string `json:"deviceId"`
	PublicKeyPEM  string `json:"publicKeyPem"`
	PrivateKeyPEM string `json:"privateKeyPem"`
}

type DevicePaired struct {
	DeviceID       string   `json:"deviceId"`
	Platform       string   `json:"platform"`
	ClientID       string   `json:"clientId"`
	ClientMode     string   `json:"clientMode"`
	ApprovedScopes []string `json:"approvedScopes"`
	Scopes         []string `json:"scopes"`
	Tokens         map[string]struct {
		Token string `json:"token"`
	} `json:"tokens"`
}

func ParseDeviceIdentity(deviceData, pairedData []byte) (*identity.Identity, *DevicePaired, error) {
	var device IdentityDevice
	if err := json.Unmarshal(deviceData, &device); err != nil {
		return nil, nil, fmt.Errorf("parse sandbox openclaw device identity: %w", err)
	}

	var paired map[string]DevicePaired
	if err := json.Unmarshal(pairedData, &paired); err != nil {
		return nil, nil, fmt.Errorf("parse sandbox openclaw paired devices: %w", err)
	}

	pairedDevice, ok := paired[device.DeviceID]
	if !ok {
		return nil, nil, fmt.Errorf("sandbox openclaw device %s is not paired", device.DeviceID)
	}

	block, _ := pem.Decode([]byte(device.PrivateKeyPEM))
	if block == nil {
		return nil, nil, fmt.Errorf("decode sandbox openclaw private key pem")
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse sandbox openclaw private key: %w", err)
	}

	privateKey, ok := privateKeyAny.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("sandbox openclaw private key is not ed25519")
	}

	block, _ = pem.Decode([]byte(device.PublicKeyPEM))
	if block == nil {
		return nil, nil, fmt.Errorf("decode sandbox openclaw public key pem")
	}

	publicKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse sandbox openclaw public key: %w", err)
	}

	publicKey, ok := publicKeyAny.(ed25519.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("sandbox openclaw public key is not ed25519")
	}

	type rawIdentity struct {
		DeviceID        string
		PublicKeyB64URL string
		PrivateKey      ed25519.PrivateKey
	}

	return (*identity.Identity)(unsafe.Pointer(&rawIdentity{
		DeviceID:        device.DeviceID,
		PublicKeyB64URL: base64.RawURLEncoding.EncodeToString(publicKey),
		PrivateKey:      privateKey,
	})), &pairedDevice, nil
}
