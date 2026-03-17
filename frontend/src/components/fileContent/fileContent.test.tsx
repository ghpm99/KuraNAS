import { render, screen } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import FileContent from './fileContent';
import type { FileData } from '../providers/fileProvider/fileContext';

const mockUseFile = jest.fn();
const mockOpenMediaItem = jest.fn();

jest.mock('../providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));
jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({
        openMediaItem: (...args: any[]) => mockOpenMediaItem(...args),
    }),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

jest.mock('../fileCard', () => ({ title, metadata, onClick, onClickStar }: any) => (
    <div>
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
        expect(screen.getByText('ERROR_LOADING_FILES')).toBeInTheDocument();
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
            expect.objectContaining({ id: 1, name: 'song' })
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
            expect.objectContaining({ id: 3, name: 'child' })
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
            expect.objectContaining({ id: 1, name: 'movie.mp4' })
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
        expect(screen.getByRole('button', { name: 'song' })).toBeInTheDocument();
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
});
