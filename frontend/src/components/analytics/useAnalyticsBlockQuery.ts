import { useQuery, type QueryKey } from '@tanstack/react-query';
import { analyticsStaleTimeMs } from '@/components/providers/queryFreshness';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

interface AnalyticsBlockQueryOptions<BlockData> {
    queryKey: QueryKey;
    fetchBlock: () => Promise<BlockData>;
}

export type AnalyticsBlockQueryResult<BlockData> =
    | { status: 'pending'; data: undefined; errorMessage: undefined; retry: () => void }
    | { status: 'error'; data: undefined; errorMessage: string | undefined; retry: () => void }
    | { status: 'success'; data: BlockData; errorMessage: undefined; retry: () => void };

export const useAnalyticsBlockQuery = <BlockData>({
    queryKey,
    fetchBlock,
}: AnalyticsBlockQueryOptions<BlockData>): AnalyticsBlockQueryResult<BlockData> => {
    const query = useQuery({
        queryKey,
        queryFn: fetchBlock,
        retry: false,
        staleTime: analyticsStaleTimeMs,
        placeholderData: (previousData) => previousData,
    });

    const retry = () => void query.refetch();

    if (query.isError) {
        return {
            status: 'error',
            data: undefined,
            errorMessage: extractBackendErrorMessage(query.error),
            retry,
        };
    }
    if (query.isPending) {
        return { status: 'pending', data: undefined, errorMessage: undefined, retry };
    }
    return { status: 'success', data: query.data, errorMessage: undefined, retry };
};

export default useAnalyticsBlockQuery;
