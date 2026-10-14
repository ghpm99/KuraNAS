import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoLibrarySection from './VideoLibrarySection';

const video: VideoFileDto = {
    id: 5,
    name: 'Episode 5',
    path: '/shows/episode-5.mp4',
    parent_path: '/shows',
    format: 'mp4',
    size: 500,
};

describe('VideoLibrarySection details action', () => {
    it('opens the details panel for the chosen video and closes it again', async () => {
        renderWithoutBackend(
            <VideoLibrarySection
                videos={[video]}
                search=""
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={jest.fn()}
            />
        );

        expect(screen.queryByText('VIDEO_DETAILS_TITLE')).toBeNull();

        fireEvent.click(await screen.findByRole('button', { name: 'FILES_DETAILS' }));

        expect(
            await screen.findByRole('link', { name: 'VIDEO_ERROR_DOWNLOAD_ORIGINAL' })
        ).toHaveAttribute('href', expect.stringMatching(/\/files\/download\/5$/));

        fireEvent.click(screen.getByRole('button', { name: 'CLOSE' }));
        expect(screen.queryByRole('link', { name: 'VIDEO_ERROR_DOWNLOAD_ORIGINAL' })).toBeNull();
    });
});
