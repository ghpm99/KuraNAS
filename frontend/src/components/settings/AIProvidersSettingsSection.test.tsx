import { fireEvent, render, screen } from '@testing-library/react';
import AIProvidersSettingsSection from './AIProvidersSettingsSection';
import useAIProvidersSettings from './useAIProvidersSettings';

jest.mock('./useAIProvidersSettings');

const mockedUseHook = useAIProvidersSettings as jest.Mock;

const baseHook = {
	t: (key: string) => key,
	providers: [
		{
			name: 'ollama',
			enabled: true,
			model: 'llama3.1',
			base_url: 'http://localhost:11434',
			priority: 0,
			params: { timeout_seconds: 120, keep_alive: '5m' },
			requires_api_key: false,
			api_key_configured: true,
		},
		{
			name: 'openai',
			enabled: false,
			model: 'gpt-4o-mini',
			base_url: 'https://api.openai.com/v1',
			priority: 1,
			params: { timeout_seconds: 30 },
			requires_api_key: true,
			api_key_configured: false,
		},
		{
			name: 'anthropic',
			enabled: false,
			model: 'claude-sonnet-4',
			base_url: '',
			priority: 2,
			params: {},
			requires_api_key: true,
			api_key_configured: true,
		},
	],
	isLoading: false,
	hasError: false,
	isSaving: false,
	ollamaStatus: {
		reachable: true,
		version: '0.5.0',
		base_url: 'http://localhost:11434',
		models: [{ name: 'llama3.1:latest', size: 1024, digest: 'abc', modified_at: '', parameter_size: '8B' }],
	},
	ollamaLoading: false,
	toggleEnabled: jest.fn(),
	setField: jest.fn(),
	setParam: jest.fn(),
	saveProvider: jest.fn(),
	pullModelName: '',
	setPullModelName: jest.fn(),
	handlePull: jest.fn(),
	isPulling: false,
	pullProgress: 0,
	handleDeleteModel: jest.fn(),
	isDeleting: false,
};

