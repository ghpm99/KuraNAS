package files

import (
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func ensureTestIcons(t *testing.T) {
	t.Helper()

	testRoot := filepath.Join("etc", "kuranas")
	t.Cleanup(func() {
		_ = os.RemoveAll(testRoot)
	})

	iconDir := filepath.Join(testRoot, "icons")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		t.Fatalf("failed to create icon dir: %v", err)
	}

	writeIcon := func(name string) {
		t.Helper()
		path := filepath.Join(iconDir, name+".png")
		if _, err := os.Stat(path); err == nil {
			return
		}

		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("failed to create icon file %s: %v", path, err)
		}
		defer f.Close()

		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})
		img.Set(1, 0, color.RGBA{G: 255, A: 255})
		img.Set(0, 1, color.RGBA{B: 255, A: 255})
		img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("failed to encode icon %s: %v", path, err)
		}
	}

	for _, name := range []string{"folder", "unknown", "mp4", "mp3", "pdf"} {
		writeIcon(name)
	}
}

type filesRepoMock struct {
	db *database.DbContext

	createFileFn               func(transaction *sql.Tx, file FileModel) (FileModel, error)
	getFilesFn                 func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	updateFileFn               func(transaction *sql.Tx, file FileModel) (bool, error)
	getDirectoryContentCountFn func(fileId int, parentPath string) (int, error)
	getCountByTypeFn           func(fileType FileType) (int, error)
	getTotalSpaceUsedFn        func() (int, error)
	getReportSizeByFormatFn    func() ([]SizeReportModel, error)
	getTopFilesBySizeFn        func(limit int) ([]FileModel, error)
	getDuplicateFilesFn        func(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error)
	getImagesFn                func(page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getMusicFn                 func(page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getVideosFn                func(page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getMusicArtistsFn          func(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error)
	getMusicByArtistFn         func(artist string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getMusicAlbumsFn           func(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error)
	getMusicByAlbumFn          func(album string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getMusicGenresFn           func(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error)
	getMusicByGenreFn          func(genre string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getMusicFoldersFn          func(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error)
}

func (m *filesRepoMock) GetDbContext() *database.DbContext { return m.db }
func (m *filesRepoMock) CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error) {
	if m.createFileFn != nil {
		return m.createFileFn(transaction, file)
	}
	return file, nil
}
func (m *filesRepoMock) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getFilesFn != nil {
		return m.getFilesFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) UpdateFile(transaction *sql.Tx, file FileModel) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(transaction, file)
	}
	return true, nil
}
func (m *filesRepoMock) GetDirectoryContentCount(fileId int, parentPath string) (int, error) {
	if m.getDirectoryContentCountFn != nil {
		return m.getDirectoryContentCountFn(fileId, parentPath)
	}
	return 0, nil
}
func (m *filesRepoMock) GetCountByType(fileType FileType) (int, error) {
	if m.getCountByTypeFn != nil {
		return m.getCountByTypeFn(fileType)
	}
	return 0, nil
}
func (m *filesRepoMock) GetTotalSpaceUsed() (int, error) {
	if m.getTotalSpaceUsedFn != nil {
		return m.getTotalSpaceUsedFn()
	}
	return 0, nil
}
func (m *filesRepoMock) GetReportSizeByFormat() ([]SizeReportModel, error) {
	if m.getReportSizeByFormatFn != nil {
		return m.getReportSizeByFormatFn()
	}
	return nil, nil
}
func (m *filesRepoMock) GetTopFilesBySize(limit int) ([]FileModel, error) {
	if m.getTopFilesBySizeFn != nil {
		return m.getTopFilesBySizeFn(limit)
	}
	return nil, nil
}
func (m *filesRepoMock) GetDuplicateFiles(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {
	if m.getDuplicateFilesFn != nil {
		return m.getDuplicateFilesFn(page, pageSize)
	}
	return utils.PaginationResponse[DuplicateFilesModel]{Items: []DuplicateFilesModel{}}, nil
}
func (m *filesRepoMock) GetImages(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getImagesFn != nil {
		return m.getImagesFn(page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetMusic(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getMusicFn != nil {
		return m.getMusicFn(page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetVideos(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getVideosFn != nil {
		return m.getVideosFn(page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
	if m.getMusicArtistsFn != nil {
		return m.getMusicArtistsFn(page, pageSize)
	}
	return utils.PaginationResponse[MusicArtistDto]{Items: []MusicArtistDto{}}, nil
}
func (m *filesRepoMock) GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getMusicByArtistFn != nil {
		return m.getMusicByArtistFn(artist, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
	if m.getMusicAlbumsFn != nil {
		return m.getMusicAlbumsFn(page, pageSize)
	}
	return utils.PaginationResponse[MusicAlbumDto]{Items: []MusicAlbumDto{}}, nil
}
func (m *filesRepoMock) GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getMusicByAlbumFn != nil {
		return m.getMusicByAlbumFn(album, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
	if m.getMusicGenresFn != nil {
		return m.getMusicGenresFn(page, pageSize)
	}
	return utils.PaginationResponse[MusicGenreDto]{Items: []MusicGenreDto{}}, nil
}
func (m *filesRepoMock) GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getMusicByGenreFn != nil {
		return m.getMusicByGenreFn(genre, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
	if m.getMusicFoldersFn != nil {
		return m.getMusicFoldersFn(page, pageSize)
	}
	return utils.PaginationResponse[MusicFolderDto]{Items: []MusicFolderDto{}}, nil
}

type metadataRepoMock struct {
	upsertImageFn func(transaction *sql.Tx, metadata ImageMetadataModel) (ImageMetadataModel, error)
	upsertAudioFn func(transaction *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error)
	upsertVideoFn func(transaction *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error)
}

func (m *metadataRepoMock) GetImageMetadataByID(id int) (ImageMetadataModel, error) {
	return ImageMetadataModel{}, nil
}
func (m *metadataRepoMock) UpsertImageMetadata(transaction *sql.Tx, metadata ImageMetadataModel) (ImageMetadataModel, error) {
	if m.upsertImageFn != nil {
		return m.upsertImageFn(transaction, metadata)
	}
	return metadata, nil
}
func (m *metadataRepoMock) DeleteImageMetadata(id int) error { return nil }
func (m *metadataRepoMock) GetAudioMetadataByID(id int) (AudioMetadataModel, error) {
	return AudioMetadataModel{}, nil
}
func (m *metadataRepoMock) UpsertAudioMetadata(transaction *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error) {
	if m.upsertAudioFn != nil {
		return m.upsertAudioFn(transaction, metadata)
	}
	return metadata, nil
}
func (m *metadataRepoMock) DeleteAudioMetadata(id int) error { return nil }
func (m *metadataRepoMock) GetVideoMetadataByID(id int) (VideoMetadataModel, error) {
	return VideoMetadataModel{}, nil
}
func (m *metadataRepoMock) UpsertVideoMetadata(transaction *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error) {
	if m.upsertVideoFn != nil {
		return m.upsertVideoFn(transaction, metadata)
	}
	return metadata, nil
}
func (m *metadataRepoMock) DeleteVideoMetadata(id int) error { return nil }

func newFilesServiceForTest(t *testing.T, repo *filesRepoMock, metadata *metadataRepoMock) *Service {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo.db = database.NewDbContext(db)
	return &Service{
		Repository:         repo,
		MetadataRepository: metadata,
		Tasks:              make(chan utils.Task, 4),
	}
}

func sampleModel(id int, name string, typ FileType) FileModel {
	return FileModel{
		ID:         id,
		Name:       name,
		Path:       "/tmp/" + name,
		ParentPath: "/tmp",
		Type:       typ,
		Format:     ".txt",
		Size:       10,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
}

func TestFileService_GetAndFind(t *testing.T) {
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			if filter.Name.HasValue && filter.Name.Value == "none" {
				return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
			}
			if filter.Name.HasValue && filter.Name.Value == "multi" {
				return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(1, "a", File), sampleModel(2, "b", File)}}, nil
			}
			if filter.ID.HasValue && filter.ID.Value == 123 {
				return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(123, "one", File)}}, nil
			}
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(1, "one", File)}}, nil
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	if _, err := s.GetFileByNameAndPath("none", "/tmp/none"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if _, err := s.GetFileByNameAndPath("multi", "/tmp/multi"); err == nil {
		t.Fatalf("expected multi-file error")
	}
	file, err := s.GetFileByNameAndPath("one", "/tmp/one")
	if err != nil || file.ID == 0 {
		t.Fatalf("expected one file, err=%v", err)
	}
	idFile, err := s.GetFileById(123)
	if err != nil || idFile.ID != 123 {
		t.Fatalf("expected id 123, got %+v err=%v", idFile, err)
	}
}

func TestFileService_GetFilesAndDirectoryCount(t *testing.T) {
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items: []FileModel{
					sampleModel(1, "dir", Directory),
					sampleModel(2, "file", File),
				},
				Pagination: utils.Pagination{Page: 1, PageSize: 10},
			}, nil
		},
		getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
			if fileId == 1 {
				return 4, nil
			}
			return 0, errors.New("not directory")
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	result, err := s.GetFiles(FileFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected two items")
	}
	if result.Items[0].DirectoryContentCount != 4 {
		t.Fatalf("expected directory content count 4")
	}
}

func TestFileService_CreateUpdateAndMetadata(t *testing.T) {
	repo := &filesRepoMock{
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			file.ID = 77
			return file, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return true, nil
		},
	}
	metadata := &metadataRepoMock{
		upsertImageFn: func(transaction *sql.Tx, m ImageMetadataModel) (ImageMetadataModel, error) {
			m.Path = "img"
			return m, nil
		},
		upsertAudioFn: func(transaction *sql.Tx, m AudioMetadataModel) (AudioMetadataModel, error) {
			m.Path = "aud"
			return m, nil
		},
		upsertVideoFn: func(transaction *sql.Tx, m VideoMetadataModel) (VideoMetadataModel, error) {
			m.Path = "vid"
			return m, nil
		},
	}
	s := newFilesServiceForTest(t, repo, metadata)

	file, err := s.CreateFile(FileDto{
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Type:       File,
	})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if file.ID != 77 {
		t.Fatalf("expected id 77")
	}

	ok, err := s.UpdateFile(FileDto{
		ID:         77,
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Type:       File,
		Metadata:   ImageMetadataModel{},
	})
	if err != nil || !ok {
		t.Fatalf("expected update success, ok=%v err=%v", ok, err)
	}

	if _, err := s.UpsertMetadata(nil, FileDto{ID: 1, Metadata: AudioMetadataModel{}}); err != nil {
		t.Fatalf("expected audio upsert success: %v", err)
	}
	if _, err := s.UpsertMetadata(nil, FileDto{ID: 1, Metadata: VideoMetadataModel{}}); err != nil {
		t.Fatalf("expected video upsert success: %v", err)
	}
	if _, err := s.UpsertMetadata(nil, FileDto{ID: 1, Metadata: "unknown"}); err != nil {
		t.Fatalf("expected unknown metadata type to be ignored: %v", err)
	}
}

func TestFileService_ScanAndExistsAndBlob(t *testing.T) {
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			if filter.ID.HasValue && filter.ID.Value == 999 {
				return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
			}
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(10, "blob.txt", File)}}, nil
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	s.ScanFilesTask("x")
	s.ScanDirTask("/tmp")
	if len(s.Tasks) < 3 {
		t.Fatalf("expected tasks to be enqueued")
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "blob.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if !s.CheckFileExistsByPath(tmpFile) {
		t.Fatalf("expected file to exist")
	}
	if s.CheckFileExistsByPath(filepath.Join(tmpDir, "missing")) {
		t.Fatalf("expected missing file to not exist")
	}

	// Rebind repo return path to existing file for blob retrieval.
	repo.getFilesFn = func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
		return utils.PaginationResponse[FileModel]{
			Items: []FileModel{
				{
					ID:         10,
					Name:       "blob.txt",
					Path:       tmpFile,
					ParentPath: tmpDir,
					Type:       File,
					Format:     ".txt",
					UpdatedAt:  time.Now(),
					CreatedAt:  time.Now(),
				},
			},
		}, nil
	}

	blob, err := s.GetFileBlobById(10)
	if err != nil || len(blob.Blob) == 0 {
		t.Fatalf("expected blob bytes, err=%v", err)
	}
	if !s.CheckFileExists(10) {
		t.Fatalf("expected CheckFileExists true for existing file")
	}
	repo.getFilesFn = func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
		return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
	}
	if s.CheckFileExists(999) {
		t.Fatalf("expected CheckFileExists false for missing file")
	}
}

func TestFileService_ReportsAndWrappers(t *testing.T) {
	repo := &filesRepoMock{
		getTotalSpaceUsedFn: func() (int, error) { return 1000, nil },
		getCountByTypeFn: func(fileType FileType) (int, error) {
			if fileType == File {
				return 50, nil
			}
			return 20, nil
		},
		getReportSizeByFormatFn: func() ([]SizeReportModel, error) {
			return []SizeReportModel{
				{Format: ".mp3", Total: 3, Size: 300},
				{Format: ".jpg", Total: 2, Size: 200},
			}, nil
		},
		getTopFilesBySizeFn: func(limit int) ([]FileModel, error) {
			return []FileModel{sampleModel(1, "top", File)}, nil
		},
		getDuplicateFilesFn: func(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {
			return utils.PaginationResponse[DuplicateFilesModel]{
				Items: []DuplicateFilesModel{{Name: "d", Size: 10, Copies: 2, Paths: "/a,/b"}},
				Pagination: utils.Pagination{
					Page: 1, PageSize: 10,
				},
			}, nil
		},
		getImagesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(1, "i", File)}}, nil
		},
		getMusicFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(2, "m", File)}}, nil
		},
		getVideosFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(3, "v", File)}}, nil
		},
		getMusicArtistsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
			return utils.PaginationResponse[MusicArtistDto]{Items: []MusicArtistDto{{Artist: "a"}}}, nil
		},
		getMusicByArtistFn: func(artist string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(4, "ma", File)}}, nil
		},
		getMusicAlbumsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
			return utils.PaginationResponse[MusicAlbumDto]{Items: []MusicAlbumDto{{Album: "al"}}}, nil
		},
		getMusicByAlbumFn: func(album string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(5, "mb", File)}}, nil
		},
		getMusicGenresFn: func(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
			return utils.PaginationResponse[MusicGenreDto]{Items: []MusicGenreDto{{Genre: "g"}}}, nil
		},
		getMusicByGenreFn: func(genre string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(6, "mg", File)}}, nil
		},
		getMusicFoldersFn: func(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
			return utils.PaginationResponse[MusicFolderDto]{Items: []MusicFolderDto{{Folder: "f"}}}, nil
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	if total, _ := s.GetTotalSpaceUsed(); total != 1000 {
		t.Fatalf("expected total space 1000")
	}
	if filesCount, _ := s.GetTotalFiles(); filesCount != 50 {
		t.Fatalf("expected total files 50")
	}
	if dirsCount, _ := s.GetTotalDirectory(); dirsCount != 20 {
		t.Fatalf("expected total dirs 20")
	}

	report, err := s.GetReportSizeByFormat()
	if err != nil || len(report) == 0 {
		t.Fatalf("expected report entries, err=%v", err)
	}
	top, err := s.GetTopFilesBySize(1)
	if err != nil || len(top) != 1 {
		t.Fatalf("expected one top file, err=%v", err)
	}
	dups, err := s.GetDuplicateFiles(1, 10)
	if err != nil || dups.TotalFiles != 2 {
		t.Fatalf("expected duplicates report, err=%v", err)
	}

	if _, err := s.GetImages(1, 10); err != nil {
		t.Fatalf("expected images success, err=%v", err)
	}
	if _, err := s.GetMusic(1, 10); err != nil {
		t.Fatalf("expected music success, err=%v", err)
	}
	if _, err := s.GetVideos(1, 10); err != nil {
		t.Fatalf("expected videos success, err=%v", err)
	}
	if _, err := s.GetMusicArtists(1, 10); err != nil {
		t.Fatalf("expected artists success, err=%v", err)
	}
	if _, err := s.GetMusicByArtist("a", 1, 10); err != nil {
		t.Fatalf("expected music by artist success, err=%v", err)
	}
	if _, err := s.GetMusicAlbums(1, 10); err != nil {
		t.Fatalf("expected albums success, err=%v", err)
	}
	if _, err := s.GetMusicByAlbum("x", 1, 10); err != nil {
		t.Fatalf("expected by album success, err=%v", err)
	}
	if _, err := s.GetMusicGenres(1, 10); err != nil {
		t.Fatalf("expected genres success, err=%v", err)
	}
	if _, err := s.GetMusicByGenre("g", 1, 10); err != nil {
		t.Fatalf("expected by genre success, err=%v", err)
	}
	if _, err := s.GetMusicFolders(1, 10); err != nil {
		t.Fatalf("expected folders success, err=%v", err)
	}
}

