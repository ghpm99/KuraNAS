import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import AIProvidersSettingsSection from './AIProvidersSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useAIProvidersSettings +
// service/aiProviders.ts + service/ollama.ts run, so each command asserts the
// exact endpoint/payload the backend aiproviders/ollama handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

const ollamaProvider = {
	name: 'ollama',
	enabled: true,
	model: 'llama3',
	base_url: 'http://localhost:11434',
	priority: 1,
	params: { timeout_seconds: 120, keep_alive: '5m' },
	requires_api_key: false,
	api_key_configured: false,
};

const ollamaStatus = {
	reachable: true,
	version: '0.5.0',
	base_url: 'http://localhost:11434',
	models: [{ name: 'llama3.1', size: 1000000, parameter_size: '7B', quantization_level: 'Q4' }],
};

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<AIProvidersSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/AIProvidersSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/ai/providers') return Promise.resolve({ data: [ollamaProvider] });
			if (url === '/ai/ollama/status') return Promise.resolve({ data: ollamaStatus });
			if (url.startsWith('/jobs/')) return Promise.resolve({ data: { id: 5, status: 'running' } });
			return Promise.reject(new Error(`unexpected GET ${url}`));
		});
		mockedApi.put.mockResolvedValue({ data: ollamaProvider });
		mockedApi.post.mockResolvedValue({ data: { job_id: 5 } });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('saving a provider issues PUT /ai/providers/:name with the edited request', async () => {
		renderSection();
		const model = await screen.findByLabelText('AI_PROVIDERS_MODEL');

		fireEvent.change(model, { target: { value: 'llama3.1' } });
		fireEvent.click(screen.getByText('AI_PROVIDERS_SAVE'));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith(
				'/ai/providers/ollama',
				expect.objectContaining({ enabled: true, model: 'llama3.1', base_url: 'http://localhost:11434', priority: 1 })
			)
		);
	});

	it('pulling a model issues POST /ai/ollama/models/pull with the typed model', async () => {
		renderSection();
		const pullInput = await screen.findByLabelText('AI_OLLAMA_PULL_PLACEHOLDER');

		fireEvent.change(pullInput, { target: { value: 'mistral' } });
		fireEvent.click(screen.getByText('AI_OLLAMA_PULL'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/ai/ollama/models/pull', { model: 'mistral' })
		);
	});

	it('deleting a model issues DELETE /ai/ollama/models/:name', async () => {
		renderSection();
		await screen.findByText('llama3.1');

		fireEvent.click(screen.getByText('AI_OLLAMA_DELETE_MODEL'));

		await waitFor(() =>
			expect(mockedApi.delete).toHaveBeenCalledWith('/ai/ollama/models/llama3.1')
		);
	});
});
