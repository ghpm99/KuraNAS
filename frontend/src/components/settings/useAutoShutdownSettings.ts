import useI18n from '@/components/i18n/provider/i18nContext';
import {
	getAutoShutdownSettings,
	getSuggestedShutdownTime,
	updateAutoShutdownSettings,
} from '@/service/autoShutdown';
import type { AutoShutdownSettings, SuggestedShutdownTime } from '@/types/autoShutdown';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const defaultSettings: AutoShutdownSettings = {
	enabled: false,
	time: '03:00',
	grace_period_seconds: 60,
};

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

const useAutoShutdownSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	// draft holds unsaved edits; while null the form mirrors the server state.
	const [draft, setDraft] = useState<AutoShutdownSettings | null>(null);
	const [suggestion, setSuggestion] = useState<SuggestedShutdownTime | null>(null);

	const settingsQuery = useQuery({
		queryKey: ['auto-shutdown-settings'],
		queryFn: getAutoShutdownSettings,
		retry: false,
	});

	const form = draft ?? settingsQuery.data ?? defaultSettings;

	const setField = useCallback(
		<K extends keyof AutoShutdownSettings>(field: K, value: AutoShutdownSettings[K]) => {
			setDraft({ ...form, [field]: value });
		},
		[form]
	);

	const saveMutation = useMutation({
		mutationFn: updateAutoShutdownSettings,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ['auto-shutdown-settings'] });
		},
	});

	const suggestMutation = useMutation({
		mutationFn: getSuggestedShutdownTime,
	});

	const handleSave = useCallback(async () => {
		try {
			await saveMutation.mutateAsync(form);
			setDraft(null);
			enqueueSnackbar(t('SETTINGS_AUTO_SHUTDOWN_SAVE_SUCCESS'), { variant: 'success' });
		} catch (error) {
			// Backend errors arrive already translated — show them verbatim.
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('SETTINGS_AUTO_SHUTDOWN_SAVE_ERROR'), {
				variant: 'error',
			});
		}
	}, [enqueueSnackbar, form, saveMutation, t]);

	const handleSuggest = useCallback(async () => {
		try {
			const result = await suggestMutation.mutateAsync();
			setSuggestion(result);
			if (result.available) {
				setField('time', result.time);
			}
		} catch (error) {
			const backendMessage = extractBackendError(error);
			enqueueSnackbar(backendMessage ?? t('SETTINGS_AUTO_SHUTDOWN_LOAD_ERROR'), {
				variant: 'error',
			});
		}
	}, [enqueueSnackbar, setField, suggestMutation, t]);

	return {
		t,
		form,
		suggestion,
		isLoading: settingsQuery.isLoading,
		isSaving: saveMutation.isPending,
		isSuggesting: suggestMutation.isPending,
		hasError: settingsQuery.isError,
		hasUnsavedChanges: draft !== null,
		setField,
		handleSave,
		handleSuggest,
	};
};

export default useAutoShutdownSettings;
