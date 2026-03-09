package playlist

import (
	"sort"
)

// ---------------------------------------------------------------------------
// Template Method — PlaylistEngine orquestra o pipeline completo.
//
// Pipeline fixo:
//   1. Classificar videos           (Rule Engine / Specification)
//   2. Construir contexto           (indices, estado, comportamento)
//   3. Selecionar estrategias       (Chain of Responsibility)
//   4. Gerar candidatos             (Strategy Pattern)
//   5. Pontuar e ranquear           (Scoring Engine)
//   6. Aplicar regras de negocio    (dedup, filtros, limites)
//   7. Retornar playlists finais
//
// Cada passo pode ser customizado injetando componentes diferentes.
// ---------------------------------------------------------------------------

type PlaylistEngine struct {
	Classifier *VideoClassifier
	Chain      *StrategyChain
	Scorer     *ScoringEngine
	Analyzer   *BehaviorAnalyzer

	// Configuracoes
	MinPlaylistSize  int // playlists com menos videos sao descartadas
	MaxPlaylistCount int // maximo de playlists geradas
}

// NewPlaylistEngine cria um engine com configuracao default.
func NewPlaylistEngine() *PlaylistEngine {
	return &PlaylistEngine{
		Classifier:       NewVideoClassifier(),
		Chain:            DefaultStrategyChain(),
		Scorer:           NewScoringEngine(),
		Analyzer:         NewBehaviorAnalyzer(),
		MinPlaylistSize:  1,
		MaxPlaylistCount: 200,
	}
}

// BuildInput contem tudo que o engine precisa para gerar playlists.
type BuildInput struct {
	Videos          []VideoEntry
	ClientID        string
	PlaybackState   *PlaybackSnapshot
	BehaviorEvents  []BehaviorEvent
	Exclusions      map[int]map[int]bool // playlistID -> set[videoID]
}

// BuildResult contem o resultado do engine.
type BuildResult struct {
	Candidates       []PlaylistCandidate
	StrategyReasons  []string
	ClassifiedVideos []ClassifiedVideo
}

// Build executa o pipeline completo de geracao de playlists.
// Este e o Template Method — a sequencia e fixa, os componentes sao plugaveis.
func (e *PlaylistEngine) Build(input BuildInput) BuildResult {
	// ── Passo 1: Classificar todos os videos ──
	classified := e.Classifier.ClassifyAll(input.Videos)

	// ── Passo 2: Construir contexto enriquecido ──
	ctx := e.buildContext(classified, input)

	// ── Passo 3: Selecionar estrategias via Chain of Responsibility ──
	decision := e.Chain.Resolve(ctx)

	// ── Passo 4: Executar todas as estrategias selecionadas ──
	var allCandidates []PlaylistCandidate
	for _, strategy := range decision.Strategies {
		candidates := strategy.Build(ctx)
		allCandidates = append(allCandidates, candidates...)
	}

	// ── Passo 5: Scoring e ranking ──
	allCandidates = e.Scorer.ScoreAll(
		allCandidates,
		classified,
		ctx.Behavior,
		input.Exclusions,
	)

	// ── Passo 6: Regras de negocio finais ──
	allCandidates = e.applyBusinessRules(allCandidates)

	return BuildResult{
		Candidates:       allCandidates,
		StrategyReasons:  decision.Reasons,
		ClassifiedVideos: classified,
	}
}

// ---------------------------------------------------------------------------
// Passos internos do pipeline
// ---------------------------------------------------------------------------

func (e *PlaylistEngine) buildContext(classified []ClassifiedVideo, input BuildInput) *PlaylistContext {
	// Indice por ID
	byID := make(map[int]*ClassifiedVideo, len(classified))
	for i := range classified {
		byID[classified[i].Video.ID] = &classified[i]
	}

	// Indice por pasta
	byFolder := make(map[string][]*ClassifiedVideo)
	for i := range classified {
		folder := classified[i].Video.ParentPath
		byFolder[folder] = append(byFolder[folder], &classified[i])
	}

	// Construir perfil de comportamento
	var behavior *BehaviorProfile
	if len(input.BehaviorEvents) > 0 {
		behavior = e.Analyzer.BuildProfile(input.BehaviorEvents, byID)
	}

	return &PlaylistContext{
		Videos:         classified,
		VideoByID:      byID,
		VideosByFolder: byFolder,
		ClientID:       input.ClientID,
		PlaybackState:  input.PlaybackState,
		Behavior:       behavior,
		Exclusions:     input.Exclusions,
	}
}

