import {
    Box,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
} from '@mui/material';
import { Folder, Play } from 'lucide-react';
import { useSearchParams } from 'react-router-dom';
import CollectionContextMenu from '@/features/music/components/contextMenu/CollectionContextMenu';
import CategoryHeader from '@/features/music/components/CategoryHeader';
import { createFolderPlaybackContext } from '@/features/music/components/playbackContext';
import { queueToTracks, findStartIndex } from '@/features/music/components/musicQueueTracks';
import { shuffleItems } from '@/utils/shuffleItems';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { IMusicData } from '@/types/music';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    getMusicByFolder,
    getMusicFolders,
    getMusicFolderSummary,
    getMusicQueueByFolder,
} from '@/service/music';
import { MusicFolder, MusicGroupSummary } from '@/types/music';
import { getFolderName, handleKeyboardActivation, MUSIC_COLLECTION_PAGE_SIZE } from './shared';
import MusicCollectionListFeedback from './components/MusicCollectionListFeedback';
import MusicCollectionTrackList from './components/MusicCollectionTrackList';
import MusicSortControl from './components/MusicSortControl';
import { useMusicGroupSummary } from './useMusicGroupSummary';
import { useMusicInfinitePages } from './useMusicInfinitePages';
import { useMusicListSort } from './useMusicListSort';

const loadFolderTracks = (folderPath: string) =>
    getMusicQueueByFolder(folderPath).then(queueToTracks);

export default function FoldersView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedFolderPath = searchParams.get('folder') ?? '';
    const { listSort, changeField, toggleOrder } = useMusicListSort('folders');
    const foldersQuery = useMusicInfinitePages<MusicFolder>(
        ['music-folders', listSort],
        (pageNumber) => getMusicFolders(pageNumber, MUSIC_COLLECTION_PAGE_SIZE, listSort)
    );
    const folders = foldersQuery.items;

    const handleSelectFolder = (folder: string) => {
        setSearchParams((current) => {
            const next = new URLSearchParams(current);
            next.set('folder', folder);
            return next;
        });
    };

    const handleBack = () => {
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete('folder');
                return next;
            },
            { replace: true }
        );
    };

    if (selectedFolderPath) {
        return <FolderTracksView folder={selectedFolderPath} onBack={handleBack} />;
    }

    return (
        <>
            <MusicSortControl
                view="folders"
                listSort={listSort}
                onFieldChange={changeField}
                onOrderToggle={toggleOrder}
            />
            <FolderListView
                folders={folders}
                isLoading={foldersQuery.isLoading}
                isError={foldersQuery.isError}
                errorMessage={foldersQuery.errorMessage}
                onRetry={foldersQuery.retry}
                fetchNextPage={foldersQuery.fetchNextPage}
                hasNextPage={foldersQuery.hasNextPage}
                isFetchingNextPage={foldersQuery.isFetchingNextPage}
                onSelect={handleSelectFolder}
            />
        </>
    );
}

