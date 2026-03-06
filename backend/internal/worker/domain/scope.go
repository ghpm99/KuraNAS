package domain

type ScopeType string

const (
	ScopeTypeFile ScopeType = "file"
	ScopeTypePath ScopeType = "path"
	ScopeTypeRoot ScopeType = "root"
)

type ScopePayload struct {
	Type ScopeType  `json:"type"`
	File *FileScope `json:"file,omitempty"`
	Path *PathScope `json:"path,omitempty"`
	Root *RootScope `json:"root,omitempty"`
}

type FileScope struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type PathScope struct {
	Path string `json:"path"`
}

type RootScope struct {
	Root string `json:"root"`
}

func NewFileScopePayload(file FileScope) ScopePayload {
	return ScopePayload{
		Type: ScopeTypeFile,
		File: &file,
	}
}

func NewPathScopePayload(path string) ScopePayload {
	return ScopePayload{
		Type: ScopeTypePath,
		Path: &PathScope{Path: path},
	}
}

func NewRootScopePayload(root string) ScopePayload {
	return ScopePayload{
		Type: ScopeTypeRoot,
		Root: &RootScope{Root: root},
	}
}
