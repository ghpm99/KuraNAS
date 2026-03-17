import { useQuery } from '@tanstack/react-query';
import {
    useAllVideoFiles,
    useVideoHomeCatalog,
    useVideoPlaylistDetail,
    useVideoPlaylists,
    useVideosWithoutPlaylist,
} from './useVideoQueries';

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoPlaylists: jest.fn(() => Promise.resolve(['playlists'])),
    getVideoPlaylistById: jest.fn((id: number) => Promise.resolve({ id })),
    getVideosWithoutPlaylist: jest.fn((limit: number) => Promise.resolve([{ limit }])),
    getAllVideoFiles: jest.fn((limit: number) => Promise.resolve([{ limit }])),
    getVideoHomeCatalog: jest.fn((limit: number) => Promise.resolve({ limit })),
}));

const mockedUseQuery = useQuery as jest.Mock;

describe('providers/videoContentProvider/useVideoQueries', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseQuery.mockReturnValue({ data: null, status: 'success' });
    });

    it('builds query options for video playlists', async () => {
        useVideoPlaylists();
        const options = mockedUseQuery.mock.calls[0][0];

        expect(options.queryKey).toEqual(['video', 'playlists']);
        await expect(options.queryFn()).resolves.toEqual(['playlists']);
    });

    it('builds query options for playlist details with enabled flag', async () => {
        useVideoPlaylistDetail();
        let options = mockedUseQuery.mock.calls[0][0];
        expect(options.queryKey).toEqual(['video', 'playlist-detail', undefined]);
        expect(options.enabled).toBe(false);

        jest.clearAllMocks();
        useVideoPlaylistDetail(7);
        options = mockedUseQuery.mock.calls[0][0];
        expect(options.queryKey).toEqual(['video', 'playlist-detail', 7]);
        expect(options.enabled).toBe(true);
        await expect(options.queryFn()).resolves.toEqual({ id: 7 });
    });

    it('builds query options for unassigned/all files and home catalog', async () => {
        useVideosWithoutPlaylist();
        useAllVideoFiles();
        useVideoHomeCatalog();

        const unassigned = mockedUseQuery.mock.calls[0][0];
        const allFiles = mockedUseQuery.mock.calls[1][0];
        const homeCatalog = mockedUseQuery.mock.calls[2][0];

        expect(unassigned.queryKey).toEqual(['video', 'unassigned']);
        await expect(unassigned.queryFn()).resolves.toEqual([{ limit: 2000 }]);

        expect(allFiles.queryKey).toEqual(['video', 'all-files']);
        await expect(allFiles.queryFn()).resolves.toEqual([{ limit: 3000 }]);

        expect(homeCatalog.queryKey).toEqual(['video', 'home-catalog']);
        await expect(homeCatalog.queryFn()).resolves.toEqual({ limit: 24 });
    });
});
