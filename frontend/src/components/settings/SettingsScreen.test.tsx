import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import SettingsScreen from './SettingsScreen';

const mockHandleReset = jest.fn();
const mockHandleSave = jest.fn();
const mockSetPlayersField = jest.fn();
const mockSetAppearanceField = jest.fn();
const mockSetLanguageField = jest.fn();
const mockUseSettingsScreen = jest.fn();

jest.mock('./useSettingsScreen', () => ({
	__esModule: true,
	default: () => mockUseSettingsScreen(),
}));

const createScreenState = (overrides: Record<string, any> = {}) => ({
	t: (key: string, params?: Record<string, string>) => {
		const map: Record<string, string> = {
			SETTINGS_PAGE_TITLE: 'Settings',
			SETTINGS_PAGE_DESCRIPTION: 'Manage runtime preferences.',
			SETTINGS_SUMMARY_ROOT: 'Root',
			SETTINGS_SUMMARY_WORKERS: 'Workers',
			SETTINGS_STATUS_ENABLED: 'Enabled',
			SETTINGS_STATUS_DISABLED: 'Disabled',
			LANGUAGE: 'Language',
			SETTINGS_SECTION_LIBRARY: 'Library',
			SETTINGS_SECTION_LIBRARY_DESCRIPTION: 'Library preferences.',
			SETTINGS_LIBRARY_WATCHED_PATHS: 'Watched paths',
			SETTINGS_LIBRARY_WATCHED_PATHS_HELP: 'One path per line.',
			SETTINGS_LIBRARY_REMEMBER_LAST_LOCATION: 'Remember last location',
			SETTINGS_LIBRARY_PRIORITIZE_FAVORITES: 'Prioritize favorites',
			SETTINGS_LIBRARY_RUNTIME_ROOT: `Runtime root: ${params?.path ?? ''}`,
			SETTINGS_SECTION_INDEXING: 'Indexing',
			SETTINGS_SECTION_INDEXING_DESCRIPTION: 'Indexing behavior.',
			SETTINGS_INDEXING_SCAN_ON_STARTUP: 'Scan on startup',
			SETTINGS_INDEXING_EXTRACT_METADATA: 'Extract metadata',
			SETTINGS_INDEXING_GENERATE_PREVIEWS: 'Generate previews',
			SETTINGS_INDEXING_WORKERS_ON: 'Workers are enabled.',
			SETTINGS_INDEXING_WORKERS_OFF: 'Workers are disabled.',
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
		library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
		indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
	},
	draft: {
		library: { watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
		indexing: { scan_on_startup: true, extract_metadata: true, generate_previews: true },
		players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 8 },
		appearance: { accent_color: 'violet', reduce_motion: false },
		language: { current: 'en-US' },
	},
	isLoading: false,
	isSaving: false,
	hasError: false,
	hasUnsavedChanges: true,
	languageOptions: [{ value: 'en-US', label: 'English' }],
	accentOptions: [
		{ value: 'violet', label: 'Violet' },
		{ value: 'cyan', label: 'Cyan' },
	],
	slideshowOptions: [
		{ value: 4, label: '4 seconds' },
		{ value: 8, label: '8 seconds' },
	],
	watchedPathsText: '/data',
	setLibraryField: jest.fn(),
	setIndexingField: jest.fn(),
	setPlayersField: mockSetPlayersField,
	setAppearanceField: mockSetAppearanceField,
	setLanguageField: mockSetLanguageField,
	handleWatchedPathsChange: jest.fn(),
	handleReset: mockHandleReset,
	handleSave: mockHandleSave,
	...overrides,
});

describe('components/settings/SettingsScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseSettingsScreen.mockReturnValue(createScreenState());
	});

	it('renders settings sections and triggers actions', () => {
		render(
			<MemoryRouter>
				<SettingsScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('Settings')).toBeInTheDocument();
		expect(screen.getByText('Library')).toBeInTheDocument();
		expect(screen.getByText('Appearance')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Save changes' })).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
		fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

		expect(mockHandleReset).toHaveBeenCalledTimes(1);
		expect(mockHandleSave).toHaveBeenCalledTimes(1);
	});

	it('renders fallback summaries and disabled warnings when runtime configuration is unavailable', () => {
		mockUseSettingsScreen.mockReturnValue(
			createScreenState({
				settings: {
					library: { runtime_root_path: '', watched_paths: ['/data', '/backup'], remember_last_location: true, prioritize_favorites: true },
					indexing: { workers_enabled: false, scan_on_startup: false, extract_metadata: false, generate_previews: false },
				},
				draft: {
					library: { watched_paths: ['/data', '/backup'], remember_last_location: true, prioritize_favorites: true },
					indexing: { scan_on_startup: false, extract_metadata: false, generate_previews: false },
					players: { remember_music_queue: false, remember_video_progress: false, autoplay_next_video: false, image_slideshow_seconds: 4 },
					appearance: { accent_color: 'cyan', reduce_motion: true },
					language: { current: 'pt-BR' },
				},
				languageOptions: [{ value: 'pt-BR', label: 'Portuguese' }],
				hasUnsavedChanges: false,
			}),
		);

		render(
			<MemoryRouter>
				<SettingsScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('Root: -')).toBeInTheDocument();
		expect(screen.getByText('Workers: Disabled')).toBeInTheDocument();
		expect(screen.getByText('/backup')).toBeInTheDocument();
		expect(screen.getByText('Workers are disabled.')).toBeInTheDocument();
		expect(screen.getByText('Runtime root: -')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
	});

	it('surfaces load errors and disables actions while saving', () => {
		mockUseSettingsScreen.mockReturnValue(
			createScreenState({
				hasError: true,
				isSaving: true,
				hasUnsavedChanges: true,
			}),
		);

		render(
			<MemoryRouter>
				<SettingsScreen />
			</MemoryRouter>,
		);

		expect(screen.getByText('Unable to load settings.')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
		expect(screen.getByRole('button', { name: 'Saving...' })).toBeDisabled();
		expect(screen.getByRole('switch', { name: 'Remember music queue' })).toBeDisabled();
		expect(screen.getByRole('switch', { name: 'Reduce motion' })).toBeDisabled();
	});
});
