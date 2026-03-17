import {
    ActivityDiaryData,
    ActivityDiaryFormData,
    ActivityDiarySummary,
} from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import { Pagination } from '@/types/pagination';
import { apiBase } from '.';

export const getActivityDiarySummary = async (): Promise<ActivityDiarySummary> => {
    const response = await apiBase.get<ActivityDiarySummary>('/diary/summary');
    return response.data;
};

export const getActivityDiaryEntries = async (): Promise<Pagination<ActivityDiaryData>> => {
    const response = await apiBase.get<Pagination<ActivityDiaryData>>('/diary/');
    return response.data;
};

export const createActivityDiaryEntry = async (
    form: ActivityDiaryFormData
): Promise<ActivityDiaryData> => {
    const response = await apiBase.post<ActivityDiaryData>('/diary/', {
        name: form.name,
        description: form.description,
    });
    return response.data;
};

export const duplicateActivityDiaryEntry = async (diaryId: number): Promise<ActivityDiaryData> => {
    const response = await apiBase.post<ActivityDiaryData>('/diary/copy', {
        ID: diaryId,
    });
    return response.data;
};
