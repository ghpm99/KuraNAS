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

type ListActivityDiaryEntriesParams = {
    page?: number;
    pageSize?: number;
};

export const getActivityDiaryEntries = async ({
    page = 1,
    pageSize = 20,
}: ListActivityDiaryEntriesParams = {}): Promise<Pagination<ActivityDiaryData>> => {
    const response = await apiBase.get<Pagination<ActivityDiaryData>>('/diary/', {
        params: { page, page_size: pageSize },
    });
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
        id: diaryId,
    });
    return response.data;
};
