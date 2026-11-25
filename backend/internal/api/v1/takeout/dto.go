package takeout

import "time"

type InitTakeoutUploadDto struct {
	FileName string `json:"file_name" binding:"required"`
	Size     int64  `json:"size"`
}

type InitTakeoutUploadResultDto struct {
	UploadID     string `json:"upload_id"`
	ChunkSize    int64  `json:"chunk_size"`
	ReceivedSize int64  `json:"received_size"`
	Resumed      bool   `json:"resumed"`
}

type UploadTakeoutChunkDto struct {
	UploadID string `json:"upload_id" form:"upload_id" binding:"required"`
	Offset   int64  `json:"offset" form:"offset"`
}

type CompleteTakeoutUploadDto struct {
	UploadID string `json:"upload_id" binding:"required"`
}

type TakeoutImportResultDto struct {
	JobID    int    `json:"job_id"`
	ImportID int    `json:"import_id"`
	Message  string `json:"message"`
}

type TakeoutUploadStatusDto struct {
	UploadID     string `json:"upload_id"`
	Size         int64  `json:"size"`
	ReceivedSize int64  `json:"received_size"`
}

type TakeoutImportDto struct {
	ID               int        `json:"id"`
	JobID            *int       `json:"job_id"`
	FileName         string     `json:"file_name"`
	Size             int64      `json:"size"`
	State            string     `json:"state"`
	TotalEntries     int        `json:"total_entries"`
	ProcessedEntries int        `json:"processed_entries"`
	Imported         int        `json:"imported"`
	AlreadyExisting  int        `json:"already_existing"`
	Ignored          int        `json:"ignored"`
	FailedEntries    int        `json:"failed_entries"`
	Error            string     `json:"error"`
	CreatedAt        time.Time  `json:"created_at"`
	EndedAt          *time.Time `json:"ended_at"`
}
