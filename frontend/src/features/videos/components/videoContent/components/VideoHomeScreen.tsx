import { getVideoRoute } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { type VideoCatalogItemDto, type VideoPlaylistDto } from '@/service/videoPlayback';
import VideoCatalogRail from './VideoCatalogRail';
import VideoSectionPlaylistGrid, { VideoSectionActionLink } from './VideoSectionPlaylistGrid';

type VideoHomeScreenProps = {
    continuePlaylists: VideoPlaylistDto[];
    seriesPlaylists: VideoPlaylistDto[];
    moviePlaylists: VideoPlaylistDto[];
    personalPlaylists: VideoPlaylistDto[];
    clipPlaylists: VideoPlaylistDto[];
    folderPlaylists: VideoPlaylistDto[];
    recentCatalogItems: VideoCatalogItemDto[];
    onSelectPlaylist: (playlist: VideoPlaylistDto) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

export default function VideoHomeScreen({
    continuePlaylists,
    seriesPlaylists,
    moviePlaylists,
    personalPlaylists,
    clipPlaylists,
    folderPlaylists,
    recentCatalogItems,
    onSelectPlaylist,
    onPlayVideo,
}: VideoHomeScreenProps) {
    const { t } = useI18n();

    return (
        <>
            <VideoSectionPlaylistGrid
                titleKey="VIDEO_SECTION_CONTINUE"
                descriptionKey="VIDEO_SECTION_CONTINUE_DESCRIPTION"
                emptyKey="VIDEO_NO_RECENT_PLAYLISTS"
                playlists={continuePlaylists.slice(0, 4)}
                onSelectPlaylist={onSelectPlaylist}
                onPlayVideo={onPlayVideo}
                badge={t('VIDEO_CONTINUE_BADGE_RESUME')}
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
