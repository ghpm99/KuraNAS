import { fireEvent, render, screen } from '@testing-library/react';
import AccessControlSettingsSection from './AccessControlSettingsSection';

const mockHandleAdd = jest.fn();
const mockHandleAddCurrentDevice = jest.fn();
const mockHandleToggle = jest.fn();
const mockHandleDelete = jest.fn();
const mockSetCidr = jest.fn();
const mockSetLabel = jest.fn();
const mockUseAccessControlSettings = jest.fn();

jest.mock('./useAccessControlSettings', () => ({
	__esModule: true,
	default: () => mockUseAccessControlSettings(),
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string, params?: Record<string, string>) =>
		params ? `${key}:${Object.values(params).join(',')}` : key,
	entries: [
		{ id: 1, cidr: '192.168.1.10/32', label: 'notebook', enabled: true, created_at: '2026-06-11' },
	],
	clientIP: '192.168.1.77',
	isLoading: false,
	isSaving: false,
	hasError: false,
	cidr: '',
	label: '',
	setCidr: mockSetCidr,
	setLabel: mockSetLabel,
	handleAdd: mockHandleAdd,
	handleAddCurrentDevice: mockHandleAddCurrentDevice,
	handleToggle: mockHandleToggle,
	handleDelete: mockHandleDelete,
	...overrides,
});

describe('components/settings/AccessControlSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseAccessControlSettings.mockReturnValue(createState());
	});

	it('renders entries with toggle and remove actions', () => {
		render(<AccessControlSettingsSection />);

		expect(screen.getByText('SETTINGS_ACCESS_CONTROL_TITLE')).toBeInTheDocument();
		expect(screen.getByText('192.168.1.10/32')).toBeInTheDocument();
		expect(screen.getByText('notebook')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_ACCESS_CONTROL_YOUR_IP:192.168.1.77')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('checkbox'));
		expect(mockHandleToggle).toHaveBeenCalledWith(1, false);

		fireEvent.click(screen.getByText('SETTINGS_ACCESS_CONTROL_REMOVE'));
		expect(mockHandleDelete).toHaveBeenCalledWith(1);

		fireEvent.click(screen.getByText('SETTINGS_ACCESS_CONTROL_ADD_CURRENT'));
		expect(mockHandleAddCurrentDevice).toHaveBeenCalled();
	});

	it('shows the empty warning when no IP is registered', () => {
		mockUseAccessControlSettings.mockReturnValue(createState({ entries: [] }));
		render(<AccessControlSettingsSection />);

		expect(screen.getByText('SETTINGS_ACCESS_CONTROL_EMPTY')).toBeInTheDocument();
	});

	it('disables the add button until a cidr is typed and submits it', () => {
		mockUseAccessControlSettings.mockReturnValue(createState({ cidr: '192.168.1.20' }));
		render(<AccessControlSettingsSection />);

		const addButton = screen.getByText('SETTINGS_ACCESS_CONTROL_ADD');
		fireEvent.click(addButton);
		expect(mockHandleAdd).toHaveBeenCalled();
	});

	it('renders the loading state', () => {
		mockUseAccessControlSettings.mockReturnValue(createState({ isLoading: true }));
		render(<AccessControlSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_ACCESS_CONTROL_ADD')).not.toBeInTheDocument();
	});

	it('renders the load error alert', () => {
		mockUseAccessControlSettings.mockReturnValue(createState({ hasError: true }));
		render(<AccessControlSettingsSection />);

		expect(screen.getByText('SETTINGS_ACCESS_CONTROL_LOAD_ERROR')).toBeInTheDocument();
	});
});
