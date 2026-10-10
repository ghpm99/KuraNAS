import { getVideoRoute } from '@/app/routes';
import {
    type VideoCatalogItemDto,
    type VideoContinueItemDto,
    type VideoPlaylistDto,
} from '@/service/videoPlayback';
import VideoCatalogRail from './VideoCatalogRail';
import VideoContinueWatchingSection from './VideoContinueWatchingSection';
import VideoSectionPlaylistGrid, { VideoSectionActionLink } from './VideoSectionPlaylistGrid';

type VideoHomeScreenProps = {
    continueWatchingItems: VideoContinueItemDto[];
    seriesPlaylists: VideoPlaylistDto[];
    moviePlaylists: VideoPlaylistDto[];
    personalPlaylists: VideoPlaylistDto[];
    clipPlaylists: VideoPlaylistDto[];
    folderPlaylists: VideoPlaylistDto[];
    recentCatalogItems: VideoCatalogItemDto[];
    onSelectPlaylist: (playlist: VideoPlaylistDto) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
    onSetWatched?: (videoId: number, watched: boolean) => void;
};

export default function VideoHomeScreen({
    continueWatchingItems,
    seriesPlaylists,
    moviePlaylists,
    personalPlaylists,
    clipPlaylists,
    folderPlaylists,
    recentCatalogItems,
    onSelectPlaylist,
    onPlayVideo,
    onSetWatched,
}: VideoHomeScreenProps) {
    return (
        <>
            <VideoContinueWatchingSection
                items={continueWatchingItems.slice(0, 4)}
                onPlayVideo={onPlayVideo}
                onSetWatched={onSetWatched}
                action={<VideoSectionActionLink to={getVideoRoute('continue')} />}
            />
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_SERIES"
                descriptionKey="VIDEO_SECTION_SERIES_DESCRIPTION"
                emptyKey="VIDEO_SECTION_SERIES_EMPTY"
                playlists={seriesPlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute('series')} />}
            />
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_MOVIES"
                descriptionKey="VIDEO_SECTION_MOVIES_DESCRIPTION"
                emptyKey="VIDEO_SECTION_MOVIES_EMPTY"
                playlists={moviePlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute('movies')} />}
            />
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_PERSONAL"
                descriptionKey="VIDEO_SECTION_PERSONAL_DESCRIPTION"
                emptyKey="VIDEO_SECTION_PERSONAL_EMPTY"
                playlists={personalPlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute('personal')} />}
            />
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_CLIPS"
                descriptionKey="VIDEO_SECTION_CLIPS_DESCRIPTION"
                emptyKey="VIDEO_SECTION_CLIPS_EMPTY"
                playlists={clipPlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute('clips')} />}
            />
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_FOLDERS"
                descriptionKey="VIDEO_SECTION_FOLDERS_DESCRIPTION"
                emptyKey="VIDEO_SECTION_FOLDERS_EMPTY"
                playlists={folderPlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute('folders')} />}
            />
            <VideoCatalogRail
                titleKey="VIDEO_HOME_RECENT"
                descriptionKey="VIDEO_HOME_RECENT_DESCRIPTION"
                items={recentCatalogItems}
                onPlayVideo={onPlayVideo}
            />
        </>
    );
}
