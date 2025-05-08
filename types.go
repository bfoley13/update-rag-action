package main

type RagDocument struct {
	DocumentId  string            `json:"doc_id"`
	Text        string            `json:"text"`
	HashValue   string            `json:"hash_value"`
	Metadata    map[string]string `json:"metadata"`
	IsTruncated bool              `json:"is_truncated"`
}

type ListRagDocumentResponse struct {
	Documents []*RagDocument `json:"documents"`
	Count     int            `json:"count"`
}

type CreateIndexRequest struct {
	IndexName string         `json:"index_name"`
	Documents []*RagDocument `json:"documents"`
}

type UpdateDocumentRequest struct {
	Documents []*RagDocument `json:"documents"`
}

type UpdateDocumentResponse struct {
	UpdatedDocuments   []*RagDocument `json:"updated_documents"`
	UnchangedDocuments []*RagDocument `json:"unchanged_documents"`
	NotFoundDocuments  []*RagDocument `json:"not_found_documents"`
}
