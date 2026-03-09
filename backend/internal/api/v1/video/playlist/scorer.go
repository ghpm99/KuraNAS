package playlist

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Scoring/Ranking Engine — o coracao da "inteligencia"
//
// Cada ScoringRule contribui com um score ponderado. O engine agrega
// todos os scores para cada video dentro de cada playlist candidate.
// ---------------------------------------------------------------------------

// ScoringRule avalia um video num contexto e retorna um score.
type ScoringRule interface {
	Name() string
	Weight() float64
	Score(video ClassifiedVideo, ctx *ScoringContext) float64
}

// ScoringContext contem informacoes necessarias para scoring.
type ScoringContext struct {
	Candidate  *PlaylistCandidate
	AllVideos  []ClassifiedVideo
	Behavior   *BehaviorProfile
	Exclusions map[int]bool // videos excluidos desta playlist
}

// ScoringEngine aplica todas as regras e calcula scores finais.
type ScoringEngine struct {
	rules []ScoringRule
}

func NewScoringEngine(rules ...ScoringRule) *ScoringEngine {
	if len(rules) == 0 {
		rules = DefaultScoringRules()
	}
	return &ScoringEngine{rules: rules}
}

// ScoreCandidate aplica scoring a todos os videos de um candidate.
func (e *ScoringEngine) ScoreCandidate(candidate *PlaylistCandidate, ctx *ScoringContext) {
	totalPlaylistScore := 0.0

	for i := range candidate.Videos {
		sv := &candidate.Videos[i]
		finalScore := sv.Score // manter score base da estrategia

		for _, rule := range e.rules {
			ruleScore := rule.Score(sv.Video, ctx) * rule.Weight()
			if ruleScore != 0 {
				finalScore += ruleScore
				sv.Reasons = append(sv.Reasons,
					fmt.Sprintf("%s:%+.1f", rule.Name(), ruleScore))
			}
		}

		sv.Score = finalScore
		totalPlaylistScore += finalScore
	}

	candidate.TotalScore = totalPlaylistScore

	// Reordenar videos por score (maior primeiro) exceto para series (manter ordem)
	if candidate.GroupMode != "folder" && candidate.GroupMode != "prefix" {
		sort.Slice(candidate.Videos, func(i, j int) bool {
			return candidate.Videos[i].Score > candidate.Videos[j].Score
		})
	}
}

// ScoreAll aplica scoring a todos os candidates.
func (e *ScoringEngine) ScoreAll(candidates []PlaylistCandidate, allVideos []ClassifiedVideo, behavior *BehaviorProfile, exclusions map[int]map[int]bool) []PlaylistCandidate {
	for i := range candidates {
		excl := map[int]bool{}
		// Buscar exclusions relevantes para este candidate (por source key)
		// Na pratica, o playlistID e resolvido na persistencia
		ctx := &ScoringContext{
			Candidate:  &candidates[i],
			AllVideos:  allVideos,
			Behavior:   behavior,
			Exclusions: excl,
		}
		e.ScoreCandidate(&candidates[i], ctx)
	}

	// Ordenar candidates por score total decrescente
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalScore > candidates[j].TotalScore
	})

	return candidates
}

// ---------------------------------------------------------------------------
// Regras de scoring concretas
// ---------------------------------------------------------------------------

// EpisodeSequenceRule: bonus para videos que formam sequencia detectavel.
type EpisodeSequenceRule struct{}

func (r *EpisodeSequenceRule) Name() string    { return "episode_sequence" }
func (r *EpisodeSequenceRule) Weight() float64 { return 1.5 }
func (r *EpisodeSequenceRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	name := strings.ToLower(v.Video.Name)
	if EpisodePattern.MatchString(name) || EpisodeNumeric.MatchString(name) {
		return 20.0
	}
	if SequentialNumber.MatchString(name) {
		return 10.0
	}
	return 0
}

// MetadataRichnessRule: bonus para videos com metadados completos.
// Videos com metadados permitem classificacao mais confiavel.
type MetadataRichnessRule struct{}

func (r *MetadataRichnessRule) Name() string    { return "metadata_richness" }
func (r *MetadataRichnessRule) Weight() float64 { return 0.5 }
func (r *MetadataRichnessRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	if v.Video.Meta == nil {
		return -5.0 // leve penalidade por falta de metadados
	}
	score := 0.0
	if v.Video.Meta.Duration > 0 {
		score += 3
	}
	if v.Video.Meta.Width > 0 && v.Video.Meta.Height > 0 {
		score += 2
	}
	if v.Video.Meta.AudioChannels > 0 {
		score += 1
	}
	return score
}

// QualityRule: videos em alta resolucao recebem bonus.
type QualityRule struct{}

