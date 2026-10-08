import { act, fireEvent, render, screen } from '@testing-library/react';
import FileContent from './fileContent';
import GlobalShortcutsProvider from '@/components/shortcuts/GlobalShortcutsProvider';
import { MemoryRouter } from 'react-router-dom';
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

const listedFiles = [1, 2, 3].map((id) => createTestFile(id));

const renderContent = (viewMode: 'grid' | 'list' = 'grid') => {
    const handleSelectItem = jest.fn();
    const onGoToParent = jest.fn();
    const onFocusSearch = jest.fn();
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
        <MemoryRouter>
            <GlobalShortcutsProvider>
                <SelectionTestHarness fileContext={fileContext} seedFiles={[]}>
                    <FileContent
                        showHeading={false}
                        viewMode={viewMode}
                        onGoToParent={onGoToParent}
                        onFocusSearch={onFocusSearch}
                    />
                </SelectionTestHarness>
            </GlobalShortcutsProvider>
        </MemoryRouter>
    );
    return { handleSelectItem, onGoToParent, onFocusSearch };
};

const press = (key: string, init: KeyboardEventInit = {}) => {
    act(() => {
        fireEvent.keyDown(document.body, { key, ...init });
    });
};

const itemLinks = () => document.querySelectorAll<HTMLElement>('[data-file-id]');

describe('FileContent keyboard shortcuts', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockOpenMediaItem.mockReturnValue(false);
    });

    it('exposes a single roving tab stop on the first item', () => {
        renderContent();

        expect(Array.from(itemLinks()).map((link) => link.tabIndex)).toEqual([0, -1, -1]);
    });

    it('moves the tab stop and DOM focus with the arrow keys in list mode', () => {
        renderContent('list');

        press('ArrowDown');
        press('ArrowDown');

        expect(Array.from(itemLinks()).map((link) => link.tabIndex)).toEqual([-1, 0, -1]);
        expect(itemLinks()[1]).toHaveFocus();
    });

    it('opens the focused item with Enter', () => {
        const { handleSelectItem } = renderContent();

        press('ArrowDown');
        press('ArrowRight');
        press('Enter');

        expect(handleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 2 }));
    });

    it('selects everything with Ctrl+A', () => {
        renderContent();

        press('a', { ctrlKey: true });
        expect(screen.getByTestId('selected-count')).toHaveTextContent('3');
    });

    it('opens the delete dialog for the focused item with Delete', async () => {
        renderContent();

        press('ArrowDown');
        press('Delete');

        expect(await screen.findByText('FILES_DELETE_TRASH_NOTICE')).toBeInTheDocument();
    });

    it('opens the rename dialog for the focused item with F2', async () => {
        renderContent();

        press('ArrowDown');
        press('F2');

        expect(await screen.findByLabelText('NAME')).toHaveValue('file-1.txt');
    });

    it('extends the selection with Shift+arrows', () => {
        renderContent();

        press('ArrowDown');
        press('ArrowRight', { shiftKey: true });

        expect(screen.getByTestId('selected-count')).toHaveTextContent('2');
    });

    it('forwards Backspace and / to the screen handlers', () => {
        const { onGoToParent, onFocusSearch } = renderContent();

        press('Backspace');
        press('/');

        expect(onGoToParent).toHaveBeenCalledTimes(1);
        expect(onFocusSearch).toHaveBeenCalledTimes(1);
    });

    it('lists the file shortcuts in the global help dialog while mounted', async () => {
        renderContent();

        press('?', { shiftKey: true });

        expect(await screen.findByText('SHORTCUTS_SECTION_CURRENT_PAGE')).toBeInTheDocument();
        expect(screen.getByText('FILES_SHORTCUT_OPEN')).toBeInTheDocument();
    });
});
