package search

import "time"

type FileResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Starred    bool
	Size       int64
	UpdatedAt  time.Time
	IsCold     bool
}

type FolderResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Starred    bool
	Size       int64
	UpdatedAt  time.Time
	IsCold     bool
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
	UpdatedAt  time.Time
}

type ImageResultModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	UpdatedAt  time.Time
	Category   string
	Context    string
}

type TrackResultModel struct {
	FileID     int
	Title      string
	Artist     string
	Album      string
	AlbumOwner string
	Duration   float64
	Path       string
}
