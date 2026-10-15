import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { useVideoPlaylistDetail, useVideoSectionPlaylists } from './useVideoQueries';

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
    useInfiniteQuery: jest.fn(),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoPlaylistsBySection: jest.fn((section: string, page: number, pageSize: number) =>
        Promise.resolve({ section, page, pageSize })
    ),
    getVideoPlaylistById: jest.fn((id: number) => Promise.resolve({ id })),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseInfiniteQuery = useInfiniteQuery as jest.Mock;

describe('providers/videoContentProvider/useVideoQueries', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseQuery.mockReturnValue({ data: null, status: 'success' });
    });

    it('builds infinite query options for the playlists of a section', async () => {
        useVideoSectionPlaylists('series', 24, true);
        const options = mockedUseInfiniteQuery.mock.calls[0][0];

        expect(options.queryKey).toEqual(['video', 'playlists', 'section', 'series', 24]);
        expect(options.enabled).toBe(true);
        expect(options.initialPageParam).toBe(1);
        await expect(options.queryFn({ pageParam: 3 })).resolves.toEqual({
            section: 'series',
            page: 3,
            pageSize: 24,
        });
        expect(options.getNextPageParam({ pagination: { page: 3, has_next: true } })).toBe(4);
        expect(options.getNextPageParam({ pagination: { page: 3, has_next: false } })).toBe(
            undefined
        );
        expect(options.getNextPageParam(undefined)).toBe(undefined);
    });

    it('keeps the section query disabled when it is not needed', () => {
        useVideoSectionPlaylists('movies', 4, false);

        expect(mockedUseInfiniteQuery.mock.calls[0][0].enabled).toBe(false);
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
});
