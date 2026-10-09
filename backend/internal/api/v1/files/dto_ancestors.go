package files

type FileAncestorDto struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Path string   `json:"path"`
	Type FileType `json:"type"`
}
