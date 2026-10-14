package video

import (
	"database/sql"
	"errors"
	"testing"
)

func posterRepositoryWithFolders(imagesByFolder map[string][]PosterCandidateModel) *videoRepoMock {
	return &videoRepoMock{
		listPosterCandidatesInFolderFn: func(folderPath string) ([]PosterCandidateModel, error) {
			return imagesByFolder[folderPath], nil
		},
	}
}

func TestFindPlaylistPosterUsesFolderOfFolderPlaylist(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{"/data/Show": {{ID: 7, Name: "poster.jpg"}}})
	repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{ID: id, SourcePath: "folder:/data/Show"}, nil
	}

	posterFileID, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1)
	if err != nil || posterFileID != 7 {
		t.Fatalf("poster = %d err = %v, want 7", posterFileID, err)
	}
}

func TestFindPlaylistPosterFallsBackToParentOfSeasonFolder(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{
		"/data/Show":          {{ID: 9, Name: "folder.png"}},
		"/data/Show/Season 1": {{ID: 3, Name: "unrelated.jpg"}},
	})
	repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{ID: id, SourcePath: "folder:/data/Show/Season 1"}, nil
	}

	posterFileID, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1)
	if err != nil || posterFileID != 9 {
		t.Fatalf("poster = %d err = %v, want 9", posterFileID, err)
	}
}

func TestFindPlaylistPosterPrefersSeasonFolderOverParent(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{
		"/data/Show":          {{ID: 9, Name: "poster.png"}},
		"/data/Show/Season 1": {{ID: 3, Name: "cover.jpg"}},
	})
	repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{ID: id, SourcePath: "folder:/data/Show/Season 1"}, nil
	}

	posterFileID, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1)
	if err != nil || posterFileID != 3 {
		t.Fatalf("poster = %d err = %v, want 3", posterFileID, err)
	}
}

func TestFindPlaylistPosterDoesNotClimbFromRegularFolder(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{"/data": {{ID: 9, Name: "poster.png"}}})
	repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{ID: id, SourcePath: "folder:/data/Show"}, nil
	}

	if _, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1); !errors.Is(err, ErrPosterNotFound) {
		t.Fatalf("err = %v, want ErrPosterNotFound", err)
	}
}

func TestFindPlaylistPosterUsesDominantFolderOfSeriesPlaylist(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{"/data/Show": {{ID: 4, Name: "cover.webp"}}})
	repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{ID: id, SourcePath: "series:show"}, nil
	}
	repository.getPlaylistDominantFolderFn = func(playlistID int) (string, error) { return "/data/Show", nil }

	posterFileID, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1)
	if err != nil || posterFileID != 4 {
		t.Fatalf("poster = %d err = %v, want 4", posterFileID, err)
	}
}

func TestFindPlaylistPosterReportsNotFoundForSeriesWithoutItemsAndOtherSources(t *testing.T) {
	for _, sourcePath := range []string{"series:show", "random:smart_mix", "custom:abc"} {
		repository := posterRepositoryWithFolders(nil)
		repository.getVideoPlaylistByIDFn = func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id, SourcePath: sourcePath}, nil
		}
		repository.getPlaylistDominantFolderFn = func(playlistID int) (string, error) { return "", sql.ErrNoRows }

		if _, err := newVideoServiceForTest(t, repository).FindPlaylistPosterFileID(1); !errors.Is(err, ErrPosterNotFound) {
			t.Errorf("%s: err = %v, want ErrPosterNotFound", sourcePath, err)
		}
	}
}

func TestFindMoviePosterUsesFolderPosterOnlyForSingleVideoFolder(t *testing.T) {
	repository := posterRepositoryWithFolders(map[string][]PosterCandidateModel{"/data/Movies": {{ID: 5, Name: "poster.jpg"}}})
	repository.getVideoFileByIDFn = func(id int) (VideoFileModel, error) {
		return VideoFileModel{ID: id, Name: "Film.mkv", ParentPath: "/data/Movies"}, nil
	}

	repository.countVideosInFolderFn = func(folderPath string) (int, error) { return 2, nil }
	if _, err := newVideoServiceForTest(t, repository).FindMoviePosterFileID(1); !errors.Is(err, ErrPosterNotFound) {
		t.Fatalf("shared folder: err = %v, want ErrPosterNotFound", err)
	}

	repository.countVideosInFolderFn = func(folderPath string) (int, error) { return 1, nil }
	posterFileID, err := newVideoServiceForTest(t, repository).FindMoviePosterFileID(1)
	if err != nil || posterFileID != 5 {
		t.Fatalf("own folder: poster = %d err = %v, want 5", posterFileID, err)
	}
}

func TestFindMoviePosterPropagatesRepositoryErrors(t *testing.T) {
	repository := posterRepositoryWithFolders(nil)
	repository.getVideoFileByIDFn = func(id int) (VideoFileModel, error) { return VideoFileModel{}, sql.ErrNoRows }
	if _, err := newVideoServiceForTest(t, repository).FindMoviePosterFileID(1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}

	failure := errors.New("boom")
	repository.getVideoFileByIDFn = func(id int) (VideoFileModel, error) {
		return VideoFileModel{Name: "a.mkv", ParentPath: "/x"}, nil
	}
	repository.listPosterCandidatesInFolderFn = func(folderPath string) ([]PosterCandidateModel, error) { return nil, failure }
	if _, err := newVideoServiceForTest(t, repository).FindMoviePosterFileID(1); !errors.Is(err, failure) {
		t.Fatalf("err = %v, want boom", err)
	}

	repository.listPosterCandidatesInFolderFn = func(folderPath string) ([]PosterCandidateModel, error) {
		return []PosterCandidateModel{{ID: 1, Name: "poster.jpg"}}, nil
	}
	repository.countVideosInFolderFn = func(folderPath string) (int, error) { return 0, failure }
	if _, err := newVideoServiceForTest(t, repository).FindMoviePosterFileID(1); !errors.Is(err, failure) {
		t.Fatalf("err = %v, want boom", err)
	}
}
