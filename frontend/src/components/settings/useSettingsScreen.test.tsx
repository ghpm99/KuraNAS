import { act, renderHook } from '@testing-library/react';
import useSettingsScreen from './useSettingsScreen';

const mockSaveSettings = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string>) => {
			const map: Record<string, string> = {
				'SETTINGS_LANGUAGE_OPTION_en-US': 'English',
				'SETTINGS_LANGUAGE_OPTION_pt-BR': 'Portuguese',
				SETTINGS_APPEARANCE_ACCENT_VIOLET: 'Violet',
				SETTINGS_APPEARANCE_ACCENT_CYAN: 'Cyan',
				SETTINGS_APPEARANCE_ACCENT_ROSE: 'Rose',
				SETTINGS_PLAYERS_SLIDESHOW_OPTION: `${params?.seconds ?? ''} seconds`,
				SETTINGS_SAVE_SUCCESS: 'Saved',
				SETTINGS_SAVE_ERROR: 'Failed',
			};
			return map[key] ?? key;
		},
	}),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
	useSettings: () => ({
		settings: {
			library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 4 },
			appearance: { accent_color: 'violet', reduce_motion: false },
			language: { current: 'en-US', available: ['en-US', 'pt-BR'] },
		},
		isLoading: false,
		isSaving: false,
		hasError: false,
		saveSettings: mockSaveSettings,
	}),
}));

describe('components/settings/useSettingsScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('derives draft state and allows editing watched paths', () => {
		const { result } = renderHook(() => useSettingsScreen());

		expect(result.current.languageOptions).toHaveLength(2);
		expect(result.current.watchedPathsText).toBe('/data');
		expect(result.current.hasUnsavedChanges).toBe(false);

		act(() => {
			result.current.handleWatchedPathsChange('/media\n/archive');
		});

		expect(result.current.draft.library.watched_paths).toEqual(['/media', '/archive']);
		expect(result.current.hasUnsavedChanges).toBe(true);
	});

	it('saves changes and reports success', async () => {
		mockSaveSettings.mockResolvedValue(undefined);
		const { result } = renderHook(() => useSettingsScreen());

		await act(async () => {
			result.current.setLanguageField('pt-BR');
			await result.current.handleSave();
		});

		expect(mockSaveSettings).toHaveBeenCalledTimes(1);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Saved', { variant: 'success' });
	});

	it('reports save failures', async () => {
		mockSaveSettings.mockRejectedValue(new Error('save failed'));
		const { result } = renderHook(() => useSettingsScreen());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed', { variant: 'error' });
	});
});
