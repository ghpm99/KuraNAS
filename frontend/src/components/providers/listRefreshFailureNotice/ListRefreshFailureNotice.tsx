import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';

type FetchMoreMeta = { fetchMore?: { direction?: string } };

const hasLoadedPages = (queryData: unknown): boolean =>
    Array.isArray((queryData as { pages?: unknown } | undefined)?.pages);

const isNextPageFailure = (fetchMeta: unknown): boolean =>
    (fetchMeta as FetchMoreMeta | null | undefined)?.fetchMore?.direction === 'forward';

/**
 * Warns once when a paginated list that already has data fails to refresh,
 * ignoring next-page failures, which the list itself reports.
 */
const ListRefreshFailureNotice = () => {
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();

    useEffect(() => {
        const unsubscribe = queryClient.getQueryCache().subscribe((cacheEvent) => {
            if (cacheEvent.type !== 'updated' || cacheEvent.action.type !== 'error') return;
            const { data, fetchMeta } = cacheEvent.query.state;
            if (!hasLoadedPages(data) || isNextPageFailure(fetchMeta)) return;
            enqueueSnackbar(t('LIST_REFRESH_FAILED'), {
                variant: 'warning',
                key: 'list-refresh-failed',
                preventDuplicate: true,
            });
        });
        return unsubscribe;
    }, [queryClient, enqueueSnackbar, t]);

    return null;
};

export default ListRefreshFailureNotice;
