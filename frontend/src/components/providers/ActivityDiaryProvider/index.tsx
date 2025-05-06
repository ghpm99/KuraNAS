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

const reducerFormData = (
	state: ActivityDiaryFormData,
	action: { type: string; payload?: any }
): ActivityDiaryFormData => {
	switch (action.type) {
		case 'SET_NAME':
			return { ...state, name: action.payload };
		case 'SET_DESCRIPTION':
			return { ...state, description: action.payload };
		case 'RESET':
			return initialFormState;
		default:
			throw new Error(`Unknown action: ${action.type}`);
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
