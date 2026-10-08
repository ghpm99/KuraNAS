import { fireEvent, render, screen } from '@testing-library/react';
import FileContent from './fileContent';
import { createTestFile } from '../selection/testFileFactory';
import { SelectionTestHarness } from '../selection/selectionTestHarness';
import { createFileContextStub } from '../selection/fileContextStub';
import type { FileSearchListing } from './fileContent';

const mockOpenMediaItem = jest.fn();
const mockSentinelProps = jest.fn();

jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({ openMediaItem: (...args: unknown[]) => mockOpenMediaItem(...args) }),
}));
jest.mock('@/components/loadMoreSentinel/loadMoreSentinel', () => ({
    __esModule: true,
    default: (props: unknown) => {
        mockSentinelProps(props);
        return null;
    },
}));

const searchedFiles = [
    createTestFile(1, { name: 'relatorio.txt', parent_path: '/docs/2024' }),
    createTestFile(2, { name: 'relatorio-final.txt', parent_path: '/docs' }),
];

const buildListing = (overrides: Partial<FileSearchListing> = {}): FileSearchListing => ({
    items: searchedFiles,
    status: 'success',
    hasNextPage: true,
    isFetchingNextPage: false,
    fetchNextPage: jest.fn(),
    ...overrides,
});

const renderSearch = (
    listing: FileSearchListing,
    viewMode: 'grid' | 'list' = 'grid',
    contextOverrides = {}
) => {
    const handleSelectItem = jest.fn();
    const fileContext = createFileContextStub({
        status: 'pending',
        files: [],
        handleSelectItem,
        handleStarredItem: jest.fn(),
        fileListFilter: 'all',
        hasNextPage: false,
        isFetchingNextPage: false,
        fetchNextPage: jest.fn(),
        ...contextOverrides,
    });
    render(
        <SelectionTestHarness fileContext={fileContext} seedFiles={[]}>
            <FileContent
                showHeading={false}
                viewMode={viewMode}
                searchListing={listing}
                emptyStateMessage="FILES_SEARCH_EMPTY"
            />
        </SelectionTestHarness>
    );
    return { handleSelectItem };
};

describe('FileContent search listing', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockOpenMediaItem.mockReturnValue(false);
    });

    it('ignores the folder listing status and shows each result with its parent path', () => {
        renderSearch(buildListing());

        expect(screen.getByText('relatorio.txt')).toBeInTheDocument();
        expect(screen.getByText('/docs/2024')).toBeInTheDocument();
        expect(screen.getByText('/docs')).toBeInTheDocument();
    });

    it('shows the parent path in list mode as well', () => {
        renderSearch(buildListing(), 'list');

        expect(screen.getByText('/docs/2024')).toBeInTheDocument();
    });

    it('keeps selection working on results', () => {
        renderSearch(buildListing());

        fireEvent.click(screen.getAllByRole('checkbox')[0]!);

        expect(screen.getByTestId('selected-count').textContent).toBe('1');
    });

    it('opens a result by navigating to it', () => {
        const { handleSelectItem } = renderSearch(buildListing());

        fireEvent.click(screen.getByText('relatorio-final.txt'));

        expect(handleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 2 }));
    });

    it('paginates through the search listing instead of the folder listing', () => {
        const fetchNextPage = jest.fn();
        renderSearch(buildListing({ fetchNextPage, isFetchingNextPage: true }), 'grid', {
            hasNextPage: false,
        });

        expect(mockSentinelProps).toHaveBeenCalledWith({
            hasNextPage: true,
            isFetchingNextPage: true,
            fetchNextPage,
        });
    });

    it('shows the empty state message when nothing matched', () => {
        renderSearch(buildListing({ items: [], hasNextPage: false }));

        expect(screen.getByText('FILES_SEARCH_EMPTY')).toBeInTheDocument();
    });

    it('shows loading and error states of the search itself', () => {
        const { unmount } = render(
            <SelectionTestHarness fileContext={createFileContextStub({ status: 'success' })} seedFiles={[]}>
                <FileContent searchListing={buildListing({ status: 'pending' })} />
            </SelectionTestHarness>
        );
        expect(screen.getByText('LOADING')).toBeInTheDocument();
        unmount();

        render(
            <SelectionTestHarness fileContext={createFileContextStub({ status: 'success' })} seedFiles={[]}>
                <FileContent searchListing={buildListing({ status: 'error' })} />
            </SelectionTestHarness>
        );
        expect(screen.getByText('ERROR_LOADING_FILES')).toBeInTheDocument();
    });

    it('lists results even when a file is currently selected in the explorer', () => {
        renderSearch(buildListing(), 'grid', { selectedItem: createTestFile(9) });

        expect(screen.getByText('relatorio.txt')).toBeInTheDocument();
    });
});
