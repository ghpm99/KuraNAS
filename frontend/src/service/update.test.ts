jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
	},
}));

import { apiBase } from './index';
import { applyUpdate, getUpdateStatus } from './update';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
};

describe('service/update', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets update status', async () => {
		const payload = { status: 'idle', available: false };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getUpdateStatus();

		expect(mockedApi.get).toHaveBeenCalledWith('/update/status');
		expect(result).toEqual(payload);
	});

	it('applies update', async () => {
		const payload = { started: true };
		mockedApi.post.mockResolvedValue({ data: payload });

		const result = await applyUpdate();

		expect(mockedApi.post).toHaveBeenCalledWith('/update/apply');
		expect(result).toEqual(payload);
	});
});

