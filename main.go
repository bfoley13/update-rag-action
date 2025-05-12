package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	githubactions "github.com/sethvargo/go-githubactions"
)

func main() {
	ragHost := githubactions.GetInput("ragHost")
	if ragHost == "" {
		githubactions.Fatalf("ragHost is required")
	}

	ragPort := githubactions.GetInput("ragPort")
	if ragPort == "" {
		githubactions.Fatalf("ragPort is required")
	}

	branch := githubactions.GetInput("branch")
	if branch == "" {
		githubactions.Fatalf("branch is required")
	}

	token := githubactions.GetInput("token")
	if token == "" {
		githubactions.Fatalf("token is required")
	}

	gitHubSha := os.Getenv("GITHUB_SHA")
	githubRepo := os.Getenv("GITHUB_REPOSITORY")
	githubRepoOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	gitHubHeadRef := os.Getenv("GITHUB_HEAD_REF")
	gitHubBaseRef := os.Getenv("GITHUB_BASE_REF")
	gitHubRef := os.Getenv("GITHUB_REF")
	gitHubRefName := os.Getenv("GITHUB_REF_NAME")
	gitHubEnv := os.Getenv("GITHUB_ENV")
	gitHubSetupSummary := os.Getenv("GITHUB_SETUP_SUMMARY")

	githubactions.Infof("GITHUB_REPOSITORY: %s", githubRepo)
	githubactions.Infof("GITHUB_REPOSITORY_OWNER: %s", githubRepoOwner)
	githubactions.Infof("GITHUB_HEAD_REF: %s", gitHubHeadRef)
	githubactions.Infof("GITHUB_BASE_REF: %s", gitHubBaseRef)
	githubactions.Infof("GITHUB_REF: %s", gitHubRef)
	githubactions.Infof("GITHUB_REF_NAME: %s", gitHubRefName)
	githubactions.Infof("GITHUB_ENV: %s", gitHubEnv)
	githubactions.Infof("GITHUB_SETUP_SUMMARY: %s", gitHubSetupSummary)

	if githubRepo == "" {
		githubactions.Fatalf("GITHUB_REPOSITORY is required")
	}
	if githubRepoOwner == "" {
		githubactions.Fatalf("GITHUB_REPOSITORY_OWNER is required")
	}

	ghClient := NewGitHubClient(githubRepoOwner, githubRepo, branch, token)

	githubactions.Infof("ragHost: %s | ragPort: %s | branch: %s", ragHost, ragPort, branch)
	ragClient := NewRagClient(ragHost, ragPort, branch)
	indexExists, err := ragClient.CheckIfIndexExists()
	if err != nil {
		githubactions.Fatalf("failed to check if index exists: %v", err)
	}

	if !indexExists {
		githubactions.Infof("Index does not exist, creating index")
		createIndex(ragClient)
	} else {
		githubactions.Infof("Index already exists, updating index")
		updatedFiles, err := getUpdatedFiles(ghClient)
		if err != nil {
			githubactions.Fatalf("failed to get updated files: %v", err)
		}

		if len(updatedFiles) == 0 {
			githubactions.Infof("No updated files found")
			return
		}

		updateIndex(ragClient, updatedFiles)
		githubactions.Infof("Index updated successfully")
	}

	docs, err := ragClient.GetIndexDocuments()
	if err != nil {
		githubactions.Fatalf("failed to get index documents: %v", err)
	}

	githubactions.Infof("Index documents retrieved successfully")
	githubactions.Infof("Documents: %v", docs)
	githubactions.Infof("Document count: %d", len(docs))
}

func createIndex(ragClient *RagClient) {
	documents := []*RagDocument{}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		githubactions.Infof("Processing file: %s", path)

		fileBytes, err := os.ReadFile(path)
		if err != nil {
			githubactions.Fatalf("failed to read file: %v", err)
		}

		fileContent := string(fileBytes)
		documents = append(documents, &RagDocument{
			Text:     fileContent,
			Metadata: map[string]string{"file_path": path, "file_name": d.Name()},
		})

		return nil
	})

	if err != nil {
		githubactions.Fatalf("failed to create documents for index create: %v", err)
	}

	_, err = ragClient.CreateIndex(documents)
	if err != nil {
		githubactions.Fatalf("failed to create index: %v", err)
	}

	githubactions.Infof("Index created successfully")
}

func getUpdatedFiles(ghClient *GitHubClient) ([]string, error) {
	files := []string{}
	files, err := ghClient.GetCommitFiles(context.Background(), os.Getenv("GITHUB_SHA"))
	if err != nil {
		githubactions.Fatalf("failed to get commit files: %v", err)
		return nil, err
	}
	githubactions.Infof("Updated files: %v", files)
	return files, nil
}

func updateIndex(ragClient *RagClient, updatedFiles []string) {
	documents := []*RagDocument{}
	for _, file := range updatedFiles {
		githubactions.Infof("Processing file: %s", file)
		if !strings.HasSuffix(file, ".go") {
			githubactions.Infof("Skipping file: %s", file)
			continue
		}

		fileBytes, err := os.ReadFile(file)
		if err != nil {
			githubactions.Fatalf("failed to read file: %v", err)
		}

		fileContent := string(fileBytes)
		documents = append(documents, &RagDocument{
			Text:     fileContent,
			Metadata: map[string]string{"file_path": file, "file_name": filepath.Base(file)},
		})
	}

	updateResponse, err := ragClient.UpdateDocuments(documents)
	if err != nil {
		githubactions.Fatalf("failed to update index: %v", err)
	}
	githubactions.Infof("Index updated successfully")
	githubactions.Infof("Updated documents: %v", updateResponse.UpdatedDocuments)
	githubactions.Infof("Unchanged documents: %v", updateResponse.UnchangedDocuments)
	githubactions.Infof("Not found documents: %v", updateResponse.NotFoundDocuments)
}
