import { Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './fileSearch.module.css';

interface FileSearchResultsHeaderProps {
    query?: string;
    resultCount?: number;
    hasMoreResults?: boolean;
    isContentSearch?: boolean;
    onClear?: () => void;
}

const FileSearchResultsHeader = ({
    query = '',
    resultCount = 0,
    hasMoreResults = false,
    isContentSearch = false,
    onClear,
}: FileSearchResultsHeaderProps) => {
    const { t } = useI18n();

    const labelKeyPrefix = isContentSearch ? 'FILES_SEARCH_CONTENT_RESULTS' : 'FILES_SEARCH_RESULTS';

    const resultsLabel = (() => {
        if (hasMoreResults) {
            return t(`${labelKeyPrefix}_MORE`, { count: String(resultCount), query });
        }
        if (resultCount === 1) {
            return t(`${labelKeyPrefix}_ONE`, { query });
        }
        return t(`${labelKeyPrefix}_TOTAL`, { count: String(resultCount), query });
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
