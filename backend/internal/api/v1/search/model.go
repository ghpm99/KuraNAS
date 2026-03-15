package search

import "time"

type FileResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Starred    bool
}

type FolderResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Starred    bool
}

type ArtistResultModel struct {
	Artist     string
	TrackCount int
	AlbumCount int
}

type AlbumResultModel struct {
	Artist     string
	Album      string
	Year       string
	TrackCount int
}

type MusicPlaylistResultModel struct {
	ID          int
	Name        string
	Description string
	IsSystem    bool
	UpdatedAt   time.Time
	TrackCount  int
}

type VideoPlaylistResultModel struct {
	ID             int
	Name           string
	Type           string
	Classification string
	SourcePath     string
	IsAuto         bool
	UpdatedAt      time.Time
	ItemCount      int
}

type VideoResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
}

type ImageResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Category   string
	Context    string
}
