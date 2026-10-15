package files

type StarredCountsDto struct {
	Total   int `json:"total"`
	Folders int `json:"folders"`
	Files   int `json:"files"`
	Media   int `json:"media"`
}
