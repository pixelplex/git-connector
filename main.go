package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func CheckIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func UpdateGitlab(repo *git.Repository, ref string) {
	w, err := repo.Worktree()
	CheckIfError(err)

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})

	if err != git.NoErrAlreadyUpToDate {
		CheckIfError(err)
	}

	branch := strings.TrimPrefix(ref, "refs/heads/")
	finalRef := "+refs/remotes/origin/" + branch + ":" + ref

	err = repo.Push(&git.PushOptions{RemoteName: "gitlab", RefSpecs: []config.RefSpec{config.RefSpec(finalRef)}})
	CheckIfError(err)
}

func main() {
	var params ServerParams
	if err := params.ParseParams(); err != nil {
		panic("cannot parse params:" + err.Error())
	}

	repo, err := git.PlainClone(params.DirPath, false, &git.CloneOptions{
		URL:      params.GitHubUrl,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		repo, err = git.PlainOpen(params.DirPath)
		CheckIfError(err)
	} else {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "gitlab",
			URLs: []string{params.GitLabUrl},
		})

		if err != git.ErrRemoteExists {
			CheckIfError(err)
		}
	}

	hook := github.New(&github.Config{Secret: params.Secret})

	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		pl := payload.(github.PushPayload)

		UpdateGitlab(repo, pl.Ref)
	}, github.PushEvent)

	err = webhooks.Run(hook, fmt.Sprintf(":%s", params.Port), "/webhook")
	CheckIfError(err)

}