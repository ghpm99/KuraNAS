package video

import "testing"

func TestPickSeriesPosterFollowsPosterFolderCoverPriority(t *testing.T) {
	candidates := []PosterCandidateModel{
		{ID: 1, Name: "cover.jpg"},
		{ID: 2, Name: "Folder.PNG"},
		{ID: 3, Name: "POSTER.webp"},
		{ID: 4, Name: "random.jpg"},
	}

	poster, isFound := pickSeriesPoster(candidates)
	if !isFound || poster.ID != 3 {
		t.Fatalf("poster = %+v found=%v, want id 3", poster, isFound)
	}

	poster, isFound = pickSeriesPoster(candidates[:2])
	if !isFound || poster.ID != 2 {
		t.Fatalf("poster = %+v found=%v, want folder id 2", poster, isFound)
	}

	poster, isFound = pickSeriesPoster(candidates[:1])
	if !isFound || poster.ID != 1 {
		t.Fatalf("poster = %+v found=%v, want cover id 1", poster, isFound)
	}

	if _, isFound = pickSeriesPoster(candidates[3:]); isFound {
		t.Fatal("unrelated image must not be a poster")
	}
}

func TestPickSeriesPosterPrefersJpgOverOtherExtensionsOfSameName(t *testing.T) {
	poster, isFound := pickSeriesPoster([]PosterCandidateModel{
		{ID: 1, Name: "poster.webp"},
		{ID: 2, Name: "poster.png"},
		{ID: 3, Name: "poster.jpg"},
	})
	if !isFound || poster.ID != 3 {
		t.Fatalf("poster = %+v found=%v, want id 3", poster, isFound)
	}
}

func TestPickSeriesPosterIgnoresUnsupportedExtensions(t *testing.T) {
	if _, isFound := pickSeriesPoster([]PosterCandidateModel{{ID: 1, Name: "poster.gif"}, {ID: 2, Name: "poster"}}); isFound {
		t.Fatal("unsupported extension must not be a poster")
	}
}

func TestIsSeasonFolderName(t *testing.T) {
	seasonNames := []string{"Season 1", "season 01", "Temporada 2", "S03", "Temp.4", "season_5"}
	for _, name := range seasonNames {
		if !isSeasonFolderName(name) {
			t.Errorf("%q should be a season folder", name)
		}
	}
	otherNames := []string{"Breaking Bad", "Seasonal", "Season", "S1 Extras", "Specials", "Season 1 Extras"}
	for _, name := range otherNames {
		if isSeasonFolderName(name) {
			t.Errorf("%q should not be a season folder", name)
		}
	}
}

func TestSeasonParentFolder(t *testing.T) {
	parentFolder, isSeasonFolder := seasonParentFolder("/data/Series/Show/Season 2")
	if !isSeasonFolder || parentFolder != "/data/Series/Show" {
		t.Fatalf("parent = %q season=%v", parentFolder, isSeasonFolder)
	}
	if _, isSeasonFolder = seasonParentFolder("/data/Series/Show"); isSeasonFolder {
		t.Fatal("regular folder has no season parent")
	}
}

func TestPickMoviePosterPrefersNamedPosterThenSameNameImage(t *testing.T) {
	candidates := []PosterCandidateModel{
		{ID: 1, Name: "poster.jpg"},
		{ID: 2, Name: "Matrix.PNG"},
		{ID: 3, Name: "matrix-poster.jpg"},
	}

	poster, isFound := pickMoviePoster("Matrix.mkv", true, candidates)
	if !isFound || poster.ID != 3 {
		t.Fatalf("poster = %+v found=%v, want id 3", poster, isFound)
	}

	poster, isFound = pickMoviePoster("Matrix.mkv", true, candidates[:2])
	if !isFound || poster.ID != 2 {
		t.Fatalf("poster = %+v found=%v, want id 2", poster, isFound)
	}
}

func TestPickMovieFolderPosterOnlyWhenMovieIsAloneInFolder(t *testing.T) {
	candidates := []PosterCandidateModel{{ID: 1, Name: "folder.jpg"}, {ID: 2, Name: "poster.png"}}

	poster, isFound := pickMoviePoster("Alone.mkv", true, candidates)
	if !isFound || poster.ID != 2 {
		t.Fatalf("poster = %+v found=%v, want id 2", poster, isFound)
	}

	if _, isFound = pickMoviePoster("Shared.mkv", false, candidates); isFound {
		t.Fatal("folder poster must not apply to a movie sharing its folder")
	}
}

func TestPickMoviePosterDoesNotUseCoverImage(t *testing.T) {
	if _, isFound := pickMoviePoster("Alone.mkv", true, []PosterCandidateModel{{ID: 1, Name: "cover.jpg"}}); isFound {
		t.Fatal("cover.* is a series-only convention")
	}
}
