package allinone

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/a3tai/openclaw-go/chatcompletions"
	"github.com/a3tai/openclaw-go/gateway"
	"github.com/a3tai/openclaw-go/identity"
	"github.com/a3tai/openclaw-go/protocol"
	"github.com/llm-infra/secvirt/sdk-go/allinone/openclaw"
	"github.com/llm-infra/secvirt/sdk-go/desktop"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/json"
	"github.com/sirupsen/logrus"
)

const (
	openClawConfigDir   = "/opt/openclaw-config"
	openClawUserDirName = ".openclaw"
	openClawConfigFile  = "openclaw.json"
	openClawDeviceFile  = "identity/device.json"
	openClawPairedFile  = "devices/paired.json"
)

var ErrOpenClawNotRunning = errors.New("openclaw is not running")

type Claw struct {
	port           int
	token          string
	deviceIdentity *identity.Identity
	deviceToken    string
	clientInfo     protocol.ClientInfo
	scopes         []protocol.Scope

	pid uint32
}

func (s *Sandbox) InitOpenClaw(ctx context.Context, opts ...desktop.Option) error {
	configDir := filepath.Join(s.HomeDir(), openClawUserDirName)

	// 检查配置
	exist, err := s.Filesystem().Exist(ctx, configDir)
	if err != nil {
		var connectErr *connect.Error
		if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeNotFound {
			return err
		}
	}

	if !exist {
		if _, err := s.Cmd().Run(
			ctx,
			fmt.Sprintf("cp -r %s %s", openClawConfigDir, configDir),
			nil,
			"",
			false,
		); err != nil {
			return err
		}
	}

	// 从配置读取Port、Token、Device信息
	claw, err := s.loadOpenClawRuntime(ctx)
	if err != nil {
		return err
	}
	s.claw = *claw

	// 通过端口检查OpenClaw是否已经在运行
	// 如果已经运行了，初始化handle、client、chatClient
	ok, err := s.attachOpenClawIfRunning(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return s.RestartOpenClaw(ctx, opts...)
	}
	return nil
}

func (s *Sandbox) RestartOpenClaw(ctx context.Context, opts ...desktop.Option) error {
	opt := desktop.NewOptions(s.HomeDir())
	for _, o := range opts {
		o(opt)
	}

	if s.claw.pid != 0 {
		s.Cmd().Run(ctx, fmt.Sprintf("kill -9 %d", s.claw.pid), nil, "", false)
		s.claw.pid = 0
	}

	handle, err := s.Cmd().Start(ctx,
		"openclaw gateway",
		opt.Envs(),
		opt.Cwd(),
		opt.Stdin())
	if err != nil {
		return err
	}

	once := sync.Once{}
	success := make(chan struct{})
	go func(pid uint32) {
		handle.Wait(ctx,
			commands.WithStdout(
				func(b []byte) {
					once.Do(func() { success <- struct{}{} })
					logrus.WithContext(ctx).
						Debugf("gateway %d, stdout: %s", pid, string(b))
				},
			),
			commands.WithStderr(
				func(b []byte) {
					logrus.WithContext(ctx).
						Errorf("gateway %d, stderr: %s", pid, string(b))
				},
			),
		)

		logrus.WithContext(ctx).Debugf("gateway %d, logger exit", pid)
	}(handle.Pid())

	<-success
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			client, err := s.ClawClient(ctx)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			defer client.Close()
			return nil
		}
	}
}

func (s *Sandbox) ReloadOpenClaw(ctx context.Context, opts ...desktop.Option) error {
	if s.claw.pid == 0 {
		return ErrOpenClawNotRunning
	}

	_, err := s.Cmd().Run(ctx,
		fmt.Sprintf("kill -USR1 %d", s.claw.pid),
		nil, "", false,
	)
	if err != nil {
		return s.RestartOpenClaw(ctx)
	}
	return nil
}

func (s *Sandbox) loadOpenClawRuntime(ctx context.Context) (*Claw, error) {
	configDir := filepath.Join(s.HomeDir(), openClawUserDirName)

	configData, err := s.Filesystem().Read(ctx, filepath.Join(configDir, openClawConfigFile))
	if err != nil {
		return nil, fmt.Errorf("read sandbox openclaw config: %w", err)
	}

	config, err := openclaw.LoadConfig(configData)
	if err != nil {
		return nil, err
	}

	deviceData, err := s.Filesystem().Read(ctx, filepath.Join(configDir, openClawDeviceFile))
	if err != nil {
		return nil, fmt.Errorf("read sandbox openclaw device identity: %w", err)
	}

	pairedData, err := s.Filesystem().Read(ctx, filepath.Join(configDir, openClawPairedFile))
	if err != nil {
		return nil, fmt.Errorf("read sandbox openclaw paired devices: %w", err)
	}

	identity, device, err := openclaw.ParseDeviceIdentity(deviceData, pairedData)
	if err != nil {
		return nil, err
	}

	deviceToken := ""
	if token, ok := device.Tokens["operator"]; ok {
		deviceToken = token.Token
	}

	scopeStrs := device.ApprovedScopes
	if len(scopeStrs) == 0 {
		scopeStrs = device.Scopes
	}
	scopes := make([]protocol.Scope, 0, len(scopeStrs))
	for _, scope := range scopeStrs {
		scopes = append(scopes, protocol.Scope(scope))
	}

	return &Claw{
		port:           config.Gateway.Port,
		token:          config.Gateway.Auth.Token,
		deviceIdentity: identity,
		deviceToken:    deviceToken,
		clientInfo: protocol.ClientInfo{
			ID:       device.ClientID,
			Version:  "0.1.0",
			Platform: device.Platform,
			Mode:     device.ClientMode,
		},
		scopes: scopes,
	}, nil
}

