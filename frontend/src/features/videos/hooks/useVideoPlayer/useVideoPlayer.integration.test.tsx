import { act, renderHook } from '@testing-library/react';
import useVideoPlayer from './useVideoPlayer';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real useVideoPlayer +
// service/videoPlayback.ts run, so each playback command asserts the exact
// endpoint/payload the backend video playback handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; put: jest.Mock };

const session = {
	playlist: {
		id: 2,
		type: 'series',
		source_path: '/',
		name: 'Série',
		is_hidden: false,
		is_auto: false,
		group_mode: 'single',
		classification: 'personal',
		item_count: 1,
		cover_video_id: 5,
		created_at: '',
		updated_at: '',
		last_played_at: null,
		items: [],
	},
	playback_state: {
		id: 1,
		client_id: 'client',
		playlist_id: 2,
		video_id: 5,
		current_time: 0,
		duration: 100,
		is_paused: false,
		completed: false,
	},
};

describe('features/videos/useVideoPlayer (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.post.mockResolvedValue({ data: session });
		mockedApi.put.mockResolvedValue({ data: session.playback_state });
		mockedApi.get.mockResolvedValue({ data: session });
	});

	it('playVideo POSTs /video/playback/start with video_id and playlist_id', async () => {
		const { result } = renderHook(() => useVideoPlayer({ videoId: '5', playlistId: 2 }));

		await act(async () => {
			await result.current.playVideo();
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/video/playback/start', {
			video_id: 5,
			playlist_id: 2,
		});
	});

	it('nextVideo POSTs /video/playback/next', async () => {
		const { result } = renderHook(() => useVideoPlayer({ videoId: '5', playlistId: 2 }));

		await act(async () => {
			await result.current.nextVideo();
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/video/playback/next');
	});

	it('previousVideo POSTs /video/playback/previous', async () => {
		const { result } = renderHook(() => useVideoPlayer({ videoId: '5', playlistId: 2 }));

		await act(async () => {
			await result.current.previousVideo();
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/video/playback/previous');
	});

	it('onVideoEnded PUTs /video/playback/state with the playback fields', async () => {
		const { result } = renderHook(() => useVideoPlayer({ videoId: '5', playlistId: 2 }));

		await act(async () => {
			await result.current.playVideo();
		});
		await act(async () => {
			await result.current.onVideoEnded();
		});

		expect(mockedApi.put).toHaveBeenCalledWith(
			'/video/playback/state',
			expect.objectContaining({ playlist_id: 2, video_id: 5, completed: true, is_paused: true })
		);
	});
});
