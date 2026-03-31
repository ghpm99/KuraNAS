jest.mock('./index', () => ({
	apiBase: {
		post: jest.fn(),
	},
}));

import { apiBase } from './index';
import { completeTakeoutUpload, initTakeoutUpload, uploadTakeoutChunk } from './takeout';

const mockedApi = apiBase as unknown as {
	post: jest.Mock;
};

describe('service/takeout', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('initializes takeout upload', async () => {
		mockedApi.post.mockResolvedValue({ data: { upload_id: 'u1', chunk_size: 1024 } });
		const result = await initTakeoutUpload('takeout.zip', 100);
		expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/init', {
			file_name: 'takeout.zip',
			size: 100,
		});
		expect(result.upload_id).toBe('u1');
	});

	it('uploads a chunk', async () => {
		mockedApi.post.mockResolvedValue({ data: { received: true } });
		const result = await uploadTakeoutChunk('u1', new Blob(['abc']), 0);
		expect(mockedApi.post).toHaveBeenCalledWith(
			'/takeout/upload/chunk',
			expect.any(FormData),
			expect.objectContaining({
				headers: expect.objectContaining({
					'Content-Type': 'multipart/form-data',
				}),
			})
		);
		expect(result.received).toBe(true);
	});

	it('completes takeout upload', async () => {
		mockedApi.post.mockResolvedValue({ data: { job_id: 20, message: 'ok' } });
		const result = await completeTakeoutUpload('u1');
		expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/complete', {
			upload_id: 'u1',
		});
		expect(result.job_id).toBe(20);
	});
});
