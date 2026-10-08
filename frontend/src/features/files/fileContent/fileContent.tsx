import { useState, type MouseEvent } from 'react';
import { FileType } from '@/utils';
import { formatSize } from '@/shared/utils/formatSize';
import FileCard from '../fileCard';
import FileListRow from '../fileListRow';
import FileContextMenu, { type FileContextMenuAnchor } from '../fileContextMenu/fileContextMenu';
import { useFileSelectionContext } from '../selection/fileSelectionContext';
import { resolveListedFiles } from '../selection/listedFiles';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { FileData } from '@/features/files/providers/fileProvider/fileContext';
import FileViewer from './components/fileViewer/fileViewer';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useMediaOpener from '@/components/hooks/useMediaOpener/useMediaOpener';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import styles from './fileContent.module.css';

type ContextMenuState = { file: FileData; anchorPosition: FileContextMenuAnchor };

export type FileSearchListing = {
    items: FileData[];
    status: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
};

interface FileContentProps {
    searchListing?: FileSearchListing;
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
    searchListing,
}: FileContentProps) => {
    const {
        status,
        handleSelectItem,
        selectedItem,
        files,
        handleStarredItem,
        fileListFilter,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useFile();
    const { t } = useI18n();
    const { openMediaItem } = useMediaOpener();
    const fileSelection = useFileSelectionContext();
    const [contextMenuState, setContextMenuState] = useState<ContextMenuState | null>(null);
    const isSelectionEnabled = items === undefined;
    const listingStatus = searchListing?.status ?? status;
    const pagination = searchListing ?? { hasNextPage, isFetchingNextPage, fetchNextPage };
    const currentListTitle =
        fileListFilter === 'starred'
            ? t('STARRED_FILES')
            : fileListFilter === 'recent'
              ? t('RECENT_FILES')
              : t('FILES');

    if (listingStatus === 'pending') {
        return <div className={styles.fileContent}>{t('LOADING')}</div>;
    }
    if (listingStatus === 'error') {
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

    const handleItemClick = (
        file: FileData,
        event: MouseEvent<HTMLElement>,
        orderedFiles: FileData[]
    ) => {
        if (isSelectionEnabled && event.shiftKey) {
            fileSelection.selectRange(file, orderedFiles);
            return;
        }
        const isTogglingSelection =
            isSelectionEnabled &&
            (event.ctrlKey || event.metaKey || fileSelection.hasSelection);
        if (isTogglingSelection) {
            fileSelection.toggle(file);
            return;
        }
        handleOpenItem(file);
    };

    const handleToggleSelectionClick = (
        file: FileData,
        event: MouseEvent<HTMLElement>,
        orderedFiles: FileData[]
    ) => {
        if (event.shiftKey) {
            fileSelection.selectRange(file, orderedFiles);
            return;
        }
        fileSelection.toggle(file);
    };

    const openContextMenu = (file: FileData, event: MouseEvent<HTMLElement>) => {
        event.preventDefault();
        setContextMenuState({
            file,
            anchorPosition: { top: event.clientY, left: event.clientX },
        });
    };

    const openContextMenuFromButton = (file: FileData, event: MouseEvent<HTMLElement>) => {
        event.stopPropagation();
        const buttonBounds = event.currentTarget.getBoundingClientRect();
        setContextMenuState({
            file,
            anchorPosition: { top: buttonBounds.bottom, left: buttonBounds.left },
        });
    };

    const buildInteractionProps = (file: FileData, orderedFiles: FileData[]) => ({
        isSelected: isSelectionEnabled && fileSelection.isSelected(file.id),
        isSelectionActive: isSelectionEnabled && fileSelection.hasSelection,
        onToggleSelection: isSelectionEnabled
            ? (event: MouseEvent<HTMLElement>) =>
                  handleToggleSelectionClick(file, event, orderedFiles)
            : undefined,
        onOpenMenu: (event: MouseEvent<HTMLElement>) => openContextMenuFromButton(file, event),
        onContextMenu: (event: MouseEvent<HTMLElement>) => openContextMenu(file, event),
    });

    const contextMenuTargetFiles = (): FileData[] => {
        if (!contextMenuState) return [];
        const { file } = contextMenuState;
        const isPartOfMultiSelection =
            isSelectionEnabled && fileSelection.isSelected(file.id) && fileSelection.selectedCount > 1;
        return isPartOfMultiSelection ? fileSelection.selectedFiles : [file];
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
                            <FileListRow
                                key={file.id}
                                title={file.name}
                                starred={file.starred}
                                isCold={file.tier === 'cold'}
                                metadata={fileMetadata(file)}
                                secondaryText={searchListing ? file.parent_path : undefined}
                                thumbnail={thumbnailUrl(file.id)}
                                onClick={(event) => handleItemClick(file, event, collectionItems)}
                                onClickStar={() => handleStarredItem(file.id)}
                                {...buildInteractionProps(file, collectionItems)}
                            />
                        ))}
                    </div>
                ) : (
                    <div className={styles.fileGrid}>
                        {collectionItems.map((file) => (
                            <FileCard
                                key={file.id}
                                title={file.name}
                                starred={file.starred}
                                isCold={file.tier === 'cold'}
                                metadata={fileMetadata(file)}
                                secondaryText={searchListing ? file.parent_path : undefined}
                                thumbnail={thumbnailUrl(file.id)}
                                onClick={(event) => handleItemClick(file, event, collectionItems)}
                                onClickStar={() => handleStarredItem(file.id)}
                                {...buildInteractionProps(file, collectionItems)}
                            />
                        ))}
                    </div>
                )}
                <FileContextMenu
                    anchorPosition={contextMenuState?.anchorPosition ?? null}
                    targetFiles={contextMenuTargetFiles()}
                    onClose={() => setContextMenuState(null)}
                    onOpenFile={handleOpenItem}
                />
                {items ? null : (
                    <LoadMoreSentinel
                        hasNextPage={pagination.hasNextPage}
                        isFetchingNextPage={pagination.isFetchingNextPage}
                        fetchNextPage={pagination.fetchNextPage}
                    />
                )}
            </div>
        );
    };

    if (searchListing) {
        return renderCollection(title ?? '', searchListing.items);
    }

    const currentItems = items ?? resolveListedFiles(selectedItem, files);
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
