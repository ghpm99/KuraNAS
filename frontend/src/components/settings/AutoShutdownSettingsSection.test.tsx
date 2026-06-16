import { fireEvent, render, screen } from '@testing-library/react';
import AutoShutdownSettingsSection from './AutoShutdownSettingsSection';

const mockSetField = jest.fn();
const mockHandleSave = jest.fn();
const mockHandleSuggest = jest.fn();
const mockUseAutoShutdownSettings = jest.fn();

jest.mock('./useAutoShutdownSettings', () => ({
	__esModule: true,
	default: () => mockUseAutoShutdownSettings(),
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
	t: (key: string, vars?: Record<string, string>) =>
		vars ? `${key}:${JSON.stringify(vars)}` : key,
	form: {
		enabled: true,
		time: '03:00',
		grace_period_seconds: 60,
	},
	suggestion: null,
	isLoading: false,
	isSaving: false,
	isSuggesting: false,
	hasError: false,
	hasUnsavedChanges: true,
	setField: mockSetField,
	handleSave: mockHandleSave,
	handleSuggest: mockHandleSuggest,
	...overrides,
});

describe('components/settings/AutoShutdownSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseAutoShutdownSettings.mockReturnValue(createState());
	});

	it('renders the title and the enable toggle', () => {
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByText('SETTINGS_AUTO_SHUTDOWN_TITLE')).toBeInTheDocument();
		expect(screen.getByRole('switch')).toBeInTheDocument();
	});

	it('forwards form edits to the hook', () => {
		render(<AutoShutdownSettingsSection />);

		fireEvent.click(screen.getByRole('switch'));
		expect(mockSetField).toHaveBeenCalledWith('enabled', false);

		const grace = screen.getByRole('spinbutton');
		fireEvent.change(grace, { target: { value: '120' } });
		expect(mockSetField).toHaveBeenCalledWith('grace_period_seconds', 120);
	});

	it('triggers the suggest and save actions', () => {
		render(<AutoShutdownSettingsSection />);

		fireEvent.click(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SUGGEST_BUTTON'));
		expect(mockHandleSuggest).toHaveBeenCalled();

		fireEvent.click(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SAVE'));
		expect(mockHandleSave).toHaveBeenCalled();
	});

	it('shows the suggestion hint with the sample size', () => {
		mockUseAutoShutdownSettings.mockReturnValue(
			createState({ suggestion: { available: true, time: '02:30', sample_size: 7 } })
		);
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByText(/SETTINGS_AUTO_SHUTDOWN_SUGGESTION:/)).toHaveTextContent('02:30');
		expect(screen.getByText(/SETTINGS_AUTO_SHUTDOWN_SUGGESTION:/)).toHaveTextContent('7');
	});

	it('shows the empty suggestion message when unavailable', () => {
		mockUseAutoShutdownSettings.mockReturnValue(
			createState({ suggestion: { available: false, time: '', sample_size: 1 } })
		);
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SUGGESTION_EMPTY')).toBeInTheDocument();
	});

	it('disables the save button without unsaved changes', () => {
		mockUseAutoShutdownSettings.mockReturnValue(createState({ hasUnsavedChanges: false }));
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SAVE')).toBeDisabled();
	});

	it('disables the time/grace fields when the feature is off', () => {
		mockUseAutoShutdownSettings.mockReturnValue(
			createState({ form: { enabled: false, time: '03:00', grace_period_seconds: 60 } })
		);
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByRole('spinbutton')).toBeDisabled();
	});

	it('renders the loading state', () => {
		mockUseAutoShutdownSettings.mockReturnValue(createState({ isLoading: true }));
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('SETTINGS_AUTO_SHUTDOWN_SAVE')).not.toBeInTheDocument();
	});

	it('renders the load error alert', () => {
		mockUseAutoShutdownSettings.mockReturnValue(createState({ hasError: true }));
		render(<AutoShutdownSettingsSection />);

		expect(screen.getByText('SETTINGS_AUTO_SHUTDOWN_LOAD_ERROR')).toBeInTheDocument();
	});
});
