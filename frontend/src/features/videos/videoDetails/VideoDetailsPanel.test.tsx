import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoDetailsPanel from './VideoDetailsPanel';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const video: VideoFileDto = {
    id: 7,
    name: 'Movie.mkv',
    path: '/media/films/Movie.mkv',
    parent_path: '/media/films',
    format: '.mkv',
    size: 2048,
    created_at: '2026-01-01T10:00:00Z',
    updated_at: '0001-01-01T00:00:00Z',
    height: 1080,
    width: 1920,
};

const summary = {
    duration: '3600',
    width: 1920,
    height: 1080,
    codec_name: 'h264',
    frame_rate: 24,
    bit_rate: '8000000',
    audio_codec: 'aac',
    format_name: 'matroska,webm',
};

describe('VideoDetailsPanel', () => {
    beforeEach(() => mockGet.mockReset());

    it('shows listing data, technical metadata and omits missing dates', async () => {
        mockGet.mockImplementation((url: string) =>
            url === '/video/metadata/7'
                ? Promise.resolve({ data: summary })
                : Promise.reject(new Error(url))
        );

        renderWithoutBackend(<VideoDetailsPanel video={video} onClose={jest.fn()} />);

        expect(await screen.findByText('h264')).toBeInTheDocument();
        expect(screen.getByText('Movie.mkv')).toBeInTheDocument();
        expect(screen.getByText('2.00 KB')).toBeInTheDocument();
        expect(screen.getByText('MKV')).toBeInTheDocument();
        expect(screen.getByText('1920 × 1080 (VIDEO_BADGE_RESOLUTION_FHD)')).toBeInTheDocument();
        expect(screen.getByText('1:00:00')).toBeInTheDocument();
        expect(screen.getByText('24 fps')).toBeInTheDocument();
        expect(screen.getByText('8000 kbps')).toBeInTheDocument();
        expect(screen.getByText('aac')).toBeInTheDocument();
        expect(screen.getByText('matroska,webm')).toBeInTheDocument();
        expect(screen.getByText('CREATED')).toBeInTheDocument();
        expect(screen.queryByText('MODIFIED')).not.toBeInTheDocument();
        expect(screen.queryByText('VIDEO_DETAILS_METADATA_UNAVAILABLE')).not.toBeInTheDocument();
        expect(mockGet).toHaveBeenCalledWith('/video/metadata/7');
    });

    it('links the folder to the files screen and the download to the original', async () => {
        mockGet.mockRejectedValue(new Error('absent'));

        renderWithoutBackend(<VideoDetailsPanel video={video} onClose={jest.fn()} />);

        expect(await screen.findByRole('link', { name: '/media/films' })).toHaveAttribute(
            'href',
            '/files/media/films'
        );
        expect(screen.getByRole('link', { name: 'VIDEO_ERROR_DOWNLOAD_ORIGINAL' })).toHaveAttribute(
            'href',
            expect.stringMatching(/\/files\/download\/7$/)
        );
    });

    it('keeps listing data and renders the backend message verbatim on metadata error', async () => {
        mockGet.mockRejectedValue({ response: { data: { error: 'Metadados não encontrados' } } });

        renderWithoutBackend(<VideoDetailsPanel video={video} onClose={jest.fn()} />);

        expect(await screen.findByText('Metadados não encontrados')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_DETAILS_METADATA_UNAVAILABLE')).toBeInTheDocument();
        expect(screen.getByText('1920 × 1080 (VIDEO_BADGE_RESOLUTION_FHD)')).toBeInTheDocument();
    });

    it('calls onClose from the close button', async () => {
        mockGet.mockRejectedValue(new Error('absent'));
        const onClose = jest.fn();

        renderWithoutBackend(<VideoDetailsPanel video={video} onClose={onClose} />);

        fireEvent.click(await screen.findByRole('button', { name: 'CLOSE' }));
        expect(onClose).toHaveBeenCalled();
    });
});
