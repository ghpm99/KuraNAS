import { useInfiniteQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getVideoLibraryMovies, type VideoMovieSort } from '@/service/videoPlayback';

export const videoMoviesPageSize = 24;

export const useVideoMovieBrowser = () => {
    const [sort, setSort] = useState<VideoMovieSort>('name');

    const movieQuery = useInfiniteQuery({
        queryKey: ['videos', 'library-movies', sort],
        queryFn: ({ pageParam }) => getVideoLibraryMovies(sort, pageParam, videoMoviesPageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage, loadedPages) =>
            lastPage?.pagination?.has_next ? loadedPages.length + 1 : undefined,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    return {
        sort,
        setSort,
        movieQuery,
        movies: movieQuery.data?.pages.flatMap((page) => page.items ?? []) ?? [],
    };
};
