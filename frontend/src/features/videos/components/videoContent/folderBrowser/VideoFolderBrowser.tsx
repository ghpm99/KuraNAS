import { CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import ErrorState from '@/components/errorState/errorState';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import VideoFolderBreadcrumb from './VideoFolderBreadcrumb';
import VideoFolderCards from './VideoFolderCards';
import VideoFolderVideoList from './VideoFolderVideoList';
import { useVideoFolderBrowser } from './useVideoFolderBrowser';
import styles from './VideoFolderBrowser.module.css';

type VideoFolderBrowserProps = {
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

export default function VideoFolderBrowser({ onPlayVideo }: VideoFolderBrowserProps) {
    const { t } = useI18n();
    const { selectedFolder, selectFolder, folderQuery, videoQuery, folders, videos } =
        useVideoFolderBrowser();

    const isRootEmpty =
        selectedFolder === '' && folderQuery.status === 'success' && folders.length === 0;
    const isFolderEmpty =
        selectedFolder !== '' &&
        folderQuery.status === 'success' &&
        videoQuery.status === 'success' &&
        folders.length === 0 &&
        videos.length === 0;

    const renderFolders = () => {
        if (folderQuery.status === 'pending') {
            return <CircularProgress size={28} aria-label={t('LOADING')} />;
        }
        if (folderQuery.status === 'error') {
            return (
                <ErrorState
                    title={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_SECTION_FOLDERS') })}
                    backendMessage={extractBackendErrorMessage(folderQuery.error)}
                    onRetry={() => void folderQuery.refetch()}
                />
            );
        }
        return (
            <VideoFolderCards
                folders={folders}
                hasNextPage={folderQuery.hasNextPage}
                isFetchingNextPage={folderQuery.isFetchingNextPage}
                fetchNextPage={() => void folderQuery.fetchNextPage()}
                onSelectFolder={selectFolder}
            />
        );
    };

    const renderVideos = () => {
        if (selectedFolder === '') {
            return null;
        }
        if (videoQuery.status === 'error') {
            return (
                <ErrorState
                    title={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_ALL') })}
                    backendMessage={extractBackendErrorMessage(videoQuery.error)}
                    onRetry={() => void videoQuery.refetch()}
                />
            );
        }
        return (
            <VideoFolderVideoList
                videos={videos}
                hasNextPage={videoQuery.hasNextPage}
                isFetchingNextPage={videoQuery.isFetchingNextPage}
                fetchNextPage={() => void videoQuery.fetchNextPage()}
                onPlayVideo={onPlayVideo}
            />
        );
    };

    return (
        <section className={styles.browser}>
            <VideoFolderBreadcrumb selectedFolder={selectedFolder} onSelectFolder={selectFolder} />
            {renderFolders()}
            {renderVideos()}
            {isRootEmpty || isFolderEmpty ? (
                <p className={styles.empty}>{t('VIDEO_FOLDERS_EMPTY')}</p>
            ) : null}
        </section>
    );
}
