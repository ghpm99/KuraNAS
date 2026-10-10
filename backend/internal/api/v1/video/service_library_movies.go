package video

import (
	"fmt"

	"nas-go/api/pkg/utils"
)

func (s *Service) ListLibraryMovies(request LibraryMoviesRequest) (utils.PaginationResponse[VideoFileDto], error) {
	movieModels, err := s.Repository.ListLibraryMovies(
		request.Sort,
		request.PageSize+1,
		utils.CalculateOffset(request.Page, request.PageSize),
	)
	if err != nil {
		return utils.PaginationResponse[VideoFileDto]{}, fmt.Errorf("ListLibraryMovies: %w", err)
	}

	movieDtos := make([]VideoFileDto, 0, len(movieModels))
	for _, movieModel := range movieModels {
		movieDtos = append(movieDtos, movieModel.ToDto())
	}

	response := utils.PaginationResponse[VideoFileDto]{
		Items:      movieDtos,
		Pagination: utils.Pagination{Page: request.Page, PageSize: request.PageSize},
	}
	response.UpdatePagination()
	return response, nil
}
