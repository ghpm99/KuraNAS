import { act, renderHook } from '@testing-library/react';
import useMusicStateSync from './useMusicStateSync';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real useMusicStateSync +
// service/playerState.ts run. playerState has no button — it is a debounced
// background sync — so this asserts the debounced PUT /music/player-state/ with
// the exact body the backend music player-state handler decodes.
jest.mock('@/service', () => ({
	apiBase: { put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { put: jest.Mock };

describe('features/music/useMusicStateSync (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		jest.useFakeTimers();
		mockedApi.put.mockResolvedValue({ data: {} });
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('debounced syncState issues PUT /music/player-state/ with the assembled body', async () => {
		const { result } = renderHook(() =>
			useMusicStateSync({
				getCurrentTrackId: () => 5,
				getCurrentTime: () => 30,
				volume: 0.8,
				shuffle: false,
				repeatMode: 'none',
				playbackContext: { playlistId: 2 } as never,
			})
		);

		act(() => result.current.syncState());
		act(() => jest.advanceTimersByTime(2000));

		expect(mockedApi.put).toHaveBeenCalledWith('/music/player-state/', {
			playlist_id: 2,
			current_file_id: 5,
			current_position: 30,
			volume: 0.8,
			shuffle: false,
			repeat_mode: 'none',
		});
	});
});
