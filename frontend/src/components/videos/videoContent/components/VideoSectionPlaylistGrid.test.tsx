import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VideoSectionPlaylistGrid, { VideoSectionActionLink } from './VideoSectionPlaylistGrid';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => {
			const map: Record<string, string> = {
				VIDEO_OPEN_SECTION: 'Open Section',
				TITLE_KEY: 'Section Title',
				DESCRIPTION_KEY: 'Section Description',
				EMPTY_KEY: 'Nothing here',
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

const playlist: VideoPlaylistDto = {
	id: 8,
	type: 'series',
	source_path: '/shows',
	name: 'Playlist 8',
	is_hidden: false,
	is_auto: true,
	group_mode: 'prefix',
	classification: 'series',
	item_count: 2,
	cover_video_id: 80,
	created_at: '2026-03-16T10:00:00Z',
	updated_at: '2026-03-16T10:00:00Z',
	last_played_at: null,
	items: [],
};

describe('videos/videoContent/VideoSectionPlaylistGrid', () => {
	it('renders the empty state and custom action', () => {
		render(
			<MemoryRouter>
				<VideoSectionPlaylistGrid
					titleKey='TITLE_KEY'
					descriptionKey='DESCRIPTION_KEY'
					emptyKey='EMPTY_KEY'
					playlists={[]}
					onSelectPlaylist={jest.fn()}
					onPlayVideo={jest.fn()}
					action={<VideoSectionActionLink to='/videos/library' />}
				/>
			</MemoryRouter>,
		);

		expect(screen.getByText('Section Title')).toBeInTheDocument();
		expect(screen.getByText('Section Description')).toBeInTheDocument();
		expect(screen.getByText('Nothing here')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Open Section' })).toHaveAttribute('href', '/videos/library');
	});

	it('renders playlists and forwards card callbacks', () => {
		const onSelectPlaylist = jest.fn();
		const onPlayVideo = jest.fn();

		render(
			<VideoSectionPlaylistGrid
				titleKey='TITLE_KEY'
				descriptionKey='DESCRIPTION_KEY'
				emptyKey='EMPTY_KEY'
				playlists={[playlist]}
				onSelectPlaylist={onSelectPlaylist}
				onPlayVideo={onPlayVideo}
				badge='Resume'
			/>,
		);

		expect(screen.getByText('Resume')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'select-8' }));
		fireEvent.click(screen.getByRole('button', { name: 'play-8' }));

		expect(onSelectPlaylist).toHaveBeenCalledWith(playlist);
		expect(onPlayVideo).toHaveBeenCalledWith(80, 8);
	});
});
