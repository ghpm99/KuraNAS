import ActionBar from '@/components/actionBar';
import FileContent from '@/features/files/fileContent';
import FileDetails from '@/features/files/fileDetails';
import FileDetailsProvider from '@/features/files/fileDetails/fileDetailsProvider';
import useFileDetails from '@/features/files/fileDetails/useFileDetails';
import FileSelectionToolbar from '@/features/files/selection/fileSelectionToolbar';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';
import FilesSortControl from '@/features/files/filesSortControl/filesSortControl';
import useFile from '@/features/files/providers/fileProvider/fileContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import FolderTree from '@/components/layout/Sidebar/components/folderTree';
import Tabs from '@/components/tabs';
import {
    Button,
    Drawer,
    IconButton,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
    useMediaQuery,
} from '@mui/material';
import { LayoutGrid, List, PanelLeft, Search } from 'lucide-react';
import { useRef, useState } from 'react';
import UploadDropZone from '@/features/files/upload/uploadDropZone';
import FindByDiskPathDialog from '@/features/files/findByDiskPath/findByDiskPathDialog';
import { FileType } from '@/utils';
import { useNavigate } from 'react-router-dom';
import { buildFilesUrl } from '@/features/files/providers/fileProvider/fileProviderUtils';
import FileSearchBar from '@/features/files/search/FileSearchBar';
import FileSearchResultsHeader from '@/features/files/search/FileSearchResultsHeader';
import useFileSearchQuery from '@/features/files/search/useFileSearchQuery';
import useFileSearchResults from '@/features/files/search/useFileSearchResults';
import FilesBreadcrumb from './FilesBreadcrumb';
import useFilesExplorerScreen from './useFilesExplorerScreen';
import styles from './FilesExplorerScreen.module.css';

const phoneMediaQuery = '(max-width:640px)';

const parentFolderPath = (parentPath: string | undefined): string =>
    !parentPath || parentPath === '/' ? '' : parentPath;