func TestFileService_DeleteAndChecksumBranches(t *testing.T) {
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items: []FileModel{sampleModel(filter.ID.Value, "x", File)},
			}, nil
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	alreadyDeleted := FileDto{
		ID:        1,
		Name:      "x",
		Path:      "/tmp/x",
		DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now()},
	}
	if err := s.DeleteFile(alreadyDeleted, false); err == nil {
		t.Fatalf("expected already deleted error")
	}

	recent := FileDto{
		ID:              2,
		Name:            "y",
		Path:            "/tmp/y",
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now()},
	}
	if err := s.DeleteFile(recent, true); err == nil {
		t.Fatalf("expected recently accessed deletion error")
	}
	if err := s.DeleteFile(recent, false); err != nil {
		t.Fatalf("expected manual delete to ignore recent-access rule, got %v", err)
	}

	// Force default branch in UpdateCheckSum.
	repo.getFilesFn = func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
		m := sampleModel(filter.ID.Value, "unknown", FileType(99))
		return utils.PaginationResponse[FileModel]{Items: []FileModel{m}}, nil
	}
	if err := s.UpdateCheckSum(9); err == nil {
		t.Fatalf("expected unknown file type error")
	}
}

func TestFileService_ChecksumThumbnailAndDeleteSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(filePath, []byte("checksum"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	updateCalls := 0
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			if filter.ID.HasValue && filter.ID.Value == 1 {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{
						{
							ID:         1,
							Name:       "data.txt",
							Path:       filePath,
							ParentPath: tmpDir,
							Type:       File,
							Format:     ".txt",
							UpdatedAt:  time.Now(),
							CreatedAt:  time.Now(),
						},
					},
				}, nil
			}
			if filter.ID.HasValue && filter.ID.Value == 2 {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{
						{
							ID:         2,
							Name:       "dir",
							Path:       tmpDir,
							ParentPath: filepath.Dir(tmpDir),
							Type:       Directory,
							UpdatedAt:  time.Now(),
							CreatedAt:  time.Now(),
						},
					},
				}, nil
			}
			if filter.ParentPath.HasValue {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{
						{
							ID:         3,
							Name:       "child1",
							Path:       filepath.Join(tmpDir, "c1"),
							ParentPath: tmpDir,
							Type:       File,
							CheckSum:   "abcd",
							UpdatedAt:  time.Now(),
							CreatedAt:  time.Now(),
						},
					},
					Pagination: utils.Pagination{Page: 1, PageSize: 1000, HasNext: false},
				}, nil
			}
			return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			updateCalls++
			return true, nil
		},
	}
	s := newFilesServiceForTest(t, repo, &metadataRepoMock{})

	if err := s.UpdateCheckSum(1); err != nil {
		t.Fatalf("expected file checksum update success, got %v", err)
	}
	if err := s.UpdateCheckSum(2); err != nil {
		t.Fatalf("expected dir checksum update success, got %v", err)
	}
	if updateCalls < 2 {
		t.Fatalf("expected at least two update calls, got %d", updateCalls)
	}

	oldAccess := time.Now().Add(-48 * time.Hour)
	err := s.DeleteFile(FileDto{
		ID:              4,
		Name:            "old.txt",
		Path:            filePath,
		Type:            File,
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: oldAccess},
	}, true)
	if err != nil {
		t.Fatalf("expected DeleteFile success for old access file, got %v", err)
	}

	thumbData, err := s.GetFileThumbnail(FileDto{
		ID:              5,
		Name:            "missing.txt",
		Path:            filepath.Join(tmpDir, "missing.txt"),
		ParentPath:      tmpDir,
		Type:            File,
		Format:          ".txt",
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: oldAccess},
	}, 100, 100)
	if err == nil || thumbData != nil {
		t.Fatalf("expected GetFileThumbnail to fail on missing file with nil data")
	}

	if _, err := s.GetVideoThumbnail(FileDto{
		ID:   6,
		Path: filepath.Join(tmpDir, "missing.mp4"),
		Type: File,
	}, 320, 180); err == nil {
		t.Fatalf("expected missing video thumbnail error")
	}

	if _, err := s.GetVideoPreviewGif(FileDto{
		ID:   7,
		Path: filepath.Join(tmpDir, "missing.mp4"),
		Type: File,
	}, 320, 180); err == nil {
		t.Fatalf("expected missing video preview error")
	}
}

