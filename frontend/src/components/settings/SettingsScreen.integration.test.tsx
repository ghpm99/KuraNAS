import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import SettingsProvider from '@/components/providers/settingsProvider';
import { apiBase } from '@/service';
import SettingsScreen from './SettingsScreen';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('./AccessControlSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./AIProvidersSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./ImageClassificationBackfill', () => ({ __esModule: true, default: () => null }));
jest.mock('./AutoShutdownSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./BackupSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./EmailSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./LibrarySettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./StorageRootsSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./TieringSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./YtDlpSettingsSection', () => ({ __esModule: true, default: () => null }));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const serverSettings = {
    indexing: {
        workers_enabled: true,
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
    },
    captures: { save_path: '/data/Capturas', default_path: '/data/Capturas', storage_roots: [] },
    ai: { image_classification: true },
    players: {
        remember_music_queue: true,
        remember_video_progress: true,
        autoplay_next_video: true,
        image_slideshow_seconds: 4,
    },
    appearance: { accent_color: 'violet', reduce_motion: false, theme_mode: 'dark' },
    language: { current: 'en-US', available: ['en-US'] },
};

const waitForEnabledThemeSelect = async () => {
    const themeSelect = await screen.findByLabelText('SETTINGS_APPEARANCE_THEME');
    await waitFor(() => expect(themeSelect).not.toHaveAttribute('aria-disabled', 'true'));
    return themeSelect;
};

const renderScreen = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <SettingsProvider>
                <MemoryRouter>
                    <SettingsScreen />
                </MemoryRouter>
            </SettingsProvider>
        </QueryClientProvider>
    );
};

describe('components/settings/SettingsScreen theme control (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.put.mockImplementation(async (_url: string, payload: unknown) => ({
            data: { ...serverSettings, ...(payload as object) },
        }));
    });

    it('renders the appearance section without any backend response', async () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));

        renderScreen();

        expect(await screen.findByText('SETTINGS_SECTION_APPEARANCE')).toBeInTheDocument();
        expect(screen.getByLabelText('SETTINGS_APPEARANCE_THEME')).toBeInTheDocument();
    });

    it('saving a different theme issues PUT /configuration/settings with appearance.theme_mode', async () => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });

        renderScreen();

        const themeSelect = await waitForEnabledThemeSelect();
        expect(themeSelect).toHaveTextContent('SETTINGS_APPEARANCE_THEME_DARK');
        fireEvent.mouseDown(themeSelect);
        fireEvent.click(screen.getByRole('option', { name: 'SETTINGS_APPEARANCE_THEME_LIGHT' }));
        fireEvent.click(screen.getByRole('button', { name: 'SETTINGS_SAVE' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
        expect(mockedApi.put).toHaveBeenCalledWith(
            '/configuration/settings',
            expect.objectContaining({
                appearance: { accent_color: 'violet', reduce_motion: false, theme_mode: 'light' },
            })
        );
    });

    it('offers dark, light and system', async () => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });

        renderScreen();

        fireEvent.mouseDown(await waitForEnabledThemeSelect());

        expect(screen.getAllByRole('option').map((option) => option.textContent)).toEqual(
            expect.arrayContaining([
                'SETTINGS_APPEARANCE_THEME_DARK',
                'SETTINGS_APPEARANCE_THEME_LIGHT',
                'SETTINGS_APPEARANCE_THEME_SYSTEM',
            ])
        );
    });
});
