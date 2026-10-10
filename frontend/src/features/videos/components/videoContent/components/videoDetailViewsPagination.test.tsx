import { fireEvent, render, screen } from '@testing-library/react';
import VideoContextDetailView from './VideoContextDetailView';
import VideoPlaylistDetailView from './VideoPlaylistDetailView';
import VideoSeriesDetailView from './VideoSeriesDetailView';
import type { VideoPlaylistDto, VideoPlaylistItemDto } from '@/service/videoPlayback';

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const buildItem = (order: number, name: string): VideoPlaylistItemDto => ({
    id: order + 1,
    order_index: order,
    source_kind: 'auto',
    status: 'not_started',
    progress_pct: 0,
    video: {
        id: 500 + order,
        name,
        path: `/series/${name}`,
        parent_path: '/series',
        format: 'mkv',
        size: 10,
    },
});

const buildPlaylist = (items: VideoPlaylistItemDto[], itemCount: number): VideoPlaylistDto => ({
    id: 1,
    type: 'series',
    source_path: '/series',
    name: 'Show',
    is_hidden: false,
    is_auto: true,
    group_mode: 'prefix',
    classification: 'series',
    item_count: itemCount,
    cover_video_id: null,
    created_at: '2026-03-14T00:00:00Z',
    updated_at: '2026-03-14T00:00:00Z',
    last_played_at: null,
    items,
});

const firstPageItems = [buildItem(0, 'Show S01E01.mkv'), buildItem(1, 'Show S01E02.mkv')];
const secondPageItems = [buildItem(2, 'Show S02E01.mkv'), buildItem(3, 'Show S02E02.mkv')];

const renderPlaylistView = (
    playlist: VideoPlaylistDto,
    paging: { hasMoreItems?: boolean; onLoadMoreItems?: () => void } = {}
) => (
    <VideoPlaylistDetailView
        playlist={playlist}
        isRenaming={false}
        isRemoving={false}
        isReordering={false}
        onBack={jest.fn()}
        onOpenVideo={jest.fn()}
        onRename={jest.fn()}
        onRemoveVideo={jest.fn()}
        onMoveItem={jest.fn()}
        {...paging}
    />
);

describe('detail views item paging', () => {
    it.each([
        [
            'playlist',
            (
                playlist: VideoPlaylistDto,
                paging: { hasMoreItems?: boolean; onLoadMoreItems?: () => void }
            ) => renderPlaylistView(playlist, paging),
        ],
        [
            'series',
            (
                playlist: VideoPlaylistDto,
                paging: { hasMoreItems?: boolean; onLoadMoreItems?: () => void }
            ) => (
                <VideoSeriesDetailView
                    playlist={playlist}
                    onBack={jest.fn()}
                    onOpenVideo={jest.fn()}
                    {...paging}
                />
            ),
        ],
        [
            'context',
            (
                playlist: VideoPlaylistDto,
                paging: { hasMoreItems?: boolean; onLoadMoreItems?: () => void }
            ) => (
                <VideoContextDetailView
                    playlist={playlist}
                    onBack={jest.fn()}
                    onOpenVideo={jest.fn()}
                    {...paging}
                />
            ),
        ],
    ])(
        '%s view hides load more on the last page and fetches the next page otherwise',
        (_viewName, buildView) => {
            const playlist = buildPlaylist(firstPageItems, 4);
            const { rerender } = render(buildView(playlist, { hasMoreItems: false }));

            expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).not.toBeInTheDocument();

            const onLoadMoreItems = jest.fn();
            rerender(buildView(playlist, { hasMoreItems: true, onLoadMoreItems }));
            fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

            expect(onLoadMoreItems).toHaveBeenCalledTimes(1);
        }
    );

    it('series view groups seasons incrementally as further pages arrive', () => {
        const buildSeries = (items: VideoPlaylistItemDto[]) => (
            <VideoSeriesDetailView
                playlist={buildPlaylist(items, 4)}
                onBack={jest.fn()}
                onOpenVideo={jest.fn()}
                hasMoreItems={items.length < 4}
                onLoadMoreItems={jest.fn()}
            />
        );
        const { rerender } = render(buildSeries(firstPageItems));

        expect(screen.queryByRole('button', { name: /VIDEO_DETAIL_SEASON_LABEL/ })).toBeNull();
        expect(screen.getByText('S01E01')).toBeInTheDocument();
        expect(screen.queryByText('S02E01')).not.toBeInTheDocument();

        rerender(buildSeries([...firstPageItems, ...secondPageItems]));

        expect(screen.queryAllByRole('button', { name: 'LOAD_MORE' })).toHaveLength(0);
        const seasonChips = screen
            .getAllByRole('button')
            .filter((button) => button.textContent === 'VIDEO_DETAIL_SEASON_LABEL');
        expect(seasonChips).toHaveLength(2);

        fireEvent.click(seasonChips[1] as HTMLElement);

        expect(screen.getByText('S02E01')).toBeInTheDocument();
        expect(screen.queryByText('S01E01')).not.toBeInTheDocument();
    });

    it('series view keeps appending episodes to the season already open', () => {
        const pageOne = [buildItem(0, 'Show S01E01.mkv'), buildItem(1, 'Show S01E02.mkv')];
        const pageTwo = [buildItem(2, 'Show S01E03.mkv')];
        const { rerender } = render(
            <VideoSeriesDetailView
                playlist={buildPlaylist(pageOne, 3)}
                onBack={jest.fn()}
                onOpenVideo={jest.fn()}
                hasMoreItems
            />
        );

        expect(screen.queryByText('S01E03')).not.toBeInTheDocument();

        rerender(
            <VideoSeriesDetailView
                playlist={buildPlaylist([...pageOne, ...pageTwo], 3)}
                onBack={jest.fn()}
                onOpenVideo={jest.fn()}
            />
        );

        expect(screen.getByText('S01E03')).toBeInTheDocument();
        expect(screen.getByText('S01E01')).toBeInTheDocument();
    });
});
