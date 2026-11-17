import { CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
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

    const retryFolders = useRetryFailedPage(folderQuery);
    const retryVideos = useRetryFailedPage(videoQuery);

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
        return (
            <ListFeedback
                isEmpty={folders.length === 0}
                isError={folderQuery.isError}
                isNextPageError={folderQuery.isFetchNextPageError}
                errorTitle={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_SECTION_FOLDERS') })}
                errorMessage={extractBackendErrorMessage(folderQuery.error)}
                onRetry={retryFolders}
                hasNextPage={folderQuery.hasNextPage}
                isFetchingNextPage={folderQuery.isFetchingNextPage}
                onLoadMore={() => void folderQuery.fetchNextPage()}
                emptyState={null}
            >
                <VideoFolderCards folders={folders} onSelectFolder={selectFolder} />
            </ListFeedback>
        );
    };

    const renderVideos = () => {
        if (selectedFolder === '') {
            return null;
        }
        return (
            <ListFeedback
                isEmpty={videos.length === 0}
                isError={videoQuery.isError}
                isNextPageError={videoQuery.isFetchNextPageError}
                errorTitle={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_ALL') })}
                errorMessage={extractBackendErrorMessage(videoQuery.error)}
                onRetry={retryVideos}
                hasNextPage={videoQuery.hasNextPage}
                isFetchingNextPage={videoQuery.isFetchingNextPage}
                onLoadMore={() => void videoQuery.fetchNextPage()}
                emptyState={null}
            >
                <VideoFolderVideoList videos={videos} onPlayVideo={onPlayVideo} />
            </ListFeedback>
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
