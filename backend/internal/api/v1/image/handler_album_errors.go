package image

import (
	"errors"
	"net/http"
)

type albumErrorResponse struct {
	statusCode int
	messageKey string
}

var albumErrorResponses = []struct {
	knownError error
	response   albumErrorResponse
}{
	{ErrAlbumNotFound, albumErrorResponse{http.StatusNotFound, "ERROR_IMAGE_ALBUM_NOT_FOUND"}},
	{ErrAlbumNameTaken, albumErrorResponse{http.StatusConflict, "ERROR_IMAGE_ALBUM_NAME_TAKEN"}},
	{ErrAlbumNameRequired, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_NAME_REQUIRED"}},
	{ErrAlbumNameTooLong, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_NAME_TOO_LONG"}},
	{ErrAlbumInvalidID, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_INVALID_ID"}},
	{ErrAlbumInvalidFileIDs, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_INVALID_FILE_IDS"}},
	{ErrAlbumCoverNotInAlbum, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_COVER_NOT_IN_ALBUM"}},
	{ErrAlbumNothingToUpdate, albumErrorResponse{http.StatusBadRequest, "ERROR_IMAGE_ALBUM_NOTHING_TO_UPDATE"}},
}

func albumErrorResponseFor(err error) albumErrorResponse {
	for _, candidate := range albumErrorResponses {
		if errors.Is(err, candidate.knownError) {
			return candidate.response
		}
	}
	if libraryMessageKey := libraryErrorMessageKey(err); libraryMessageKey != "ERROR_INVALID_REQUEST" {
		return albumErrorResponse{http.StatusBadRequest, libraryMessageKey}
	}
	return albumErrorResponse{http.StatusInternalServerError, "ERROR_INTERNAL"}
}