func TestFileService_ThumbnailAndVideoFallbacks(t *testing.T) {
	ensureTestIcons(t)

	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "not-image.txt")
	if err := os.WriteFile(existingFile, []byte("plain text"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fakeVideo := filepath.Join(tmpDir, "video.mp4")
	if err := os.WriteFile(fakeVideo, []byte("not-a-real-video"), 0644); err != nil {
		t.Fatalf("failed to create fake video file: %v", err)
	}

	s := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{})

	dirThumb, err := s.GetFileThumbnail(FileDto{
		ID:   101,
		Path: tmpDir,
		Type: Directory,
	}, -10, 120)
	if err != nil {
		t.Fatalf("expected directory thumbnail success, got %v", err)
	}
	if len(dirThumb) == 0 {
		t.Fatalf("expected non-empty directory thumbnail")
	}

	fileThumb, err := s.GetFileThumbnail(FileDto{
		ID:     102,
		Path:   existingFile,
		Type:   File,
		Format: ".txt",
	}, 4096, 120)
	if err != nil {
		t.Fatalf("expected file fallback thumbnail success, got %v", err)
	}
	if len(fileThumb) == 0 {
		t.Fatalf("expected non-empty file thumbnail")
	}

	videoThumb, err := s.GetVideoThumbnail(FileDto{
		ID:   103,
		Path: fakeVideo,
		Type: File,
	}, -1, -1)
	if err != nil {
		t.Fatalf("expected video thumbnail fallback success, got %v", err)
	}
	if len(videoThumb) == 0 {
		t.Fatalf("expected non-empty video thumbnail fallback")
	}

	previewGif, err := s.GetVideoPreviewGif(FileDto{
		ID:   104,
		Path: fakeVideo,
		Type: File,
	}, -1, -1)
	if err != nil {
		t.Fatalf("expected video preview fallback success, got %v", err)
	}
	if len(previewGif) == 0 {
		t.Fatalf("expected non-empty video preview fallback")
	}
}

