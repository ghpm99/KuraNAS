import { fireEvent, render, screen } from '@testing-library/react';
import BackupSettingsSection from './BackupSettingsSection';

const mockSetField = jest.fn();
const mockHandleSave = jest.fn();
const mockUseBackupSettings = jest.fn();

jest.mock('./useBackupSettings', () => ({
	__esModule: true,
	default: () => mockUseBackupSettings(),
	backupStatusKey: (status: string) => `STATUS_${status.toUpperCase()}`,
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string) => key,
	form: {
		enabled: true,
		destination_path: '/mnt/backup',
		retention_days: 30,
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
	pendingFiles: 5,
	isLoading: false,
	isSaving: false,
	hasError: false,
	hasUnsavedChanges: true,
	setField: mockSetField,
	handleSave: mockHandleSave,
	...overrides,
});

describe('components/settings/BackupSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseBackupSettings.mockReturnValue(createState());
	});

	it('renders the last run status and pending counter', () => {
		render(<BackupSettingsSection />);

		expect(screen.getByText('SETTINGS_BACKUP_TITLE')).toBeInTheDocument();
		expect(screen.getByText(/STATUS_COMPLETED/)).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_BACKUP_PENDING_FILES')).toBeInTheDocument();
	});

	it('saves the edited settings', () => {
		render(<BackupSettingsSection />);

		fireEvent.click(screen.getByText('SETTINGS_BACKUP_SAVE'));
		expect(mockHandleSave).toHaveBeenCalled();
	});

	it('forwards form edits to the hook', () => {
		render(<BackupSettingsSection />);

		fireEvent.click(screen.getByRole('switch'));
		expect(mockSetField).toHaveBeenCalledWith('enabled', false);

		const destination = screen.getByRole('textbox');
		fireEvent.change(destination, { target: { value: '/mnt/cold' } });
		expect(mockSetField).toHaveBeenCalledWith('destination_path', '/mnt/cold');

		const [retention, interval] = screen.getAllByRole('spinbutton');
		fireEvent.change(retention!, { target: { value: '15' } });
		expect(mockSetField).toHaveBeenCalledWith('retention_days', 15);
		fireEvent.change(interval!, { target: { value: '12' } });
		expect(mockSetField).toHaveBeenCalledWith('interval_hours', 12);
	});

	it('disables the save button without unsaved changes', () => {
		mockUseBackupSettings.mockReturnValue(createState({ hasUnsavedChanges: false }));
		render(<BackupSettingsSection />);

		expect(screen.getByText('SETTINGS_BACKUP_SAVE')).toBeDisabled();
	});

	it('shows the never-ran state and the last error verbatim', () => {
		mockUseBackupSettings.mockReturnValue(
			createState({
				status: {
					enabled: true,
					has_run: false,
					status: '',
					started_at: null,
					ended_at: null,
					last_error: 'disco cheio',
				},
				pendingFiles: undefined,
			})
		);
		render(<BackupSettingsSection />);

		expect(screen.getByText('SETTINGS_BACKUP_NEVER_RAN')).toBeInTheDocument();
		expect(screen.getByText('disco cheio')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_BACKUP_PENDING_FILES')).not.toBeInTheDocument();
	});

	it('renders the loading state', () => {
		mockUseBackupSettings.mockReturnValue(createState({ isLoading: true }));
		render(<BackupSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_BACKUP_SAVE')).not.toBeInTheDocument();
	});

	it('renders the load error alert', () => {
		mockUseBackupSettings.mockReturnValue(createState({ hasError: true }));
		render(<BackupSettingsSection />);

		expect(screen.getByText('SETTINGS_BACKUP_LOAD_ERROR')).toBeInTheDocument();
	});
});
