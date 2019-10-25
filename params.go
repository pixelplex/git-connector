package main

import (
	"errors"
	"strings"
	"strconv"

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

	AppID           int
	InstallID       int
}

// Names of params
const (
	portParam = "port"
	secretParam = "secret"

	GitHubUrlParam = "github"
	GitLabUrlParam  = "gitlab"

	DirPathParam = "local-path"
	PrivkeyParam = "privkey"

	AppIDParam = "app-id"
	InstallIDParam = "installation-id"
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

	AppIDParam: &flagData{Value: "44467", Help: "Application ID from GitHub App"},
	InstallIDParam: &flagData{Value: "3771539", Help: "Installation ID"},
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

	appID, err := strconv.Atoi(configFlags[AppIDParam].Value)
	CheckIfError(err)
	
	instID, err := strconv.Atoi(configFlags[InstallIDParam].Value)
	CheckIfError(err)

	sp.AppID = appID
	sp.InstallID = instID
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

	if err := checkAllKeys(); err != nil {
		return err
	}

	sp.initServerParams()

	return nil
}
