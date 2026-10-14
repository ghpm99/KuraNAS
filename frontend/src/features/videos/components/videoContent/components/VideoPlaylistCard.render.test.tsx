import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import VideoPlaylistCard from './VideoPlaylistCard';

describe('VideoPlaylistCard without mocks', () => {
    it('mounts with a bare playlist and no cover against an absent backend', () => {
        renderWithoutBackend(
            <VideoPlaylistCard
                playlist={{ id: 3, name: 'Bare' } as VideoPlaylistDto}
                onSelect={jest.fn()}
                onPlay={jest.fn()}
            />
        );

        expect(screen.getByText('Bare')).toBeInTheDocument();
    });
});