type FolderListViewProps = {
    folders: MusicFolder[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    onRetry: () => void;
    fetchNextPage: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (folder: string) => void;
};

function FolderListView({
    folders,
    isLoading,
    isError,
    errorMessage,
    onRetry,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: FolderListViewProps) {
    const { t } = useI18n();
    const { replaceQueue } = useGlobalMusic();

    const handlePlayFolder = async (event: React.MouseEvent, folderPath: string) => {
        event.stopPropagation();
        const tracks = await loadFolderTracks(folderPath);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createFolderPlaybackContext(folderPath));
        }
    };

    return (
        <MusicCollectionListFeedback
            isLoading={isLoading}
            isError={isError}
            errorMessage={errorMessage}
            isEmpty={folders.length === 0}
            emptyTitleKey="MUSIC_FOLDERS_EMPTY"
            errorTitleKey="MUSIC_LIST_ERROR_TITLE"
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onRetry={onRetry}
            fetchNextPage={fetchNextPage}
        >
            <Box sx={{ p: 1 }}>
                <List sx={{ width: '100%' }}>
                    {folders.map((folder) => (
                        <CollectionContextMenu
                            key={folder.folder}
                            collectionName={getFolderName(folder.folder)}
                            playbackContext={createFolderPlaybackContext(folder.folder)}
                            loadTracks={() => loadFolderTracks(folder.folder)}
                            layout="row"
                        >
                            <ListItem
                                key={folder.folder}
                                disablePadding
                                sx={{
                                    '&:hover .folder-play': { opacity: 1 },
                                }}
                            >
                                <ListItemButton
                                    component="div"
                                    role="button"
                                    tabIndex={0}
                                    onClick={() => onSelect(folder.folder)}
                                    onKeyDown={(event) =>
                                        handleKeyboardActivation(event, () =>
                                            onSelect(folder.folder)
                                        )
                                    }
                                    sx={{ borderRadius: 1.5, py: 1, pl: 1.5, pr: 6, gap: 1 }}
                                >
                                    <ListItemIcon sx={{ minWidth: 40 }}>
                                        <Box
                                            sx={{
                                                width: 40,
                                                height: 40,
                                                borderRadius: 1,
                                                bgcolor: 'rgba(var(--app-color-primary-rgb), 0.12)',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                            }}
                                        >
                                            <Folder size={20} color="var(--app-color-primary)" />
                                        </Box>
                                    </ListItemIcon>
                                    <ListItemText
                                        primary={getFolderName(folder.folder)}
                                        secondary={`${folder.track_count} ${t('MUSIC_TRACKS_COUNT')}`}
                                        primaryTypographyProps={{ fontWeight: 500 }}
                                    />
                                    <IconButton
                                        className="folder-play"
                                        onClick={(event) =>
                                            void handlePlayFolder(event, folder.folder)
                                        }
                                        sx={{
                                            opacity: 0,
                                            transition: 'all 0.2s ease',
                                            color: 'primary.main',
                                            '&:hover': {
                                                bgcolor: 'rgba(var(--app-color-primary-rgb), 0.12)',
                                            },
                                        }}
                                    >
                                        <Play size={18} fill="var(--app-color-primary)" />
                                    </IconButton>
                                </ListItemButton>
                            </ListItem>
                        </CollectionContextMenu>
                    ))}
                </List>
            </Box>
        </MusicCollectionListFeedback>
    );
}

function FolderTracksView({ folder, onBack }: { folder: string; onBack: () => void }) {
    const { replaceQueue } = useGlobalMusic();
    const summary = useMusicGroupSummary<MusicGroupSummary>(['music-folder-summary', folder], () =>
        getMusicFolderSummary(folder)
    );
    const tracksQuery = useMusicInfinitePages<IMusicData>(
        ['music-by-folder', folder],
        (pageNumber) => getMusicByFolder(folder, pageNumber, MUSIC_COLLECTION_PAGE_SIZE)
    );
    const tracks = tracksQuery.items;
    const playbackContext = createFolderPlaybackContext(folder);

    const queueFolderTracks = async (trackId?: number, shuffle = false) => {
        const allTracks = await loadFolderTracks(folder);
        if (allTracks.length === 0) {
            return;
        }

        if (shuffle) {
            replaceQueue(shuffleItems(allTracks), 0, playbackContext);
            return;
        }

        const startIndex = findStartIndex(allTracks, trackId);
        replaceQueue(allTracks, startIndex, playbackContext);
    };

    return (
        <Box sx={{ p: 2 }}>
            <CategoryHeader
                title={getFolderName(folder)}
                subtitle={folder}
                trackCount={summary?.track_count}
                totalLengthSeconds={summary?.total_length_seconds}
                icon={<Folder size={48} opacity={0.7} />}
                gradientFrom="var(--app-color-primary)"
                onBack={onBack}
                onPlayAll={() => void queueFolderTracks()}
                onShuffleAll={() => void queueFolderTracks(undefined, true)}
            />

            <MusicCollectionTrackList
                tracks={tracks}
                isLoading={tracksQuery.isLoading}
                isError={tracksQuery.isError}
                errorMessage={tracksQuery.errorMessage}
                hasNextPage={tracksQuery.hasNextPage}
                isFetchingNextPage={tracksQuery.isFetchingNextPage}
                onPlayTrack={(track) => void queueFolderTracks(track.id)}
                onRetry={tracksQuery.retry}
                fetchNextPage={tracksQuery.fetchNextPage}
            />
        </Box>
    );
}
