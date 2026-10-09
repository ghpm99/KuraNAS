import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import FindByDiskPathDialog from './findByDiskPathDialog';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

describe('FindByDiskPathDialog (seam)', () => {
    it('GETs /files/by-disk-path with only the path query param and hands the FileDto over', async () => {
        const file = { id: 4, name: 'a.txt', path: '/docs/a.txt' };
        mockedApi.get.mockResolvedValue({ data: file });
        const onFileFound = jest.fn();
        render(<FindByDiskPathDialog open onClose={jest.fn()} onFileFound={onFileFound} />);

        fireEvent.change(screen.getByRole('textbox'), {
            target: { value: 'D:\\data\\docs\\a.txt' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_SUBMIT' }));

        await waitFor(() => expect(onFileFound).toHaveBeenCalledWith(file));
        expect(mockedApi.get).toHaveBeenCalledWith('/files/by-disk-path', {
            params: { path: 'D:\\data\\docs\\a.txt' },
        });
    });
});
