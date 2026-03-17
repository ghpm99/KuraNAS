import { FileType, formatSize } from '@/utils';
import FileCard from '../fileCard';
import useI18n from '../i18n/provider/i18nContext';
import useFile, { FileData } from '../providers/fileProvider/fileContext';
import FileViewer from './components/fileViewer/fileViewer';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useMediaOpener from '@/components/hooks/useMediaOpener/useMediaOpener';
import styles from './fileContent.module.css';

interface FileContentProps {
    showHeading?: boolean;
    viewMode?: 'grid' | 'list';
    items?: FileData[];
    title?: string;
    emptyStateMessage?: string;
}

const FileContent = ({
    showHeading = true,
    viewMode = 'grid',
    items,
    title,
    emptyStateMessage,
}: FileContentProps) => {
    const { status, handleSelectItem, selectedItem, files, handleStarredItem, fileListFilter } =
        useFile();
    const { t } = useI18n();
    const { openMediaItem } = useMediaOpener();
    const currentListTitle =
        fileListFilter === 'starred'
            ? t('STARRED_FILES')
            : fileListFilter === 'recent'
              ? t('RECENT_FILES')
              : t('FILES');

    if (status === 'pending') {
        return <div className={styles.fileContent}>{t('LOADING')}</div>;
    }
    if (status === 'error') {
        return <div className={styles.fileContent}>{t('ERROR_LOADING_FILES')}</div>;
    }

    const fileMetadata = (file: FileData): string => {
        if (file.type === FileType.File) {
            const format = file.format ? `${file.format} - ` : '';
            const fileSize = formatSize(file.size);

            return `${format}${fileSize}`;
        }
        const directoryContentCount = file.directory_content_count;
        const countText = directoryContentCount > 1 ? t('ITENS') : t('ITEM');
        return `${t('FOLDER')} - ${directoryContentCount} ${countText}`;
    };

    const thumbnailUrl = (id: number) => `${getApiV1BaseUrl()}/files/thumbnail/${id}`;

    const handleOpenItem = (file: FileData) => {
        if (!openMediaItem(file)) {
            handleSelectItem(file);
        }
    };

    const renderCollection = (collectionTitle: string, collectionItems: FileData[]) => {
        if (collectionItems.length === 0) {
            return (
                <div className={styles.fileContent}>
                    {showHeading ? <h1 className={styles.title}>{collectionTitle}</h1> : null}
                    <div className={styles.emptyState}>
                        {emptyStateMessage ?? t('EMPTY_FILE_LIST')}
                    </div>
                </div>
            );
        }

        return (
            <div className={styles.fileContent}>
                {showHeading ? <h1 className={styles.title}>{collectionTitle}</h1> : null}
                {viewMode === 'list' ? (
                    <div className={styles.fileList}>
                        {collectionItems.map((file) => (
                            <div key={file.id} className={styles.listRow}>
                                <button
                                    type="button"
                                    className={styles.listButton}
                                    onClick={() => handleOpenItem(file)}
                                    aria-label={file.name}
                                >
                                    <img
                                        src={thumbnailUrl(file.id)}
                                        alt={file.name}
                                        loading="lazy"
                                        className={styles.listThumbnail}
                                    />
                                    <div className={styles.listContent}>
                                        <span className={styles.listTitle}>{file.name}</span>
                                        <span className={styles.listMetadata}>
                                            {fileMetadata(file)}
                                        </span>
                                    </div>
                                </button>
                                <button
                                    type="button"
                                    className={styles.listStarButton}
                                    onClick={() => handleStarredItem(file.id)}
                                >
                                    {file.starred ? '★' : '☆'}
                                </button>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className={styles.fileGrid}>
                        {collectionItems.map((file) => (
                            <FileCard
                                key={file.id}
                                title={file.name}
                                starred={file.starred}
                                metadata={fileMetadata(file)}
                                thumbnail={thumbnailUrl(file.id)}
                                onClick={() => handleOpenItem(file)}
                                onClickStar={() => handleStarredItem(file.id)}
                            />
                        ))}
                    </div>
                )}
            </div>
        );
    };

    const currentItems =
        items ??
        (!selectedItem
            ? (files ?? [])
            : selectedItem.type === FileType.Directory
              ? (selectedItem.file_children ?? [])
              : []);
    const currentTitle = title ?? (!selectedItem ? currentListTitle : selectedItem.name);

    if (!selectedItem) {
        return renderCollection(currentTitle, currentItems);
    }

    if (selectedItem.type === FileType.Directory) {
        return renderCollection(currentTitle, currentItems);
    }

    return (
        <div className={styles.previewContainer}>
            <FileViewer file={selectedItem} />
        </div>
    );
};

export default FileContent;