func TestFileService_GetFileThumbnailCacheHit(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{})
	cacheDir := config.GetBuildConfig("ThumbnailPath")
	cacheFile := filepath.Join(cacheDir, "42_320.png")
	cached := []byte("cached-png")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(cacheFile, cached, 0644); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cacheFile) })

	data, err := s.GetFileThumbnail(FileDto{
		ID:   42,
		Path: "/path/not/used/on-cache-hit",
		Type: File,
	}, 0, 100)
	if err != nil {
		t.Fatalf("expected cache hit without error, got %v", err)
	}
	if string(data) != string(cached) {
		t.Fatalf("expected cached data, got %q", string(data))
	}
}

func TestFileService_GetFileThumbnailMissingFileDeleteFailure(t *testing.T) {
	tmpDir := t.TempDir()
	s := newFilesServiceForTest(t, &filesRepoMock{
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return false, errors.New("update failure")
		},
	}, &metadataRepoMock{})

	_, err := s.GetFileThumbnail(FileDto{
		ID:     99,
		Path:   filepath.Join(tmpDir, "missing.txt"),
		Type:   File,
		Format: ".txt",
	}, 100, 100)
	if err == nil {
		t.Fatalf("expected error for missing file with delete failure")
	}
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase wrapping, got %v", err)
	}
}