describe('components/settings/AIProvidersSettingsSection', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseHook.mockReturnValue(baseHook);
	});

	it('shows a loading state', () => {
		mockedUseHook.mockReturnValue({ ...baseHook, isLoading: true });
		render(<AIProvidersSettingsSection />);
		expect(screen.getByText('AI_PROVIDERS_TITLE')).toBeInTheDocument();
	});

	it('renders providers and the Ollama models panel', () => {
		render(<AIProvidersSettingsSection />);

		expect(screen.getByText('ollama')).toBeInTheDocument();
		expect(screen.getByText('openai')).toBeInTheDocument();
		// cloud provider without a key is flagged
		expect(screen.getByText('AI_PROVIDERS_NO_API_KEY')).toBeInTheDocument();
		// installed model is listed
		expect(screen.getByText('llama3.1:latest')).toBeInTheDocument();
	});

	it('saves a provider when the save button is clicked', () => {
		render(<AIProvidersSettingsSection />);

		const saveButtons = screen.getAllByText('AI_PROVIDERS_SAVE');
		fireEvent.click(saveButtons[0]!);

		expect(baseHook.saveProvider).toHaveBeenCalledWith('ollama');
	});

	it('triggers a model pull', () => {
		mockedUseHook.mockReturnValue({ ...baseHook, pullModelName: 'mistral' });
		render(<AIProvidersSettingsSection />);

		fireEvent.click(screen.getByText('AI_OLLAMA_PULL'));
		expect(baseHook.handlePull).toHaveBeenCalled();
	});

	it('deletes an installed model', () => {
		render(<AIProvidersSettingsSection />);

		fireEvent.click(screen.getByText('AI_OLLAMA_DELETE_MODEL'));
		expect(baseHook.handleDeleteModel).toHaveBeenCalledWith('llama3.1:latest');
	});

	it('shows the progress bar and disables inputs while pulling', () => {
		mockedUseHook.mockReturnValue({ ...baseHook, isPulling: true, pullProgress: 42 });
		const { container } = render(<AIProvidersSettingsSection />);

		expect(screen.getByText('AI_OLLAMA_PULLING')).toBeInTheDocument();
		expect(container.querySelector('.MuiLinearProgress-root')).toBeInTheDocument();
	});

	it('renders an unreachable daemon state with no models', () => {
		mockedUseHook.mockReturnValue({
			...baseHook,
			ollamaStatus: { reachable: false, base_url: 'http://localhost:11434', models: [] },
		});
		render(<AIProvidersSettingsSection />);

		expect(screen.getByText('AI_OLLAMA_UNREACHABLE')).toBeInTheDocument();
	});

	it('shows the empty-models hint when the daemon has none', () => {
		mockedUseHook.mockReturnValue({
			...baseHook,
			ollamaStatus: { ...baseHook.ollamaStatus, models: [] },
		});
		render(<AIProvidersSettingsSection />);

		expect(screen.getByText('AI_OLLAMA_NO_MODELS')).toBeInTheDocument();
	});

	it('hides the Ollama panel when the provider is disabled', () => {
		mockedUseHook.mockReturnValue({
			...baseHook,
			providers: baseHook.providers.map((p) =>
				p.name === 'ollama' ? { ...p, enabled: false } : p
			),
		});
		render(<AIProvidersSettingsSection />);

		expect(screen.queryByText('AI_OLLAMA_MODELS_TITLE')).not.toBeInTheDocument();
	});

	it('renders an error alert when loading fails', () => {
		mockedUseHook.mockReturnValue({ ...baseHook, hasError: true });
		render(<AIProvidersSettingsSection />);

		expect(screen.getByText('AI_PROVIDERS_LOAD_ERROR')).toBeInTheDocument();
	});

	it('formats model sizes and optional details', () => {
		mockedUseHook.mockReturnValue({
			...baseHook,
			ollamaStatus: {
				...baseHook.ollamaStatus,
				models: [
					{ name: 'zero', size: 0, digest: 'a', modified_at: '' },
					{ name: 'tiny', size: 5, digest: 'a2', modified_at: '' },
					{ name: 'bytes', size: 500, digest: 'b', modified_at: '' },
					{
						name: 'big',
						size: 2 * 1024 * 1024 * 1024,
						digest: 'c',
						modified_at: '',
						parameter_size: '7B',
						quantization_level: 'Q4_0',
					},
					{ name: 'huge', size: 1024 ** 5, digest: 'd', modified_at: '' },
				],
			},
		});
		render(<AIProvidersSettingsSection />);

		expect(screen.getByText('zero')).toBeInTheDocument();
		expect(screen.getByText(/500 B/)).toBeInTheDocument();
		expect(screen.getByText(/2.0 GB/)).toBeInTheDocument();
		expect(screen.getByText(/Q4_0/)).toBeInTheDocument();
	});

	it('renders empty inputs when a provider has no tuning params', () => {
		mockedUseHook.mockReturnValue({
			...baseHook,
			providers: [{ ...baseHook.providers[0], params: {} }],
		});
		render(<AIProvidersSettingsSection />);

		expect(screen.getByLabelText('AI_PROVIDERS_TIMEOUT')).toHaveValue(null);
		expect(screen.getByLabelText('AI_PROVIDERS_KEEP_ALIVE')).toHaveValue('');
	});

	it('edits provider fields and params through the inputs', () => {
		const { container } = render(<AIProvidersSettingsSection />);

		fireEvent.change(screen.getAllByLabelText('AI_PROVIDERS_MODEL')[0]!, {
			target: { value: 'qwen2.5' },
		});
		expect(baseHook.setField).toHaveBeenCalledWith('ollama', 'model', 'qwen2.5');

		fireEvent.change(screen.getAllByLabelText('AI_PROVIDERS_BASE_URL')[0]!, {
			target: { value: 'http://nas:11434' },
		});
		expect(baseHook.setField).toHaveBeenCalledWith('ollama', 'base_url', 'http://nas:11434');

		fireEvent.change(screen.getAllByLabelText('AI_PROVIDERS_PRIORITY')[0]!, {
			target: { value: '3' },
		});
		expect(baseHook.setField).toHaveBeenCalledWith('ollama', 'priority', 3);

		fireEvent.change(screen.getAllByLabelText('AI_PROVIDERS_TIMEOUT')[0]!, {
			target: { value: '300' },
		});
		expect(baseHook.setParam).toHaveBeenCalledWith('ollama', 'timeout_seconds', 300);

		fireEvent.change(screen.getByLabelText('AI_PROVIDERS_KEEP_ALIVE'), {
			target: { value: '30m' },
		});
		expect(baseHook.setParam).toHaveBeenCalledWith('ollama', 'keep_alive', '30m');

		// ollama starts enabled, so clicking its switch toggles it off
		const switches = container.querySelectorAll('input[type="checkbox"]');
		fireEvent.click(switches[0]!);
		expect(baseHook.toggleEnabled).toHaveBeenCalledWith('ollama', false);
	});
});
