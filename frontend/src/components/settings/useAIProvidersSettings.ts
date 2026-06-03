import useI18n from '@/components/i18n/provider/i18nContext';
import { getAIProviders, updateAIProvider } from '@/service/aiProviders';
import { getJob } from '@/service/jobs';
import { deleteOllamaModel, getOllamaStatus, pullOllamaModel } from '@/service/ollama';
import type { AIProviderDto, AIProviderName, UpdateAIProviderRequest } from '@/types/aiProviders';
import { isJobFinished } from '@/types/jobs';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useEffect, useMemo, useState } from 'react';

type ProviderEdits = Partial<Record<AIProviderName, Partial<AIProviderDto>>>;

const toUpdateRequest = (provider: AIProviderDto): UpdateAIProviderRequest => ({
	enabled: provider.enabled,
	model: provider.model,
	base_url: provider.base_url,
	priority: provider.priority,
	params: provider.params,
});

const useAIProvidersSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();

	const providersQuery = useQuery({
		queryKey: ['ai-providers'],
		queryFn: getAIProviders,
		retry: false,
	});

	const [pullJobId, setPullJobId] = useState<number | null>(null);

	const ollamaQuery = useQuery({
		queryKey: ['ollama-status'],
		queryFn: getOllamaStatus,
		retry: false,
		refetchInterval: pullJobId ? 5000 : false,
	});

	const jobQuery = useQuery({
		queryKey: ['job', pullJobId],
		queryFn: () => getJob(pullJobId as number),
		enabled: pullJobId !== null,
		refetchInterval: pullJobId ? 1500 : false,
	});

	const updateMutation = useMutation({
		mutationFn: ({ name, request }: { name: AIProviderName; request: UpdateAIProviderRequest }) =>
			updateAIProvider(name, request),
		onSuccess: (updated) => {
			queryClient.setQueryData<AIProviderDto[]>(['ai-providers'], (current = []) =>
				current.map((provider) => (provider.name === updated.name ? updated : provider))
			);
		},
	});

	const pullMutation = useMutation({ mutationFn: pullOllamaModel });
	const deleteMutation = useMutation({ mutationFn: deleteOllamaModel });

	const [edits, setEdits] = useState<ProviderEdits>({});
	const [pullModelName, setPullModelName] = useState('');

	const providers = useMemo(() => {
		const resolved = providersQuery.data ?? [];
		return resolved.map((provider) => ({
			...provider,
			...edits[provider.name],
		}));
	}, [providersQuery.data, edits]);

	const setField = useCallback(
		<K extends keyof AIProviderDto>(name: AIProviderName, field: K, value: AIProviderDto[K]) => {
			setEdits((current) => ({
				...current,
				[name]: { ...current[name], [field]: value },
			}));
		},
		[]
	);

	const persist = useCallback(
		async (provider: AIProviderDto, successKey: string) => {
			try {
				await updateMutation.mutateAsync({
					name: provider.name,
					request: toUpdateRequest(provider),
				});
				setEdits((current) => ({ ...current, [provider.name]: undefined }));
				enqueueSnackbar(t(successKey), { variant: 'success' });
			} catch {
				enqueueSnackbar(t('AI_PROVIDERS_SAVE_ERROR'), { variant: 'error' });
			}
		},
		[enqueueSnackbar, t, updateMutation]
	);

	const toggleEnabled = useCallback(
		async (name: AIProviderName, enabled: boolean) => {
			const provider = providers.find((item) => item.name === name);
			if (!provider) return;
			await persist({ ...provider, enabled }, enabled ? 'AI_PROVIDERS_ENABLED' : 'AI_PROVIDERS_DISABLED');
		},
		[persist, providers]
	);

	const saveProvider = useCallback(
		async (name: AIProviderName) => {
			const provider = providers.find((item) => item.name === name);
			if (!provider) return;
			await persist(provider, 'AI_PROVIDERS_SAVED');
		},
		[persist, providers]
	);

	const handlePull = useCallback(async () => {
		const model = pullModelName.trim();
		if (!model) return;
		try {
			const response = await pullMutation.mutateAsync(model);
			setPullJobId(response.job_id);
			enqueueSnackbar(t('AI_OLLAMA_PULL_STARTED'), { variant: 'info' });
		} catch {
			enqueueSnackbar(t('AI_OLLAMA_PULL_ERROR'), { variant: 'error' });
		}
	}, [enqueueSnackbar, pullModelName, pullMutation, t]);

	const handleDeleteModel = useCallback(
		async (name: string) => {
			try {
				await deleteMutation.mutateAsync(name);
				enqueueSnackbar(t('AI_OLLAMA_MODEL_DELETED'), { variant: 'success' });
				await queryClient.invalidateQueries({ queryKey: ['ollama-status'] });
			} catch {
				enqueueSnackbar(t('AI_OLLAMA_MODEL_DELETE_ERROR'), { variant: 'error' });
			}
		},
		[deleteMutation, enqueueSnackbar, queryClient, t]
	);

	// When the active pull job finishes, stop polling, refresh the models list
	// and notify the user.
	useEffect(() => {
		const job = jobQuery.data;
		if (!job || pullJobId === null) return;
		if (!isJobFinished(job.status)) return;

		setPullJobId(null);
		setPullModelName('');
		void queryClient.invalidateQueries({ queryKey: ['ollama-status'] });
		if (job.status === 'completed') {
			enqueueSnackbar(t('AI_OLLAMA_PULL_COMPLETED'), { variant: 'success' });
		} else {
			enqueueSnackbar(t('AI_OLLAMA_PULL_ERROR'), { variant: 'error' });
		}
	}, [jobQuery.data, pullJobId, queryClient, enqueueSnackbar, t]);

	return {
		t,
		providers,
		isLoading: providersQuery.isLoading,
		hasError: providersQuery.isError,
		isSaving: updateMutation.isPending,
		ollamaStatus: ollamaQuery.data,
		ollamaLoading: ollamaQuery.isLoading,
		toggleEnabled,
		setField,
		saveProvider,
		pullModelName,
		setPullModelName,
		handlePull,
		isPulling: pullMutation.isPending || pullJobId !== null,
		pullProgress: jobQuery.data?.progress.progress ?? 0,
		handleDeleteModel,
		isDeleting: deleteMutation.isPending,
	};
};

export default useAIProvidersSettings;
