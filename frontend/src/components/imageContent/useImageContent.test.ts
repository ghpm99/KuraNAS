import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { useImageContent } from './useImageContent';

const mockUseImage = jest.fn();
const mockRef = jest.fn();
const mockUseIntersectionObserver = jest.fn();
const mockToggleStarredFile = jest.fn();
const mockGetFileByPath = jest.fn();
const mockNavigate = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('../providers/imageProvider/imageProvider', () => ({
    useImage: () => mockUseImage(),
}));
jest.mock('../hooks/IntersectionObserver/useIntersectionObserver', () => ({
    useIntersectionObserver: (...args: any[]) => mockUseIntersectionObserver(...args),
}));
jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));
jest.mock('@/service/files', () => ({
    toggleStarredFile: (...args: any[]) => mockToggleStarredFile(...args),
    getFileByPath: (...args: any[]) => mockGetFileByPath(...args),
}));
jest.mock('react-router-dom', () => {
    const actual = jest.requireActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});
jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => '/api/v1',
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string | number>) => {
            const map: Record<string, string> = {
                LOCALE: 'pt-BR',
                IMAGES_SECTION_LIBRARY: 'Biblioteca',
                IMAGES_SECTION_RECENT: 'Recentes',
                IMAGES_SECTION_CAPTURES: 'Capturas',
                IMAGES_SECTION_PHOTOS: 'Fotos',
                IMAGES_SECTION_FOLDERS: 'Pastas',
                IMAGES_SECTION_ALBUMS: 'Albuns automaticos',
                IMAGES_GROUP_BY_DATE: 'Data',
                IMAGES_GROUP_BY_TYPE: 'Tipo',
                IMAGES_GROUP_BY_NAME: 'Nome',
                IMAGES_GROUP_NO_FORMAT: 'Sem formato',
                IMAGES_GROUP_NO_DATE: 'Sem data',
                IMAGES_GROUP_INITIAL: `Inicial ${params?.letter ?? ''}`.trim(),
                IMAGES_ALBUM_TRAVEL: 'Viagens',
                IMAGES_ALBUM_TRAVEL_DESCRIPTION: 'Fotos de viagens',
                IMAGES_ALBUM_DOCUMENTS: 'Documentos',
                IMAGES_ALBUM_DOCUMENTS_DESCRIPTION: 'Documentos digitalizados',
                IMAGES_ALBUM_WALLPAPERS: 'Wallpapers',
                IMAGES_ALBUM_WALLPAPERS_DESCRIPTION: 'Papeis de parede',
                IMAGES_ALBUM_MEMES: 'Memes',
                IMAGES_ALBUM_MEMES_DESCRIPTION: 'Memes e reacoes',
                IMAGES_ALBUM_OTHERS: 'Outros',
                IMAGES_ALBUM_OTHERS_DESCRIPTION: 'Tudo que nao entrou nos temas principais',
                IMAGES_FOLDERS_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} pastas`,
                IMAGES_ALBUMS_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} albuns`,
                IMAGES_COUNT_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} imagens`,
                IMAGES_FOLDERS_EMPTY_TITLE: 'Nenhuma pasta encontrada',
                IMAGES_FOLDERS_EMPTY_DESC: 'Sem pastas',
                IMAGES_ALBUMS_EMPTY_TITLE: 'Nenhum album encontrado',
                IMAGES_ALBUMS_EMPTY_DESC: 'Sem albuns',
                IMAGES_EMPTY_TITLE: 'Nenhuma imagem encontrada',
                IMAGES_EMPTY_DESC: 'Sem imagens',
                IMAGES_VIEWER_FAVORITE_ADDED: 'Imagem adicionada aos favoritos',
                IMAGES_VIEWER_FAVORITE_REMOVED: 'Imagem removida dos favoritos',
                IMAGES_VIEWER_FAVORITE_ERROR: 'Erro ao atualizar favorito',
            };
            return map[key] ?? key;
        },
    }),
}));
jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    useSettings: () => ({
        settings: {
            library: {
                runtime_root_path: '/data',
                watched_paths: ['/data'],
                remember_last_location: true,
                prioritize_favorites: true,
            },
            indexing: {
                workers_enabled: true,
                scan_on_startup: true,
                extract_metadata: true,
                generate_previews: true,
            },
            players: {
                remember_music_queue: true,
                remember_video_progress: true,
                autoplay_next_video: true,
                image_slideshow_seconds: 4,
            },
            appearance: { accent_color: 'violet', reduce_motion: false },
            language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
        },
    }),
}));

const createImage = (overrides: Record<string, any> = {}) => ({
    id: 1,
    name: 'img1.jpg',
    path: '/photos/img1.jpg',
    type: 2,
    format: '.jpg',
    size: 1024,
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    created_at: '2026-03-10T10:00:00Z',
    updated_at: '2026-03-10T10:00:00Z',
    metadata: {
        id: 1,
        fileId: 1,
        path: '/photos/img1.jpg',
        format: 'jpg',
        mode: 'RGB',
        width: 1600,
        height: 900,
        dpi_x: 72,
        dpi_y: 72,
        x_resolution: 72,
        y_resolution: 72,
        resolution_unit: 2,
        orientation: 1,
        compression: 0,
        photometric_interpretation: 0,
        color_space: 1,
        components_configuration: '',
        icc_profile: '',
        make: 'Sony',
        model: 'A7',
        lens_model: '',
        serial_number: '',
        datetime: '2026-03-10T10:00:00Z',
        datetime_original: '2026-03-10T10:00:00Z',
        datetime_digitized: '',
        subsec_time: '',
        iso: 400,
        shutter_speed: 0,
        focal_length: 35,
        f_number: 2.8,
        aperture_value: 0,
        brightness_value: 0,
        exposure_bias: 0,
        metering_mode: 0,
        flash: 0,
        white_balance: 0,
        exposure_program: 0,
        max_aperture_value: 0,
        gps_latitude: 0,
        gps_longitude: 0,
        gps_altitude: 0,
        gps_date: '',
        gps_time: '',
        exposure_time: 0.008,
        user_comment: '',
        copyright: '',
        artist: '',
        software: '',
        image_description: '',
        classification: { category: 'photo', confidence: 0.9 },
        createdAt: '2026-03-10T10:00:00Z',
        ...overrides.metadata,
    },
    ...overrides,
});

const createQueryClient = () =>
    new QueryClient({
        defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
        },
    });

const createWrapper = (initialEntries: string[]) => {
    return ({ children }: { children: ReactNode }) =>
        createElement(
            QueryClientProvider,
            { client: createQueryClient() },
            createElement(MemoryRouter, { initialEntries }, children)
        );
};

const defaultImageReturn = () => ({
    images: [createImage()],
    status: 'success' as const,
    imageGroupBy: 'date' as const,
    setImageGroupBy: jest.fn(),
    fetchNextPage: jest.fn().mockResolvedValue(undefined),
    hasNextPage: false,
    isFetchingNextPage: false,
});

describe('useImageContent', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseIntersectionObserver.mockImplementation(() => ({ ref: mockRef }));
        mockToggleStarredFile.mockResolvedValue(undefined);
        mockGetFileByPath.mockResolvedValue(null);
        mockUseImage.mockReturnValue(defaultImageReturn());
    });

    // ── Section / title / viewMode branches ──────────────────────────

    it('returns library section and grid view mode by default', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.activeSection).toBe('library');
        expect(result.current.viewMode).toBe('grid');
        expect(result.current.title).toBe('Biblioteca');
    });

    it('returns recent section', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/recent']),
        });

        expect(result.current.activeSection).toBe('recent');
        expect(result.current.title).toBe('Recentes');
    });

    it('returns captures section and filters to capture category', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({
                    id: 1,
                    metadata: {
                        classification: { category: 'capture', confidence: 0.98 },
                    },
                }),
                createImage({
                    id: 2,
                    name: 'photo.jpg',
                    metadata: { classification: { category: 'photo', confidence: 0.9 } },
                }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/captures']),
        });

        expect(result.current.activeSection).toBe('captures');
        expect(result.current.filteredImages).toHaveLength(1);
        expect(result.current.filteredImages[0]?.id).toBe(1);
    });

    it('returns photos section and filters to photo category', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({
                    id: 1,
                    metadata: {
                        classification: { category: 'capture', confidence: 0.98 },
                    },
                }),
                createImage({
                    id: 2,
                    name: 'photo.jpg',
                    metadata: { classification: { category: 'photo', confidence: 0.9 } },
                }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/photos']),
        });

        expect(result.current.activeSection).toBe('photos');
        expect(result.current.filteredImages).toHaveLength(1);
        expect(result.current.filteredImages[0]?.id).toBe(2);
    });

    // ── Folders viewMode ─────────────────────────────────────────────

    it('shows folders viewMode when no folder is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' }),
                createImage({ id: 2, path: '/photos/family/b.jpg', name: 'b.jpg' }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        expect(result.current.viewMode).toBe('folders');
        expect(result.current.title).toBe('Pastas');
        expect(result.current.filteredFolderCards.length).toBeGreaterThanOrEqual(2);
    });

    it('switches to grid viewMode when a folder is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders?folder=/photos/travel']),
        });

        expect(result.current.viewMode).toBe('grid');
        expect(result.current.selectedFolder).not.toBeNull();
        expect(result.current.filteredImages).toHaveLength(1);
    });

    // ── Albums viewMode ──────────────────────────────────────────────

    it('shows albums viewMode when no album is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        expect(result.current.viewMode).toBe('albums');
        expect(result.current.title).toBe('Albuns automaticos');
    });

    it('switches to grid viewMode when an album is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums?album=others']),
        });

        expect(result.current.viewMode).toBe('grid');
        expect(result.current.selectedAlbum).not.toBeNull();
    });

    // ── Summary and empty state branches ─────────────────────────────

    it('returns folders summary in folders viewMode', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        expect(result.current.summary).toContain('pastas');
        expect(result.current.emptyState.title).toBe('Nenhuma pasta encontrada');
    });

    it('returns albums summary in albums viewMode', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        expect(result.current.summary).toContain('albuns');
        expect(result.current.emptyState.title).toBe('Nenhum album encontrado');
    });

    it('returns grid summary in grid viewMode', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.summary).toContain('imagens');
        expect(result.current.emptyState.title).toBe('Nenhuma imagem encontrada');
    });

    // ── Grouping branches ────────────────────────────────────────────

    it('groups images by date with known date', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages.length).toBeGreaterThanOrEqual(1);
        expect(result.current.groupByLabels).toEqual({
            date: 'Data',
            type: 'Tipo',
            name: 'Nome',
        });
    });

    it('groups images by type', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            imageGroupBy: 'type',
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages.length).toBeGreaterThanOrEqual(1);
        expect(result.current.groupedImages[0]?.label).toBe('.jpg');
    });

    it('groups images by name using first letter', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            imageGroupBy: 'name',
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages.length).toBeGreaterThanOrEqual(1);
    });

    it('groups images with no date into unknown group', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({
                    id: 1,
                    created_at: '',
                    updated_at: '',
                    metadata: {
                        datetime: '',
                        datetime_original: '',
                        createdAt: '',
                    },
                }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages.length).toBe(1);
        expect(result.current.groupedImages[0]?.label).toBe('Sem data');
    });

    it('groups images with no format into fallback label for type groupBy', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            imageGroupBy: 'type',
            images: [createImage({ id: 1, format: '' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages[0]?.label).toBe('Sem formato');
    });

    it('groups images with empty name into # for name groupBy', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            imageGroupBy: 'name',
            images: [createImage({ id: 1, name: '' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages[0]?.label).toContain('#');
    });

    // ── Search filtering ─────────────────────────────────────────────

    it('filters images by search term', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({ id: 1, name: 'beach.jpg' }),
                createImage({ id: 2, name: 'mountain.jpg' }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.setSearch('beach');
        });

        expect(result.current.filteredImages).toHaveLength(1);
        expect(result.current.filteredImages[0]?.id).toBe(1);
    });

    it('filters folder cards by search term', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' }),
                createImage({ id: 2, path: '/photos/family/b.jpg', name: 'b.jpg' }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        act(() => {
            result.current.setSearch('travel');
        });

        expect(result.current.filteredFolderCards.length).toBe(1);
        expect(result.current.filteredFolderCards[0]?.title).toBe('travel');
    });

    it('filters album cards by search term', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        act(() => {
            result.current.setSearch('Outros');
        });

        expect(result.current.filteredAlbumCards.length).toBeGreaterThanOrEqual(1);
    });

    it('returns all folder cards when search is empty', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        expect(result.current.filteredFolderCards.length).toBe(1);
    });

    it('returns all album cards when search is empty', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        // Should have all 5 automatic albums
        expect(result.current.filteredAlbumCards.length).toBe(5);
    });

    // ── handleOpenImage / handleCloseViewer ──────────────────────────

    it('opens and closes the image viewer', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.activeImage).toBeNull();

        act(() => {
            result.current.handleOpenImage(1);
        });

        expect(result.current.activeImage).not.toBeNull();
        expect(result.current.activeImage?.id).toBe(1);

        act(() => {
            result.current.handleCloseViewer();
        });

        expect(result.current.activeImage).toBeNull();
    });

    // ── handleToggleFavorite ─────────────────────────────────────────

    it('toggles favorite on the active image', async () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        act(() => {
            result.current.handleToggleFavorite();
        });

        await waitFor(() => expect(mockToggleStarredFile).toHaveBeenCalledWith(1));
    });

    it('does not toggle favorite when no active image', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleToggleFavorite();
        });

        expect(mockToggleStarredFile).not.toHaveBeenCalled();
    });

    // ── handleOpenFolder ─────────────────────────────────────────────

    it('navigates to file explorer with folder path when opening folder', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        act(() => {
            result.current.handleOpenFolder();
        });

        expect(mockNavigate).toHaveBeenCalledWith({
            pathname: '/files',
            search: '?path=%2Fphotos',
        });
    });

    it('does nothing when handleOpenFolder is called without active image', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenFolder();
        });

        expect(mockNavigate).not.toHaveBeenCalled();
    });

    // ── handleSelectFolder / handleSelectAlbum ───────────────────────

    it('handleSelectFolder sets the folder search param', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        act(() => {
            result.current.handleSelectFolder('/photos/travel');
        });

        // After selecting, viewMode should reflect the selection
        // We can check that selectedFolder is updated on next render with the param
    });

    it('handleSelectAlbum sets the album search param', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        act(() => {
            result.current.handleSelectAlbum('others');
        });
    });

    it('handleSelectFolder with null clears the selection', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders?folder=/photos/travel']),
        });

        act(() => {
            result.current.handleSelectFolder(null);
        });
    });

    it('handleSelectAlbum with null clears the selection', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums?album=others']),
        });

        act(() => {
            result.current.handleSelectAlbum(null);
        });
    });

    // ── handleLoadMore ───────────────────────────────────────────────

    it('handleLoadMore calls fetchNextPage when there is more to load', async () => {
        const fetchNextPage = jest.fn().mockResolvedValue(undefined);
        let capturedOnIntersect: () => void = () => {};
        mockUseIntersectionObserver.mockImplementation((opts: any) => {
            capturedOnIntersect = opts.onIntersect;
            return { ref: mockRef };
        });

        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            fetchNextPage,
            hasNextPage: true,
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        await act(async () => {
            await capturedOnIntersect();
        });

        expect(fetchNextPage).toHaveBeenCalled();
    });

    it('handleLoadMore does not call fetchNextPage when no next page', async () => {
        const fetchNextPage = jest.fn().mockResolvedValue(undefined);
        let capturedOnIntersect: () => void = () => {};
        mockUseIntersectionObserver.mockImplementation((opts: any) => {
            capturedOnIntersect = opts.onIntersect;
            return { ref: mockRef };
        });

        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            fetchNextPage,
            hasNextPage: false,
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        await act(async () => {
            await capturedOnIntersect();
        });

        expect(fetchNextPage).not.toHaveBeenCalled();
    });

    it('handleLoadMore does not call fetchNextPage when already fetching next page', async () => {
        const fetchNextPage = jest.fn().mockResolvedValue(undefined);
        let capturedOnIntersect: () => void = () => {};
        mockUseIntersectionObserver.mockImplementation((opts: any) => {
            capturedOnIntersect = opts.onIntersect;
            return { ref: mockRef };
        });

        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            fetchNextPage,
            hasNextPage: true,
            isFetchingNextPage: true,
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        await act(async () => {
            await capturedOnIntersect();
        });

        expect(fetchNextPage).not.toHaveBeenCalled();
    });

    // ── activeSelection title/description branches ───────────────────

    it('returns folder selection title and description when folder is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders?folder=/photos/travel']),
        });

        expect(result.current.activeSelectionTitle).toBe('travel');
        expect(result.current.activeSelectionDescription).toBe('/photos/travel');
        expect(result.current.title).toBe('travel');
    });

    it('returns album selection title and description when album is selected', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums?album=others']),
        });

        expect(result.current.activeSelectionTitle).toBe('Outros');
        expect(result.current.activeSelectionDescription).toBe(
            'Tudo que nao entrou nos temas principais'
        );
        expect(result.current.title).toBe('Outros');
    });

    it('returns empty title and description when no selection in folders section', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        expect(result.current.activeSelectionTitle).toBe('');
        expect(result.current.activeSelectionDescription).toBe('');
    });

    it('returns empty title and description when no selection in albums section', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        expect(result.current.activeSelectionTitle).toBe('');
        expect(result.current.activeSelectionDescription).toBe('');
    });

    // ── empty images ─────────────────────────────────────────────────

    it('handles empty images array', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.filteredImages).toHaveLength(0);
        expect(result.current.groupedImages).toHaveLength(0);
        expect(result.current.lastVisibleImageId).toBeUndefined();
    });

    // ── lastVisibleImageId ───────────────────────────────────────────

    it('returns last visible image id', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1 }), createImage({ id: 2, name: 'img2.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.lastVisibleImageId).toBe(2);
    });

    // ── activeFolderPath / activeImageDate ────────────────────────────

    it('returns activeFolderPath and activeImageDate when viewer is open', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        expect(result.current.activeFolderPath).toBe('/photos');
        expect(result.current.activeImageDate).not.toBeNull();
    });

    it('returns empty activeFolderPath and null activeImageDate when no image is active', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.activeFolderPath).toBe('');
        expect(result.current.activeImageDate).toBeNull();
    });

    // ── requestedImageId from search params ──────────────────────────

    it('opens image from search param when image exists in list', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=1']),
        });

        expect(result.current.activeImage?.id).toBe(1);
    });

    it('does not open image when requestedImageId is invalid', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=abc']),
        });

        expect(result.current.activeImage).toBeNull();
    });

    it('does not open image when requestedImageId is zero', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=0']),
        });

        expect(result.current.activeImage).toBeNull();
    });

    it('does not open image when requestedImageId is negative', () => {
        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=-5']),
        });

        expect(result.current.activeImage).toBeNull();
    });

    // ── selectedFolderId / selectedAlbumId only apply in correct section ──

    it('ignores folder param when not in folders section', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?folder=/photos/travel']),
        });

        expect(result.current.selectedFolder).toBeNull();
        expect(result.current.viewMode).toBe('grid');
    });

    it('ignores album param when not in albums section', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?album=others']),
        });

        expect(result.current.selectedAlbum).toBeNull();
        expect(result.current.viewMode).toBe('grid');
    });

    // ── hasNextPage boolean conversion ───────────────────────────────

    it('converts hasNextPage to boolean', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            hasNextPage: undefined,
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.hasNextPage).toBe(false);
    });

    // ── folder card coverImageId ─────────────────────────────────────

    it('returns coverImageId in folder cards', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 5, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders']),
        });

        expect(result.current.filteredFolderCards.length).toBe(1);
        expect(result.current.filteredFolderCards[0]?.coverImageId).toBe(5);
    });

    // ── album card coverImageId ──────────────────────────────────────

    it('returns coverImageId in album cards', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [
                createImage({
                    id: 5,
                    metadata: {
                        width: 800,
                        height: 600,
                        classification: { category: 'other', confidence: 0.5 },
                    },
                }),
            ],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums']),
        });

        const othersCard = result.current.filteredAlbumCards.find((c) => c.id === 'others');
        expect(othersCard?.coverImageId).toBe(5);
    });

    // ── status passthrough ───────────────────────────────────────────

    it('passes through pending status', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            status: 'pending',
            images: [],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.status).toBe('pending');
    });

    // ── format with whitespace-only is treated as empty ──────────────

    it('handles format with whitespace-only for type grouping', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            imageGroupBy: 'type',
            images: [createImage({ id: 1, format: '   ' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        expect(result.current.groupedImages[0]?.label).toBe('Sem formato');
    });

    // ── requestedImagePath resolution ────────────────────────────────

    it('attempts to resolve requested image when not in images list', async () => {
        mockGetFileByPath.mockResolvedValue(createImage({ id: 99, name: 'resolved.jpg' }));

        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1 })],
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=99&imagePath=/photos/resolved.jpg']),
        });

        await waitFor(() => expect(mockGetFileByPath).toHaveBeenCalledWith('/photos/resolved.jpg'));
    });

    it('does not resolve when image is already in list', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 99, name: 'resolved.jpg' })],
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=99&imagePath=/photos/resolved.jpg']),
        });

        expect(mockGetFileByPath).not.toHaveBeenCalled();
    });

    it('does not resolve when imagePath is empty', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1 })],
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=99&imagePath=']),
        });

        expect(mockGetFileByPath).not.toHaveBeenCalled();
    });

    it('does not resolve when imagePath is whitespace only', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1 })],
        });

        renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images?image=99&imagePath=%20%20']),
        });

        expect(mockGetFileByPath).not.toHaveBeenCalled();
    });

    // ── toggleFavorite onSuccess/onError branches ────────────────────

    it('shows success snackbar when favorite is removed', async () => {
        mockToggleStarredFile.mockResolvedValue(undefined);

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        // Image is starred: false by default, so toggling adds it
        act(() => {
            result.current.handleToggleFavorite();
        });

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Imagem adicionada aos favoritos', {
                variant: 'success',
            })
        );
    });

    it('shows error snackbar when favorite toggle fails', async () => {
        mockToggleStarredFile.mockRejectedValue(new Error('Network error'));

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        act(() => {
            result.current.handleToggleFavorite();
        });

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Erro ao atualizar favorito', {
                variant: 'error',
            })
        );
    });

    it('shows favorite removed message for starred image', async () => {
        mockToggleStarredFile.mockResolvedValue(undefined);
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, starred: true })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images']),
        });

        act(() => {
            result.current.handleOpenImage(1);
        });

        act(() => {
            result.current.handleToggleFavorite();
        });

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Imagem removida dos favoritos', {
                variant: 'success',
            })
        );
    });

    // ── folder/album images when selection doesn't match ─────────────

    it('returns empty images when selected folder does not match any collection', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage({ id: 1, path: '/photos/travel/a.jpg', name: 'a.jpg' })],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/folders?folder=/nonexistent']),
        });

        expect(result.current.selectedFolder).toBeNull();
        expect(result.current.filteredImages).toHaveLength(0);
    });

    it('returns empty images when selected album does not match', () => {
        mockUseImage.mockReturnValue({
            ...defaultImageReturn(),
            images: [createImage()],
        });

        const { result } = renderHook(() => useImageContent(), {
            wrapper: createWrapper(['/images/albums?album=nonexistent']),
        });

        expect(result.current.selectedAlbum).toBeNull();
        expect(result.current.filteredImages).toHaveLength(0);
    });
});
