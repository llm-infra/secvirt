package codeide

type PackagesResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type JupyterOutput struct {
	Type      string `json:"type,omitempty"`
	Timestamp any    `json:"timestamp,omitempty"`
	Data      any    `json:"data,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
	Traceback string `json:"traceback,omitempty"`
}

type RunCodeResponseV1 struct {
	Output  string `json:"output,omitempty"`
	Console string `json:"console,omitempty"`
}

type RunCodeResponse struct {
	Result  any             `json:"result,omitempty"`
	Errors  []JupyterOutput `json:"errors,omitempty"`
	Stdouts []JupyterOutput `json:"stdouts,omitempty"`
	Stderrs []JupyterOutput `json:"stderrs,omitempty"`
}
