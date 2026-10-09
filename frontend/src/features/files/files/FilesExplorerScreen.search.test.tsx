import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
import FilesExplorerScreen from './FilesExplorerScreen';

const mockUseFile = jest.fn();
const mockSearchFiles = jest.fn();

jest.mock('@/service/files', () => ({
    searchFiles: (...args: unknown[]) => mockSearchFiles(...args),
}));

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

jest.mock('@/features/files/providers/fileProvider/useFileAncestors', () => ({
    __esModule: true,
    default: () => ({ data: undefined }),
}));

jest.mock('@/components/actionBar', () => () => <div>ActionBarMock</div>);
jest.mock('@/features/files/fileContent', () => ({ searchListing, emptyStateMessage }: any) => (
    <div data-testid="file-content">
        {searchListing
            ? searchListing.items.map((item: any) => <span key={item.id}>{item.name}</span>)
            : 'FolderListing'}
        {searchListing?.items.length === 0 ? emptyStateMessage : null}
    </div>
));
jest.mock('@/features/files/fileDetails', () => () => <div>FileDetailsMock</div>);
jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => <div>TreeMock</div>);
jest.mock('@/components/tabs', () => () => <div>TabsMock</div>);
jest.mock('@/features/files/findByDiskPath/findByDiskPathDialog', () => () => null);

const folder = {
    id: 7,
    name: 'docs',
    path: '/docs',
    parent_path: '/',
    type: 1,
    file_children: [],
};

const LocationProbe = () => <span data-testid="location">{useLocation().search}</span>;

const renderScreen = (initialUrl: string) =>
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <MemoryRouter initialEntries={[initialUrl]}>
                <FilesExplorerScreen />
                <LocationProbe />
            </MemoryRouter>
        </QueryClientProvider>
    );

describe('FilesExplorerScreen search', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue({
            files: [],
            selectedItem: folder,
            handleSelectItem: jest.fn(),
            fileListFilter: 'all',
            filesSort: { key: 'name', order: 'asc' },
            setFilesSort: jest.fn(),
        });
        mockSearchFiles.mockResolvedValue({
            items: [{ id: 1, name: 'relatorio.txt' }],
            pagination: { page: 1, pageSize: 100, hasNext: false },
        });
    });

    it('shows the normal listing and never searches without q', () => {
        renderScreen('/files/docs');

        expect(screen.getByText('FolderListing')).toBeInTheDocument();
        expect(screen.getByText('TabsMock')).toBeInTheDocument();
        expect(screen.queryByText(/FILES_SEARCH_RESULTS/)).toBeNull();
        expect(mockSearchFiles).not.toHaveBeenCalled();
    });

    it('searches the open folder recursively when q is in the url', async () => {
        renderScreen('/files/docs?q=relatorio');

        expect(await screen.findByText('relatorio.txt')).toBeInTheDocument();
        expect(mockSearchFiles).toHaveBeenCalledWith({
            q: 'relatorio',
            parentId: 7,
            recursive: true,
            page: 1,
            pageSize: 100,
            refinements: expect.objectContaining({ kinds: undefined, tier: undefined }),
        });
        expect(
            screen.getByText('FILES_SEARCH_RESULTS_ONE:{"query":"relatorio"}')
        ).toBeInTheDocument();
        expect(screen.queryByText('TabsMock')).toBeNull();
        expect(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' })).toBeInTheDocument();
    });

    it('sends the filters kept in the url with the search', async () => {
        renderScreen('/files/docs?q=relatorio&kind=image&tier=cold&starred=true&sort=size');

        await screen.findByText('relatorio.txt');

        expect(mockSearchFiles).toHaveBeenCalledWith(
            expect.objectContaining({
                refinements: expect.objectContaining({
                    kinds: ['image'],
                    tier: 'cold',
                    starred: true,
                    sort: 'size',
                }),
            })
        );
    });

    it('hides the filter bar when there is no search', async () => {
        renderScreen('/files/docs');

        expect(await screen.findByText('TabsMock')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' })).toBeNull();
    });

    it('searches only direct children after unticking include subfolders', async () => {
        renderScreen('/files/docs?q=relatorio');
        await screen.findByText('relatorio.txt');

        fireEvent.click(screen.getByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS'));

        await waitFor(() =>
            expect(mockSearchFiles).toHaveBeenLastCalledWith(
                expect.objectContaining({ parentId: 7, recursive: false })
            )
        );
    });

    it('searches globally when no folder is open', async () => {
        mockUseFile.mockReturnValue({ ...mockUseFile(), selectedItem: null });
        renderScreen('/files?q=foto');

        await screen.findByText('relatorio.txt');

        expect(mockSearchFiles).toHaveBeenCalledWith(
            expect.objectContaining({ q: 'foto', parentId: undefined })
        );
        expect(screen.queryByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS')).toBeNull();
    });

    it('shows the empty state when nothing matched', async () => {
        mockSearchFiles.mockResolvedValue({
            items: [],
            pagination: { page: 1, pageSize: 100, hasNext: false },
        });
        renderScreen('/files/docs?q=zzz');

        expect(await screen.findByText('FILES_SEARCH_EMPTY')).toBeInTheDocument();
    });

    it('writes the typed term to the url after the debounce and clears it', async () => {
        renderScreen('/files/docs');

        fireEvent.change(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER'), {
            target: { value: 'relatorio' },
        });
        expect(screen.getByTestId('location').textContent).toBe('');

        await waitFor(() => expect(screen.getByTestId('location').textContent).toBe('?q=relatorio'));
        expect(await screen.findByText('relatorio.txt')).toBeInTheDocument();

        fireEvent.click(screen.getAllByRole('button', { name: 'FILES_SEARCH_CLEAR' })[0]!);

        expect(screen.getByTestId('location').textContent).toBe('');
        expect(screen.getByText('FolderListing')).toBeInTheDocument();
    });

    it('does not search for a single character', () => {
        renderScreen('/files/docs?q=a');

        expect(screen.getByText('FolderListing')).toBeInTheDocument();
        expect(mockSearchFiles).not.toHaveBeenCalled();
    });
});
