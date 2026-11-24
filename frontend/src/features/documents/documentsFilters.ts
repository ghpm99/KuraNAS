import type {
    DocumentFilters,
    DocumentFormat,
    DocumentIndexStateName,
    DocumentSort,
    DocumentSortField,
    DocumentSortOrder,
} from '@/service/documents';

export const documentFormats: DocumentFormat[] = [
    'pdf',
    'text',
    'spreadsheet',
    'presentation',
    'image',
    'other',
];
export const documentIndexStates: DocumentIndexStateName[] = [
    'indexed',
    'partial',
    'name_only',
    'queued',
    'failed',
];
export const documentSortFields: DocumentSortField[] = [
    'modified_at',
    'name',
    'size',
    'created_at',
];
export const documentSortOrders: DocumentSortOrder[] = ['asc', 'desc'];

export type DocumentsFilterState = {
    format: DocumentFormat | '';
    modifiedFrom: string;
    modifiedTo: string;
    favoritesOnly: boolean;
    indexState: DocumentIndexStateName | '';
};

export type DocumentsSortState = {
    sort: DocumentSortField;
    order: DocumentSortOrder;
};

export const emptyDocumentsFilters: DocumentsFilterState = {
    format: '',
    modifiedFrom: '',
    modifiedTo: '',
    favoritesOnly: false,
    indexState: '',
};

export const defaultDocumentsSort: DocumentsSortState = { sort: 'modified_at', order: 'desc' };

const defaultOrderBySortField: Record<DocumentSortField, DocumentSortOrder> = {
    modified_at: 'desc',
    name: 'asc',
    size: 'desc',
    created_at: 'desc',
};

const isoDatePattern = /^\d{4}-\d{2}-\d{2}$/;

const filterParamNames = [
    'format',
    'modified_from',
    'modified_to',
    'favorites_only',
    'index_state',
];
const sortParamNames = ['sort', 'order'];

const pickKnown = <Allowed extends string>(
    rawValue: string | null,
    allowedValues: readonly Allowed[]
): Allowed | undefined => allowedValues.find((allowedValue) => allowedValue === rawValue);

const pickIsoDate = (rawValue: string | null): string =>
    rawValue !== null && isoDatePattern.test(rawValue) ? rawValue : '';

export const defaultOrderFor = (sort: DocumentSortField): DocumentSortOrder =>
    defaultOrderBySortField[sort];

export const parseDocumentsFilters = (searchParams: URLSearchParams): DocumentsFilterState => ({
    format: pickKnown(searchParams.get('format'), documentFormats) ?? '',
    modifiedFrom: pickIsoDate(searchParams.get('modified_from')),
    modifiedTo: pickIsoDate(searchParams.get('modified_to')),
    favoritesOnly: searchParams.get('favorites_only') === 'true',
    indexState: pickKnown(searchParams.get('index_state'), documentIndexStates) ?? '',
});

export const parseDocumentsSort = (searchParams: URLSearchParams): DocumentsSortState | null => {
    const sort = pickKnown(searchParams.get('sort'), documentSortFields);
    if (!sort) return null;
    const order = pickKnown(searchParams.get('order'), documentSortOrders) ?? defaultOrderFor(sort);
    return { sort, order };
};

export const writeDocumentsFilters = (
    currentParams: URLSearchParams,
    filters: DocumentsFilterState
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    filterParamNames.forEach((paramName) => nextParams.delete(paramName));
    if (filters.format) nextParams.set('format', filters.format);
    if (filters.modifiedFrom) nextParams.set('modified_from', filters.modifiedFrom);
    if (filters.modifiedTo) nextParams.set('modified_to', filters.modifiedTo);
    if (filters.favoritesOnly) nextParams.set('favorites_only', 'true');
    if (filters.indexState) nextParams.set('index_state', filters.indexState);
    return nextParams;
};

export const parseDocumentsFolderId = (searchParams: URLSearchParams): number | undefined => {
    const rawFolderId = searchParams.get('folder_id');
    if (rawFolderId === null || !/^\d+$/.test(rawFolderId)) return undefined;
    const folderId = Number(rawFolderId);
    return folderId > 0 ? folderId : undefined;
};

export const writeDocumentsFolderId = (
    currentParams: URLSearchParams,
    folderId: number | undefined
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    nextParams.delete('folder_id');
    if (folderId !== undefined) nextParams.set('folder_id', String(folderId));
    return nextParams;
};

export const writeDocumentsSort = (
    currentParams: URLSearchParams,
    sort: DocumentsSortState
): URLSearchParams => {
    const nextParams = new URLSearchParams(currentParams);
    sortParamNames.forEach((paramName) => nextParams.delete(paramName));
    nextParams.set('sort', sort.sort);
    nextParams.set('order', sort.order);
    return nextParams;
};

export const countActiveDocumentsFilters = (filters: DocumentsFilterState): number =>
    [
        filters.format !== '',
        filters.modifiedFrom !== '' || filters.modifiedTo !== '',
        filters.favoritesOnly,
        filters.indexState !== '',
    ].filter(Boolean).length;

export const toDocumentFilters = (filters: DocumentsFilterState): DocumentFilters => ({
    format: filters.format || undefined,
    modifiedFrom: filters.modifiedFrom || undefined,
    modifiedTo: filters.modifiedTo || undefined,
    favoritesOnly: filters.favoritesOnly || undefined,
    indexState: filters.indexState || undefined,
});

export const toDocumentSort = (sort: DocumentsSortState): DocumentSort => ({
    sort: sort.sort,
    order: sort.order,
});
