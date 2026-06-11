import useI18n from '@/components/i18n/provider/i18nContext';
import {
	createAllowedIP,
	deleteAllowedIP,
	getAllowedIPs,
	getClientIP,
	updateAllowedIP,
} from '@/service/accessControl';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

const useAccessControlSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	const [cidr, setCidr] = useState('');
	const [label, setLabel] = useState('');

	const allowedIPsQuery = useQuery({
		queryKey: ['access-control', 'ips'],
		queryFn: getAllowedIPs,
		retry: false,
	});
	const clientIPQuery = useQuery({
		queryKey: ['access-control', 'client-ip'],
		queryFn: getClientIP,
		retry: false,
	});

	const invalidate = useCallback(() => {
		void queryClient.invalidateQueries({ queryKey: ['access-control', 'ips'] });
	}, [queryClient]);

	const notifyError = useCallback(
		(error: unknown) => {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('ACCESS_CONTROL_SAVE_ERROR'), { variant: 'error' });
		},
		[enqueueSnackbar, t]
	);

	const createMutation = useMutation({
		mutationFn: createAllowedIP,
		onSuccess: invalidate,
	});
	const updateMutation = useMutation({
		mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
			updateAllowedIP(id, { enabled }),
		onSuccess: invalidate,
	});
	const deleteMutation = useMutation({
		mutationFn: deleteAllowedIP,
		onSuccess: invalidate,
	});

	const handleCreate = useCallback(
		async (cidrValue: string, labelValue: string) => {
			const trimmed = cidrValue.trim();
			if (trimmed.length === 0) {
				return;
			}
			try {
				await createMutation.mutateAsync({ cidr: trimmed, label: labelValue.trim() });
				setCidr('');
				setLabel('');
				enqueueSnackbar(t('SETTINGS_ACCESS_CONTROL_SAVED'), { variant: 'success' });
			} catch (error) {
				notifyError(error);
			}
		},
		[createMutation, enqueueSnackbar, notifyError, t]
	);

	const handleAdd = useCallback(() => handleCreate(cidr, label), [cidr, handleCreate, label]);

	const handleAddCurrentDevice = useCallback(async () => {
		const clientIP = clientIPQuery.data?.ip;
		if (!clientIP) {
			return;
		}
		await handleCreate(clientIP, label);
	}, [clientIPQuery.data?.ip, handleCreate, label]);

	const handleToggle = useCallback(
		async (id: number, enabled: boolean) => {
			try {
				await updateMutation.mutateAsync({ id, enabled });
				enqueueSnackbar(t('SETTINGS_ACCESS_CONTROL_SAVED'), { variant: 'success' });
			} catch (error) {
				notifyError(error);
			}
		},
		[enqueueSnackbar, notifyError, t, updateMutation]
	);

	const handleDelete = useCallback(
		async (id: number) => {
			try {
				await deleteMutation.mutateAsync(id);
				enqueueSnackbar(t('SETTINGS_ACCESS_CONTROL_SAVED'), { variant: 'success' });
			} catch (error) {
				notifyError(error);
			}
		},
		[deleteMutation, enqueueSnackbar, notifyError, t]
	);

	return {
		t,
		entries: allowedIPsQuery.data ?? [],
		clientIP: clientIPQuery.data?.ip ?? '',
		isLoading: allowedIPsQuery.isLoading,
		isSaving: createMutation.isPending || updateMutation.isPending || deleteMutation.isPending,
		hasError: allowedIPsQuery.isError,
		cidr,
		label,
		setCidr,
		setLabel,
		handleAdd,
		handleAddCurrentDevice,
		handleToggle,
		handleDelete,
	};
};

export default useAccessControlSettings;
