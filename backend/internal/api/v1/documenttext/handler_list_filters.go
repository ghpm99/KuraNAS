package documenttext

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const documentDateLayout = "2006-01-02"

var documentFormatGroups = map[string]bool{
	GroupPDF: true, GroupText: true, GroupSpreadsheet: true, GroupPresentation: true, GroupOther: true, FormatImage: true,
}

var documentIndexStates = map[string]bool{
	StateIndexed: true, StatePartial: true, StateNameOnly: true, StateQueued: true, StateFailed: true,
}

var documentSorts = map[string]bool{
	SortModifiedAt: true, SortName: true, SortSize: true, SortCreatedAt: true,
}

func parseDocumentFormat(rawFormat string) (string, bool) {
	if rawFormat == "" || documentFormatGroups[rawFormat] {
		return rawFormat, true
	}
	return "", false
}

func parseDocumentDate(rawDate string) (*time.Time, bool) {
	if rawDate == "" {
		return nil, true
	}
	parsedDate, err := time.ParseInLocation(documentDateLayout, rawDate, time.UTC)
	if err != nil {
		return nil, false
	}
	return &parsedDate, true
}

func parseDocumentFavoritesOnly(rawFavorites string) (bool, bool) {
	switch rawFavorites {
	case "", "false":
		return false, true
	case "true":
		return true, true
	}
	return false, false
}

func parseDocumentIndexState(rawIndexState string) (string, bool) {
	if rawIndexState == "" || documentIndexStates[rawIndexState] {
		return rawIndexState, true
	}
	return "", false
}

func parseDocumentFolderID(rawFolderID string) (int, bool) {
	if rawFolderID == "" {
		return 0, true
	}
	folderID, err := strconv.Atoi(rawFolderID)
	if err != nil || folderID <= 0 {
		return 0, false
	}
	return folderID, true
}

func parseDocumentSort(rawSort string) (string, bool) {
	if rawSort == "" {
		return SortModifiedAt, true
	}
	return rawSort, documentSorts[rawSort]
}

func parseDocumentOrder(rawOrder string, sort string) (string, bool) {
	switch rawOrder {
	case OrderAscending, OrderDescending:
		return rawOrder, true
	case "":
		return defaultOrderOf(sort), true
	}
	return "", false
}

func defaultOrderOf(sort string) string {
	if sort == SortName {
		return OrderAscending
	}
	return OrderDescending
}

func parseDocumentFilter(c *gin.Context) (DocumentFilter, int, string) {
	format, isFormatValid := parseDocumentFormat(c.Query("format"))
	if !isFormatValid {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_FORMAT"
	}
	folderID, isFolderValid := parseDocumentFolderID(c.Query("folder_id"))
	if !isFolderValid {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_FOLDER"
	}
	modifiedFrom, isFromValid := parseDocumentDate(c.Query("modified_from"))
	modifiedTo, isToValid := parseDocumentDate(c.Query("modified_to"))
	if !isFromValid || !isToValid {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_DATE"
	}
	if modifiedFrom != nil && modifiedTo != nil && modifiedFrom.After(*modifiedTo) {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_RANGE"
	}
	favoritesOnly, isFavoritesValid := parseDocumentFavoritesOnly(c.Query("favorites_only"))
	if !isFavoritesValid {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_FAVORITES"
	}
	indexState, isIndexStateValid := parseDocumentIndexState(c.Query("index_state"))
	if !isIndexStateValid {
		return DocumentFilter{}, 0, "ERROR_DOCUMENT_INVALID_INDEX_STATE"
	}
	return DocumentFilter{
		Format:        format,
		ModifiedFrom:  modifiedFrom,
		ModifiedTo:    modifiedTo,
		FavoritesOnly: favoritesOnly,
		IndexState:    indexState,
	}, folderID, ""
}
