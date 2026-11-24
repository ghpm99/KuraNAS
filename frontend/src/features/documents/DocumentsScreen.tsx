import { useState } from 'react';
import { Button, ToggleButton, ToggleButtonGroup } from '@mui/material';
import { FileText, LayoutGrid, List, RefreshCw } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import EmptyState from '@/components/emptyState/emptyState';
import DocumentCard from './DocumentCard';
import DocumentRow from './DocumentRow';
import DocumentsListSkeleton from './DocumentsListSkeleton';
import {
    loadDocumentsViewMode,
    saveDocumentsViewMode,
    type DocumentsViewMode,
} from './documentsViewModePreference';
import useDocumentsList from './useDocumentsList';
import styles from './DocumentsScreen.module.css';

const DocumentsScreen = () => {
    const { t } = useI18n();
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
    } = useDocumentsList();

    const changeViewMode = (nextViewMode: DocumentsViewMode | null) => {
        if (!nextViewMode) return;
        setViewMode(nextViewMode);
        saveDocumentsViewMode(nextViewMode);
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
                        <EmptyState
                            icon={<FileText size={32} aria-hidden="true" />}
                            title={t('DOCUMENTS_EMPTY_TITLE')}
                            description={t('DOCUMENTS_EMPTY_DESCRIPTION')}
                        />
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
        </PageContainer>
    );
};

export default DocumentsScreen;
