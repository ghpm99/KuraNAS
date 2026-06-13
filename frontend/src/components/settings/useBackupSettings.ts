import useI18n from '@/components/i18n/provider/i18nContext';
import {
	getBackupPending,
	getBackupSettings,
	getBackupStatus,
	updateBackupSettings,
} from '@/service/backup';
import type { BackupSettings } from '@/types/backup';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const defaultSettings: BackupSettings = {
	enabled: false,
	destination_path: '',
	retention_days: 30,
	interval_hours: 24,
};

const statusKeyByValue: Record<string, string> = {
	queued: 'SETTINGS_BACKUP_STATUS_QUEUED',
	running: 'SETTINGS_BACKUP_STATUS_RUNNING',
	completed: 'SETTINGS_BACKUP_STATUS_COMPLETED',
	failed: 'SETTINGS_BACKUP_STATUS_FAILED',
	partial_fail: 'SETTINGS_BACKUP_STATUS_PARTIAL_FAIL',
	canceled: 'SETTINGS_BACKUP_STATUS_CANCELED',
};

export const backupStatusKey = (status: string): string =>
	statusKeyByValue[status] ?? 'SETTINGS_BACKUP_STATUS_QUEUED';

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

const useBackupSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	// draft holds unsaved edits; while null the form mirrors the server state.
	const [draft, setDraft] = useState<BackupSettings | null>(null);

	const settingsQuery = useQuery({
		queryKey: ['backup-settings'],
		queryFn: getBackupSettings,
		retry: false,
	});
	const statusQuery = useQuery({
		queryKey: ['backup-status'],
		queryFn: getBackupStatus,
		retry: false,
	});
	const pendingQuery = useQuery({
		queryKey: ['backup-pending'],
		queryFn: getBackupPending,
		retry: false,
	});

	const form = draft ?? settingsQuery.data ?? defaultSettings;

	const setField = useCallback(
		<K extends keyof BackupSettings>(field: K, value: BackupSettings[K]) => {
			setDraft({ ...form, [field]: value });
		},
		[form]
	);

	const saveMutation = useMutation({
		mutationFn: updateBackupSettings,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ['backup-settings'] });
			void queryClient.invalidateQueries({ queryKey: ['backup-status'] });
			void queryClient.invalidateQueries({ queryKey: ['backup-pending'] });
		},
	});

	const handleSave = useCallback(async () => {
		try {
			await saveMutation.mutateAsync(form);
			setDraft(null);
			enqueueSnackbar(t('SETTINGS_BACKUP_SAVED'), { variant: 'success' });
		} catch (error) {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('SETTINGS_BACKUP_SAVE_ERROR'), {
				variant: 'error',
			});
		}
	}, [enqueueSnackbar, form, saveMutation, t]);

	return {
		t,
		form,
		status: statusQuery.data,
		pendingFiles: pendingQuery.data?.pending_files,
		isLoading: settingsQuery.isLoading,
		isSaving: saveMutation.isPending,
		hasError: settingsQuery.isError,
		hasUnsavedChanges: draft !== null,
		setField,
		handleSave,
	};
};

export default useBackupSettings;
