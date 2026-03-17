/* eslint-disable react-refresh/only-export-components */
import { Pagination } from '@/types/pagination';
import {
    FetchNextPageOptions,
    InfiniteData,
    InfiniteQueryObserverResult,
    useInfiniteQuery,
} from '@tanstack/react-query';
import { createContext, useContext } from 'react';
import { useIntersectionObserver } from '@/components/hooks/IntersectionObserver/useIntersectionObserver';
import { getMusic } from '@/service/music';

export interface IMusicMetadata {
    id: number;
    fileId: number;
    path: string;
    format: string;
    title: string;
    artist: string;
    album: string;
    year: number;
    genre: string;
    track: number;
    disc: number;
    duration: number;
    bitrate: number;
    sampleRate: number;
    channels: number;
    createdAt: string;
}

export interface IMusicData {
    id: number;
    name: string;
    path: string;
    type: number;
    format: string;
    size: number;
    updated_at: string;
    created_at: string;
    deleted_at: string;
    last_interaction: string;
    last_backup: string;
    check_sum: string;
    directory_content_count: number;
    starred: boolean;
    metadata?: IMusicMetadata;
}

export interface IMusicContext {
    music: IMusicData[];
    status: 'error' | 'success' | 'pending';
    fetchNextPage: (
        options?: FetchNextPageOptions | undefined
    ) => Promise<InfiniteQueryObserverResult<InfiniteData<PaginationResponse, unknown>, Error>>;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    lastItemRef: (node: HTMLLIElement | null) => (() => void) | undefined;
}

type PaginationResponse = Pagination<IMusicData>;

const MusicContext = createContext<IMusicContext | undefined>(undefined);

export const MusicContextProvider = MusicContext.Provider;

const pageSize = 200;

export const MusicProvider = ({ children }: { children: React.ReactNode }) => {
    const { status, data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music'],
        queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> => getMusic(pageParam, pageSize),
        initialPageParam: 1,
        getNextPageParam: (lastPage) => {
            if (lastPage.pagination.has_next) {
                return lastPage.pagination.page + 1;
            }
            return undefined;
        },
        staleTime: 0,
    });

    const { ref: lastItemRef } = useIntersectionObserver<HTMLLIElement>({
        enabled: hasNextPage && !isFetchingNextPage,
        rootMargin: '400px',
        onIntersect: () => {
            if (hasNextPage && !isFetchingNextPage) {
                fetchNextPage();
            }
        },
    });

    const allMusic = data?.pages.flatMap((page) => page.items) ?? [];

    return (
        <MusicContextProvider
            value={{
                music: allMusic,
                status,
                fetchNextPage,
                hasNextPage,
                isFetchingNextPage,
                lastItemRef,
            }}
        >
            {children}
        </MusicContextProvider>
    );
};

export const useMusic = () => {
    const context = useContext(MusicContext);
    if (!context) {
        throw new Error('useMusic must be used within a MusicContextProvider');
    }
    return context;
};
