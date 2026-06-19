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
            indexing: {
                workers_enabled: true,
                scan_on_startup: true,
                extract_metadata: true,
                generate_previews: true,
            },
            captures: {
                save_path: '/srv/capturas',
                default_path: '/srv/capturas',
                storage_roots: ['/data'],
            },
            ai: {
                image_classification: true,
            },
            players: {
                remember_music_queue: true,
                remember_video_progress: true,
                autoplay_next_video: true,
                image_slideshow_seconds: 4,
            },
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

    it('derives draft state from provider settings', () => {
        const { result } = renderHook(() => useSettingsScreen());

        expect(result.current.languageOptions).toHaveLength(2);
        expect(result.current.hasUnsavedChanges).toBe(false);

        act(() => {
            result.current.setLanguageField('pt-BR');
        });

        expect(result.current.draft.language.current).toBe('pt-BR');
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
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Saved', {
            variant: 'success',
        });
    });

    it('reports save failures', async () => {
        mockSaveSettings.mockRejectedValue(new Error('save failed'));
        const { result } = renderHook(() => useSettingsScreen());

        await act(async () => {
            await result.current.handleSave();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed', {
            variant: 'error',
        });
    });

    it('computes accent options from constant values', () => {
        const { result } = renderHook(() => useSettingsScreen());

        expect(result.current.accentOptions).toEqual([
            { value: 'violet', label: 'Violet' },
            { value: 'cyan', label: 'Cyan' },
            { value: 'rose', label: 'Rose' },
        ]);
    });

    it('computes slideshow options from constant values', () => {
        const { result } = renderHook(() => useSettingsScreen());

        expect(result.current.slideshowOptions).toEqual([
            { value: 4, label: '4 seconds' },
            { value: 8, label: '8 seconds' },
            { value: 12, label: '12 seconds' },
            { value: 20, label: '20 seconds' },
        ]);
    });

    it('setIndexingField updates indexing fields in draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setIndexingField('scan_on_startup', false);
        });
        expect(result.current.draft.indexing.scan_on_startup).toBe(false);

        act(() => {
            result.current.setIndexingField('extract_metadata', false);
        });
        expect(result.current.draft.indexing.extract_metadata).toBe(false);

        act(() => {
            result.current.setIndexingField('generate_previews', false);
        });
        expect(result.current.draft.indexing.generate_previews).toBe(false);
    });

    it('setAIField updates AI fields in draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setAIField('image_classification', false);
        });
        expect(result.current.draft.ai.image_classification).toBe(false);
    });

    it('setPlayersField updates players fields in draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setPlayersField('remember_music_queue', false);
        });
        expect(result.current.draft.players.remember_music_queue).toBe(false);

        act(() => {
            result.current.setPlayersField('remember_video_progress', false);
        });
        expect(result.current.draft.players.remember_video_progress).toBe(false);

        act(() => {
            result.current.setPlayersField('autoplay_next_video', false);
        });
        expect(result.current.draft.players.autoplay_next_video).toBe(false);

        act(() => {
            result.current.setPlayersField('image_slideshow_seconds', 12);
        });
        expect(result.current.draft.players.image_slideshow_seconds).toBe(12);
    });

    it('setAppearanceField updates appearance fields in draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setAppearanceField('accent_color', 'rose');
        });
        expect(result.current.draft.appearance.accent_color).toBe('rose');

        act(() => {
            result.current.setAppearanceField('reduce_motion', true);
        });
        expect(result.current.draft.appearance.reduce_motion).toBe(true);
    });

    it('setLanguageField updates language current in draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setLanguageField('pt-BR');
        });
        expect(result.current.draft.language.current).toBe('pt-BR');
    });

    it('handleReset reverts draft to baseline', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setLanguageField('pt-BR');
            result.current.setAppearanceField('accent_color', 'rose');
        });
        expect(result.current.hasUnsavedChanges).toBe(true);

        act(() => {
            result.current.handleReset();
        });
        expect(result.current.hasUnsavedChanges).toBe(false);
        expect(result.current.draft.language.current).toBe('en-US');
        expect(result.current.draft.appearance.accent_color).toBe('violet');
    });

    it('hasUnsavedChanges is false when draft equals baseline', () => {
        const { result } = renderHook(() => useSettingsScreen());
        expect(result.current.hasUnsavedChanges).toBe(false);
    });

    it('passes the current draft to saveSettings', async () => {
        mockSaveSettings.mockResolvedValue(undefined);
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setPlayersField('image_slideshow_seconds', 20);
        });

        await act(async () => {
            await result.current.handleSave();
        });

        const savedDraft = mockSaveSettings.mock.calls[0][0];
        expect(savedDraft.players.image_slideshow_seconds).toBe(20);
    });

    it('setCapturesField updates the captures save path in the draft', () => {
        const { result } = renderHook(() => useSettingsScreen());

        act(() => {
            result.current.setCapturesField('save_path', '/srv/outras-capturas');
        });
        expect(result.current.draft.captures.save_path).toBe('/srv/outras-capturas');
    });

    it('handleSave surfaces the backend error message when the save is rejected', async () => {
        mockSaveSettings.mockRejectedValue({
            response: { data: { error: 'Captures path must be outside the storage roots' } },
        });
        const { result } = renderHook(() => useSettingsScreen());

        await act(async () => {
            await result.current.handleSave();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'Captures path must be outside the storage roots',
            { variant: 'error' }
        );
    });

    it('handleSave falls back to the generic error toast without a backend message', async () => {
        mockSaveSettings.mockRejectedValue(new Error('network down'));
        const { result } = renderHook(() => useSettingsScreen());

        await act(async () => {
            await result.current.handleSave();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed', { variant: 'error' });
    });

    it('exposes settings and loading states from provider', () => {
        const { result } = renderHook(() => useSettingsScreen());

        expect(result.current.settings.indexing.workers_enabled).toBe(true);
        expect(result.current.isLoading).toBe(false);
        expect(result.current.isSaving).toBe(false);
        expect(result.current.hasError).toBe(false);
    });
});
