import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import type { IMusicData } from '@/types/music';
import TrackContextMenu from './TrackContextMenu';

const mockPlayNext = jest.fn();
const mockAddToQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ playNext: mockPlayNext, addToQueue: mockAddToQueue }),
}));

const track = {
    id: 7,
    metadata: { artist: 'Band', album: 'Hits' },
} as IMusicData;

const LocationProbe = () => {
    const location = useLocation();
    return <div data-testid="location">{`${location.pathname}${location.search}`}</div>;
};

const renderMenu = (
    menuTrack: IMusicData,
    handlers = { onClose: jest.fn(), onAddToPlaylist: jest.fn() }
) => {
    render(
        <MemoryRouter>
            <Routes>
                <Route
                    path="*"
                    element={
                        <>
                            <TrackContextMenu
                                track={menuTrack}
                                position={{ top: 10, left: 10 }}
                                {...handlers}
                            />
                            <LocationProbe />
                        </>
                    }
                />
            </Routes>
        </MemoryRouter>
    );
    return handlers;
};

describe('TrackContextMenu', () => {
    beforeEach(() => jest.clearAllMocks());

    it('queues the track right after the current one', () => {
        const { onClose } = renderMenu(track);

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_PLAY_NEXT' }));

        expect(mockPlayNext).toHaveBeenCalledWith([track]);
        expect(onClose).toHaveBeenCalled();
    });

    it('appends the track to the end of the queue', () => {
        renderMenu(track);

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_ADD_TO_QUEUE' }));

        expect(mockAddToQueue).toHaveBeenCalledWith([track]);
    });

    it('delegates adding to a playlist', () => {
        const { onAddToPlaylist } = renderMenu(track);

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_ADD_TO_PLAYLIST' }));

        expect(onAddToPlaylist).toHaveBeenCalled();
    });

    it('navigates to the album and to the artist', () => {
        renderMenu(track);

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_GO_TO_ALBUM' }));
        expect(screen.getByTestId('location')).toHaveTextContent(
            '/music/albums?album=band%3A%3Ahits'
        );

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_GO_TO_ARTIST' }));
        expect(screen.getByTestId('location')).toHaveTextContent('/music/artists?artist=band');
    });

    it('hides navigation entries when the track has no album or artist', () => {
        renderMenu({ id: 8 } as IMusicData);

        expect(screen.queryByRole('menuitem', { name: 'MUSIC_GO_TO_ALBUM' })).toBeNull();
        expect(screen.queryByRole('menuitem', { name: 'MUSIC_GO_TO_ARTIST' })).toBeNull();
    });
});
