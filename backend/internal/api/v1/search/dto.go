package search

type GlobalSearchResponseDto struct {
	Query      string              `json:"query"`
	Suggestion string              `json:"suggestion,omitempty"`
	Files      []FileResultDto     `json:"files"`
	Folders    []FolderResultDto   `json:"folders"`
	Artists    []ArtistResultDto   `json:"artists"`
	Albums     []AlbumResultDto    `json:"albums"`
	Playlists  []PlaylistResultDto `json:"playlists"`
	Videos     []VideoResultDto    `json:"videos"`
	Images     []ImageResultDto    `json:"images"`
}

type FileResultDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Format     string `json:"format"`
	Starred    bool   `json:"starred"`
}

type FolderResultDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Starred    bool   `json:"starred"`
}

type ArtistResultDto struct {
	Key        string `json:"key"`
	Artist     string `json:"artist"`
	TrackCount int    `json:"track_count"`
	AlbumCount int    `json:"album_count"`
}

type AlbumResultDto struct {
	Key        string `json:"key"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	Year       string `json:"year"`
	TrackCount int    `json:"track_count"`
}

type PlaylistResultDto struct {
	Scope          string `json:"scope"`
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Count          int    `json:"count"`
	Classification string `json:"classification"`
	SourcePath     string `json:"source_path"`
	IsAuto         bool   `json:"is_auto"`
}

type VideoResultDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Format     string `json:"format"`
}

type ImageResultDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Format     string `json:"format"`
	Category   string `json:"category"`
	Context    string `json:"context"`
}