func (s *Sandbox) attachOpenClawIfRunning(ctx context.Context) (bool, error) {
	pid, ok, _ := s.findListeningPID(ctx, s.claw.port)
	if !ok {
		return false, nil
	}

	s.claw.pid = pid

	client, err := s.ClawClient(ctx)
	if err != nil {
		return false, nil
	}
	defer client.Close()

	return true, nil
}

func (s *Sandbox) findListeningPID(ctx context.Context, port int) (uint32, bool, error) {
	var listeningPIDPattern = regexp.MustCompile(`pid=(\d+)`)

	res, err := s.Cmd().Run(ctx,
		fmt.Sprintf("ss -ltnp 'sport = :%d'", port),
		nil, "", false,
	)
	if err != nil && res == nil {
		return 0, false, err
	}

	matches := listeningPIDPattern.FindSubmatch([]byte(res.Stdout))
	if len(matches) != 2 {
		return 0, false, nil
	}

	var pid uint32
	if _, err := fmt.Sscanf(string(matches[1]), "%d", &pid); err != nil {
		return 0, false, fmt.Errorf("parse listening pid: %w", err)
	}
	return pid, true, nil
}

/*******************************/
/********** 客户端管理 **********/
/*******************************/
func (s *Sandbox) ClawClient(ctx context.Context) (*gateway.Client, error) {
	client := gateway.NewClient(
		gateway.WithToken(s.claw.token),
		gateway.WithIdentity(s.claw.deviceIdentity, s.claw.deviceToken),
		gateway.WithClientInfo(s.claw.clientInfo),
		gateway.WithRole(protocol.RoleOperator),
		gateway.WithScopes(s.claw.scopes...),
		gateway.WithHeaders(
			spec.GenSandboxHeader(s.claw.port, s.Name, ""),
		),
	)

	return client, client.Connect(ctx, fmt.Sprintf("ws://%s:%d", s.ProxyHost, s.ProxyPort))
}

func (s *Sandbox) ChatClient(opts ...ClawOption) *chatcompletions.Client {
	opt := &ClawOptions{}
	for _, o := range opts {
		o(opt)
	}

	return &chatcompletions.Client{
		BaseURL: s.ProxyBaseURL(),
		Token:   s.claw.token,
		HTTPClient: &http.Client{
			Transport: spec.NewHeaderRoundTripper(
				spec.GenSandboxHeader(s.claw.port, s.Name, ""),
				http.DefaultTransport,
			),
		},
		AgentID:    opt.agentID,
		SessionKey: opt.sessionKey,
	}
}

type ClawOption func(*ClawOptions)

type ClawOptions struct {
	client     *gateway.Client
	sessionKey string
	agentID    string
}

func WithClawClient(client *gateway.Client) ClawOption {
	return func(co *ClawOptions) { co.client = client }
}

func WithSessionKey(sessionKey string) ClawOption {
	return func(co *ClawOptions) { co.sessionKey = sessionKey }
}

func WithAgentID(agentID string) ClawOption {
	return func(co *ClawOptions) { co.agentID = agentID }
}

/*******************************/
/*********** 配置管理 ***********/
/*******************************/
type ConfigResult struct {
	Hash   string          `json:"hash"`
	Config openclaw.Config `json:"config"`
}

func (s *Sandbox) GetConfig(ctx context.Context, opts ...ClawOption) (*ConfigResult, error) {
	opt := &ClawOptions{}
	for _, o := range opts {
		o(opt)
	}

	if opt.client == nil {
		client, err := s.ClawClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
		}
		defer client.Close()

		opt.client = client
	}

	var res ConfigResult
	configData, err := opt.client.ConfigGet(ctx)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configData, &res); err != nil {
		return nil, fmt.Errorf("unmarshal config get response: %w", err)
	}

	return &res, nil
}

/******************************/
/*********** 会话管理 **********/
/******************************/
func (s *Sandbox) GetSessions(ctx context.Context) (any, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SessionsList(ctx, protocol.SessionsListParams{})
}

func (s *Sandbox) GetSession(ctx context.Context, sessionID string, limit int) (any, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SessionsGet(ctx, protocol.SessionsGetParams{
		Key:   sessionID,
		Limit: &limit,
	})
}

func (s *Sandbox) PatchSession(ctx context.Context, params protocol.SessionsPatchParams) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SessionsPatch(ctx, params)
}

