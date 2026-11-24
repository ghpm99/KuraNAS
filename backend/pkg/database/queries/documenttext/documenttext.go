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

//go:embed select_document_for_indexing.sql
var SelectDocumentForIndexingQuery string

//go:embed select_index_state.sql
var SelectIndexStateQuery string
