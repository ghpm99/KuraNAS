jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
	},
}));

import { apiBase } from './index';
import { getYtDlpStatus, updateYtDlp } from './ingest';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
};

describe('service/ingest', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets the yt-dlp status', async () => {
		const status = {
			installed: true,
			current_version: '2024.08.06',
			latest_version: '2024.09.01',
			update_available: true,
			release_url: 'http://x',
			release_date: '2024-09-01',
		};
		mockedApi.get.mockResolvedValue({ data: status });

		const result = await getYtDlpStatus();

		expect(mockedApi.get).toHaveBeenCalledWith('/ingest/ytdlp/status');
		expect(result).toEqual(status);
	});

	it('triggers a yt-dlp update', async () => {
		mockedApi.post.mockResolvedValue({ data: { message: 'ok' } });

		await updateYtDlp();

		expect(mockedApi.post).toHaveBeenCalledWith('/ingest/ytdlp/update');
	});
});
