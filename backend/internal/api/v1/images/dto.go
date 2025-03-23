package images

type ImageDto struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Format          string `json:"format"`
	Size            int    `json:"size"`
	UpdatedAt       string `json:"updated_at"`
	CreatedAt       string `json:"created_at"`
	LastInteraction string `json:"last_interaction"`
	LastBackup      string `json:"last_backup"`
}
