package config

import (
	"github.com/rascals77/consul-init/consul"
)

// Policies defines what ACL policies will be created
type Policies struct {
	Name        string
	Description string
	Rules       string
}

// Tokens defines what ACL tokens will be created
type Tokens struct {
	Name     string
	Policies []string
}

// Config describes the contents of the configuration file parameters
type Config struct {
	TokenFile         string          `mapstructure:"token_file",validate:"required"`
	Address           string          `mapstructure:"address",validate:"required"`
	Port              int             `mapstructure:"port",validate:"required"`
	Scheme            string          `mapstructure:"scheme",validate:"required"`
	CACert            string          `mapstructure:"cacert",validate:"required"`
	Cert              string          `mapstructure:"cert",validate:"required"`
	Key               string          `mapstructure:"key",validate:"required"`
	Members           []consul.Member `validate:"required"`
	NodeAgentTemplate string          `mapstructure:"node_agent_template",validate:"required"`
	Policies          []Policies      `validate:"required"`
	Tokens            []Tokens        `validate:"required"`
	TokenSecretsFile  string          `mapstructure:"token_secrets_file",validate:"required"`
}
