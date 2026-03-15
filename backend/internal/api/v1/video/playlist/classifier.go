package playlist

import (
	"path/filepath"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Specification Pattern — cada regra e uma spec testavel isoladamente
// ---------------------------------------------------------------------------

// ClassificationSpec avalia se um video satisfaz uma condicao de classificacao.
type ClassificationSpec interface {
	IsSatisfiedBy(video VideoEntry) bool
	Name() string
	Confidence() float64 // quao confiante essa spec e quando satisfeita
}

// ClassificationRule associa uma spec a uma classificacao.
type ClassificationRule struct {
	Spec           ClassificationSpec
	Classification VideoClassification
	Priority       int // menor = maior prioridade
}

// VideoClassifier aplica um conjunto de regras em ordem de prioridade.
// A primeira regra satisfeita com maior prioridade vence.
type VideoClassifier struct {
	rules []ClassificationRule
}

func NewVideoClassifier() *VideoClassifier {
	return &VideoClassifier{
		rules: defaultClassificationRules(),
	}
}

// Classify retorna a classificacao do video e a confianca.
func (c *VideoClassifier) Classify(video VideoEntry) ClassifiedVideo {
	var bestMatch *ClassificationRule
	var bestConfidence float64

	for i := range c.rules {
		rule := &c.rules[i]
		if rule.Spec.IsSatisfiedBy(video) {
			conf := rule.Spec.Confidence()
			// Prioridade menor vence; em caso de empate, maior confianca vence
			if bestMatch == nil || rule.Priority < bestMatch.Priority ||
				(rule.Priority == bestMatch.Priority && conf > bestConfidence) {
				bestMatch = rule
				bestConfidence = conf
			}
		}
	}

	if bestMatch == nil {
		return ClassifiedVideo{
			Video:          video,
			Classification: ClassPersonal,
			Confidence:     0.1,
			MatchedRules:   []string{"fallback_personal"},
		}
	}

	return ClassifiedVideo{
		Video:          video,
		Classification: bestMatch.Classification,
		Confidence:     bestConfidence,
		MatchedRules:   []string{bestMatch.Spec.Name()},
	}
}

// ClassifyAll classifica todos os videos e monta os indices do contexto.
func (c *VideoClassifier) ClassifyAll(videos []VideoEntry) []ClassifiedVideo {
	result := make([]ClassifiedVideo, 0, len(videos))
	for _, v := range videos {
		result = append(result, c.Classify(v))
	}
	return result
}

// ---------------------------------------------------------------------------
// Specs concretas
// ---------------------------------------------------------------------------

// pathContainsSpec checa se o path contem alguma das keywords.
type pathContainsSpec struct {
	name       string
	keywords   []string
	confidence float64
}

func (s *pathContainsSpec) Name() string        { return s.name }
func (s *pathContainsSpec) Confidence() float64 { return s.confidence }
func (s *pathContainsSpec) IsSatisfiedBy(v VideoEntry) bool {
	lower := strings.ToLower(v.Path + " " + v.ParentPath + " " + v.Name)
	for _, kw := range s.keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// episodePatternSpec detecta padroes de episodio no nome do arquivo.
type episodePatternSpec struct {
	pattern *regexp.Regexp
}

func (s *episodePatternSpec) Name() string        { return "episode_pattern" }
func (s *episodePatternSpec) Confidence() float64 { return 0.9 }
func (s *episodePatternSpec) IsSatisfiedBy(v VideoEntry) bool {
	return s.pattern.MatchString(strings.ToLower(v.Name))
}

// durationRangeSpec classifica por duracao (requer metadados).
type durationRangeSpec struct {
	name       string
	minSeconds float64
	maxSeconds float64
	confidence float64
}

func (s *durationRangeSpec) Name() string        { return s.name }
func (s *durationRangeSpec) Confidence() float64 { return s.confidence }
func (s *durationRangeSpec) IsSatisfiedBy(v VideoEntry) bool {
	if v.Meta == nil || v.Meta.Duration <= 0 {
		return false
	}
	return v.Meta.Duration >= s.minSeconds && v.Meta.Duration <= s.maxSeconds
}

// resolutionSpec classifica por resolucao (ex: >= 720p para conteudo "profissional").
type resolutionSpec struct {
	name       string
	minHeight  int
	confidence float64
}

func (s *resolutionSpec) Name() string        { return s.name }
func (s *resolutionSpec) Confidence() float64 { return s.confidence }
func (s *resolutionSpec) IsSatisfiedBy(v VideoEntry) bool {
	if v.Meta == nil {
		return false
	}
	return v.Meta.Height >= s.minHeight
}

// coursePatternSpec detecta padroes de curso/tutorial (numeracao sequencial + keywords).
type coursePatternSpec struct {
	pattern *regexp.Regexp
}

func (s *coursePatternSpec) Name() string        { return "course_pattern" }
func (s *coursePatternSpec) Confidence() float64 { return 0.85 }
func (s *coursePatternSpec) IsSatisfiedBy(v VideoEntry) bool {
	lower := strings.ToLower(v.Path + " " + v.ParentPath + " " + v.Name)
	keywords := []string{"curso", "course", "aula", "lesson", "lecture", "tutorial", "module", "modulo"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return s.pattern.MatchString(strings.ToLower(v.Name))
}

// clipSpec detecta clips curtos (< 60s ou keywords como "clip", "meme", "shorts").
type clipSpec struct{}

func (s *clipSpec) Name() string        { return "clip_short_video" }
func (s *clipSpec) Confidence() float64 { return 0.7 }
func (s *clipSpec) IsSatisfiedBy(v VideoEntry) bool {
	lower := strings.ToLower(v.Name + " " + v.ParentPath)
	keywords := []string{"clip", "meme", "shorts", "tiktok", "reel"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	// Se temos metadados e duracao < 60s, e provavelmente um clip
	if v.Meta != nil && v.Meta.Duration > 0 && v.Meta.Duration < 60 {
		return true
	}
	return false
}

// musicVideoSpec detecta videos musicais.
type musicVideoSpec struct{}

func (s *musicVideoSpec) Name() string        { return "music_video" }
func (s *musicVideoSpec) Confidence() float64 { return 0.75 }
func (s *musicVideoSpec) IsSatisfiedBy(v VideoEntry) bool {
	lower := strings.ToLower(v.Path + " " + v.ParentPath + " " + v.Name)
	keywords := []string{"music", "musica", "mv", "videoclip", "videoclipe", "karaoke"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// compositeSpec combina multiplas specs com AND.
type compositeSpec struct {
	name       string
	specs      []ClassificationSpec
	confidence float64
}

func (s *compositeSpec) Name() string        { return s.name }
func (s *compositeSpec) Confidence() float64 { return s.confidence }
func (s *compositeSpec) IsSatisfiedBy(v VideoEntry) bool {
	for _, spec := range s.specs {
		if !spec.IsSatisfiedBy(v) {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Regras default ordenadas por prioridade
// ---------------------------------------------------------------------------

func defaultClassificationRules() []ClassificationRule {
	episodeRe := regexp.MustCompile(`(?i)s\d{1,2}e\d{1,2}|\d{1,2}x\d{1,2}|ep\.?\s?\d+|epis[oó]dio\s?\d+|season\s?\d+\s*episode\s?\d+|cap[ií]tulo\s?\d+`)
	courseRe := regexp.MustCompile(`(?i)(?:aula|lesson|lecture|module|modulo)\s*\d+`)

	return []ClassificationRule{
		// Prioridade 1: Padroes fortes baseados em convencoes de nomeacao
		{
			Spec:           &episodePatternSpec{pattern: episodeRe},
			Classification: ClassSeries,
			Priority:       1,
		},
		{
			Spec:           &coursePatternSpec{pattern: courseRe},
			Classification: ClassCourse,
			Priority:       1,
		},
		{
			Spec: &pathContainsSpec{
				name: "personal_capture_path",
				keywords: []string{
					"/camera", "/dcim", "/captura", "/captures", "/screenrecord",
					"/screen-record", "/screen recording", "/gravacoes", "/gravacao",
					"/obs", "/dvr", "/webcam", "/whatsapp video",
				},
				confidence: 0.85,
			},
			Classification: ClassPersonal,
			Priority:       1,
		},

		// Prioridade 2: Keywords no path
		{
			Spec:           &pathContainsSpec{name: "anime_path", keywords: []string{"/anime", "/animes"}, confidence: 0.85},
			Classification: ClassAnime,
			Priority:       2,
		},
		{
			Spec:           &pathContainsSpec{name: "series_path", keywords: []string{"/series", "/season", "/temporada"}, confidence: 0.8},
			Classification: ClassSeries,
			Priority:       2,
		},
		{
			Spec:           &pathContainsSpec{name: "movie_path", keywords: []string{"/movies", "/filmes", "/movie", "/filme"}, confidence: 0.8},
			Classification: ClassMovie,
			Priority:       2,
		},
		{
			Spec:           &musicVideoSpec{},
			Classification: ClassMusic,
			Priority:       2,
		},
		{
			Spec:           &pathContainsSpec{name: "program_path", keywords: []string{"steam", "program", "sample", "benchmark"}, confidence: 0.7},
			Classification: ClassProgram,
			Priority:       2,
		},

		// Prioridade 3: Classificacao por metadados
		{
			Spec:           &clipSpec{},
			Classification: ClassClip,
			Priority:       3,
		},
		{
			// Filme: video longo (> 60min) em alta resolucao, sem padrao de episodio
			Spec: &compositeSpec{
				name: "long_hd_movie",
				specs: []ClassificationSpec{
					&durationRangeSpec{name: "long_video", minSeconds: 3600, maxSeconds: 14400, confidence: 0.6},
					&resolutionSpec{name: "hd_resolution", minHeight: 720, confidence: 0.5},
				},
				confidence: 0.65,
			},
			Classification: ClassMovie,
			Priority:       3,
		},

		// Prioridade 4: Classificacao por duracao quando nenhuma outra regra match
		{
			Spec:           &durationRangeSpec{name: "medium_episode", minSeconds: 900, maxSeconds: 3600, confidence: 0.4},
			Classification: ClassSeries,
			Priority:       4,
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers exportados para uso no scoring
// ---------------------------------------------------------------------------

var (
	EpisodePattern   = regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})|(\d{1,2})x(\d{1,2})`)
	EpisodeNumeric   = regexp.MustCompile(`(?i)(?:ep\.?\s?|epis[oó]dio\s?|cap[ií]tulo\s?)(\d+)`)
	SequentialNumber = regexp.MustCompile(`(?:^|\D)(\d{1,3})(?:\D|$)`)
)

// InferTitlePrefix extrai o prefixo do titulo removendo sufixos de episodio.
func InferTitlePrefix(name string) string {
	noExt := strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
	if noExt == "" {
		return ""
	}

	bracketCleanup := regexp.MustCompile(`\[[^\]]+\]|\([^\)]+\)`)
	episodeInline := regexp.MustCompile(`(?i)[\s._-]*(s\d{1,2}e\d{1,2}|\d{1,2}x\d{1,2})[\s._-]*`)
	episodeSuffix := regexp.MustCompile(`(?i)[\s._-]*(ep\.?\s?\d+|epis[oó]dio\s?\d+|cap[ií]tulo\s?\d+|part\s?\d+|parte\s?\d+|\d{1,3})$`)
	spaceCollapse := regexp.MustCompile(`[\s._-]+`)

	value := strings.ToLower(noExt)
	value = bracketCleanup.ReplaceAllString(value, "")
	value = episodeInline.ReplaceAllString(value, " ")
	value = episodeSuffix.ReplaceAllString(value, "")
	value = spaceCollapse.ReplaceAllString(value, " ")
	value = strings.TrimSpace(value)
	return value
}

// IsGenericFolderName retorna true para nomes de pasta genericos demais.
func IsGenericFolderName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return true
	}
	generic := map[string]bool{
		"videos": true, "video": true, "movies": true, "filmes": true,
		"downloads": true, "clips": true, "desktop": true,
		"documentos": true, "documents": true, "media": true,
		"home": true, "root": true, "tmp": true, "temp": true,
		"new folder": true, "nova pasta": true, "misc": true,
		"outros": true, "other": true, "stuff": true,
	}
	return generic[lower]
}
