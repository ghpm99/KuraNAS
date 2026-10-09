package image

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeAlbumRepository struct {
	albums         []AlbumModel
	getErr         error
	createErr      error
	updateErr      error
	deleteErr      error
	changeErr      error
	changedCount   int
	listLimit      int
	listOffset     int
	createdName    string
	update         AlbumUpdate
	deletedAlbumID int
	changedFileIDs []int
	changedAlbumID int
}

func (f *fakeAlbumRepository) ListAlbums(limit int, offset int) ([]AlbumModel, error) {
	f.listLimit, f.listOffset = limit, offset
	return f.albums, f.getErr
}

func (f *fakeAlbumRepository) GetAlbum(albumID int) (AlbumModel, error) {
	if f.getErr != nil {
		return AlbumModel{}, f.getErr
	}
	coverFileID := 9
	return AlbumModel{ID: albumID, Name: "Viagem", CoverFileID: &coverFileID, ItemCount: 3}, nil
}

func (f *fakeAlbumRepository) CreateAlbum(name string) (int, error) {
	f.createdName = name
	return 4, f.createErr
}

func (f *fakeAlbumRepository) UpdateAlbum(update AlbumUpdate) error {
	f.update = update
	return f.updateErr
}

func (f *fakeAlbumRepository) DeleteAlbum(albumID int) error {
	f.deletedAlbumID = albumID
	return f.deleteErr
}

func (f *fakeAlbumRepository) AddAlbumItems(albumID int, fileIDs []int) (int, error) {
	f.changedAlbumID, f.changedFileIDs = albumID, fileIDs
	return f.changedCount, f.changeErr
}

func (f *fakeAlbumRepository) RemoveAlbumItems(albumID int, fileIDs []int) (int, error) {
	return f.AddAlbumItems(albumID, fileIDs)
}

func TestCreateAlbumTrimsAndValidatesName(t *testing.T) {
	repository := &fakeAlbumRepository{}
	service := NewAlbumService(repository, &fakeLibraryService{})

	album, err := service.CreateAlbum("  Viagem  ")
	if err != nil || repository.createdName != "Viagem" || album.ID != 4 || album.ItemCount != 3 {
		t.Fatalf("album = %+v name %q err %v", album, repository.createdName, err)
	}

	if _, err := service.CreateAlbum("   "); !errors.Is(err, ErrAlbumNameRequired) {
		t.Fatalf("blank name error = %v", err)
	}
	if _, err := service.CreateAlbum(strings.Repeat("á", maxAlbumNameLength+1)); !errors.Is(err, ErrAlbumNameTooLong) {
		t.Fatalf("long name error = %v", err)
	}
	if _, err := service.CreateAlbum(strings.Repeat("á", maxAlbumNameLength)); err != nil {
		t.Fatalf("name at the limit must pass: %v", err)
	}
}

func TestCreateAlbumPropagatesDuplicateName(t *testing.T) {
	service := NewAlbumService(&fakeAlbumRepository{createErr: ErrAlbumNameTaken}, &fakeLibraryService{})
	if _, err := service.CreateAlbum("Viagem"); !errors.Is(err, ErrAlbumNameTaken) {
		t.Fatalf("error = %v", err)
	}
}

