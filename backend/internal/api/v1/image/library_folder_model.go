package image

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
	ImageCount  int
	CoverFileID int
}
