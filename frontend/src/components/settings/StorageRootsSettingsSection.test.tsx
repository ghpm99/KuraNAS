import { fireEvent, render, screen } from '@testing-library/react';
import StorageRootsSettingsSection from './StorageRootsSettingsSection';

const mockHandleAdd = jest.fn();
const mockHandleToggle = jest.fn();
const mockHandleDelete = jest.fn();
const mockSetPath = jest.fn();
const mockSetLabel = jest.fn();
const mockUseStorageRootsSettings = jest.fn();

jest.mock('./useStorageRootsSettings', () => ({
	__esModule: true,
	default: () => mockUseStorageRootsSettings(),
}));

const sampleRoots = [
	{ id: 1, path: '/mnt/dados', label: 'Dados', enabled: true, created_at: '2026-06-12' },
	{ id: 2, path: '/mnt/midia', label: 'Midia', enabled: true, created_at: '2026-06-12' },
];

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string) => key,
	roots: sampleRoots,
	primaryRootId: 1,
	isLoading: false,
	isSaving: false,
	hasError: false,
	path: '',
	label: '',
	setPath: mockSetPath,
	setLabel: mockSetLabel,
	handleAdd: mockHandleAdd,
	handleToggle: mockHandleToggle,
	handleDelete: mockHandleDelete,
	...overrides,
});

describe('components/settings/StorageRootsSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseStorageRootsSettings.mockReturnValue(createState());
	});

	it('renders the roots with the primary protected', () => {
		render(<StorageRootsSettingsSection />);

		expect(screen.getByText('SETTINGS_STORAGE_ROOTS_TITLE')).toBeInTheDocument();
		expect(screen.getByText('/mnt/dados')).toBeInTheDocument();
		expect(screen.getByText('/mnt/midia')).toBeInTheDocument();
		expect(screen.getByText('SETTINGS_STORAGE_ROOTS_PRIMARY')).toBeInTheDocument();

		// The primary root's switch and remove button stay disabled.
		const switches = screen.getAllByRole('checkbox');
		expect(switches[0]).toBeDisabled();
		expect(switches[1]).not.toBeDisabled();

		const removeButtons = screen.getAllByText('SETTINGS_STORAGE_ROOTS_REMOVE');
		expect(removeButtons[0]).toBeDisabled();

		fireEvent.click(switches[1]!);
		expect(mockHandleToggle).toHaveBeenCalledWith(2, false);

		fireEvent.click(removeButtons[1]!);
		expect(mockHandleDelete).toHaveBeenCalledWith(2);
	});

	it('shows the empty warning when no root is registered', () => {
		mockUseStorageRootsSettings.mockReturnValue(createState({ roots: [], primaryRootId: undefined }));
		render(<StorageRootsSettingsSection />);

		expect(screen.getByText('SETTINGS_STORAGE_ROOTS_EMPTY')).toBeInTheDocument();
	});

	it('submits a new root when the path is filled', () => {
		mockUseStorageRootsSettings.mockReturnValue(createState({ path: '/mnt/backup' }));
		render(<StorageRootsSettingsSection />);

		fireEvent.click(screen.getByText('SETTINGS_STORAGE_ROOTS_ADD'));
		expect(mockHandleAdd).toHaveBeenCalled();
	});

	it('forwards typed values to the form state', () => {
		render(<StorageRootsSettingsSection />);

		const [pathInput, labelInput] = screen.getAllByRole('textbox');
		fireEvent.change(pathInput!, { target: { value: '/mnt/novo' } });
		expect(mockSetPath).toHaveBeenCalledWith('/mnt/novo');
		fireEvent.change(labelInput!, { target: { value: 'Novo' } });
		expect(mockSetLabel).toHaveBeenCalledWith('Novo');
	});

	it('renders the loading state', () => {
		mockUseStorageRootsSettings.mockReturnValue(createState({ isLoading: true }));
		render(<StorageRootsSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_STORAGE_ROOTS_ADD')).not.toBeInTheDocument();
	});

	it('renders the load error alert', () => {
		mockUseStorageRootsSettings.mockReturnValue(createState({ hasError: true }));
		render(<StorageRootsSettingsSection />);

		expect(screen.getByText('SETTINGS_STORAGE_ROOTS_LOAD_ERROR')).toBeInTheDocument();
	});
});
