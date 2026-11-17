import { Pagination } from '@/types/pagination';
import {
    MusicAlbum,
    MusicAlbumSummary,
    MusicArtist,
    MusicArtistSummary,
    MusicFolder,
    MusicGenre,
    MusicGroupSummary,
    MusicHomeCatalog,
    MusicListSort,
    MusicMostPlayedPeriod,
    MusicPlayedTrack,
    MusicQueue,
    BatchTrackMetadataResult,
    BatchTrackMetadataSelector,
    RenamedMusicGenre,
    TrackMetadataFieldName,
    TrackMetadataFields,
} from '@/types/music';
import { IMusicData } from '@/types/music';
import { isAxiosError } from 'axios';
import { apiBase } from '.';
import { trackMetadataFieldNames } from './trackMetadataFieldNames';

const toSortParams = (listSort?: MusicListSort) =>
    listSort ? { sort: listSort.sort, order: listSort.order } : {};

export const getMusicHomeCatalog = async (limit: number, listSort?: MusicListSort) => {
    const response = await apiBase.get<MusicHomeCatalog>('/music/library/home', {
        params: { limit, ...toSortParams(listSort) },
    });
    return response.data;
};

export const getMusicArtists = async (page: number, pageSize: number, listSort?: MusicListSort) => {
    const response = await apiBase.get<Pagination<MusicArtist>>('/music/library/artists', {
        params: { page, page_size: pageSize, ...toSortParams(listSort) },
    });
    return response.data;
};

export const getMusicByArtist = async (artistKey: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>(
        `/music/library/artists/${encodeURIComponent(artistKey)}/tracks`,
        {
            params: { page, page_size: pageSize },
        }
    );
    return response.data;
};

export const getMusicAlbums = async (page: number, pageSize: number, listSort?: MusicListSort) => {
    const response = await apiBase.get<Pagination<MusicAlbum>>('/music/library/albums', {
        params: { page, page_size: pageSize, ...toSortParams(listSort) },
    });
    return response.data;
};

export const getMusicByAlbum = async (albumKey: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>(
        `/music/library/albums/${encodeURIComponent(albumKey)}/tracks`,
        {
            params: { page, page_size: pageSize },
        }
    );
    return response.data;
};

export const getMusicGenres = async (page: number, pageSize: number, listSort?: MusicListSort) => {
    const response = await apiBase.get<Pagination<MusicGenre>>('/music/library/genres', {
        params: { page, page_size: pageSize, ...toSortParams(listSort) },
    });
    return response.data;
};

export const getMusicByGenre = async (genreKey: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>(
        `/music/library/genres/${encodeURIComponent(genreKey)}/tracks`,
        {
            params: { page, page_size: pageSize },
        }
    );
    return response.data;
};

export const getMusicFolders = async (page: number, pageSize: number, listSort?: MusicListSort) => {
    const response = await apiBase.get<Pagination<MusicFolder>>('/music/library/folders', {
        params: { page, page_size: pageSize, ...toSortParams(listSort) },
    });
    return response.data;
};

export const getMusic = async (page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>('/music/library', {
        params: { page, page_size: pageSize },
    });
    return response.data;
};

export const searchMusicTracks = async (searchText: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>('/music/search', {
        params: { q: searchText, page, page_size: pageSize },
    });
    return response.data;
};

export const getMusicByFolder = async (folder: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<IMusicData>>(
        `/music/library/folders/${encodeURIComponent(folder)}/tracks`,
        {
            params: { page, page_size: pageSize },
        }
    );
    return response.data;
};

const getLibrarySummary = async <SummaryType>(groupPath: string) => {
    const response = await apiBase.get<SummaryType>(`/music/library/${groupPath}`);
    return response.data;
};

export const getMusicAlbumSummary = (albumKey: string) =>
    getLibrarySummary<MusicAlbumSummary>(`albums/${encodeURIComponent(albumKey)}`);

export const getMusicArtistSummary = (artistKey: string) =>
    getLibrarySummary<MusicArtistSummary>(`artists/${encodeURIComponent(artistKey)}`);

export const getMusicGenreSummary = (genreKey: string) =>
    getLibrarySummary<MusicGroupSummary>(`genres/${encodeURIComponent(genreKey)}`);

export const getMusicFolderSummary = (folderPath: string) =>
    getLibrarySummary<MusicGroupSummary>(`folders/${encodeURIComponent(folderPath)}`);

