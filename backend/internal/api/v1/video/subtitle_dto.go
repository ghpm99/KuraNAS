package video

type SubtitleFileModel struct {
	ID   int
	Name string
}

type SubtitleDto struct {
	ID       int    `json:"id"`
	Label    string `json:"label"`
	Language string `json:"language"`
	Format   string `json:"format"`
	IsForced bool   `json:"is_forced"`
}
