import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import * as musicService from '@/service/music';
import { buildServerPage } from '@/shared/test/buildServerPage';
import FoldersView from './FoldersView';

jest.mock('@/service/music');
jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => undefined,
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const mockedMusic = jest.mocked(musicService);
const emptyPageOf = <ItemType,>() =>
    buildServerPage<ItemType>([], { page: 1, pageSize: 50, hasNext: false });

describe('FoldersView folder edit action', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedMusic.getMusicFolders.mockResolvedValue(emptyPageOf());
        mockedMusic.getMusicByFolder.mockResolvedValue(emptyPageOf());
        mockedMusic.getMusicFolderSummary.mockResolvedValue({
            key: '/music/rock',
            name: 'rock',
            track_count: 2,
            total_length_seconds: 0,
        });
        mockedMusic.getMusicArtists.mockResolvedValue(emptyPageOf());
        mockedMusic.getMusicGenres.mockResolvedValue(emptyPageOf());
        mockedMusic.updateTracksMetadata.mockResolvedValue({ matched: 2, updated: 0, failed: [] });
    });

    it('opens the batch dialog for the selected folder path', async () => {
        render(
            <QueryClientProvider client={new QueryClient()}>
                <MemoryRouter initialEntries={['/?folder=%2Fmusic%2Frock']}>
                    <FoldersView />
                </MemoryRouter>
            </QueryClientProvider>
        );

        fireEvent.click(await screen.findByRole('button', { name: 'MUSIC_EDIT_FOLDER' }));
        fireEvent.change(screen.getByLabelText('MUSIC_METADATA_FIELD_ALBUM'), {
            target: { value: 'Rock Hits' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_REVIEW' }));

        await waitFor(() =>
            expect(mockedMusic.updateTracksMetadata).toHaveBeenCalledWith(
                { folderKey: '/music/rock' },
                { album: 'Rock Hits' },
                { isDryRun: true }
            )
        );
    });
});
