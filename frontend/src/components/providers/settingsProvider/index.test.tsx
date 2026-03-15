import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SettingsProvider from './index';
import { useSettings } from './settingsContext';
import { getSettingsConfiguration, updateSettingsConfiguration } from '@/service/configuration';

jest.mock('@/service/configuration', () => ({
	getSettingsConfiguration: jest.fn(),
	updateSettingsConfiguration: jest.fn(),
}));

const mockedGetSettingsConfiguration = getSettingsConfiguration as jest.Mock;
const mockedUpdateSettingsConfiguration = updateSettingsConfiguration as jest.Mock;

const createWrapper = () => {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});

	return ({ children }: { children: React.ReactNode }) => (
		<QueryClientProvider client={queryClient}>
			<SettingsProvider>{children}</SettingsProvider>
		</QueryClientProvider>
	);
};

describe('components/providers/settingsProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		document.documentElement.removeAttribute('data-app-motion');
		document.documentElement.style.removeProperty('--app-color-primary');
	});

	it('loads settings and applies appearance runtime values', async () => {
		mockedGetSettingsConfiguration.mockResolvedValue({
			library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 8 },
			appearance: { accent_color: 'cyan', reduce_motion: true },
			language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
		});

		const { result } = renderHook(() => useSettings(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoading).toBe(false));

		expect(result.current.settings.language.current).toBe('pt-BR');
		expect(document.documentElement.style.getPropertyValue('--app-color-primary')).toBe('#06B6D4');
		expect(document.documentElement.dataset.appMotion).toBe('reduced');
	});

	it('saves settings through the mutation boundary', async () => {
		mockedGetSettingsConfiguration.mockResolvedValue({
			library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 4 },
			appearance: { accent_color: 'violet', reduce_motion: false },
			language: { current: 'en-US', available: ['en-US', 'pt-BR'] },
		});
		mockedUpdateSettingsConfiguration.mockResolvedValue({
			library: { runtime_root_path: '/data', watched_paths: ['/media'], remember_last_location: false, prioritize_favorites: false },
			indexing: { workers_enabled: true, scan_on_startup: false, extract_metadata: true, generate_previews: false },
			players: { remember_music_queue: false, remember_video_progress: true, autoplay_next_video: false, image_slideshow_seconds: 12 },
			appearance: { accent_color: 'rose', reduce_motion: true },
			language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
		});

		const { result } = renderHook(() => useSettings(), { wrapper: createWrapper() });

		await waitFor(() => expect(result.current.isLoading).toBe(false));

		await act(async () => {
			await result.current.saveSettings({
				library: { watched_paths: ['/media'], remember_last_location: false, prioritize_favorites: false },
				indexing: { scan_on_startup: false, extract_metadata: true, generate_previews: false },
				players: { remember_music_queue: false, remember_video_progress: true, autoplay_next_video: false, image_slideshow_seconds: 12 },
				appearance: { accent_color: 'rose', reduce_motion: true },
				language: { current: 'pt-BR' },
			});
		});

		expect(mockedUpdateSettingsConfiguration).toHaveBeenCalledTimes(1);
	});
});
