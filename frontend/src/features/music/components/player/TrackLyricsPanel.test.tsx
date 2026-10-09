import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import type { IMusicData } from '@/types/music';
import { getAudioSummary } from '@/service/fileTypeMetadata';
import TrackLyricsPanel from './TrackLyricsPanel';

jest.mock('@/service/fileTypeMetadata', () => ({
    getAudioSummary: jest.fn(),
}));

const mockedGetAudioSummary = getAudioSummary as jest.Mock;

const renderPanel = (track?: Partial<IMusicData>) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <TrackLyricsPanel track={track as IMusicData | undefined} />
        </QueryClientProvider>
    );

describe('TrackLyricsPanel', () => {
    beforeEach(() => {
        mockedGetAudioSummary.mockReset();
    });

    it('renders the empty state without a track and never calls the backend', () => {
        renderPanel(undefined);
        expect(screen.getByText('PLAYER_LYRICS_UNAVAILABLE')).toBeInTheDocument();
        expect(mockedGetAudioSummary).not.toHaveBeenCalled();
    });

    it('shows lyrics from the queue metadata preserving line breaks without fetching', () => {
        renderPanel({ id: 1, name: 'Song', metadata: { lyrics: 'line one\nline two' } });
        const lyricsBlock = screen.getByTestId('track-lyrics');
        expect(lyricsBlock.textContent).toBe('line one\nline two');
        expect(mockedGetAudioSummary).not.toHaveBeenCalled();
    });

    it('fetches the lyrics once when the queue metadata has none', async () => {
        mockedGetAudioSummary.mockResolvedValue({ lyrics: 'fetched lyrics' });
        renderPanel({ id: 7, name: 'Song', metadata: {} });
        expect(screen.getByRole('status')).toBeInTheDocument();
        expect(await screen.findByText('fetched lyrics')).toBeInTheDocument();
        expect(mockedGetAudioSummary).toHaveBeenCalledTimes(1);
        expect(mockedGetAudioSummary).toHaveBeenCalledWith(7);
    });

    it('shows the empty state when the fetched metadata has no lyrics', async () => {
        mockedGetAudioSummary.mockResolvedValue({});
        renderPanel({ id: 8, name: 'Song' });
        expect(await screen.findByText('PLAYER_LYRICS_UNAVAILABLE')).toBeInTheDocument();
    });

    it('shows the empty state when the fetch fails', async () => {
        mockedGetAudioSummary.mockRejectedValue(new Error('offline'));
        renderPanel({ id: 9, name: 'Song' });
        expect(await screen.findByText('PLAYER_LYRICS_UNAVAILABLE')).toBeInTheDocument();
    });
});
