import { useQuery } from '@tanstack/react-query';

export function useMusicGroupSummary<SummaryType>(
    queryKey: readonly unknown[],
    fetchSummary: () => Promise<SummaryType>
): SummaryType | null {
    const summaryQuery = useQuery({
        queryKey,
        queryFn: async () => (await fetchSummary()) ?? null,
        retry: false,
    });
    return summaryQuery.data ?? null;
}
