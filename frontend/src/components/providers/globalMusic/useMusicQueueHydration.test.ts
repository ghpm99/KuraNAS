import { act, renderHook } from '@testing-library/react';
import useMusicQueueHydration from './useMusicQueueHydration';
import { getPlayerState } from '@/service/playerState';
import { getNowPlayingPlaylist, getPlaylistTracks } from '@/service/playlist';

jest.mock('@/service/playerState', () => ({
	getPlayerState: jest.fn(),
}));

jest.mock('@/service/playlist', () => ({
	getNowPlayingPlaylist: jest.fn(),
	getPlaylistTracks: jest.fn(),
}));

describe('useMusicQueueHydration', () => {
	const callbacks = {
		setQueue: jest.fn(),
		setCurrentIndex: jest.fn(),
		setPlaybackContext: jest.fn(),
	};

	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('hydrates queue when enabled and valid data is available', async () => {
		(getPlayerState as jest.Mock).mockResolvedValue({
			current_file_id: 10,
		});
		(getNowPlayingPlaylist as jest.Mock).mockResolvedValue({
			id: 5,
			track_count: 1,
			name: 'Playlist',
		});
		(getPlaylistTracks as jest.Mock).mockResolvedValue({
			items: [
				{
					file: {
						id: 10,
						name: 'track',
						path: '/music',
						type: 0,
						format: 'mp3',
						size: 100,
						updated_at: '',
						created_at: '',
						deleted_at: '',
						last_interaction: '',
						last_backup: '',
						check_sum: '',
						directory_content_count: 0,
						starred: false,
					},
				},
			],
		});

		renderHook(() => useMusicQueueHydration(true, callbacks));

		await act(async () => {
			await Promise.resolve();
		});

		expect(callbacks.setQueue).toHaveBeenCalled();
		expect(callbacks.setCurrentIndex).toHaveBeenCalledWith(0);
		expect(callbacks.setPlaybackContext).toHaveBeenCalled();
	});

	it('does nothing when disabled', async () => {
		renderHook(() => useMusicQueueHydration(false, callbacks));
		await act(async () => {
			await Promise.resolve();
		});
		expect(callbacks.setQueue).not.toHaveBeenCalled();
		expect(callbacks.setCurrentIndex).not.toHaveBeenCalled();
		expect(callbacks.setPlaybackContext).not.toHaveBeenCalled();
	});
});
