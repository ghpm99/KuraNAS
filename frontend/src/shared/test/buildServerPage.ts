import type { Pagination } from '@/types/pagination';

interface ServerPageOptions {
    page: number;
    pageSize: number;
    hasNext: boolean;
}

/**
 * Builds a paginated response with exactly the envelope the backend serializes.
 * has_prev is derived from the page number, as the backend does.
 */
export function buildServerPage<T>(items: T[], options: ServerPageOptions): Pagination<T> {
    return {
        items,
        pagination: {
            page: options.page,
            page_size: options.pageSize,
            has_next: options.hasNext,
            has_prev: options.page > 1,
        },
    };
}
