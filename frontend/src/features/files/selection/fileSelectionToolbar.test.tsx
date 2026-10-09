import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import FileSelectionToolbar from './fileSelectionToolbar';
import { createTestFile } from './testFileFactory';
import { SelectionTestHarness } from './selectionTestHarness';
import { createFileContextStub } from './fileContextStub';

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
        open ? <button onClick={() => onSelect({ folderId: 99 })}>CONFIRM_PICKER</button> : null,
}));

const listedFiles = [1, 2, 3].map((id) => createTestFile(id));

const renderToolbar = (fileContextOverrides = {}, seedFiles = listedFiles) => {
    const fileContext = createFileContextStub({ files: listedFiles, ...fileContextOverrides });
    render(
        <SelectionTestHarness fileContext={fileContext} seedFiles={seedFiles}>
            <FileSelectionToolbar />
        </SelectionTestHarness>
    );
    fireEvent.click(screen.getByText('seed-selection'));
    return fileContext;
};

describe('FileSelectionToolbar', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders an empty selection without any service mock', () => {
        render(
            <SelectionTestHarness fileContext={createFileContextStub()} seedFiles={[]}>
                <FileSelectionToolbar />
            </SelectionTestHarness>
        );

        expect(screen.getByRole('toolbar')).toBeInTheDocument();
        expect(screen.getByText('FILES_SELECTION_COUNT')).toBeInTheDocument();
    });

    it('selects all listed files and clears the selection', () => {
        renderToolbar({}, [listedFiles[0]!]);
        expect(screen.getByTestId('selected-count')).toHaveTextContent('1');

        fireEvent.click(screen.getByText('FILES_SELECT_ALL'));
        expect(screen.getByTestId('selected-count')).toHaveTextContent('3');

        fireEvent.click(screen.getByLabelText('FILES_CLEAR_SELECTION'));
        expect(screen.getByTestId('selected-count')).toHaveTextContent('0');
    });

    it('offers rename only for a single selected file', () => {
        renderToolbar({}, [listedFiles[0]!]);
        expect(screen.getByText('RENAME')).toBeInTheDocument();
    });

    it('hides rename when several files are selected', () => {
        renderToolbar();
        expect(screen.queryByText('RENAME')).toBeNull();
    });

    it('downloads a single file by url and several files as one zip', () => {
        renderToolbar();
        fireEvent.click(screen.getByText('DOWNLOAD'));
        expect(mockTriggerBrowserDownload).toHaveBeenCalledWith(
            expect.stringContaining('download-zip?ids=1,2,3')
        );
    });

    it('moves every selected file after one folder pick and clears the selection', async () => {
        const fileContext = renderToolbar();

        fireEvent.click(screen.getByText('MOVE'));
        fireEvent.click(screen.getByText('CONFIRM_PICKER'));

        await waitFor(() => expect(fileContext.moveFile).toHaveBeenCalledTimes(3));
        expect(fileContext.moveFile).toHaveBeenCalledWith(2, 99, undefined);
        await waitFor(() =>
            expect(screen.getByTestId('selected-count')).toHaveTextContent('0')
        );
    });

    it('copies every selected file after one folder pick', async () => {
        const fileContext = renderToolbar();

        fireEvent.click(screen.getByText('COPY'));
        fireEvent.click(screen.getByText('CONFIRM_PICKER'));

        await waitFor(() => expect(fileContext.copyFile).toHaveBeenCalledTimes(3));
    });

    it('toggles favorites for the selection', async () => {
        const fileContext = renderToolbar();

        fireEvent.click(screen.getByText('FILES_FAVORITE'));

        await waitFor(() => expect(fileContext.toggleStarred).toHaveBeenCalledTimes(3));
    });

    it('shows the unfavorite label when every selected file is a favorite', () => {
        renderToolbar({}, [createTestFile(1, { starred: true })]);
        expect(screen.getByText('FILES_UNFAVORITE')).toBeInTheDocument();
    });

    it('confirms once, deletes all, keeps only the failed file selected and reports', async () => {
        const deleteFile = jest.fn(async (fileId: number) => {
            if (fileId === 2) throw { response: { data: { error: 'Pasta protegida' } } };
        });
        renderToolbar({ deleteFile });

        fireEvent.click(screen.getByText('DELETE'));
        const confirmation = await screen.findByRole('dialog');
        expect(within(confirmation).getByText('FILES_CONFIRM_DELETE_MANY')).toBeInTheDocument();
        fireEvent.click(within(confirmation).getByRole('button', { name: 'DELETE' }));

        await waitFor(() => expect(deleteFile).toHaveBeenCalledTimes(3));
        await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(expect.any(String), { variant: 'warning' });
        await waitFor(() =>
            expect(screen.getByTestId('selected-count')).toHaveTextContent('1')
        );
    });

    it('renames the single selected file', async () => {
        const fileContext = renderToolbar({}, [listedFiles[0]!]);

        fireEvent.click(screen.getByText('RENAME'));
        const dialog = await screen.findByRole('dialog');
        fireEvent.change(within(dialog).getByRole('textbox'), { target: { value: 'novo.txt' } });
        fireEvent.click(within(dialog).getByRole('button', { name: 'RENAME' }));

        await waitFor(() => expect(fileContext.renameFile).toHaveBeenCalledWith(1, 'novo.txt'));
    });
});
