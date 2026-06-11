import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
	createAllowedIP,
	deleteAllowedIP,
	updateAllowedIP,
} from '@/service/accessControl';
import useAccessControlSettings from './useAccessControlSettings';

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
		t: (key: string, params?: Record<string, string>) =>
			params ? `${key}:${Object.values(params).join(',')}` : key,
	}),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/accessControl', () => ({
	getAllowedIPs: jest.fn(),
	getClientIP: jest.fn(),
	createAllowedIP: jest.fn(),
	updateAllowedIP: jest.fn(),
	deleteAllowedIP: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedCreateAllowedIP = createAllowedIP as jest.Mock;
const mockedUpdateAllowedIP = updateAllowedIP as jest.Mock;
const mockedDeleteAllowedIP = deleteAllowedIP as jest.Mock;

const sampleEntries = [
	{ id: 1, cidr: '192.168.1.10/32', label: 'notebook', enabled: true, created_at: '2026-06-11' },
	{ id: 2, cidr: '192.168.1.0/24', label: '', enabled: false, created_at: '2026-06-11' },
];

describe('components/settings/useAccessControlSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: string[] }) => {
			if (queryKey[1] === 'client-ip') {
				return { data: { ip: '192.168.1.77' }, isLoading: false, isError: false };
			}
			return { data: sampleEntries, isLoading: false, isError: false };
		});
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

	it('exposes entries and the requester client IP', () => {
		const { result } = renderHook(() => useAccessControlSettings());

		expect(result.current.entries).toHaveLength(2);
		expect(result.current.clientIP).toBe('192.168.1.77');
		expect(result.current.isLoading).toBe(false);
		expect(result.current.hasError).toBe(false);
	});

	it('creates an entry with trimmed values and resets the form', async () => {
		mockedCreateAllowedIP.mockResolvedValue(sampleEntries[0]);
		const { result } = renderHook(() => useAccessControlSettings());

		act(() => {
			result.current.setCidr('  192.168.1.10  ');
			result.current.setLabel(' notebook ');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockedCreateAllowedIP).toHaveBeenCalledWith({
			cidr: '192.168.1.10',
			label: 'notebook',
		});
		expect(mockInvalidateQueries).toHaveBeenCalled();
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_ACCESS_CONTROL_SAVED', {
			variant: 'success',
		});
		expect(result.current.cidr).toBe('');
		expect(result.current.label).toBe('');
	});

	it('ignores an empty cidr on add', async () => {
		const { result } = renderHook(() => useAccessControlSettings());

		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockedCreateAllowedIP).not.toHaveBeenCalled();
	});

	it('registers the current device using the detected client IP', async () => {
		mockedCreateAllowedIP.mockResolvedValue(sampleEntries[0]);
		const { result } = renderHook(() => useAccessControlSettings());

		await act(async () => {
			await result.current.handleAddCurrentDevice();
		});

		expect(mockedCreateAllowedIP).toHaveBeenCalledWith({ cidr: '192.168.1.77', label: '' });
	});

	it('shows the backend error verbatim when the create fails', async () => {
		mockedCreateAllowedIP.mockRejectedValue({
			response: { data: { error: 'mensagem traduzida do backend' } },
		});
		const { result } = renderHook(() => useAccessControlSettings());

		act(() => {
			result.current.setCidr('bogus');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('mensagem traduzida do backend', {
			variant: 'error',
		});
	});

	it('falls back to the generic error key when the failure has no backend message', async () => {
		mockedCreateAllowedIP.mockRejectedValue(new Error('network down'));
		const { result } = renderHook(() => useAccessControlSettings());

		act(() => {
			result.current.setCidr('192.168.1.10');
		});
		await act(async () => {
			await result.current.handleAdd();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACCESS_CONTROL_SAVE_ERROR', {
			variant: 'error',
		});
	});

	it('toggles and deletes entries', async () => {
		mockedUpdateAllowedIP.mockResolvedValue({ ...sampleEntries[0], enabled: false });
		mockedDeleteAllowedIP.mockResolvedValue(undefined);
		const { result } = renderHook(() => useAccessControlSettings());

		await act(async () => {
			await result.current.handleToggle(1, false);
		});
		expect(mockedUpdateAllowedIP).toHaveBeenCalledWith(1, { enabled: false });

		await act(async () => {
			await result.current.handleDelete(2);
		});
		expect(mockedDeleteAllowedIP).toHaveBeenCalledWith(2);
		expect(mockInvalidateQueries).toHaveBeenCalledTimes(2);
	});

	it('surfaces toggle/delete failures through the snackbar', async () => {
		mockedUpdateAllowedIP.mockRejectedValue({ response: { data: { error: 'erro do servidor' } } });
		mockedDeleteAllowedIP.mockRejectedValue(new Error('boom'));
		const { result } = renderHook(() => useAccessControlSettings());

		await act(async () => {
			await result.current.handleToggle(1, false);
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('erro do servidor', { variant: 'error' });

		await act(async () => {
			await result.current.handleDelete(2);
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACCESS_CONTROL_SAVE_ERROR', {
			variant: 'error',
		});
	});
});
