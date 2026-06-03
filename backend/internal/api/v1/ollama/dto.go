package ollama

import "time"

// ModelDto describes a model installed on the daemon.
type ModelDto struct {
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	Digest            string    `json:"digest"`
	ModifiedAt        time.Time `json:"modified_at"`
	Family            string    `json:"family,omitempty"`
	ParameterSize     string    `json:"parameter_size,omitempty"`
	QuantizationLevel string    `json:"quantization_level,omitempty"`
}

// StatusDto reports daemon reachability plus installed models.
type StatusDto struct {
	Reachable bool       `json:"reachable"`
	Version   string     `json:"version,omitempty"`
	BaseURL   string     `json:"base_url"`
	Models    []ModelDto `json:"models"`
}

// PullModelRequest is the body for triggering a model download.
type PullModelRequest struct {
	Model string `json:"model" binding:"required"`
}

// PullModelResponse returns the job id tracking the background download.
type PullModelResponse struct {
	JobID int `json:"job_id"`
}
