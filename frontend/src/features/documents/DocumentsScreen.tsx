import { Button } from '@mui/material';
import { FileText, RefreshCw } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import EmptyState from '@/components/emptyState/emptyState';
import DocumentRow from './DocumentRow';
import DocumentsListSkeleton from './DocumentsListSkeleton';
import useDocumentsList from './useDocumentsList';
import styles from './DocumentsScreen.module.css';

const DocumentsScreen = () => {
    const { t } = useI18n();
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

    const subtitle = total === undefined ? undefined : t('DOCUMENTS_TOTAL', { count: String(total) });
    const isListFullyLoaded = documents.length > 0 && !hasNextPage && !isNextPageError;

    return (
        <PageContainer>
            <PageHeader
                title={t('NAV_DOCUMENTS')}
                subtitle={subtitle}
                actions={
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<RefreshCw size={16} />}
                        onClick={reload}
                    >
                        {t('DOCUMENTS_RELOAD')}
                    </Button>
                }
            />
            {isLoading && !isError ? (
                <DocumentsListSkeleton />
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
                    <ul className={styles.list} aria-label={t('NAV_DOCUMENTS')}>
                        {documents.map((document) => (
                            <DocumentRow key={document.id} document={document} />
                        ))}
                    </ul>
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