func (s *Sandbox) ResetSession(ctx context.Context, sessionID string) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SessionsReset(ctx, protocol.SessionsResetParams{
		Key: sessionID,
	})
}

func (s *Sandbox) DeleteSession(ctx context.Context, sessionID string) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	client.SessionsPatch(ctx, protocol.SessionsPatchParams{})

	return client.SessionsDelete(ctx, protocol.SessionsDeleteParams{
		Key: sessionID,
	})
}

/*******************************/
/********** 智能体管理 **********/
/*******************************/

func (s *Sandbox) GetAgents(ctx context.Context) (*protocol.AgentsListResult, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.AgentsList(ctx)
}

/*******************************/
/*********** 模型管理 ***********/
/*******************************/

// 读取模型
func (s *Sandbox) GetModels(ctx context.Context) (map[string]openclaw.ModelProvider, error) {
	result, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	if result.Config.Models == nil || len(result.Config.Models.Providers) == 0 {
		return map[string]openclaw.ModelProvider{}, nil
	}

	return result.Config.Models.Providers, nil
}

func (s *Sandbox) SetModel(ctx context.Context, provider string, model openclaw.ModelProvider) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	result, err := s.GetConfig(ctx, WithClawClient(client))
	if err != nil {
		return err
	}

	if len(model.API) == 0 {
		model.API = openclaw.APIOpenAICompletions
	}

	for i, v := range model.Models {
		if len(v.API) == 0 {
			model.Models[i].API = model.API
		}
	}

	return client.ConfigPatch(ctx, protocol.ConfigPatchParams{
		BaseHash: result.Hash,
		Raw: json.MarshalString(
			openclaw.Config{
				Models: &openclaw.ModelsConfig{
					Providers: map[string]openclaw.ModelProvider{
						provider: model,
					},
				},
			},
		),
	})
}

// func (s *Sandbox) DeleteModel(ctx context.Context, provider string) error {
// 	result, err := s.GetConfig(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	delete(result.Config.Models.Providers, provider)
// 	return s.claw.client.ConfigPatch(ctx, protocol.ConfigPatchParams{
// 		BaseHash: result.Hash,
// 		Raw: json.MarshalString(
// 			openclaw.Config{
// 				Models: &openclaw.ModelsConfig{
// 					Providers: result.Config.Models.Providers,
// 				},
// 			},
// 		),
// 	})
// }

func (s *Sandbox) SetDefaultModel(ctx context.Context, modelRef openclaw.ModelRef) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	result, err := s.GetConfig(ctx, WithClawClient(client))
	if err != nil {
		return err
	}

	for p, v := range result.Config.Models.Providers {
		for _, m := range v.Models {
			if fmt.Sprintf("%s/%s", p, m.ID) == modelRef.Primary {
				return client.ConfigPatch(ctx, protocol.ConfigPatchParams{
					BaseHash: result.Hash,
					Raw: json.MarshalString(
						openclaw.Config{
							Agents: &openclaw.AgentsConfig{
								Defaults: &openclaw.AgentDefaults{
									Model: &modelRef,
								},
							},
						},
					),
				})
			}
		}
	}

	return fmt.Errorf("model %s not found in config", modelRef.Primary)
}

/*******************************/
/*********** 技能管理 ***********/
/*******************************/
func (s *Sandbox) GetSkills(ctx context.Context, agentID string) (any, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SkillsStatus(ctx, protocol.SkillsStatusParams{
		AgentID: agentID,
	})
}

func (s *Sandbox) UpdateSkill(ctx context.Context, params protocol.SkillsUpdateParams) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.SkillsUpdate(ctx, params)
}

// func (s *Sandbox) InstallSkill(ctx context.Context, req protocol.SkillsInstallParams) (any, error) {
// 	client, err := s.ClawClient(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
// 	}
// 	defer client.Close()

// 	return client.SkillsInstall(ctx, req)

// }

/*******************************/
/*********** 频道管理 ***********/
/*******************************/
// func (s *Sandbox) GetChannels(ctx context.Context) error {
// 	client, err := s.ClawClient(ctx)
// 	if err != nil {
// 		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
// 	}
// 	defer client.Close()

// }

/*******************************/
/*********** 定时任务 ***********/
/*******************************/
func (s *Sandbox) GetCrons(ctx context.Context) (any, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	includeDisabled := true
	return client.CronList(ctx, protocol.CronListParams{
		IncludeDisabled: &includeDisabled,
	})
}

func (s *Sandbox) AddCron(ctx context.Context, params protocol.CronAddParams) (any, error) {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.CronAdd(ctx, params)
}

func (s *Sandbox) UpdateCron(ctx context.Context, params protocol.CronUpdateParams) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.CronUpdate(ctx, params)
}

func (s *Sandbox) DeleteCron(ctx context.Context, id string) error {
	client, err := s.ClawClient(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenClawNotRunning, err)
	}
	defer client.Close()

	return client.CronRemove(ctx, protocol.CronRemoveParams{
		ID: id,
	})
}
