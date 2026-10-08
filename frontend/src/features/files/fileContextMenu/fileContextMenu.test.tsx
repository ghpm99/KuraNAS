import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import FileContextMenu from './fileContextMenu';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import { SelectionTestHarness } from '@/features/files/selection/selectionTestHarness';
import { createFileContextStub } from '@/features/files/selection/fileContextStub';

const mockEnqueueSnackbar = jest.fn();
const mockTriggerBrowserDownload = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));
jest.mock('@/service/browserDownload', () => ({
    triggerBrowserDownload: (...args: unknown[]) => mockTriggerBrowserDownload(...args),
}));
jest.mock('@/components/folderPicker/folderPicker', () => ({
    __esModule: true,
    default: ({
        open,
        onSelect,
    }: {
        open: boolean;
        onSelect: (destination: { folderId: number }) => void;
    }) =>
        open ? <button onClick={() => onSelect({ folderId: 7 })}>CONFIRM_PICKER</button> : null,
}));

const anchorPosition = { top: 10, left: 10 };

const renderMenu = (targetFiles = [createTestFile(1)], overrides = {}) => {
    const fileContext = createFileContextStub(overrides);
    const onClose = jest.fn();
    const onOpenFile = jest.fn();
    render(
        <SelectionTestHarness fileContext={fileContext} seedFiles={targetFiles}>
            <FileContextMenu
                anchorPosition={anchorPosition}
                targetFiles={targetFiles}
                onClose={onClose}
                onOpenFile={onOpenFile}
            />
        </SelectionTestHarness>
    );
    return { fileContext, onClose, onOpenFile };
};

describe('FileContextMenu', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders nothing visible when closed and no backend exists', () => {
        render(
            <SelectionTestHarness fileContext={createFileContextStub()} seedFiles={[]}>
                <FileContextMenu
                    anchorPosition={null}
                    targetFiles={[]}
                    onClose={jest.fn()}
                    onOpenFile={jest.fn()}
                />
            </SelectionTestHarness>
        );

        expect(screen.queryByRole('menu')).toBeNull();
    });

    it('opens the target file', () => {
        const { onOpenFile, onClose } = renderMenu();

        fireEvent.click(screen.getByText('FILES_OPEN'));

        expect(onClose).toHaveBeenCalled();
        expect(onOpenFile).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }));
    });

    it('disables open and rename for a multi-selection', () => {
        renderMenu([createTestFile(1), createTestFile(2)]);

        expect(screen.getByText('FILES_OPEN').closest('[role="menuitem"]')).toHaveAttribute(
            'aria-disabled',
            'true'
        );
        expect(screen.getByText('RENAME').closest('[role="menuitem"]')).toHaveAttribute(
            'aria-disabled',
            'true'
        );
    });

    it('renames a single file through the dialog', async () => {
        const { fileContext } = renderMenu();

        fireEvent.click(screen.getByText('RENAME'));
        const nameInput = await screen.findByRole('textbox');
        fireEvent.change(nameInput, { target: { value: 'renamed.txt' } });
        fireEvent.click(screen.getAllByRole('button', { name: 'RENAME' }).pop()!);

        await waitFor(() => expect(fileContext.renameFile).toHaveBeenCalledWith(1, 'renamed.txt'));
    });

    it('moves and copies all targets through the folder picker', async () => {
        const targetFiles = [createTestFile(1), createTestFile(2)];
        const { fileContext } = renderMenu(targetFiles);

        fireEvent.click(screen.getByText('MOVE'));
        fireEvent.click(screen.getByText('CONFIRM_PICKER'));
        await waitFor(() => expect(fileContext.moveFile).toHaveBeenCalledTimes(2));

        fireEvent.click(screen.getByText('COPY'));
        fireEvent.click(screen.getByText('CONFIRM_PICKER'));
        await waitFor(() => expect(fileContext.copyFile).toHaveBeenCalledTimes(2));
    });

    it('downloads the targets', () => {
        renderMenu();

        fireEvent.click(screen.getByText('DOWNLOAD'));

        expect(mockTriggerBrowserDownload).toHaveBeenCalledWith(
            expect.stringContaining('/files/download/1'),
            'file-1.txt'
        );
    });

    it('favorites and unfavorites', async () => {
        const { fileContext } = renderMenu();
        fireEvent.click(screen.getByText('FILES_FAVORITE'));
        await waitFor(() => expect(fileContext.toggleStarred).toHaveBeenCalledWith(1));
    });

    it('labels the action as unfavorite for starred targets', () => {
        renderMenu([createTestFile(1, { starred: true })]);
        expect(screen.getByText('FILES_UNFAVORITE')).toBeInTheDocument();
    });

    it('copies the path guarded by clipboard availability', async () => {
        const writeText = jest.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
        renderMenu();

        fireEvent.click(screen.getByText('FILES_COPY_PATH'));

        await waitFor(() => expect(writeText).toHaveBeenCalledWith('/library/file-1.txt'));
    });

    it('deletes after confirmation', async () => {
        const { fileContext } = renderMenu();

        fireEvent.click(screen.getByText('DELETE'));
        const confirmButtons = await screen.findAllByRole('button', { name: 'DELETE' });
        fireEvent.click(confirmButtons[confirmButtons.length - 1]!);

        await waitFor(() => expect(fileContext.deleteFile).toHaveBeenCalledWith(1));
    });
});
