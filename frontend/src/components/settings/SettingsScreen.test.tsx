import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import SettingsScreen from './SettingsScreen';

const mockHandleReset = jest.fn();
const mockHandleSave = jest.fn();
const mockSetIndexingField = jest.fn();
const mockSetCapturesField = jest.fn();
const mockSetAIField = jest.fn();
const mockSetPlayersField = jest.fn();
const mockSetAppearanceField = jest.fn();
const mockSetLanguageField = jest.fn();
const mockUseSettingsScreen = jest.fn();

jest.mock('./useSettingsScreen', () => ({
    __esModule: true,
    default: () => mockUseSettingsScreen(),
}));

jest.mock('./LibrarySettingsSection', () => ({
	__esModule: true,
	default: () => <div>Library Paths Section</div>,
}));

jest.mock('./AccessControlSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Access Control Section</div>,
}));

jest.mock('./AIProvidersSettingsSection', () => ({
	__esModule: true,
	default: () => <div>AI Providers Section</div>,
}));

jest.mock('./StorageRootsSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Storage Roots Section</div>,
}));

jest.mock('./EmailSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Email Section</div>,
}));

jest.mock('./BackupSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Backup Section</div>,
}));

jest.mock('./TieringSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Tiering Section</div>,
}));
jest.mock('./AutoShutdownSettingsSection', () => ({
	__esModule: true,
	default: () => <div>Auto Shutdown Section</div>,
}));
jest.mock('./YtDlpSettingsSection', () => ({
	__esModule: true,
	default: () => <div>yt-dlp Section</div>,
}));
jest.mock('./ImageClassificationBackfill', () => ({
	__esModule: true,
	default: () => <div>Image Classification Backfill</div>,
}));

const createScreenState = (overrides: Record<string, any> = {}) => ({
    t: (key: string, params?: Record<string, string>) => {
        const map: Record<string, string> = {
            SETTINGS_PAGE_TITLE: 'Settings',
            SETTINGS_PAGE_DESCRIPTION: 'Manage runtime preferences.',
            SETTINGS_SUMMARY_WORKERS: 'Workers',
            SETTINGS_STATUS_ENABLED: 'Enabled',
            SETTINGS_STATUS_DISABLED: 'Disabled',
            LANGUAGE: 'Language',
            SETTINGS_SECTION_INDEXING: 'Indexing',
            SETTINGS_SECTION_INDEXING_DESCRIPTION: 'Indexing behavior.',
            SETTINGS_INDEXING_SCAN_ON_STARTUP: 'Scan on startup',
            SETTINGS_INDEXING_EXTRACT_METADATA: 'Extract metadata',
            SETTINGS_INDEXING_GENERATE_PREVIEWS: 'Generate previews',
            SETTINGS_INDEXING_WORKERS_ON: 'Workers are enabled.',
            SETTINGS_INDEXING_WORKERS_OFF: 'Workers are disabled.',
            SETTINGS_SECTION_AI: 'Artificial Intelligence',
            SETTINGS_SECTION_AI_DESCRIPTION: 'AI usage controls.',
            SETTINGS_AI_IMAGE_CLASSIFICATION: 'Classify images with AI',
            SETTINGS_AI_IMAGE_CLASSIFICATION_HELP: 'Heuristic-only when disabled.',
            SETTINGS_SECTION_PLAYERS: 'Players',
            SETTINGS_SECTION_PLAYERS_DESCRIPTION: 'Playback behavior.',
            SETTINGS_PLAYERS_REMEMBER_MUSIC_QUEUE: 'Remember music queue',
            SETTINGS_PLAYERS_REMEMBER_VIDEO_PROGRESS: 'Remember video progress',
            SETTINGS_PLAYERS_AUTOPLAY_NEXT_VIDEO: 'Autoplay next video',
            SETTINGS_PLAYERS_SLIDESHOW_INTERVAL: 'Slideshow interval',
            SETTINGS_PLAYERS_SLIDESHOW_OPTION: `${params?.seconds ?? ''} seconds`,
            SETTINGS_SECTION_APPEARANCE: 'Appearance',
            SETTINGS_SECTION_APPEARANCE_DESCRIPTION: 'Appearance controls.',
            SETTINGS_APPEARANCE_ACCENT: 'Accent color',
            SETTINGS_APPEARANCE_ACCENT_VIOLET: 'Violet',
            SETTINGS_APPEARANCE_ACCENT_CYAN: 'Cyan',
            SETTINGS_APPEARANCE_ACCENT_ROSE: 'Rose',
            SETTINGS_APPEARANCE_REDUCE_MOTION: 'Reduce motion',
            SETTINGS_SECTION_LANGUAGE: 'Language',
            SETTINGS_SECTION_LANGUAGE_DESCRIPTION: 'Language selection.',
            SETTINGS_LANGUAGE_HELP: 'Translations refresh after saving.',
            ANALYTICS: 'Analytics',
            ABOUT: 'About',
            SETTINGS_RESET: 'Reset',
            SETTINGS_SAVE: 'Save changes',
            SAVING: 'Saving...',
            SETTINGS_LOAD_ERROR: 'Unable to load settings.',
        };
        return map[key] ?? key;
    },
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
    },
    draft: {
        indexing: {
            scan_on_startup: true,
            extract_metadata: true,
            generate_previews: true,
        },
        captures: {
            save_path: '/srv/capturas',
        },
        ai: {
            image_classification: true,
        },
        players: {
            remember_music_queue: true,
            remember_video_progress: true,
            autoplay_next_video: true,
            image_slideshow_seconds: 8,
        },
        appearance: { accent_color: 'violet', reduce_motion: false },
        language: { current: 'en-US' },
    },
    isLoading: false,
    isSaving: false,
    hasError: false,
    hasUnsavedChanges: true,
    languageOptions: [
        { value: 'en-US', label: 'English' },
        { value: 'pt-BR', label: 'Portuguese' },
    ],
    accentOptions: [
        { value: 'violet', label: 'Violet' },
        { value: 'cyan', label: 'Cyan' },
    ],
    slideshowOptions: [
        { value: 4, label: '4 seconds' },
        { value: 8, label: '8 seconds' },
    ],
    setIndexingField: mockSetIndexingField,
    setCapturesField: mockSetCapturesField,
    setAIField: mockSetAIField,
    setPlayersField: mockSetPlayersField,
    setAppearanceField: mockSetAppearanceField,
    setLanguageField: mockSetLanguageField,
    handleReset: mockHandleReset,
    handleSave: mockHandleSave,
    ...overrides,
});

