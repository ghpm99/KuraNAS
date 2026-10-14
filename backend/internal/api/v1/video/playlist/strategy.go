package playlist

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"nas-go/api/pkg/utils"
)

type PlaylistStrategy interface {
	Name() string
	Build(ctx *PlaylistContext) []PlaylistCandidate
}

type ByFolderStrategy struct {
	MinItemsPerFolder int
}

func NewByFolderStrategy() *ByFolderStrategy {
	return &ByFolderStrategy{MinItemsPerFolder: 2}
}

func (s *ByFolderStrategy) Name() string { return "by_folder" }

func (s *ByFolderStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	var candidates []PlaylistCandidate

	for folder, videos := range ctx.VideosByFolder {
		folderBase := strings.TrimSpace(filepath.Base(folder))
		if len(videos) < s.MinItemsPerFolder || IsGenericFolderName(folderBase) {
			continue
		}

		classCount := map[VideoClassification]int{}
		for _, v := range videos {
			classCount[v.Classification]++
		}
		dominant := dominantClassification(classCount)

		scored := make([]ScoredVideo, 0, len(videos))
		for _, v := range videos {
			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   1.0,
				Reasons: []string{"same_folder"},
			})
		}

		sort.SliceStable(scored, func(i, j int) bool {
			return lessByNaturalNameThenID(scored[i].Video.Video, scored[j].Video.Video)
		})

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "folder:" + folder,
			Name:           folderBase,
			PlaylistType:   playlistTypeFor(dominant),
			GroupMode:      "folder",
			Classification: dominant,
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

type SequentialSeriesStrategy struct {
	MinItemsPerGroup int
}

func NewSequentialSeriesStrategy() *SequentialSeriesStrategy {
	return &SequentialSeriesStrategy{MinItemsPerGroup: 2}
}

func (s *SequentialSeriesStrategy) Name() string { return "sequential_series" }

func (s *SequentialSeriesStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	prefixGroups := map[string][]*ClassifiedVideo{}
	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		prefix := InferTitlePrefix(v.Video.Name)
		if prefix != "" {
			prefixGroups[prefix] = append(prefixGroups[prefix], v)
		}
	}

	var candidates []PlaylistCandidate
	for prefix, videos := range prefixGroups {
		if len(videos) < s.MinItemsPerGroup {
			continue
		}

		scored := make([]ScoredVideo, 0, len(videos))
		for _, v := range videos {
			reasons := []string{"shared_prefix:" + prefix}

			if EpisodePattern.MatchString(strings.ToLower(v.Video.Name)) {
				reasons = append(reasons, "episode_pattern")
			}

			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   1.0,
				Reasons: reasons,
			})
		}

		sort.SliceStable(scored, func(i, j int) bool {
			return lessByNaturalNameThenID(scored[i].Video.Video, scored[j].Video.Video)
		})

		classCount := map[VideoClassification]int{}
		for _, v := range videos {
			classCount[v.Classification]++
		}

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "series:" + prefix,
			Name:           formatSeriesName(prefix),
			NameKey:        seriesNameKey(prefix),
			PlaylistType:   "series",
			GroupMode:      "prefix",
			Classification: dominantClassification(classCount),
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

type CapturedSeriesStrategy struct{}

func NewCapturedSeriesStrategy() *CapturedSeriesStrategy {
	return &CapturedSeriesStrategy{}
}

func (s *CapturedSeriesStrategy) Name() string { return "captured_series" }

func (s *CapturedSeriesStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	type seriesGroup struct {
		title  string
		videos []*ClassifiedVideo
	}

	groups := map[string]*seriesGroup{}
	order := make([]string, 0)

	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		if v.Video.Series == nil {
			continue
		}
		title := strings.TrimSpace(v.Video.Series.Title)
		if title == "" {
			continue
		}
		key := normalizeSeriesKey(title)
		g, ok := groups[key]
		if !ok {
			g = &seriesGroup{title: title}
			groups[key] = g
			order = append(order, key)
		}
		g.videos = append(g.videos, v)
	}

	candidates := make([]PlaylistCandidate, 0, len(order))
	for _, key := range order {
		g := groups[key]

		sort.SliceStable(g.videos, func(i, j int) bool {
			return seriesProvenanceLess(
				g.videos[i].Video.Series, g.videos[i].Video.Name,
				g.videos[j].Video.Series, g.videos[j].Video.Name,
			)
		})

		classCount := map[VideoClassification]int{}
		scored := make([]ScoredVideo, 0, len(g.videos))
		for _, v := range g.videos {
			classCount[v.Classification]++
			prov := v.Video.Series
			reasons := []string{"captured_series:" + key}
			if prov.Season != nil {
				reasons = append(reasons, fmt.Sprintf("s%de%d", *prov.Season, prov.Episode))
			} else {
				reasons = append(reasons, fmt.Sprintf("e%d", prov.Episode))
			}
			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   1.0,
				Reasons: reasons,
			})
		}

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "series:capture:" + key,
			Name:           g.title,
			PlaylistType:   "series",
			GroupMode:      "captured_series",
			Classification: dominantClassification(classCount),
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

