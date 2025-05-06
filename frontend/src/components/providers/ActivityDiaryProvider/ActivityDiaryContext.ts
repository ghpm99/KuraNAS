import { createContext, useContext } from 'react';

export type ActivityDiaryFormData = {
	name: string;
	description?: string;
};

export type ActivityDiaryData = {
	id: number;
	name: string;
	description: string;
	start_time: string; // ISO 8601
	end_time: string | null;
	duration_seconds: number | null;
	duration_formatted: string | null;
	in_progress?: boolean;
};

export type ActivityDiarySummary = {
	date: string; // "YYYY-MM-DD"
	total_activities: number;
	total_time_spent_seconds: number;
	total_time_spent_formatted: string;
	longest_activity?: {
		name: string;
		duration_seconds: number;
		duration_formatted: string;
	};
};

export type ActivityDiaryResponse = {
	summary: ActivityDiarySummary | undefined;
	entries: ActivityDiaryData[];
};

export type ActivityDiaryType = {
	form: ActivityDiaryFormData;
	loading: boolean;
	error?: string;
	data: ActivityDiaryResponse | null;
};

const ActivityDiaryContext = createContext<ActivityDiaryType | undefined>(undefined);

export const ActivityDiaryContextProvider = ActivityDiaryContext.Provider;

export const useActivityDiary = () => {
	const context = useContext(ActivityDiaryContext);
	if (!context) {
		throw new Error('useActivityDiary must be used within a ActivityDiaryProvider');
	}

	return context;
};
