/* eslint-disable react-refresh/only-export-components */
import {
    useInfiniteQuery,
    useQuery,
    type FetchNextPageOptions,
    type InfiniteData,
    type InfiniteQueryObserverResult,
    type QueryObserverResult,
} from '@tanstack/react-query';
import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useLocation, useSearchParams } from 'react-router-dom';
import { getImageSectionFromPath } from '@/components/images/navigation';
import {
    parseImageLibraryView,
    type ImageLibraryView,
} from '@/components/imageContent/imageLibraryView';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import {
    getImageLibraryCount,
    getImageLibraryPage,
    getImageLibraryTimeline,
} from '@/service/image';
import type { ImageLibraryItem, ImageLibraryPage, ImageTimelineBucket } from '@/types/imageLibrary';

export const imageLibraryPageSize = 60;

export type ImageLibraryPageParam = { cursor?: string; page: number };

export interface IImageContext {
    view: ImageLibraryView;
    items: ImageLibraryItem[];
    status: 'error' | 'success' | 'pending';
    error: unknown;
    isFetchNextPageError: boolean;
    total: number | null;
    timeline: ImageTimelineBucket[];
    fetchNextPage: (
        options?: FetchNextPageOptions | undefined
    ) => Promise<InfiniteQueryObserverResult<InfiniteData<ImageLibraryPage, unknown>, Error>>;
    refetch: () => Promise<QueryObserverResult<InfiniteData<ImageLibraryPage, unknown>, Error>>;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
}

const ImageContext = createContext<IImageContext | undefined>(undefined);

export const ImageContextProvider = ImageContext.Provider;

const firstPageParam: ImageLibraryPageParam = { page: 1 };

export const ImageProvider = ({ children }: { children: ReactNode }) => {
    const location = useLocation();
    const [searchParams] = useSearchParams();
    const section = getImageSectionFromPath(location.pathname);
    const view = useMemo(
        () => parseImageLibraryView(section, searchParams),
        [section, searchParams]
    );
    const { filters, ordering, takenBefore, isKeyset } = view;
    const isAlbumPicker = section === 'albums' && !view.selectedAlbum;

    const {
        data: libraryData,
        status,
        error,
        isFetchNextPageError,
        fetchNextPage,
        refetch,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteQuery({
        queryKey: ['images', 'library', filters, ordering, takenBefore],
        queryFn: ({ pageParam }): Promise<ImageLibraryPage> =>
            getImageLibraryPage({
                filters,
                ordering,
                pageSize: imageLibraryPageSize,
                cursor: pageParam.cursor,
                page: isKeyset ? undefined : pageParam.page,
                takenBefore: pageParam.cursor ? undefined : takenBefore,
            }),
        initialPageParam: firstPageParam,
        getNextPageParam: (lastPage, loadedPages): ImageLibraryPageParam | undefined => {
            if (!lastPage.has_next) {
                return undefined;
            }
            if (!isKeyset) {
                return { page: (lastPage.page ?? loadedPages.length) + 1 };
            }
            return lastPage.next_cursor
                ? { cursor: lastPage.next_cursor, page: loadedPages.length + 1 }
                : undefined;
        },
        enabled: !isAlbumPicker,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const countQuery = useQuery({
        queryKey: ['images', 'count', filters],
        queryFn: () => getImageLibraryCount(filters),
        enabled: !isAlbumPicker,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const timelineQuery = useQuery({
        queryKey: ['images', 'timeline', filters],
        queryFn: () => getImageLibraryTimeline(filters),
        enabled: !isAlbumPicker && isKeyset,
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
    });

    const items = useMemo(
        () => libraryData?.pages.flatMap((page) => page.items ?? []) ?? [],
        [libraryData]
    );

    const contextValue = useMemo<IImageContext>(
        () => ({
            view,
            items,
            status,
            error,
            isFetchNextPageError,
            total: countQuery.data ?? null,
            timeline: timelineQuery.data ?? [],
            fetchNextPage,
            refetch,
            hasNextPage,
            isFetchingNextPage,
        }),
        [
            view,
            items,
            status,
            error,
            isFetchNextPageError,
            countQuery.data,
            timelineQuery.data,
            fetchNextPage,
            refetch,
            hasNextPage,
            isFetchingNextPage,
        ]
    );

    return <ImageContextProvider value={contextValue}>{children}</ImageContextProvider>;
};

export const useImage = () => {
    const context = useContext(ImageContext);
    if (!context) {
        throw new Error('useImage must be used within an ImageContextProvider');
    }
    return context;
};