type ContinueWatchingStrategy struct{}

func NewContinueWatchingStrategy() *ContinueWatchingStrategy {
	return &ContinueWatchingStrategy{}
}

func (s *ContinueWatchingStrategy) Name() string { return "continue_watching" }

func (s *ContinueWatchingStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	if ctx.PlaybackState == nil {
		return nil
	}

	state := ctx.PlaybackState
	current, ok := ctx.VideoByID[state.VideoID]
	if !ok {
		return nil
	}

	if state.Completed || state.CurrentTime <= 0 {
		return nil
	}

	scored := []ScoredVideo{
		{
			Video:   *current,
			Score:   100.0,
			Reasons: []string{"in_progress", fmt.Sprintf("position:%.0fs", state.CurrentTime)},
		},
	}

	sameFolder := ctx.VideosByFolder[current.Video.ParentPath]
	for _, v := range sameFolder {
		if v.Video.ID == current.Video.ID {
			continue
		}
		if utils.CompareNaturalOrder(v.Video.Name, current.Video.Name) > 0 {
			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   50.0,
				Reasons: []string{"next_in_folder"},
			})
		}
	}

	if len(scored) == 0 {
		return nil
	}

	return []PlaylistCandidate{
		{
			SourceKey:      fmt.Sprintf("continue:%s:%d", ctx.ClientID, state.VideoID),
			NameKey:        NameKeyContinueWatching,
			PlaylistType:   "continue",
			GroupMode:      "behavior",
			Classification: current.Classification,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

type RelatedContentStrategy struct {
	MaxResults        int
	DurationTolerance float64
}

func NewRelatedContentStrategy() *RelatedContentStrategy {
	return &RelatedContentStrategy{
		MaxResults:        20,
		DurationTolerance: 0.3,
	}
}

func (s *RelatedContentStrategy) Name() string { return "related_content" }

func (s *RelatedContentStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	byClass := map[VideoClassification][]*ClassifiedVideo{}
	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		if v.Video.Series != nil {
			continue
		}
		byClass[v.Classification] = append(byClass[v.Classification], v)
	}

	var candidates []PlaylistCandidate

	for class, videos := range byClass {
		if class == ClassProgram || class == ClassClip {
			continue
		}
		if len(videos) < 2 {
			continue
		}

		scored := make([]ScoredVideo, 0, len(videos))
		for _, v := range videos {
			score := 1.0
			reasons := []string{fmt.Sprintf("classification:%s", class)}

			if v.Video.Meta != nil {
				reasons = append(reasons, "has_metadata")
				score += 0.5
			}

			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   score,
				Reasons: reasons,
			})
		}

		if len(scored) > s.MaxResults {
			sort.Slice(scored, func(i, j int) bool {
				return scored[i].Score > scored[j].Score
			})
			scored = scored[:s.MaxResults]
		}

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "related:" + string(class),
			NameKey:        classificationNameKey(class),
			PlaylistType:   playlistTypeFor(class),
			GroupMode:      "classification",
			Classification: class,
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

type PriorityFavoritesStrategy struct {
	MaxResults int
}

func NewPriorityFavoritesStrategy() *PriorityFavoritesStrategy {
	return &PriorityFavoritesStrategy{MaxResults: 30}
}

func (s *PriorityFavoritesStrategy) Name() string { return "priority_favorites" }

func (s *PriorityFavoritesStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	if ctx.Behavior == nil {
		return nil
	}

	behavior := ctx.Behavior

	recentlyWatched := toSet(behavior.RecentlyWatched)
	recentlyCompleted := toSet(behavior.RecentlyCompleted)
	recentlySkipped := toSet(behavior.RecentlySkipped)

	var scored []ScoredVideo
	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		score := 0.0
		var reasons []string

		if !recentlyWatched[v.Video.ID] && !recentlyCompleted[v.Video.ID] {
			score += 30
			reasons = append(reasons, "unwatched:+30")
		}

		if affinity, ok := behavior.PreferredTypes[v.Classification]; ok && affinity > 0.5 {
			bonus := affinity * 25
			score += bonus
			reasons = append(reasons, fmt.Sprintf("preferred_type:+%.0f", bonus))
		}

		if v.Video.Meta != nil && behavior.PreferredDurations.MaxSeconds > 0 {
			if v.Video.Meta.Duration >= behavior.PreferredDurations.MinSeconds &&
				v.Video.Meta.Duration <= behavior.PreferredDurations.MaxSeconds {
				score += 15
				reasons = append(reasons, "preferred_duration:+15")
			}
		}

		if recentlySkipped[v.Video.ID] {
			score -= 40
			reasons = append(reasons, "recently_skipped:-40")
		}

		if recentlyCompleted[v.Video.ID] {
			score -= 20
			reasons = append(reasons, "already_completed:-20")
		}

		if !v.Video.CreatedAt.IsZero() {
			daysSinceCreation := daysSince(v.Video.CreatedAt)
			if daysSinceCreation < 7 {
				score += 10
				reasons = append(reasons, "new_file:+10")
			}
		}

		if score > 0 && len(reasons) > 0 {
			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   score,
				Reasons: reasons,
			})
		}
	}

	if len(scored) == 0 {
		return nil
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > s.MaxResults {
		scored = scored[:s.MaxResults]
	}

	return []PlaylistCandidate{
		{
			SourceKey:      fmt.Sprintf("favorites:%s", ctx.ClientID),
			NameKey:        NameKeyRecommended,
			PlaylistType:   "mixed",
			GroupMode:      "behavior",
			Classification: ClassPersonal,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

type RandomSmartStrategy struct {
	MaxResults int
}

func NewRandomSmartStrategy() *RandomSmartStrategy {
	return &RandomSmartStrategy{MaxResults: 20}
}

func (s *RandomSmartStrategy) Name() string { return "random_smart" }

func (s *RandomSmartStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	if len(ctx.Videos) < 3 {
		return nil
	}

	byClass := map[VideoClassification][]ClassifiedVideo{}
	for _, v := range ctx.Videos {
		if v.Classification != ClassProgram {
			byClass[v.Classification] = append(byClass[v.Classification], v)
		}
	}

	var scored []ScoredVideo
	classKeys := make([]VideoClassification, 0, len(byClass))
	for k := range byClass {
		classKeys = append(classKeys, k)
	}
	sort.Slice(classKeys, func(i, j int) bool {
		return string(classKeys[i]) < string(classKeys[j])
	})

	indices := map[VideoClassification]int{}
	for len(scored) < s.MaxResults {
		added := false
		for _, class := range classKeys {
			videos := byClass[class]
			idx := indices[class]
			if idx >= len(videos) {
				continue
			}
			v := videos[idx]
			indices[class] = idx + 1
			added = true

			scored = append(scored, ScoredVideo{
				Video:   v,
				Score:   float64(s.MaxResults - len(scored)),
				Reasons: []string{"variety_mix", fmt.Sprintf("type:%s", class)},
			})

			if len(scored) >= s.MaxResults {
				break
			}
		}
		if !added {
			break
		}
	}

	if len(scored) == 0 {
		return nil
	}

	return []PlaylistCandidate{
		{
			SourceKey:      "random:smart_mix",
			NameKey:        NameKeyMixed,
			PlaylistType:   "mixed",
			GroupMode:      "random",
			Classification: ClassPersonal,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

func normalizeSeriesKey(title string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(title))), " ")
}

func seriesProvenanceLess(a *SeriesProvenance, nameA string, b *SeriesProvenance, nameB string) bool {
	seasonA, seasonB := 0, 0
	if a.Season != nil {
		seasonA = *a.Season
	}
	if b.Season != nil {
		seasonB = *b.Season
	}
	if seasonA != seasonB {
		return seasonA < seasonB
	}
	if a.Episode != b.Episode {
		return a.Episode < b.Episode
	}
	return nameA < nameB
}

var classificationTieBreakOrder = []VideoClassification{
	ClassSeries,
	ClassAnime,
	ClassCourse,
	ClassMovie,
	ClassMusic,
	ClassProgram,
	ClassClip,
	ClassPersonal,
}

func dominantClassification(counts map[VideoClassification]int) VideoClassification {
	best := ClassPersonal
	bestCount := 0
	for _, class := range classificationTieBreakOrder {
		if counts[class] > bestCount {
			best = class
			bestCount = counts[class]
		}
	}
	return best
}

func playlistTypeFor(class VideoClassification) string {
	switch class {
	case ClassSeries, ClassAnime:
		return "series"
	case ClassMovie:
		return "movie"
	case ClassCourse:
		return "course"
	default:
		return "folder"
	}
}

func seriesNameKey(prefix string) string {
	if prefix == "" {
		return NameKeyUnknownSeries
	}
	return ""
}

func formatSeriesName(prefix string) string {
	words := strings.Fields(prefix)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func toSet(ids []int) map[int]bool {
	s := make(map[int]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

func lessByNaturalNameThenID(left, right VideoEntry) bool {
	if comparison := utils.CompareNaturalOrder(left.Name, right.Name); comparison != 0 {
		return comparison < 0
	}
	return left.ID < right.ID
}
