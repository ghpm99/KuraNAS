package video

type LibraryFolderScope struct {
	Prefix      string
	IsWholeRoot bool
	Label       string
}

type LibraryFolderQuery struct {
	Scopes    []LibraryFolderScope
	Separator string
	Limit     int
	Offset    int
}

type LibraryFolderModel struct {
	Path        string
	Name        string
	VideoCount  int
	CoverFileID int
}

type LibraryFolderRequest struct {
	ParentPath string
	Page       int
	PageSize   int
}

type LibraryFolderVideosRequest struct {
	FolderPath string
	Page       int
	PageSize   int
}
