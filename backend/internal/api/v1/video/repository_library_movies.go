package video

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

const movieClassification = "movie"

func libraryMoviesQueryFor(sort LibraryMovieSort) string {
	if sort == LibraryMovieSortRecent {
		return queries.GetLibraryMoviesByRecentQuery
	}
	return queries.GetLibraryMoviesByNameQuery
}

func (r *Repository) ListLibraryMovies(sort LibraryMovieSort, limit int, offset int) ([]VideoFileModel, error) {
	movies := []VideoFileModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(libraryMoviesQueryFor(sort), pq.Array(utils.VideoFormats), movieClassification, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var movie VideoFileModel
			if err := rows.Scan(
				&movie.ID,
				&movie.Name,
				&movie.Path,
				&movie.ParentPath,
				&movie.Format,
				&movie.Size,
				&movie.CreatedAt,
				&movie.UpdatedAt,
			); err != nil {
				return err
			}
			movies = append(movies, movie)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListLibraryMovies: %w", err)
	}
	return movies, nil
}
