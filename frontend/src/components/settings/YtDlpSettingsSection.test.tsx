import { fireEvent, render, screen } from '@testing-library/react';
import YtDlpSettingsSection from './YtDlpSettingsSection';

const mockHandleUpdate = jest.fn();
const mockUseYtDlpSettings = jest.fn();

jest.mock('./useYtDlpSettings', () => ({
	__esModule: true,
	default: () => mockUseYtDlpSettings(),
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string) => key,
	status: {
		installed: true,
		current_version: '2024.08.06',
		latest_version: '2024.09.01',
		update_available: true,
		release_url: 'http://x',
		release_date: '2024-09-01',
	},
	isLoading: false,
	hasError: false,
	isUpdating: false,
	handleUpdate: mockHandleUpdate,
	...overrides,
});

describe('components/settings/YtDlpSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseYtDlpSettings.mockReturnValue(createState());
	});

	it('shows a loading indicator while fetching the status', () => {
		mockUseYtDlpSettings.mockReturnValue(createState({ isLoading: true, status: undefined }));
		render(<YtDlpSettingsSection />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('shows an error alert when the status fails to load', () => {
		mockUseYtDlpSettings.mockReturnValue(createState({ hasError: true, status: undefined }));
		render(<YtDlpSettingsSection />);
		expect(screen.getByText('SETTINGS_YTDLP_LOAD_ERROR')).toBeInTheDocument();
	});

	it('offers the update when one is available and triggers it', () => {
		render(<YtDlpSettingsSection />);

		expect(screen.getByText('SETTINGS_YTDLP_UPDATE_AVAILABLE')).toBeInTheDocument();
		expect(screen.getByText(/SETTINGS_YTDLP_LATEST/)).toBeInTheDocument();

		const button = screen.getByRole('button', { name: 'SETTINGS_YTDLP_UPDATE_BUTTON' });
		expect(button).not.toBeDisabled();
		fireEvent.click(button);
		expect(mockHandleUpdate).toHaveBeenCalled();
	});

	it('disables the button and shows the up-to-date chip when current', () => {
		mockUseYtDlpSettings.mockReturnValue(
			createState({
				status: {
					installed: true,
					current_version: '2024.09.01',
					latest_version: '2024.09.01',
					update_available: false,
					release_url: '',
					release_date: '',
				},
			})
		);
		render(<YtDlpSettingsSection />);

		expect(screen.getByText('SETTINGS_YTDLP_UP_TO_DATE')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'SETTINGS_YTDLP_UPDATE_BUTTON' })).toBeDisabled();
		expect(screen.queryByText('SETTINGS_YTDLP_RELEASE_NOTES')).not.toBeInTheDocument();
	});

	it('reports a missing binary and hides the latest chip', () => {
		mockUseYtDlpSettings.mockReturnValue(
			createState({
				status: {
					installed: false,
					current_version: '',
					latest_version: '',
					update_available: true,
					release_url: '',
					release_date: '',
				},
			})
		);
		render(<YtDlpSettingsSection />);

		expect(screen.getByText(/SETTINGS_YTDLP_NOT_INSTALLED/)).toBeInTheDocument();
		expect(screen.queryByText(/SETTINGS_YTDLP_LATEST/)).not.toBeInTheDocument();
	});

	it('shows the updating label while the mutation runs', () => {
		mockUseYtDlpSettings.mockReturnValue(createState({ isUpdating: true }));
		render(<YtDlpSettingsSection />);

		const button = screen.getByRole('button', { name: 'SETTINGS_YTDLP_UPDATING' });
		expect(button).toBeDisabled();
	});
});
