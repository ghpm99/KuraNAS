import { fireEvent, renderHook, screen, waitFor, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import { useImageSelection, type ImageSelection } from '../useImageSelection';
import ImageSelectionToolbar from './ImageSelectionToolbar';

const mockMoveFile = jest.fn();
const mockDeleteFile = jest.fn();
const mockToggleStarredFile = jest.fn();
const mockTriggerBrowserDownload = jest.fn();

jest.mock('@/service/files', () => ({
    moveFile: (...args: unknown[]) => mockMoveFile(...args),
    deleteFile: (...args: unknown[]) => mockDeleteFile(...args),
    toggleStarredFile: (...args: unknown[]) => mockToggleStarredFile(...args),
    getFileDownloadUrl: (fileId: number) => `/download/${fileId}`,
    getFilesZipDownloadUrl: (fileIds: number[]) => `/zip?ids=${fileIds.join(',')}`,
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
    }) => (open ? <button onClick={() => onSelect({ folderId: 99 })}>CONFIRM_PICKER</button> : null),
}));

const loadedImages = [1, 2, 3].map((fileId) => buildImageLibraryItem({ file_id: fileId }));

const buildSelection = (selectedImages = loadedImages.slice(0, 2)): ImageSelection => {
    const { result } = renderHook(() => useImageSelection('scope'));
    return { ...result.current, selectedItems: selectedImages, selectedCount: selectedImages.length };
};

describe('ImageSelectionToolbar', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockMoveFile.mockResolvedValue('/moved');
        mockDeleteFile.mockResolvedValue(undefined);
        mockToggleStarredFile.mockResolvedValue(undefined);
    });

    it('renders an empty selection without any service mock being called', () => {
        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection([])} loadedImages={[]} />
        );

        expect(screen.getByRole('toolbar')).toBeInTheDocument();
        expect(mockMoveFile).not.toHaveBeenCalled();
    });

    it('selects all loaded images and clears', () => {
        const selection = { ...buildSelection(), selectAll: jest.fn(), clear: jest.fn() };
        renderWithoutBackend(
            <ImageSelectionToolbar selection={selection} loadedImages={loadedImages} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_SELECT_ALL_LOADED' }));
        expect(selection.selectAll).toHaveBeenCalledWith(loadedImages);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_CLEAR_SELECTION' }));
        expect(selection.clear).toHaveBeenCalled();
    });

    it('downloads one image directly and several as a zip', () => {
        const { unmount } = renderWithoutBackend(
            <ImageSelectionToolbar
                selection={buildSelection([loadedImages[0]!])}
                loadedImages={loadedImages}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'DOWNLOAD' }));
        expect(mockTriggerBrowserDownload).toHaveBeenCalledWith('/download/1', 'Trip.jpg');
        unmount();

        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection()} loadedImages={loadedImages} />
        );
        fireEvent.click(screen.getByRole('button', { name: 'DOWNLOAD' }));
        expect(mockTriggerBrowserDownload).toHaveBeenLastCalledWith('/zip?ids=1,2');
    });

    it('favorites every selected image and deselects the ones that succeeded', async () => {
        const selection = { ...buildSelection(), deselect: jest.fn() };
        renderWithoutBackend(
            <ImageSelectionToolbar selection={selection} loadedImages={loadedImages} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'FILES_FAVORITE' }));

        await waitFor(() => expect(selection.deselect).toHaveBeenCalledWith(loadedImages.slice(0, 2)));
        expect(mockToggleStarredFile).toHaveBeenCalledTimes(2);
    });

    it('offers unfavorite when every selected image is starred', () => {
        const starredImages = loadedImages.map((image) => ({ ...image, starred: true }));
        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection(starredImages)} loadedImages={starredImages} />
        );

        expect(screen.getByRole('button', { name: 'FILES_UNFAVORITE' })).toBeInTheDocument();
    });

    it('moves the selection to the picked folder', async () => {
        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection()} loadedImages={loadedImages} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
        fireEvent.click(screen.getByText('CONFIRM_PICKER'));

        await waitFor(() => expect(mockMoveFile).toHaveBeenCalledTimes(2));
        expect(mockMoveFile).toHaveBeenCalledWith(1, 99, undefined);
    });

    it('deletes to the trash by default', async () => {
        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection()} loadedImages={loadedImages} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        fireEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'DELETE' }));

        await waitFor(() => expect(mockDeleteFile).toHaveBeenCalledTimes(2));
        expect(mockDeleteFile).toHaveBeenCalledWith(1, false);
    });

    it('deletes permanently when the option is checked', async () => {
        renderWithoutBackend(
            <ImageSelectionToolbar selection={buildSelection()} loadedImages={loadedImages} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.click(within(dialog).getByRole('checkbox', { name: 'FILES_DELETE_PERMANENTLY' }));
        fireEvent.click(within(dialog).getByRole('button', { name: 'FILES_DELETE_PERMANENTLY' }));

        await waitFor(() => expect(mockDeleteFile).toHaveBeenCalledWith(2, true));
    });
});