const FilesExplorerScreenContent = () => {
    const { t } = useI18n();
    const {
        breadcrumbSegments,
        closeMobileTree,
        contextLabel,
        itemCountLabel,
        mobileTreeOpen,
        openMobileTree,
        selectedItem,
        setViewMode,
        viewMode,
    } = useFilesExplorerScreen();
    const { filesSort, setFilesSort, handleSelectItem } = useFile();
    const { hasSelection } = useFileSelectionContext();
    const { explicitTarget, closeDetails } = useFileDetails();
    const isPhone = useMediaQuery(phoneMediaQuery);
    const [isFindByDiskPathOpen, setIsFindByDiskPathOpen] = useState(false);
    const navigate = useNavigate();
    const searchInputRef = useRef<HTMLInputElement | null>(null);
    const [isSearchRecursive, setIsSearchRecursive] = useState(true);
    const { inputValue, setInputValue, activeQuery, clearQuery } = useFileSearchQuery();
    const searchFolderId = selectedItem?.type === FileType.Directory ? selectedItem.id : undefined;
    const searchResults = useFileSearchResults({
        query: activeQuery,
        parentId: searchFolderId,
        isRecursive: isSearchRecursive,
    });
    const isSearchActive = activeQuery !== '';
    const isFileSelected = selectedItem?.type === FileType.File;
    const openedFileDetailsTarget = isFileSelected && !isPhone ? selectedItem : null;
    const detailsTarget = explicitTarget ?? openedFileDetailsTarget;
    const isDetailsInSideColumn = detailsTarget !== null && !isPhone;
    const isDetailsInDrawer = explicitTarget !== null && isPhone;
    const workspaceClassName = isDetailsInSideColumn
        ? `${styles.workspace} ${styles.workspaceWithPreview}`
        : styles.workspace;

    const goToParentFolder = selectedItem
        ? () => navigate(buildFilesUrl(parentFolderPath(selectedItem.parent_path)))
        : undefined;
    const focusSearchInput = () => searchInputRef.current?.focus();

    const closeDetailsPanel = () => {
        const isShowingOpenedFile =
            openedFileDetailsTarget !== null &&
            (explicitTarget === null || explicitTarget.id === openedFileDetailsTarget.id);
        closeDetails();
        if (isShowingOpenedFile) handleSelectItem(null);
    };

    return (
        <PageContainer>
            <PageHeader title={t('FILES_PAGE_TITLE')} subtitle={t('FILES_PAGE_DESCRIPTION')} />

            <div className={workspaceClassName}>
                <div className={styles.mainColumn}>
                    <section className={styles.panel}>
                        <div className={styles.contextHeader}>
                            <div>
                                <p className={styles.contextTitle}>{t('FILES_CURRENT_LOCATION')}</p>
                                <FilesBreadcrumb
                                    segments={breadcrumbSegments}
                                    onNavigate={(segment) =>
                                        navigate(buildFilesUrl(segment.path ?? ''))
                                    }
                                />
                            </div>

                            <div className={styles.contextActions}>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    startIcon={<PanelLeft size={16} />}
                                    onClick={openMobileTree}
                                    className={styles.treeButton}
                                >
                                    {t('FILES_OPEN_TREE')}
                                </Button>
                                <Tooltip title={t('FILES_FIND_BY_DISK_PATH')}>
                                    <IconButton
                                        size="small"
                                        aria-label={t('FILES_FIND_BY_DISK_PATH')}
                                        onClick={() => setIsFindByDiskPathOpen(true)}
                                    >
                                        <Search size={16} />
                                    </IconButton>
                                </Tooltip>
                                <FilesSortControl sort={filesSort} onChange={setFilesSort} />
                                <ToggleButtonGroup
                                    size="small"
                                    value={viewMode}
                                    exclusive
                                    onChange={(_, nextViewMode) => {
                                        if (nextViewMode) {
                                            setViewMode(nextViewMode);
                                        }
                                    }}
                                    aria-label={t('FILES_VIEW_SWITCH')}
                                >
                                    <ToggleButton value="grid" aria-label={t('FILES_VIEW_GRID')}>
                                        <LayoutGrid size={16} />
                                        <span>{t('FILES_VIEW_GRID')}</span>
                                    </ToggleButton>
                                    <ToggleButton value="list" aria-label={t('FILES_VIEW_LIST')}>
                                        <List size={16} />
                                        <span>{t('FILES_VIEW_LIST')}</span>
                                    </ToggleButton>
                                </ToggleButtonGroup>
                            </div>
                        </div>

                        <FileSearchBar
                            value={inputValue}
                            onChange={setInputValue}
                            onClear={clearQuery}
                            isFolderScope={searchFolderId !== undefined}
                            isRecursive={isSearchRecursive}
                            onRecursiveChange={setIsSearchRecursive}
                            inputRef={searchInputRef}
                        />

                        <div className={styles.contextMeta}>
                            <span>{contextLabel}</span>
                            <span>{itemCountLabel}</span>
                            {selectedItem ? <span>{selectedItem.name}</span> : null}
                        </div>
                    </section>

                    <section className={`${styles.panel} ${styles.toolbarCard}`}>
                        {hasSelection ? <FileSelectionToolbar /> : <ActionBar />}
                    </section>

                    {!isFileSelected && !isSearchActive ? (
                        <section className={`${styles.panel} ${styles.tabsCard}`}>
                            <Tabs />
                        </section>
                    ) : null}

                    <section className={`${styles.panel} ${styles.contentCard}`}>
                        {isSearchActive ? (
                            <>
                                <FileSearchResultsHeader
                                    query={activeQuery}
                                    resultCount={searchResults.items.length}
                                    hasMoreResults={searchResults.hasNextPage}
                                    onClear={clearQuery}
                                />
                                <FileContent
                                    showHeading={false}
                                    viewMode={viewMode}
                                    searchListing={searchResults}
                                    emptyStateMessage={t('FILES_SEARCH_EMPTY')}
                                    onGoToParent={goToParentFolder}
                                    onFocusSearch={focusSearchInput}
                                />
                            </>
                        ) : (
                            <UploadDropZone>
                                <FileContent
                                    showHeading={false}
                                    viewMode={viewMode}
                                    onGoToParent={goToParentFolder}
                                    onFocusSearch={focusSearchInput}
                                />
                            </UploadDropZone>
                        )}
                    </section>
                </div>

                {isDetailsInSideColumn ? (
                    <aside className={styles.previewColumn}>
                        <section className={`${styles.panel} ${styles.previewCard}`}>
                            <FileDetails file={detailsTarget} onClose={closeDetailsPanel} />
                        </section>
                    </aside>
                ) : null}
            </div>

            <FindByDiskPathDialog
                open={isFindByDiskPathOpen}
                onClose={() => setIsFindByDiskPathOpen(false)}
                onFileFound={handleSelectItem}
            />

            <Drawer anchor="bottom" open={isDetailsInDrawer} onClose={closeDetailsPanel}>
                <div className={styles.detailsDrawerContent}>
                    {explicitTarget ? (
                        <FileDetails file={explicitTarget} onClose={closeDetailsPanel} />
                    ) : null}
                </div>
            </Drawer>

            <Drawer anchor="left" open={mobileTreeOpen} onClose={closeMobileTree}>
                <div className={styles.drawerContent}>
                    <p className={styles.drawerTitle}>{t('FILES_OPEN_TREE')}</p>
                    <FolderTree />
                </div>
            </Drawer>
        </PageContainer>
    );
};

const FilesExplorerScreen = () => (
    <FileDetailsProvider>
        <FilesExplorerScreenContent />
    </FileDetailsProvider>
);

export default FilesExplorerScreen;
