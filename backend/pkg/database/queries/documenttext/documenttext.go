package queries

import (
	_ "embed"
)

//go:embed upsert_document_text.sql
var UpsertDocumentTextQuery string

//go:embed select_pending_indexing.sql
var SelectPendingIndexingQuery string

//go:embed search_documents_columns.sql
var SearchDocumentsColumnsQuery string

//go:embed search_documents_filter_match.sql
var SearchDocumentsFilterMatchQuery string

//go:embed search_documents_order.sql
var SearchDocumentsOrderQuery string

//go:embed search_documents_content_hit.sql
var SearchDocumentsContentHitQuery string

//go:embed select_document_for_indexing.sql
var SelectDocumentForIndexingQuery string

//go:embed select_index_state.sql
var SelectIndexStateQuery string

//go:embed list_documents_columns.sql
var ListDocumentsColumnsQuery string

//go:embed count_documents_columns.sql
var CountDocumentsColumnsQuery string

//go:embed documents_base.sql
var DocumentsBaseQuery string

//go:embed documents_filter_folder.sql
var DocumentsFilterFolderQuery string

//go:embed documents_filter_modified_from.sql
var DocumentsFilterModifiedFromQuery string

//go:embed documents_filter_modified_before.sql
var DocumentsFilterModifiedBeforeQuery string

//go:embed documents_filter_starred.sql
var DocumentsFilterStarredQuery string

//go:embed documents_filter_index_state.sql
var DocumentsFilterIndexStateQuery string

//go:embed documents_order_name_asc.sql
var DocumentsOrderNameAscQuery string

//go:embed documents_order_name_desc.sql
var DocumentsOrderNameDescQuery string

//go:embed documents_order_size_asc.sql
var DocumentsOrderSizeAscQuery string

//go:embed documents_order_size_desc.sql
var DocumentsOrderSizeDescQuery string

//go:embed documents_order_modified_at_asc.sql
var DocumentsOrderModifiedAtAscQuery string

//go:embed documents_order_modified_at_desc.sql
var DocumentsOrderModifiedAtDescQuery string

//go:embed documents_order_created_at_asc.sql
var DocumentsOrderCreatedAtAscQuery string

//go:embed documents_order_created_at_desc.sql
var DocumentsOrderCreatedAtDescQuery string

//go:embed documents_page.sql
var DocumentsPageQuery string

//go:embed select_active_folder_path.sql
var SelectActiveFolderPathQuery string

//go:embed select_document_text.sql
var SelectDocumentTextQuery string
