import { fireEvent, render, screen } from '@testing-library/react';
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

    it('shows one error per playlist section with retry when playlists fail', () => {
        const retry = jest.fn();
        renderHome({
            playlistsFailure: { message: 'playlists down', retry },
            isLoadingContinueWatching: true,
            isLoadingHomeCatalog: true,
        });
        expect(screen.getAllByText('playlists down')).toHaveLength(5);
        expect(screen.getAllByTestId('video-section-skeleton')).toHaveLength(2);
        fireEvent.click(screen.getAllByRole('button')[0] as HTMLElement);
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('shows the continue watching error without affecting other sections', () => {
        const retry = jest.fn();
        renderHome({ continueWatchingFailure: { message: 'continue down', retry } });
        expect(screen.getAllByRole('alert')).toHaveLength(1);
        expect(screen.getByText('continue down')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /.+/ }));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('shows the recent rail error without affecting other sections', () => {
        const retry = jest.fn();
        renderHome({ homeCatalogFailure: { message: 'catalog down', retry } });
        expect(screen.getAllByRole('alert')).toHaveLength(1);
        expect(screen.getByText('catalog down')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /.+/ }));
        expect(retry).toHaveBeenCalledTimes(1);
    });
});
