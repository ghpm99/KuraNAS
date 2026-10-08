import { createEvent, fireEvent, render, screen, waitFor } from '@testing-library/react';
import UploadDropZone from './uploadDropZone';

const mockUploadEntries = jest.fn();

jest.mock('./useUploadToCurrentFolder', () => ({
    __esModule: true,
    default: () => ({ uploadEntries: mockUploadEntries }),
}));

const buildFileDrag = (types: string[] = ['Files']) =>
    ({ types, items: [], files: [new File(['x'], 'a.txt')] }) as unknown as DataTransfer;

describe('UploadDropZone behavior', () => {
    beforeEach(() => mockUploadEntries.mockReset());

    const renderZone = () =>
        render(
            <UploadDropZone>
                <span>listing</span>
            </UploadDropZone>
        );

    it('shows the overlay while files are dragged over and hides it on leave', () => {
        renderZone();
        const zone = screen.getByText('listing').parentElement!;

        fireEvent.dragEnter(zone, { dataTransfer: buildFileDrag() });
        expect(screen.getByText('FILES_UPLOAD_DROP_HINT')).toBeInTheDocument();

        fireEvent.dragLeave(zone, { dataTransfer: buildFileDrag() });
        expect(screen.queryByText('FILES_UPLOAD_DROP_HINT')).not.toBeInTheDocument();
    });

    it('ignores drags that do not carry files', () => {
        renderZone();
        const zone = screen.getByText('listing').parentElement!;

        fireEvent.dragEnter(zone, { dataTransfer: buildFileDrag(['text/plain']) });
        fireEvent.dragOver(zone, { dataTransfer: buildFileDrag(['text/plain']) });
        fireEvent.dragLeave(zone, { dataTransfer: buildFileDrag(['text/plain']) });
        fireEvent.drop(zone, { dataTransfer: buildFileDrag(['text/plain']) });

        expect(screen.queryByText('FILES_UPLOAD_DROP_HINT')).not.toBeInTheDocument();
        expect(mockUploadEntries).not.toHaveBeenCalled();
    });

    it('allows dropping by cancelling dragover for file drags', () => {
        renderZone();
        const zone = screen.getByText('listing').parentElement!;
        const dragOver = createEvent.dragOver(zone, { dataTransfer: buildFileDrag() });

        fireEvent(zone, dragOver);

        expect(dragOver.defaultPrevented).toBe(true);
    });

    it('collects the dropped files and hands them to the upload queue', async () => {
        renderZone();
        const zone = screen.getByText('listing').parentElement!;

        fireEvent.dragEnter(zone, { dataTransfer: buildFileDrag() });
        fireEvent.drop(zone, { dataTransfer: buildFileDrag() });

        await waitFor(() => expect(mockUploadEntries).toHaveBeenCalledTimes(1));
        const droppedEntries = mockUploadEntries.mock.calls[0]![0];
        expect(droppedEntries).toHaveLength(1);
        expect(droppedEntries[0].file.name).toBe('a.txt');
        expect(screen.queryByText('FILES_UPLOAD_DROP_HINT')).not.toBeInTheDocument();
    });
});
