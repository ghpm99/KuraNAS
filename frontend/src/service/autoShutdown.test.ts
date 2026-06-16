jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getAutoShutdownSettings,
	getSuggestedShutdownTime,
	updateAutoShutdownSettings,
} from './autoShutdown';
import type { AutoShutdownSettings } from '@/types/autoShutdown';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

const sampleSettings: AutoShutdownSettings = {
	enabled: true,
	time: '03:00',
	grace_period_seconds: 60,
};

describe('service/autoShutdown', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('loads the auto-shutdown settings', async () => {
		mockedApi.get.mockResolvedValue({ data: sampleSettings });
		const result = await getAutoShutdownSettings();
		expect(mockedApi.get).toHaveBeenCalledWith('/auto-shutdown/settings');
		expect(result).toEqual(sampleSettings);
	});

	it('saves the auto-shutdown settings', async () => {
		mockedApi.put.mockResolvedValue({ data: sampleSettings });
		const result = await updateAutoShutdownSettings(sampleSettings);
		expect(mockedApi.put).toHaveBeenCalledWith('/auto-shutdown/settings', sampleSettings);
		expect(result).toEqual(sampleSettings);
	});

	it('loads the suggested shutdown time', async () => {
		const suggestion = { available: true, time: '02:30', sample_size: 7 };
		mockedApi.get.mockResolvedValue({ data: suggestion });
		const result = await getSuggestedShutdownTime();
		expect(mockedApi.get).toHaveBeenCalledWith('/auto-shutdown/suggested-time');
		expect(result).toEqual(suggestion);
	});
});
