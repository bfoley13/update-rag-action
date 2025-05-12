package main

import (
	"io/fs"
	"os"
	"os/exec"
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
		updatedFiles, err := getUpdatedFiles()
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

func getUpdatedFiles() ([]string, error) {
	cmd := exec.Command("ls", "-la")
	output, err := cmd.Output()
	githubactions.Infof("output bytes: %s", string(output))
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("git", "--version", "origin")
	output, err = cmd.Output()
	githubactions.Infof("output bytes: %s", string(output))
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("git", "status")
	output, err = cmd.Output()
	githubactions.Infof("output bytes: %s", string(output))
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("git", "diff", "--name-only", "HEAD", "HEAD~1")
	output, err = cmd.Output()
	githubactions.Infof("output bytes: %s", string(output))
	if err != nil {
		return nil, err
	}

	githubactions.Infof("Updated files: %s", string(output))

	files := strings.Split(string(output), "\n")
	for i := range files {
		files[i] = strings.TrimSpace(files[i])
	}
	return files, nil
}

func updateIndex(ragClient *RagClient, updatedFiles []string) {
	documents := []*RagDocument{}
	for _, file := range updatedFiles {
		githubactions.Infof("Processing file: %s", file)

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
