package files

import (
	"path/filepath"
	"strings"
)

const uploadTemporarySuffix = ".kuranas-part"

func uploadTemporaryPathFor(destinationPath string) string {
	return filepath.Join(filepath.Dir(destinationPath), "."+filepath.Base(destinationPath)+uploadTemporarySuffix)
}

// IsUploadTemporaryFile reports whether path is the hidden staging copy that a
// cross-volume upload publication writes beside its destination. Indexing and
// watching must ignore it: it is renamed to the final name once complete.
func IsUploadTemporaryFile(path string) bool {
	return strings.HasSuffix(filepath.Base(path), uploadTemporarySuffix)
}
