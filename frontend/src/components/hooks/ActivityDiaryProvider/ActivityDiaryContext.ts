import { Optional } from '@/types/optional';
import { Pagination } from '@/types/pagination';
import { ChangeEvent, createContext, FormEvent, useContext } from 'react';

export type ActivityDiaryFormData = {
	name: string;
	description?: string;
};

export type ActivityDiaryData = {
	id: number;
	name: string;
	description: string;
	start_time: string;
	end_time: Optional<string>;
	duration: number | undefined;
	duration_formatted: string | null;
	in_progress?: boolean;
};

export type ActivityDiarySummary = {
	date: string;
	total_activities: number;
	total_time_spent_seconds: number;
	longest_activity?: {
		name: string;
		duration_seconds: number;
		duration_formatted: string;
	};
};

export type ActivityDiaryResponse = {
	summary: ActivityDiarySummary | undefined;
	entries: Pagination<ActivityDiaryData>;
};

export type messageType = 'success' | 'error' | 'info';

export type ActivityDiaryType = {
	form: ActivityDiaryFormData;
	handleSubmit: (e: FormEvent) => void;
	handleNameChange: (e: ChangeEvent<HTMLInputElement>) => void;
	handleDescriptionChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
	loading: boolean;
	message?: { text: string; type: messageType };
	error?: string;
	data: ActivityDiaryResponse | null;
	getCurrentDuration: (dateString: string) => number;
	currentTime: Date;
	copyActivity: (activity: ActivityDiaryData) => void;
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
