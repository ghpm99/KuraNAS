import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
	createGoogleAuthUrl,
	deleteEmailAccount,
	startMicrosoftDeviceCode,
	updateEmailAccountSyncEnabled,
} from '@/service/email';
import useEmailSettings from './useEmailSettings';

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

jest.mock('@/service/email', () => ({
	getEmailAccounts: jest.fn(),
	deleteEmailAccount: jest.fn(),
	updateEmailAccountSyncEnabled: jest.fn(),
	createGoogleAuthUrl: jest.fn(),
	startMicrosoftDeviceCode: jest.fn(),
	getMicrosoftDeviceCodeStatus: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedCreateGoogleAuthUrl = createGoogleAuthUrl as jest.Mock;
const mockedStartMicrosoftDeviceCode = startMicrosoftDeviceCode as jest.Mock;
const mockedUpdateSyncEnabled = updateEmailAccountSyncEnabled as jest.Mock;
const mockedDeleteEmailAccount = deleteEmailAccount as jest.Mock;

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
];

type QueryState = {
	data?: unknown;
	isLoading?: boolean;
	isError?: boolean;
	error?: unknown;
};

const setupQueries = (accounts: QueryState, deviceStatus: QueryState = {}) => {
	mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: string[] }) => {
		if (queryKey[0] === 'email-accounts') {
			return { isLoading: false, isError: false, ...accounts };
		}
		return { isLoading: false, isError: false, ...deviceStatus };
	});
};

describe('components/settings/useEmailSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		setupQueries({ data: sampleAccounts });
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

	it('exposes the linked accounts', () => {
		const { result } = renderHook(() => useEmailSettings());

		expect(result.current.accounts).toHaveLength(1);
		expect(result.current.deviceCode).toBeNull();
		expect(result.current.deviceStatus).toBeNull();
		expect(result.current.hasError).toBe(false);
	});

	it('shows the backend message verbatim when loading fails (feature disabled)', () => {
		setupQueries({
			data: undefined,
			isError: true,
			error: { response: { data: { error: 'integração desligada: sem chave' } } },
		});
		const { result } = renderHook(() => useEmailSettings());

		expect(result.current.hasError).toBe(true);
		expect(result.current.loadErrorMessage).toBe('integração desligada: sem chave');
	});

	it('falls back to the catalog key when the load failure has no backend message', () => {
		setupQueries({ data: undefined, isError: true, error: new Error('network') });
		const { result } = renderHook(() => useEmailSettings());

		expect(result.current.loadErrorMessage).toBe('ERROR_EMAIL_ACCOUNTS_LOAD');
	});

	it('opens the google consent url in a new tab', async () => {
		mockedCreateGoogleAuthUrl.mockResolvedValue({ auth_url: 'https://accounts.google.com/x' });
		const openSpy = jest.spyOn(window, 'open').mockReturnValue(null);
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleLinkGoogle();
		});

		expect(openSpy).toHaveBeenCalledWith('https://accounts.google.com/x', '_blank', 'noopener');
		openSpy.mockRestore();
	});

	it('surfaces the backend error when the google link fails', async () => {
		mockedCreateGoogleAuthUrl.mockRejectedValue({
			response: { data: { error: 'client não configurado' } },
		});
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleLinkGoogle();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('client não configurado', {
			variant: 'error',
		});
	});

	it('stores the device code and reports pending status after starting the microsoft flow', async () => {
		const dto = {
			user_code: 'ABC123',
			verification_uri: 'https://microsoft.com/devicelogin',
			expires_in: 900,
			message: 'abra e digite',
		};
		mockedStartMicrosoftDeviceCode.mockResolvedValue(dto);
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleLinkMicrosoft();
		});

		expect(result.current.deviceCode).toEqual(dto);
		expect(result.current.deviceStatus).toBe('pending');
	});

	it('invalidates the accounts when the device link completes', async () => {
		mockedStartMicrosoftDeviceCode.mockResolvedValue({
			user_code: 'ABC123',
			verification_uri: 'v',
			expires_in: 900,
			message: 'm',
		});
		setupQueries({ data: sampleAccounts }, { data: { status: 'linked' } });
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleLinkMicrosoft();
		});

		expect(result.current.deviceStatus).toBe('linked');
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['email-accounts'] });
	});

	it('falls back to the generic link error without a backend message', async () => {
		mockedStartMicrosoftDeviceCode.mockRejectedValue(new Error('boom'));
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleLinkMicrosoft();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('EMAIL_ACCOUNT_LINK_FAILED', {
			variant: 'error',
		});
	});

	it('toggles sync and invalidates the listing', async () => {
		mockedUpdateSyncEnabled.mockResolvedValue(undefined);
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleToggleSync(1, false);
		});

		expect(mockedUpdateSyncEnabled).toHaveBeenCalledWith(1, false);
		expect(mockInvalidateQueries).toHaveBeenCalled();
	});

	it('surfaces toggle failures through the snackbar', async () => {
		mockedUpdateSyncEnabled.mockRejectedValue({
			response: { data: { error: 'conta não encontrada' } },
		});
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleToggleSync(9, true);
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('conta não encontrada', {
			variant: 'error',
		});
	});

	it('removes an account and shows the backend confirmation', async () => {
		mockedDeleteEmailAccount.mockResolvedValue('conta removida');
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleRemove(1);
		});

		expect(mockedDeleteEmailAccount).toHaveBeenCalledWith(1);
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('conta removida', { variant: 'success' });
		expect(mockInvalidateQueries).toHaveBeenCalled();
	});

	it('surfaces remove failures through the snackbar', async () => {
		mockedDeleteEmailAccount.mockRejectedValue(new Error('boom'));
		const { result } = renderHook(() => useEmailSettings());

		await act(async () => {
			await result.current.handleRemove(1);
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('EMAIL_ACCOUNT_LINK_FAILED', {
			variant: 'error',
		});
	});
});
