package main

import (
	"io/fs"
	"path/filepath"

	githubactions "github.com/sethvargo/go-githubactions"
)

func main() {
	// ragHost := githubactions.GetInput("ragHost")
	// if ragHost == "" {
	// 	githubactions.Fatalf("ragHost is required")
	// }

	// ragPort := githubactions.GetInput("ragPort")
	// if ragPort == "" {
	// 	githubactions.Fatalf("ragPort is required")
	// }

	// branch := githubactions.GetInput("branch")
	// if branch == "" {
	// 	githubactions.Fatalf("branch is required")
	// }

	// githubactions.Infof("ragHost: %s | ragPort: %s | branch: %s", ragHost, ragPort, branch)
	// ragClient := NewRagClient(ragHost, ragPort, branch)
	// _, err := ragClient.CheckIfIndexExists()
	// if err != nil {
	// 	githubactions.Fatalf("failed to check if index exists: %v", err)
	// }

	createIndex(nil)
}

func createIndex(ragClient *RagClient) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		githubactions.Infof("Processing file: %s", path)
		return nil
	})
	if err != nil {
		githubactions.Fatalf("failed to walk directory: %v", err)
	}

	githubactions.Infof("Index created successfully")
}
