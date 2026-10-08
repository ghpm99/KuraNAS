import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import FindByDiskPathDialog from './findByDiskPathDialog';
import { getFileByDiskPath } from '@/service/files';

jest.mock('@/service/files', () => ({
    ...jest.requireActual('@/service/files'),
    getFileByDiskPath: jest.fn(),
}));

const mockedGetFileByDiskPath = getFileByDiskPath as jest.Mock;

describe('FindByDiskPathDialog behavior', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('looks the path up, reports the found file and closes', async () => {
        const file = { id: 4, name: 'a.txt', path: '/docs/a.txt' };
        mockedGetFileByDiskPath.mockResolvedValue(file);
        const onClose = jest.fn();
        const onFileFound = jest.fn();
        render(<FindByDiskPathDialog open onClose={onClose} onFileFound={onFileFound} />);

        fireEvent.change(screen.getByRole('textbox'), { target: { value: '  F:\\cold\\a.txt  ' } });
        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_SUBMIT' }));

        await waitFor(() => expect(onFileFound).toHaveBeenCalledWith(file));
        expect(mockedGetFileByDiskPath).toHaveBeenCalledWith('F:\\cold\\a.txt');
        expect(onClose).toHaveBeenCalled();
    });

    it('renders the backend message verbatim on a 404 and keeps the dialog open', async () => {
        mockedGetFileByDiskPath.mockRejectedValue({
            response: { status: 404, data: { error: 'Nenhum arquivo indexado neste caminho' } },
        });
        const onClose = jest.fn();
        const onFileFound = jest.fn();
        render(<FindByDiskPathDialog open onClose={onClose} onFileFound={onFileFound} />);

        fireEvent.change(screen.getByRole('textbox'), { target: { value: '/nope' } });
        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_SUBMIT' }));

        expect(
            await screen.findByText('Nenhum arquivo indexado neste caminho')
        ).toBeInTheDocument();
        expect(onFileFound).not.toHaveBeenCalled();
        expect(onClose).not.toHaveBeenCalled();
    });

    it('does nothing for a blank path and clears state on cancel', () => {
        const onClose = jest.fn();
        render(<FindByDiskPathDialog open onClose={onClose} />);

        fireEvent.submit(screen.getByRole('textbox').closest('form') as HTMLFormElement);
        expect(mockedGetFileByDiskPath).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_CANCEL' }));
        expect(onClose).toHaveBeenCalled();
    });
});
