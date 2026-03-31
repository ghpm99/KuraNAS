jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import { getLibraries, updateLibrary } from './libraries';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

describe('service/libraries', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets libraries', async () => {
		const response = [
			{ category: 'images', path: '/data/Imagens' },
			{ category: 'videos', path: '/data/Videos' },
		];
		mockedApi.get.mockResolvedValue({ data: response });

		const result = await getLibraries();

		expect(mockedApi.get).toHaveBeenCalledWith('/libraries');
		expect(result).toEqual(response);
	});

	it('updates a library path', async () => {
		const response = { category: 'music', path: '/data/Musicas' };
		mockedApi.put.mockResolvedValue({ data: response });

		const result = await updateLibrary('music', { path: '/data/Musicas' });

		expect(mockedApi.put).toHaveBeenCalledWith('/libraries/music', {
			path: '/data/Musicas',
		});
		expect(result).toEqual(response);
	});
});
