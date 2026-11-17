import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
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
jest.mock('./VacuumSettingsSection', () => ({ __esModule: true, default: () => null }));
jest.mock('./YtDlpSettingsSection', () => ({ __esModule: true, default: () => null }));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const serverSettings = {
    indexing: {
        workers_enabled: true,
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
        index_document_text: true,
        scan_interval_hours: 24,
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

const renderScreen = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <SettingsProvider>
                    <MemoryRouter>
                        <SettingsScreen />
                    </MemoryRouter>
                </SettingsProvider>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

const findEnabledIntervalField = async () => {
    const intervalField = await screen.findByLabelText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS');
    await waitFor(() => expect(intervalField).toBeEnabled());
    return intervalField;
};

const clickSave = () => fireEvent.click(screen.getByRole('button', { name: 'SETTINGS_SAVE' }));

describe('components/settings indexing preferences (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.put.mockImplementation(async (_url: string, payload: unknown) => ({
            data: { ...serverSettings, ...(payload as object) },
        }));
    });

    it('renders the defaults when the server is unreachable', async () => {
        mockedApi.get.mockRejectedValue(new Error('offline'));

        renderScreen();

        expect(await findEnabledIntervalField()).toHaveValue(24);
        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT' })
        ).toBeChecked();
    });

    it('resolves keys missing from a partial server payload to the defaults', async () => {
        mockedApi.get.mockResolvedValue({
            data: {
                ...serverSettings,
                indexing: { workers_enabled: true, extract_metadata: false },
            },
        });

        renderScreen();

        expect(await findEnabledIntervalField()).toHaveValue(24);
        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_EXTRACT_METADATA' })
        ).not.toBeChecked();
        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT' })
        ).toBeChecked();
    });

    it('does not show the scan on startup preference', async () => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });

        renderScreen();
        await findEnabledIntervalField();

        expect(screen.queryByText('SETTINGS_INDEXING_SCAN_ON_STARTUP')).not.toBeInTheDocument();
    });

    it('saving sends the four indexing preferences to PUT /configuration/settings', async () => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });

        renderScreen();
        const intervalField = await findEnabledIntervalField();

        fireEvent.click(screen.getByRole('switch', { name: 'SETTINGS_INDEXING_EXTRACT_METADATA' }));
        fireEvent.click(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_GENERATE_PREVIEWS' })
        );
        fireEvent.click(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT' })
        );
        fireEvent.change(intervalField, { target: { value: '72' } });
        clickSave();

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
        expect(mockedApi.put).toHaveBeenCalledWith(
            '/configuration/settings',
            expect.objectContaining({
                indexing: {
                    scan_on_startup: true,
                    extract_metadata: false,
                    generate_previews: false,
                    index_document_text: false,
                    scan_interval_hours: 72,
                },
            })
        );
    });

    it.each(['0', '169', ''])('blocks saving when the interval is %p', async (typedHours) => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });

        renderScreen();
        const intervalField = await findEnabledIntervalField();

        fireEvent.change(intervalField, { target: { value: typedHours } });

        expect(screen.getByRole('button', { name: 'SETTINGS_SAVE' })).toBeDisabled();
        expect(mockedApi.put).not.toHaveBeenCalled();
    });

    it('shows the server message as it came when the interval is refused', async () => {
        mockedApi.get.mockResolvedValue({ data: serverSettings });
        mockedApi.put.mockRejectedValue({
            response: { data: { error: 'Intervalo da varredura fora da faixa' } },
        });

        renderScreen();
        const intervalField = await findEnabledIntervalField();

        fireEvent.change(intervalField, { target: { value: '100' } });
        clickSave();

        expect(await screen.findByText('Intervalo da varredura fora da faixa')).toBeInTheDocument();
    });
});
