import { Button, LinearProgress, Typography } from '@mui/material';
import ErrorState from '@/components/errorState/errorState';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import styles from './ActiveIngestDownloads.module.css';
import { formatRemainingTime } from '@/shared/transfer/formatRemainingTime';
import { useActiveIngestDownloads } from './useActiveIngestDownloads';

const ActiveIngestDownloads = () => {
    const { t } = useI18n();
    const { downloads, presetLabelFor, cancelDownload, isCanceling, isError, errorMessage, retry } =
        useActiveIngestDownloads();

    if (isError) {
        return (
            <section className={styles.block}>
                <ErrorState
                    title={t('INGEST_ACTIVE_LOAD_ERROR')}
                    backendMessage={errorMessage}
                    onRetry={retry}
                />
            </section>
        );
    }

    if (downloads.length === 0) return null;

    return (
        <section className={styles.block}>
            <Typography variant="h6">{t('INGEST_ACTIVE_TITLE')}</Typography>
            <ul className={styles.list}>
                {downloads.map((download) => {
                    const destination = download.subfolder
                        ? `${download.target_root_name}/${download.subfolder}`
                        : download.target_root_name;
                    const progress = Math.min(100, Math.max(0, download.progress));
                    return (
                        <li key={download.id} className={styles.item}>
                            <Typography variant="subtitle1">
                                {download.title ?? download.url}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                                {t('INGEST_ACTIVE_DESTINATION', { destination })}
                                {' · '}
                                {t('INGEST_ACTIVE_FORMAT', {
                                    format: presetLabelFor(download.preset),
                                })}
                            </Typography>
                            <div className={styles.progressRow}>
                                <LinearProgress
                                    className={styles.progressBar}
                                    variant="determinate"
                                    value={progress}
                                />
                                <Typography variant="body2">{progress}%</Typography>
                            </div>
                            <Typography variant="body2" color="text.secondary">
                                {download.speed_bytes_per_second !== null &&
                                    `${t('INGEST_ACTIVE_SPEED', {
                                        speed: formatSize(download.speed_bytes_per_second),
                                    })} · `}
                                {download.eta_seconds !== null
                                    ? t('INGEST_ACTIVE_ETA', {
                                          time: formatRemainingTime(download.eta_seconds, t),
                                      })
                                    : t('INGEST_ACTIVE_ETA_UNKNOWN')}
                            </Typography>
                            {download.job_id !== null && (
                                <Button
                                    size="small"
                                    color="error"
                                    disabled={isCanceling(download.job_id)}
                                    onClick={() => void cancelDownload(download.job_id as number)}
                                    sx={{ alignSelf: 'flex-start' }}
                                >
                                    {t('INGEST_ACTIVE_CANCEL')}
                                </Button>
                            )}
                        </li>
                    );
                })}
            </ul>
        </section>
    );
};

export default ActiveIngestDownloads;
