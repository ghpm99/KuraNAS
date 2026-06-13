package queries

import (
	_ "embed"
)

//go:embed get_tiering_settings.sql
var GetTieringSettingsQuery string

//go:embed upsert_tiering_settings.sql
var UpsertTieringSettingsQuery string

//go:embed list_demotion_candidates.sql
var ListDemotionCandidatesQuery string

//go:embed list_promotion_candidates.sql
var ListPromotionCandidatesQuery string

//go:embed set_physical_path.sql
var SetPhysicalPathQuery string

//go:embed get_last_tiering_job.sql
var GetLastTieringJobQuery string

//go:embed get_tier_counts.sql
var GetTierCountsQuery string
