import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import FileContent from './fileContent';
import { createTestFile } from '../selection/testFileFactory';
import { SelectionTestHarness } from '../selection/selectionTestHarness';
import { createFileContextStub } from '../selection/fileContextStub';
import type { FileSearchListing } from './fileContent';

jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({ openMediaItem: () => false }),
}));
jest.mock('@/components/loadMoreSentinel/loadMoreSentinel', () => ({
    __esModule: true,
    default: () => null,
}));

const emptyFolder = createTestFile(10, { name: 'Empty', type: 1, file_children: [] });

const renderContent = (
    contextOverrides = {},
    contentProps: Partial<React.ComponentProps<typeof FileContent>> = {}
) => {
    const fileContext = createFileContextStub({
        status: 'success',
        files: [],
        handleSelectItem: jest.fn(),
        handleStarredItem: jest.fn(),
        fileListFilter: 'all',
        hasNextPage: false,
        isFetchingNextPage: false,
        fetchNextPage: jest.fn(),
        retryListing: jest.fn(),
        ...contextOverrides,
    });
    render(
        <SelectionTestHarness fileContext={fileContext} seedFiles={[]}>
            <FileContent showHeading={false} {...contentProps} />
        </SelectionTestHarness>
    );
    return fileContext;
};

describe('FileContent states', () => {
    it('renders a skeleton placeholder in grid mode while the first page loads', () => {
        renderContent({ status: 'pending' });

        expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
        expect(screen.getByText('LOADING')).toBeInTheDocument();
    });

    it('renders a skeleton placeholder in list mode while the first page loads', () => {
        renderContent({ status: 'pending' }, { viewMode: 'list' });

        expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
    });

    it('shows the backend error message verbatim with a retry button', () => {
        const retryListing = jest.fn();
        renderContent({
            status: 'error',
            listingErrorMessage: 'Disco indisponível',
            retryListing,
        });

        expect(screen.getByRole('alert')).toHaveTextContent('FILES_LISTING_ERROR_TITLE');
        expect(screen.getByText('Disco indisponível')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));
        expect(retryListing).toHaveBeenCalledTimes(1);
    });

    it('shows only the generic title when the backend sent no message', () => {
        renderContent({ status: 'error', listingErrorMessage: undefined });

        expect(screen.getByRole('alert')).toHaveTextContent('FILES_LISTING_ERROR_TITLE');
    });

    it('uses the search listing error message and retry', () => {
        const retry = jest.fn();
        const listing: FileSearchListing = {
            items: [],
            status: 'error',
            hasNextPage: false,
            isFetchingNextPage: false,
            fetchNextPage: jest.fn(),
            errorMessage: 'Busca falhou',
            retry,
        };
        renderContent({}, { searchListing: listing });

        expect(screen.getByText('Busca falhou')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('shows the empty folder state with upload and create-folder actions', async () => {
        const createFolder = jest.fn().mockResolvedValue(undefined);
        const clickSpy = jest.spyOn(HTMLInputElement.prototype, 'click').mockImplementation(() => undefined);
        renderContent({ selectedItem: emptyFolder, createFolder });

        expect(screen.getByText('FILES_EMPTY_FOLDER_TITLE')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'UPLOAD_FILE' }));
        expect(clickSpy).toHaveBeenCalledTimes(1);
        clickSpy.mockRestore();

        fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: 'Docs' } });
        fireEvent.click(screen.getAllByRole('button', { name: 'NEW_FOLDER' }).pop()!);

        await waitFor(() => expect(createFolder).toHaveBeenCalledWith('Docs', 10));
    });

    it('shows the empty favorites state', () => {
        renderContent({ fileListFilter: 'starred' });

        expect(screen.getByText('FILES_EMPTY_FAVORITES_TITLE')).toBeInTheDocument();
    });

    it('shows the empty recent state', () => {
        renderContent({ fileListFilter: 'recent' });

        expect(screen.getByText('FILES_EMPTY_RECENT_TITLE')).toBeInTheDocument();
    });

    it('shows the generic empty state at the root', () => {
        renderContent();

        expect(screen.getByText('EMPTY_FILE_LIST')).toBeInTheDocument();
    });

    it('prefers a custom empty message such as the empty search', () => {
        renderContent({ selectedItem: emptyFolder }, { emptyStateMessage: 'FILES_SEARCH_EMPTY' });

        expect(screen.getByText('FILES_SEARCH_EMPTY')).toBeInTheDocument();
        expect(screen.queryByText('FILES_EMPTY_FOLDER_TITLE')).not.toBeInTheDocument();
    });
});
