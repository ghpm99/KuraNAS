package updater

type ServiceInterface interface {
	CheckForUpdate() (UpdateStatusDto, error)
	DownloadAndApply() error
}
