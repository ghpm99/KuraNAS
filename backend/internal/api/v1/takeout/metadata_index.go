package takeout

import (
	"path"
	"strings"
)

const supplementalMetadataMarker = "supplemental-metadata"

type metadataCandidate struct {
	mediaNameStem string
	metadata      TakeoutMetadata
}

type metadataIndex struct {
	metadataByName     map[string]TakeoutMetadata
	metadataByTitle    map[string]TakeoutMetadata
	candidatesByFolder map[string][]metadataCandidate
}

func newMetadataIndex() *metadataIndex {
	return &metadataIndex{
		metadataByName:     map[string]TakeoutMetadata{},
		metadataByTitle:    map[string]TakeoutMetadata{},
		candidatesByFolder: map[string][]metadataCandidate{},
	}
}

func (index *metadataIndex) add(metadataEntryPath string, metadata TakeoutMetadata) {
	folder := path.Dir(metadataEntryPath)
	mediaNameStem := mediaNameStemOfMetadataFile(path.Base(metadataEntryPath))
	index.metadataByName[metadataEntryKey(folder, mediaNameStem)] = metadata
	if title := strings.TrimSpace(metadata.Title); title != "" {
		index.metadataByTitle[metadataEntryKey(folder, title)] = metadata
	}
	index.candidatesByFolder[folder] = append(index.candidatesByFolder[folder], metadataCandidate{
		mediaNameStem: mediaNameStem,
		metadata:      metadata,
	})
}

func (index *metadataIndex) find(mediaEntryPath string) *TakeoutMetadata {
	folder := path.Dir(mediaEntryPath)
	mediaName := path.Base(mediaEntryPath)
	entryKey := metadataEntryKey(folder, mediaName)

	if metadata, found := index.metadataByName[entryKey]; found {
		return &metadata
	}
	if metadata, found := index.metadataByTitle[entryKey]; found {
		return &metadata
	}
	return findByTruncatedName(index.candidatesByFolder[folder], mediaName)
}

func findByTruncatedName(candidates []metadataCandidate, mediaName string) *TakeoutMetadata {
	var bestCandidate *metadataCandidate
	for position := range candidates {
		candidate := &candidates[position]
		if !isTruncatedNameOf(candidate, mediaName) {
			continue
		}
		if bestCandidate == nil || len(candidate.mediaNameStem) > len(bestCandidate.mediaNameStem) {
			bestCandidate = candidate
		}
	}
	if bestCandidate == nil {
		return nil
	}
	return &bestCandidate.metadata
}

func isTruncatedNameOf(candidate *metadataCandidate, mediaName string) bool {
	if candidate.mediaNameStem == "" || len(candidate.mediaNameStem) >= len(mediaName) {
		return false
	}
	if !strings.HasPrefix(mediaName, candidate.mediaNameStem) {
		return false
	}
	return strings.TrimSpace(candidate.metadata.Title) == ""
}

func mediaNameStemOfMetadataFile(metadataFileName string) string {
	withoutJSONExtension := strings.TrimSuffix(metadataFileName, path.Ext(metadataFileName))
	return trimSupplementalMetadataSuffix(withoutJSONExtension)
}

func trimSupplementalMetadataSuffix(fileName string) string {
	suffix := strings.TrimPrefix(path.Ext(fileName), ".")
	if suffix == "" || !strings.HasPrefix(supplementalMetadataMarker, strings.ToLower(suffix)) {
		return fileName
	}
	return strings.TrimSuffix(fileName, "."+suffix)
}

func metadataEntryKey(folder string, mediaName string) string {
	return folder + "/" + mediaName
}
