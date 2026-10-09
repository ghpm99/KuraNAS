import { Box, Grid } from '@mui/material';
import { useSearchParams } from 'react-router-dom';
import { getMusicAlbums } from '@/service/music';
import { MusicAlbum } from '@/types/music';
import { MUSIC_COLLECTION_PAGE_SIZE } from './shared';
import AlbumCard from './components/AlbumCard';
import AlbumPage from './components/AlbumPage';
import MusicCollectionListFeedback from './components/MusicCollectionListFeedback';
import MusicSortControl from './components/MusicSortControl';
import { useMusicInfinitePages } from './useMusicInfinitePages';
import { useMusicListSort } from './useMusicListSort';

export default function AlbumsView() {
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedAlbumKey = searchParams.get('album') ?? '';
    const { listSort, changeField, toggleOrder } = useMusicListSort('albums');
    const albumsQuery = useMusicInfinitePages<MusicAlbum>(
        ['music-albums', listSort],
        (pageNumber) => getMusicAlbums(pageNumber, MUSIC_COLLECTION_PAGE_SIZE, listSort)
    );
    const albums = albumsQuery.items;

    const handleSelectAlbum = (album: MusicAlbum) => {
        setSearchParams((current) => {
            const next = new URLSearchParams(current);
            next.set('album', album.key);
            return next;
        });
    };

    const handleBack = () => {
        setSearchParams(
            (current) => {
                const next = new URLSearchParams(current);
                next.delete('album');
                return next;
            },
            { replace: true }
        );
    };

    if (selectedAlbumKey) {
        return <AlbumPage albumKey={selectedAlbumKey} onBack={handleBack} />;
    }

    return (
        <>
            <MusicSortControl
                view="albums"
                listSort={listSort}
                onFieldChange={changeField}
                onOrderToggle={toggleOrder}
            />
            <AlbumListView
                albums={albums}
                isLoading={albumsQuery.isLoading}
                isError={albumsQuery.isError}
                errorMessage={albumsQuery.errorMessage}
                onRetry={albumsQuery.retry}
                fetchNextPage={albumsQuery.fetchNextPage}
                hasNextPage={albumsQuery.hasNextPage}
                isFetchingNextPage={albumsQuery.isFetchingNextPage}
                onSelect={handleSelectAlbum}
            />
        </>
    );
}

type AlbumListViewProps = {
    albums: MusicAlbum[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    onRetry: () => void;
    fetchNextPage: () => void;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (album: MusicAlbum) => void;
};

function AlbumListView({
    albums,
    isLoading,
    isError,
    errorMessage,
    onRetry,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
}: AlbumListViewProps) {
    return (
        <MusicCollectionListFeedback
            isLoading={isLoading}
            isError={isError}
            errorMessage={errorMessage}
            isEmpty={albums.length === 0}
            emptyTitleKey="MUSIC_ALBUMS_EMPTY"
            errorTitleKey="MUSIC_LIST_ERROR_TITLE"
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onRetry={onRetry}
            fetchNextPage={fetchNextPage}
        >
            <Box sx={{ p: 2 }}>
                <Grid container spacing={2}>
                    {albums.map((album) => (
                        <Grid key={album.key} size={{ xs: 6, sm: 4, md: 3, lg: 2.4 }}>
                            <AlbumCard album={album} onSelect={onSelect} />
                        </Grid>
                    ))}
                </Grid>
            </Box>
        </MusicCollectionListFeedback>
    );
}