func (e *PlaylistEngine) applyBusinessRules(candidates []PlaylistCandidate) []PlaylistCandidate {
	// Regra 1: Remover playlists muito pequenas (exceto continue_watching e singletons intencionais)
	filtered := make([]PlaylistCandidate, 0, len(candidates))
	for _, c := range candidates {
		if len(c.Videos) >= e.MinPlaylistSize || c.PlaylistType == "continue" {
			filtered = append(filtered, c)
		}
	}

	// Regra 2: Deduplicar — se dois candidates compartilham >80% dos videos,
	// manter o com maior score.
	filtered = e.deduplicateCandidates(filtered)

	// Regra 3: Remover videos com score muito negativo de cada candidate
	for i := range filtered {
		var cleaned []ScoredVideo
		for _, sv := range filtered[i].Videos {
			if sv.Score > -100 { // threshold: excluidos pelo usuario ficam abaixo
				cleaned = append(cleaned, sv)
			}
		}
		filtered[i].Videos = cleaned
	}

	// Regra 4: Remover candidates que ficaram vazios apos limpeza
	var final []PlaylistCandidate
	for _, c := range filtered {
		if len(c.Videos) > 0 {
			final = append(final, c)
		}
	}

	// Regra 5: Limitar quantidade total
	if len(final) > e.MaxPlaylistCount {
		// Ja estao ordenados por score total
		final = final[:e.MaxPlaylistCount]
	}

	return final
}

func (e *PlaylistEngine) deduplicateCandidates(candidates []PlaylistCandidate) []PlaylistCandidate {
	if len(candidates) <= 1 {
		return candidates
	}

	// Ja estao ordenados por TotalScore descrescente
	removed := map[int]bool{}

	for i := 0; i < len(candidates); i++ {
		if removed[i] {
			continue
		}
		setI := videoIDSet(candidates[i])

		for j := i + 1; j < len(candidates); j++ {
			if removed[j] {
				continue
			}
			setJ := videoIDSet(candidates[j])
			overlap := overlapPercent(setI, setJ)

			if overlap > 0.8 {
				// Remover o com menor score (j, pois ja esta ordenado)
				removed[j] = true
			}
		}
	}

	result := make([]PlaylistCandidate, 0, len(candidates)-len(removed))
	for i, c := range candidates {
		if !removed[i] {
			result = append(result, c)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func videoIDSet(c PlaylistCandidate) map[int]bool {
	s := make(map[int]bool, len(c.Videos))
	for _, v := range c.Videos {
		s[v.Video.Video.ID] = true
	}
	return s
}

func overlapPercent(a, b map[int]bool) float64 {
	smaller, larger := a, b
	if len(a) > len(b) {
		smaller, larger = b, a
	}
	if len(smaller) == 0 {
		return 0
	}

	overlap := 0
	for id := range smaller {
		if larger[id] {
			overlap++
		}
	}
	return float64(overlap) / float64(len(smaller))
}

// ---------------------------------------------------------------------------
// Utility: converter output do engine para o formato antigo (smartGroup)
// para manter compatibilidade com o service existente.
// ---------------------------------------------------------------------------

type SmartGroup struct {
	SourceKey      string
	Name           string
	PlaylistType   string
	GroupMode      string
	Classification string
	VideoIDs       []int
}

// ToSmartGroups converte o resultado do engine para o formato que o
// service.RebuildSmartPlaylists() espera. Isso permite migracao incremental.
func (r BuildResult) ToSmartGroups() []SmartGroup {
	groups := make([]SmartGroup, 0, len(r.Candidates))

	for _, c := range r.Candidates {
		ids := make([]int, 0, len(c.Videos))
		for _, sv := range c.Videos {
			ids = append(ids, sv.Video.Video.ID)
		}
		sort.Ints(ids)

		groups = append(groups, SmartGroup{
			SourceKey:      c.SourceKey,
			Name:           c.Name,
			PlaylistType:   c.PlaylistType,
			GroupMode:      c.GroupMode,
			Classification: string(c.Classification),
			VideoIDs:       ids,
		})
	}

	return groups
}
