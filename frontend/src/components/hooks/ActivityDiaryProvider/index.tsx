import { apiBase } from '@/service';
import { useMutation, useQuery } from '@tanstack/react-query';
import { ChangeEvent, FormEvent, useCallback, useEffect, useMemo, useReducer, useState } from 'react';
import {
	ActivityDiaryContextProvider,
	ActivityDiaryData,
	ActivityDiaryFormData,
	ActivityDiarySummary,
	ActivityDiaryType,
	messageType,
} from './ActivityDiaryContext';
import { Pagination } from '@/types/pagination';

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

	const {
		data: summaryData,
		error: summaryError,
		refetch: refetchSummary,
	} = useQuery({
		queryKey: ['activity-diary-summary'],
		queryFn: async (): Promise<ActivityDiarySummary> => {
			const response = await apiBase.get<ActivityDiarySummary>('/diary/summary');
			return response.data;
		},
	});

	const {
		data: diaryData,
		error,
		refetch: refetchList,
	} = useQuery({
		queryKey: ['activity-diary-list'],
		queryFn: async (): Promise<Pagination<ActivityDiaryData>> => {
			const response = await apiBase.get<Pagination<ActivityDiaryData>>('/diary/');
			return response.data;
		},
	});

	const createDiaryMutation = useMutation({
		mutationFn: async (form: ActivityDiaryFormData): Promise<ActivityDiaryData> => {
			const response = await apiBase.post<ActivityDiaryData>('/diary/', {
				name: form.name,
				description: form.description,
			});
			return response.data;
		},
		onSuccess: (data) => {
			setMessage({ text: 'Atividade adicionada com sucesso!', type: 'success' });
			console.log('Diário criado:', data);
			refetchList();
			refetchSummary();
		},
		onError: (error) => {
			setMessage({ text: 'Erro ao adicionar atividade.', type: 'error' });
			console.error('Erro ao criar diário:', error);
		},
	});

	const duplicateDiaryMutation = useMutation({
		mutationFn: async (diaryId: number): Promise<ActivityDiaryData> => {
			const response = await apiBase.post<ActivityDiaryData>('/diary/copy', {
				ID: diaryId,
			});
			return response.data;
		},
		onSuccess: (data) => {
			setMessage({ text: 'Atividade duplicada com sucesso!', type: 'success' });
			console.log('Diário criado:', data);
			refetchList();
			refetchSummary();
		},
		onError: (error) => {
			setMessage({ text: 'Erro ao duplicar atividade.', type: 'error' });
			console.error('Erro ao criar diário:', error);
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
				setMessage({ text: 'O nome deve ter no máximo 50 caracteres.', type: 'error' });
				return;
			}
			if (formData.name.length < 3) {
				setMessage({ text: 'O nome deve ter no mínimo 3 caracteres.', type: 'error' });
				return;
			}
			if (!/^[a-zA-Z0-9 ]+$/.test(formData.name)) {
				setMessage({ text: 'O nome só pode conter letras, números e espaços.', type: 'error' });
				return;
			}
			if (formData.name.trim() === '') {
				setMessage({ text: 'O nome não pode ser vazio.', type: 'error' });
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
		[formData, createDiaryMutation, message]
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

	const copyActivity = (activity: ActivityDiaryData) => {
		duplicateDiaryMutation.mutate(activity.id);
	};

	const contextValue: ActivityDiaryType = useMemo(
		() => ({
			form: formData,
			handleSubmit,
			handleNameChange,
			handleDescriptionChange,
			loading: true,
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
		]
	);

	return <ActivityDiaryContextProvider value={contextValue}>{children}</ActivityDiaryContextProvider>;
};

export default ActivityDiaryProvider;
