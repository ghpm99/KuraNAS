import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import FileProvider from '@/features/files/providers/fileProvider';
import FileListFilterSync from '@/features/files/files/FileListFilterSync';
import FavoritesScreen from './FavoritesScreen';

const mockGetStarredFiles = jest.fn();
const mockGetFilesTree = jest.fn();
const mockGetFileByPath = jest.fn();
const mockGetFileAncestors = jest.fn();
const mockSetFilesStarred = jest.fn();

jest.mock('@/service/files', () => ({
    ...jest.requireActual('@/service/files'),
    getStarredFiles: (...args: unknown[]) => mockGetStarredFiles(...args),
    getFilesTree: (...args: unknown[]) => mockGetFilesTree(...args),
    getFileByPath: (...args: unknown[]) => mockGetFileByPath(...args),
    getFileAncestors: (...args: unknown[]) => mockGetFileAncestors(...args),
    setFilesStarred: (...args: unknown[]) => mockSetFilesStarred(...args),
    getRecentAccessByFileId: jest.fn().mockResolvedValue([]),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/features/files/fileDetails', () => ({ file }: { file: FileData }) => (
    <div>FileDetailsMock:{file.name}</div>
));

const buildEntry = (id: number, name: string, path: string, overrides = {}): FileData => ({
    id,
    name,
    path,
    parent_path: path.slice(0, path.lastIndexOf('/')) || '/',
    type: 2,
    format: '.txt',
    size: 1,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    ...overrides,
});

const starredFolder = buildEntry(10, 'Projects', '/Projects', {
    type: 1,
    starred: true,
    directory_content_count: 2,
});
const plainFile = buildEntry(11, 'plain-notes.txt', '/Projects/plain-notes.txt');
const plainFolder = buildEntry(12, 'plain-folder', '/Projects/plain-folder', { type: 1 });

const toPage = (items: FileData[]) => ({
    items,
    pagination: { hasNext: false, hasPrevious: false, page: 1, pageSize: 200 },
});

const LocationProbe = () => <div data-testid="location">{useLocation().pathname}</div>;

const renderFavorites = (initialPath: string) =>
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <SettingsProvider>
                <GlobalMusicProvider>
                    <MemoryRouter initialEntries={[initialPath]}>
                        <FileProvider>
                            <FileListFilterSync filter="starred" />
                            <Routes>
                                <Route path="/favorites/*" element={<FavoritesScreen />} />
                                <Route path="*" element={<div>outside favorites</div>} />
                            </Routes>
                            <LocationProbe />
                        </FileProvider>
                    </MemoryRouter>
                </GlobalMusicProvider>
            </SettingsProvider>
        </QueryClientProvider>
    );

const selectItemAt = async (position: number) => {
    const selectCheckboxes = await screen.findAllByLabelText('FILES_SELECT_ITEM');
    const checkbox: HTMLElement | undefined = selectCheckboxes[position];
    if (!checkbox) throw new Error(`no selectable item at ${position}`);
    fireEvent.click(checkbox);
};

