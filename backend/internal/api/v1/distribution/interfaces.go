package distribution

type RepositoryInterface interface {
	// ListArtifacts returns every artifact described by the manifest whose file
	// is actually present on disk, with SizeBytes and AbsPath filled in.
	ListArtifacts() ([]Artifact, error)
	// GetArtifact returns a single present artifact by id, with AbsPath filled in.
	GetArtifact(id string) (Artifact, error)
}

type ServiceInterface interface {
	// ListDownloads returns the catalog of available client apps, with names
	// already translated to the active locale.
	ListDownloads() ([]DownloadItemDto, error)
	// ResolveDownload returns the absolute file path and the download filename
	// for a given artifact id.
	ResolveDownload(id string) (path string, filename string, err error)
}