func TestFileService_GetVideoThumbAndPreviewCacheHit(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{})
	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create video cache dir: %v", err)
	}

	thumbPath := filepath.Join(cacheDir, "501_320x180.png")
	thumbBytes := []byte("cached-video-thumb")
	if err := os.WriteFile(thumbPath, thumbBytes, 0644); err != nil {
		t.Fatalf("failed to write thumb cache: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(thumbPath) })

	thumb, err := s.GetVideoThumbnail(FileDto{ID: 501, Path: "/missing.mp4", Type: File}, 320, 180)
	if err != nil {
		t.Fatalf("expected video thumbnail cache hit, got %v", err)
	}
	if string(thumb) != string(thumbBytes) {
		t.Fatalf("expected cached thumbnail bytes")
	}

	previewPath := filepath.Join(cacheDir, "502_320x180_preview.gif")
	previewBytes := []byte("cached-video-preview")
	if err := os.WriteFile(previewPath, previewBytes, 0644); err != nil {
		t.Fatalf("failed to write preview cache: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(previewPath) })

	preview, err := s.GetVideoPreviewGif(FileDto{ID: 502, Path: "/missing.mp4", Type: File}, 320, 180)
	if err != nil {
		t.Fatalf("expected video preview cache hit, got %v", err)
	}
	if string(preview) != string(previewBytes) {
		t.Fatalf("expected cached preview bytes")
	}
}

func TestFileService_ErrorBranches(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("get files failed")
		},
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			return FileModel{}, errors.New("create failed")
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return false, errors.New("update failed")
		},
		getImagesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("images failed")
		},
		getMusicFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("music failed")
		},
		getVideosFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("videos failed")
		},
		getMusicByArtistFn: func(artist string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("artist failed")
		},
		getMusicByAlbumFn: func(album string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("album failed")
		},
		getMusicByGenreFn: func(genre string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("genre failed")
		},
	}, &metadataRepoMock{
		upsertImageFn: func(transaction *sql.Tx, metadata ImageMetadataModel) (ImageMetadataModel, error) {
			return ImageMetadataModel{}, errors.New("upsert image failed")
		},
	})

	if _, err := s.GetFiles(FileFilter{}, 1, 10); err == nil {
		t.Fatalf("expected GetFiles error")
	}
	if _, err := s.CreateFile(FileDto{Name: "x", Path: "/tmp/x", ParentPath: "/tmp", Type: File}); err == nil {
		t.Fatalf("expected CreateFile error")
	}
	if _, err := s.UpdateFile(FileDto{ID: 1, Name: "x", Path: "/tmp/x", ParentPath: "/tmp", Type: File}); err == nil {
		t.Fatalf("expected UpdateFile error")
	}
	if _, err := s.UpdateFile(FileDto{ID: 1, Name: "x", Path: "/tmp/x", ParentPath: "/tmp", Type: File, Metadata: ImageMetadataModel{}}); err == nil {
		t.Fatalf("expected UpdateFile metadata error")
	}

	if _, err := s.GetImages(1, 10); err == nil {
		t.Fatalf("expected GetImages error")
	}
	if _, err := s.GetMusic(1, 10); err == nil {
		t.Fatalf("expected GetMusic error")
	}
	if _, err := s.GetVideos(1, 10); err == nil {
		t.Fatalf("expected GetVideos error")
	}
	if _, err := s.GetMusicByArtist("a", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByArtist error")
	}
	if _, err := s.GetMusicByAlbum("a", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByAlbum error")
	}
	if _, err := s.GetMusicByGenre("a", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByGenre error")
	}
}

