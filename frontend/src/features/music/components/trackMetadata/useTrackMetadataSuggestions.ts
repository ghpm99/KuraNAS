import { useQuery } from '@tanstack/react-query';
import { getMusicArtists, getMusicGenres } from '@/service/music';

const suggestionPageSize = 500;
const nameAscending = { sort: 'name', order: 'asc' } as const;

export default function useTrackMetadataSuggestions() {
    const { data: artistPage } = useQuery({
        queryKey: ['music-metadata-suggestions', 'artists'],
        queryFn: () => getMusicArtists(1, suggestionPageSize, nameAscending),
    });
    const { data: genrePage } = useQuery({
        queryKey: ['music-metadata-suggestions', 'genres'],
        queryFn: () => getMusicGenres(1, suggestionPageSize, nameAscending),
    });

    const genreGroups = genrePage?.items ?? [];

    return {
        artistNames: artistPage?.items?.map((artistGroup) => artistGroup.artist) ?? [],
        genreGroups,
        genreNames: genreGroups.map((genreGroup) => genreGroup.genre),
    };
}
