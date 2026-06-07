package distribution

// Artifact is the on-disk shape of a distributable client app, as described by
// downloads/manifest.json. The manifest is produced by CI when it builds the
// Android APKs and packages the browser extension; the server only hosts the
// files, it never builds them.
type Artifact struct {
	ID             string `json:"id"`
	Platform       string `json:"platform"`
	NameKey        string `json:"name_key"`
	DescriptionKey string `json:"description_key"`
	File           string `json:"file"`
	Version        string `json:"version"`
	MinOS          string `json:"min_os"`
	SHA256         string `json:"sha256"`

	// SizeBytes is filled by the repository from the file on disk, not read
	// from the manifest, so it always reflects the artifact actually served.
	SizeBytes int64 `json:"-"`
	// AbsPath is the resolved location of the file on disk, filled by the
	// repository. It is never serialized.
	AbsPath string `json:"-"`
}

// manifest is the root document of downloads/manifest.json.
type manifest struct {
	Artifacts []Artifact `json:"artifacts"`
}
