package main

import (
	"context"

	"github.com/google/go-github/v72/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	RepoOwner string
	RepoName  string
	Branch    string
	client    *github.Client
}

func NewGitHubClient(repoOwner, repoName, branch, token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return &GitHubClient{
		RepoOwner: repoOwner,
		RepoName:  repoName,
		Branch:    branch,
		client:    client,
	}
}

func (c *GitHubClient) GetFilesInPR(ctx context.Context, prNum int) ([]string, error) {
	files := []string{}
	commits, _, err := c.client.PullRequests.ListCommits(ctx, c.RepoOwner, c.RepoName, prNum, nil)
	if err != nil {
		return nil, err
	}

	for _, commit := range commits {
		for _, file := range commit.Files {
			if file.Filename != nil && *file.Filename != "" {
				files = append(files, *file.Filename)
			}
		}
	}

	return files, nil
}

func (c *GitHubClient) GetCommitFiles(ctx context.Context, commitSHA string) ([]string, error) {
	files := []string{}
	commit, _, err := c.client.Repositories.GetCommit(ctx, c.RepoOwner, c.RepoName, commitSHA, nil)
	if err != nil {
		return nil, err
	}

	githubactions.Infof("Commit: %+v", *commit)

	for _, file := range commit.Files {
		if file.Filename != nil && *file.Filename != "" {
			files = append(files, *file.Filename)
		}
	}

	return files, nil
}
