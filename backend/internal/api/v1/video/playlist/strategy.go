package playlist

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Strategy Pattern — cada estrategia sabe montar playlists de um jeito
// ---------------------------------------------------------------------------

type PlaylistStrategy interface {
	Name() string
	Build(ctx *PlaylistContext) []PlaylistCandidate
}

// ---------------------------------------------------------------------------
// ByFolderStrategy: agrupa videos que compartilham a mesma pasta,
// mas so se a pasta for "forte" (>= minItems e nome nao generico).
// ---------------------------------------------------------------------------

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

		// Determinar classificacao dominante da pasta
		classCount := map[VideoClassification]int{}
		for _, v := range videos {
			classCount[v.Classification]++
		}
		dominant := dominantClassification(classCount)

		scored := make([]ScoredVideo, 0, len(videos))
		for _, v := range videos {
			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   1.0, // score base, sera ajustado pelo scorer
				Reasons: []string{"same_folder"},
			})
		}

		// Ordenar por nome para manter ordem natural de episodios
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Video.Video.Name < scored[j].Video.Video.Name
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

// ---------------------------------------------------------------------------
// SequentialSeriesStrategy: detecta series por prefixo compartilhado
// e padrao de episodio, mesmo que os arquivos estejam em pastas diferentes.
// Isso resolve o problema de agrupamento cross-folder.
// ---------------------------------------------------------------------------

type SequentialSeriesStrategy struct {
	MinItemsPerGroup int
}

func NewSequentialSeriesStrategy() *SequentialSeriesStrategy {
	return &SequentialSeriesStrategy{MinItemsPerGroup: 2}
}

func (s *SequentialSeriesStrategy) Name() string { return "sequential_series" }

func (s *SequentialSeriesStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	// Agrupar por prefixo de titulo (cross-folder)
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

			// Bonus se tem padrao de episodio explicito
			if EpisodePattern.MatchString(strings.ToLower(v.Video.Name)) {
				reasons = append(reasons, "episode_pattern")
			}

			scored = append(scored, ScoredVideo{
				Video:   *v,
				Score:   1.0,
				Reasons: reasons,
			})
		}

		// Ordenar por nome para ordem natural de episodios
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Video.Video.Name < scored[j].Video.Video.Name
		})

		// Classificacao dominante do grupo
		classCount := map[VideoClassification]int{}
		for _, v := range videos {
			classCount[v.Classification]++
		}

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "series:" + prefix,
			Name:           formatSeriesName(prefix),
			PlaylistType:   "series",
			GroupMode:      "prefix",
			Classification: dominantClassification(classCount),
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

