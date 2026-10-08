import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import FileDetails from './fileDetails';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => ({ isLoadingAccessData: false, recentAccessFiles: [] }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: jest.fn() }),
}));

const respondWith = (routes: Record<string, unknown>) => {
    mockGet.mockImplementation((url: string) =>
        url in routes ? Promise.resolve({ data: routes[url] }) : Promise.reject(new Error(url))
    );
};

const renderDetails = (file: Record<string, unknown>) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <FileDetails file={file as any} onClose={jest.fn()} />
        </QueryClientProvider>
    );

const fileBase = {
    id: 7,
    name: 'x',
    type: 2,
    size: 10,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    path: '/x',
};

const requestedUrls = () => mockGet.mock.calls.map((call) => call[0]);

describe('FileDetails seam (apiBase only)', () => {
    beforeEach(() => mockGet.mockReset());

    it('computes recursive folder stats from the folder-stats endpoint', async () => {
        respondWith({
            '/files/folder-stats/5': { file_count: 12, folder_count: 3, total_size_bytes: 2048 },
        });

        renderDetails({ ...fileBase, id: 5, type: 1, format: '' });

        expect(
            screen.getByText('FOLDER_STATS_CALCULATING', { selector: 'span' })
        ).toBeInTheDocument();
        expect(await screen.findByText('12')).toBeInTheDocument();
        expect(screen.getByText('3')).toBeInTheDocument();
        expect(screen.getByText('2.00 KB')).toBeInTheDocument();
        expect(mockGet).toHaveBeenCalledWith('/files/folder-stats/5');
    });

    it('hides the folder stats silently when the request fails', async () => {
        respondWith({});

        renderDetails({ ...fileBase, id: 5, type: 1, format: '' });

        await waitFor(() =>
            expect(screen.queryByText('FOLDER_STATS_CALCULATING', { selector: 'span' })).toBeNull()
        );
        expect(screen.queryByText('FOLDER_STATS_FILE_COUNT')).toBeNull();
    });

    it('composes image EXIF rows from the image endpoint only', async () => {
        respondWith({
            '/image/metadata/7': {
                width: 4000,
                height: 3000,
                make: 'Canon',
                model: 'R5',
                lens_model: '',
                datetime_original: '2026:01:01 10:00:00',
                exposure_time: 0.01,
                f_number: 1.8,
                iso: 200,
                focal_length: 50,
            },
        });

        renderDetails({ ...fileBase, format: '.jpg' });

        expect(await screen.findByText('4000 × 3000')).toBeInTheDocument();
        expect(screen.getByText('Canon R5')).toBeInTheDocument();
        expect(screen.getByText('1/100 s')).toBeInTheDocument();
        expect(screen.getByText('f/1.8')).toBeInTheDocument();
        expect(screen.getByText('50 mm')).toBeInTheDocument();
        expect(screen.queryByText('FILE_METADATA_LENS')).toBeNull();
        expect(requestedUrls()).toContain('/image/metadata/7');
        expect(requestedUrls()).not.toContain('/music/metadata/7');
        expect(requestedUrls()).not.toContain('/video/metadata/7');
    });

    it('composes audio tag rows from the music endpoint only', async () => {
        respondWith({
            '/music/metadata/7': {
                title: 'Song',
                artist: 'Band',
                album: 'Album',
                genre: 'Rock',
                year: '2020',
                track_number: '3',
                length: 185.4,
                bitrate: 320000,
                sample_rate: 44100,
                channels: 2,
            },
        });

        renderDetails({ ...fileBase, format: '.mp3' });

        expect(await screen.findByText('Song')).toBeInTheDocument();
        expect(screen.getByText('3:05')).toBeInTheDocument();
        expect(screen.getByText('320 kbps')).toBeInTheDocument();
        expect(screen.getByText('44100 Hz')).toBeInTheDocument();
        expect(requestedUrls()).toContain('/music/metadata/7');
        expect(requestedUrls()).not.toContain('/image/metadata/7');
    });

    it('composes video rows from the video endpoint only', async () => {
        respondWith({
            '/video/metadata/7': {
                duration: '3725.5',
                width: 1920,
                height: 1080,
                codec_name: 'h264',
                frame_rate: 23.976,
                bit_rate: '5000000',
                audio_codec: 'aac',
                format_name: 'matroska',
            },
        });

        renderDetails({ ...fileBase, format: '.mkv' });

        expect(await screen.findByText('1:02:06')).toBeInTheDocument();
        expect(screen.getByText('1920 × 1080')).toBeInTheDocument();
        expect(screen.getByText('h264')).toBeInTheDocument();
        expect(screen.getByText('23.98 fps')).toBeInTheDocument();
        expect(screen.getByText('5000 kbps')).toBeInTheDocument();
        expect(requestedUrls()).toContain('/video/metadata/7');
    });

    it('hides the metadata section silently when the type endpoint fails', async () => {
        respondWith({});

        renderDetails({ ...fileBase, format: '.mp3' });

        await waitFor(() => expect(requestedUrls()).toContain('/music/metadata/7'));
        expect(screen.queryByText('FILE_METADATA_TITLE')).toBeNull();
    });

    it('does not request any type metadata for documents', async () => {
        respondWith({});

        renderDetails({ ...fileBase, format: '.pdf' });

        await waitFor(() => expect(requestedUrls()).toContain('/files/location/7'));
        expect(requestedUrls().filter((url) => String(url).includes('/metadata/'))).toEqual([]);
    });
});
