import useI18n from '@/components/i18n/provider/i18nContext';
import {
	createGoogleAuthUrl,
	deleteEmailAccount,
	getEmailAccounts,
	getEmailAiProvider,
	getMicrosoftDeviceCodeStatus,
	startMicrosoftDeviceCode,
	updateEmailAccountSyncEnabled,
	updateEmailAiProvider,
} from '@/service/email';
import type { EmailAiProvider, EmailDeviceCodeDto, EmailDeviceCodeStatus } from '@/types/email';

// Cloud providers send e-mail content off-box, so the UI shows a privacy warning
// before they are selected. Local Ollama and the default chain do not.
const CLOUD_PROVIDERS: ReadonlySet<EmailAiProvider> = new Set<EmailAiProvider>([
	'openai',
	'anthropic',
]);

export const isCloudEmailProvider = (provider: EmailAiProvider): boolean =>
	CLOUD_PROVIDERS.has(provider);
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useEffect, useState } from 'react';

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

const DEVICE_STATUS_POLL_MS = 3000;

const useEmailSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	const [deviceCode, setDeviceCode] = useState<EmailDeviceCodeDto | null>(null);

	const accountsQuery = useQuery({
		queryKey: ['email-accounts'],
		queryFn: getEmailAccounts,
		retry: false,
	});

	const deviceStatusQuery = useQuery({
		queryKey: ['email-device-status'],
		queryFn: getMicrosoftDeviceCodeStatus,
		enabled: deviceCode !== null,
		refetchInterval: (query) =>
			!query.state.data || query.state.data.status === 'pending'
				? DEVICE_STATUS_POLL_MS
				: false,
		retry: false,
	});

	const deviceStatus: EmailDeviceCodeStatus | null = deviceCode
		? (deviceStatusQuery.data?.status ?? 'pending')
		: null;

	const providerQuery = useQuery({
		queryKey: ['email-ai-provider'],
		queryFn: getEmailAiProvider,
		retry: false,
	});

	const invalidateAccounts = useCallback(() => {
		void queryClient.invalidateQueries({ queryKey: ['email-accounts'] });
	}, [queryClient]);

	useEffect(() => {
		if (deviceStatus === 'linked') {
			invalidateAccounts();
		}
	}, [deviceStatus, invalidateAccounts]);

	const notifyError = useCallback(
		(error: unknown) => {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('EMAIL_ACCOUNT_LINK_FAILED'), {
				variant: 'error',
			});
		},
		[enqueueSnackbar, t]
	);

	const syncMutation = useMutation({
		mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
			updateEmailAccountSyncEnabled(id, enabled),
		onSuccess: invalidateAccounts,
	});
	const deleteMutation = useMutation({
		mutationFn: deleteEmailAccount,
		onSuccess: invalidateAccounts,
	});
	const providerMutation = useMutation({
		mutationFn: updateEmailAiProvider,
		onSuccess: (data) => {
			queryClient.setQueryData(['email-ai-provider'], data);
		},
	});

	const handleLinkGoogle = useCallback(async () => {
		try {
			const { auth_url: authUrl } = await createGoogleAuthUrl();
			window.open(authUrl, '_blank', 'noopener');
		} catch (error) {
			notifyError(error);
		}
	}, [notifyError]);

	const handleLinkMicrosoft = useCallback(async () => {
		try {
			setDeviceCode(await startMicrosoftDeviceCode());
		} catch (error) {
			notifyError(error);
		}
	}, [notifyError]);

	const handleToggleSync = useCallback(
		async (id: number, enabled: boolean) => {
			try {
				await syncMutation.mutateAsync({ id, enabled });
			} catch (error) {
				notifyError(error);
			}
		},
		[notifyError, syncMutation]
	);

	const handleRemove = useCallback(
		async (id: number) => {
			try {
				const message = await deleteMutation.mutateAsync(id);
				if (message) {
					// The backend answers the confirmation already translated.
					enqueueSnackbar(message, { variant: 'success' });
				}
			} catch (error) {
				notifyError(error);
			}
		},
		[deleteMutation, enqueueSnackbar, notifyError]
	);

	const handleChangeProvider = useCallback(
		async (provider: EmailAiProvider) => {
			try {
				await providerMutation.mutateAsync(provider);
			} catch (error) {
				notifyError(error);
			}
		},
		[notifyError, providerMutation]
	);

	return {
		t,
		accounts: accountsQuery.data ?? [],
		isLoading: accountsQuery.isLoading,
		isSaving: syncMutation.isPending || deleteMutation.isPending,
		hasError: accountsQuery.isError,
		// A disabled feature (no EMAIL_TOKEN_KEY) explains itself through the
		// backend's own translated message.
		loadErrorMessage:
			extractBackendError(accountsQuery.error) ?? t('ERROR_EMAIL_ACCOUNTS_LOAD'),
		deviceCode,
		deviceStatus,
		aiProvider: providerQuery.data?.provider ?? 'ollama',
		isProviderSaving: providerMutation.isPending,
		handleLinkGoogle,
		handleLinkMicrosoft,
		handleToggleSync,
		handleRemove,
		handleChangeProvider,
	};
};

export default useEmailSettings;
