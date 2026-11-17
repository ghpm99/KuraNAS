import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
import * as musicService from '@/service/music';
import { buildServerPage } from '@/shared/test/buildServerPage';
import GenresView from './GenresView';

jest.mock('@/service/music');
jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const mockedMusic = jest.mocked(musicService);
const emptyPageOf = <ItemType,>() =>
    buildServerPage<ItemType>([], { page: 1, pageSize: 50, hasNext: false });

const LocationProbe = () => <span data-testid="location">{useLocation().search}</span>;

const renderGenresView = (initialEntry: string) =>
    render(
        <QueryClientProvider client={new QueryClient()}>
            <MemoryRouter initialEntries={[initialEntry]}>
                <GenresView />
                <LocationProbe />
            </MemoryRouter>
        </QueryClientProvider>
    );

describe('GenresView rename action', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedMusic.getMusicGenres.mockResolvedValue(
            buildServerPage([{ key: 'rock', genre: 'Rock', track_count: 3 }], {
                page: 1,
                pageSize: 50,
                hasNext: false,
            })
        );
        mockedMusic.getMusicByGenre.mockResolvedValue(emptyPageOf());
        mockedMusic.getMusicArtists.mockResolvedValue(emptyPageOf());
        mockedMusic.getMusicGenreSummary.mockResolvedValue({
            key: 'rock',
            name: 'Rock',
            track_count: 3,
            total_length_seconds: 0,
        });
    });

    it('is not offered on the genre list', async () => {
        renderGenresView('/');

        await screen.findByText('Rock');

        expect(
            screen.queryByRole('button', { name: 'MUSIC_GENRE_RENAME' })
        ).not.toBeInTheDocument();
    });

    it('is offered on the selected genre once its summary is loaded and starts from the current name', async () => {
        renderGenresView('/?genre=rock');

        fireEvent.click(await screen.findByRole('button', { name: 'MUSIC_GENRE_RENAME' }));

        expect(screen.getByLabelText('MUSIC_GENRE_RENAME_FIELD_NAME')).toHaveValue('Rock');
    });

    it('is not offered when the genre summary cannot be loaded', async () => {
        mockedMusic.getMusicGenreSummary.mockRejectedValue(new Error('not found'));
        renderGenresView('/?genre=unknown-key');

        await waitFor(() => expect(mockedMusic.getMusicGenreSummary).toHaveBeenCalledTimes(1));

        expect(
            screen.queryByRole('button', { name: 'MUSIC_GENRE_RENAME' })
        ).not.toBeInTheDocument();
    });

    it('opens the resulting genre after a rename replaces the old key', async () => {
        mockedMusic.renameMusicGenre.mockResolvedValue({
            key: 'hard-rock',
            name: 'Hard Rock',
            trackCount: 3,
            merged: false,
        });
        renderGenresView('/?genre=rock');

        fireEvent.click(await screen.findByRole('button', { name: 'MUSIC_GENRE_RENAME' }));
        fireEvent.change(screen.getByLabelText('MUSIC_GENRE_RENAME_FIELD_NAME'), {
            target: { value: 'Hard Rock' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_REVIEW' }));
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_CONFIRM' }));

        await waitFor(() =>
            expect(screen.getByTestId('location').textContent).toBe('?genre=hard-rock')
        );
        expect(mockedMusic.renameMusicGenre).toHaveBeenCalledWith('rock', 'Hard Rock');
    });
});
