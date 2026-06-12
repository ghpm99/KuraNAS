import { fireEvent, render, screen } from '@testing-library/react';
import EmailSettingsSection from './EmailSettingsSection';

const mockHandleLinkGoogle = jest.fn();
const mockHandleLinkMicrosoft = jest.fn();
const mockHandleToggleSync = jest.fn();
const mockHandleRemove = jest.fn();
const mockUseEmailSettings = jest.fn();

jest.mock('./useEmailSettings', () => ({
	__esModule: true,
	default: () => mockUseEmailSettings(),
}));

const sampleAccounts = [
	{
		id: 1,
		provider: 'google',
		address: 'owner@gmail.com',
		display_name: 'Owner',
		status: 'linked',
		sync_enabled: true,
		last_sync_at: null,
		last_error: '',
		created_at: '2026-06-12',
	},
	{
		id: 2,
		provider: 'microsoft',
		address: 'owner@hotmail.com',
		display_name: 'Owner',
		status: 'reauth_required',
		sync_enabled: false,
		last_sync_at: null,
		last_error: 'invalid_grant',
		created_at: '2026-06-12',
	},
];

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string) => key,
	accounts: sampleAccounts,
	isLoading: false,
	isSaving: false,
	hasError: false,
	loadErrorMessage: '',
	deviceCode: null,
	deviceStatus: null,
	handleLinkGoogle: mockHandleLinkGoogle,
	handleLinkMicrosoft: mockHandleLinkMicrosoft,
	handleToggleSync: mockHandleToggleSync,
	handleRemove: mockHandleRemove,
	...overrides,
});

describe('components/settings/EmailSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseEmailSettings.mockReturnValue(createState());
	});

	it('renders the linked accounts with status chips', () => {
		render(<EmailSettingsSection />);

		expect(screen.getByText('SETTINGS_EMAIL_TITLE')).toBeInTheDocument();
		expect(screen.getByText('owner@gmail.com')).toBeInTheDocument();
		expect(screen.getByText('owner@hotmail.com')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_EMAIL_STATUS_LINKED')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_EMAIL_STATUS_REAUTH_REQUIRED')).toBeInTheDocument();
		// The backend's last_error arrives translated and is shown verbatim.
		expect(screen.getByText('invalid_grant')).toBeInTheDocument();
	});

	it('starts each link flow from its button', () => {
		render(<EmailSettingsSection />);

		fireEvent.click(screen.getByText('SETTINGS_EMAIL_ADD_GOOGLE'));
		expect(mockHandleLinkGoogle).toHaveBeenCalled();

		fireEvent.click(screen.getByText('SETTINGS_EMAIL_ADD_MICROSOFT'));
		expect(mockHandleLinkMicrosoft).toHaveBeenCalled();
	});

	it('toggles sync and removes accounts', () => {
		render(<EmailSettingsSection />);

		const switches = screen.getAllByRole('checkbox');
		fireEvent.click(switches[0]!);
		expect(mockHandleToggleSync).toHaveBeenCalledWith(1, false);

		const removeButtons = screen.getAllByText('SETTINGS_EMAIL_REMOVE');
		fireEvent.click(removeButtons[1]!);
		expect(mockHandleRemove).toHaveBeenCalledWith(2);
	});

	it('shows the device code block while a microsoft link is pending', () => {
		mockUseEmailSettings.mockReturnValue(
			createState({
				deviceCode: {
					user_code: 'ABC123',
					verification_uri: 'https://microsoft.com/devicelogin',
					expires_in: 900,
					message: 'abra o link e digite o código',
				},
				deviceStatus: 'pending',
			})
		);
		render(<EmailSettingsSection />);

		expect(screen.getByText('abra o link e digite o código')).toBeInTheDocument();
		expect(screen.getByText('ABC123')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_EMAIL_DEVICE_PENDING')).toBeInTheDocument();
	});

	it('celebrates a finished device link', () => {
		mockUseEmailSettings.mockReturnValue(
			createState({
				deviceCode: {
					user_code: 'ABC123',
					verification_uri: 'v',
					expires_in: 900,
					message: 'm',
				},
				deviceStatus: 'linked',
			})
		);
		render(<EmailSettingsSection />);

		expect(screen.getByText('SETTINGS_EMAIL_DEVICE_LINKED')).toBeInTheDocument();
	});

	it('shows the empty warning without linked accounts', () => {
		mockUseEmailSettings.mockReturnValue(createState({ accounts: [] }));
		render(<EmailSettingsSection />);

		expect(screen.getByText('SETTINGS_EMAIL_NO_ACCOUNTS')).toBeInTheDocument();
	});

	it('renders the loading state', () => {
		mockUseEmailSettings.mockReturnValue(createState({ isLoading: true }));
		render(<EmailSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_EMAIL_ADD_GOOGLE')).not.toBeInTheDocument();
	});

	it('renders the backend load error verbatim (feature disabled)', () => {
		mockUseEmailSettings.mockReturnValue(
			createState({ hasError: true, loadErrorMessage: 'integração desligada: sem chave' })
		);
		render(<EmailSettingsSection />);

		expect(screen.getByText('integração desligada: sem chave')).toBeInTheDocument();
	});
});
