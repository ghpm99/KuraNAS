import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { updateAIProvider } from '@/service/aiProviders';
import { deleteOllamaModel, pullOllamaModel } from '@/service/ollama';
import useAIProvidersSettings from './useAIProvidersSettings';

const mockEnqueueSnackbar = jest.fn();
const mockInvalidateQueries = jest.fn().mockResolvedValue(undefined);
const mockSetQueryData = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/aiProviders', () => ({
	getAIProviders: jest.fn(),
	updateAIProvider: jest.fn(),
}));

jest.mock('@/service/ollama', () => ({
	getOllamaStatus: jest.fn(),
	pullOllamaModel: jest.fn(),
	deleteOllamaModel: jest.fn(),
}));

jest.mock('@/service/jobs', () => ({
	getJob: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedUpdateAIProvider = updateAIProvider as jest.Mock;
const mockedPullModel = pullOllamaModel as jest.Mock;
const mockedDeleteModel = deleteOllamaModel as jest.Mock;

const providersData = [
	{
		name: 'ollama',
		enabled: false,
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
];

const ollamaStatus = {
	reachable: true,
	version: '0.5.0',
	base_url: 'http://localhost:11434',
	models: [{ name: 'llama3.1:latest', size: 100, digest: 'abc', modified_at: '' }],
};

let jobData: unknown;

describe('components/settings/useAIProvidersSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		jobData = undefined;
		mockedUseQueryClient.mockReturnValue({
			setQueryData: (_key: unknown, updater: unknown) => {
				mockSetQueryData(_key, updater);
				if (typeof updater === 'function') {
					(updater as (current: unknown) => unknown)(providersData);
				}
			},
			invalidateQueries: mockInvalidateQueries,
		});
		mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: unknown[] }) => {
			if (queryKey[0] === 'ai-providers') {
				return { data: providersData, isLoading: false, isError: false };
			}
			if (queryKey[0] === 'ollama-status') {
				return { data: ollamaStatus, isLoading: false };
			}
			return { data: jobData };
		});
		mockedUseMutation.mockImplementation(
			({ mutationFn, onSuccess }: { mutationFn: (arg: unknown) => Promise<unknown>; onSuccess?: (r: unknown) => void }) => ({
				mutateAsync: async (arg: unknown) => {
					const result = await mutationFn(arg);
					onSuccess?.(result);
					return result;
				},
				isPending: false,
			})
		);
		mockedUpdateAIProvider.mockResolvedValue(providersData[0]);
		mockedPullModel.mockResolvedValue({ job_id: 99 });
		mockedDeleteModel.mockResolvedValue(undefined);
	});

	it('merges providers with local edits via setField and setParam', () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setField('ollama', 'model', 'qwen2.5');
			result.current.setParam('ollama', 'timeout_seconds', 300);
			// second param edit merges onto the already-edited params
			result.current.setParam('ollama', 'max_retries', 5);
		});

		const ollama = result.current.providers.find((p) => p.name === 'ollama');
		expect(ollama?.model).toBe('qwen2.5');
		expect(ollama?.params.timeout_seconds).toBe(300);
		expect(ollama?.params.max_retries).toBe(5);
		// untouched param preserved
		expect(ollama?.params.keep_alive).toBe('5m');
	});

	it('toggles a provider enabled and persists the override', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.toggleEnabled('ollama', true);
		});

		expect(mockedUpdateAIProvider).toHaveBeenCalledWith(
			'ollama',
			expect.objectContaining({ enabled: true, model: 'llama3.1' })
		);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_PROVIDERS_ENABLED', { variant: 'success' });
	});

	it('disables a provider with the matching message', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.toggleEnabled('ollama', false);
		});

		expect(mockedUpdateAIProvider).toHaveBeenCalledWith(
			'ollama',
			expect.objectContaining({ enabled: false })
		);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_PROVIDERS_DISABLED', { variant: 'success' });
	});

	it('saves edited provider fields', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setParam('openai', 'timeout_seconds', 45);
		});
		await act(async () => {
			await result.current.saveProvider('openai');
		});

		expect(mockedUpdateAIProvider).toHaveBeenCalledWith(
			'openai',
			expect.objectContaining({ params: expect.objectContaining({ timeout_seconds: 45 }) })
		);
	});

	it('starts a model pull', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setPullModelName('llama3.1');
		});
		await act(async () => {
			await result.current.handlePull();
		});

		expect(mockedPullModel).toHaveBeenCalledWith('llama3.1');
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_OLLAMA_PULL_STARTED', { variant: 'info' });
	});

	it('deletes a model and refreshes status', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.handleDeleteModel('llama3.1:latest');
		});

		expect(mockedDeleteModel).toHaveBeenCalledWith('llama3.1:latest');
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['ollama-status'] });
	});

	it('surfaces an error snackbar when saving fails', async () => {
		mockedUpdateAIProvider.mockRejectedValue(new Error('boom'));
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.saveProvider('ollama');
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_PROVIDERS_SAVE_ERROR', { variant: 'error' });
	});

	it('surfaces an error snackbar when the pull request fails', async () => {
		mockedPullModel.mockRejectedValue(new Error('down'));
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setPullModelName('llama3.1');
		});
		await act(async () => {
			await result.current.handlePull();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_OLLAMA_PULL_ERROR', { variant: 'error' });
	});

	it('does not pull when the model name is blank', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.handlePull();
		});

		expect(mockedPullModel).not.toHaveBeenCalled();
	});

	it('surfaces an error snackbar when deleting a model fails', async () => {
		mockedDeleteModel.mockRejectedValue(new Error('nope'));
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.handleDeleteModel('llama3.1:latest');
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_OLLAMA_MODEL_DELETE_ERROR', { variant: 'error' });
	});

	it('finishes the pull when the tracked job completes', async () => {
		jobData = { id: 99, type: 'ollama_pull', status: 'completed', progress: { progress: 100 } };
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setPullModelName('llama3.1');
		});
		await act(async () => {
			await result.current.handlePull();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_OLLAMA_PULL_COMPLETED', { variant: 'success' });
	});

	it('ignores toggle and save for an unknown provider', async () => {
		const { result } = renderHook(() => useAIProvidersSettings());

		await act(async () => {
			await result.current.toggleEnabled('bogus' as never, true);
			await result.current.saveProvider('bogus' as never);
		});

		expect(mockedUpdateAIProvider).not.toHaveBeenCalled();
	});

	it('handles missing provider and status data gracefully', () => {
		mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: unknown[] }) => {
			if (queryKey[0] === 'ai-providers') {
				return { data: undefined, isLoading: true, isError: false };
			}
			if (queryKey[0] === 'ollama-status') {
				return { data: undefined, isLoading: true };
			}
			return { data: jobData };
		});

		const { result } = renderHook(() => useAIProvidersSettings());

		expect(result.current.providers).toEqual([]);
		expect(result.current.ollamaStatus).toBeUndefined();
		expect(result.current.pullProgress).toBe(0);
	});

	it('reports a failed pull job', async () => {
		jobData = { id: 99, type: 'ollama_pull', status: 'failed', progress: { progress: 0 } };
		const { result } = renderHook(() => useAIProvidersSettings());

		act(() => {
			result.current.setPullModelName('llama3.1');
		});
		await act(async () => {
			await result.current.handlePull();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('AI_OLLAMA_PULL_ERROR', { variant: 'error' });
	});
});