describe('Favorites folder navigation', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockGetStarredFiles.mockResolvedValue(toPage([starredFolder]));
        mockGetFilesTree.mockImplementation(async ({ fileParent }: { fileParent?: number }) =>
            toPage(fileParent === 10 ? [plainFile, plainFolder] : [])
        );
        mockGetFileByPath.mockImplementation(async (path: string) =>
            path === '/Projects' ? starredFolder : path === plainFile.path ? plainFile : null
        );
        mockGetFileAncestors.mockResolvedValue([]);
        mockSetFilesStarred.mockImplementation(async (ids: number[], isStarred: boolean) => ({
            starred: isStarred,
            updated_ids: ids,
            not_found_ids: [],
        }));
    });

    it('opens a starred folder inside favorites and lists all its children', async () => {
        renderFavorites('/favorites');

        fireEvent.click(await screen.findByText('Projects'));

        await waitFor(() =>
            expect(screen.getByTestId('location')).toHaveTextContent('/favorites/Projects')
        );
        expect(await screen.findByText('plain-notes.txt')).toBeInTheDocument();
        expect(screen.getByText('plain-folder')).toBeInTheDocument();
        expect(mockGetFilesTree).toHaveBeenCalledWith(
            expect.objectContaining({ fileParent: 10, category: 'all' })
        );
    });

    it('returns to the starred list through the Favoritos breadcrumb', async () => {
        renderFavorites('/favorites/Projects');

        expect(await screen.findByText('plain-notes.txt')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'STARRED_FILES' }));

        await waitFor(() =>
            expect(screen.getByTestId('location')).toHaveTextContent(/^\/favorites$/)
        );
        expect(await screen.findByText('Projects')).toBeInTheDocument();
        expect(screen.queryByText('plain-notes.txt')).not.toBeInTheDocument();
    });

    it('opens the preview panel for a file without leaving favorites', async () => {
        renderFavorites('/favorites/Projects');

        fireEvent.click(await screen.findByText('plain-notes.txt'));

        expect(await screen.findByText('FileDetailsMock:plain-notes.txt')).toBeInTheDocument();
        expect(screen.getByTestId('location')).toHaveTextContent(
            '/favorites/Projects/plain-notes.txt'
        );
        expect(screen.queryByText('outside favorites')).not.toBeInTheDocument();
    });

    it('renders the folder from a direct favorites deep link', async () => {
        renderFavorites('/favorites/Projects');

        expect(await screen.findByText('plain-folder')).toBeInTheDocument();
        expect(mockGetFileByPath).toHaveBeenCalledWith('/Projects');
    });

    it('keeps item links inside favorites for new-tab opening', async () => {
        renderFavorites('/favorites/Projects');

        expect(await screen.findByRole('link', { name: /plain-notes\.txt/ })).toHaveAttribute(
            'href',
            '/favorites/Projects/plain-notes.txt'
        );
        expect(screen.getByRole('link', { name: /plain-folder/ })).toHaveAttribute(
            'href',
            '/favorites/Projects/plain-folder'
        );
    });

    it('keeps starred folder links inside favorites at the root listing', async () => {
        renderFavorites('/favorites');

        expect(await screen.findByRole('link', { name: /Projects/ })).toHaveAttribute(
            'href',
            '/favorites/Projects'
        );
    });

    describe('selection', () => {
        const secondStarredFolder = buildEntry(13, 'Archive', '/Archive', {
            type: 1,
            starred: true,
        });

        it('shows the selection toolbar after selecting two starred items', async () => {
            mockGetStarredFiles.mockResolvedValue(toPage([starredFolder, secondStarredFolder]));
            renderFavorites('/favorites');

            await selectItemAt(0);
            await selectItemAt(1);
            expect(screen.getByRole('toolbar', { name: 'FILES_SELECTION_TOOLBAR' })).toBeInTheDocument();
            expect(screen.getByText('FILES_SELECTION_COUNT')).toBeInTheDocument();
        });

        it('unfavorites the selection in one call and removes the items from the list', async () => {
            mockGetStarredFiles.mockResolvedValue(toPage([starredFolder, secondStarredFolder]));
            mockSetFilesStarred.mockImplementationOnce(async (ids: number[]) => {
                mockGetStarredFiles.mockResolvedValue(toPage([]));
                return { starred: false, updated_ids: ids, not_found_ids: [] };
            });
            renderFavorites('/favorites');

            await selectItemAt(0);
            await selectItemAt(1);
            fireEvent.click(screen.getByRole('button', { name: 'FILES_UNFAVORITE' }));

            await waitFor(() =>
                expect(mockSetFilesStarred).toHaveBeenCalledWith(
                    expect.arrayContaining([10, 13]),
                    false
                )
            );
            await waitFor(() => expect(screen.queryByText('Projects')).not.toBeInTheDocument());
            expect(screen.queryByText('Archive')).not.toBeInTheDocument();
            expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
        });

        it('selects every listed item with the select-all shortcut', async () => {
            mockGetStarredFiles.mockResolvedValue(toPage([starredFolder, secondStarredFolder]));
            renderFavorites('/favorites');
            await screen.findByText('Archive');

            fireEvent.keyDown(document, { key: 'a', ctrlKey: true });

            expect(await screen.findByText('FILES_SELECTION_COUNT')).toBeInTheDocument();
            expect(screen.getAllByRole('checkbox', { checked: true })).toHaveLength(2);
        });

        it('selects items inside a starred folder', async () => {
            renderFavorites('/favorites/Projects');

            await selectItemAt(0);

            expect(screen.getByRole('toolbar', { name: 'FILES_SELECTION_TOOLBAR' })).toBeInTheDocument();
        });
    });
});
