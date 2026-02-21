import { apiBase } from '@/service';
import { Pagination } from '@/types/pagination';
import {
	FetchNextPageOptions,
	InfiniteData,
	InfiniteQueryObserverResult,
	useInfiniteQuery,
} from '@tanstack/react-query';
import { createContext, useContext, useState } from 'react';
import { useIntersectionObserver } from '../IntersectionObserver/useIntersectionObserver';
import { formatSize } from '@/utils';

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
		options?: FetchNextPageOptions | undefined,
	) => Promise<InfiniteQueryObserverResult<InfiniteData<PaginationResponse, unknown>, Error>>;
	hasNextPage: boolean;
	isFetchingNextPage: boolean;
	getMusicArtist: (music: IMusicData) => string;
	formatDuration: (seconds: number) => string;
	musicMetadata: (music: { format: string; size: number; metadata?: IMusicMetadata }) => string;
	getMusicTitle: (music: IMusicData) => string;
	lastItemRef: (node: HTMLLIElement | null) => (() => void) | undefined;
	playTrack: (track: IMusicData) => void;
	playlist: IMusicData[];
	hasTrackInPlaylist: boolean;
	currentTrack: number | undefined;
	setCurrentTrack: (index: number) => void;
}

type PaginationResponse = Pagination<IMusicData>;

const MusicContext = createContext<IMusicContext | undefined>(undefined);

export const MusicContextProvider = MusicContext.Provider;

const pageSize = 200;

export const MusicProvider = ({ children }: { children: React.ReactNode }) => {
	const [playlist, setPlaylist] = useState<IMusicData[]>([]);
	const [currentTrack, setCurrentTrack] = useState<number | undefined>(undefined);

	console.log(playlist);
	console.log(currentTrack);

	const { status, data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['music'],
		queryFn: async ({ pageParam = 1 }): Promise<PaginationResponse> => {
			const response = await apiBase.get<PaginationResponse>(`/files/music`, {
				params: { page: pageParam, page_size: pageSize },
			});
			return response.data;
		},
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

	const getMusicArtist = (music: IMusicData): string => {
		if (music.metadata?.artist) {
			return music.metadata.artist;
		}
		return 'Unknown Artist';
	};

	const formatDuration = (seconds: number): string => {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	};

	const musicMetadata = (music: { format: string; size: number; metadata?: IMusicMetadata }): string => {
		const format = music.format ? `${music.format} - ` : '';
		const fileSize = formatSize(music.size);
		const duration = music.metadata?.duration ? formatDuration(music.metadata.duration) : '';
		return `${format}${fileSize}${duration ? ` - ${duration}` : ''}`;
	};

	const getMusicTitle = (music: IMusicData): string => {
		if (music.metadata?.title) {
			return music.metadata.title;
		}
		return music.name;
	};

	const playTrack = (track: IMusicData) => {
		if (playlist.some((t) => t.id === track.id)) {
			return;
		}
		setPlaylist((prev) => [...prev, track]);
		if (currentTrack === undefined) {
			setCurrentTrack(0);
		}
	};

	const hasTrackInPlaylist = playlist.length > 0;

	return (
		<MusicContextProvider
			value={{
				music: allMusic,
				status,
				fetchNextPage,
				hasNextPage,
				isFetchingNextPage,
				getMusicArtist,
				formatDuration,
				musicMetadata,
				getMusicTitle,
				lastItemRef,
				playTrack,
				playlist,
				hasTrackInPlaylist,
				currentTrack,
				setCurrentTrack,
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
