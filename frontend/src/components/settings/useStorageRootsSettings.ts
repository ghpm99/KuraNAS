import useI18n from '@/components/i18n/provider/i18nContext';
import {
	createStorageRoot,
	deleteStorageRoot,
	getStorageRoots,
	updateStorageRoot,
} from '@/service/storageRoots';
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

const useStorageRootsSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	const [path, setPath] = useState('');
	const [label, setLabel] = useState('');

	const rootsQuery = useQuery({
		queryKey: ['storage-roots'],
		queryFn: getStorageRoots,
		retry: false,
	});

	const invalidate = useCallback(() => {
		void queryClient.invalidateQueries({ queryKey: ['storage-roots'] });
		// The tree's level zero changes with the registered roots.
		void queryClient.invalidateQueries({ queryKey: ['files'] });
	}, [queryClient]);

	const notifyError = useCallback(
		(error: unknown) => {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('SETTINGS_STORAGE_ROOTS_SAVE_ERROR'), {
				variant: 'error',
			});
		},
		[enqueueSnackbar, t]
	);

	const createMutation = useMutation({
		mutationFn: createStorageRoot,
		onSuccess: invalidate,
	});
	const updateMutation = useMutation({
		mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
			updateStorageRoot(id, { enabled }),
		onSuccess: invalidate,
	});
	const deleteMutation = useMutation({
		mutationFn: deleteStorageRoot,
		onSuccess: invalidate,
	});

	const handleAdd = useCallback(async () => {
		const trimmedPath = path.trim();
		if (trimmedPath.length === 0) {
			return;
		}
		try {
			await createMutation.mutateAsync({ path: trimmedPath, label: label.trim() });
			setPath('');
			setLabel('');
			enqueueSnackbar(t('SETTINGS_STORAGE_ROOTS_SAVED'), { variant: 'success' });
		} catch (error) {
			notifyError(error);
		}
	}, [createMutation, enqueueSnackbar, label, notifyError, path, t]);

	const handleToggle = useCallback(
		async (id: number, enabled: boolean) => {
			try {
				await updateMutation.mutateAsync({ id, enabled });
				enqueueSnackbar(t('SETTINGS_STORAGE_ROOTS_SAVED'), { variant: 'success' });
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
				enqueueSnackbar(t('SETTINGS_STORAGE_ROOTS_SAVED'), { variant: 'success' });
			} catch (error) {
				notifyError(error);
			}
		},
		[deleteMutation, enqueueSnackbar, notifyError, t]
	);

	const roots = rootsQuery.data ?? [];

	return {
		t,
		roots,
		primaryRootId: roots[0]?.id,
		isLoading: rootsQuery.isLoading,
		isSaving: createMutation.isPending || updateMutation.isPending || deleteMutation.isPending,
		hasError: rootsQuery.isError,
		path,
		label,
		setPath,
		setLabel,
		handleAdd,
		handleToggle,
		handleDelete,
	};
};

export default useStorageRootsSettings;
