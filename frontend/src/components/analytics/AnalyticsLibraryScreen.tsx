import { appRoutes } from '@/app/routes';
import AnalyticsKpiCard from '@/components/analyticsV2/AnalyticsKpiCard';
import AnalyticsSection from '@/components/analyticsV2/AnalyticsSection';
import { type AnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import {
    Box,
    Chip,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import styles from './AnalyticsScreen.module.css';

interface AnalyticsLibraryScreenProps {
    state: AnalyticsScreenState;
}

const AnalyticsLibraryScreen = ({ state }: AnalyticsLibraryScreenProps) => {
    const navigate = useNavigate();
    const {
        t,
        data,
        loading,
        error,
        formatDate,
        healthStatusLabel,
        healthStatusColor,
        processingFailureTotal,
    } = state;

    return (
        <div className={styles.screen}>
            <div className={styles.intro}>
                <h2 className={styles.title}>{t('ANALYTICS_LIBRARY_TITLE')}</h2>
                <p className={styles.description}>{t('ANALYTICS_LIBRARY_DESCRIPTION')}</p>
            </div>

            <div className={styles.grid}>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_INDEXED_FILES')}
                        value={(data?.health.indexed_files ?? 0).toLocaleString()}
                        helpText={`${t('ANALYTICS_INDEXED_FOLDERS')}: ${(data?.counts.folders ?? 0).toLocaleString()}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_CATEGORIZED_MEDIA')}
                        value={(data?.library.categorized_media ?? 0).toLocaleString()}
                        helpText={t('ANALYTICS_CATEGORIZED_MEDIA_HELP')}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_IMAGES_CLASSIFIED')}
                        value={(data?.library.image_classified ?? 0).toLocaleString()}
                        helpText={t('ANALYTICS_IMAGES_CLASSIFIED_HELP')}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_METADATA_PENDING')}
                        value={(data?.processing.metadata_pending ?? 0).toLocaleString()}
                        helpText={`${t('ANALYTICS_METADATA_FAILED')}: ${(data?.processing.metadata_failed ?? 0).toLocaleString()}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_THUMBNAIL_PENDING')}
                        value={(data?.processing.thumbnail_pending ?? 0).toLocaleString()}
                        helpText={`${t('ANALYTICS_THUMBNAIL_FAILED')}: ${(data?.processing.thumbnail_failed ?? 0).toLocaleString()}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_PROCESSING_FAILURES')}
                        value={processingFailureTotal.toLocaleString()}
                        helpText={`${t('ANALYTICS_ERRORS_24H')}: ${(data?.health.errors_last_24h ?? 0).toLocaleString()}`}
                    />
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_MEDIA_COVERAGE')}
                        loading={loading}
                        errorKey={error}
                        empty={!data}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('TYPE')}</TableCell>
                                    <TableCell align="right">
                                        {t('ANALYTICS_ITEMS_INDEXED')}
                                    </TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                <TableRow>
                                    <TableCell>{t('NAV_MUSIC')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.library.audio_with_metadata ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                                <TableRow>
                                    <TableCell>{t('NAV_VIDEOS')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.library.video_with_metadata ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                                <TableRow>
                                    <TableCell>{t('NAV_IMAGES')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.library.image_with_metadata ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_IMAGES_CLASSIFIED')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.library.image_classified ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_PROCESSING_QUEUE')}
                        loading={loading}
                        errorKey={error}
                        empty={!data}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('TYPE')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_PENDING')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_FAILED')}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_METADATA_LABEL')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.processing.metadata_pending ?? 0).toLocaleString()}
                                    </TableCell>
                                    <TableCell align="right">
                                        {(data?.processing.metadata_failed ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_THUMBNAIL_LABEL')}</TableCell>
                                    <TableCell align="right">
                                        {(data?.processing.thumbnail_pending ?? 0).toLocaleString()}
                                    </TableCell>
                                    <TableCell align="right">
                                        {(data?.processing.thumbnail_failed ?? 0).toLocaleString()}
                                    </TableCell>
                                </TableRow>
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_INDEX_HEALTH')}
                        loading={loading}
                        errorKey={error}
                    >
                        <Box className={styles.statusHeader}>
                            <Chip label={healthStatusLabel} color={healthStatusColor} />
                            <Typography color="text.secondary">
                                {t('ANALYTICS_INDEXED_FILES')}:{' '}
                                {(data?.health.indexed_files ?? 0).toLocaleString()} •{' '}
                                {t('ANALYTICS_ERRORS_24H')}: {data?.health.errors_last_24h ?? 0}
                            </Typography>
                            <Typography color="text.secondary">
                                {t('ANALYTICS_LAST_SCAN')}:{' '}
                                {formatDate(data?.health.last_scan_at ?? '')}
                            </Typography>
                        </Box>
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_RECENT_ERRORS')}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {(data?.health.recent_errors.length ?? 0) === 0 ? (
                                    <TableRow>
                                        <TableCell>{t('ANALYTICS_NO_ERRORS')}</TableCell>
                                    </TableRow>
                                ) : (
                                    data?.health.recent_errors.map((item) => (
                                        <TableRow key={item}>
                                            <TableCell>{item}</TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_RECENT_FILES')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.recent_files.length}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('NAME')}</TableCell>
                                    <TableCell>{t('PATH')}</TableCell>
                                    <TableCell align="right">{t('CREATED')}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.recent_files.slice(0, 10).map((file) => (
                                    <TableRow
                                        key={file.id}
                                        hover
                                        onClick={() => navigate(appRoutes.files)}
                                        className={styles.tableCellAction}
                                    >
                                        <TableCell>{file.name}</TableCell>
                                        <TableCell>{file.parent_path}</TableCell>
                                        <TableCell align="right">
                                            {formatDate(file.created_at)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
            </div>
        </div>
    );
};

export default AnalyticsLibraryScreen;
