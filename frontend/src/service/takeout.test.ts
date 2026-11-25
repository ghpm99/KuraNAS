jest.mock('./index', () => ({
	apiBase: {
		post: jest.fn(),
		get: jest.fn(),
		delete: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	cancelTakeoutUpload,
	completeTakeoutUpload,
	getTakeoutImport,
	getTakeoutUploadStatus,
	initTakeoutUpload,
	listTakeoutImports,
	uploadTakeoutChunk,
} from './takeout';

const mockedApi = apiBase as unknown as {
	post: jest.Mock;
	get: jest.Mock;
	delete: jest.Mock;
};

const takeoutImport = {
	id: 5,
	job_id: null,
	file_name: 'takeout.zip',
	size: 1000,
	state: 'processing',
	total_entries: 10,
	processed_entries: 4,
	imported: 2,
	already_existing: 1,
	ignored: 0,
	failed_entries: 1,
	error: '',
	created_at: '2026-11-20T10:00:00Z',
	ended_at: null,
};

describe('service/takeout', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('initializes takeout upload', async () => {
		mockedApi.post.mockResolvedValue({
			data: { upload_id: 'u1', chunk_size: 1024, received_size: 512, resumed: true },
		});
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

	it('exposes resume fields when initializing an upload', async () => {
		mockedApi.post.mockResolvedValue({
			data: { upload_id: 'u1', chunk_size: 1024, received_size: 512, resumed: true },
		});
		const result = await initTakeoutUpload('takeout.zip', 100);
		expect(result.received_size).toBe(512);
		expect(result.resumed).toBe(true);
	});

	it('reads the upload status', async () => {
		mockedApi.get.mockResolvedValue({
			data: { upload_id: 'u1', size: 100, received_size: 40 },
		});
		const result = await getTakeoutUploadStatus('u1');
		expect(mockedApi.get).toHaveBeenCalledWith('/takeout/upload/u1');
		expect(result).toEqual({ upload_id: 'u1', size: 100, received_size: 40 });
	});

	it('cancels an upload', async () => {
		mockedApi.delete.mockResolvedValue({ status: 204, data: '' });
		await expect(cancelTakeoutUpload('u1')).resolves.toBeUndefined();
		expect(mockedApi.delete).toHaveBeenCalledWith('/takeout/upload/u1');
	});

	it('lists imports with pagination params', async () => {
		const listing = {
			items: [takeoutImport],
			pagination: { page: 2, page_size: 10, has_next: false, has_prev: true },
		};
		mockedApi.get.mockResolvedValue({ data: listing });
		const result = await listTakeoutImports({ page: 2, pageSize: 10 });
		expect(mockedApi.get).toHaveBeenCalledWith('/takeout/imports', {
			params: { page: 2, page_size: 10 },
		});
		expect(result.items[0]?.job_id).toBeNull();
		expect(result.items[0]?.ended_at).toBeNull();
		expect(result.pagination.has_prev).toBe(true);
	});

	it('uses default pagination when listing imports', async () => {
		mockedApi.get.mockResolvedValue({ data: { items: [], pagination: {} } });
		await listTakeoutImports();
		expect(mockedApi.get).toHaveBeenCalledWith('/takeout/imports', {
			params: { page: 1, page_size: 20 },
		});
	});

	it('throws when the imports listing has no items array', async () => {
		mockedApi.get.mockResolvedValue({ data: { items: null } });
		await expect(listTakeoutImports()).rejects.toThrow();
		mockedApi.get.mockResolvedValue({ data: undefined });
		await expect(listTakeoutImports()).rejects.toThrow();
	});

	it('reads a single import', async () => {
		mockedApi.get.mockResolvedValue({ data: { ...takeoutImport, job_id: 7, state: 'done' } });
		const result = await getTakeoutImport(5);
		expect(mockedApi.get).toHaveBeenCalledWith('/takeout/imports/5');
		expect(result.job_id).toBe(7);
		expect(result.failed_entries).toBe(1);
	});
});
