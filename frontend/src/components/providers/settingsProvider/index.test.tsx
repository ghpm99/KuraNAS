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

const fullSettings = (overrides: Record<string, any> = {}) => ({
    library: {
        runtime_root_path: '/data',
        watched_paths: ['/data'],
        remember_last_location: true,
        prioritize_favorites: true,
    },
    indexing: {
        workers_enabled: true,
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
    },
    players: {
        remember_music_queue: true,
        remember_video_progress: true,
        autoplay_next_video: true,
        image_slideshow_seconds: 4,
    },
    appearance: { accent_color: 'violet' as const, reduce_motion: false },
    language: { current: 'en-US', available: ['en-US', 'pt-BR'] },
    ...overrides,
});

describe('components/providers/settingsProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        document.documentElement.removeAttribute('data-app-motion');
        document.documentElement.style.removeProperty('--app-color-primary');
        document.documentElement.style.removeProperty('--app-color-primary-hover');
        document.documentElement.style.removeProperty('--app-shadow-active-primary');
    });

    it('loads settings and applies appearance runtime values', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(
            fullSettings({
                appearance: { accent_color: 'cyan', reduce_motion: true },
                language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
            })
        );

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.settings.language.current).toBe('pt-BR');
        expect(document.documentElement.style.getPropertyValue('--app-color-primary')).toBe(
            '#06B6D4'
        );
        expect(document.documentElement.style.getPropertyValue('--app-color-primary-hover')).toBe(
            '#22D3EE'
        );
        expect(document.documentElement.dataset.appMotion).toBe('reduced');
    });

    it('saves settings through the mutation boundary', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(fullSettings());
        mockedUpdateSettingsConfiguration.mockResolvedValue(
            fullSettings({
                library: {
                    runtime_root_path: '/data',
                    watched_paths: ['/media'],
                    remember_last_location: false,
                    prioritize_favorites: false,
                },
                appearance: { accent_color: 'rose', reduce_motion: true },
                language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
            })
        );

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        await act(async () => {
            await result.current.saveSettings({
                library: {
                    watched_paths: ['/media'],
                    remember_last_location: false,
                    prioritize_favorites: false,
                },
                indexing: {
                    scan_on_startup: false,
                    extract_metadata: true,
                    generate_previews: false,
                },
                players: {
                    remember_music_queue: false,
                    remember_video_progress: true,
                    autoplay_next_video: false,
                    image_slideshow_seconds: 12,
                },
                appearance: { accent_color: 'rose', reduce_motion: true },
                language: { current: 'pt-BR' },
            });
        });

        expect(mockedUpdateSettingsConfiguration).toHaveBeenCalledTimes(1);
    });

    it('removes data-app-motion when reduce_motion is false', async () => {
        document.documentElement.dataset.appMotion = 'reduced';

        mockedGetSettingsConfiguration.mockResolvedValue(
            fullSettings({
                appearance: { accent_color: 'violet', reduce_motion: false },
            })
        );

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(document.documentElement.dataset.appMotion).toBeUndefined();
    });

    it('applies violet accent palette CSS variables', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(fullSettings());

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(document.documentElement.style.getPropertyValue('--app-color-primary')).toBe(
            '#6D5DF6'
        );
        expect(document.documentElement.style.getPropertyValue('--app-color-primary-hover')).toBe(
            '#7C70FF'
        );
    });

    it('applies rose accent palette CSS variables', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(
            fullSettings({
                appearance: { accent_color: 'rose', reduce_motion: false },
            })
        );

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(document.documentElement.style.getPropertyValue('--app-color-primary')).toBe(
            '#E11D48'
        );
        expect(document.documentElement.style.getPropertyValue('--app-color-primary-hover')).toBe(
            '#FB7185'
        );
    });

    it('falls back to default settings when query fails', async () => {
        mockedGetSettingsConfiguration.mockRejectedValue(new Error('network error'));

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.hasError).toBe(true);
        expect(result.current.settings.library.runtime_root_path).toBe('');
        expect(result.current.settings.appearance.accent_color).toBe('violet');
    });

    it('exposes isSaving as false before mutation and refresh function', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(fullSettings());

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.isSaving).toBe(false);
        expect(typeof result.current.refresh).toBe('function');
    });

    it('refresh re-fetches settings data', async () => {
        mockedGetSettingsConfiguration.mockResolvedValue(fullSettings());

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        mockedGetSettingsConfiguration.mockResolvedValue(
            fullSettings({
                language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
            })
        );

        await act(async () => {
            await result.current.refresh();
        });

        await waitFor(() => expect(result.current.settings.language.current).toBe('pt-BR'));
    });

    it('throws when useSettings is called outside provider', () => {
        expect(() => {
            renderHook(() => useSettings());
        }).toThrow('useSettings must be used within a SettingsProvider');
    });

    it('updates settings in query cache after successful save', async () => {
        const updatedSettings = fullSettings({
            appearance: { accent_color: 'cyan', reduce_motion: false },
        });
        // Both the initial load and re-fetch after invalidation return the updated value
        mockedGetSettingsConfiguration.mockResolvedValue(updatedSettings);
        mockedUpdateSettingsConfiguration.mockResolvedValue(updatedSettings);

        const { result } = renderHook(() => useSettings(), {
            wrapper: createWrapper(),
        });

        await waitFor(() => expect(result.current.isLoading).toBe(false));

        await act(async () => {
            await result.current.saveSettings({
                library: {
                    watched_paths: ['/data'],
                    remember_last_location: true,
                    prioritize_favorites: true,
                },
                indexing: {
                    scan_on_startup: true,
                    extract_metadata: true,
                    generate_previews: true,
                },
                players: {
                    remember_music_queue: true,
                    remember_video_progress: true,
                    autoplay_next_video: true,
                    image_slideshow_seconds: 4,
                },
                appearance: { accent_color: 'cyan', reduce_motion: false },
                language: { current: 'en-US' },
            });
        });

        await waitFor(() => expect(result.current.settings.appearance.accent_color).toBe('cyan'));
        expect(document.documentElement.style.getPropertyValue('--app-color-primary')).toBe(
            '#06B6D4'
        );
    });
});
