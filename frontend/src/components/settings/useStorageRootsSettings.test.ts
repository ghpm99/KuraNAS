import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
	createStorageRoot,
	deleteStorageRoot,
	updateStorageRoot,
} from '@/service/storageRoots';
import useStorageRootsSettings from './useStorageRootsSettings';

const mockEnqueueSnackbar = jest.fn();
const mockInvalidateQueries = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/storageRoots', () => ({
	getStorageRoots: jest.fn(),
	createStorageRoot: jest.fn(),
	updateStorageRoot: jest.fn(),
	deleteStorageRoot: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedCreateStorageRoot = createStorageRoot as jest.Mock;
const mockedUpdateStorageRoot = updateStorageRoot as jest.Mock;
const mockedDeleteStorageRoot = deleteStorageRoot as jest.Mock;

const sampleRoots = [
	{ id: 1, path: '/mnt/dados', label: 'Dados', enabled: true, created_at: '2026-06-12' },
	{ id: 2, path: '/mnt/midia', label: 'Midia', enabled: false, created_at: '2026-06-12' },
];

describe('components/settings/useStorageRootsSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockedUseQuery.mockReturnValue({ data: sampleRoots, isLoading: false, isError: false });
		// Run the real mutationFn + onSuccess so the hook is exercised end to end.
		mockedUseMutation.mockImplementation(
			({
				mutationFn,
				onSuccess,
			}: {
				mutationFn: (args: unknown) => Promise<unknown>;
				onSuccess?: (result: unknown) => void;
			}) => ({
				mutateAsync: async (args: unknown) => {
					const result = await mutationFn(args);
					onSuccess?.(result);
					return result;
				},
				isPending: false,
			})
		);
	});

	it('exposes roots and flags the first as primary', () => {
		const { result } = renderHook(() => useStorageRootsSettings());

		expect(result.current.roots).toHaveLength(2);
		expect(result.current.primaryRootId).toBe(1);
		expect(result.current.isLoading).toBe(false);
		expect(result.current.hasError).toBe(false);
	});

	it('has no primary id without registered roots', () => {
		mockedUseQuery.mockReturnValue({ data: undefined, isLoading: false, isError: false });
		const { result } = renderHook(() => useStorageRootsSettings());

		expect(result.current.roots).toHaveLength(0);
		expect(result.current.primaryRootId).toBeUndefined();
	});

	it('creates a root with trimmed values and resets the form', async () => {
		mockedCreateStorageRoot.mockResolvedValue(sampleRoots[1]);
		const { result } = renderHook(() => useStorageRootsSettings());

		act(() => {
			result.current.setPath('  /mnt/midia  ');
			result.current.setLabel(' Midia ');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockedCreateStorageRoot).toHaveBeenCalledWith({ path: '/mnt/midia', label: 'Midia' });
		expect(mockInvalidateQueries).toHaveBeenCalled();
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_STORAGE_ROOTS_SAVED', {
			variant: 'success',
		});
		expect(result.current.path).toBe('');
		expect(result.current.label).toBe('');
	});

	it('ignores an empty path on add', async () => {
		const { result } = renderHook(() => useStorageRootsSettings());

		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockedCreateStorageRoot).not.toHaveBeenCalled();
	});

	it('shows the backend error verbatim when the create fails', async () => {
		mockedCreateStorageRoot.mockRejectedValue({
			response: { data: { error: 'mensagem traduzida do backend' } },
		});
		const { result } = renderHook(() => useStorageRootsSettings());

		act(() => {
			result.current.setPath('/mnt/sobreposto');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('mensagem traduzida do backend', {
			variant: 'error',
		});
	});

	it('falls back to the generic error key when the failure has no backend message', async () => {
		mockedCreateStorageRoot.mockRejectedValue(new Error('network down'));
		const { result } = renderHook(() => useStorageRootsSettings());

		act(() => {
			result.current.setPath('/mnt/novo');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_STORAGE_ROOTS_SAVE_ERROR', {
			variant: 'error',
		});
	});

	it('toggles and deletes roots', async () => {
		mockedUpdateStorageRoot.mockResolvedValue({ ...sampleRoots[1], enabled: true });
		mockedDeleteStorageRoot.mockResolvedValue(undefined);
		const { result } = renderHook(() => useStorageRootsSettings());

		await act(async () => {
			await result.current.handleToggle(2, true);
		});
		expect(mockedUpdateStorageRoot).toHaveBeenCalledWith(2, { enabled: true });

		await act(async () => {
			await result.current.handleDelete(2);
		});
		expect(mockedDeleteStorageRoot).toHaveBeenCalledWith(2);
		expect(mockInvalidateQueries).toHaveBeenCalledTimes(4);
	});

	it('surfaces toggle/delete failures through the snackbar', async () => {
		mockedUpdateStorageRoot.mockRejectedValue({ response: { data: { error: 'erro do servidor' } } });
		mockedDeleteStorageRoot.mockRejectedValue(new Error('boom'));
		const { result } = renderHook(() => useStorageRootsSettings());

		await act(async () => {
			await result.current.handleToggle(2, true);
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('erro do servidor', { variant: 'error' });

		await act(async () => {
			await result.current.handleDelete(2);
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_STORAGE_ROOTS_SAVE_ERROR', {
			variant: 'error',
		});
	});
});
