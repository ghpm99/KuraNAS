import { render, screen } from '@testing-library/react';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import type { IMusicData } from '@/types/music';
import TrackListItem from '../TrackListItem';

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    useSettings: () => ({
        settings: { players: { remember_music_queue: false } },
        isLoading: true,
    }),
}));

describe('TrackListItem with the real provider and no router or services', () => {
    beforeAll(() => {
        window.HTMLMediaElement.prototype.pause = () => undefined;
    });

    it('renders a partial track with its more actions button', () => {
        render(
            <GlobalMusicProvider>
                <TrackListItem
                    track={{ id: 3, name: 'bare' } as IMusicData}
                    index={0}
                    onPlay={jest.fn()}
                />
            </GlobalMusicProvider>
        );

        expect(
            screen.getByRole('button', { name: 'MUSIC_TRACK_MORE_ACTIONS' })
        ).toBeInTheDocument();
    });
});
