package playlist

import (
	"errors"
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// State Machine — controla o ciclo de vida da reproducao com transicoes
// bem definidas. Cada estado so aceita certos eventos.
//
// Diagrama:
//
//   idle ──start──> loading ──loaded──> playing
//                                         │
//                          ┌───pause──────┘
//                          │              │
//                        paused ──resume──┘
//                          │
//                   ┌──────┤
//                   │      └──abandon──> abandoned
//                   │
//            playing ──complete──> completed
//                   │
//                   └──skip──> skipped ──start──> loading
//                   │
//                   └──error──> errored ──retry──> loading
//
// ---------------------------------------------------------------------------

type PlaybackState string

const (
	StateIdle      PlaybackState = "idle"
	StateLoading   PlaybackState = "loading"
	StatePlaying   PlaybackState = "playing"
	StatePaused    PlaybackState = "paused"
	StateCompleted PlaybackState = "completed"
	StateAbandoned PlaybackState = "abandoned"
	StateSkipped   PlaybackState = "skipped"
	StateErrored   PlaybackState = "errored"
)

type PlaybackEvent string

const (
	EventStart    PlaybackEvent = "start"
	EventLoaded   PlaybackEvent = "loaded"
	EventPause    PlaybackEvent = "pause"
	EventResume   PlaybackEvent = "resume"
	EventComplete PlaybackEvent = "complete"
	EventSkip     PlaybackEvent = "skip"
	EventAbandon  PlaybackEvent = "abandon"
	EventError    PlaybackEvent = "error"
	EventRetry    PlaybackEvent = "retry"
)

// Transition define uma transicao valida.
type Transition struct {
	From  PlaybackState
	Event PlaybackEvent
	To    PlaybackState
}

// PlaybackStateMachine gerencia o estado de reproducao de um cliente.
type PlaybackStateMachine struct {
	ClientID    string
	VideoID     int
	PlaylistID  int
	State       PlaybackState
	Position    float64
	Duration    float64
	StartedAt   time.Time
	LastEventAt time.Time

	transitions map[transitionKey]PlaybackState
	listeners   []StateChangeListener
}

type transitionKey struct {
	from  PlaybackState
	event PlaybackEvent
}

// StateChangeListener e notificado quando o estado muda.
// Usado para emitir BehaviorEvents.
type StateChangeListener interface {
	OnStateChange(machine *PlaybackStateMachine, from PlaybackState, event PlaybackEvent, to PlaybackState)
}

func NewPlaybackStateMachine(clientID string) *PlaybackStateMachine {
	sm := &PlaybackStateMachine{
		ClientID: clientID,
		State:    StateIdle,
	}
	sm.transitions = buildTransitionTable()
	return sm
}

// AddListener registra um listener de mudanca de estado.
func (sm *PlaybackStateMachine) AddListener(l StateChangeListener) {
	sm.listeners = append(sm.listeners, l)
}

// HandleEvent processa um evento e transiciona o estado.
func (sm *PlaybackStateMachine) HandleEvent(event PlaybackEvent, videoID int, playlistID int, position float64, duration float64) error {
	key := transitionKey{from: sm.State, event: event}
	nextState, ok := sm.transitions[key]
	if !ok {
		return fmt.Errorf("transicao invalida: estado=%s evento=%s", sm.State, event)
	}

	prevState := sm.State

	// Atualizar campos conforme o evento
	switch event {
	case EventStart:
		sm.VideoID = videoID
		sm.PlaylistID = playlistID
		sm.Position = 0
		sm.Duration = duration
		sm.StartedAt = time.Now()
	case EventLoaded:
		sm.Duration = duration
	case EventPause, EventResume:
		sm.Position = position
	case EventComplete:
		sm.Position = duration
	case EventSkip:
		sm.Position = position
	case EventAbandon:
		sm.Position = position
	}

	sm.State = nextState
	sm.LastEventAt = time.Now()

	// Notificar listeners
	for _, l := range sm.listeners {
		l.OnStateChange(sm, prevState, event, nextState)
	}

	return nil
}

// CanTransition verifica se um evento e valido no estado atual.
func (sm *PlaybackStateMachine) CanTransition(event PlaybackEvent) bool {
	key := transitionKey{from: sm.State, event: event}
	_, ok := sm.transitions[key]
	return ok
}

// IsActive retorna true se esta em estado de reproducao ativa.
func (sm *PlaybackStateMachine) IsActive() bool {
	return sm.State == StatePlaying || sm.State == StatePaused || sm.State == StateLoading
}

// IsTerminal retorna true se o estado atual e terminal (precisa de novo start).
func (sm *PlaybackStateMachine) IsTerminal() bool {
	return sm.State == StateCompleted || sm.State == StateAbandoned || sm.State == StateSkipped
}

// WatchedPercent retorna a porcentagem assistida.
func (sm *PlaybackStateMachine) WatchedPercent() float64 {
	if sm.Duration <= 0 {
		return 0
	}
	pct := (sm.Position / sm.Duration) * 100
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

// ShouldAutoAdvance retorna true se a playlist deve avancar pro proximo video.
func (sm *PlaybackStateMachine) ShouldAutoAdvance() bool {
	return sm.State == StateCompleted || sm.State == StateSkipped
}

// ToBehaviorEvent converte o estado atual em um BehaviorEvent para persistencia.
func (sm *PlaybackStateMachine) ToBehaviorEvent(event PlaybackEvent) BehaviorEvent {
	var eventType BehaviorEventType
	switch event {
	case EventStart:
		eventType = EventStarted
	case EventPause:
		eventType = EventPaused
	case EventResume:
		eventType = EventResumed
	case EventComplete:
		eventType = EventCompleted
	case EventSkip:
		eventType = EventSkipped
	case EventAbandon:
		eventType = EventAbandoned
	default:
		eventType = BehaviorEventType(event)
	}

	return BehaviorEvent{
		ClientID:       sm.ClientID,
		VideoID:        sm.VideoID,
		PlaylistID:     sm.PlaylistID,
		EventType:      eventType,
		Position:       sm.Position,
		Duration:       sm.Duration,
		WatchedPercent: sm.WatchedPercent(),
		Timestamp:      time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tabela de transicoes
// ---------------------------------------------------------------------------

func buildTransitionTable() map[transitionKey]PlaybackState {
	transitions := []Transition{
		// Idle
		{StateIdle, EventStart, StateLoading},

		// Loading
		{StateLoading, EventLoaded, StatePlaying},
		{StateLoading, EventError, StateErrored},

		// Playing
		{StatePlaying, EventPause, StatePaused},
		{StatePlaying, EventComplete, StateCompleted},
		{StatePlaying, EventSkip, StateSkipped},
		{StatePlaying, EventAbandon, StateAbandoned},
		{StatePlaying, EventError, StateErrored},

		// Paused
		{StatePaused, EventResume, StatePlaying},
		{StatePaused, EventSkip, StateSkipped},
		{StatePaused, EventAbandon, StateAbandoned},
		{StatePaused, EventStart, StateLoading}, // iniciar outro video

		// Terminal states → reset
		{StateCompleted, EventStart, StateLoading},
		{StateSkipped, EventStart, StateLoading},
		{StateAbandoned, EventStart, StateLoading},
		{StateErrored, EventRetry, StateLoading},
		{StateErrored, EventStart, StateLoading},
	}

	table := make(map[transitionKey]PlaybackState, len(transitions))
	for _, t := range transitions {
		table[transitionKey{from: t.From, event: t.Event}] = t.To
	}
	return table
}

// ---------------------------------------------------------------------------
// Validacoes para uso no service
// ---------------------------------------------------------------------------

// ValidatePlaybackUpdate valida se uma atualizacao de estado e consistente.
func ValidatePlaybackUpdate(currentTime, duration float64, completed bool) error {
	if currentTime < 0 {
		return errors.New("posicao nao pode ser negativa")
	}
	if duration < 0 {
		return errors.New("duracao nao pode ser negativa")
	}
	if duration > 0 && currentTime > duration {
		return errors.New("posicao nao pode exceder duracao")
	}
	if completed && duration > 0 && currentTime < duration*0.8 {
		return errors.New("video marcado como completo mas posicao indica menos de 80% assistido")
	}
	return nil
}
