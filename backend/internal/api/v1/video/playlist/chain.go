package playlist

type StrategyDecision struct {
	Strategies []PlaylistStrategy
	Reasons    []string
}

type ChainHandler interface {
	Handle(ctx *PlaylistContext, decision *StrategyDecision) bool
	Name() string
}

type StrategyChain struct {
	handlers []ChainHandler
}

func NewStrategyChain(handlers ...ChainHandler) *StrategyChain {
	return &StrategyChain{handlers: handlers}
}

func (c *StrategyChain) Resolve(ctx *PlaylistContext) StrategyDecision {
	decision := StrategyDecision{}

	for _, handler := range c.handlers {
		handler.Handle(ctx, &decision)
	}

	if len(decision.Strategies) == 0 {
		decision.Strategies = []PlaylistStrategy{NewByFolderStrategy()}
		decision.Reasons = append(decision.Reasons, "fallback:by_folder")
	}

	return decision
}

type ContinueWatchingHandler struct{}

func (h *ContinueWatchingHandler) Name() string { return "continue_watching_check" }
func (h *ContinueWatchingHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	if ctx.PlaybackState != nil && !ctx.PlaybackState.Completed && ctx.PlaybackState.CurrentTime > 0 {
		d.Strategies = append(d.Strategies, NewContinueWatchingStrategy())
		d.Reasons = append(d.Reasons, "active_playback_detected")
		return true
	}
	return false
}

type CapturedSeriesHandler struct{}

func (h *CapturedSeriesHandler) Name() string { return "captured_series_check" }
func (h *CapturedSeriesHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	for _, v := range ctx.Videos {
		if v.Video.Series != nil {
			d.Strategies = append(d.Strategies, NewCapturedSeriesStrategy())
			d.Reasons = append(d.Reasons, "captured_series_detected")
			return true
		}
	}
	return false
}

type SeriesDetectionHandler struct{}

func (h *SeriesDetectionHandler) Name() string { return "series_detection" }
func (h *SeriesDetectionHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	episodeCount := 0
	for _, v := range ctx.Videos {
		if v.Classification == ClassSeries || v.Classification == ClassAnime {
			episodeCount++
		}
	}

	if episodeCount >= 2 {
		d.Strategies = append(d.Strategies, NewSequentialSeriesStrategy())
		d.Reasons = append(d.Reasons, "series_content_detected")
		return true
	}
	return false
}

type FolderGroupingHandler struct{}

func (h *FolderGroupingHandler) Name() string { return "folder_grouping" }
func (h *FolderGroupingHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	strongFolders := 0
	for _, videos := range ctx.VideosByFolder {
		if len(videos) >= 2 {
			strongFolders++
		}
	}

	if strongFolders > 0 {
		d.Strategies = append(d.Strategies, NewByFolderStrategy())
		d.Reasons = append(d.Reasons, "folder_structure_available")
		return true
	}
	return false
}

type RelatedContentHandler struct{}

func (h *RelatedContentHandler) Name() string { return "related_content" }
func (h *RelatedContentHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	classCount := map[VideoClassification]int{}
	for _, v := range ctx.Videos {
		classCount[v.Classification]++
	}

	diverseClasses := 0
	for _, count := range classCount {
		if count >= 2 {
			diverseClasses++
		}
	}

	if diverseClasses >= 2 {
		d.Strategies = append(d.Strategies, NewRelatedContentStrategy())
		d.Reasons = append(d.Reasons, "diverse_content_detected")
		return true
	}
	return false
}

type BehaviorBasedHandler struct{}

func (h *BehaviorBasedHandler) Name() string { return "behavior_based" }
func (h *BehaviorBasedHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	if ctx.Behavior == nil {
		return false
	}

	totalWatched := len(ctx.Behavior.RecentlyWatched) + len(ctx.Behavior.RecentlyCompleted)
	if totalWatched < 5 {
		return false
	}

	d.Strategies = append(d.Strategies, NewPriorityFavoritesStrategy())
	d.Reasons = append(d.Reasons, "behavior_data_available")
	return true
}

type SmartMixHandler struct{}

func (h *SmartMixHandler) Name() string { return "smart_mix" }
func (h *SmartMixHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	if len(ctx.Videos) >= 5 {
		d.Strategies = append(d.Strategies, NewRandomSmartStrategy())
		d.Reasons = append(d.Reasons, "enough_videos_for_mix")
		return true
	}
	return false
}

func DefaultStrategyChain() *StrategyChain {
	return NewStrategyChain(
		&ContinueWatchingHandler{},
		&CapturedSeriesHandler{},
		&SeriesDetectionHandler{},
		&FolderGroupingHandler{},
		&RelatedContentHandler{},
		&BehaviorBasedHandler{},
		&SmartMixHandler{},
	)
}
