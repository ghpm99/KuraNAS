import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoDetailsPanel from './VideoDetailsPanel';

describe('VideoDetailsPanel without mocks', () => {
    it('renders nothing when there is no video', () => {
        const { container } = renderWithoutBackend(
            <VideoDetailsPanel video={null} onClose={jest.fn()} />
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('mounts a bare video against an absent backend and keeps the download link', async () => {
        renderWithoutBackend(<VideoDetailsPanel video={{ id: 9 } as any} onClose={jest.fn()} />);

        const downloadLink = await screen.findByRole('link', {
            name: 'VIDEO_ERROR_DOWNLOAD_ORIGINAL',
        });
        expect(downloadLink).toHaveAttribute('href', expect.stringContaining('/files/download/9'));
        expect(downloadLink).toHaveAttribute('download');
        expect(await screen.findByText('VIDEO_DETAILS_METADATA_UNAVAILABLE')).toBeInTheDocument();
        expect(screen.queryByText('0')).not.toBeInTheDocument();
        expect(screen.queryByText('undefined')).not.toBeInTheDocument();
    });
});
