package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	githubactions "github.com/sethvargo/go-githubactions"
)

type RagClient struct {
	Host   string
	Port   string
	Branch string
}

func NewRagClient(host string, port string, branch string) *RagClient {
	return &RagClient{
		Host:   host,
		Port:   port,
		Branch: branch,
	}
}

// GetHost returns the host of the RAG client.
func (c *RagClient) GetIndexedDocuments(fileNames []string) ([]*RagDocument, error) {
	metadataFilter := map[string]string{
		"branch": c.Branch,
	}
	if len(fileNames) == 0 {
		return nil, nil
	}

	respDocs := []*RagDocument{}
	for _, fileName := range fileNames {
		metadataFilter["file_name"] = fileName
		filterBytes, err := json.Marshal(&metadataFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata filter: %w", err)
		}

		resp, err := http.Get(fmt.Sprintf("http://%s:%s/indexes/%s/documents?metadata_fileter=%s", c.Host, c.Port, c.Branch, string(filterBytes)))
		if err != nil {
			return nil, err
		}

		var listDocResponse ListRagDocumentResponse
		if err := json.NewDecoder(resp.Body).Decode(&listDocResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()
		respJson, _ := json.Marshal(listDocResponse)
		githubactions.Infof("Response JSON: %s", string(respJson))
		githubactions.Infof("Documents found for file %s: %+v", fileName, listDocResponse.Documents)

		respDocs = append(respDocs, listDocResponse.Documents...)
	}

	return respDocs, nil
}

func (c *RagClient) UpdateDocuments(documents []*RagDocument) (*UpdateDocumentResponse, error) {
	for _, doc := range documents {
		doc.Metadata["branch"] = c.Branch
	}

	updateRequest := UpdateDocumentRequest{
		Documents: documents,
	}

	updateBytes, err := json.Marshal(&updateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://%s:%s/indexes/%s/documents", c.Host, c.Port, c.Branch), "application/json", bytes.NewBuffer(updateBytes))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update documents: %s", resp.Status)
	}
	defer resp.Body.Close()

	var updateResponse UpdateDocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updateResponse, nil
}

func (c *RagClient) CreateIndex(documents []*RagDocument) ([]*RagDocument, error) {

	for _, doc := range documents {
		doc.Metadata["branch"] = c.Branch
	}

	createRequest := CreateIndexRequest{
		IndexName: c.Branch,
		Documents: documents,
	}

	createBytes, err := json.Marshal(createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create index request: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://%s:%s/index", c.Host, c.Port), "application/json", bytes.NewBuffer(createBytes))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create index: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response []*RagDocument
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

func (c *RagClient) CheckIfIndexExists() (bool, error) {
	indexes, err := c.ListIndexs()
	if err != nil {
		return false, err
	}
	for _, index := range indexes {
		if strings.EqualFold(index, c.Branch) {
			return true, nil
		}
	}
	return false, nil
}

func (c *RagClient) ListIndexs() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/indexes", c.Host, c.Port))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get index documents: %s", resp.Status)
	}
	defer resp.Body.Close()

	var indexes []string
	if err := json.NewDecoder(resp.Body).Decode(&indexes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return indexes, nil
}
