package libraries

type LibraryDto struct {
	Category string `json:"category"`
	Path     string `json:"path"`
}

type UpdateLibraryDto struct {
	Path string `json:"path" binding:"required"`
}

func (m *LibraryModel) ToDto() LibraryDto {
	return LibraryDto{
		Category: string(m.Category),
		Path:     m.Path,
	}
}

func (d *LibraryDto) ToModel() LibraryModel {
	return LibraryModel{
		Category: LibraryCategory(d.Category),
		Path:     d.Path,
	}
}
