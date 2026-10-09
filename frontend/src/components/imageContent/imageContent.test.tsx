import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { ImageProvider } from '@/components/providers/imageProvider/imageProvider';
import { apiBase } from '@/service';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import {
    defaultSettingsConfiguration,
    SettingsContextProvider,
} from '@/components/providers/settingsProvider/settingsContext';
import ImageContent from './imageContent';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => '/api/v1',
}));

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useNavigate: () => mockNavigate,
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: jest.fn() }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            key === 'LOCALE' ? 'en-US' : params ? `${key}:${Object.values(params).join(',')}` : key,
    }),
}));

const mockedApiGet = apiBase.get as jest.Mock;
const mockedApiPost = apiBase.post as jest.Mock;

class FakeIntersectionObserver {
    static instances: FakeIntersectionObserver[] = [];
    isDisconnected = false;
    constructor(private readonly callback: (entries: { isIntersecting: boolean }[]) => void) {
        FakeIntersectionObserver.instances.push(this);
    }
    observe = jest.fn();
    unobserve = jest.fn();
    disconnect = () => {
        this.isDisconnected = true;
    };
    trigger = () => this.callback([{ isIntersecting: true }]);
}

const scrollToSentinel = () => {
    const liveObservers = FakeIntersectionObserver.instances.filter(
        (observer) => !observer.isDisconnected
    );
    act(() => liveObservers[liveObservers.length - 1]!.trigger());
};

const mockMatchMedia = (isDesktop: boolean) => {
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            matches: isDesktop,
            media: query,
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            addListener: jest.fn(),
            removeListener: jest.fn(),
            dispatchEvent: jest.fn(),
            onchange: null,
        }),
    });
};

type LibraryPageResponse = {
    items: ImageLibraryItem[];
    next_cursor?: string;
    has_next?: boolean;
    page?: number;
};

const libraryPage = (response: LibraryPageResponse) => ({
    data: { next_cursor: '', has_next: false, page_size: 60, ...response },
});

type FolderRow = { path: string; name: string; image_count: number; cover_file_id: number };
type FolderPage = { items: FolderRow[]; hasNext: boolean };

type InstallApiOptions = {
    total?: number;
    timeline?: unknown[];
    folders?: Record<string, FolderRow[]>;
    folderPages?: FolderPage[];
};

const folderResponse = (items: FolderRow[], page: number, hasNext: boolean) => ({
    data: { items, pagination: { page, page_size: 48, has_next: hasNext, has_prev: page > 1 } },
});

const installApi = (pages: (LibraryPageResponse | Error)[], options: InstallApiOptions = {}) => {
    let pageIndex = 0;
    mockedApiGet.mockImplementation(
        (
            url: string,
            config?: { params?: { page_size?: number; parent?: string; page: number } }
        ) => {
            if (url === '/image/library') {
                if (config?.params?.page_size === 1) {
                    return Promise.resolve(
                        libraryPage({ items: [buildImageLibraryItem({ file_id: 900 })] })
                    );
                }
                const page = pages[Math.min(pageIndex++, pages.length - 1)]!;
                return page instanceof Error
                    ? Promise.reject(page)
                    : Promise.resolve(libraryPage(page));
            }
            if (url === '/image/library/folders') {
                const params = config?.params as { parent?: string; page: number };
                if (options.folderPages) {
                    const folderPage = options.folderPages[params.page - 1]!;
                    return Promise.resolve(
                        folderResponse(folderPage.items, params.page, folderPage.hasNext)
                    );
                }
                return Promise.resolve(
                    folderResponse(options.folders?.[params.parent ?? ''] ?? [], params.page, false)
                );
            }
            if (url === '/image/library/count') {
                return Promise.resolve({ data: { total: options.total ?? 0 } });
            }
            if (url === '/image/library/timeline') {
                return Promise.resolve({ data: options.timeline ?? [] });
            }
            return Promise.reject(new Error(`unexpected GET ${url}`));
        }
    );
};

const libraryParams = () =>
    mockedApiGet.mock.calls
        .filter(([url, config]) => url === '/image/library' && config.params.page_size !== 1)
        .map(([, config]) => config.params);

