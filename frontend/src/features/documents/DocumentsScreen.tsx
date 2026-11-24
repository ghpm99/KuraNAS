import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, Drawer, ToggleButton, ToggleButtonGroup, useMediaQuery } from '@mui/material';
import { FileText, Folder, LayoutGrid, List, RefreshCw } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import EmptyState from '@/components/emptyState/emptyState';
import { buildDocumentViewerUrl } from '@/app/routes';
import FileDetails from '@/features/files/fileDetails';
import FileDetailsProvider from '@/features/files/fileDetails/fileDetailsProvider';
import useFileDetails from '@/features/files/fileDetails/useFileDetails';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';
import FileSearchBar from '@/features/files/search/FileSearchBar';
import useFileSearchQuery from '@/features/files/search/useFileSearchQuery';
import DocumentCard from './DocumentCard';
import DocumentItemMenu from './DocumentItemMenu';
import DocumentsBatchBar from './DocumentsBatchBar';
import { documentToFileData, searchResultToFileData } from './documentFileAdapter';
import useDocumentItemInteractions from './useDocumentItemInteractions';
import DocumentRow from './DocumentRow';
import { viewportMediaQueries } from '@/theme/visualTokens';
import DocumentsUploadActions from './DocumentsUploadActions';
import DocumentsFolderPanel from './DocumentsFolderPanel';
import DocumentsFilterBar from './DocumentsFilterBar';
import DocumentsSortControl from './DocumentsSortControl';
import DocumentsSearchResults from './DocumentsSearchResults';
import DocumentsListSkeleton from './DocumentsListSkeleton';
import {
    loadDocumentsViewMode,
    saveDocumentsViewMode,
    type DocumentsViewMode,
} from './documentsViewModePreference';
import useDocumentsFilters from './useDocumentsFilters';
import useDocumentsList from './useDocumentsList';
import useDocumentsSearch from './useDocumentsSearch';
import styles from './DocumentsScreen.module.css';

const DocumentsScreenContent = () => {
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
    const navigate = useNavigate();
    const location = useLocation();
    const { inputValue, setInputValue, activeQuery, clearQuery } = useFileSearchQuery({
        shouldAlwaysReplaceHistory: true,
    });
    const isSearching = activeQuery !== '';
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
    } = useDocumentsList(requestFilters, requestSort, !isSearching);
    const search = useDocumentsSearch(activeQuery, requestFilters);
    const { hasSelection, clear: clearSelection } = useFileSelectionContext();
    const { explicitTarget, closeDetails } = useFileDetails();
    const { interactionFor, menuAnchorPosition, menuTargetFiles, closeMenu } =
        useDocumentItemInteractions();

    const selectionResetKey = JSON.stringify([requestFilters, requestSort, activeQuery]);
    useEffect(() => {
        clearSelection();
    }, [selectionResetKey, clearSelection]);

    const selectableFiles = useMemo(
        () =>
            isSearching
                ? search.results.map(searchResultToFileData)
                : documents.map(documentToFileData),
        [isSearching, search.results, documents]
    );

    const openDocument = (fileId: number, pageNumber?: number) => {
        navigate(buildDocumentViewerUrl(fileId, location.search, pageNumber));
    };

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
        isSearching || total === undefined
            ? undefined
            : t('DOCUMENTS_TOTAL', { count: String(total) });
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
                        <DocumentsUploadActions selectedFolderId={folderId} />
                        <DocumentsSortControl
                            sortState={sortState}
                            onChange={setSort}
                            isRelevanceLocked={isSearching}
                        />
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
                            onClick={isSearching ? search.retry : reload}
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
                    <FileSearchBar
                        value={inputValue}
                        onChange={setInputValue}
                        onClear={clearQuery}
                        placeholder={t('DOCUMENTS_SEARCH_PLACEHOLDER')}
                    />
                    <DocumentsFilterBar
                        filters={filters}
                        onChange={setFilters}
                        onReset={resetFilters}
                    />
                    {isError && !isSearching && folderId !== undefined ? (
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
                    {hasSelection ? <DocumentsBatchBar selectableFiles={selectableFiles} /> : null}
                    {isSearching ? (
                        search.isLoading && !search.isError ? (
                            <DocumentsListSkeleton viewMode="list" />
                        ) : (
                            <ListFeedback
                                isEmpty={search.results.length === 0}
                                isError={search.isError}
                                isNextPageError={search.isNextPageError}
                                errorTitle={t('DOCUMENTS_LOAD_ERROR')}
                                errorMessage={search.errorMessage}
                                onRetry={search.retry}
                                hasNextPage={search.hasNextPage}
                                isFetchingNextPage={search.isFetchingNextPage}
                                onLoadMore={search.fetchNextPage}
                                emptyState={
                                    <EmptyState
                                        icon={<FileText size={32} aria-hidden="true" />}
                                        title={t('DOCUMENTS_SEARCH_EMPTY_TITLE')}
                                        description={t('DOCUMENTS_SEARCH_EMPTY_DESCRIPTION')}
                                    />
                                }
                            >
                                <DocumentsSearchResults
                                    query={activeQuery}
                                    results={search.results}
                                    onOpenDocument={openDocument}
                                    getInteraction={(result) =>
                                        interactionFor(searchResultToFileData(result))
                                    }
                                />
                            </ListFeedback>
                        )
                    ) : isLoading && !isError ? (
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
                                        <DocumentCard
                                            key={document.id}
                                            document={document}
                                            interaction={interactionFor(
                                                documentToFileData(document)
                                            )}
                                            onOpen={() => openDocument(document.id)}
                                        />
                                    ))}
                                </ul>
                            ) : (
                                <ul className={styles.list} aria-label={t('NAV_DOCUMENTS')}>
                                    {documents.map((document) => (
                                        <DocumentRow
                                            key={document.id}
                                            document={document}
                                            interaction={interactionFor(
                                                documentToFileData(document)
                                            )}
                                            onOpen={() => openDocument(document.id)}
                                        />
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
            <Drawer anchor="right" open={explicitTarget !== null} onClose={closeDetails}>
                <div className={styles.detailsDrawerContent}>
                    {explicitTarget ? (
                        <FileDetails file={explicitTarget} onClose={closeDetails} />
                    ) : null}
                </div>
            </Drawer>
            <DocumentItemMenu
                anchorPosition={menuAnchorPosition}
                targetFiles={menuTargetFiles}
                onClose={closeMenu}
            />
        </PageContainer>
    );
};

const DocumentsScreen = () => (
    <FileDetailsProvider>
        <DocumentsScreenContent />
    </FileDetailsProvider>
);

export default DocumentsScreen;