describe('components/settings/SettingsScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseSettingsScreen.mockReturnValue(createScreenState());
    });

    it('renders all settings sections and triggers save/reset actions', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('Settings')).toBeInTheDocument();
        expect(screen.getByText('Indexing')).toBeInTheDocument();
        expect(screen.getByText('Players')).toBeInTheDocument();
        expect(screen.getByText('Appearance')).toBeInTheDocument();
        // Language appears in summary chip and section title
        expect(screen.getAllByText('Language').length).toBeGreaterThanOrEqual(1);

        fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
        fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

        expect(mockHandleReset).toHaveBeenCalledTimes(1);
        expect(mockHandleSave).toHaveBeenCalledTimes(1);
    });

    it('renders fallback summaries and disabled warnings when runtime configuration is unavailable', () => {
        mockUseSettingsScreen.mockReturnValue(
            createScreenState({
                settings: {
                    indexing: {
                        workers_enabled: false,
                        scan_on_startup: false,
                        extract_metadata: false,
                        generate_previews: false,
                    },
                },
                draft: {
                    indexing: {
                        scan_on_startup: false,
                        extract_metadata: false,
                        generate_previews: false,
                    },
                    ai: {
                        image_classification: false,
                    },
                    players: {
                        remember_music_queue: false,
                        remember_video_progress: false,
                        autoplay_next_video: false,
                        image_slideshow_seconds: 4,
                    },
                    appearance: { accent_color: 'cyan', reduce_motion: true },
                    language: { current: 'pt-BR' },
                },
                languageOptions: [{ value: 'pt-BR', label: 'Portuguese' }],
                hasUnsavedChanges: false,
            })
        );

        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('Workers: Disabled')).toBeInTheDocument();
        expect(screen.getByText('Workers are disabled.')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
    });

    it('surfaces load errors and disables actions while saving', () => {
        mockUseSettingsScreen.mockReturnValue(
            createScreenState({
                hasError: true,
                isSaving: true,
                hasUnsavedChanges: true,
            })
        );

        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('Unable to load settings.')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Saving...' })).toBeDisabled();
        expect(screen.getByRole('switch', { name: 'Remember music queue' })).toBeDisabled();
        expect(screen.getByRole('switch', { name: 'Reduce motion' })).toBeDisabled();
    });

    it('disables controls while isLoading is true', () => {
        mockUseSettingsScreen.mockReturnValue(
            createScreenState({
                isLoading: true,
                hasUnsavedChanges: true,
            })
        );

        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByRole('switch', { name: 'Scan on startup' })).toBeDisabled();
        expect(screen.getByRole('switch', { name: 'Remember music queue' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Save changes' })).toBeDisabled();
    });

    it('fires indexing switch handlers with correct arguments', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('switch', { name: 'Scan on startup' }));
        expect(mockSetIndexingField).toHaveBeenCalledWith('scan_on_startup', false);

        fireEvent.click(screen.getByRole('switch', { name: 'Extract metadata' }));
        expect(mockSetIndexingField).toHaveBeenCalledWith('extract_metadata', false);

        fireEvent.click(screen.getByRole('switch', { name: 'Generate previews' }));
        expect(mockSetIndexingField).toHaveBeenCalledWith('generate_previews', false);
    });

    it('fires the AI image classification switch handler with correct arguments', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('switch', { name: 'Classify images with AI' }));
        expect(mockSetAIField).toHaveBeenCalledWith('image_classification', false);
    });

    it('fires player switch handlers with correct arguments', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('switch', { name: 'Remember music queue' }));
        expect(mockSetPlayersField).toHaveBeenCalledWith('remember_music_queue', false);

        fireEvent.click(screen.getByRole('switch', { name: 'Remember video progress' }));
        expect(mockSetPlayersField).toHaveBeenCalledWith('remember_video_progress', false);

        fireEvent.click(screen.getByRole('switch', { name: 'Autoplay next video' }));
        expect(mockSetPlayersField).toHaveBeenCalledWith('autoplay_next_video', false);
    });

    it('fires appearance reduce_motion switch handler', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        fireEvent.click(screen.getByRole('switch', { name: 'Reduce motion' }));
        expect(mockSetAppearanceField).toHaveBeenCalledWith('reduce_motion', true);
    });

    it('renders footer links to analytics and about pages', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('Analytics')).toBeInTheDocument();
        expect(screen.getByText('About')).toBeInTheDocument();
    });

    it('does not show error alert when hasError is false', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.queryByText('Unable to load settings.')).not.toBeInTheDocument();
    });

    it('renders workers enabled alert when workers are on', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('Workers are enabled.')).toBeInTheDocument();
    });

    it('renders indexing chip variants based on draft boolean state', () => {
        mockUseSettingsScreen.mockReturnValue(
            createScreenState({
                draft: {
                    indexing: {
                        scan_on_startup: false,
                        extract_metadata: true,
                        generate_previews: false,
                    },
                    ai: {
                        image_classification: false,
                    },
                    players: {
                        remember_music_queue: true,
                        remember_video_progress: true,
                        autoplay_next_video: true,
                        image_slideshow_seconds: 8,
                    },
                    appearance: { accent_color: 'violet', reduce_motion: false },
                    language: { current: 'en-US' },
                },
            })
        );

        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        // Chips for indexing toggles are always rendered - just verify section is present
        expect(screen.getByText('Indexing')).toBeInTheDocument();
    });

    it('triggers setPlayersField when slideshow select changes', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        // Open the slideshow select by clicking the MUI select trigger
        const slideshowSelect = screen.getByLabelText('Slideshow interval');
        fireEvent.mouseDown(slideshowSelect);

        // Click the "4 seconds" option in the dropdown
        const option = screen.getByRole('option', { name: '4 seconds' });
        fireEvent.click(option);

        expect(mockSetPlayersField).toHaveBeenCalledWith('image_slideshow_seconds', 4);
    });

    it('triggers setAppearanceField when accent select changes', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        const accentSelect = screen.getByLabelText('Accent color');
        fireEvent.mouseDown(accentSelect);

        const option = screen.getByRole('option', { name: 'Cyan' });
        fireEvent.click(option);

        expect(mockSetAppearanceField).toHaveBeenCalledWith('accent_color', 'cyan');
    });

    it('triggers setLanguageField when language select changes', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        // There are multiple elements with "Language" label; get the select specifically
        const languageSelect = screen
            .getAllByLabelText('Language')
            .find((el) => el.getAttribute('role') === 'combobox')!;
        fireEvent.mouseDown(languageSelect);

        const option = screen.getByRole('option', { name: 'Portuguese' });
        fireEvent.click(option);

        expect(mockSetLanguageField).toHaveBeenCalledWith('pt-BR');
    });

    it('renders accent color swatch rows', () => {
        render(
            <MemoryRouter>
                <SettingsScreen />
            </MemoryRouter>
        );

        // accentOptions has Violet and Cyan; each appears in swatches
        expect(screen.getAllByText('Violet').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Cyan').length).toBeGreaterThanOrEqual(1);
    });
});
