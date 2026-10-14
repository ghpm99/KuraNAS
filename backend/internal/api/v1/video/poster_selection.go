package video

import (
	"path/filepath"
	"regexp"
	"strings"
)

var posterImageExtensions = []string{".jpg", ".jpeg", ".png", ".webp"}

var seriesPosterBaseNames = []string{"poster", "folder", "cover"}

var movieFolderPosterBaseNames = []string{"poster", "folder"}

var seasonFolderNamePattern = regexp.MustCompile(`(?i)^(?:season|temporada|temp|s)[\s._-]*\d{1,3}$`)

type PosterCandidateModel struct {
	ID   int
	Name string
}

func isSeasonFolderName(folderName string) bool {
	return seasonFolderNamePattern.MatchString(strings.TrimSpace(folderName))
}

func seasonParentFolder(folderPath string) (string, bool) {
	if !isSeasonFolderName(filepath.Base(folderPath)) {
		return "", false
	}
	parentFolder := filepath.Dir(folderPath)
	if parentFolder == folderPath {
		return "", false
	}
	return parentFolder, true
}

func posterExtensionRank(fileName string) int {
	extension := strings.ToLower(filepath.Ext(fileName))
	for rank, posterExtension := range posterImageExtensions {
		if extension == posterExtension {
			return rank
		}
	}
	return -1
}

func findPosterByBaseName(candidates []PosterCandidateModel, baseName string) (PosterCandidateModel, bool) {
	bestRank := len(posterImageExtensions)
	var best PosterCandidateModel
	for _, candidate := range candidates {
		rank := posterExtensionRank(candidate.Name)
		if rank < 0 || rank >= bestRank {
			continue
		}
		candidateBaseName := strings.TrimSuffix(candidate.Name, filepath.Ext(candidate.Name))
		if !strings.EqualFold(candidateBaseName, baseName) {
			continue
		}
		best = candidate
		bestRank = rank
	}
	return best, bestRank < len(posterImageExtensions)
}

func findPosterByPriority(candidates []PosterCandidateModel, baseNames []string) (PosterCandidateModel, bool) {
	for _, baseName := range baseNames {
		if poster, isFound := findPosterByBaseName(candidates, baseName); isFound {
			return poster, true
		}
	}
	return PosterCandidateModel{}, false
}

func pickSeriesPoster(candidates []PosterCandidateModel) (PosterCandidateModel, bool) {
	return findPosterByPriority(candidates, seriesPosterBaseNames)
}

func pickMoviePoster(movieFileName string, isMovieAloneInFolder bool, candidates []PosterCandidateModel) (PosterCandidateModel, bool) {
	movieBaseName := strings.TrimSuffix(movieFileName, filepath.Ext(movieFileName))
	namedBaseNames := []string{movieBaseName + "-poster", movieBaseName}
	if poster, isFound := findPosterByPriority(candidates, namedBaseNames); isFound {
		return poster, true
	}
	if !isMovieAloneInFolder {
		return PosterCandidateModel{}, false
	}
	return findPosterByPriority(candidates, movieFolderPosterBaseNames)
}
