package queries

import (
	_ "embed"
)

//go:embed list_accounts.sql
var ListAccountsQuery string

//go:embed get_account_by_id.sql
var GetAccountByIDQuery string

//go:embed upsert_account.sql
var UpsertAccountQuery string

//go:embed update_account_tokens.sql
var UpdateAccountTokensQuery string

//go:embed update_account_sync_enabled.sql
var UpdateAccountSyncEnabledQuery string

//go:embed delete_account.sql
var DeleteAccountQuery string

//go:embed update_account_last_sync.sql
var UpdateAccountLastSyncQuery string

//go:embed insert_message.sql
var InsertMessageQuery string

//go:embed list_messages.sql
var ListMessagesQuery string

//go:embed list_pending_messages.sql
var ListPendingMessagesQuery string

//go:embed update_message_prefilter.sql
var UpdateMessagePrefilterQuery string

//go:embed purge_messages_before.sql
var PurgeMessagesBeforeQuery string

//go:embed list_messages_for_analysis.sql
var ListMessagesForAnalysisQuery string

//go:embed upsert_analysis.sql
var UpsertAnalysisQuery string

//go:embed update_message_analyzed.sql
var UpdateMessageAnalyzedQuery string

//go:embed get_analysis_by_message.sql
var GetAnalysisByMessageQuery string
