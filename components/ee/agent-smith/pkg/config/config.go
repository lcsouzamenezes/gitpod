package config

import (
	"encoding/json"
	"github.com/gitpod-io/gitpod/agent-smith/pkg/agent"
	"golang.org/x/xerrors"
	"io/ioutil"
)

func GetConfig(cfgFile string) (*Config, error) {
	if cfgFile == "" {
		return nil, xerrors.Errorf("missing --config")
	}

	fc, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, xerrors.Errorf("cannot read config: %v", err)
	}

	var cfg Config
	err = json.Unmarshal(fc, &cfg)
	if err != nil {
		return nil, xerrors.Errorf("cannot unmarshal config: %v", err)
	}

	if cfg.ProbePath == "" {
		cfg.ProbePath = "/app/probe.o"
	}

	return &cfg, nil
}

// Config is the struct holding the configuration for agent-smith
// if you are considering changing this struct, remember
// to update the config schema using:
// $ go run main.go config-schema > config-schema.json
// And also update the examples accordingly.
type Config struct {
	agent.Config

	Namespace string `json:"namespace,omitempty"`

	PProfAddr      string `json:"pprofAddr,omitempty"`
	PrometheusAddr string `json:"prometheusAddr,omitempty"`

	// We have had memory leak issues with agent smith in the past due to experimental gRPC use.
	// This upper limit causes agent smith to stop itself should it go above this limit.
	MaxSysMemMib uint64 `json:"systemMemoryLimitMib,omitempty"`

	HostURL        string `json:"hostURL,omitempty"`
	GitpodAPIToken string `json:"gitpodAPIToken,omitempty"`
}
