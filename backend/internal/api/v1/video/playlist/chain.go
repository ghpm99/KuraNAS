package playlist

// ---------------------------------------------------------------------------
// Chain of Responsibility — decide quais estrategias usar baseado no contexto.
// Cada handler avalia o contexto e pode adicionar/remover estrategias.
// A chain permite extensao sem modificar o codigo existente.
// ---------------------------------------------------------------------------

// StrategyDecision e o resultado acumulado da chain.
type StrategyDecision struct {
	Strategies []PlaylistStrategy
	Reasons    []string
}

// ChainHandler avalia o contexto e pode modificar a decisao.
// Retorna true se processou (proximo handler ainda roda, mas a decisao ja contem algo).
type ChainHandler interface {
	Handle(ctx *PlaylistContext, decision *StrategyDecision) bool
	Name() string
}

// StrategyChain e a chain completa de decisao.
type StrategyChain struct {
	handlers []ChainHandler
}

func NewStrategyChain(handlers ...ChainHandler) *StrategyChain {
	return &StrategyChain{handlers: handlers}
}

// Resolve executa todos os handlers da chain e retorna a decisao final.
func (c *StrategyChain) Resolve(ctx *PlaylistContext) StrategyDecision {
	decision := StrategyDecision{}

	for _, handler := range c.handlers {
		handler.Handle(ctx, &decision)
	}

	// Se nenhum handler adicionou estrategia, usar fallback
	if len(decision.Strategies) == 0 {
		decision.Strategies = []PlaylistStrategy{NewByFolderStrategy()}
		decision.Reasons = append(decision.Reasons, "fallback:by_folder")
	}

	return decision
}

// ---------------------------------------------------------------------------
// Handlers concretos
// ---------------------------------------------------------------------------

// ContinueWatchingHandler: se o usuario tem um video em progresso, adiciona
// a estrategia de "continuar assistindo" como prioridade.
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

// SeriesDetectionHandler: se existem videos com padrao de episodio,
// adiciona a estrategia de series sequenciais.
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

// FolderGroupingHandler: sempre adiciona agrupamento por pasta como base.
type FolderGroupingHandler struct{}

func (h *FolderGroupingHandler) Name() string { return "folder_grouping" }
func (h *FolderGroupingHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	// Contar pastas com multiplos videos
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

// RelatedContentHandler: se existem videos classificados, adiciona
// agrupamento por conteudo relacionado.
type RelatedContentHandler struct{}

func (h *RelatedContentHandler) Name() string { return "related_content" }
func (h *RelatedContentHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	classCount := map[VideoClassification]int{}
	for _, v := range ctx.Videos {
		classCount[v.Classification]++
	}

	// So adicionar se temos pelo menos 2 classificacoes diferentes com 2+ videos
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

// BehaviorBasedHandler: se temos dados de comportamento do usuario,
// adiciona estrategia de favoritos e mix inteligente.
type BehaviorBasedHandler struct{}

func (h *BehaviorBasedHandler) Name() string { return "behavior_based" }
func (h *BehaviorBasedHandler) Handle(ctx *PlaylistContext, d *StrategyDecision) bool {
	if ctx.Behavior == nil {
		return false
	}

	// So ativar se temos dados suficientes (pelo menos 5 videos assistidos)
	totalWatched := len(ctx.Behavior.RecentlyWatched) + len(ctx.Behavior.RecentlyCompleted)
	if totalWatched < 5 {
		return false
	}

	d.Strategies = append(d.Strategies, NewPriorityFavoritesStrategy())
	d.Reasons = append(d.Reasons, "behavior_data_available")
	return true
}

// SmartMixHandler: sempre adiciona o mix variado como ultima opcao.
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

// ---------------------------------------------------------------------------
// Chain default: pipeline padrao de decisao
// ---------------------------------------------------------------------------

func DefaultStrategyChain() *StrategyChain {
	return NewStrategyChain(
		&ContinueWatchingHandler{},
		&SeriesDetectionHandler{},
		&FolderGroupingHandler{},
		&RelatedContentHandler{},
		&BehaviorBasedHandler{},
		&SmartMixHandler{},
	)
}
