import { useState } from 'react';
import { Button, Drawer, ToggleButton, ToggleButtonGroup, useMediaQuery } from '@mui/material';
import { FileText, Folder, LayoutGrid, List, RefreshCw } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import EmptyState from '@/components/emptyState/emptyState';
import DocumentCard from './DocumentCard';
import DocumentRow from './DocumentRow';
import { viewportMediaQueries } from '@/theme/visualTokens';
import DocumentsFolderPanel from './DocumentsFolderPanel';
import DocumentsFilterBar from './DocumentsFilterBar';
import DocumentsSortControl from './DocumentsSortControl';
import DocumentsListSkeleton from './DocumentsListSkeleton';
import {
    loadDocumentsViewMode,
    saveDocumentsViewMode,
    type DocumentsViewMode,
} from './documentsViewModePreference';
import useDocumentsFilters from './useDocumentsFilters';
import useDocumentsList from './useDocumentsList';
import styles from './DocumentsScreen.module.css';

const DocumentsScreen = () => {
    const { t } = useI18n();
    const {
        filters,
        sortState,
        folderId,
        requestFilters,
        requestSort,
        activeFilterCount,
        setFilters,
        setSort,
        setFolderId,
        resetFilters,
    } = useDocumentsFilters();
    const isNarrowScreen = useMediaQuery(viewportMediaQueries.belowTablet);
    const [isFolderDrawerOpen, setIsFolderDrawerOpen] = useState(false);
    const [viewMode, setViewMode] = useState<DocumentsViewMode>(loadDocumentsViewMode);
    const {
        documents,
        total,
        isLoading,
        isError,
        isNextPageError,
        hasNextPage,
        isFetchingNextPage,
        fetchNextPage,
        errorMessage,
        retry,
        reload,
    } = useDocumentsList(requestFilters, requestSort);

    const changeViewMode = (nextViewMode: DocumentsViewMode | null) => {
        if (!nextViewMode) return;
        setViewMode(nextViewMode);
        saveDocumentsViewMode(nextViewMode);
    };

    const selectFolder = (nextFolderId: number | undefined) => {
        setFolderId(nextFolderId);
        setIsFolderDrawerOpen(false);
    };

    const subtitle =
        total === undefined ? undefined : t('DOCUMENTS_TOTAL', { count: String(total) });
    const isListFullyLoaded = documents.length > 0 && !hasNextPage && !isNextPageError;

    return (
        <PageContainer>
            <PageHeader
                title={t('NAV_DOCUMENTS')}
                subtitle={subtitle}
                actions={
                    <>
                        {isNarrowScreen ? (
                            <Button
                                size="small"
                                variant="outlined"
                                startIcon={<Folder size={16} />}
                                onClick={() => setIsFolderDrawerOpen(true)}
                            >
                                {t('DOCUMENTS_FOLDERS_TITLE')}
                            </Button>
                        ) : null}
                        <DocumentsSortControl sortState={sortState} onChange={setSort} />
                        <ToggleButtonGroup
                            size="small"
                            value={viewMode}
                            exclusive
                            onChange={(_, nextViewMode) => changeViewMode(nextViewMode)}
                            aria-label={t('DOCUMENTS_VIEW_SWITCH')}
                        >
                            <ToggleButton value="grid" aria-label={t('DOCUMENTS_VIEW_GRID')}>
                                <LayoutGrid size={16} />
                                <span>{t('DOCUMENTS_VIEW_GRID')}</span>
                            </ToggleButton>
                            <ToggleButton value="list" aria-label={t('DOCUMENTS_VIEW_LIST')}>
                                <List size={16} />
                                <span>{t('DOCUMENTS_VIEW_LIST')}</span>
                            </ToggleButton>
                        </ToggleButtonGroup>
                        <Button
                            size="small"
                            variant="outlined"
                            startIcon={<RefreshCw size={16} />}
                            onClick={reload}
                        >
                            {t('DOCUMENTS_RELOAD')}
                        </Button>
                    </>
                }
            />
            <div
                className={
                    isNarrowScreen
                        ? `${styles.workspace} ${styles.workspaceNarrow}`
                        : styles.workspace
                }
            >
                {isNarrowScreen ? null : (
                    <aside className={styles.folderColumn}>
                        <DocumentsFolderPanel selectedFolderId={folderId} onSelect={selectFolder} />
                    </aside>
                )}
                <div className={styles.mainColumn}>
                    <DocumentsFilterBar
                        filters={filters}
                        onChange={setFilters}
                        onReset={resetFilters}
                    />
                    {isError && folderId !== undefined ? (
                        <div className={styles.folderError}>
                            <Button
                                size="small"
                                variant="outlined"
                                onClick={() => setFolderId(undefined)}
                            >
                                {t('DOCUMENTS_FOLDERS_ALL')}
                            </Button>
                        </div>
                    ) : null}
                    {isLoading && !isError ? (
                        <DocumentsListSkeleton viewMode={viewMode} />
                    ) : (
                        <ListFeedback
                            isEmpty={documents.length === 0}
                            isError={isError}
                            isNextPageError={isNextPageError}
                            errorTitle={t('DOCUMENTS_LOAD_ERROR')}
                            errorMessage={errorMessage}
                            onRetry={retry}
                            hasNextPage={hasNextPage}
                            isFetchingNextPage={isFetchingNextPage}
                            onLoadMore={fetchNextPage}
                            emptyState={
                                activeFilterCount > 0 ? (
                                    <EmptyState
                                        icon={<FileText size={32} aria-hidden="true" />}
                                        title={t('DOCUMENTS_FILTER_EMPTY_TITLE')}
                                        description={t('DOCUMENTS_FILTER_EMPTY_DESCRIPTION')}
                                        actions={
                                            <Button
                                                size="small"
                                                variant="outlined"
                                                onClick={resetFilters}
                                            >
                                                {t('DOCUMENTS_FILTERS_CLEAR')}
                                            </Button>
                                        }
                                    />
                                ) : (
                                    <EmptyState
                                        icon={<FileText size={32} aria-hidden="true" />}
                                        title={t('DOCUMENTS_EMPTY_TITLE')}
                                        description={t('DOCUMENTS_EMPTY_DESCRIPTION')}
                                    />
                                )
                            }
                        >
                            {viewMode === 'grid' ? (
                                <ul className={styles.grid} aria-label={t('NAV_DOCUMENTS')}>
                                    {documents.map((document) => (
                                        <DocumentCard key={document.id} document={document} />
                                    ))}
                                </ul>
                            ) : (
                                <ul className={styles.list} aria-label={t('NAV_DOCUMENTS')}>
                                    {documents.map((document) => (
                                        <DocumentRow key={document.id} document={document} />
                                    ))}
                                </ul>
                            )}
                            {isListFullyLoaded ? (
                                <p className={styles.listEndTotal}>
                                    {t('DOCUMENTS_TOTAL', { count: String(documents.length) })}
                                </p>
                            ) : null}
                        </ListFeedback>
                    )}
                </div>
            </div>
            <Drawer
                anchor="left"
                open={isNarrowScreen && isFolderDrawerOpen}
                onClose={() => setIsFolderDrawerOpen(false)}
            >
                <div className={styles.drawerContent}>
                    <DocumentsFolderPanel selectedFolderId={folderId} onSelect={selectFolder} />
                </div>
            </Drawer>
        </PageContainer>
    );
};

export default DocumentsScreen;
