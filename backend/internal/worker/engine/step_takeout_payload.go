package engine

type TakeoutStepPayload struct {
	ZipPath  string `json:"zip_path"`
	UploadID string `json:"upload_id"`
	FileName string `json:"file_name"`
}