func TestListAlbumsComputesPaginationFromExtraRow(t *testing.T) {
	repository := &fakeAlbumRepository{albums: []AlbumModel{{ID: 1}, {ID: 2}, {ID: 3}}}
	service := NewAlbumService(repository, &fakeLibraryService{})

	page, err := service.ListAlbums(AlbumListRequest{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if repository.listLimit != 3 || repository.listOffset != 2 {
		t.Fatalf("limit/offset = %d/%d", repository.listLimit, repository.listOffset)
	}
	if len(page.Items) != 2 || !page.Pagination.HasNext || !page.Pagination.HasPrev || page.Pagination.Page != 2 {
		t.Fatalf("page = %+v", page)
	}
}

func TestListAlbumsPropagatesRepositoryError(t *testing.T) {
	service := NewAlbumService(&fakeAlbumRepository{getErr: errors.New("db down")}, &fakeLibraryService{})
	if _, err := service.ListAlbums(AlbumListRequest{Page: 1, PageSize: 10}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateAlbumValidation(t *testing.T) {
	service := NewAlbumService(&fakeAlbumRepository{}, &fakeLibraryService{})
	blankName := " "
	invalidCover := 0

	if _, err := service.UpdateAlbum(AlbumUpdate{AlbumID: 1}); !errors.Is(err, ErrAlbumNothingToUpdate) {
		t.Fatalf("empty update error = %v", err)
	}
	if _, err := service.UpdateAlbum(AlbumUpdate{AlbumID: 1, Name: &blankName}); !errors.Is(err, ErrAlbumNameRequired) {
		t.Fatalf("blank name error = %v", err)
	}
	if _, err := service.UpdateAlbum(AlbumUpdate{AlbumID: 1, CoverFileID: &invalidCover}); !errors.Is(err, ErrAlbumInvalidFileIDs) {
		t.Fatalf("invalid cover error = %v", err)
	}
}

func TestUpdateAlbumTrimsNameAndReturnsAlbum(t *testing.T) {
	repository := &fakeAlbumRepository{}
	service := NewAlbumService(repository, &fakeLibraryService{})
	rawName := " Férias "
	coverFileID := 7

	album, err := service.UpdateAlbum(AlbumUpdate{AlbumID: 5, Name: &rawName, CoverFileID: &coverFileID})
	if err != nil || album.ID != 5 {
		t.Fatalf("album = %+v err %v", album, err)
	}
	if *repository.update.Name != "Férias" || *repository.update.CoverFileID != 7 {
		t.Fatalf("update = %+v", repository.update)
	}
}

func TestUpdateAlbumMapsUnknownAlbumAndRepositoryErrors(t *testing.T) {
	rawName := "Novo"
	if _, err := NewAlbumService(&fakeAlbumRepository{getErr: ErrAlbumNotFound}, &fakeLibraryService{}).UpdateAlbum(AlbumUpdate{AlbumID: 1, Name: &rawName}); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("unknown album error = %v", err)
	}
	if _, err := NewAlbumService(&fakeAlbumRepository{updateErr: ErrAlbumNameTaken}, &fakeLibraryService{}).UpdateAlbum(AlbumUpdate{AlbumID: 1, Name: &rawName}); !errors.Is(err, ErrAlbumNameTaken) {
		t.Fatalf("taken name error = %v", err)
	}
}

func TestDeleteAlbumDelegatesToRepository(t *testing.T) {
	repository := &fakeAlbumRepository{}
	service := NewAlbumService(repository, &fakeLibraryService{})
	if err := service.DeleteAlbum(8); err != nil || repository.deletedAlbumID != 8 {
		t.Fatalf("delete err %v id %d", err, repository.deletedAlbumID)
	}

	failingService := NewAlbumService(&fakeAlbumRepository{deleteErr: ErrAlbumNotFound}, &fakeLibraryService{})
	if err := failingService.DeleteAlbum(8); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("error = %v", err)
	}
}

func TestChangeAlbumItemsDeduplicatesAndReportsCounts(t *testing.T) {
	repository := &fakeAlbumRepository{changedCount: 2}
	service := NewAlbumService(repository, &fakeLibraryService{})

	added, err := service.AddAlbumItems(3, []int{5, 6, 5, 7})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if added.Requested != 3 || added.Changed != 2 || !reflect.DeepEqual(repository.changedFileIDs, []int{5, 6, 7}) || repository.changedAlbumID != 3 {
		t.Fatalf("added = %+v ids %v", added, repository.changedFileIDs)
	}

	removed, err := service.RemoveAlbumItems(3, []int{5})
	if err != nil || removed.Requested != 1 {
		t.Fatalf("removed = %+v err %v", removed, err)
	}
}

func TestChangeAlbumItemsRejectsInvalidSelections(t *testing.T) {
	service := NewAlbumService(&fakeAlbumRepository{}, &fakeLibraryService{})
	tooManyFileIDs := make([]int, maxAlbumItemsPerCall+1)
	for position := range tooManyFileIDs {
		tooManyFileIDs[position] = position + 1
	}
	atLimitFileIDs := tooManyFileIDs[:maxAlbumItemsPerCall]

	for name, fileIDs := range map[string][]int{"empty": nil, "zero": {0}, "negative": {-1, 2}, "too many": tooManyFileIDs} {
		if _, err := service.AddAlbumItems(1, fileIDs); !errors.Is(err, ErrAlbumInvalidFileIDs) {
			t.Fatalf("%s: error = %v", name, err)
		}
	}
	if _, err := service.AddAlbumItems(1, atLimitFileIDs); err != nil {
		t.Fatalf("selection at the limit must pass: %v", err)
	}
}

func TestChangeAlbumItemsPropagatesErrors(t *testing.T) {
	if _, err := NewAlbumService(&fakeAlbumRepository{getErr: ErrAlbumNotFound}, &fakeLibraryService{}).AddAlbumItems(1, []int{1}); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("unknown album error = %v", err)
	}
	if _, err := NewAlbumService(&fakeAlbumRepository{changeErr: errors.New("db down")}, &fakeLibraryService{}).RemoveAlbumItems(1, []int{1}); err == nil {
		t.Fatal("expected repository error")
	}
}

func TestListAlbumItemsRestrictsLibraryListingToTheAlbum(t *testing.T) {
	library := &fakeLibraryService{page: LibraryPageDto{Items: []LibraryItemDto{{FileID: 1}}}}
	service := NewAlbumService(&fakeAlbumRepository{}, library)

	page, err := service.ListAlbumItems(AlbumItemsRequest{AlbumID: 6, Listing: LibraryListRequest{PageSize: 10}})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("page = %+v err %v", page, err)
	}
	if library.listRequest.Filter.AlbumID != 6 || library.listRequest.PageSize != 10 {
		t.Fatalf("library request = %+v", library.listRequest)
	}
}

func TestListAlbumItemsErrors(t *testing.T) {
	if _, err := NewAlbumService(&fakeAlbumRepository{getErr: ErrAlbumNotFound}, &fakeLibraryService{}).ListAlbumItems(AlbumItemsRequest{AlbumID: 1}); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("unknown album error = %v", err)
	}
	if _, err := NewAlbumService(&fakeAlbumRepository{}, &fakeLibraryService{err: errors.New("db down")}).ListAlbumItems(AlbumItemsRequest{AlbumID: 1}); err == nil {
		t.Fatal("expected library error")
	}
}

func TestGetAlbumReturnsDtoAndPropagatesNotFound(t *testing.T) {
	album, err := NewAlbumService(&fakeAlbumRepository{}, &fakeLibraryService{}).GetAlbum(3)
	if err != nil || album.ID != 3 || album.ItemCount != 3 {
		t.Fatalf("album = %+v err %v", album, err)
	}
	if _, err := NewAlbumService(&fakeAlbumRepository{getErr: ErrAlbumNotFound}, &fakeLibraryService{}).GetAlbum(3); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("error = %v", err)
	}
}
