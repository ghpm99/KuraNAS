package playlist

import (
	"math"
	"sort"
)

// ---------------------------------------------------------------------------
// Event-Driven Behavior Learning
//
// Este modulo analisa eventos de comportamento do usuario para construir
// um perfil de preferencias. O perfil e usado pelo ScoringEngine e pelas
// Strategies para tomar decisoes mais inteligentes.
//
// Fluxo:
//   1. Cada acao do player gera um BehaviorEvent (via State Machine)
//   2. Eventos sao persistidos no banco
//   3. BehaviorAnalyzer processa os eventos e gera um BehaviorProfile
//   4. O profile alimenta o Scoring Engine e as estrategias
// ---------------------------------------------------------------------------

// BehaviorAnalyzer processa eventos historicos e gera perfis.
type BehaviorAnalyzer struct {
	// Quantos eventos recentes considerar para o perfil
	MaxEvents int
	// Peso dado a eventos mais recentes (exponential decay)
	RecencyDecayFactor float64
}

func NewBehaviorAnalyzer() *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		MaxEvents:          500,
		RecencyDecayFactor: 0.95,
	}
}

// BuildProfile constroi um BehaviorProfile a partir de eventos historicos
// e das classificacoes dos videos.
func (a *BehaviorAnalyzer) BuildProfile(
	events []BehaviorEvent,
	classifiedVideos map[int]*ClassifiedVideo,
) *BehaviorProfile {
	if len(events) == 0 {
		return nil
	}

	// Limitar a eventos mais recentes
	if len(events) > a.MaxEvents {
		events = events[len(events)-a.MaxEvents:]
	}

	profile := &BehaviorProfile{
		PreferredTypes: make(map[VideoClassification]float64),
	}

	// Contadores para calculos
	var completedCount, startedCount int
	var totalWatchDuration float64
	var watchedDurations []float64
	sessionVideos := map[string]int{} // clientID -> count para sessao media

	completedSet := map[int]bool{}
	skippedSet := map[int]bool{}
	watchedSet := map[int]bool{}

	// Contadores por classificacao
	typeCompleted := map[VideoClassification]int{}
	typeStarted := map[VideoClassification]int{}

	for _, event := range events {
		classification := ClassPersonal
		if cv, ok := classifiedVideos[event.VideoID]; ok {
			classification = cv.Classification
		}

		switch event.EventType {
		case EventStarted:
			startedCount++
			watchedSet[event.VideoID] = true
			sessionVideos[event.ClientID]++
			typeStarted[classification]++

		case EventCompleted:
			completedCount++
			completedSet[event.VideoID] = true
			typeCompleted[classification]++
			if event.Duration > 0 {
				totalWatchDuration += event.Duration
				watchedDurations = append(watchedDurations, event.Duration)
			}

		case EventSkipped:
			skippedSet[event.VideoID] = true

		case EventAbandoned:
			if event.Duration > 0 && event.Position > 0 {
				totalWatchDuration += event.Position
				watchedDurations = append(watchedDurations, event.Position)
			}
		}
	}

	// Taxa de completacao
	if startedCount > 0 {
		profile.CompletionRate = float64(completedCount) / float64(startedCount)
	}

	// Duracao media assistida
	if len(watchedDurations) > 0 {
		profile.AvgWatchDuration = totalWatchDuration / float64(len(watchedDurations))
	}

	// Sessao media (videos por sessao)
	if len(sessionVideos) > 0 {
		totalSessionVideos := 0
		for _, count := range sessionVideos {
			totalSessionVideos += count
		}
		profile.AvgSessionLength = totalSessionVideos / len(sessionVideos)
	}

	// Faixa de duracao preferida (percentil 25-75)
	if len(watchedDurations) >= 3 {
		sort.Float64s(watchedDurations)
		p25 := percentile(watchedDurations, 25)
		p75 := percentile(watchedDurations, 75)
		profile.PreferredDurations = DurationRange{
			MinSeconds: p25,
			MaxSeconds: p75,
		}
	}

	// Afinidade por tipo de conteudo
	for class, completed := range typeCompleted {
		started := typeStarted[class]
		if started > 0 {
			// Afinidade = taxa de completacao do tipo * peso por volume
			completionRate := float64(completed) / float64(started)
			volumeWeight := math.Min(float64(started)/10.0, 1.0) // normalizar
			profile.PreferredTypes[class] = completionRate * volumeWeight
		}
	}

	// Padroes de skip
	for class, started := range typeStarted {
		skipped := 0
		for _, event := range events {
			if event.EventType == EventSkipped {
				if cv, ok := classifiedVideos[event.VideoID]; ok && cv.Classification == class {
					skipped++
				}
			}
		}
		if started >= 3 && skipped > 0 {
			rate := float64(skipped) / float64(started)
			if rate >= 0.3 { // skip rate >= 30% e significativo
				profile.SkipPatterns = append(profile.SkipPatterns, SkipPattern{
					Classification: class,
					SkipRate:       rate,
					SampleSize:     started,
				})
			}
		}
	}

	// Listas recentes (ultimos 50 unicos)
	profile.RecentlyWatched = recentUniqueIDs(watchedSet, 50)
	profile.RecentlyCompleted = recentUniqueIDs(completedSet, 50)
	profile.RecentlySkipped = recentUniqueIDs(skippedSet, 50)

	return profile
}

// ---------------------------------------------------------------------------
// StateChangeListener que emite BehaviorEvents
// ---------------------------------------------------------------------------

// BehaviorEventEmitter implementa StateChangeListener e coleta eventos
// para posterior persistencia.
type BehaviorEventEmitter struct {
	Events []BehaviorEvent
}

func NewBehaviorEventEmitter() *BehaviorEventEmitter {
	return &BehaviorEventEmitter{}
}

func (e *BehaviorEventEmitter) OnStateChange(
	machine *PlaybackStateMachine,
	from PlaybackState,
	event PlaybackEvent,
	to PlaybackState,
) {
	// Emitir evento para estados significativos
	switch event {
	case EventStart, EventComplete, EventSkip, EventAbandon:
		e.Events = append(e.Events, machine.ToBehaviorEvent(event))
	case EventPause:
		// So emitir pause se assistiu pelo menos 10%
		if machine.WatchedPercent() >= 10 {
			e.Events = append(e.Events, machine.ToBehaviorEvent(event))
		}
	}
}

// Flush retorna todos os eventos acumulados e limpa o buffer.
func (e *BehaviorEventEmitter) Flush() []BehaviorEvent {
	events := e.Events
	e.Events = nil
	return events
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func recentUniqueIDs(set map[int]bool, limit int) []int {
	ids := make([]int, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	// Sort descending (mais recentes primeiro — IDs maiores = mais novos normalmente)
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids
}
