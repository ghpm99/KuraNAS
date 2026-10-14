import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoHiddenPlaylistsLink from './VideoHiddenPlaylistsLink';

describe('VideoHiddenPlaylistsLink', () => {
    it('mounts without any backend and links to the hidden playlists route', () => {
        renderWithoutBackend(<VideoHiddenPlaylistsLink />);

        expect(screen.getByRole('link', { name: 'VIDEO_HIDDEN_PLAYLISTS_LINK' })).toHaveAttribute(
            'href',
            '/videos/hidden'
        );
    });
});
