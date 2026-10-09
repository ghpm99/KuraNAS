import {
    fileSearchKinds,
    type FileSearchKind,
    type FileSearchOrder,
    type FileSearchRefinements,
    type FileSearchSort,
    type FileSearchTier,
} from '@/types/fileSearch';

export const fileSearchSizePresets = ['small', 'medium', 'large', 'huge'] as const;
export const fileSearchSorts: FileSearchSort[] = ['relevance', 'name', 'size', 'modified'];
export const fileSearchTiers: FileSearchTier[] = ['hot', 'cold'];

export type FileSearchSizePreset = (typeof fileSearchSizePresets)[number];

export type FileSearchFilters = {
    kinds: FileSearchKind[];
    modifiedFrom: string;
    modifiedTo: string;
    sizePreset: FileSearchSizePreset | '';
    tier: FileSearchTier | '';
    onlyStarred: boolean;
    sort: FileSearchSort;
    order: FileSearchOrder | '';
};

export const emptyFileSearchFilters: FileSearchFilters = {
    kinds: [],
    modifiedFrom: '',
    modifiedTo: '',
    sizePreset: '',
    tier: '',
    onlyStarred: false,
    sort: 'relevance',
    order: '',
};

const bytesPerMegabyte = 1024 * 1024;
const bytesPerGigabyte = 1024 * bytesPerMegabyte;

const sizeBoundsByPreset: Record<FileSearchSizePreset, { minSize?: number; maxSize?: number }> = {
    small: { maxSize: bytesPerMegabyte - 1 },
    medium: { minSize: bytesPerMegabyte, maxSize: 100 * bytesPerMegabyte - 1 },
    large: { minSize: 100 * bytesPerMegabyte, maxSize: bytesPerGigabyte - 1 },
    huge: { minSize: bytesPerGigabyte },
};

const isoDatePattern = /^\d{4}-\d{2}-\d{2}$/;

const filterParamNames = [
    'kind',
    'modified_from',
    'modified_to',
    'size',
    'tier',
    'starred',
    'sort',
    'order',
];

const pickKnown = <Allowed extends string>(
    rawValue: string | null,
    allowedValues: readonly Allowed[]
): Allowed | undefined => allowedValues.find((allowedValue) => allowedValue === rawValue);

const pickIsoDate = (rawValue: string | null): string =>
    rawValue !== null && isoDatePattern.test(rawValue) ? rawValue : '';

export const parseFileSearchFilters = (searchParams: URLSearchParams): FileSearchFilters => ({
    kinds: fileSearchKinds.filter((kind) => searchParams.getAll('kind').includes(kind)),
    modifiedFrom: pickIsoDate(searchParams.get('modified_from')),
    modifiedTo: pickIsoDate(searchParams.get('modified_to')),
    sizePreset: pickKnown(searchParams.get('size'), fileSearchSizePresets) ?? '',
    tier: pickKnown(searchParams.get('tier'), fileSearchTiers) ?? '',
    onlyStarred: searchParams.get('starred') === 'true',
    sort: pickKnown(searchParams.get('sort'), fileSearchSorts) ?? 'relevance',
    order: pickKnown(searchParams.get('order'), ['asc', 'desc'] as const) ?? '',
});

export const writeFileSearchFilters = (
    currentParams: URLSearchParams,
    filters: FileSearchFilters
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    filterParamNames.forEach((paramName) => nextParams.delete(paramName));
    filters.kinds.forEach((kind) => nextParams.append('kind', kind));
    if (filters.modifiedFrom) nextParams.set('modified_from', filters.modifiedFrom);
    if (filters.modifiedTo) nextParams.set('modified_to', filters.modifiedTo);
    if (filters.sizePreset) nextParams.set('size', filters.sizePreset);
    if (filters.tier) nextParams.set('tier', filters.tier);
    if (filters.onlyStarred) nextParams.set('starred', 'true');
    if (filters.sort !== 'relevance') nextParams.set('sort', filters.sort);
    if (filters.order) nextParams.set('order', filters.order);
    return nextParams;
};

export const countActiveFileSearchFilters = (filters: FileSearchFilters): number =>
    [
        filters.kinds.length > 0,
        filters.modifiedFrom !== '' || filters.modifiedTo !== '',
        filters.sizePreset !== '',
        filters.tier !== '',
        filters.onlyStarred,
        filters.sort !== 'relevance',
    ].filter(Boolean).length;

export const toFileSearchRefinements = (filters: FileSearchFilters): FileSearchRefinements => ({
    kinds: filters.kinds.length > 0 ? filters.kinds : undefined,
    modifiedFrom: filters.modifiedFrom || undefined,
    modifiedTo: filters.modifiedTo || undefined,
    ...(filters.sizePreset ? sizeBoundsByPreset[filters.sizePreset] : {}),
    tier: filters.tier || undefined,
    starred: filters.onlyStarred || undefined,
    sort: filters.sort === 'relevance' ? undefined : filters.sort,
    order: filters.sort === 'relevance' ? undefined : filters.order || undefined,
});
