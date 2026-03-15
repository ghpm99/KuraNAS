import { searchGlobal } from './search';
import { apiBase } from '.';

jest.mock('.', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

const mockedApiGet = apiBase.get as jest.Mock;

describe('service/search', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('requests the global search endpoint with query and limit', async () => {
		mockedApiGet.mockResolvedValue({
			data: {
				query: 'mix',
				files: [],
				folders: [],
				artists: [],
				albums: [],
				playlists: [],
				videos: [],
				images: [],
			},
		});

		const result = await searchGlobal('mix', 8);

		expect(mockedApiGet).toHaveBeenCalledWith('/search/global', {
			params: {
				q: 'mix',
				limit: 8,
			},
		});
		expect(result.query).toBe('mix');
	});

	it('uses the default per-section limit when omitted', async () => {
		mockedApiGet.mockResolvedValue({
			data: {
				query: '',
				files: [],
				folders: [],
				artists: [],
				albums: [],
				playlists: [],
				videos: [],
				images: [],
			},
		});

		await searchGlobal('');

		expect(mockedApiGet).toHaveBeenCalledWith('/search/global', {
			params: {
				q: '',
				limit: 6,
			},
		});
	});
});
