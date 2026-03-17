import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlaylistCard from './VideoPlaylistCard';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			if (key === 'VIDEO_PREVIEW_ALT') return `Preview ${params?.name ?? ''}`.trim();
			if (key === 'VIDEO_PLAYLIST_ITEM_COUNT') return `${params?.count ?? '0'} items`;
			if (key === 'VIDEO_PLAY') return 'Play';
			return key;
		},
	}),
}));

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const createPlaylist = (overrides?: Partial<VideoPlaylistDto>): VideoPlaylistDto => ({
	id: 7,
	type: 'series',
	source_path: '/series/show',
	name: 'Showcase',
	is_hidden: false,
	is_auto: true,
	group_mode: 'prefix',
	classification: 'series',
	item_count: 4,
	cover_video_id: 70,
	created_at: '2026-03-16T10:00:00Z',
	updated_at: '2026-03-16T10:00:00Z',
	last_played_at: null,
	items: [],
	...overrides,
});

describe('videos/videoContent/VideoPlaylistCard', () => {
	it('renders artwork, badge and forwards selection/play callbacks', () => {
		const onSelect = jest.fn();
		const onPlay = jest.fn();
		const playlist = createPlaylist();

		render(
			<VideoPlaylistCard
				playlist={playlist}
				onSelect={onSelect}
				onPlay={onPlay}
				focusVideoId={99}
				badge='Resume'
			/>,
		);

		expect(screen.getByText('4 items')).toBeInTheDocument();
		expect(screen.getByText('SERIES')).toBeInTheDocument();
		expect(screen.getByText('Resume')).toBeInTheDocument();
		expect(screen.getByRole('img', { name: 'Showcase' })).toHaveAttribute(
			'src',
			'http://localhost:8000/api/v1/files/video-thumbnail/99?width=640&height=360',
		);
		expect(screen.getByRole('img', { name: 'Preview Showcase' })).toHaveAttribute(
			'src',
			'http://localhost:8000/api/v1/files/video-preview/99?width=640&height=360',
		);

		fireEvent.click(screen.getByRole('button', { name: /Showcase/i }));
		fireEvent.click(screen.getByRole('button', { name: 'Play' }));

		expect(onSelect).toHaveBeenCalledWith(playlist);
		expect(onPlay).toHaveBeenCalledWith(99, 7);
	});

	it('falls back to personal classification and disables play when there is no cover', () => {
		const onPlay = jest.fn();

		render(
			<VideoPlaylistCard
				playlist={createPlaylist({
					classification: '' as unknown as VideoPlaylistDto['classification'],
					cover_video_id: null,
				})}
				onSelect={jest.fn()}
				onPlay={onPlay}
			/>,
		);

		expect(screen.getByText('PERSONAL')).toBeInTheDocument();
		expect(screen.queryByRole('img', { name: 'Showcase' })).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Play' })).toBeDisabled();
		expect(onPlay).not.toHaveBeenCalled();
	});
});
