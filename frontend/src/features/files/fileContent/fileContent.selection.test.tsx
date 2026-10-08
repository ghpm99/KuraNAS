import { fireEvent, render, screen } from '@testing-library/react';
import FileContent from './fileContent';
import { createTestFile } from '../selection/testFileFactory';
import { SelectionTestHarness } from '../selection/selectionTestHarness';
import { createFileContextStub } from '../selection/fileContextStub';

const mockOpenMediaItem = jest.fn();

jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({ openMediaItem: (...args: unknown[]) => mockOpenMediaItem(...args) }),
}));
jest.mock('@/components/loadMoreSentinel/loadMoreSentinel', () => ({
    __esModule: true,
    default: () => null,
}));

const listedFiles = [1, 2, 3, 4].map((id) => createTestFile(id));

const renderContent = (viewMode: 'grid' | 'list' = 'grid', contentItems?: typeof listedFiles) => {
    const handleSelectItem = jest.fn();
    const fileContext = createFileContextStub({
        status: 'success',
        files: listedFiles,
        handleSelectItem,
        handleStarredItem: jest.fn(),
        fileListFilter: 'all',
        hasNextPage: false,
        isFetchingNextPage: false,
        fetchNextPage: jest.fn(),
    });
    render(
        <SelectionTestHarness fileContext={fileContext} seedFiles={[]}>
            <FileContent showHeading={false} viewMode={viewMode} items={contentItems} />
        </SelectionTestHarness>
    );
    return { handleSelectItem };
};

const selectedCount = () => screen.getByTestId('selected-count').textContent;

describe('FileContent selection', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockOpenMediaItem.mockReturnValue(false);
    });

    it('renders the listing with real cards and no selection', () => {
        renderContent();

        expect(screen.getByText('file-1.txt')).toBeInTheDocument();
        expect(selectedCount()).toBe('0');
    });

    it('opens on plain click while nothing is selected', () => {
        const { handleSelectItem } = renderContent();

        fireEvent.click(screen.getByText('file-2.txt'));

        expect(handleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 2 }));
    });

    it('toggles with the checkbox and then toggles on plain click instead of opening', () => {
        const { handleSelectItem } = renderContent();

        fireEvent.click(screen.getAllByRole('checkbox')[0]!);
        expect(selectedCount()).toBe('1');

        fireEvent.click(screen.getByText('file-2.txt'));
        expect(selectedCount()).toBe('2');
        expect(handleSelectItem).not.toHaveBeenCalled();
    });

    it('toggles with ctrl and meta click', () => {
        renderContent();

        fireEvent.click(screen.getByText('file-1.txt'), { ctrlKey: true });
        fireEvent.click(screen.getByText('file-2.txt'), { metaKey: true });

        expect(selectedCount()).toBe('2');
    });

    it('selects a range with shift click from the anchor, in the list view too', () => {
        renderContent('list');

        fireEvent.click(screen.getByRole('button', { name: 'file-1.txt' }), { ctrlKey: true });
        fireEvent.click(screen.getByRole('button', { name: 'file-3.txt' }), { shiftKey: true });

        expect(selectedCount()).toBe('3');
    });

    it('selects a range with shift click on a checkbox', () => {
        renderContent();

        fireEvent.click(screen.getByText('file-1.txt'), { ctrlKey: true });
        fireEvent.click(screen.getAllByRole('checkbox')[3]!, { shiftKey: true });

        expect(selectedCount()).toBe('4');
    });

    it('does not offer selection for an externally provided item list but keeps opening', () => {
        const { handleSelectItem } = renderContent('grid', listedFiles);

        expect(screen.queryByRole('checkbox')).toBeNull();
        fireEvent.click(screen.getByText('file-1.txt'), { ctrlKey: true });

        expect(selectedCount()).toBe('0');
        expect(handleSelectItem).toHaveBeenCalledTimes(1);
    });

    it('opens the context menu on right click and from the menu button', () => {
        renderContent();

        fireEvent.contextMenu(screen.getByText('file-1.txt'));
        expect(screen.getByRole('menu')).toBeInTheDocument();
    });

    it('applies the context menu to the whole selection when the target is selected', () => {
        renderContent('list');

        fireEvent.click(screen.getByRole('button', { name: 'file-1.txt' }), { ctrlKey: true });
        fireEvent.click(screen.getByRole('button', { name: 'file-2.txt' }), { ctrlKey: true });
        fireEvent.contextMenu(screen.getByRole('button', { name: 'file-2.txt' }));

        expect(screen.getByText('FILES_OPEN').closest('[role="menuitem"]')).toHaveAttribute(
            'aria-disabled',
            'true'
        );
    });

    it('opens the menu from the item menu button and closes it with escape', () => {
        renderContent();

        fireEvent.click(screen.getAllByRole('button', { name: 'FILES_ITEM_MENU' })[0]!);
        expect(screen.getByRole('menu')).toBeInTheDocument();
        expect(screen.getByText('FILES_OPEN').closest('[role="menuitem"]')).not.toHaveAttribute(
            'aria-disabled',
            'true'
        );

        fireEvent.click(screen.getByText('FILES_OPEN'));
    });
});
