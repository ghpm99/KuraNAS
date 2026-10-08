import { useCallback, useEffect, useState } from 'react';

interface UseInfiniteScrollSentinelOptions {
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    rootMargin?: string;
}

export function useInfiniteScrollSentinel({
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    rootMargin = '200px',
}: UseInfiniteScrollSentinelOptions) {
    const [sentinelElement, setSentinelElement] = useState<HTMLElement | null>(null);
    const canLoadMore = hasNextPage && !isFetchingNextPage;

    const sentinelRef = useCallback((node: HTMLElement | null) => {
        setSentinelElement(node);
    }, []);

    const loadMore = useCallback(() => {
        if (canLoadMore) {
            fetchNextPage();
        }
    }, [canLoadMore, fetchNextPage]);

    useEffect(() => {
        if (!sentinelElement || !canLoadMore || typeof IntersectionObserver === 'undefined') {
            return;
        }
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries.some((entry) => entry.isIntersecting)) {
                    fetchNextPage();
                }
            },
            { rootMargin }
        );
        observer.observe(sentinelElement);
        return () => observer.disconnect();
    }, [sentinelElement, canLoadMore, fetchNextPage, rootMargin]);

    return { sentinelRef, loadMore, canLoadMore };
}

export default useInfiniteScrollSentinel;
