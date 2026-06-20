package queries

import (
	_ "embed"
)

//go:embed upsert_image_metadata.sql
var UpsertImageMetadataQuery string

//go:embed get_image_metadata_by_id.sql
var GetImageMetadataByIDQuery string

//go:embed delete_image_metadata.sql
var DeleteImageMetadataQuery string

//go:embed get_images.sql
var GetImagesQuery string

//go:embed count_pending_ai_classification.sql
var CountPendingAIClassificationQuery string

//go:embed select_pending_ai_classification.sql
var SelectPendingAIClassificationQuery string
