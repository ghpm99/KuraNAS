import { fireEvent, render, screen } from '@testing-library/react';
import VideoCatalogSections from './VideoCatalogSections';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => {
			const map: Record<string, string> = {
				VIDEO_CONTINUE_WATCHING: 'Continue Watching',
				VIDEO_RECENT_PLAYLISTS_DESC: 'Recent playlists description',
				VIDEO_NO_RECENT_PLAYLISTS: 'No recent playlists',
				VIDEO_PLAYLISTS: 'Playlists',
				VIDEO_PLAYLISTS_DESC: 'Playlists description',
				VIDEO_CONTINUE_BADGE_RESUME: 'Resume',
			};
			return map[key] ?? key;
		},
	}),
}));

jest.mock('./VideoPlaylistCard', () => ({
	__esModule: true,
	default: ({
		playlist,
		onSelect,
		onPlay,
		badge,
	}: {
		playlist: VideoPlaylistDto;
		onSelect: (playlist: VideoPlaylistDto) => void;
		onPlay: (videoId: number, playlistId?: number | null) => void;
		badge?: string;
	}) => (
		<div data-testid={`playlist-${playlist.id}`}>
			<span>{playlist.name}</span>
			{badge && <span>{badge}</span>}
			<button type='button' onClick={() => onSelect(playlist)}>
				select-{playlist.id}
			</button>
			<button type='button' onClick={() => onPlay(playlist.cover_video_id ?? 0, playlist.id)}>
				play-{playlist.id}
			</button>
		</div>
	),
}));

const createPlaylist = (id: number, name: string): VideoPlaylistDto => ({
	id,
	type: 'series',
	source_path: `/videos/${id}`,
	name,
	is_hidden: false,
	is_auto: true,
	group_mode: 'prefix',
	classification: 'series',
	item_count: 3,
	cover_video_id: id * 10,
	created_at: '2026-03-16T10:00:00Z',
	updated_at: '2026-03-16T10:00:00Z',
	last_played_at: null,
	items: [],
});

describe('videos/videoContent/VideoCatalogSections', () => {
	it('renders the empty continue section and skips empty grouped categories', () => {
		render(
			<VideoCatalogSections
				continuePlaylists={[]}
				groupedPlaylists={{
					classificationTitle: {
						series: 'Series Title',
					},
					grouped: {
						series: [createPlaylist(1, 'Alpha')],
						clips: [createPlaylist(2, 'Beta')],
						empty: [],
					},
				}}
				onSelectPlaylist={jest.fn()}
				onPlayVideo={jest.fn()}
			/>,
		);

		expect(screen.getByText('No recent playlists')).toBeInTheDocument();
		expect(screen.getByText('Series Title')).toBeInTheDocument();
		expect(screen.getByText('clips')).toBeInTheDocument();
		expect(screen.queryByText('empty')).not.toBeInTheDocument();
	});

	it('renders continue cards with badge and forwards select/play handlers', () => {
		const onSelectPlaylist = jest.fn();
		const onPlayVideo = jest.fn();
		const playlist = createPlaylist(5, 'Gamma');

		render(
			<VideoCatalogSections
				continuePlaylists={[playlist]}
				groupedPlaylists={{
					classificationTitle: {},
					grouped: {},
				}}
				onSelectPlaylist={onSelectPlaylist}
				onPlayVideo={onPlayVideo}
			/>,
		);

		expect(screen.getByText('Resume')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'select-5' }));
		fireEvent.click(screen.getByRole('button', { name: 'play-5' }));

		expect(onSelectPlaylist).toHaveBeenCalledWith(playlist);
		expect(onPlayVideo).toHaveBeenCalledWith(50, 5);
	});
});
