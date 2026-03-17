import { act, renderHook } from '@testing-library/react';
import useGlobalSearchProvider from './useGlobalSearchProvider';

const mockNavigate = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

jest.mock('react-router-dom', () => ({
    useNavigate: () => mockNavigate,
    useLocation: () => ({ pathname: '/files', search: '' }),
}));

const mockSearchData = {
    files: [
        {
            id: 1,
            name: 'FILE',
            path: '/path/file',
            format: 'mp4',
            type: 1,
            parent_path: '/',
            size: 1,
        },
    ],
    folders: [{ id: 2, name: 'Folder', path: '/folder' }],
    artists: [{ key: 'artist-1', artist: 'Artist', track_count: 2, album_count: 1 }],
    albums: [
        {
            key: 'album-1',
            album: 'Album',
            artist: 'Artist',
            track_count: 5,
            year: '2020',
        },
    ],
    playlists: [
        {
            id: 3,
            name: 'Music Playlist',
            scope: 'music',
            count: 5,
            source_path: '',
            classification: 'personal',
            description: 'desc',
        },
        {
            id: 4,
            name: 'Video Playlist',
            scope: 'video',
            count: 3,
            source_path: '/path',
            classification: 'series',
            description: 'folder',
        },
        {
            id: 5,
            name: 'Custom Video Playlist',
            scope: 'video',
            count: 2,
            source_path: '',
            classification: 'movie',
            description: 'custom desc',
        },
    ],
    videos: [{ id: 6, name: 'Video', path: '/video.mp4', format: 'video/mp4' }],
    images: [
        {
            id: 7,
            name: 'Image',
            path: '/image.jpg',
            context: 'Library',
            category: 'folder',
        },
    ],
};

let mockUseQueryReturn: {
    data: typeof mockSearchData | undefined;
    isFetching: boolean;
} = { data: undefined, isFetching: false };

jest.mock('@tanstack/react-query', () => ({
    useQuery: () => mockUseQueryReturn,
}));

const mockGetVideoSectionForPlaylist = jest.fn().mockReturnValue('series');
const mockGetVideoDetailRoute = jest.fn().mockReturnValue('/videos/series/video-playlist');

jest.mock('@/components/videos/navigation', () => ({
    getVideoDetailRoute: (...args: unknown[]) => mockGetVideoDetailRoute(...args),
    getVideoSectionForPlaylist: (...args: unknown[]) => mockGetVideoSectionForPlaylist(...args),
}));

jest.mock('@/app/routes', () => ({
    appRoutes: {
        home: '/home',
        files: '/files',
        favorites: '/favorites',
        legacyFavorites: '/starred',
        settings: '/settings',
        about: '/about',
        images: '/images',
        music: '/music',
        videos: '/videos',
        analytics: '/analytics',
        videoPlayerBase: '/video',
    },
    getMusicRoute: (section: string) => `/music/${section}`,
    getVideoRoute: (section: string) => `/videos/${section}`,
    getAnalyticsRoute: (section: string) => `/analytics/${section}`,
}));

