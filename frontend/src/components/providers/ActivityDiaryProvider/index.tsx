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
	const { data: summaryData, error } = useQuery({
		queryKey: ['activity-diary-summary'],
		queryFn: async (): Promise<ActivityDiarySummary> => {
			const response = await apiBase.get<ActivityDiarySummary>('/diary/summary');
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
		},
		onError: (error) => {
			setMessage({ text: 'Erro ao adicionar atividade.', type: 'error' });
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
			if (formData.name.trim()) {
				createDiaryMutation.mutate({
					name: formData.name,
					description: formData.description,
				});

				setFormData({ type: 'RESET' });
			}
		},
		[formData, createDiaryMutation]
	);

	const handleNameChange = ({ target }: ChangeEvent<HTMLInputElement>) => {
		setFormData({ type: 'SET_NAME', payload: target.value });
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

	const contextValue: ActivityDiaryType = useMemo(
		() => ({
			form: formData,
			handleSubmit,
			handleNameChange,
			handleDescriptionChange,
			loading: true,
			message: message,
			data: {
				entries: [],
				summary: summaryData,
			},
			getCurrentDuration,
			error: error?.message,
		}),
		[error?.message, formData, getCurrentDuration, handleSubmit, message, summaryData]
	);

	return <ActivityDiaryContextProvider value={contextValue}>{children}</ActivityDiaryContextProvider>;
};

export default ActivityDiaryProvider;
