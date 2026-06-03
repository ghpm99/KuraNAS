package queries

import (
	_ "embed"
)

//go:embed get_ai_providers.sql
var GetAIProvidersQuery string

//go:embed get_ai_provider_by_name.sql
var GetAIProviderByNameQuery string

//go:embed insert_ai_provider_if_absent.sql
var InsertAIProviderIfAbsentQuery string

//go:embed update_ai_provider.sql
var UpdateAIProviderQuery string
