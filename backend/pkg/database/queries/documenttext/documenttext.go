package queries

import (
	_ "embed"
)

//go:embed upsert_document_text.sql
var UpsertDocumentTextQuery string

//go:embed select_pending_indexing.sql
var SelectPendingIndexingQuery string

//go:embed search_documents.sql
var SearchDocumentsQuery string
