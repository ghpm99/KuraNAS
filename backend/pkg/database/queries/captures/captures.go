package queries

import (
	_ "embed"
)

//go:embed insert_capture.sql
var InsertCaptureQuery string

//go:embed get_captures.sql
var GetCapturesQuery string

//go:embed get_capture_by_id.sql
var GetCaptureByIDQuery string

//go:embed get_capture_by_episode_key.sql
var GetCaptureByEpisodeKeyQuery string

//go:embed delete_capture.sql
var DeleteCaptureQuery string