export const getMusicAlbumsByArtist = async (artistKey: string, page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<MusicAlbum>>(
        `/music/library/artists/${encodeURIComponent(artistKey)}/albums`,
        { params: { page, page_size: pageSize } }
    );
    return response.data;
};

const getMusicQueue = async (contextPath: string) => {
    const response = await apiBase.get<MusicQueue>(`/music/library/${contextPath}/queue`);
    return response.data;
};

export const getMusicQueueByArtist = (artistKey: string) =>
    getMusicQueue(`artists/${encodeURIComponent(artistKey)}`);

export const getMusicQueueByAlbum = (albumKey: string) =>
    getMusicQueue(`albums/${encodeURIComponent(albumKey)}`);

export const getMusicQueueByGenre = (genreKey: string) =>
    getMusicQueue(`genres/${encodeURIComponent(genreKey)}`);

export const getMusicQueueByFolder = (folderPath: string) =>
    getMusicQueue(`folders/${encodeURIComponent(folderPath)}`);

export const recordMusicPlay = async (fileId: number, playedSeconds: number): Promise<void> => {
    await apiBase.post('/music/plays', { file_id: fileId, played_seconds: playedSeconds });
};

export const getMostPlayedTracks = async (
    page: number,
    pageSize: number,
    period: MusicMostPlayedPeriod = 'all'
) => {
    const response = await apiBase.get<Pagination<MusicPlayedTrack>>('/music/library/most-played', {
        params: { page, page_size: pageSize, period },
    });
    return response.data;
};

export const getRecentlyPlayedTracks = async (page: number, pageSize: number) => {
    const response = await apiBase.get<Pagination<MusicPlayedTrack>>(
        '/music/library/recent-plays',
        { params: { page, page_size: pageSize } }
    );
    return response.data;
};

export const updateTrackMetadata = async (
    fileId: number,
    fields: TrackMetadataFields
): Promise<IMusicData> => {
    const response = await apiBase.put<IMusicData>(`/music/tracks/${fileId}/metadata`, fields);
    return response.data;
};

interface BatchTrackMetadataResponse {
    matched: number;
    updated: number;
    failed: { file_id: number; error: string }[] | null;
}

const toBatchRequestBody = (selector: BatchTrackMetadataSelector, fields: TrackMetadataFields) => {
    if ('albumKey' in selector) {
        return { fields, album_key: selector.albumKey };
    }
    if ('folderKey' in selector) {
        return { fields, folder_key: selector.folderKey };
    }
    return { fields, file_ids: selector.fileIds };
};

export const updateTracksMetadata = async (
    selector: BatchTrackMetadataSelector,
    fields: TrackMetadataFields,
    options: { isDryRun: boolean }
): Promise<BatchTrackMetadataResult> => {
    const url = '/music/tracks/metadata';
    const body = toBatchRequestBody(selector, fields);
    const response = options.isDryRun
        ? await apiBase.put<BatchTrackMetadataResponse>(url, body, { params: { dry_run: true } })
        : await apiBase.put<BatchTrackMetadataResponse>(url, body);
    return {
        matched: response.data.matched,
        updated: response.data.updated,
        failed: (response.data.failed ?? []).map((failure) => ({
            fileId: failure.file_id,
            error: failure.error,
        })),
    };
};

const readRefusedField = (error: unknown): unknown => {
    if (!isAxiosError(error) || error.response?.status !== 400) {
        return undefined;
    }
    return (error.response.data as { field?: unknown } | undefined)?.field;
};

export const readRefusedMetadataField = (error: unknown): TrackMetadataFieldName | undefined => {
    const refusedField = readRefusedField(error);
    return trackMetadataFieldNames.find((fieldName) => fieldName === refusedField);
};

export const isGenreNameRefused = (error: unknown): boolean => readRefusedField(error) === 'name';

interface RenamedMusicGenreResponse {
    key: string;
    name: string;
    track_count: number;
    merged: boolean;
}

export const renameMusicGenre = async (
    genreKey: string,
    name: string
): Promise<RenamedMusicGenre> => {
    const response = await apiBase.put<RenamedMusicGenreResponse>(
        `/music/library/genres/${encodeURIComponent(genreKey)}`,
        { name }
    );
    return {
        key: response.data.key,
        name: response.data.name,
        trackCount: response.data.track_count,
        merged: response.data.merged,
    };
};
