import type { CaptureStatus } from '@/types/captures';

export const captureStatuses: CaptureStatus[] = ['uploaded', 'promoting', 'promoted', 'failed'];

export type CapturesFilters = {
    name: string;
    status: CaptureStatus | '';
    platform: string;
};

export const emptyCapturesFilters: CapturesFilters = { name: '', status: '', platform: '' };

const nameParamName = 'q';
const statusParamName = 'status';
const platformParamName = 'platform';

export const parseCapturesFilters = (searchParams: URLSearchParams): CapturesFilters => ({
    name: searchParams.get(nameParamName) ?? '',
    status: captureStatuses.find((status) => status === searchParams.get(statusParamName)) ?? '',
    platform: searchParams.get(platformParamName) ?? '',
});

export const writeCapturesFilters = (
    currentParams: URLSearchParams,
    filters: CapturesFilters
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    nextParams.delete(nameParamName);
    nextParams.delete(statusParamName);
    nextParams.delete(platformParamName);
    if (filters.name) nextParams.set(nameParamName, filters.name);
    if (filters.status) nextParams.set(statusParamName, filters.status);
    if (filters.platform) nextParams.set(platformParamName, filters.platform);
    return nextParams;
};

export const hasActiveCapturesFilters = (filters: CapturesFilters): boolean =>
    filters.name !== '' || filters.status !== '' || filters.platform !== '';
