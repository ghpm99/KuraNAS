package takeout

type InitTakeoutUploadDto struct {
	FileName string `json:"file_name" binding:"required"`
	Size     int64  `json:"size"`
}

type InitTakeoutUploadResultDto struct {
	UploadID  string `json:"upload_id"`
	ChunkSize int64  `json:"chunk_size"`
}

type UploadTakeoutChunkDto struct {
	UploadID string `json:"upload_id" form:"upload_id" binding:"required"`
	Offset   int64  `json:"offset" form:"offset"`
}

type CompleteTakeoutUploadDto struct {
	UploadID string `json:"upload_id" binding:"required"`
}

type TakeoutImportResultDto struct {
	JobID   int    `json:"job_id"`
	Message string `json:"message"`
}
