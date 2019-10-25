package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"context"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"

	gitHubHook "gopkg.in/go-playground/webhooks.v5/github"
	gitLabHook "gopkg.in/go-playground/webhooks.v5/gitlab"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func CheckIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func UpdateGitlab(repo *git.Repository) {
	w, err := repo.Worktree()
	CheckIfError(err)

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})

	if err != git.NoErrAlreadyUpToDate {
		CheckIfError(err)
	}

	finalRef := "+refs/remotes/origin/*:refs/heads/*"

	err = repo.Push(&git.PushOptions{RemoteName: "gitlab", RefSpecs: []config.RefSpec{config.RefSpec(finalRef)}})
	CheckIfError(err)
}

func main() {
	var params ServerParams
	if err := params.ParseParams(); err != nil {
		panic("cannot parse params:" + err.Error())
	}

	context := context.Background()

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int64(params.AppID), int64(params.InstallID), params.PrivKeyPath)
	CheckIfError(err)

	client := github.NewClient(&http.Client{Transport: itr})

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

	GitHubHook, _ := gitHubHook.New(gitHubHook.Options.Secret(params.Secret))
	GitLabHook, _ := gitLabHook.New(gitLabHook.Options.Secret(params.Secret))

	mapCheckRun := make(map[string]int64)
	conclDict := make(map[string]string)

	conclDict["success"] = "success"
	conclDict["canceled"] = "cancelled"
	conclDict["failed"] = "failure"
	conclDict["skipped"] = "cancelled"

	http.HandleFunc("/gitlabhooks", func(w http.ResponseWriter, r *http.Request) {
		payload, err := GitLabHook.Parse(r, gitLabHook.PipelineEvents)

		if err != nil {
			if err == gitLabHook.ErrEventNotFound {
				fmt.Println("EventNotFound")
			}
		}

		switch payload.(type) {

		case gitLabHook.PipelineEventPayload:
			pl := payload.(gitLabHook.PipelineEventPayload)

			if pl.ObjectAttributes.Status != "pending" && pl.ObjectAttributes.Status != "running" {
				opt := github.UpdateCheckRunOptions{Name: "CheckRun", Status: github.String("completed"), Conclusion: github.String(conclDict[pl.ObjectAttributes.Status])}
	
				id := mapCheckRun[pl.ObjectAttributes.SHA]
				client.Checks.UpdateCheckRun(context, params.Owner, params.Repository, id, opt)		
			}

		}
	})

	http.HandleFunc("/githubhooks", func(w http.ResponseWriter, r *http.Request) {
		payload, err := GitHubHook.Parse(r, gitHubHook.PushEvent, gitHubHook.CheckSuiteEvent)

		if err != nil {
			if err == gitHubHook.ErrEventNotFound {
				fmt.Println("EventNotFound")
			}
		}

		switch payload.(type) {

		case gitHubHook.CheckSuitePayload:
			pl := payload.(gitHubHook.CheckSuitePayload)
			if pl.Action == "requested" {
				opt := github.CreateCheckRunOptions{Name: "CheckRun", HeadSHA: pl.CheckSuite.HeadSHA, Status: github.String("in_progress")}
				checkRun, _, _ := client.Checks.CreateCheckRun(context, params.Owner, params.Repository, opt)
				mapCheckRun[pl.CheckSuite.HeadSHA] = *checkRun.ID
			}
		case gitHubHook.PushPayload:
			UpdateGitlab(repo)
		}
	})

	http.ListenAndServe(":" +  params.Port, nil)

}