// ---------------------------------------------------------------------------
// ContinueWatchingStrategy: prioriza videos em progresso e sugere proximos.
// Funciona por cliente — precisa de PlaybackState no contexto.
// ---------------------------------------------------------------------------

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

	// So gera playlist "continue watching" se o video esta em progresso
	if state.Completed || state.CurrentTime <= 0 {
		return nil
	}

	scored := []ScoredVideo{
		{
			Video:   *current,
			Score:   100.0, // score maximo para o video em andamento
			Reasons: []string{"in_progress", fmt.Sprintf("position:%.0fs", state.CurrentTime)},
		},
	}

	// Adicionar videos da mesma pasta ou serie como proximos
	sameFolder := ctx.VideosByFolder[current.Video.ParentPath]
	for _, v := range sameFolder {
		if v.Video.ID == current.Video.ID {
			continue
		}
		// So adicionar se vem depois (ordem alfabetica)
		if v.Video.Name > current.Video.Name {
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
			Name:           "Continue assistindo",
			PlaylistType:   "continue",
			GroupMode:      "behavior",
			Classification: current.Classification,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

// ---------------------------------------------------------------------------
// RelatedContentStrategy: busca videos similares por metadados.
// Combina duracao parecida + mesma classificacao + resolucao similar.
// ---------------------------------------------------------------------------

type RelatedContentStrategy struct {
	MaxResults        int
	DurationTolerance float64 // percentual de tolerancia (0.3 = 30%)
}

func NewRelatedContentStrategy() *RelatedContentStrategy {
	return &RelatedContentStrategy{
		MaxResults:        20,
		DurationTolerance: 0.3,
	}
}

func (s *RelatedContentStrategy) Name() string { return "related_content" }

func (s *RelatedContentStrategy) Build(ctx *PlaylistContext) []PlaylistCandidate {
	// Agrupar por classificacao
	byClass := map[VideoClassification][]*ClassifiedVideo{}
	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		byClass[v.Classification] = append(byClass[v.Classification], v)
	}

	var candidates []PlaylistCandidate

	for class, videos := range byClass {
		if class == ClassProgram || class == ClassClip {
			continue // nao criar playlists para programas/samples ou clips avulsos
		}
		if len(videos) < 2 {
			continue
		}

		scored := make([]ScoredVideo, 0, len(videos))
		for _, v := range videos {
			score := 1.0
			reasons := []string{fmt.Sprintf("classification:%s", class)}

			// Bonus por ter metadados ricos
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

		// Limitar resultados
		if len(scored) > s.MaxResults {
			sort.Slice(scored, func(i, j int) bool {
				return scored[i].Score > scored[j].Score
			})
			scored = scored[:s.MaxResults]
		}

		candidates = append(candidates, PlaylistCandidate{
			SourceKey:      "related:" + string(class),
			Name:           classificationDisplayName(class),
			PlaylistType:   playlistTypeFor(class),
			GroupMode:      "classification",
			Classification: class,
			Strategy:       s.Name(),
			Videos:         scored,
		})
	}

	return candidates
}

// ---------------------------------------------------------------------------
// PriorityFavoritesStrategy: mistura favoritos, recentes e nao assistidos.
// Usa comportamento do usuario para priorizar.
// ---------------------------------------------------------------------------

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

	// Criar set dos recentemente assistidos e completados
	recentlyWatched := toSet(behavior.RecentlyWatched)
	recentlyCompleted := toSet(behavior.RecentlyCompleted)
	recentlySkipped := toSet(behavior.RecentlySkipped)

	var scored []ScoredVideo
	for i := range ctx.Videos {
		v := &ctx.Videos[i]
		score := 0.0
		var reasons []string

		// Nunca assistido = oportunidade de descoberta
		if !recentlyWatched[v.Video.ID] && !recentlyCompleted[v.Video.ID] {
			score += 30
			reasons = append(reasons, "unwatched:+30")
		}

		// Tipo preferido do usuario
		if affinity, ok := behavior.PreferredTypes[v.Classification]; ok && affinity > 0.5 {
			bonus := affinity * 25
			score += bonus
			reasons = append(reasons, fmt.Sprintf("preferred_type:+%.0f", bonus))
		}

		// Duracao na faixa preferida
		if v.Video.Meta != nil && behavior.PreferredDurations.MaxSeconds > 0 {
			if v.Video.Meta.Duration >= behavior.PreferredDurations.MinSeconds &&
				v.Video.Meta.Duration <= behavior.PreferredDurations.MaxSeconds {
				score += 15
				reasons = append(reasons, "preferred_duration:+15")
			}
		}

		// Penalizar videos pulados recentemente
		if recentlySkipped[v.Video.ID] {
			score -= 40
			reasons = append(reasons, "recently_skipped:-40")
		}

		// Penalizar videos ja completados (menos incentivo a re-assistir)
		if recentlyCompleted[v.Video.ID] {
			score -= 20
			reasons = append(reasons, "already_completed:-20")
		}

		// Recencia do arquivo (videos mais novos ganham leve bonus)
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

	// Ordenar por score decrescente
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > s.MaxResults {
		scored = scored[:s.MaxResults]
	}

	return []PlaylistCandidate{
		{
			SourceKey:      fmt.Sprintf("favorites:%s", ctx.ClientID),
			Name:           "Recomendados para voce",
			PlaylistType:   "mixed",
			GroupMode:      "behavior",
			Classification: ClassPersonal,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

// ---------------------------------------------------------------------------
// RandomSmartStrategy: embaralha, mas com inteligencia — evita repeticao,
// mantem coerencia de tipo e balanceia duracao.
// ---------------------------------------------------------------------------

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

	// Estrategia: distribuir tipos uniformemente para variedade
	byClass := map[VideoClassification][]ClassifiedVideo{}
	for _, v := range ctx.Videos {
		if v.Classification != ClassProgram {
			byClass[v.Classification] = append(byClass[v.Classification], v)
		}
	}

	// Round-robin por classificacao
	var scored []ScoredVideo
	classKeys := make([]VideoClassification, 0, len(byClass))
	for k := range byClass {
		classKeys = append(classKeys, k)
	}
	// Sort for deterministic output
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
				Score:   float64(s.MaxResults - len(scored)), // score decrescente = ordem de inclusao
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
			Name:           "Mix variado",
			PlaylistType:   "mixed",
			GroupMode:      "random",
			Classification: ClassPersonal,
			Strategy:       s.Name(),
			Videos:         scored,
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func dominantClassification(counts map[VideoClassification]int) VideoClassification {
	best := ClassPersonal
	bestCount := 0
	for class, count := range counts {
		if count > bestCount {
			best = class
			bestCount = count
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

func classificationDisplayName(class VideoClassification) string {
	names := map[VideoClassification]string{
		ClassSeries:   "Series",
		ClassMovie:    "Filmes",
		ClassAnime:    "Animes",
		ClassCourse:   "Cursos",
		ClassPersonal: "Videos pessoais",
		ClassClip:     "Clips",
		ClassMusic:    "Videos musicais",
		ClassProgram:  "Programas",
	}
	if name, ok := names[class]; ok {
		return name
	}
	return "Videos"
}

func formatSeriesName(prefix string) string {
	if prefix == "" {
		return "Serie desconhecida"
	}
	// Capitalizar primeira letra de cada palavra
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
