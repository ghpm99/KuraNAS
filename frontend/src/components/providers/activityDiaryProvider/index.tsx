import { useMutation, useQuery } from '@tanstack/react-query';
import { ChangeEvent, FormEvent, useCallback, useEffect, useMemo, useReducer, useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
	createActivityDiaryEntry,
	duplicateActivityDiaryEntry,
	getActivityDiaryEntries,
	getActivityDiarySummary,
} from '@/service/activityDiary';
import {
	ActivityDiaryContextProvider,
	ActivityDiaryData,
	ActivityDiaryFormData,
	ActivityDiaryType,
	messageType,
} from './ActivityDiaryContext';
import { useSnackbar } from 'notistack';

const initialFormState: ActivityDiaryFormData = {
	name: '',
	description: '',
};

export type FormAction =
	| { type: 'SET_NAME'; payload: string }
	| { type: 'SET_DESCRIPTION'; payload: string }
	| { type: 'RESET' };

const reducerFormData = (state: ActivityDiaryFormData, action: FormAction): ActivityDiaryFormData => {
	switch (action.type) {
		case 'SET_NAME':
			return { ...state, name: action.payload };
		case 'SET_DESCRIPTION':
			return { ...state, description: action.payload };
		case 'RESET':
			return initialFormState;
		default:
			throw new Error(`Unknown action: ${action}`);
	}
};

const ActivityDiaryProvider = ({ children }: { children: React.ReactNode }) => {
	const [currentTime, setCurrentTime] = useState(new Date());
	const [formData, setFormData] = useReducer(reducerFormData, initialFormState);
	const [message, setMessage] = useState<{ text: string; type: messageType } | undefined>(undefined);
	const { enqueueSnackbar } = useSnackbar();
	const { t } = useI18n();

	const {
		data: summaryData,
		error: summaryError,
		refetch: refetchSummary,
	} = useQuery({
		queryKey: ['activity-diary-summary'],
		queryFn: getActivityDiarySummary,
	});

	const {
		data: diaryData,
		error,
		refetch: refetchList,
	} = useQuery({
		queryKey: ['activity-diary-list'],
		queryFn: getActivityDiaryEntries,
	});

	const createDiaryMutation = useMutation({
		mutationFn: (form: ActivityDiaryFormData): Promise<ActivityDiaryData> => createActivityDiaryEntry(form),
		onSuccess: () => {
			enqueueSnackbar(t('ACTIVITY_CREATE_SUCCESS'), { variant: 'success' });
			refetchList();
			refetchSummary();
		},
		onError: () => {
			enqueueSnackbar(t('ACTIVITY_CREATE_ERROR'), { variant: 'error' });
		},
	});

	const duplicateDiaryMutation = useMutation({
		mutationFn: (diaryId: number): Promise<ActivityDiaryData> => duplicateActivityDiaryEntry(diaryId),
		onSuccess: () => {
			enqueueSnackbar(t('ACTIVITY_DUPLICATE_SUCCESS'), { variant: 'success' });
			refetchList();
			refetchSummary();
		},
		onError: () => {
			enqueueSnackbar(t('ACTIVITY_DUPLICATE_ERROR'), { variant: 'error' });
		},
	});

	useEffect(() => {
		const timer = setInterval(() => {
			setCurrentTime(new Date());
		}, 1000);

		return () => clearInterval(timer);
	}, []);

	const handleSubmit = useCallback(
		(e: FormEvent) => {
			e.preventDefault();
			if (formData.name.length > 50) {
				setMessage({ text: t('ACTIVITY_NAME_MAX_ERROR'), type: 'error' });
				return;
			}
			if (formData.name.length < 3) {
				setMessage({ text: t('ACTIVITY_NAME_MIN_ERROR'), type: 'error' });
				return;
			}
			if (!/^[a-zA-Z0-9 ]+$/.test(formData.name)) {
				setMessage({ text: t('ACTIVITY_NAME_INVALID_ERROR'), type: 'error' });
				return;
			}
			if (formData.name.trim() === '') {
				setMessage({ text: t('ACTIVITY_NAME_EMPTY_ERROR'), type: 'error' });
				return;
			}
			if (message) {
				setMessage(undefined);
			}
			const name = formData.name.trim().toLowerCase();

			createDiaryMutation.mutate({
				name: name,
				description: formData.description,
			});

			setFormData({ type: 'RESET' });
		},
		[formData, createDiaryMutation, message, t]
	);

	const handleNameChange = ({ target }: ChangeEvent<HTMLInputElement>) => {
		const { value } = target;

		setFormData({ type: 'SET_NAME', payload: value });
	};

	const handleDescriptionChange = ({ target }: ChangeEvent<HTMLTextAreaElement>) => {
		setFormData({ type: 'SET_DESCRIPTION', payload: target.value });
	};

	const getCurrentDuration = useCallback(
		(dateString: string): number => {
			const date = new Date(dateString);
			return Math.floor((currentTime.getTime() - date.getTime()) / 1000);
		},
		[currentTime]
	);

	const copyActivity = useCallback(
		(activity: ActivityDiaryData) => {
			duplicateDiaryMutation.mutate(activity.id);
		},
		[duplicateDiaryMutation]
	);

	const contextValue: ActivityDiaryType = useMemo(
		() => ({
				form: formData,
				handleSubmit,
				handleNameChange,
				handleDescriptionChange,
				loading: createDiaryMutation.isPending || duplicateDiaryMutation.isPending,
				message: message,
			data: {
				entries: diaryData || { items: [], pagination: { page: 1, has_next: false, has_prev: false, page_size: 10 } },
				summary: summaryData,
			},
			getCurrentDuration,
			error: error?.message || summaryError?.message,
			currentTime,
			copyActivity,
		}),
		[
			error?.message,
			summaryError?.message,
			formData,
			getCurrentDuration,
			handleSubmit,
				message,
				diaryData,
				summaryData,
				currentTime,
				copyActivity,
				createDiaryMutation.isPending,
				duplicateDiaryMutation.isPending,
			]
		);

	return <ActivityDiaryContextProvider value={contextValue}>{children}</ActivityDiaryContextProvider>;
};

export default ActivityDiaryProvider;