describe('useGlobalSearchProvider', () => {
    let platformSpy: jest.SpyInstance;

    beforeEach(() => {
        platformSpy = jest.spyOn(window.navigator, 'platform', 'get').mockReturnValue('Win32');
        mockNavigate.mockReset();
        mockGetVideoSectionForPlaylist.mockReturnValue('series');
        mockGetVideoDetailRoute.mockReturnValue('/videos/series/video-playlist');
        mockUseQueryReturn = { data: undefined, isFetching: false };
    });

    afterEach(() => {
        platformSpy.mockRestore();
    });

    describe('platform detection (shortcut)', () => {
        it('returns Ctrl+K on non-Mac platforms', () => {
            platformSpy.mockReturnValue('Win32');
            const { result } = renderHook(() => useGlobalSearchProvider());
            expect(result.current.shortcut).toBe('Ctrl+K');
        });

        it('returns Cmd+K on Mac platforms', () => {
            platformSpy.mockReturnValue('MacIntel');
            const { result } = renderHook(() => useGlobalSearchProvider());
            expect(result.current.shortcut).toBe('Cmd+K');
        });
    });

    describe('keyboard shortcut listener (Ctrl+K / Cmd+K)', () => {
        it('toggles open state on Ctrl+K', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());
            expect(result.current.open).toBe(false);

            act(() => {
                window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }));
            });
            expect(result.current.open).toBe(true);

            act(() => {
                window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }));
            });
            expect(result.current.open).toBe(false);
        });

        it('toggles open state on Cmd+K (metaKey)', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                window.dispatchEvent(new KeyboardEvent('keydown', { key: 'K', metaKey: true }));
            });
            expect(result.current.open).toBe(true);
        });

        it('does not toggle on plain K key without modifier', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k' }));
            });
            expect(result.current.open).toBe(false);
        });

        it('cleans up event listener on unmount', () => {
            const removeSpy = jest.spyOn(window, 'removeEventListener');
            const { unmount } = renderHook(() => useGlobalSearchProvider());
            unmount();
            expect(removeSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
            removeSpy.mockRestore();
        });
    });

    describe('quick actions and matchesQuery', () => {
        it('returns all quick actions when query is empty', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const actionSection = result.current.sections.find((s) => s.id === 'actions');
            expect(actionSection).toBeDefined();
            expect(actionSection!.items.length).toBe(14);
        });

        it('filters quick actions based on query matching label or description', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('SETTINGS');
            });

            const actionSection = result.current.sections.find((s) => s.id === 'actions');
            expect(actionSection).toBeDefined();
            expect(actionSection!.items.some((item) => item.id === 'action-settings')).toBe(true);
        });

        it('returns empty sections when no actions match and no data', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('xyznonexistent999');
            });

            expect(result.current.sections.length).toBe(0);
        });

        it('navigates to correct routes for each quick action', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const actionSection = result.current.sections.find((s) => s.id === 'actions')!;
            const expectedRoutes: Record<string, string | object> = {
                'action-home': '/home',
                'action-files': '/files',
                'action-favorites': '/favorites',
                'action-images': '/images',
                'action-music': '/music',
                'action-music-artists': '/music/artists',
                'action-music-albums': '/music/albums',
                'action-music-playlists': '/music/playlists',
                'action-videos': '/videos',
                'action-videos-continue': '/videos/continue',
                'action-analytics': '/analytics',
                'action-analytics-library': '/analytics/library',
                'action-settings': '/settings',
                'action-about': '/about',
            };

            for (const item of actionSection.items) {
                mockNavigate.mockClear();
                act(() => {
                    item.onSelect();
                });
                expect(mockNavigate).toHaveBeenCalledWith(expectedRoutes[item.id]);
            }
        });
    });

    describe('search result mapping', () => {
        beforeEach(() => {
            mockUseQueryReturn = { data: mockSearchData, isFetching: false };
        });

        it('maps file results correctly', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const filesSection = result.current.sections.find((s) => s.id === 'files');
            expect(filesSection).toBeDefined();
            const fileItem = filesSection!.items[0];
            expect(fileItem.id).toBe('file-1');
            expect(fileItem.kind).toBe('file');
            expect(fileItem.label).toBe('FILE');
            expect(fileItem.description).toBe('/path/file');
            expect(fileItem.meta).toBe('mp4');

            act(() => {
                fileItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith('/files/path/file');
        });

        it('maps folder results correctly', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const foldersSection = result.current.sections.find((s) => s.id === 'folders');
            expect(foldersSection).toBeDefined();
            const folderItem = foldersSection!.items[0];
            expect(folderItem.id).toBe('folder-2');
            expect(folderItem.kind).toBe('folder');

            act(() => {
                folderItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith('/files/folder');
        });

        it('maps artist results correctly', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const artistsSection = result.current.sections.find((s) => s.id === 'artists');
            expect(artistsSection).toBeDefined();
            const artistItem = artistsSection!.items[0];
            expect(artistItem.id).toBe('artist-artist-1');
            expect(artistItem.kind).toBe('artist');
            expect(artistItem.label).toBe('Artist');

            act(() => {
                artistItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith({
                pathname: '/music/artists',
                search: '?artist=artist-1',
            });
        });

        it('maps album results correctly', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const albumsSection = result.current.sections.find((s) => s.id === 'albums');
            expect(albumsSection).toBeDefined();
            const albumItem = albumsSection!.items[0];
            expect(albumItem.id).toBe('album-album-1');
            expect(albumItem.kind).toBe('album');
            expect(albumItem.label).toBe('Album');
            expect(albumItem.meta).toBe('2020');

            act(() => {
                albumItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith({
                pathname: '/music/albums',
                search: '?album=album-1',
            });
        });

        it('maps music playlist results and navigates to music playlists route', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            expect(playlistsSection).toBeDefined();
            const musicPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-music-3');
            expect(musicPlaylist).toBeDefined();
            expect(musicPlaylist!.kind).toBe('playlist');

            act(() => {
                musicPlaylist!.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith({
                pathname: '/music/playlists',
                search: '?playlist=3',
            });
        });

        it('maps video playlist with source_path (folder type) and navigates via getVideoDetailRoute', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            const videoPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-video-4');
            expect(videoPlaylist).toBeDefined();
            expect(videoPlaylist!.meta).toBe('series');

            mockNavigate.mockClear();
            act(() => {
                videoPlaylist!.onSelect();
            });

            expect(mockGetVideoSectionForPlaylist).toHaveBeenCalledWith({
                type: 'folder',
                classification: 'series',
            });
            expect(mockGetVideoDetailRoute).toHaveBeenCalledWith('series', 'video-playlist');
            expect(mockNavigate).toHaveBeenCalled();
        });

        it('maps video playlist without source_path (custom type) and navigates via getVideoDetailRoute', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            const customPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-video-5');
            expect(customPlaylist).toBeDefined();
            expect(customPlaylist!.meta).toBe('movie');

            mockNavigate.mockClear();
            mockGetVideoSectionForPlaylist.mockClear();
            act(() => {
                customPlaylist!.onSelect();
            });

            expect(mockGetVideoSectionForPlaylist).toHaveBeenCalledWith({
                type: 'custom',
                classification: 'movie',
            });
            expect(mockNavigate).toHaveBeenCalled();
        });

        it('maps video results and navigates with state', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const videosSection = result.current.sections.find((s) => s.id === 'videos');
            expect(videosSection).toBeDefined();
            const videoItem = videosSection!.items[0];
            expect(videoItem.id).toBe('video-6');
            expect(videoItem.kind).toBe('video');

            act(() => {
                videoItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith('/video/6', {
                state: { from: '/files', playlistId: null },
            });
        });

        it('maps image results and navigates with query params', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const imagesSection = result.current.sections.find((s) => s.id === 'images');
            expect(imagesSection).toBeDefined();
            const imageItem = imagesSection!.items[0];
            expect(imageItem.id).toBe('image-7');
            expect(imageItem.kind).toBe('image');
            expect(imageItem.meta).toBe('Library');

            act(() => {
                imageItem.onSelect();
            });
            expect(mockNavigate).toHaveBeenCalledWith({
                pathname: '/images',
                search: '?image=7&imagePath=%2Fimage.jpg',
            });
        });

        it('uses category as meta when context is empty', () => {
            mockUseQueryReturn = {
                data: {
                    ...mockSearchData,
                    images: [
                        {
                            id: 8,
                            name: 'NoCtx',
                            path: '/img2.jpg',
                            context: '',
                            category: 'photos',
                        },
                    ],
                },
                isFetching: false,
            };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const imagesSection = result.current.sections.find((s) => s.id === 'images');
            const imageItem = imagesSection!.items[0];
            expect(imageItem.meta).toBe('photos');
        });

        it('music playlist description uses NAV_MUSIC scope', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            const musicPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-music-3');
            expect(musicPlaylist!.description).toContain('NAV_MUSIC');
        });

        it('video playlist description uses NAV_VIDEOS scope', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            const videoPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-video-4');
            expect(videoPlaylist!.description).toContain('NAV_VIDEOS');
        });

        it('music playlist meta uses description field', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('search');
            });

            const playlistsSection = result.current.sections.find((s) => s.id === 'playlists');
            const musicPlaylist = playlistsSection!.items.find((i) => i.id === 'playlist-music-3');
            expect(musicPlaylist!.meta).toBe('desc');
        });
    });

    describe('empty state logic', () => {
        it('showEmptyState is true when query >= 2 chars, not fetching, and no sections', () => {
            mockUseQueryReturn = {
                data: {
                    files: [],
                    folders: [],
                    artists: [],
                    albums: [],
                    playlists: [],
                    videos: [],
                    images: [],
                },
                isFetching: false,
            };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('xyznonexistent999');
            });

            expect(result.current.showEmptyState).toBe(true);
        });

        it('showEmptyState is false when still fetching', () => {
            mockUseQueryReturn = { data: undefined, isFetching: true };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('xyznonexistent999');
            });

            expect(result.current.showEmptyState).toBe(false);
        });

        it('showEmptyState is false when query is too short', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('a');
            });

            expect(result.current.showEmptyState).toBe(false);
        });
    });

    describe('keyboard navigation (handleInputKeyDown)', () => {
        it('does nothing when flattenedItems is empty', () => {
            mockUseQueryReturn = {
                data: {
                    files: [],
                    folders: [],
                    artists: [],
                    albums: [],
                    playlists: [],
                    videos: [],
                    images: [],
                },
                isFetching: false,
            };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('xyznonexistent999');
            });

            expect(result.current.sections.length).toBe(0);

            const event = { key: 'ArrowDown', preventDefault: jest.fn() } as any;
            act(() => {
                result.current.handleInputKeyDown(event);
            });
            expect(event.preventDefault).not.toHaveBeenCalled();
        });

        it('ArrowDown wraps around to first item from last item', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const totalItems = result.current.sections.flatMap((s) => s.items).length;
            expect(totalItems).toBeGreaterThan(0);

            // Navigate to last item
            for (let i = 0; i < totalItems - 1; i++) {
                act(() => {
                    result.current.handleInputKeyDown({
                        key: 'ArrowDown',
                        preventDefault: jest.fn(),
                    } as any);
                });
            }

            // Verify we're at the last item
            const lastItem = result.current.sections.flatMap((s) => s.items)[totalItems - 1];
            expect(result.current.activeItemId).toBe(lastItem.id);

            // One more ArrowDown should wrap to first
            act(() => {
                result.current.handleInputKeyDown({
                    key: 'ArrowDown',
                    preventDefault: jest.fn(),
                } as any);
            });
            const firstItem = result.current.sections.flatMap((s) => s.items)[0];
            expect(result.current.activeItemId).toBe(firstItem.id);
        });

        it('ArrowUp wraps around to last item from first item', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const totalItems = result.current.sections.flatMap((s) => s.items).length;
            expect(totalItems).toBeGreaterThan(0);

            // At index 0, ArrowUp should go to last item
            const event = { key: 'ArrowUp', preventDefault: jest.fn() } as any;
            act(() => {
                result.current.handleInputKeyDown(event);
            });
            expect(event.preventDefault).toHaveBeenCalled();

            const lastItem = result.current.sections.flatMap((s) => s.items)[totalItems - 1];
            expect(result.current.activeItemId).toBe(lastItem.id);
        });

        it('Enter key activates the current item and closes search', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const event = { key: 'Enter', preventDefault: jest.fn() } as any;
            act(() => {
                result.current.handleInputKeyDown(event);
            });

            expect(event.preventDefault).toHaveBeenCalled();
            expect(mockNavigate).toHaveBeenCalledWith('/home');
            expect(result.current.open).toBe(false);
            expect(result.current.query).toBe('');
        });

        it('ignores unrelated keys', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const event = { key: 'Escape', preventDefault: jest.fn() } as any;
            act(() => {
                result.current.handleInputKeyDown(event);
            });
            expect(event.preventDefault).not.toHaveBeenCalled();
        });
    });

    describe('onSelect handler (activateItem)', () => {
        it('calls onSelect and closes search dialog', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            const firstItem = result.current.sections[0].items[0];
            mockNavigate.mockClear();

            act(() => {
                result.current.activateItem(firstItem);
            });

            expect(mockNavigate).toHaveBeenCalled();
            expect(result.current.open).toBe(false);
            expect(result.current.query).toBe('');
        });
    });

    describe('openSearch and closeSearch', () => {
        it('openSearch resets activeIndex and opens dialog', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            expect(result.current.open).toBe(true);
        });

        it('closeSearch resets query and activeIndex', () => {
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('test');
            });

            act(() => {
                result.current.closeSearch();
            });

            expect(result.current.open).toBe(false);
            expect(result.current.query).toBe('');
        });
    });

    describe('setQuery resets activeIndex', () => {
        it('resets activeIndex when query changes', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
            });

            // Move index forward
            act(() => {
                result.current.handleInputKeyDown({
                    key: 'ArrowDown',
                    preventDefault: jest.fn(),
                } as any);
            });

            // Changing query should reset index back to 0
            act(() => {
                result.current.setQuery('HOME');
            });

            const firstItem = result.current.sections.flatMap((s) => s.items)[0];
            expect(firstItem).toBeDefined();
            expect(result.current.activeItemId).toBe(firstItem.id);
        });
    });

    describe('data absent returns only actions', () => {
        it('returns only action sections when data is undefined', () => {
            mockUseQueryReturn = { data: undefined, isFetching: false };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('HOME');
            });

            expect(result.current.sections.length).toBe(1);
            expect(result.current.sections[0].id).toBe('actions');
        });
    });

    describe('sections with empty result arrays are excluded', () => {
        it('omits sections when their result arrays are empty', () => {
            mockUseQueryReturn = {
                data: {
                    files: [],
                    folders: [],
                    artists: [],
                    albums: [],
                    playlists: [],
                    videos: [],
                    images: [
                        {
                            id: 10,
                            name: 'OnlyImage',
                            path: '/only.jpg',
                            context: 'ctx',
                            category: 'cat',
                        },
                    ],
                },
                isFetching: false,
            };
            const { result } = renderHook(() => useGlobalSearchProvider());

            act(() => {
                result.current.openSearch();
                result.current.setQuery('OnlyImage');
            });

            const sectionIds = result.current.sections.map((s) => s.id);
            expect(sectionIds).not.toContain('files');
            expect(sectionIds).not.toContain('folders');
            expect(sectionIds).not.toContain('artists');
            expect(sectionIds).not.toContain('albums');
            expect(sectionIds).not.toContain('playlists');
            expect(sectionIds).not.toContain('videos');
            expect(sectionIds).toContain('images');
        });
    });
});
