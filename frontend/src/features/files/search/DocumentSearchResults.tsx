import { Link } from 'react-router-dom';
import { FileText } from 'lucide-react';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import HighlightedText from '@/components/search/HighlightedText';
import { buildFilesUrl } from '@/app/routes';
import { formatDocumentSnippet } from '@/shared/utils/formatDocumentSnippet';
import { formatSize } from '@/shared/utils/formatSize';
import { formatShortDate } from '@/shared/utils/formatShortDate';
import type { DocumentSearchResult } from '@/service/documents';
import styles from './fileSearch.module.css';

interface DocumentSearchResultsProps {
    query?: string;
    items?: DocumentSearchResult[];
    status?: string;
    hasNextPage?: boolean;
    isFetchingNextPage?: boolean;
    fetchNextPage?: () => void;
    errorMessage?: string;
    retry?: () => void;
}

const DocumentSearchResults = ({
    query = '',
    items = [],
    status = 'pending',
    hasNextPage = false,
    isFetchingNextPage = false,
    fetchNextPage,
    errorMessage,
    retry,
}: DocumentSearchResultsProps) => {
    const { t } = useI18n();

    if (status === 'pending') {
        return (
            <p className={styles.contentStatus} role="status">
                {t('LOADING')}
            </p>
        );
    }

    if (status === 'error') {
        return (
            <ErrorState
                title={t('FILES_LISTING_ERROR_TITLE')}
                backendMessage={errorMessage}
                onRetry={retry}
            />
        );
    }

    if (items.length === 0) {
        return <p className={styles.contentStatus}>{t('FILES_SEARCH_CONTENT_EMPTY')}</p>;
    }

    return (
        <div>
            <ul className={styles.documentList} aria-label={t('FILES_SEARCH_MODE_CONTENT')}>
                {items.map((documentResult) => {
                    const snippetText = formatDocumentSnippet(documentResult.snippet);
                    const modifiedText = documentResult.updated_at
                        ? formatShortDate(documentResult.updated_at)
                        : '';
                    const detailsText = [formatSize(documentResult.size ?? 0), modifiedText]
                        .filter(Boolean)
                        .join(' · ');

                    return (
                        <li key={documentResult.file_id}>
                            <Link to={buildFilesUrl(documentResult.path)} className={styles.documentRow}>
                                <FileText size={18} className={styles.documentIcon} />
                                <span className={styles.documentBody}>
                                    <span className={styles.documentName}>
                                        <HighlightedText text={documentResult.name} query={query} />
                                    </span>
                                    {snippetText ? (
                                        <span className={styles.documentSnippet}>
                                            <HighlightedText text={snippetText} query={query} />
                                        </span>
                                    ) : null}
                                    <span className={styles.documentMeta}>
                                        {[documentResult.parent_path, detailsText]
                                            .filter(Boolean)
                                            .join(' · ')}
                                    </span>
                                </span>
                            </Link>
                        </li>
                    );
                })}
            </ul>
            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
            />
        </div>
    );
};

export default DocumentSearchResults;
