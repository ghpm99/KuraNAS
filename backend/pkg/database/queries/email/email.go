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
