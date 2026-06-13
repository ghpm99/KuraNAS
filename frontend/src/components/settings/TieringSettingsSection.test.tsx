import { fireEvent, render, screen } from '@testing-library/react';
import TieringSettingsSection from './TieringSettingsSection';

const mockSetField = jest.fn();
const mockHandleSave = jest.fn();
const mockUseTieringSettings = jest.fn();

jest.mock('./useTieringSettings', () => ({
	__esModule: true,
	default: () => mockUseTieringSettings(),
	tieringStatusKey: (status: string) => `STATUS_${status.toUpperCase()}`,
}));

jest.mock('@/utils', () => ({
	formatSize: (bytes: number) => `${bytes}B`,
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string) => key,
	form: {
		enabled: true,
		cold_dir_path: '/mnt/cold',
		min_age_days: 90,
		min_size_bytes: 1048576,
		interval_hours: 24,
	},
	status: {
		enabled: true,
		has_run: true,
		status: 'completed',
		started_at: '2026-06-12T10:00:00Z',
		ended_at: '2026-06-12T10:05:00Z',
		last_error: '',
	},
	usage: { hot_files: 10, hot_bytes: 1000, cold_files: 4, cold_bytes: 400 },
	isLoading: false,
	isSaving: false,
	hasError: false,
	hasUnsavedChanges: true,
	setField: mockSetField,
	handleSave: mockHandleSave,
	...overrides,
});

describe('components/settings/TieringSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseTieringSettings.mockReturnValue(createState());
	});

	it('renders the last run status and the hot/cold usage', () => {
		render(<TieringSettingsSection />);

		expect(screen.getByText('SETTINGS_TIERING_TITLE')).toBeInTheDocument();
		expect(screen.getByText(/STATUS_COMPLETED/)).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_TIERING_HOT_USAGE')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_TIERING_COLD_USAGE')).toBeInTheDocument();
	});

	it('saves the edited settings', () => {
		render(<TieringSettingsSection />);

		fireEvent.click(screen.getByText('SETTINGS_TIERING_SAVE'));
		expect(mockHandleSave).toHaveBeenCalled();
	});

	it('forwards form edits to the hook, converting MiB to bytes', () => {
		render(<TieringSettingsSection />);

		fireEvent.click(screen.getByRole('switch'));
		expect(mockSetField).toHaveBeenCalledWith('enabled', false);

		const coldDir = screen.getByRole('textbox');
		fireEvent.change(coldDir, { target: { value: '/mnt/frio' } });
		expect(mockSetField).toHaveBeenCalledWith('cold_dir_path', '/mnt/frio');

		const [age, sizeMib, interval] = screen.getAllByRole('spinbutton');
		fireEvent.change(age!, { target: { value: '30' } });
		expect(mockSetField).toHaveBeenCalledWith('min_age_days', 30);
		fireEvent.change(sizeMib!, { target: { value: '5' } });
		expect(mockSetField).toHaveBeenCalledWith('min_size_bytes', 5 * 1024 * 1024);
		fireEvent.change(interval!, { target: { value: '12' } });
		expect(mockSetField).toHaveBeenCalledWith('interval_hours', 12);
	});

	it('disables the save button without unsaved changes', () => {
		mockUseTieringSettings.mockReturnValue(createState({ hasUnsavedChanges: false }));
		render(<TieringSettingsSection />);

		expect(screen.getByText('SETTINGS_TIERING_SAVE')).toBeDisabled();
	});

	it('shows the never-ran state and the last error verbatim', () => {
		mockUseTieringSettings.mockReturnValue(
			createState({
				status: {
					enabled: true,
					has_run: false,
					status: '',
					started_at: null,
					ended_at: null,
					last_error: 'volume frio offline',
				},
				usage: undefined,
			})
		);
		render(<TieringSettingsSection />);

		expect(screen.getByText('SETTINGS_TIERING_NEVER_RAN')).toBeInTheDocument();
		expect(screen.getByText('volume frio offline')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_TIERING_HOT_USAGE')).not.toBeInTheDocument();
	});

	it('renders the loading state', () => {
		mockUseTieringSettings.mockReturnValue(createState({ isLoading: true }));
		render(<TieringSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_TIERING_SAVE')).not.toBeInTheDocument();
	});

	it('renders the load error alert', () => {
		mockUseTieringSettings.mockReturnValue(createState({ hasError: true }));
		render(<TieringSettingsSection />);

		expect(screen.getByText('SETTINGS_TIERING_LOAD_ERROR')).toBeInTheDocument();
	});
});
