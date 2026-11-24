import { FileText } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import HighlightedText from '@/components/search/HighlightedText';
import { formatDocumentSnippet } from '@/shared/utils/formatDocumentSnippet';
import { formatSize } from '@/shared/utils/formatSize';
import { formatDate } from '@/utils';
import type { DocumentSearchResult } from '@/service/documents';
import styles from './DocumentsSearchResults.module.css';

type DocumentsSearchResultsProps = {
    query: string;
    results: DocumentSearchResult[];
    onOpenDocument: (fileId: number, page?: number) => void;
};

const DocumentsSearchResults = ({ query, results, onOpenDocument }: DocumentsSearchResultsProps) => {
    const { t } = useI18n();

    return (
        <ul className={styles.list} aria-label={t('DOCUMENTS_SEARCH_RESULTS')}>
            {results.map((result) => {
                const snippetText = formatDocumentSnippet(result.snippet);
                const pageNumber = result.page ?? undefined;
                const detailsText = [
                    result.parent_path,
                    formatSize(result.size ?? 0),
                    result.updated_at ? formatDate(result.updated_at) : '',
                ]
                    .filter(Boolean)
                    .join(' · ');

                return (
                    <li key={result.file_id}>
                        <button
                            type="button"
                            className={styles.result}
                            onClick={() => onOpenDocument(result.file_id, pageNumber)}
                        >
                            <FileText size={18} className={styles.icon} aria-hidden="true" />
                            <span className={styles.body}>
                                <span className={styles.name}>
                                    <HighlightedText text={result.name} query={query} />
                                </span>
                                {snippetText ? (
                                    <span className={styles.snippet}>
                                        <HighlightedText text={snippetText} query={query} />
                                    </span>
                                ) : null}
                                <span className={styles.meta}>{detailsText}</span>
                            </span>
                            {pageNumber !== undefined ? (
                                <span className={styles.page}>
                                    {t('DOCUMENTS_SEARCH_PAGE', { page: String(pageNumber) })}
                                </span>
                            ) : null}
                        </button>
                    </li>
                );
            })}
        </ul>
    );
};

export default DocumentsSearchResults;
