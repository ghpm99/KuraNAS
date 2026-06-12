jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
		put: jest.fn(),
		delete: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	createStorageRoot,
	deleteStorageRoot,
	getStorageRoots,
	updateStorageRoot,
} from './storageRoots';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

const sampleRoot = {
	id: 1,
	path: '/mnt/dados',
	label: 'Dados',
	enabled: true,
	created_at: '2026-06-12T10:00:00Z',
};

describe('service/storageRoots', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('lists storage roots', async () => {
		mockedApi.get.mockResolvedValue({ data: [sampleRoot] });

		const result = await getStorageRoots();

		expect(mockedApi.get).toHaveBeenCalledWith('/storage-roots');
		expect(result).toEqual([sampleRoot]);
	});

	it('creates a storage root', async () => {
		mockedApi.post.mockResolvedValue({ data: sampleRoot });

		const result = await createStorageRoot({ path: '/mnt/dados', label: 'Dados' });

		expect(mockedApi.post).toHaveBeenCalledWith('/storage-roots', {
			path: '/mnt/dados',
			label: 'Dados',
		});
		expect(result).toEqual(sampleRoot);
	});

	it('updates a storage root', async () => {
		const disabled = { ...sampleRoot, enabled: false };
		mockedApi.put.mockResolvedValue({ data: disabled });

		const result = await updateStorageRoot(1, { enabled: false });

		expect(mockedApi.put).toHaveBeenCalledWith('/storage-roots/1', { enabled: false });
		expect(result).toEqual(disabled);
	});

	it('deletes a storage root', async () => {
		mockedApi.delete.mockResolvedValue({});

		await deleteStorageRoot(1);

		expect(mockedApi.delete).toHaveBeenCalledWith('/storage-roots/1');
	});
});
