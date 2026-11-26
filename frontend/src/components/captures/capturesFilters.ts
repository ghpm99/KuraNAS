import type { CaptureStatus } from '@/types/captures';

export const captureStatuses: CaptureStatus[] = ['uploaded', 'promoting', 'promoted', 'failed'];

export type CapturesFilters = {
    name: string;
    status: CaptureStatus | '';
};

export const emptyCapturesFilters: CapturesFilters = { name: '', status: '' };

const nameParamName = 'q';
const statusParamName = 'status';

export const parseCapturesFilters = (searchParams: URLSearchParams): CapturesFilters => ({
    name: searchParams.get(nameParamName) ?? '',
    status: captureStatuses.find((status) => status === searchParams.get(statusParamName)) ?? '',
});

export const writeCapturesFilters = (
    currentParams: URLSearchParams,
    filters: CapturesFilters
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    nextParams.delete(nameParamName);
    nextParams.delete(statusParamName);
    if (filters.name) nextParams.set(nameParamName, filters.name);
    if (filters.status) nextParams.set(statusParamName, filters.status);
    return nextParams;
};

export const hasActiveCapturesFilters = (filters: CapturesFilters): boolean =>
    filters.name !== '' || filters.status !== '';
