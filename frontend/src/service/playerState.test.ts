jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import { getPlayerState, updatePlayerState } from './playerState';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

describe('service/playerState', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets player state', async () => {
		const payload = { id: 1, client_id: 'c1' };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getPlayerState();

		expect(mockedApi.get).toHaveBeenCalledWith('/music/player-state/');
		expect(result).toEqual(payload);
	});

	it('updates player state', async () => {
		const request = { current_position: 10, volume: 0.7 };
		const payload = { id: 1, current_position: 10, volume: 0.7 };
		mockedApi.put.mockResolvedValue({ data: payload });

		const result = await updatePlayerState(request);

		expect(mockedApi.put).toHaveBeenCalledWith('/music/player-state/', request);
		expect(result).toEqual(payload);
	});
});

