import type { Pagination } from '@/types/pagination';

export const getNextPageNumber = (lastPage: Pagination<unknown> | undefined): number | undefined =>
    lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined;
