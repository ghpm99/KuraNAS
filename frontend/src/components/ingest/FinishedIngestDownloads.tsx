import { Box, Button, CircularProgress, Typography } from '@mui/material';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import EmptyState from '@/components/emptyState/emptyState';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatDateTime } from '@/shared/utils/formatDateTime';
import type { IngestDownload } from '@/types/ingest';
import styles from './ActiveIngestDownloads.module.css';
import { useFinishedIngestDownloads } from './useFinishedIngestDownloads';
import { useOpenIngestDownloadLocation } from './useOpenIngestDownloadLocation';

const stateLabelKeyOf = (state: string): string => {
    if (state === 'completed') return 'INGEST_HISTORY_STATE_COMPLETED';
    if (state === 'canceled') return 'INGEST_HISTORY_STATE_CANCELED';
    return 'INGEST_HISTORY_STATE_FAILED';
};

const destinationOf = (download: IngestDownload): string =>
    download.subfolder
        ? `${download.target_root_name}/${download.subfolder}`
        : download.target_root_name;

const FinishedIngestDownloads = () => {
    const { t } = useI18n();
    const history = useFinishedIngestDownloads();
    const location = useOpenIngestDownloadLocation();

    return (
        <section className={styles.block}>
            <Typography variant="h6">{t('INGEST_HISTORY_TITLE')}</Typography>
            {history.isLoading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                    <CircularProgress size={24} aria-label={t('LOADING')} />
                </Box>
            ) : (
                <ListFeedback
                    isEmpty={history.downloads.length === 0}
                    isError={history.isError}
                    isNextPageError={history.isNextPageError}
                    errorTitle={t('INGEST_HISTORY_LOAD_ERROR')}
                    errorMessage={history.errorMessage}
                    onRetry={history.retry}
                    hasNextPage={history.hasNextPage}
                    isFetchingNextPage={history.isFetchingNextPage}
                    onLoadMore={history.fetchNextPage}
                    emptyState={<EmptyState title={t('INGEST_HISTORY_EMPTY')} />}
                >
                    <ul className={styles.list}>
                        {history.downloads.map((download) => (
                            <li key={download.id} className={styles.item}>
                                <Typography variant="subtitle1">
                                    {download.title ?? download.url}
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    {t(stateLabelKeyOf(download.state))}
                                    {' · '}
                                    {formatDateTime(download.ended_at ?? download.created_at)}
                                    {' · '}
                                    {t('INGEST_ACTIVE_DESTINATION', {
                                        destination: destinationOf(download),
                                    })}
                                </Typography>
                                {download.state !== 'completed' && download.error ? (
                                    <Typography variant="body2" color="error">
                                        {download.error}
                                    </Typography>
                                ) : null}
                                {download.state === 'completed' ? (
                                    <div className={styles.actionsRow}>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            disabled={download.file_id === null}
                                            onClick={() =>
                                                download.file_id !== null &&
                                                void location.openFile(
                                                    download.id,
                                                    download.file_id
                                                )
                                            }
                                        >
                                            {t('INGEST_HISTORY_OPEN')}
                                        </Button>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            disabled={download.file_id === null}
                                            onClick={() =>
                                                download.file_id !== null &&
                                                void location.openFolder(
                                                    download.id,
                                                    download.file_id
                                                )
                                            }
                                        >
                                            {t('INGEST_HISTORY_OPEN_FOLDER')}
                                        </Button>
                                    </div>
                                ) : null}
                                {download.state === 'failed' || download.state === 'canceled' ? (
                                    <div className={styles.actionsRow}>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            disabled={history.isRetrying(download.id)}
                                            onClick={() => void history.retryDownload(download.id)}
                                        >
                                            {t('INGEST_HISTORY_RETRY')}
                                        </Button>
                                    </div>
                                ) : null}
                                {location.failedDownloadId === download.id ? (
                                    <Typography variant="body2" color="error">
                                        {t('INGEST_HISTORY_LOCATION_ERROR')}
                                    </Typography>
                                ) : null}
                            </li>
                        ))}
                    </ul>
                    {!history.hasNextPage && !history.isNextPageError ? (
                        <Typography variant="body2" color="text.secondary">
                            {t('INGEST_HISTORY_END')}
                        </Typography>
                    ) : null}
                </ListFeedback>
            )}
        </section>
    );
};

export default FinishedIngestDownloads;
