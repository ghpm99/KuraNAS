import { useQuery } from '@tanstack/react-query';
import { getMusicAlbums, getMusicArtists, getMusicFolders, getMusicGenres } from '@/service/music';
import { PlaylistRuleOption } from './PlaylistRuleOptionsField';

const OPTIONS_PAGE_SIZE = 500;
const OPTIONS_STALE_TIME_MS = 60_000;

export const useRuleOptions = (
    sourceName: string,
    loadOptions: () => Promise<PlaylistRuleOption[]>
) =>
    useQuery({
        queryKey: ['playlist-rule-options', sourceName],
        queryFn: loadOptions,
        staleTime: OPTIONS_STALE_TIME_MS,
    });

export const loadGenreOptions = async () =>
    ((await getMusicGenres(1, OPTIONS_PAGE_SIZE)).items ?? []).map((genre) => ({
        value: genre.key,
        label: genre.genre,
    }));

export const loadArtistOptions = async () =>
    ((await getMusicArtists(1, OPTIONS_PAGE_SIZE)).items ?? []).map((artist) => ({
        value: artist.key,
        label: artist.artist,
    }));

export const loadAlbumOptions = async () =>
    ((await getMusicAlbums(1, OPTIONS_PAGE_SIZE)).items ?? []).map((album) => ({
        value: album.key,
        label: album.artist ? `${album.album} - ${album.artist}` : album.album,
    }));

export const loadFolderOptions = async () =>
    ((await getMusicFolders(1, OPTIONS_PAGE_SIZE)).items ?? []).map((folder) => ({
        value: folder.folder,
        label: folder.folder,
    }));
