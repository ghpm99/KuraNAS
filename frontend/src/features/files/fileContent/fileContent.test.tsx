import { render, screen } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import FileContent from './fileContent';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

const mockUseFile = jest.fn();
const mockOpenMediaItem = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));
jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({
        openMediaItem: (...args: any[]) => mockOpenMediaItem(...args),
    }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

jest.mock('../fileCard', () => ({ title, metadata, onClick, onClickStar, isCold }: any) => (
    <div data-cold={String(Boolean(isCold))} data-testid={`card-${title}`}>
        <button onClick={onClick}>{title}</button>
        <button onClick={onClickStar}>star-{title}</button>
        <span>{metadata}</span>
    </div>
));
jest.mock('./components/fileViewer/fileViewer', () => ({ file }: any) => (
    <div>viewer:{file.name}</div>
));

const createFile = (overrides: Partial<FileData> = {}): FileData => ({
    id: 1,
    name: 'file.txt',
    path: '/library/file.txt',
    parent_path: '/library',
    type: 2,
    format: '.txt',
    size: 1024,
    updated_at: '2026-03-10T10:00:00Z',
    created_at: '2026-03-10T10:00:00Z',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    ...overrides,
});

describe('fileContent', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockOpenMediaItem.mockReturnValue(false);
    });

    it('renders pending and error states', () => {
        mockUseFile.mockReturnValue({
            status: 'pending',
            selectedItem: null,
            files: [],
        });
        render(<FileContent />);
        expect(screen.getByText('LOADING')).toBeInTheDocument();

        mockUseFile.mockReturnValue({
            status: 'error',
            selectedItem: null,
            files: [],
        });
        render(<FileContent />);
        expect(screen.getByText('FILES_LISTING_ERROR_TITLE')).toBeInTheDocument();
    });

    it('renders root files, directory and file preview branches', () => {
        const rootHandleSelectItem = jest.fn();
        const rootHandleStarredItem = jest.fn();
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: rootHandleSelectItem,
            handleStarredItem: rootHandleStarredItem,
            selectedItem: null,
            files: [
                createFile({ id: 1, name: 'song', format: '.mp3' }),
                createFile({
                    id: 4,
                    name: 'docs',
                    path: '/library/docs',
                    type: 1,
                    format: '',
                    directory_content_count: 1,
                }),
            ],
        });
        render(<FileContent />);
        expect(screen.getByText('FILES')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'song' })).toBeInTheDocument();
        expect(screen.getByText(/FOLDER - 1 ITEM/)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'song' }));
        expect(mockOpenMediaItem).toHaveBeenCalledWith(
            expect.objectContaining({ id: 1, name: 'song' }),
            expect.arrayContaining([
                expect.objectContaining({ id: 1 }),
                expect.objectContaining({ id: 4 }),
            ])
        );
        expect(rootHandleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }));
        fireEvent.click(screen.getByRole('button', { name: 'star-song' }));
        expect(rootHandleStarredItem).toHaveBeenCalledWith(1);

        const handleSelectItem = jest.fn();
        const handleStarredItem = jest.fn();
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem,
            handleStarredItem,
            selectedItem: {
                ...createFile({
                    id: 2,
                    name: 'folder',
                    path: '/library/folder',
                    type: 1,
                    format: '',
                }),
                file_children: [
                    createFile({
                        id: 3,
                        name: 'child',
                        path: '/library/folder/child.txt',
                        parent_path: '/library/folder',
                        size: 50,
                        starred: true,
                    }),
                    createFile({
                        id: 5,
                        name: 'subfolder',
                        path: '/library/folder/subfolder',
                        parent_path: '/library/folder',
                        type: 1,
                        format: '',
                        directory_content_count: 3,
                    }),
                ],
            },
            files: [],
        });
        render(<FileContent />);
        expect(screen.getByText('folder')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'child' })).toBeInTheDocument();
        expect(screen.getByText(/FOLDER - 3 ITENS/)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'child' }));
        expect(mockOpenMediaItem).toHaveBeenCalledWith(
            expect.objectContaining({ id: 3, name: 'child' }),
            expect.arrayContaining([expect.objectContaining({ id: 3 })])
        );
        expect(handleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 3 }));
        fireEvent.click(screen.getByRole('button', { name: 'star-child' }));
        expect(handleStarredItem).toHaveBeenCalledWith(3);

        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: createFile({
                id: 9,
                name: 'report.pdf',
                path: '/library/report.pdf',
                format: '.pdf',
                size: 500,
            }),
            files: [],
        });
        render(<FileContent />);
        expect(screen.getByText('viewer:report.pdf')).toBeInTheDocument();
    });

    it('does not reselect files handled by the shared media opener', () => {
        const handleSelectItem = jest.fn();
        mockOpenMediaItem.mockReturnValue(true);
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem,
            handleStarredItem: jest.fn(),
            selectedItem: null,
            fileListFilter: 'all',
            files: [
                createFile({
                    id: 1,
                    name: 'movie.mp4',
                    path: '/library/movie.mp4',
                    format: '.mp4',
                }),
            ],
        });

        render(<FileContent />);
        fireEvent.click(screen.getByRole('button', { name: 'movie.mp4' }));

        expect(mockOpenMediaItem).toHaveBeenCalledWith(
            expect.objectContaining({ id: 1, name: 'movie.mp4' }),
            [expect.objectContaining({ id: 1 })]
        );
        expect(handleSelectItem).not.toHaveBeenCalled();
    });

    it('supports list view without duplicated heading', () => {
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            fileListFilter: 'all',
            files: [
                createFile({
                    id: 1,
                    name: 'song',
                    path: '/library/song.mp3',
                    format: '.mp3',
                }),
            ],
        });

        render(<FileContent viewMode="list" showHeading={false} />);
        expect(screen.queryByText('FILES')).not.toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'song' })).toHaveAttribute(
            'href',
            '/files/library/song.mp3'
        );
        expect(screen.getByRole('columnheader', { name: 'NAME' })).toBeInTheDocument();
    });

    it('sorts through the header columns of the list view', () => {
        const setFilesSort = jest.fn();
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            fileListFilter: 'all',
            files: [createFile({ id: 1, name: 'song' })],
            filesSort: { key: 'name', order: 'asc' },
            setFilesSort,
        });

        render(<FileContent viewMode="list" showHeading={false} />);
        fireEvent.click(screen.getByRole('button', { name: 'NAME' }));
        fireEvent.click(screen.getByRole('button', { name: 'MODIFIED' }));

        expect(setFilesSort).toHaveBeenNthCalledWith(1, { key: 'name', order: 'desc' });
        expect(setFilesSort).toHaveBeenNthCalledWith(2, { key: 'updated_at', order: 'asc' });
    });

    it('flags cold files in the grid cards and shows the indicator in list rows', () => {
        const baseContext = {
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            fileListFilter: 'all',
            files: [
                createFile({ id: 1, name: 'cold.txt', tier: 'cold' }),
                createFile({ id: 2, name: 'hot.txt', tier: 'hot' }),
                createFile({ id: 3, name: 'legacy.txt' }),
            ],
        };
        mockUseFile.mockReturnValue(baseContext);

        const grid = render(<FileContent />);
        expect(screen.getByTestId('card-cold.txt')).toHaveAttribute('data-cold', 'true');
        expect(screen.getByTestId('card-hot.txt')).toHaveAttribute('data-cold', 'false');
        expect(screen.getByTestId('card-legacy.txt')).toHaveAttribute('data-cold', 'false');
        grid.unmount();

        render(<FileContent viewMode="list" />);
        expect(screen.getAllByRole('img', { name: 'FILE_TIER_COLD_INDICATOR' })).toHaveLength(1);
    });

    it('supports custom collection data and empty state messages', () => {
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            fileListFilter: 'starred',
            files: [],
        });

        const { rerender } = render(
            <FileContent
                title="Favorites scope"
                items={[
                    createFile({
                        id: 7,
                        name: 'notes.txt',
                        path: '/library/notes.txt',
                        format: '.txt',
                        size: 12,
                        starred: true,
                    }),
                ]}
                emptyStateMessage="EMPTY_FAVORITES"
            />
        );

        expect(screen.getByText('Favorites scope')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'notes.txt' })).toBeInTheDocument();

        rerender(
            <FileContent title="Favorites scope" items={[]} emptyStateMessage="EMPTY_FAVORITES" />
        );
        expect(screen.getByText('EMPTY_FAVORITES')).toBeInTheDocument();
    });

    it('offers a load more button that fetches the next page while pages remain', () => {
        const fetchNextPage = jest.fn();
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            files: [createFile({ id: 1, name: 'song' })],
            fetchNextPage,
            hasNextPage: true,
            isFetchingNextPage: false,
        });
        render(<FileContent />);

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        expect(fetchNextPage).toHaveBeenCalledTimes(1);
    });

    it('hides the load more button when there is no next page', () => {
        mockUseFile.mockReturnValue({
            status: 'success',
            handleSelectItem: jest.fn(),
            handleStarredItem: jest.fn(),
            selectedItem: null,
            files: [createFile({ id: 1, name: 'song' })],
            fetchNextPage: jest.fn(),
            hasNextPage: false,
            isFetchingNextPage: false,
        });
        render(<FileContent />);

        expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).not.toBeInTheDocument();
    });
});
