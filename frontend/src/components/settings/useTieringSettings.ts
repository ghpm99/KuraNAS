import useI18n from '@/components/i18n/provider/i18nContext';
import {
	getTieringSettings,
	getTieringStatus,
	getTieringUsage,
	updateTieringSettings,
} from '@/service/tiering';
import type { TieringSettings } from '@/types/tiering';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const defaultSettings: TieringSettings = {
	enabled: false,
	cold_dir_path: '',
	min_age_days: 90,
	min_size_bytes: 1024 * 1024,
	interval_hours: 24,
};

const statusKeyByValue: Record<string, string> = {
	queued: 'SETTINGS_TIERING_STATUS_QUEUED',
	running: 'SETTINGS_TIERING_STATUS_RUNNING',
	completed: 'SETTINGS_TIERING_STATUS_COMPLETED',
	failed: 'SETTINGS_TIERING_STATUS_FAILED',
	partial_fail: 'SETTINGS_TIERING_STATUS_PARTIAL_FAIL',
	canceled: 'SETTINGS_TIERING_STATUS_CANCELED',
};

export const tieringStatusKey = (status: string): string =>
	statusKeyByValue[status] ?? 'SETTINGS_TIERING_STATUS_QUEUED';

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

const useTieringSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	// draft holds unsaved edits; while null the form mirrors the server state.
	const [draft, setDraft] = useState<TieringSettings | null>(null);

	const settingsQuery = useQuery({
		queryKey: ['tiering-settings'],
		queryFn: getTieringSettings,
		retry: false,
	});
	const statusQuery = useQuery({
		queryKey: ['tiering-status'],
		queryFn: getTieringStatus,
		retry: false,
	});
	const usageQuery = useQuery({
		queryKey: ['tiering-usage'],
		queryFn: getTieringUsage,
		retry: false,
	});

	const form = draft ?? settingsQuery.data ?? defaultSettings;

	const setField = useCallback(
		<K extends keyof TieringSettings>(field: K, value: TieringSettings[K]) => {
			setDraft({ ...form, [field]: value });
		},
		[form]
	);

	const saveMutation = useMutation({
		mutationFn: updateTieringSettings,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ['tiering-settings'] });
			void queryClient.invalidateQueries({ queryKey: ['tiering-status'] });
			void queryClient.invalidateQueries({ queryKey: ['tiering-usage'] });
		},
	});

	const handleSave = useCallback(async () => {
		try {
			await saveMutation.mutateAsync(form);
			setDraft(null);
			enqueueSnackbar(t('SETTINGS_TIERING_SAVED'), { variant: 'success' });
		} catch (error) {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('SETTINGS_TIERING_SAVE_ERROR'), {
				variant: 'error',
			});
		}
	}, [enqueueSnackbar, form, saveMutation, t]);

	return {
		t,
		form,
		status: statusQuery.data,
		usage: usageQuery.data,
		isLoading: settingsQuery.isLoading,
		isSaving: saveMutation.isPending,
		hasError: settingsQuery.isError,
		hasUnsavedChanges: draft !== null,
		setField,
		handleSave,
	};
};

export default useTieringSettings;
