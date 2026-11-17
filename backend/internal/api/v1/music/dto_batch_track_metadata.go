package music

type BatchMetadataFailureDto struct {
	FileID int    `json:"file_id"`
	Error  string `json:"error"`
}

type BatchMetadataResultDto struct {
	Matched int                       `json:"matched"`
	Updated int                       `json:"updated"`
	Failed  []BatchMetadataFailureDto `json:"failed"`
}
