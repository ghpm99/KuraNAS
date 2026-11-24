import { PlaylistRules } from '@/types/playlist';

export const createEmptyPlaylistRules = (): PlaylistRules => ({
    match: 'any',
    genres: [],
    artists: [],
    albums: [],
    folders: [],
    year_from: null,
    year_to: null,
    favorites_only: false,
});

export const normalizePlaylistRules = (
    partialRules: Partial<PlaylistRules> | null | undefined
): PlaylistRules => {
    const emptyRules = createEmptyPlaylistRules();
    return {
        match: partialRules?.match === 'all' ? 'all' : 'any',
        genres: partialRules?.genres ?? emptyRules.genres,
        artists: partialRules?.artists ?? emptyRules.artists,
        albums: partialRules?.albums ?? emptyRules.albums,
        folders: partialRules?.folders ?? emptyRules.folders,
        year_from: partialRules?.year_from ?? null,
        year_to: partialRules?.year_to ?? null,
        favorites_only: partialRules?.favorites_only ?? false,
    };
};

export const hasPlaylistRuleCriteria = (rules: PlaylistRules): boolean =>
    rules.genres.length > 0 ||
    rules.artists.length > 0 ||
    rules.albums.length > 0 ||
    rules.folders.length > 0 ||
    rules.year_from !== null ||
    rules.year_to !== null ||
    rules.favorites_only;
