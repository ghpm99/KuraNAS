import { useState, type MouseEvent } from 'react';
import { FileType } from '@/utils';
import { formatSize } from '@/shared/utils/formatSize';
import FileCard from '../fileCard';
import FileListRow from '../fileListRow';
import FileListHeader from '../fileListHeader/fileListHeader';
import { buildFilesUrl } from '@/features/files/providers/fileProvider/fileProviderUtils';
import { formatModifiedDate, formatSizeColumn, formatTypeColumn } from './fileListFormatting';
import { isNewTabClick } from './fileLinkNavigation';
import FileContextMenu, { type FileContextMenuAnchor } from '../fileContextMenu/fileContextMenu';
import { useFileSelectionContext } from '../selection/fileSelectionContext';
import { resolveListedFiles } from '../selection/listedFiles';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { FileData } from '@/features/files/providers/fileProvider/fileContext';
import FileViewer from './components/fileViewer/fileViewer';
import FileViewerNavigation from './components/fileViewer/fileViewerNavigation';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useMediaOpener from '@/components/hooks/useMediaOpener/useMediaOpener';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useFileActionFlow from '../fileActions/useFileActionFlow';
import { useRegisterPageShortcuts } from '@/components/shortcuts/shortcutRegistry';
import { fileShortcutDefinitions } from '../shortcuts/fileShortcutDefinitions';
import { countRenderedGridColumns } from '../shortcuts/fileItemFocus';
import useFileBrowserShortcuts from '../shortcuts/useFileBrowserShortcuts';
import FileCollectionEmptyState from './fileCollectionEmptyState';
import FileListingSkeleton from './fileListingSkeleton';
import styles from './fileContent.module.css';

type ContextMenuState = { file: FileData; anchorPosition: FileContextMenuAnchor };

export type FileSearchListing = {
    items: FileData[];
    status: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    errorMessage?: string;
    retry?: () => void;
};

interface FileContentProps {
    searchListing?: FileSearchListing;
    showHeading?: boolean;
    viewMode?: 'grid' | 'list';
    items?: FileData[];
    title?: string;
    emptyStateMessage?: string;
    onGoToParent?: () => void;
    onFocusSearch?: () => void;
}

const FileContent = ({
    showHeading = true,
    viewMode = 'grid',
    items,
    title,
    emptyStateMessage,
    searchListing,
    onGoToParent,
    onFocusSearch,
}: FileContentProps) => {
    const {
        status,
        listingErrorMessage,
        retryListing,
        handleSelectItem,
        selectedItem,
        files,
        handleStarredItem,
        fileListFilter,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        filesSort,
        setFilesSort,
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

    const { startAction, dialogs: actionDialogs } = useFileActionFlow();
    const isViewingFile = selectedItem?.type === FileType.File;
    const listedFiles = searchListing?.items ?? items ?? resolveListedFiles(selectedItem, files);

    const openListedFile = (file: FileData) => {
        if (!openMediaItem(file, listedFiles)) {
            handleSelectItem(file);
        }
    };

    const areFileShortcutsEnabled = isSelectionEnabled && !isViewingFile;
    useRegisterPageShortcuts(fileShortcutDefinitions, areFileShortcutsEnabled);

    const { tabStopFileId, focusFile } = useFileBrowserShortcuts({
        isEnabled: areFileShortcutsEnabled,
        files: listedFiles,
        selection: fileSelection,
        getColumnCount: () => (viewMode === 'list' ? 1 : countRenderedGridColumns()),
        onOpenFile: openListedFile,
        onDeleteFiles: (filesToDelete) => startAction('delete', filesToDelete),
        onRenameFile: (fileToRename) => startAction('rename', [fileToRename]),
        onGoToParent,
        onFocusSearch,
    });

    if (listingStatus === 'pending') {
        return <FileListingSkeleton viewMode={viewMode} />;
    }
    if (listingStatus === 'error') {
        return (
            <div className={styles.fileContent}>
                <ErrorState
                    title={t('FILES_LISTING_ERROR_TITLE')}
                    backendMessage={searchListing ? searchListing.errorMessage : listingErrorMessage}
                    onRetry={searchListing ? searchListing.retry : retryListing}
                />
            </div>
        );
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

    const handleOpenItem = (file: FileData, orderedFiles: FileData[] = []) => {
        if (!openMediaItem(file, orderedFiles)) {
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
            isSelectionEnabled && (event.ctrlKey || event.metaKey || fileSelection.hasSelection);
        if (isTogglingSelection) {
            fileSelection.toggle(file);
            return;
        }
        handleOpenItem(file, orderedFiles);
    };

    const handleLinkClick = (
        file: FileData,
        event: MouseEvent<HTMLElement>,
        orderedFiles: FileData[]
    ) => {
        const isSelectionModifierClick = isSelectionEnabled && fileSelection.hasSelection;
        if (isNewTabClick(event) && !isSelectionModifierClick) return;
        event.preventDefault();
        handleItemClick(file, event, orderedFiles);
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
        fileId: file.id,
        isTabStop: file.id === tabStopFileId,
        onFocusItem: () => focusFile(file.id),
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
            isSelectionEnabled &&
            fileSelection.isSelected(file.id) &&
            fileSelection.selectedCount > 1;
        return isPartOfMultiSelection ? fileSelection.selectedFiles : [file];
    };

    const isOpenFolderListing =
        selectedItem?.type === FileType.Directory && items === undefined && !searchListing;

    const renderCollection = (collectionTitle: string, collectionItems: FileData[]) => {
        if (collectionItems.length === 0) {
            return (
                <div className={styles.fileContent}>
                    {showHeading ? <h1 className={styles.title}>{collectionTitle}</h1> : null}
                    <FileCollectionEmptyState
                        customMessage={emptyStateMessage}
                        isFolderOpen={isOpenFolderListing}
                        fileListFilter={fileListFilter}
                    />
                    {actionDialogs}
                </div>
            );
        }

        return (
            <div className={styles.fileContent}>
                {showHeading ? <h1 className={styles.title}>{collectionTitle}</h1> : null}
                {viewMode === 'list' ? (
                    <div className={styles.fileList} role="table" aria-label={collectionTitle}>
                        {isSelectionEnabled && !searchListing ? (
                            <FileListHeader sort={filesSort} onSortChange={setFilesSort} />
                        ) : null}
                        {collectionItems.map((file) => (
                            <FileListRow
                                key={file.id}
                                title={file.name}
                                href={buildFilesUrl(file.path)}
                                starred={file.starred}
                                isCold={file.tier === 'cold'}
                                sizeText={formatSizeColumn(file, t)}
                                modifiedText={formatModifiedDate(file.updated_at)}
                                typeText={formatTypeColumn(file, t)}
                                secondaryText={searchListing ? file.parent_path : undefined}
                                thumbnail={thumbnailUrl(file.id)}
                                onClick={(event) => handleLinkClick(file, event, collectionItems)}
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
                                href={buildFilesUrl(file.path)}
                                starred={file.starred}
                                isCold={file.tier === 'cold'}
                                metadata={fileMetadata(file)}
                                secondaryText={searchListing ? file.parent_path : undefined}
                                thumbnail={thumbnailUrl(file.id)}
                                onClick={(event) => handleLinkClick(file, event, collectionItems)}
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
                    onOpenFile={(file) => handleOpenItem(file, collectionItems)}
                />
                {actionDialogs}
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
            <FileViewerNavigation file={selectedItem} />
            <FileViewer file={selectedItem} />
        </div>
    );
};

export default FileContent;
