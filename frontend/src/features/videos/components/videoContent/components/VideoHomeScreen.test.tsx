import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VideoHomeScreen from './VideoHomeScreen';

const renderHome = (loading: Partial<Parameters<typeof VideoHomeScreen>[0]> = {}) =>
    render(
        <MemoryRouter>
            <VideoHomeScreen
                continueWatchingItems={[]}
                seriesPlaylists={[]}
                moviePlaylists={[]}
                personalPlaylists={[]}
                clipPlaylists={[]}
                folderPlaylists={[]}
                recentCatalogItems={[]}
                onSelectPlaylist={jest.fn()}
                onPlayVideo={jest.fn()}
                {...loading}
            />
        </MemoryRouter>
    );

describe('VideoHomeScreen', () => {
    it('renders without loading flags and without data', () => {
        renderHome();
        expect(screen.queryAllByTestId('video-section-skeleton')).toHaveLength(0);
    });

    it('shows one skeleton per playlist section while playlists load', () => {
        renderHome({ isLoadingPlaylists: true });
        expect(screen.getAllByTestId('video-section-skeleton')).toHaveLength(5);
    });

    it('shows a skeleton for continue watching only while its query loads', () => {
        renderHome({ isLoadingContinueWatching: true });
        expect(screen.getAllByTestId('video-section-skeleton')).toHaveLength(1);
    });

    it('shows a skeleton for the recent rail while the catalog loads', () => {
        renderHome({ isLoadingHomeCatalog: true });
        expect(screen.getAllByTestId('video-section-skeleton')).toHaveLength(1);
    });
});