const folderCalls = () =>
    mockedApiGet.mock.calls
        .filter(([url]) => url === '/image/library/folders')
        .map(([, config]) => config.params as { parent?: string; page: number });

const lastLibraryParams = () => {
    const allParams = libraryParams();
    return allParams[allParams.length - 1];
};

function LocationProbe() {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
}

const settingsValue = {
    settings: defaultSettingsConfiguration,
    isLoading: false,
    isSaving: false,
    hasError: false,
    refresh: jest.fn(),
    saveSettings: jest.fn(),
};

const renderGallery = (route = '/images') =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MemoryRouter initialEntries={[route]}>
                <SettingsContextProvider value={settingsValue}>
                    <ImageProvider>
                        <ImageContent />
                    </ImageProvider>
                </SettingsContextProvider>
                <LocationProbe />
            </MemoryRouter>
        </QueryClientProvider>
    );

const marchImages = [
    buildImageLibraryItem({ file_id: 1, name: 'March-1.jpg', taken_at: '2026-03-20T10:00:00Z' }),
    buildImageLibraryItem({ file_id: 2, name: 'March-2.jpg', taken_at: '2026-03-02T10:00:00Z' }),
    buildImageLibraryItem({ file_id: 3, name: 'Feb-1.jpg', taken_at: '2026-02-11T10:00:00Z' }),
    buildImageLibraryItem({ file_id: 4, name: 'Undated.jpg', taken_at: null }),
];

