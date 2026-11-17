import { Link as RouterLink } from 'react-router-dom';
import { Link, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { buildFilesUrl } from '@/app/routes';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useLargestFilesPage } from '@/components/analytics/useLargestFilesPage';
import { formatSize } from '@/shared/utils/formatSize';
import styles from './AnalyticsScreen.module.css';

const AnalyticsLargestFilesScreen = () => {
    const { t } = useI18n();
    const {
        files,
        isPending,
        isError,
        isFetchNextPageError,
        isFetchingNextPage,
        hasNextPage,
        fetchNextPage,
        errorMessage,
        retry,
    } = useLargestFilesPage();

    if (isPending) {
        return <p role="status">{t('LOADING')}</p>;
    }

    return (
        <ListFeedback
            isEmpty={files.length === 0}
            isError={isError}
            isNextPageError={isFetchNextPageError}
            errorTitle={t('ANALYTICS_LARGEST_ERROR_TITLE')}
            errorMessage={errorMessage}
            onRetry={retry}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            onLoadMore={fetchNextPage}
            emptyState={<p>{t('ANALYTICS_LARGEST_EMPTY')}</p>}
        >
            <Table size="small" aria-label={t('ANALYTICS_SECTION_LARGEST')}>
                <TableHead>
                    <TableRow>
                        <TableCell>{t('NAME')}</TableCell>
                        <TableCell>{t('ANALYTICS_COLUMN_FOLDER')}</TableCell>
                        <TableCell align="right">{t('SIZE')}</TableCell>
                        <TableCell>{t('TYPE')}</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {files.map((file) => (
                        <TableRow key={file.id} hover>
                            <TableCell>
                                <Link component={RouterLink} to={buildFilesUrl(file.path)}>
                                    {file.name}
                                </Link>
                            </TableCell>
                            <TableCell className={styles.tableCellAction}>
                                <Link component={RouterLink} to={buildFilesUrl(file.parent_path)}>
                                    {file.parent_path}
                                </Link>
                            </TableCell>
                            <TableCell align="right">{formatSize(file.size)}</TableCell>
                            <TableCell>{file.format}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </ListFeedback>
    );
};

export default AnalyticsLargestFilesScreen;
