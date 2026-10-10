import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { VideoLibraryFolderDto } from '@/service/videoPlayback';
import styles from './VideoFolderBrowser.module.css';

type VideoFolderCardsProps = {
    folders: VideoLibraryFolderDto[];
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    onSelectFolder: (folderPath: string) => void;
};

const thumbnailBaseUrl = () => `${getApiV1BaseUrl()}/files/video-thumbnail`;

export default function VideoFolderCards({
    folders,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    onSelectFolder,
}: VideoFolderCardsProps) {
    const { t } = useI18n();

    return (
        <>
            <div className={styles.folderGrid}>
                {folders.map((folder) => (
                    <button
                        type="button"
                        key={folder.path}
                        className={styles.folderCard}
                        aria-label={t('VIDEO_FOLDER_OPEN', { name: folder.name })}
                        onClick={() => onSelectFolder(folder.path)}
                    >
                        <div className={styles.folderCover}>
                            {folder.cover_file_id ? (
                                <img
                                    loading="lazy"
                                    src={`${thumbnailBaseUrl()}/${folder.cover_file_id}?width=320&height=180`}
                                    alt={folder.name}
                                />
                            ) : (
                                <div className={styles.folderPlaceholder}>
                                    {folder.name.slice(0, 1).toUpperCase()}
                                </div>
                            )}
                        </div>
                        <div className={styles.folderBody}>
                            <h3>{folder.name}</h3>
                            <span className={styles.folderCount}>
                                {t('VIDEO_FOLDER_VIDEOS_COUNT', {
                                    count: String(folder.video_count),
                                })}
                            </span>
                        </div>
                    </button>
                ))}
            </div>
            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
            />
        </>
    );
}