func TestFileService_GetFileBlobByIdReadError(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items: []FileModel{sampleModel(999, "missing.bin", File)},
			}, nil
		},
	}, &metadataRepoMock{})

	if _, err := s.GetFileBlobById(999); err == nil {
		t.Fatalf("expected GetFileBlobById read error")
	}
}

func TestFileService_AdditionalErrorAndEdgeBranches(t *testing.T) {
	t.Run("GetFiles with directory count error falls back to zero", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{sampleModel(1, "dir", Directory)},
				}, nil
			},
			getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
				return 0, errors.New("count failed")
			},
		}, &metadataRepoMock{})

		out, err := s.GetFiles(FileFilter{}, 1, 10)
		if err != nil {
			t.Fatalf("expected GetFiles success, got %v", err)
		}
		if out.Items[0].DirectoryContentCount != 0 {
			t.Fatalf("expected directory count fallback to 0")
		}
	})

	t.Run("GetFileById no rows and multiple rows", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				if filter.ID.Value == 10 {
					return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
				}
				return utils.PaginationResponse[FileModel]{Items: []FileModel{
					sampleModel(1, "a", File),
					sampleModel(2, "b", File),
				}}, nil
			},
		}, &metadataRepoMock{})

		if _, err := s.GetFileById(10); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		if _, err := s.GetFileById(11); err == nil {
			t.Fatalf("expected multiple rows error")
		}
	})

	t.Run("UpdateCheckSum returns error when update does not affect rows", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "x.txt")
		if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{{
						ID:         filter.ID.Value,
						Name:       "x.txt",
						Path:       filePath,
						ParentPath: tmpDir,
						Type:       File,
						Format:     ".txt",
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					}},
				}, nil
			},
			updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
				return false, nil
			},
		}, &metadataRepoMock{})

		if err := s.UpdateCheckSum(1); err == nil {
			t.Fatalf("expected UpdateCheckSum error when update returns false")
		}
	})

	t.Run("UpdateCheckSum propagates get-file error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{}, errors.New("repository down")
			},
		}, &metadataRepoMock{})

		if err := s.UpdateCheckSum(1); err == nil {
			t.Fatalf("expected UpdateCheckSum to propagate fetch error")
		}
	})

	t.Run("GetReportSizeByFormat propagates repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getReportSizeByFormatFn: func() ([]SizeReportModel, error) {
				return nil, errors.New("report failed")
			},
		}, &metadataRepoMock{})

		if _, err := s.GetReportSizeByFormat(); err == nil {
			t.Fatalf("expected report error")
		}
	})

	t.Run("UpsertMetadata audio and video errors", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{
			upsertAudioFn: func(transaction *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error) {
				return AudioMetadataModel{}, errors.New("audio metadata failed")
			},
			upsertVideoFn: func(transaction *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error) {
				return VideoMetadataModel{}, errors.New("video metadata failed")
			},
		})

		if _, err := s.UpsertMetadata(nil, FileDto{ID: 1, Metadata: AudioMetadataModel{}}); err == nil {
			t.Fatalf("expected audio metadata error")
		}
		if _, err := s.UpsertMetadata(nil, FileDto{ID: 1, Metadata: VideoMetadataModel{}}); err == nil {
			t.Fatalf("expected video metadata error")
		}
	})

	t.Run("CheckFileExists returns false on repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{}, errors.New("query failed")
			},
		}, &metadataRepoMock{})

		if s.CheckFileExists(10) {
			t.Fatalf("expected CheckFileExists false on repository error")
		}
	})

	t.Run("DeleteFile propagates update errors", func(t *testing.T) {
		sErr := newFilesServiceForTest(t, &filesRepoMock{
			updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
				return false, errors.New("update failed")
			},
		}, &metadataRepoMock{})

		err := sErr.DeleteFile(FileDto{
			ID:              1,
			Name:            "x",
			Path:            "/tmp/x",
			Type:            File,
			LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now().Add(-48 * time.Hour)},
		}, true)
		if err == nil {
			t.Fatalf("expected DeleteFile update error")
		}

		sFalse := newFilesServiceForTest(t, &filesRepoMock{
			updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
				return false, nil
			},
		}, &metadataRepoMock{})

		err = sFalse.DeleteFile(FileDto{
			ID:              2,
			Name:            "y",
			Path:            "/tmp/y",
			Type:            File,
			LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now().Add(-48 * time.Hour)},
		}, true)
		if err == nil {
			t.Fatalf("expected DeleteFile not-found error")
		}
	})
}
