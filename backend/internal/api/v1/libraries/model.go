package libraries

import "time"

type LibraryCategory string

const (
	LibraryCategoryImages    LibraryCategory = "images"
	LibraryCategoryMusic     LibraryCategory = "music"
	LibraryCategoryVideos    LibraryCategory = "videos"
	LibraryCategoryDocuments LibraryCategory = "documents"
)

var AllCategories = []LibraryCategory{
	LibraryCategoryImages,
	LibraryCategoryMusic,
	LibraryCategoryVideos,
	LibraryCategoryDocuments,
}

func (c LibraryCategory) IsValid() bool {
	switch c {
	case LibraryCategoryImages, LibraryCategoryMusic, LibraryCategoryVideos, LibraryCategoryDocuments:
		return true
	default:
		return false
	}
}

type LibraryModel struct {
	ID        int
	Category  LibraryCategory
	Path      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c LibraryCategory) LabelKey() string {
	switch c {
	case LibraryCategoryImages:
		return "LIBRARY_IMAGES"
	case LibraryCategoryMusic:
		return "LIBRARY_MUSIC"
	case LibraryCategoryVideos:
		return "LIBRARY_VIDEOS"
	case LibraryCategoryDocuments:
		return "LIBRARY_DOCUMENTS"
	default:
		return ""
	}
}
