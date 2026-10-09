package search

import "time"

const (
	TierHot  = "hot"
	TierCold = "cold"
)

type GlobalSearchResponseDto struct {
	Query      string              `json:"query"`
	Suggestion string              `json:"suggestion,omitempty"`
	Fuzzy      bool                `json:"fuzzy,omitempty"`
	Files      []FileResultDto     `json:"files"`
	Folders    []FolderResultDto   `json:"folders"`
	Artists    []ArtistResultDto   `json:"artists"`
	Albums     []AlbumResultDto    `json:"albums"`
	Playlists  []PlaylistResultDto `json:"playlists"`
	Videos     []VideoResultDto    `json:"videos"`
	Images     []ImageResultDto    `json:"images"`
	Tracks     []TrackResultDto    `json:"tracks"`
}

type FileResultDto struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parent_path"`
	Format     string    `json:"format"`
	Starred    bool      `json:"starred"`
	Size       int64     `json:"size"`
	UpdatedAt  time.Time `json:"updated_at"`
	Tier       string    `json:"tier"`
}

type FolderResultDto struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parent_path"`
	Starred    bool      `json:"starred"`
	Size       int64     `json:"size"`
	UpdatedAt  time.Time `json:"updated_at"`
	Tier       string    `json:"tier"`
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
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parent_path"`
	Format     string    `json:"format"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ImageResultDto struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parent_path"`
	Format     string    `json:"format"`
	UpdatedAt  time.Time `json:"updated_at"`
	Category   string    `json:"category"`
	Context    string    `json:"context"`
}

type TrackResultDto struct {
	FileID   int     `json:"file_id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album"`
	AlbumKey string  `json:"album_key"`
	Duration float64 `json:"duration"`
	Path     string  `json:"path"`
}

func (response GlobalSearchResponseDto) isEmpty() bool {
	return len(response.Files) == 0 &&
		len(response.Folders) == 0 &&
		len(response.Artists) == 0 &&
		len(response.Albums) == 0 &&
		len(response.Playlists) == 0 &&
		len(response.Videos) == 0 &&
		len(response.Images) == 0 &&
		len(response.Tracks) == 0
}
