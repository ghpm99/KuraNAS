import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SettingsProvider from './index';
import { useSettings } from './settingsContext';
import type { UpdateSettingsConfigurationRequest } from '@/service/configuration';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real SettingsProvider + useSettings +
// service/configuration.ts run, so saving asserts PUT /configuration/settings
// with the exact nested payload the backend configuration handler decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const request: UpdateSettingsConfigurationRequest = {
	indexing: { scan_on_startup: true, extract_metadata: false, generate_previews: true },
	captures: { save_path: '/data/Capturas' },
	ai: { image_classification: true },
	players: {
		remember_music_queue: true,
		remember_video_progress: false,
		autoplay_next_video: true,
		image_slideshow_seconds: 7,
	},
	appearance: { accent_color: 'cyan', reduce_motion: true },
	language: { current: 'pt-BR' },
};

const Consumer = () => {
	const { saveSettings } = useSettings();
	return <button onClick={() => void saveSettings(request)}>save</button>;
};

const serverConfig = {
	indexing: { workers_enabled: false, scan_on_startup: true, extract_metadata: true, generate_previews: true },
	captures: { save_path: '', default_path: '', storage_roots: [] },
	ai: { image_classification: true },
	players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 4 },
	appearance: { accent_color: 'violet', reduce_motion: false },
	language: { current: 'en-US', available: ['en-US'] },
};

describe('components/providers/settingsProvider (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: serverConfig });
		mockedApi.put.mockResolvedValue({ data: serverConfig });
	});

	it('saveSettings issues PUT /configuration/settings with the request payload', async () => {
		const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
		render(
			<QueryClientProvider client={client}>
				<SettingsProvider>
					<Consumer />
				</SettingsProvider>
			</QueryClientProvider>
		);

		fireEvent.click(screen.getByText('save'));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/configuration/settings', request)
		);
	});
});
