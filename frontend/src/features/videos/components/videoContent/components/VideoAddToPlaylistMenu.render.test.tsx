import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoAddToPlaylistMenu from './VideoAddToPlaylistMenu';

describe('VideoAddToPlaylistMenu without mocks', () => {
    it('mounts open against an absent backend and keeps the search field usable', async () => {
        renderWithoutBackend(
            <VideoAddToPlaylistMenu
                videoId={1}
                anchorEl={document.body}
                onClose={jest.fn()}
                onSelectPlaylist={jest.fn()}
            />
        );

        expect(await screen.findByLabelText('VIDEO_PLAYLIST_SEARCH')).toBeInTheDocument();
    });
});
