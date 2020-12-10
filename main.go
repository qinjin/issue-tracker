package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tink-ab/github-devstats/bazel-github-devstats/external/com_github_google_go_github/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

const owner = "qinjin"
const repoName = "issue-tracker"
const label = "tech-pains"

type Reactions struct {
	TotalCount *int `json:"total_count,omitempty"`
	PlusOne    *int `json:"+1,omitempty"`
	MinusOne   *int `json:"-1,omitempty"`
	Laugh      *int `json:"laugh,omitempty"`
	Confused   *int `json:"confused,omitempty"`
	Heart      *int `json:"heart,omitempty"`
	Hooray     *int `json:"hooray,omitempty"`
}

type Issue struct {
	Number    int        `json:"number,omitempty"`
	Title     string     `json:"title,omitempty"`
	Label     string     `json:"label,omitempty"`
	URL       string     `json:"url,omitempty"`
	Reactions *Reactions `json:"reactions,omitempty"`
}

type RepoStatus struct {
	Name   string  `json:"name,omitempty"`
	URL    string  `json:"url,omitempty"`
	Issues []Issue `json:"issues,omitempty"`
}

func main() {
	log.Println("Starting HTTP server on port 8090")
	http.HandleFunc("/status", handleStatusRequest)
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatalf("can not start HTTP server %s", err)
	}
}

func handleStatusRequest(writer http.ResponseWriter, request *http.Request) {
	repoStatus, err := getRepoStatus(owner, repoName, label)
	if err != nil {
		log.Fatalf("can not get repo status %s", err)
	}

	//output, _ := json.MarshalIndent(repoStatus, "", "    ")
	//log.Println(string(output))

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(repoStatus)
}

func getRepoStatus(owner string, repoName string, label string) (RepoStatus, error) {
	log.Println("Accessing github...")
	if accessToken, ok := os.LookupEnv("GITHUB_ACCESS_TOKEN"); ok {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		repo, _, err := client.Repositories.Get(ctx, owner, repoName)
		if err != nil {
			log.Fatalf("Could not get repo %s", err)
		}

		log.Println("Found", *repo.OpenIssuesCount, "open issues from repo:", *repo.FullName)
		var issues []Issue
		for i := 1; i <= *repo.OpenIssuesCount; i++ {
			issue, _, err := client.Issues.Get(ctx, owner, repoName, i)
			if err != nil {
				log.Fatalf("could not list issue %d due to %s", i, err)
			}

			// Skip issues without specified label.
			if containsLabel(label, issue.Labels) != ok {
				log.Println("Skip issue", *issue.Number, "as it doesn't contain label:", label)
				continue
			}

			issues = append(issues, Issue{
				Number: *issue.Number,
				Title:  *issue.Title,
				Label:  label,
				URL:    *issue.URL,
				Reactions: &Reactions{
					TotalCount: issue.Reactions.TotalCount,
					PlusOne:    issue.Reactions.PlusOne,
					MinusOne:   issue.Reactions.MinusOne,
					Laugh:      issue.Reactions.Laugh,
					Confused:   issue.Reactions.Confused,
					Heart:      issue.Reactions.Heart,
					Hooray:     issue.Reactions.Hooray,
				},
			})
		}

		repoStatus := RepoStatus{
			Name:   *repo.FullName,
			URL:    *repo.URL,
			Issues: issues,
		}

		return repoStatus, nil
	}
	return RepoStatus{}, fmt.Errorf("can not access repo %s", repoName)
}

func containsLabel(expectedLabel string, labels []github.Label) bool {
	for _, label := range labels {
		if expectedLabel == *label.Name {
			return true
		}
	}
	return false
}
