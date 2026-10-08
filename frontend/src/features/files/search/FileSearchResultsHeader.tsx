import { Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './fileSearch.module.css';

interface FileSearchResultsHeaderProps {
    query?: string;
    resultCount?: number;
    hasMoreResults?: boolean;
    onClear?: () => void;
}

const FileSearchResultsHeader = ({
    query = '',
    resultCount = 0,
    hasMoreResults = false,
    onClear,
}: FileSearchResultsHeaderProps) => {
    const { t } = useI18n();

    const resultsLabel = (() => {
        if (hasMoreResults) {
            return t('FILES_SEARCH_RESULTS_MORE', { count: String(resultCount), query });
        }
        if (resultCount === 1) {
            return t('FILES_SEARCH_RESULTS_ONE', { query });
        }
        return t('FILES_SEARCH_RESULTS_TOTAL', { count: String(resultCount), query });
    })();

    return (
        <div className={styles.resultsHeader}>
            <span>{resultsLabel}</span>
            <Button size="small" onClick={onClear}>
                {t('FILES_SEARCH_CLEAR')}
            </Button>
        </div>
    );
};

export default FileSearchResultsHeader;
