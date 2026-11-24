package playlist

import (
	"path/filepath"
	"sort"
	"strings"
)

func isSeriesLike(class VideoClassification) bool {
	return class == ClassSeries || class == ClassAnime
}

func isOverriddenToNonSeries(video *ClassifiedVideo) bool {
	return video.Video.IsClassificationOverride && !isSeriesLike(video.Classification)
}

func folderDominantClassification(videos []*ClassifiedVideo) VideoClassification {
	classCount := map[VideoClassification]int{}
	for _, v := range videos {
		if !v.Video.IsClassificationOverride {
			classCount[v.Classification]++
		}
	}
	if len(classCount) == 0 {
		for _, v := range videos {
			classCount[v.Classification]++
		}
	}
	return dominantClassification(classCount)
}

func isOverrideConflictingWithFolder(video *ClassifiedVideo, folderDominant VideoClassification) bool {
	return video.Video.IsClassificationOverride && isSeriesLike(video.Classification) != isSeriesLike(folderDominant)
}

func videosCompatibleWithFolder(videos []*ClassifiedVideo, folderDominant VideoClassification) []*ClassifiedVideo {
	compatible := make([]*ClassifiedVideo, 0, len(videos))
	for _, v := range videos {
		if !isOverrideConflictingWithFolder(v, folderDominant) {
			compatible = append(compatible, v)
		}
	}
	return compatible
}

type OverriddenSeriesStrategy struct{}

func NewOverriddenSeriesStrategy() *OverriddenSeriesStrategy {
	return &OverriddenSeriesStrategy{}
}

func (s *OverriddenSeriesStrategy) Name() string { return "overridden_series" }

func (s *OverriddenSeriesStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	var candidates []PlaylistCandidate

	for folder, videos := range ctx.VideosByFolder {
		folderDominant := folderDominantClassification(videos)
		var overriddenToSeries []*ClassifiedVideo
		for _, v := range videos {
			if isOverrideConflictingWithFolder(v, folderDominant) {
				overriddenToSeries = append(overriddenToSeries, v)
			}
		}
		if len(overriddenToSeries) == 0 {
			continue
		}
		candidates = append(candidates, s.candidatesForFolder(folder, overriddenToSeries)...)
	}

	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].SourceKey < candidates[j].SourceKey })
	return candidates
}

func (s *OverriddenSeriesStrategy) candidatesForFolder(folder string, videos []*ClassifiedVideo) []PlaylistCandidate {
	folderBase := strings.TrimSpace(filepath.Base(folder))
	videosByPrefix := map[string][]*ClassifiedVideo{}
	for _, v := range videos {
		prefix := InferTitlePrefix(v.Video.Name)
		videosByPrefix[prefix] = append(videosByPrefix[prefix], v)
	}

	if len(videosByPrefix) > 1 || IsGenericFolderName(folderBase) {
		candidates := make([]PlaylistCandidate, 0, len(videosByPrefix))
		for prefix, group := range videosByPrefix {
			candidates = append(candidates, s.candidate("folder_series:"+folder+":"+prefix, seriesNameForPrefix(prefix, group), group))
		}
		return candidates
	}
	return []PlaylistCandidate{s.candidate("folder_series:"+folder, folderBase, videos)}
}

func seriesNameForPrefix(prefix string, videos []*ClassifiedVideo) string {
	if prefix != "" {
		return formatSeriesName(prefix)
	}
	return strings.TrimSuffix(videos[0].Video.Name, filepath.Ext(videos[0].Video.Name))
}

func (s *OverriddenSeriesStrategy) candidate(sourceKey string, name string, videos []*ClassifiedVideo) PlaylistCandidate {
	classCount := map[VideoClassification]int{}
	scored := make([]ScoredVideo, 0, len(videos))
	for _, v := range videos {
		classCount[v.Classification]++
		scored = append(scored, ScoredVideo{
			Video:   *v,
			Score:   1.0,
			Reasons: []string{"classification_override"},
		})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return lessByNaturalNameThenID(scored[i].Video.Video, scored[j].Video.Video)
	})

	return PlaylistCandidate{
		SourceKey:      sourceKey,
		Name:           name,
		PlaylistType:   "series",
		GroupMode:      "folder",
		Classification: dominantClassification(classCount),
		Strategy:       s.Name(),
		Videos:         scored,
	}
}
