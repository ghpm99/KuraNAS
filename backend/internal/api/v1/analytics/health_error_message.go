package analytics

import "nas-go/api/pkg/i18n"

const healthErrorGenericKey = "ANALYTICS_HEALTH_ERROR_GENERIC"

var healthErrorKeyByStepType = map[string]string{
	"scan_filesystem":          "ANALYTICS_HEALTH_ERROR_SCAN",
	"diff_against_db":          "ANALYTICS_HEALTH_ERROR_SCAN",
	"mark_deleted":             "ANALYTICS_HEALTH_ERROR_SCAN",
	"metadata":                 "ANALYTICS_HEALTH_ERROR_METADATA",
	"image_metadata_reconcile": "ANALYTICS_HEALTH_ERROR_METADATA",
	"checksum":                 "ANALYTICS_HEALTH_ERROR_CHECKSUM",
	"persist":                  "ANALYTICS_HEALTH_ERROR_PERSIST",
	"thumbnail":                "ANALYTICS_HEALTH_ERROR_THUMBNAIL",
	"playlist_index":           "ANALYTICS_HEALTH_ERROR_PLAYLIST",
	"ai_playlist_cluster":      "ANALYTICS_HEALTH_ERROR_PLAYLIST",
	"takeout_extract":          "ANALYTICS_HEALTH_ERROR_TAKEOUT",
	"backup_run":               "ANALYTICS_HEALTH_ERROR_BACKUP",
	"tier_migration":           "ANALYTICS_HEALTH_ERROR_TIER_MIGRATION",
	"capture_promote":          "ANALYTICS_HEALTH_ERROR_CAPTURE",
	"remote_fetch":             "ANALYTICS_HEALTH_ERROR_DOWNLOAD",
	"ollama_model_pull":        "ANALYTICS_HEALTH_ERROR_AI",
	"image_classify_enumerate": "ANALYTICS_HEALTH_ERROR_AI",
	"image_classify_batch":     "ANALYTICS_HEALTH_ERROR_AI",
	"email_fetch":              "ANALYTICS_HEALTH_ERROR_EMAIL",
	"email_prefilter":          "ANALYTICS_HEALTH_ERROR_EMAIL",
	"email_analyze":            "ANALYTICS_HEALTH_ERROR_EMAIL",
}

func translateHealthErrorStepType(stepType string) string {
	messageKey, isKnownStepType := healthErrorKeyByStepType[stepType]
	if !isKnownStepType {
		return i18n.GetMessage(healthErrorGenericKey)
	}
	return i18n.GetMessage(messageKey)
}
