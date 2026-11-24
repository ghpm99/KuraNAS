/* eslint-disable react-refresh/only-export-components */
import {
    createContext,
    useCallback,
    useContext,
    useMemo,
    useState,
    type ReactNode,
} from 'react';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistRules, PlaylistTrack } from '@/types/playlist';
import {
    getAutomaticPlaylists,
    createPlaylist,
    deletePlaylist,
    getPlaylistTracks,
    getPlaylists,
    moveTrackInPlaylist,
    removeTrackFromPlaylist,
    updatePlaylist,
} from '@/service/playlist';
import { PlaylistKindChoice, PlaylistsContextData } from './playlistsContext';
import { createEmptyPlaylistRules, normalizePlaylistRules } from './playlistRuleCriteria';
import { extractBackendErrorField } from '@/shared/utils/extractBackendErrorField';
import { useSearchParams } from 'react-router-dom';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const CREATE_RULE_GENRE_PARAM = 'create_rule_genre';

const PlaylistsContext = createContext<PlaylistsContextData | undefined>(undefined);

export function PlaylistsProvider({ children }: { children: ReactNode }) {
    const queryClient = useQueryClient();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();
    const [searchParams, setSearchParams] = useSearchParams();
    const requestedRuleGenre = searchParams.get(CREATE_RULE_GENRE_PARAM);
    const [createOpen, setCreateOpen] = useState(requestedRuleGenre !== null);
    const [newName, setNewName] = useState('');
    const [newDescription, setNewDescription] = useState('');
    const [newKind, setNewKind] = useState<PlaylistKindChoice>(
        requestedRuleGenre === null ? 'manual' : 'rule'
    );
    const [newRules, setNewRulesState] = useState<PlaylistRules>(() =>
        normalizePlaylistRules({ genres: requestedRuleGenre ? [requestedRuleGenre] : [] })
    );
    const [createErrorField, setCreateErrorField] = useState<string | null>(null);
    const [renameErrorField, setRenameErrorField] = useState<string | null>(null);

    const clearRequestedRuleGenre = useCallback(() => {
        if (!searchParams.has(CREATE_RULE_GENRE_PARAM)) {
            return;
        }
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete(CREATE_RULE_GENRE_PARAM);
                return next;
            },
            { replace: true }
        );
    }, [searchParams, setSearchParams]);

    const setNewRules = useCallback((rules: PlaylistRules) => {
        setNewRulesState(rules);
        setCreateErrorField(null);
    }, []);

    const openCreateRuleDialog = useCallback((initialRules?: Partial<PlaylistRules>) => {
        setNewName('');
        setNewDescription('');
        setNewKind('rule');
        setNewRulesState(normalizePlaylistRules(initialRules));
        setCreateErrorField(null);
        setCreateOpen(true);
    }, []);

    const automaticPlaylistsQuery = useQuery({
        queryKey: ['automatic-playlists'],
        queryFn: getAutomaticPlaylists,
    });

    const playlistQueryFn = useCallback(
        async (pageParam: number): Promise<Pagination<Playlist>> => getPlaylists(pageParam, 50),
        []
    );

    const playlistsQuery = useInfiniteQuery({
        queryKey: ['playlists'],
        queryFn: ({ pageParam = 1 }) => playlistQueryFn(pageParam),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const playlists = useMemo(
        () => [
            ...(automaticPlaylistsQuery.data ?? []),
            ...(playlistsQuery.data?.pages.flatMap((page) => page.items) ?? []),
        ],
        [automaticPlaylistsQuery.data, playlistsQuery.data]
    );

    const selectedPlaylistId = Number(searchParams.get('playlist') ?? '');
    const selectedPlaylist = useMemo(
        () =>
            Number.isInteger(selectedPlaylistId) && selectedPlaylistId !== 0
                ? (playlists.find((playlist) => playlist.id === selectedPlaylistId) ?? null)
                : null,
        [playlists, selectedPlaylistId]
    );

    const playlistTracksQueryFn = useCallback(
        async (pageParam: number): Promise<Pagination<PlaylistTrack>> => {
            if (!selectedPlaylist) {
                return {
                    items: [],
                    pagination: {
                        page: 1,
                        page_size: 50,
                        has_next: false,
                        has_prev: false,
                    },
                };
            }
            return getPlaylistTracks(selectedPlaylist.id, pageParam, 50);
        },
        [selectedPlaylist]
    );

    const tracksQuery = useInfiniteQuery({
        queryKey: ['playlist-tracks', selectedPlaylist?.id],
        enabled: Boolean(selectedPlaylist),
        queryFn: ({ pageParam = 1 }) => playlistTracksQueryFn(pageParam),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const createMutation = useMutation({
        mutationFn: () =>
            createPlaylist({
                name: newName,
                description: newDescription,
                ...(newKind === 'rule' ? { rules: newRules } : {}),
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            setCreateOpen(false);
            clearRequestedRuleGenre();
            setNewName('');
            setNewDescription('');
            setNewKind('manual');
            setNewRulesState(createEmptyPlaylistRules());
            setCreateErrorField(null);
            enqueueSnackbar(t('MUSIC_PLAYLIST_CREATED'), { variant: 'success' });
        },
        onError: (error) => {
            setCreateErrorField(extractBackendErrorField(error) ?? null);
            enqueueSnackbar(
                extractBackendErrorMessage(error) ?? t('MUSIC_PLAYLIST_CREATE_FAILED'),
                { variant: 'error' }
            );
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
            queryClient.invalidateQueries({
                queryKey: ['playlist-tracks', selectedPlaylist.id],
            });
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            enqueueSnackbar(t('MUSIC_TRACK_REMOVED'), { variant: 'success' });
        },
        onError: () => {
            enqueueSnackbar(t('MUSIC_TRACK_REMOVE_FAILED'), { variant: 'error' });
        },
    });

    const renameMutation = useMutation({
        mutationFn: ({
            name,
            description,
            rules,
        }: {
            name: string;
            description: string;
            rules?: PlaylistRules;
        }) => {
            if (!selectedPlaylist) {
                return Promise.resolve(null);
            }
            return updatePlaylist(selectedPlaylist.id, {
                name,
                description,
                ...(rules ? { rules } : {}),
            });
        },
        onSuccess: () => {
            setRenameErrorField(null);
            queryClient.invalidateQueries({ queryKey: ['playlists'] });
            enqueueSnackbar(t('MUSIC_PLAYLIST_UPDATED'), { variant: 'success' });
        },
        onError: (error) => {
            setRenameErrorField(extractBackendErrorField(error) ?? null);
            enqueueSnackbar(
                extractBackendErrorMessage(error) ?? t('MUSIC_PLAYLIST_UPDATE_FAILED'),
                {
                    variant: 'error',
                }
            );
        },
    });

    const moveTrackMutation = useMutation({
        mutationFn: ({ fileId, position }: { fileId: number; position: number }) => {
            if (!selectedPlaylist) {
                return Promise.resolve();
            }
            return moveTrackInPlaylist(selectedPlaylist.id, fileId, position);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['playlist-tracks', selectedPlaylist?.id],
            });
        },
        onError: (error) => {
            enqueueSnackbar(
                extractBackendErrorMessage(error) ?? t('MUSIC_PLAYLIST_REORDER_FAILED'),
                { variant: 'error' }
            );
            queryClient.invalidateQueries({
                queryKey: ['playlist-tracks', selectedPlaylist?.id],
            });
        },
    });

    const tracks = useMemo(
        () => tracksQuery.data?.pages.flatMap((page) => page.items) ?? [],
        [tracksQuery.data]
    );

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
        isRenamingPlaylist: renameMutation.isPending,
        isMovingTrack: moveTrackMutation.isPending,
        createOpen,
        newName,
        newDescription,
        newKind,
        newRules,
        createErrorField,
        renameErrorField,
        selectPlaylist: (playlist) => {
            setSearchParams((current) => {
                const next = new URLSearchParams(current);
                next.set('playlist', String(playlist.id));
                return next;
            });
        },
        backToList: () => {
            setSearchParams(
                (current) => {
                    const next = new URLSearchParams(current);
                    next.delete('playlist');
                    return next;
                },
                { replace: true }
            );
        },
        fetchNextPlaylistPage: playlistsQuery.fetchNextPage,
        fetchNextTrackPage: tracksQuery.fetchNextPage,
        openCreateDialog: () => {
            setNewKind('manual');
            setNewRulesState(createEmptyPlaylistRules());
            setCreateErrorField(null);
            setCreateOpen(true);
        },
        openCreateRuleDialog,
        closeCreateDialog: () => {
            setCreateOpen(false);
            clearRequestedRuleGenre();
        },
        setNewKind,
        setNewRules,
        setNewName,
        setNewDescription,
        submitCreatePlaylist: () => createMutation.mutate(),
        deletePlaylistById: (id) => deleteMutation.mutate(id),
        removeTrackByFileId: (fileId) => removeMutation.mutate(fileId),
        renameSelectedPlaylist: (name, description, onSaved, rules) => {
            setRenameErrorField(null);
            renameMutation.mutate({ name, description, rules }, { onSuccess: onSaved });
        },
        moveTrackToPosition: (fileId, position) => moveTrackMutation.mutate({ fileId, position }),
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
