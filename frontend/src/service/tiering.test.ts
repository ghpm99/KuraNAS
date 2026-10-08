jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
		post: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getTieringSettings,
	promoteFileToHot,
	getTieringStatus,
	getTieringUsage,
	updateTieringSettings,
} from './tiering';
import type { TieringSettings } from '@/types/tiering';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
	post: jest.Mock;
};

const sampleSettings: TieringSettings = {
	enabled: true,
	cold_dir_path: '/mnt/cold',
	min_age_days: 90,
	min_size_bytes: 1048576,
	interval_hours: 24,
};

describe('service/tiering', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('promotes a cold file to the hot disk', async () => {
		const location = { file_id: 9, tier: 'hot', disk_path: '/data/a.pdf' };
		mockedApi.post.mockResolvedValue({ data: location });
		const result = await promoteFileToHot(9);
		expect(mockedApi.post).toHaveBeenCalledWith('/tiering/promote/9');
		expect(result).toEqual(location);
	});

	it('loads the tiering settings', async () => {
		mockedApi.get.mockResolvedValue({ data: sampleSettings });
		const result = await getTieringSettings();
		expect(mockedApi.get).toHaveBeenCalledWith('/tiering/settings');
		expect(result).toEqual(sampleSettings);
	});

	it('saves the tiering settings', async () => {
		mockedApi.put.mockResolvedValue({ data: sampleSettings });
		const result = await updateTieringSettings(sampleSettings);
		expect(mockedApi.put).toHaveBeenCalledWith('/tiering/settings', sampleSettings);
		expect(result).toEqual(sampleSettings);
	});

	it('loads the tiering status', async () => {
		const status = {
			enabled: true,
			has_run: true,
			status: 'completed',
			started_at: '2026-06-12T10:00:00Z',
			ended_at: '2026-06-12T10:05:00Z',
			last_error: '',
		};
		mockedApi.get.mockResolvedValue({ data: status });
		const result = await getTieringStatus();
		expect(mockedApi.get).toHaveBeenCalledWith('/tiering/status');
		expect(result).toEqual(status);
	});

	it('loads the tier usage split', async () => {
		const usage = { hot_files: 10, hot_bytes: 1000, cold_files: 4, cold_bytes: 400 };
		mockedApi.get.mockResolvedValue({ data: usage });
		const result = await getTieringUsage();
		expect(mockedApi.get).toHaveBeenCalledWith('/tiering/usage');
		expect(result).toEqual(usage);
	});
});
