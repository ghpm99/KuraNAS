jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getAboutConfiguration,
	getSettingsConfiguration,
	getTranslations,
	updateSettingsConfiguration,
} from './configuration';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

describe('service/configuration', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets about configuration', async () => {
		const payload = { version: '1.0.0' };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getAboutConfiguration();

		expect(mockedApi.get).toHaveBeenCalledWith('/configuration/about');
		expect(result).toEqual(payload);
	});

	it('gets translations', async () => {
		const payload = { SETTINGS: 'Settings' };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getTranslations();

		expect(mockedApi.get).toHaveBeenCalledWith('/configuration/translation');
		expect(result).toEqual(payload);
	});

	it('gets settings configuration', async () => {
		const payload = {
			library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 4 },
			appearance: { accent_color: 'violet', reduce_motion: false },
			language: { current: 'en-US', available: ['en-US'] },
		};
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getSettingsConfiguration();

		expect(mockedApi.get).toHaveBeenCalledWith('/configuration/settings');
		expect(result).toEqual(payload);
	});

	it('updates settings configuration', async () => {
		const request = {
			library: { watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 8 },
			appearance: { accent_color: 'cyan' as const, reduce_motion: true },
			language: { current: 'pt-BR' },
		};
		const payload = {
			...request,
			library: { ...request.library, runtime_root_path: '/data' },
			indexing: { ...request.indexing, workers_enabled: true },
			language: { ...request.language, available: ['en-US', 'pt-BR'] },
		};
		mockedApi.put.mockResolvedValue({ data: payload });

		const result = await updateSettingsConfiguration(request);

		expect(mockedApi.put).toHaveBeenCalledWith('/configuration/settings', request);
		expect(result).toEqual(payload);
	});
});
