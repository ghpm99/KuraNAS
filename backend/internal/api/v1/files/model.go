package files

type FileModel struct {
	ID              int
	Name            string
	Path            string
	Format          string
	Size            int
	UpdatedAt       string
	CreatedAt       string
	LastInteraction string
	LastBackup      string
}
