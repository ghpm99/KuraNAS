/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import {
	getAutomaticPlaylists,
	createPlaylist,
	deletePlaylist,
	getPlaylistTracks,
	getPlaylists,
	removeTrackFromPlaylist,
} from '@/service/playlist';
import { PlaylistsContextData } from './playlistsContext';
import { useSearchParams } from 'react-router-dom';

const PlaylistsContext = createContext<PlaylistsContextData | undefined>(undefined);

export function PlaylistsProvider({ children }: { children: ReactNode }) {
	const queryClient = useQueryClient();
	const { enqueueSnackbar } = useSnackbar();
	const { t } = useI18n();
	const [selectedPlaylist, setSelectedPlaylist] = useState<Playlist | null>(null);
	const [searchParams, setSearchParams] = useSearchParams();
	const [createOpen, setCreateOpen] = useState(false);
	const [newName, setNewName] = useState('');
	const [newDescription, setNewDescription] = useState('');

	const playlistQueryFn = async (pageParam: number): Promise<Pagination<Playlist>> => getPlaylists(pageParam, 50);
	const playlistTracksQueryFn = async (pageParam: number): Promise<Pagination<PlaylistTrack>> => {
		if (!selectedPlaylist) {
			return {
				items: [],
				pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
			};
		}
		return getPlaylistTracks(selectedPlaylist.id, pageParam, 50);
	};

	const automaticPlaylistsQuery = useQuery({
		queryKey: ['automatic-playlists'],
		queryFn: getAutomaticPlaylists,
	});

	const playlistsQuery = useInfiniteQuery({
		queryKey: ['playlists'],
		queryFn: ({ pageParam = 1 }) => playlistQueryFn(pageParam),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const tracksQuery = useInfiniteQuery({
		queryKey: ['playlist-tracks', selectedPlaylist?.id],
		enabled: Boolean(selectedPlaylist),
		queryFn: ({ pageParam = 1 }) => playlistTracksQueryFn(pageParam),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const createMutation = useMutation({
		mutationFn: () => createPlaylist({ name: newName, description: newDescription }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			setCreateOpen(false);
			setNewName('');
			setNewDescription('');
			enqueueSnackbar(t('MUSIC_PLAYLIST_CREATED'), { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar(t('MUSIC_PLAYLIST_CREATE_FAILED'), { variant: 'error' });
		},
	});

	const deleteMutation = useMutation({
		mutationFn: (id: number) => deletePlaylist(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			enqueueSnackbar(t('MUSIC_PLAYLIST_DELETED'), { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar(t('MUSIC_PLAYLIST_DELETE_FAILED'), { variant: 'error' });
		},
	});

	const removeMutation = useMutation({
		mutationFn: (fileId: number) => {
			if (!selectedPlaylist) {
				return Promise.resolve();
			}
			return removeTrackFromPlaylist(selectedPlaylist.id, fileId);
		},
		onSuccess: () => {
			if (!selectedPlaylist) {
				return;
			}
			queryClient.invalidateQueries({ queryKey: ['playlist-tracks', selectedPlaylist.id] });
			queryClient.invalidateQueries({ queryKey: ['playlists'] });
			enqueueSnackbar(t('MUSIC_TRACK_REMOVED'), { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar(t('MUSIC_TRACK_REMOVE_FAILED'), { variant: 'error' });
		},
	});

	const playlists = useMemo(
		() => [...(automaticPlaylistsQuery.data ?? []), ...(playlistsQuery.data?.pages.flatMap((page) => page.items) ?? [])],
		[automaticPlaylistsQuery.data, playlistsQuery.data],
	);
	const tracks = useMemo(() => tracksQuery.data?.pages.flatMap((page) => page.items) ?? [], [tracksQuery.data]);
	const requestedPlaylistId = Number(searchParams.get('playlist') ?? '');

	useEffect(() => {
		if (!Number.isFinite(requestedPlaylistId) || requestedPlaylistId <= 0 || selectedPlaylist?.id === requestedPlaylistId) {
			return;
		}

		const requestedPlaylist = playlists.find((playlist) => playlist.id === requestedPlaylistId) ?? null;
		if (requestedPlaylist) {
			setSelectedPlaylist(requestedPlaylist);
		}
	}, [playlists, requestedPlaylistId, selectedPlaylist?.id]);

	const contextValue: PlaylistsContextData = {
		selectedPlaylist,
		playlists,
		tracks,
		isLoadingPlaylists: playlistsQuery.isLoading || automaticPlaylistsQuery.isLoading,
		isLoadingTracks: tracksQuery.isLoading,
		hasNextPlaylistPage: Boolean(playlistsQuery.hasNextPage),
		hasNextTrackPage: Boolean(tracksQuery.hasNextPage),
		isFetchingNextPlaylistPage: playlistsQuery.isFetchingNextPage,
		isFetchingNextTrackPage: tracksQuery.isFetchingNextPage,
		isCreatingPlaylist: createMutation.isPending,
		isDeletingPlaylist: deleteMutation.isPending,
		isRemovingTrack: removeMutation.isPending,
		createOpen,
		newName,
		newDescription,
			selectPlaylist: (playlist) => {
				setSelectedPlaylist(playlist);
				setSearchParams((current) => {
					const next = new URLSearchParams(current);
					next.set('playlist', String(playlist.id));
					return next;
				}, { replace: true });
			},
			backToList: () => {
				setSelectedPlaylist(null);
				setSearchParams((current) => {
					const next = new URLSearchParams(current);
					next.delete('playlist');
					return next;
				}, { replace: true });
			},
		fetchNextPlaylistPage: playlistsQuery.fetchNextPage,
		fetchNextTrackPage: tracksQuery.fetchNextPage,
		openCreateDialog: () => setCreateOpen(true),
		closeCreateDialog: () => setCreateOpen(false),
		setNewName,
		setNewDescription,
		submitCreatePlaylist: () => createMutation.mutate(),
		deletePlaylistById: (id) => deleteMutation.mutate(id),
		removeTrackByFileId: (fileId) => removeMutation.mutate(fileId),
		playlistQueryFn,
			playlistTracksQueryFn,
		};

	return <PlaylistsContext.Provider value={contextValue}>{children}</PlaylistsContext.Provider>;
}

export function usePlaylistsProvider() {
	const context = useContext(PlaylistsContext);
	if (!context) {
		throw new Error('usePlaylistsProvider must be used within PlaylistsProvider');
	}
	return context;
}
