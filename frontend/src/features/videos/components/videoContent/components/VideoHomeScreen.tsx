import { getVideoRoute } from '@/app/routes';
import {
    type VideoCatalogItemDto,
    type VideoContinueItemDto,
    type VideoPlaylistDto,
} from '@/service/videoPlayback';
import type { VideoQueryFailure } from '@/features/videos/providers/videoContentProvider/videoQueryFailure';
import VideoCatalogRail from './VideoCatalogRail';
import VideoContinueWatchingSection from './VideoContinueWatchingSection';
import VideoSectionError from './VideoSectionError';
import VideoSectionSkeleton from './VideoSectionSkeleton';
import VideoSectionPlaylistGrid, { VideoSectionActionLink } from './VideoSectionPlaylistGrid';

type VideoHomeScreenProps = {
    continueWatchingItems: VideoContinueItemDto[];
    seriesPlaylists: VideoPlaylistDto[];
    moviePlaylists: VideoPlaylistDto[];
    personalPlaylists: VideoPlaylistDto[];
    clipPlaylists: VideoPlaylistDto[];
    folderPlaylists: VideoPlaylistDto[];
    recentCatalogItems: VideoCatalogItemDto[];
    isLoadingContinueWatching?: boolean;
    isLoadingPlaylists?: boolean;
    isLoadingHomeCatalog?: boolean;
    playlistsFailure?: VideoQueryFailure | null;
    continueWatchingFailure?: VideoQueryFailure | null;
    homeCatalogFailure?: VideoQueryFailure | null;
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
    isLoadingContinueWatching = false,
    isLoadingPlaylists = false,
    isLoadingHomeCatalog = false,
    playlistsFailure = null,
    continueWatchingFailure = null,
    homeCatalogFailure = null,
    onSelectPlaylist,
    onPlayVideo,
    onSetWatched,
}: VideoHomeScreenProps) {
    const renderPlaylistSection = (
        sectionName: string,
        titleKey: string,
        playlists: VideoPlaylistDto[],
        route: Parameters<typeof getVideoRoute>[0]
    ) => {
        if (isLoadingPlaylists) {
            return <VideoSectionSkeleton key={sectionName} />;
        }
        if (playlistsFailure) {
            return (
                <VideoSectionError
                    key={sectionName}
                    sectionTitleKey={`VIDEO_SECTION_${titleKey}`}
                    failure={playlistsFailure}
                />
            );
        }
        return (
            <VideoSectionPlaylistGrid
                key={sectionName}
                titleKey={`VIDEO_SECTION_${titleKey}`}
                descriptionKey={`VIDEO_SECTION_${titleKey}_DESCRIPTION`}
                emptyKey={`VIDEO_SECTION_${titleKey}_EMPTY`}
                playlists={playlists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                action={<VideoSectionActionLink to={getVideoRoute(route)} />}
            />
        );
    };

    return (
        <>
            {isLoadingContinueWatching ? (
                <VideoSectionSkeleton layout="catalogRail" />
            ) : continueWatchingFailure ? (
                <VideoSectionError
                    sectionTitleKey="VIDEO_SECTION_CONTINUE"
                    failure={continueWatchingFailure}
                />
            ) : (
                <VideoContinueWatchingSection
                    items={continueWatchingItems.slice(0, 4)}
                    onPlayVideo={onPlayVideo}
                    onSetWatched={onSetWatched}
                    action={<VideoSectionActionLink to={getVideoRoute('continue')} />}
                />
            )}
            {renderPlaylistSection('series', 'SERIES', seriesPlaylists, 'series')}
            {renderPlaylistSection('movies', 'MOVIES', moviePlaylists, 'movies')}
            {renderPlaylistSection('personal', 'PERSONAL', personalPlaylists, 'personal')}
            {renderPlaylistSection('clips', 'CLIPS', clipPlaylists, 'clips')}
            {renderPlaylistSection('folders', 'FOLDERS', folderPlaylists, 'folders')}
            {isLoadingHomeCatalog ? (
                <VideoSectionSkeleton layout="catalogRail" />
            ) : homeCatalogFailure ? (
                <VideoSectionError
                    sectionTitleKey="VIDEO_HOME_RECENT"
                    failure={homeCatalogFailure}
                />
            ) : (
                <VideoCatalogRail
                    titleKey="VIDEO_HOME_RECENT"
                    descriptionKey="VIDEO_HOME_RECENT_DESCRIPTION"
                    items={recentCatalogItems}
                    onPlayVideo={onPlayVideo}
                />
            )}
        </>
    );
}
