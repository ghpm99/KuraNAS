import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import TrackStarButton from './TrackStarButton';

describe('TrackStarButton without mocks', () => {
    it('mounts against an absent backend with a partial track and shows an unstarred toggle', () => {
        const partialTrack = { id: 3, name: 'partial.mp3' } as IMusicData;

        renderWithoutBackend(<TrackStarButton track={partialTrack} />);

        expect(screen.getByRole('button', { pressed: false })).toBeInTheDocument();
    });
});
