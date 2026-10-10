package video

type PlaylistSection string

const (
	PlaylistSectionSeries   PlaylistSection = "series"
	PlaylistSectionMovies   PlaylistSection = "movies"
	PlaylistSectionPersonal PlaylistSection = "personal"
	PlaylistSectionClips    PlaylistSection = "clips"
	PlaylistSectionFolders  PlaylistSection = "folders"
)

type PlaylistSectionFilter struct {
	Classifications []string
	PlaylistType    string
}

type PlaylistSectionRequest struct {
	Section  PlaylistSection
	Page     int
	PageSize int
}

func ParsePlaylistSection(rawSection string) (PlaylistSection, bool) {
	switch PlaylistSection(rawSection) {
	case PlaylistSectionSeries, PlaylistSectionMovies, PlaylistSectionPersonal, PlaylistSectionClips, PlaylistSectionFolders:
		return PlaylistSection(rawSection), true
	}
	return "", false
}

func (section PlaylistSection) Filter() PlaylistSectionFilter {
	switch section {
	case PlaylistSectionSeries:
		return PlaylistSectionFilter{Classifications: []string{"series", "anime"}}
	case PlaylistSectionMovies:
		return PlaylistSectionFilter{Classifications: []string{"movie"}}
	case PlaylistSectionPersonal:
		return PlaylistSectionFilter{Classifications: []string{"personal"}}
	case PlaylistSectionClips:
		return PlaylistSectionFilter{Classifications: []string{"clip", "program"}}
	case PlaylistSectionFolders:
		return PlaylistSectionFilter{Classifications: []string{}, PlaylistType: "folder"}
	}
	return PlaylistSectionFilter{Classifications: []string{}}
}
