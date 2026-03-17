import { appRoutes } from '@/app/routes';
import Button from '@/components/ui/Button/Button';
import AnalyticsChartCard from '@/components/analyticsV2/AnalyticsChartCard';
import AnalyticsKpiCard from '@/components/analyticsV2/AnalyticsKpiCard';
import AnalyticsSection from '@/components/analyticsV2/AnalyticsSection';
import { type AnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import {
    Box,
    Chip,
    LinearProgress,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import styles from './AnalyticsScreen.module.css';

interface AnalyticsOverviewScreenProps {
    state: AnalyticsScreenState;
}

const AnalyticsOverviewScreen = ({ state }: AnalyticsOverviewScreenProps) => {
    const navigate = useNavigate();
    const {
        t,
        data,
        loading,
        error,
        formatBytes,
        formatPercent,
        formatDate,
        usedPercent,
        reclaimablePercent,
        healthStatusLabel,
        healthStatusColor,
        refresh,
    } = state;

    return (
        <div className={styles.screen}>
            <div className={styles.intro}>
                <h2 className={styles.title}>{t('ANALYTICS_OVERVIEW_TITLE')}</h2>
                <p className={styles.description}>{t('ANALYTICS_OVERVIEW_DESCRIPTION')}</p>
            </div>

            <div className={styles.actions}>
                <Button variant="secondary" onClick={() => navigate(appRoutes.files)}>
                    {t('ANALYTICS_ACTION_VIEW_DUPLICATES')}
                </Button>
                <Button variant="secondary" onClick={() => navigate(appRoutes.files)}>
                    {t('ANALYTICS_ACTION_VIEW_RECENT')}
                </Button>
                <Button variant="secondary" onClick={() => navigate(appRoutes.files)}>
                    {t('ANALYTICS_ACTION_VIEW_LARGEST')}
                </Button>
                <Button variant="secondary" onClick={() => void refresh()}>
                    {t('ANALYTICS_ACTION_REINDEX')}
                </Button>
            </div>

            <div className={styles.grid}>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_STORAGE')}
                        value={`${formatBytes(data?.storage.used_bytes ?? 0)} / ${formatBytes(data?.storage.total_bytes ?? 0)}`}
                        helpText={`${t('ANALYTICS_FREE')}: ${formatBytes(data?.storage.free_bytes ?? 0)}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_GROWTH')}
                        value={`${data?.storage.growth_bytes && data.storage.growth_bytes > 0 ? '+' : ''}${formatBytes(data?.storage.growth_bytes ?? 0)}`}
                        helpText={t('ANALYTICS_KPI_GROWTH_HELP')}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_FILES_ADDED')}
                        value={String(data?.counts.files_added ?? 0)}
                        helpText={`${t('ANALYTICS_FILE_COUNT')}: ${(data?.counts.files_total ?? 0).toLocaleString()}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_HOT_FOLDERS')}
                        value={String(data?.hot_folders.length ?? 0)}
                        helpText={(data?.hot_folders[0]?.path ?? '-').slice(0, 48)}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_DUPLICATES')}
                        value={String(data?.duplicates.groups ?? 0)}
                        helpText={`${t('ANALYTICS_WASTED_SPACE')}: ${formatBytes(data?.duplicates.reclaimable_size ?? 0)}`}
                    />
                </div>
                <div className={styles.span4}>
                    <AnalyticsKpiCard
                        title={t('ANALYTICS_KPI_INDEX_STATUS')}
                        value={healthStatusLabel}
                        helpText={t('ANALYTICS_KPI_INDEX_HELP')}
                    />
                </div>
            </div>

            <div className={styles.grid}>
                <div className={styles.span6}>
                    <AnalyticsChartCard
                        title={t('ANALYTICS_STORAGE_TREND')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.time_series.length}
                        emptyKey="ANALYTICS_EMPTY"
                    >
                        <Box className={styles.inlineSummary}>
                            <Typography variant="body2" color="text.secondary">
                                {t('ANALYTICS_USED')}: {formatPercent(usedPercent)}
                            </Typography>
                            <LinearProgress
                                variant="determinate"
                                value={Math.min(100, usedPercent)}
                                className={styles.progress}
                            />
                        </Box>
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_DATE')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_SPACE_USED')}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.time_series.slice(-10).map((point) => (
                                    <TableRow key={point.date}>
                                        <TableCell>{point.date}</TableCell>
                                        <TableCell align="right">
                                            {formatBytes(point.used_bytes)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </AnalyticsChartCard>
                </div>
                <div className={styles.span6}>
                    <AnalyticsChartCard
                        title={t('ANALYTICS_STORAGE_BY_CATEGORY')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.types.length}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('TYPE')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_FILE_COUNT')}</TableCell>
                                    <TableCell align="right">
                                        {t('ANALYTICS_TOTAL_SPACE')}
                                    </TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.types.map((item) => (
                                    <TableRow key={item.type}>
                                        <TableCell>{item.type}</TableCell>
                                        <TableCell align="right">
                                            {item.count.toLocaleString()}
                                        </TableCell>
                                        <TableCell align="right">
                                            {formatBytes(item.bytes)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </AnalyticsChartCard>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_TOP_FOLDERS')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.top_folders.length}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('PATH')}</TableCell>
                                    <TableCell align="right">
                                        {t('ANALYTICS_TOTAL_SPACE')}
                                    </TableCell>
                                    <TableCell align="right">{t('MODIFIED')}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.top_folders.map((folder) => (
                                    <TableRow
                                        key={folder.path}
                                        hover
                                        onClick={() => navigate(appRoutes.files)}
                                        className={styles.tableCellAction}
                                    >
                                        <TableCell>{folder.path}</TableCell>
                                        <TableCell align="right">
                                            {formatBytes(folder.bytes)}
                                        </TableCell>
                                        <TableCell align="right">
                                            {formatDate(folder.last_modified)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_EXTENSIONS')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.extensions.length}
                    >
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_EXTENSION')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_FILE_COUNT')}</TableCell>
                                    <TableCell align="right">
                                        {t('ANALYTICS_TOTAL_SPACE')}
                                    </TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.extensions.map((item) => (
                                    <TableRow key={item.ext}>
                                        <TableCell>{item.ext}</TableCell>
                                        <TableCell align="right">
                                            {item.count.toLocaleString()}
                                        </TableCell>
                                        <TableCell align="right">
                                            {formatBytes(item.bytes)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </AnalyticsSection>
                </div>
                <div className={styles.span6}>
                    <AnalyticsSection
                        title={t('ANALYTICS_DUPLICATE_FILES')}
                        loading={loading}
                        errorKey={error}
                        empty={!data?.duplicates.top_groups.length}
                    >
                        <Box className={styles.inlineSummary}>
                            <Typography variant="body2" color="text.secondary">
                                {t('ANALYTICS_WASTED_SPACE')}:{' '}
                                {formatBytes(data?.duplicates.reclaimable_size ?? 0)} (
                                {formatPercent(reclaimablePercent)})
                            </Typography>
                        </Box>
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>{t('ANALYTICS_SIGNATURE')}</TableCell>
                                    <TableCell align="right">{t('ANALYTICS_COPIES')}</TableCell>
                                    <TableCell align="right">
                                        {t('ANALYTICS_WASTED_SPACE')}
                                    </TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {data?.duplicates.top_groups.slice(0, 10).map((group) => (
                                    <TableRow key={group.signature}>
                                        <TableCell>{group.signature.slice(0, 12)}...</TableCell>
                                        <TableCell align="right">{group.copies}</TableCell>
                                        <TableCell align="right">
                                            {formatBytes(group.reclaimable_size)}
                                        </TableCell>
                                    </TableRow>
                                ))}
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
            </div>
        </div>
    );
};

export default AnalyticsOverviewScreen;