func (r *QualityRule) Name() string    { return "quality" }
func (r *QualityRule) Weight() float64 { return 0.3 }
func (r *QualityRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	if v.Video.Meta == nil {
		return 0
	}
	h := v.Video.Meta.Height
	switch {
	case h >= 2160: // 4K
		return 15
	case h >= 1080: // Full HD
		return 10
	case h >= 720: // HD
		return 5
	case h >= 480: // SD
		return 2
	default:
		return 0
	}
}

// DurationCoherenceRule: bonus quando a duracao do video e consistente
// com os outros videos da mesma playlist (indica conteudo homogeneo).
type DurationCoherenceRule struct{}

func (r *DurationCoherenceRule) Name() string    { return "duration_coherence" }
func (r *DurationCoherenceRule) Weight() float64 { return 0.8 }
func (r *DurationCoherenceRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	if v.Video.Meta == nil || v.Video.Meta.Duration <= 0 || ctx.Candidate == nil {
		return 0
	}

	// Calcular duracao media dos outros videos do candidate
	var totalDuration float64
	var count int
	for _, sv := range ctx.Candidate.Videos {
		if sv.Video.Video.ID != v.Video.ID && sv.Video.Video.Meta != nil && sv.Video.Video.Meta.Duration > 0 {
			totalDuration += sv.Video.Video.Meta.Duration
			count++
		}
	}
	if count == 0 {
		return 0
	}

	avgDuration := totalDuration / float64(count)
	diff := math.Abs(v.Video.Meta.Duration-avgDuration) / avgDuration

	if diff < 0.15 {
		return 15.0 // muito consistente
	}
	if diff < 0.3 {
		return 8.0 // razoavelmente consistente
	}
	if diff > 1.0 {
		return -10.0 // muito diferente, provavelmente nao pertence
	}
	return 0
}

// RecencyRule: videos mais recentes recebem leve bonus.
type RecencyRule struct{}

func (r *RecencyRule) Name() string    { return "recency" }
func (r *RecencyRule) Weight() float64 { return 0.4 }
func (r *RecencyRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	days := daysSince(v.Video.CreatedAt)
	switch {
	case days < 1:
		return 15
	case days < 7:
		return 10
	case days < 30:
		return 5
	case days < 90:
		return 2
	default:
		return 0
	}
}

// BehaviorRule: ajusta score com base no comportamento do usuario.
type BehaviorRule struct{}

func (r *BehaviorRule) Name() string    { return "behavior" }
func (r *BehaviorRule) Weight() float64 { return 2.0 }
func (r *BehaviorRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	if ctx.Behavior == nil {
		return 0
	}

	score := 0.0

	// Afinidade por tipo
	if affinity, ok := ctx.Behavior.PreferredTypes[v.Classification]; ok {
		score += affinity * 15 // ate +15 para tipo favorito
	}

	// Penalizar se esta na lista de skipped
	for _, id := range ctx.Behavior.RecentlySkipped {
		if id == v.Video.ID {
			score -= 25
			break
		}
	}

	// Penalizar se ja completado recentemente (evitar repeticao)
	for _, id := range ctx.Behavior.RecentlyCompleted {
		if id == v.Video.ID {
			score -= 15
			break
		}
	}

	return score
}

// ExclusionRule: penaliza fortemente videos excluidos pelo usuario.
type ExclusionRule struct{}

func (r *ExclusionRule) Name() string    { return "exclusion" }
func (r *ExclusionRule) Weight() float64 { return 1.0 }
func (r *ExclusionRule) Score(v ClassifiedVideo, ctx *ScoringContext) float64 {
	if ctx.Exclusions != nil && ctx.Exclusions[v.Video.ID] {
		return -1000 // praticamente elimina o video
	}
	return 0
}

// ClassificationConfidenceRule: bonus proporcional a confianca da classificacao.
type ClassificationConfidenceRule struct{}

func (r *ClassificationConfidenceRule) Name() string    { return "classification_confidence" }
func (r *ClassificationConfidenceRule) Weight() float64 { return 0.6 }
func (r *ClassificationConfidenceRule) Score(v ClassifiedVideo, _ *ScoringContext) float64 {
	// Confianca alta = video classificado com certeza = melhor para a playlist
	return v.Confidence * 10
}

// ---------------------------------------------------------------------------
// Default rules
// ---------------------------------------------------------------------------

func DefaultScoringRules() []ScoringRule {
	return []ScoringRule{
		&EpisodeSequenceRule{},
		&MetadataRichnessRule{},
		&QualityRule{},
		&DurationCoherenceRule{},
		&RecencyRule{},
		&BehaviorRule{},
		&ExclusionRule{},
		&ClassificationConfidenceRule{},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func daysSince(t time.Time) float64 {
	if t.IsZero() {
		return 365 // trata como muito antigo
	}
	return time.Since(t).Hours() / 24
}