describe('ImageContent', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        FakeIntersectionObserver.instances = [];
        Object.defineProperty(window, 'IntersectionObserver', {
            configurable: true,
            writable: true,
            value: FakeIntersectionObserver,
        });
        mockMatchMedia(false);
    });

    it('groups images by month with the server total and the undated group last', async () => {
        installApi([{ items: marchImages }], {
            total: 120,
            timeline: [
                { year: 2026, month: 3, count: 80 },
                { year: 2026, month: 2, count: 40 },
            ],
        });

        renderGallery();

        const headings = await screen.findAllByRole('heading', { level: 3 });
        expect(headings.map((heading) => heading.textContent)).toEqual([
            'March 2026',
            'February 2026',
            'IMAGES_GROUP_NO_DATE',
        ]);
        expect(
            screen.getByRole('heading', { level: 2, name: 'IMAGES_SECTION_LIBRARY' })
        ).toBeInTheDocument();
        await waitFor(() => expect(screen.getAllByText('IMAGES_PHOTOS_COUNT:120')).toHaveLength(1));
        expect(screen.getByText('IMAGES_PHOTOS_COUNT:80')).toBeInTheDocument();
        expect(screen.getByText('IMAGES_PHOTOS_COUNT:40')).toBeInTheDocument();
    });

    it('loads the next page from the sentinel placed after the grid', async () => {
        installApi([
            { items: marchImages, has_next: true, next_cursor: 'cursor-1' },
            {
                items: [
                    buildImageLibraryItem({
                        file_id: 5,
                        name: 'Older.jpg',
                        taken_at: '2025-12-01T10:00:00Z',
                    }),
                ],
            },
        ]);

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        scrollToSentinel();

        expect(await screen.findByRole('img', { name: 'Older.jpg' })).toBeInTheDocument();
        expect(libraryParams()[1]).toEqual(expect.objectContaining({ cursor: 'cursor-1' }));
        await waitFor(() => expect(screen.getByText('IMAGES_END_MESSAGE')).toBeInTheDocument());
    });

    it('keeps the sentinel alive when a filter leaves the first page empty but more pages exist', async () => {
        installApi([
            { items: [], has_next: true, next_cursor: 'cursor-1' },
            { items: [buildImageLibraryItem({ file_id: 6, name: 'Late-hit.jpg' })] },
        ]);

        renderGallery('/images?q=late');
        await screen.findByRole('button', { name: 'LOAD_MORE' });

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('img', { name: 'Late-hit.jpg' })).toBeInTheDocument();
    });

    it('shows a backend error with retry and recovers on retry', async () => {
        installApi([
            Object.assign(new Error('failed'), { response: { data: { error: 'Server said no' } } }),
            { items: marchImages },
        ]);

        renderGallery();

        const alert = await screen.findByRole('alert');
        expect(within(alert).getByText('IMAGES_ERROR_TITLE')).toBeInTheDocument();
        expect(within(alert).getByText('Server said no')).toBeInTheDocument();

        fireEvent.click(within(alert).getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByRole('img', { name: 'March-1.jpg' })).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('keeps loaded images and retries the next page when it fails', async () => {
        installApi([
            { items: marchImages, has_next: true, next_cursor: 'cursor-1' },
            new Error('page two failed'),
            { items: [buildImageLibraryItem({ file_id: 5, name: 'Recovered.jpg' })] },
        ]);

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));
        const alert = await screen.findByRole('alert');
        expect(screen.getByRole('img', { name: 'March-1.jpg' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).not.toBeInTheDocument();

        fireEvent.click(within(alert).getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByRole('img', { name: 'Recovered.jpg' })).toBeInTheDocument();
        expect(libraryParams()[2]).toEqual(expect.objectContaining({ cursor: 'cursor-1' }));
    });

    it('tells an empty library apart from a filtered search without results', async () => {
        installApi([{ items: [] }]);

        const empty = renderGallery('/images');
        expect(await screen.findByText('IMAGES_EMPTY_TITLE')).toBeInTheDocument();
        expect(
            screen.queryByRole('button', { name: 'IMAGES_FILTER_CLEAR' })
        ).not.toBeInTheDocument();
        empty.unmount();

        renderGallery('/images?q=nothing');
        expect(await screen.findByText('IMAGES_EMPTY_FILTERED_TITLE')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_FILTER_CLEAR' }));
        await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/images$/));
    });

    it('uses specific empty messages for favorites and for section presets', async () => {
        installApi([{ items: [] }]);

        const favorites = renderGallery('/images/favorites');
        expect(await screen.findByText('IMAGES_EMPTY_FAVORITES_TITLE')).toBeInTheDocument();
        favorites.unmount();

        renderGallery('/images/captures');
        expect(await screen.findByText('IMAGES_EMPTY_FILTERED_TITLE')).toBeInTheDocument();
    });

    it('sends the search term to the server after the debounce and keeps it in the URL', async () => {
        installApi([{ items: marchImages }]);

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        await userEvent.type(screen.getByRole('searchbox'), 'beach');

        await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('?q=beach'));
        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(expect.objectContaining({ q: 'beach' }))
        );
        expect(libraryParams().filter((params) => params.q === 'b')).toHaveLength(0);
    });

    it('applies period, format and sort chips as server filters', async () => {
        installApi([{ items: marchImages }]);

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        fireEvent.click(screen.getByRole('button', { name: 'PNG' }));
        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(expect.objectContaining({ format: ['png'] }))
        );

        const [fromInput] = Array.from(
            document.querySelectorAll<HTMLInputElement>('input[type="date"]')
        );
        fireEvent.change(fromInput!, { target: { value: '2026-01-01' } });
        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(
                expect.objectContaining({ taken_from: '2026-01-01' })
            )
        );

        fireEvent.change(screen.getByRole('combobox'), { target: { value: 'name' } });
        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(
                expect.objectContaining({ sort: 'name', order: 'asc', page: 1 })
            )
        );
        expect(screen.queryByRole('heading', { level: 3 })).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_SORT_ORDER_ASC' }));
        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(
                expect.objectContaining({ sort: 'name', order: 'desc' })
            )
        );
    });

    it('jumps to a month from the phone sheet and goes back to the latest', async () => {
        installApi([{ items: marchImages }], { timeline: [{ year: 2026, month: 3, count: 80 }] });

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        fireEvent.click(await screen.findByRole('button', { name: 'IMAGES_SCRUBBER_OPEN' }));
        fireEvent.click(screen.getByRole('button', { name: /IMAGES_SCRUBBER_MONTH_ARIA/ }));

        await waitFor(() =>
            expect(lastLibraryParams()).toEqual(
                expect.objectContaining({ taken_before: '2026-04-01' })
            )
        );
        expect(screen.getByTestId('location')).toHaveTextContent('before=2026-04-01');

        fireEvent.click(await screen.findByRole('button', { name: 'IMAGES_JUMP_BACK_LATEST' }));

        await waitFor(() =>
            expect(screen.getByTestId('location')).not.toHaveTextContent('before=')
        );
        expect(
            screen.queryByRole('button', { name: 'IMAGES_JUMP_BACK_LATEST' })
        ).not.toBeInTheDocument();
    });

    it('shows the desktop date rail and hides the scrubber for non date orderings', async () => {
        mockMatchMedia(true);
        installApi([{ items: marchImages }], { timeline: [{ year: 2026, month: 3, count: 80 }] });

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        expect(
            await screen.findByRole('navigation', { name: 'IMAGES_SCRUBBER_ARIA' })
        ).toBeInTheDocument();
        fireEvent.change(screen.getByRole('combobox'), { target: { value: 'size' } });

        await waitFor(() =>
            expect(
                screen.queryByRole('navigation', { name: 'IMAGES_SCRUBBER_ARIA' })
            ).not.toBeInTheDocument()
        );
    });

    it('stars a card optimistically and posts the toggle', async () => {
        installApi([{ items: marchImages }]);
        mockedApiPost.mockResolvedValue({});

        renderGallery();
        await screen.findByRole('img', { name: 'March-1.jpg' });

        const starButtons = screen.getAllByRole('button', {
            name: 'IMAGES_STAR_ADD_ARIA:March-1.jpg',
        });
        fireEvent.click(starButtons[0]!);

        await waitFor(() =>
            expect(
                screen.getByRole('button', { name: 'IMAGES_STAR_REMOVE_ARIA:March-1.jpg' })
            ).toHaveAttribute('aria-pressed', 'true')
        );
        expect(mockedApiPost).toHaveBeenCalledWith('/files/starred/1');
    });

    it('lists server preset collections with counts and opens one as a filtered grid', async () => {
        installApi([{ items: [buildImageLibraryItem({ file_id: 31, name: 'Receipt.jpg' })] }], {
            total: 7,
        });

        renderGallery('/images/albums');

        const documentsCard = await screen.findByRole('button', {
            name: 'IMAGES_COLLECTION_OPEN:IMAGES_ALBUM_DOCUMENTS',
        });
        await waitFor(() =>
            expect(within(documentsCard).getByText('IMAGES_PHOTOS_COUNT:7')).toBeInTheDocument()
        );
        expect(screen.queryByRole('searchbox')).not.toBeInTheDocument();
        expect(libraryParams()).toHaveLength(0);

        fireEvent.click(documentsCard);

        expect(await screen.findByRole('img', { name: 'Receipt.jpg' })).toBeInTheDocument();
        expect(libraryParams()[0]).toEqual(
            expect.objectContaining({ category: ['document', 'receipt'] })
        );

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_BACK_TO_ALBUMS' }));
        expect(
            await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:IMAGES_ALBUM_MEMES' })
        ).toBeInTheDocument();
    });

    it('lists server folders with real counts and covers, then browses into a subfolder', async () => {
        installApi([{ items: [buildImageLibraryItem({ file_id: 5, name: 'Inside.jpg' })] }], {
            total: 1,
            folders: {
                '': [{ path: '/photos', name: 'photos', image_count: 12, cover_file_id: 77 }],
                '/photos': [
                    { path: '/photos/trip', name: 'trip', image_count: 4, cover_file_id: 78 },
                ],
            },
        });

        renderGallery('/images/folders');

        const photosCard = await screen.findByRole('button', {
            name: 'IMAGES_COLLECTION_OPEN:photos',
        });
        expect(within(photosCard).getByText('IMAGES_PHOTOS_COUNT:12')).toBeInTheDocument();
        expect(within(photosCard).getByRole('img', { name: 'photos' })).toHaveAttribute(
            'src',
            expect.stringContaining('/files/thumbnail/77')
        );
        expect(libraryParams()).toHaveLength(0);

        fireEvent.click(photosCard);

        expect(await screen.findByRole('img', { name: 'Inside.jpg' })).toBeInTheDocument();
        expect(
            await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:trip' })
        ).toBeInTheDocument();
        expect(lastLibraryParams()).toEqual(expect.objectContaining({ folder: '/photos' }));
        expect(folderCalls()).toContainEqual(
            expect.objectContaining({ parent: '/photos', page: 1 })
        );
        expect(screen.getByTestId('location')).toHaveTextContent('folder=%2Fphotos');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_SECTION_FOLDERS' }));
        await waitFor(() =>
            expect(screen.getByTestId('location')).not.toHaveTextContent('folder=')
        );
        expect(
            await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:photos' })
        ).toBeInTheDocument();
    });

    it('loads more folders when the folder list sentinel is reached', async () => {
        installApi([], {
            folderPages: [
                {
                    items: [{ path: '/a', name: 'a', image_count: 1, cover_file_id: 1 }],
                    hasNext: true,
                },
                {
                    items: [{ path: '/b', name: 'b', image_count: 2, cover_file_id: 2 }],
                    hasNext: false,
                },
            ],
        });

        renderGallery('/images/folders');
        await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:a' });

        scrollToSentinel();

        expect(
            await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:b' })
        ).toBeInTheDocument();
        expect(folderCalls().map((params) => params.page)).toEqual([1, 2]);
    });

    it('shows the empty state when the server has no folders with images', async () => {
        installApi([], { folders: { '': [] } });

        renderGallery('/images/folders');

        expect(await screen.findByText('IMAGES_FOLDERS_EMPTY_TITLE')).toBeInTheDocument();
    });

    it('shows a retryable error when the folder list fails', async () => {
        mockedApiGet.mockRejectedValue(new Error('down'));

        renderGallery('/images/folders');

        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('opens the viewer from a card, navigates and closes it', async () => {
        installApi([{ items: marchImages }]);

        renderGallery();
        fireEvent.click(
            await screen.findByRole('button', { name: 'IMAGES_OPEN_IMAGE_ARIA:March-2.jpg' })
        );

        const dialog = await screen.findByRole('dialog', { name: 'March-2.jpg' });
        expect(screen.getByTestId('location')).toHaveTextContent('image=2');

        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' }));
        expect(await screen.findByRole('dialog', { name: 'Feb-1.jpg' })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_CLOSE_VIEWER' }));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(screen.getByTestId('location')).not.toHaveTextContent('image=');
    });

    it('opens the viewer from the image deep link and toggles its favorite state', async () => {
        installApi([{ items: marchImages }]);
        mockedApiPost.mockResolvedValue({});

        renderGallery('/images?image=3');

        const dialog = await screen.findByRole('dialog', { name: 'Feb-1.jpg' });
        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_VIEWER_ADD_FAVORITE' }));

        await waitFor(() => expect(mockedApiPost).toHaveBeenCalledWith('/files/starred/3'));
        await waitFor(() =>
            expect(
                within(screen.getByRole('dialog')).getByRole('button', {
                    name: 'IMAGES_VIEWER_REMOVE_FAVORITE',
                })
            ).toBeInTheDocument()
        );
    });

    it('resolves a deep-linked image that is not in the loaded pages through its path', async () => {
        installApi([{ items: marchImages }]);
        const deepLinked = {
            id: 99,
            name: 'Far-away.jpg',
            path: '/old/Far-away.jpg',
            parent_path: '/old',
            type: 2,
            format: '.jpg',
            size: 5,
            updated_at: '',
            created_at: '',
            deleted_at: '',
            last_interaction: '',
            last_backup: '',
            check_sum: '',
            directory_content_count: 0,
            starred: false,
        };
        const defaultImplementation = mockedApiGet.getMockImplementation()!;
        mockedApiGet.mockImplementation((url: string, config?: unknown) =>
            url === '/files/path'
                ? Promise.resolve({ data: { items: [deepLinked] } })
                : defaultImplementation(url, config)
        );

        renderGallery('/images?image=99&imagePath=%2Fold%2FFar-away.jpg');

        expect(await screen.findByRole('dialog', { name: 'Far-away.jpg' })).toBeInTheDocument();
        expect(mockedApiGet).toHaveBeenCalledWith('/files/path', {
            params: { path: '/old/Far-away.jpg' },
        });
    });

    it('fetches the next gallery page when next is pressed on the last loaded image', async () => {
        const firstPage = marchImages.slice(0, 2);
        const secondPage = marchImages.slice(2, 3);
        installApi([
            { items: firstPage, has_next: true, next_cursor: 'cursor-1' },
            { items: secondPage },
        ]);

        renderGallery();
        fireEvent.click(
            await screen.findByRole('button', { name: 'IMAGES_OPEN_IMAGE_ARIA:March-2.jpg' })
        );
        const dialog = await screen.findByRole('dialog', { name: 'March-2.jpg' });
        expect(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' })).toBeEnabled();

        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' }));

        expect(await screen.findByRole('dialog', { name: 'Feb-1.jpg' })).toBeInTheDocument();
        expect(libraryParams()).toHaveLength(2);
        expect(lastLibraryParams().cursor).toBe('cursor-1');
        expect(screen.getByRole('button', { name: 'IMAGES_NEXT' })).toBeDisabled();

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_PREVIOUS' }));
        expect(await screen.findByRole('dialog', { name: 'March-2.jpg' })).toBeInTheDocument();
    });

    it('does not wrap around at the true end or the true start of the gallery', async () => {
        installApi([{ items: marchImages.slice(0, 2) }]);

        renderGallery('/images?image=1');
        const dialog = await screen.findByRole('dialog', { name: 'March-1.jpg' });
        expect(within(dialog).getByRole('button', { name: 'IMAGES_PREVIOUS' })).toBeDisabled();

        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' }));
        expect(await screen.findByRole('dialog', { name: 'March-2.jpg' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'IMAGES_NEXT' })).toBeDisabled();
        expect(libraryParams()).toHaveLength(1);
    });

    it('navigates around a deep-linked image that is not loaded using the neighbors endpoint', async () => {
        installApi([{ items: marchImages }]);
        const deepLinked = {
            id: 99,
            name: 'Far-away.jpg',
            path: '/old/Far-away.jpg',
            parent_path: '/old',
            type: 2,
            format: '.jpg',
            size: 5,
            updated_at: '',
            created_at: '',
            deleted_at: '',
            last_interaction: '',
            last_backup: '',
            check_sum: '',
            directory_content_count: 0,
            starred: false,
        };
        const neighbors = {
            before: [buildImageLibraryItem({ file_id: 100, name: 'Newer.jpg' })],
            after: [buildImageLibraryItem({ file_id: 98, name: 'Older.jpg' })],
        };
        const defaultImplementation = mockedApiGet.getMockImplementation()!;
        mockedApiGet.mockImplementation((url: string, config?: unknown) => {
            if (url === '/files/path') {
                return Promise.resolve({ data: { items: [deepLinked] } });
            }
            if (url === '/image/library/neighbors/99') {
                return Promise.resolve({ data: neighbors });
            }
            return defaultImplementation(url, config);
        });

        renderGallery('/images?image=99&imagePath=%2Fold%2FFar-away.jpg');

        const dialog = await screen.findByRole('dialog', { name: 'Far-away.jpg' });
        await waitFor(() =>
            expect(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' })).toBeEnabled()
        );
        expect(mockedApiGet).toHaveBeenCalledWith(
            '/image/library/neighbors/99',
            expect.objectContaining({ params: expect.objectContaining({ count: 20 }) })
        );

        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_NEXT' }));
        expect(await screen.findByRole('dialog', { name: 'Older.jpg' })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_PREVIOUS' }));
        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_PREVIOUS' }));
        expect(await screen.findByRole('dialog', { name: 'Newer.jpg' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'IMAGES_PREVIOUS' })).toBeDisabled();
    });

    it('opens the folder of the viewed image in the files page', async () => {
        installApi([{ items: marchImages }]);

        renderGallery('/images?image=1');
        fireEvent.click(await screen.findByRole('button', { name: 'IMAGES_VIEWER_OPEN_FOLDER' }));

        expect(mockNavigate).toHaveBeenCalledWith({
            pathname: '/files',
            search: '?path=%2Fphotos%2Ftravel',
        });
    });
});
