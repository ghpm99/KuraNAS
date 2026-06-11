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
	createAllowedIP,
	deleteAllowedIP,
	getAllowedIPs,
	getClientIP,
	updateAllowedIP,
} from './accessControl';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

const sampleEntry = {
	id: 1,
	cidr: '192.168.1.10/32',
	label: 'notebook',
	enabled: true,
	created_at: '2026-06-11T10:00:00Z',
};

describe('service/accessControl', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('lists allowed IPs', async () => {
		mockedApi.get.mockResolvedValue({ data: [sampleEntry] });

		const result = await getAllowedIPs();

		expect(mockedApi.get).toHaveBeenCalledWith('/access-control/ips');
		expect(result).toEqual([sampleEntry]);
	});

	it('creates an allowed IP', async () => {
		mockedApi.post.mockResolvedValue({ data: sampleEntry });

		const result = await createAllowedIP({ cidr: '192.168.1.10', label: 'notebook' });

		expect(mockedApi.post).toHaveBeenCalledWith('/access-control/ips', {
			cidr: '192.168.1.10',
			label: 'notebook',
		});
		expect(result).toEqual(sampleEntry);
	});

	it('updates an allowed IP', async () => {
		const disabled = { ...sampleEntry, enabled: false };
		mockedApi.put.mockResolvedValue({ data: disabled });

		const result = await updateAllowedIP(1, { enabled: false });

		expect(mockedApi.put).toHaveBeenCalledWith('/access-control/ips/1', { enabled: false });
		expect(result).toEqual(disabled);
	});

	it('deletes an allowed IP', async () => {
		mockedApi.delete.mockResolvedValue({});

		await deleteAllowedIP(1);

		expect(mockedApi.delete).toHaveBeenCalledWith('/access-control/ips/1');
	});

	it('gets the requester client IP', async () => {
		mockedApi.get.mockResolvedValue({ data: { ip: '192.168.1.77' } });

		const result = await getClientIP();

		expect(mockedApi.get).toHaveBeenCalledWith('/access-control/client-ip');
		expect(result).toEqual({ ip: '192.168.1.77' });
	});
});
