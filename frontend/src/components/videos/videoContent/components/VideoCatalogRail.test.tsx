import { fireEvent, render, screen } from '@testing-library/react';
import VideoCatalogRail from './VideoCatalogRail';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => {
			const map: Record<string, string> = {
				VIDEO_PLAY: 'Play',
				VIDEO_STATUS_COMPLETED: 'Completed',
				VIDEO_STATUS_IN_PROGRESS: 'In Progress',
				VIDEO_STATUS_NOT_STARTED: 'Not Started',
				TITLE_KEY: 'Catalog Title',
				DESCRIPTION_KEY: 'Catalog Description',
			};
			return map[key] ?? key;
		},
	}),
}));

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const items = [
	{
		video: {
			id: 1,
			name: 'Episode 1',
			path: '/videos/episode-1.mp4',
			parent_path: '/videos/season-1',
			format: 'mp4',
			size: 123,
		},
		status: 'completed' as const,
		progress_pct: 100,
	},
	{
		video: {
			id: 2,
			name: 'Episode 2',
			path: '/videos/episode-2.mp4',
			parent_path: '/videos/season-1',
			format: 'mkv',
			size: 124,
		},
		status: 'in_progress' as const,
		progress_pct: 40,
	},
	{
		video: {
			id: 3,
			name: 'Episode 3',
			path: '/videos/episode-3.mp4',
			parent_path: '/videos/season-1',
			format: 'avi',
			size: 125,
		},
		status: 'not_started' as const,
		progress_pct: 0,
	},
];

describe('videos/videoContent/VideoCatalogRail', () => {
	it('returns null when there are no catalog items', () => {
		const { container } = render(
			<VideoCatalogRail titleKey='TITLE_KEY' descriptionKey='DESCRIPTION_KEY' items={[]} onPlayVideo={jest.fn()} />,
		);

		expect(container.firstChild).toBeNull();
	});

	it('renders all status variants and forwards play actions', () => {
		const onPlayVideo = jest.fn();

		render(
			<VideoCatalogRail
				titleKey='TITLE_KEY'
				descriptionKey='DESCRIPTION_KEY'
				items={items}
				onPlayVideo={onPlayVideo}
			/>,
		);

		expect(screen.getByText('Catalog Title')).toBeInTheDocument();
		expect(screen.getByText('Catalog Description')).toBeInTheDocument();
		expect(screen.getByText('Completed')).toBeInTheDocument();
		expect(screen.getByText('In Progress')).toBeInTheDocument();
		expect(screen.getByText('Not Started')).toBeInTheDocument();
		expect(screen.getByText('MP4')).toBeInTheDocument();
		expect(screen.getByText('MKV')).toBeInTheDocument();
		expect(screen.getByText('AVI')).toBeInTheDocument();

		const image = screen.getByRole('img', { name: 'Episode 1' });
		expect(image).toHaveAttribute('src', 'http://localhost:8000/api/v1/files/video-thumbnail/1?width=480&height=270');

		fireEvent.click(screen.getAllByRole('button', { name: 'Play' })[1]!);
		expect(onPlayVideo).toHaveBeenCalledWith(2, null);
	});
});
