package playlist

import "time"

// ---------------------------------------------------------------------------
// Video enrichido com metadados — unidade base de todo o engine
// ---------------------------------------------------------------------------

type VideoEntry struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Size       int64
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Metadados extraidos do worker (pode ser nil se nao processado ainda)
	Meta *VideoMeta
}

type VideoMeta struct {
	Duration        float64 // segundos
	Width           int
	Height          int
	FrameRate       float64
	CodecName       string
	AspectRatio     string
	AudioChannels   int
	AudioCodec      string
	AudioSampleRate string
}

// ---------------------------------------------------------------------------
// Classificacao de video (output do Rule Engine)
// ---------------------------------------------------------------------------

type VideoClassification string

const (
	ClassSeries   VideoClassification = "series"
	ClassMovie    VideoClassification = "movie"
	ClassAnime    VideoClassification = "anime"
	ClassCourse   VideoClassification = "course"
	ClassPersonal VideoClassification = "personal"
	ClassClip     VideoClassification = "clip"
	ClassMusic    VideoClassification = "music_video"
	ClassProgram  VideoClassification = "program"
)

type ClassifiedVideo struct {
	Video          VideoEntry
	Classification VideoClassification
	Confidence     float64  // 0.0–1.0, quao confiante a classificacao e
	MatchedRules   []string // nomes das regras que matcharam
}

// ---------------------------------------------------------------------------
// Contexto completo passado ao engine
// ---------------------------------------------------------------------------

type PlaylistContext struct {
	// Todos os videos classificados disponiveis
	Videos []ClassifiedVideo

	// Indice por ID para acesso rapido
	VideoByID map[int]*ClassifiedVideo

	// Indice por pasta para acesso rapido
	VideosByFolder map[string][]*ClassifiedVideo

	// Estado de reprodução do cliente atual (pode ser nil)
	ClientID      string
	PlaybackState *PlaybackSnapshot

	// Historico de comportamento do usuario
	Behavior *BehaviorProfile

	// Videos excluidos pelo usuario de playlists auto
	Exclusions map[int]map[int]bool // playlistID -> set[videoID]
}

// PlaybackSnapshot e um snapshot imutavel do estado de reproducao
type PlaybackSnapshot struct {
	VideoID     int
	PlaylistID  int
	CurrentTime float64
	Duration    float64
	Completed   bool
	IsPaused    bool
}

// ---------------------------------------------------------------------------
// Output do engine
// ---------------------------------------------------------------------------

type ScoredVideo struct {
	Video   ClassifiedVideo
	Score   float64
	Reasons []string // ex: ["same_folder:+20", "episode_sequence:+30"]
}

type PlaylistCandidate struct {
	SourceKey      string // chave unica: "folder:/path", "series:breaking_bad"
	Name           string
	PlaylistType   string // "folder", "series", "movie", "course", "mixed"
	GroupMode      string // "folder", "prefix", "classification", "behavior"
	Classification VideoClassification
	Strategy       string // nome da estrategia que gerou
	Videos         []ScoredVideo
	TotalScore     float64
}

// ---------------------------------------------------------------------------
// Comportamento do usuario (input para scoring e aprendizado)
// ---------------------------------------------------------------------------

type BehaviorProfile struct {
	// Taxas gerais
	CompletionRate   float64 // % de videos que o usuario termina
	AvgWatchDuration float64 // duracao media assistida em segundos
	AvgSessionLength int     // quantos videos por sessao

	// Preferencias aprendidas
	PreferredDurations DurationRange                   // faixa de duracao preferida
	PreferredTypes     map[VideoClassification]float64 // tipo -> afinidade (0-1)
	SkipPatterns       []SkipPattern

	// Historico recente
	RecentlyWatched   []int // IDs dos ultimos N videos assistidos
	RecentlySkipped   []int
	RecentlyCompleted []int
}

type DurationRange struct {
	MinSeconds float64
	MaxSeconds float64
}

type SkipPattern struct {
	Classification VideoClassification
	SkipRate       float64 // 0-1
	SampleSize     int
}

// ---------------------------------------------------------------------------
// Evento de comportamento (persistido para aprendizado)
// ---------------------------------------------------------------------------

type BehaviorEventType string

const (
	EventStarted   BehaviorEventType = "started"
	EventPaused    BehaviorEventType = "paused"
	EventResumed   BehaviorEventType = "resumed"
	EventCompleted BehaviorEventType = "completed"
	EventSkipped   BehaviorEventType = "skipped"
	EventAbandoned BehaviorEventType = "abandoned"
)

type BehaviorEvent struct {
	ClientID       string
	VideoID        int
	PlaylistID     int
	EventType      BehaviorEventType
	Position       float64 // posicao em segundos quando o evento ocorreu
	Duration       float64 // duracao total do video
	WatchedPercent float64 // % assistido ate o momento
	Timestamp      time.Time
}
