import { useMemo, useReducer } from 'react';
import {
	ActivityDiaryContextProvider,
	ActivityDiaryFormData,
	ActivityDiarySummary,
	ActivityDiaryType,
} from './ActivityDiaryContext';
import { useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

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
	const [formData, setFormData] = useReducer(reducerFormData, initialFormState);
	const { status, data: summaryData } = useQuery({
		queryKey: ['activity-diary-summary'],
		queryFn: async (): Promise<ActivityDiarySummary> => {
			const response = await apiBase.get<ActivityDiarySummary>('/activity/summary');
			return response.data;
		},
	});

	const contextValue: ActivityDiaryType = useMemo(
		() => ({
			form: formData,
			setForm: setFormData,
			loading: true,

			data: {
				entries: [],
				summary: summaryData,
			},
		}),
		[]
	);

	return <ActivityDiaryContextProvider value={contextValue}>{children}</ActivityDiaryContextProvider>;
};

export default ActivityDiaryProvider;
