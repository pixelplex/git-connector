package main

import (
	"errors"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type flagData struct {
	Value string
	Help  string
}

// ServerConfig - configuration for server
type ServerParams struct {
	Port            string
	Secret          string
	GitHubUrl       string
	GitLabUrl       string
	DirPath         string
}

// Names of params
const (
	portParam = "port"
	secretParam = "secret"

	GitHubUrlParam = "github"
	GitLabUrlParam  = "gitlab"

	DirPathParam = "local-path"
)

// Errors
const (
	configKeyError = "Unsupported key in params: "
)

var configFlags map[string]*flagData = map[string]*flagData{
	portParam: &flagData{Value: "3000", Help: "server port"},
	secretParam: &flagData{Value: "MySecret", Help: "secret of github webhook"},

	GitHubUrlParam: &flagData{Value: "https://github.com/alex-suslikov/github-mirroing", Help: "GitHub URL for mirroring"},
	GitLabUrlParam:  &flagData{Value: "git@gitlab.com:Cipa_Joe/github-mirroing.git", Help: "GitLab URL for mirroring"},

	DirPathParam: &flagData{Value: "/tmp/repo/", Help: "path to local directory"},
}

func (sp *ServerParams) initServerParams() {
	sp.Port       = configFlags[portParam].Value
	sp.Secret     = configFlags[secretParam].Value
	sp.GitHubUrl  = configFlags[GitHubUrlParam].Value
	sp.GitLabUrl  = configFlags[GitLabUrlParam].Value
	sp.DirPath    = configFlags[DirPathParam].Value
}

func checkAllKeys() error {
	for key, value := range viper.AllSettings() {
		switch value.(type) {
		case string:
			if _, ok := configFlags[key]; !ok {
				return errors.New(configKeyError + key)
			}
		case map[string]interface{}:
			submap := value.(map[string]interface{})
			for subkey := range submap {
				fullKey := key + "." + subkey
				if _, ok := configFlags[fullKey]; !ok {
					return errors.New(configKeyError + fullKey)
				}
			}
		default:
			return errors.New(configKeyError + key)
		}
	}
	return nil
}

// ParseParams - parse params from command line
func (sp *ServerParams) ParseParams() error {
	for k, v := range configFlags {
		flag.StringVar(&v.Value, k, v.Value, v.Help)
	}

	flag.Parse()

	viper.BindPFlags(flag.CommandLine)

	viper.SupportedExts = viper.SupportedExts[:0]
	viper.SupportedExts = append(viper.SupportedExts, "yml")

	if err := checkAllKeys(); err != nil {
		return err
	}

	sp.initServerParams()

	return nil
}
