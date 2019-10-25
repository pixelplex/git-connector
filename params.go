package main

import (
	"errors"
	"strings"

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
	PrivKeyPath     string
	Owner           string
	Repository      string
}

// Names of params
const (
	portParam = "port"
	secretParam = "secret"

	GitHubUrlParam = "github"
	GitLabUrlParam  = "gitlab"

	DirPathParam = "local-path"
	PrivkeyParam = "privkey"
)

// Errors
const (
	configKeyError = "Unsupported key in params: "
)

var configFlags map[string]*flagData = map[string]*flagData{
	portParam: &flagData{Value: "3000", Help: "server port"},
	secretParam: &flagData{Value: "MySecret", Help: "secret of github webhook"},

	GitHubUrlParam: &flagData{Value: "git@github.com:alex-suslikov/registrator.git", Help: "GitHub URL for mirroring"},
	GitLabUrlParam:  &flagData{Value: "git@gitlab.pixelplex.by:a.suslikov/registrator.git", Help: "GitLab URL for mirroring"},

	DirPathParam: &flagData{Value: "/tmp/repo/", Help: "path to local directory"},

	PrivkeyParam: &flagData{Value: "privkey.pem", Help: "path to private key file"},
}

func (sp *ServerParams) initServerParams() {
	sp.Port       = configFlags[portParam].Value
	sp.Secret     = configFlags[secretParam].Value
	sp.GitHubUrl  = configFlags[GitHubUrlParam].Value
	sp.GitLabUrl  = configFlags[GitLabUrlParam].Value
	sp.DirPath    = configFlags[DirPathParam].Value
	sp.PrivKeyPath    = configFlags[PrivkeyParam].Value

	str := strings.TrimPrefix(configFlags[GitHubUrlParam].Value, "git@github.com:")
	str = strings.TrimSuffix(str, ".git")

	strArr := strings.Split(str, "/")

	sp.Owner = strArr[0]
	sp.Repository = strArr[1]
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
