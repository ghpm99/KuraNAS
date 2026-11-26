import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { apiBase } from '@/service';
import TakeoutImportScreen from './TakeoutImportScreen';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

const selectFile = (container: HTMLElement, file: File) => {
    const fileInput = container.querySelector('input[type="file"]') as HTMLInputElement;
    act(() => {
        fireEvent.change(fileInput, { target: { files: { length: 1, item: () => file } } });
    });
};

describe('TakeoutImportScreen queue', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        let initCount = 0;
        mockedApi.post.mockImplementation((url: string) => {
            if (url === '/takeout/upload/init') {
                initCount += 1;
                return Promise.resolve({ data: { upload_id: `u${initCount}`, chunk_size: 1024 } });
            }
            if (url === '/takeout/upload/chunk')
                return Promise.resolve({ data: { received: true } });
            return Promise.resolve({ data: { job_id: 1, message: 'ok' } });
        });
    });

    it('lists both added files and sends both, each one finishing as complete', async () => {
        const { container } = render(<TakeoutImportScreen />);
        selectFile(container, new File(['a'], 'first.zip'));
        selectFile(container, new File(['b'], 'second.zip'));

        expect(screen.getByText(/first\.zip/)).toBeInTheDocument();
        expect(screen.getByText(/second\.zip/)).toBeInTheDocument();
        expect(screen.getAllByText('TAKEOUT_QUEUE_WAITING')).toHaveLength(2);

        fireEvent.click(screen.getByRole('button', { name: 'TAKEOUT_START_UPLOAD' }));

        await waitFor(() => expect(screen.getAllByText('TAKEOUT_UPLOAD_COMPLETE')).toHaveLength(2));
        expect(screen.getByRole('button', { name: 'TAKEOUT_START_UPLOAD' })).toBeDisabled();
    });

    it('shows the type error for a non zip file and keeps the start button disabled', () => {
        const { container } = render(<TakeoutImportScreen />);
        selectFile(container, new File(['a'], 'backup.rar'));

        expect(screen.getByText('TAKEOUT_FILE_TYPE_ERROR')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'TAKEOUT_START_UPLOAD' })).toBeDisabled();
    });

    it('removes settled entries on clear', async () => {
        const { container } = render(<TakeoutImportScreen />);
        selectFile(container, new File(['a'], 'backup.rar'));

        fireEvent.click(screen.getByRole('button', { name: 'TAKEOUT_CLEAR' }));

        expect(screen.queryByText(/backup\.rar/)).not.toBeInTheDocument();
    });
});
