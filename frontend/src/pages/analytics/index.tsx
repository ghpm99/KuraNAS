import { useMemo, useState, useEffect } from 'react';
import {
	Box,
	Chip,
	FormControl,
	Grid,
	InputLabel,
	LinearProgress,
	MenuItem,
	Select,
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableRow,
	Typography,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import Button from '@/components/ui/Button/Button';
import Layout from '@/components/layout/Layout';
import useI18n from '@/components/i18n/provider/i18nContext';
import { AnalyticsProvider } from '@/components/providers/analyticsProvider';
import { useAnalyticsOverview } from '@/components/providers/analyticsProvider/analyticsContext';
import { useAnalyticsFormatters } from '@/components/hooks/useAnalyticsFormatters/useAnalyticsFormatters';
import { useAnalyticsDerived } from '@/components/hooks/useAnalyticsDerived/useAnalyticsDerived';
import AnalyticsKpiCard from '@/components/analyticsV2/AnalyticsKpiCard';
import AnalyticsSection from '@/components/analyticsV2/AnalyticsSection';
import AnalyticsChartCard from '@/components/analyticsV2/AnalyticsChartCard';
import { AnalyticsPeriod } from '@/types/analytics';
import styles from './analytics.module.css';

const AnalyticsContent = () => {
	const { t } = useI18n();
	const navigate = useNavigate();
	const { formatBytes, formatPercent, formatDate } = useAnalyticsFormatters();
	const { period, setPeriod, data, loading, error, refresh } = useAnalyticsOverview();
	const { usedPercent, reclaimablePercent } = useAnalyticsDerived(data);
	const [now, setNow] = useState(0);

	useEffect(() => {
		const timer = setInterval(() => setNow(Date.now()), 60000);
		return () => clearInterval(timer);
	}, []);

	const updatedMinutes = (() => {
		if (!data?.generated_at) return '-';
		const generatedTime = new Date(data.generated_at).getTime();
		if (Number.isNaN(generatedTime)) return '-';
		const minutes = Math.max(0, Math.floor((now - generatedTime) / 60000));
		return String(minutes);
	})();

	const healthStatusLabel = useMemo(() => {
		switch (data?.health.status) {
			case 'scanning':
				return t('ANALYTICS_STATUS_SCANNING');
			case 'error':
				return t('ANALYTICS_STATUS_ERROR');
			default:
				return t('ANALYTICS_STATUS_OK');
		}
	}, [data?.health.status, t]);

	return (
		<div className={styles.content}>
			<div className={styles.headerRow}>
				<div>
					<Typography variant='h4'>{t('ANALYTICS_PAGE_TITLE')}</Typography>
					<Typography color='text.secondary'>{t('ANALYTICS_PAGE_DESCRIPTION')}</Typography>
				</div>
				<div className={styles.headerActions}>
					<FormControl size='small' sx={{ minWidth: 120 }}>
						<InputLabel id='analytics-period'>{t('ANALYTICS_PERIOD')}</InputLabel>
						<Select
							labelId='analytics-period'
							value={period}
							label={t('ANALYTICS_PERIOD')}
							onChange={(event) => setPeriod(event.target.value as AnalyticsPeriod)}
						>
							<MenuItem value='24h'>24h</MenuItem>
							<MenuItem value='7d'>7d</MenuItem>
							<MenuItem value='30d'>30d</MenuItem>
							<MenuItem value='90d'>90d</MenuItem>
						</Select>
					</FormControl>
					<Button onClick={() => void refresh()}>{t('ANALYTICS_REFRESH')}</Button>
				</div>
			</div>

			<div className={styles.quickActions}>
				<Button variant='secondary' onClick={() => navigate('/files')}>
					{t('ANALYTICS_ACTION_VIEW_DUPLICATES')}
				</Button>
				<Button variant='secondary' onClick={() => navigate('/files')}>
					{t('ANALYTICS_ACTION_VIEW_RECENT')}
				</Button>
				<Button variant='secondary' onClick={() => navigate('/files')}>
					{t('ANALYTICS_ACTION_VIEW_LARGEST')}
				</Button>
				<Button variant='secondary' onClick={() => navigate('/files')}>
					{t('ANALYTICS_ACTION_REINDEX')}
				</Button>
			</div>

			<Grid container spacing={2} sx={{ mb: 2 }}>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard
						title={t('ANALYTICS_KPI_STORAGE')}
						value={`${formatBytes(data?.storage.used_bytes ?? 0)} / ${formatBytes(data?.storage.total_bytes ?? 0)}`}
						helpText={`${t('ANALYTICS_FREE')}: ${formatBytes(data?.storage.free_bytes ?? 0)}`}
					/>
				</Grid>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard
						title={t('ANALYTICS_KPI_GROWTH')}
						value={`${data?.storage.growth_bytes && data.storage.growth_bytes > 0 ? '+' : ''}${formatBytes(data?.storage.growth_bytes ?? 0)}`}
						helpText={t('ANALYTICS_KPI_GROWTH_HELP')}
					/>
				</Grid>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard
						title={t('ANALYTICS_KPI_FILES_ADDED')}
						value={String(data?.counts.files_added ?? 0)}
						helpText={`${t('ANALYTICS_FILE_COUNT')}: ${(data?.counts.files_total ?? 0).toLocaleString()}`}
					/>
				</Grid>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard
						title={t('ANALYTICS_KPI_HOT_FOLDERS')}
						value={String(data?.hot_folders.length ?? 0)}
						helpText={(data?.hot_folders[0]?.path ?? '-').slice(0, 48)}
					/>
				</Grid>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard
						title={t('ANALYTICS_KPI_DUPLICATES')}
						value={String(data?.duplicates.groups ?? 0)}
						helpText={`${t('ANALYTICS_WASTED_SPACE')}: ${formatBytes(data?.duplicates.reclaimable_size ?? 0)}`}
					/>
				</Grid>
				<Grid size={{ xs: 12, md: 4 }}>
					<AnalyticsKpiCard title={t('ANALYTICS_KPI_INDEX_STATUS')} value={healthStatusLabel} helpText={t('ANALYTICS_KPI_INDEX_HELP')} />
				</Grid>
			</Grid>

			<Grid container spacing={2} sx={{ mb: 2 }}>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsChartCard
						title={t('ANALYTICS_STORAGE_TREND')}
						loading={loading}
						errorKey={error}
						empty={!data?.time_series.length}
						emptyKey='ANALYTICS_EMPTY'
					>
						<Box sx={{ mb: 2 }}>
							<Typography variant='body2' color='text.secondary'>
								{t('ANALYTICS_USED')}: {formatPercent(usedPercent)}
							</Typography>
							<LinearProgress variant='determinate' value={Math.min(100, usedPercent)} sx={{ mt: 1 }} />
						</Box>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('ANALYTICS_DATE')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_SPACE_USED')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.time_series.slice(-10).map((point) => (
									<TableRow key={point.date}>
										<TableCell>{point.date}</TableCell>
										<TableCell align='right'>{formatBytes(point.used_bytes)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsChartCard>
				</Grid>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsChartCard
						title={t('ANALYTICS_STORAGE_BY_CATEGORY')}
						loading={loading}
						errorKey={error}
						empty={!data?.types.length}
					>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('TYPE')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_FILE_COUNT')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_TOTAL_SPACE')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.types.map((item) => (
									<TableRow key={item.type}>
										<TableCell>{item.type}</TableCell>
										<TableCell align='right'>{item.count.toLocaleString()}</TableCell>
										<TableCell align='right'>{formatBytes(item.bytes)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsChartCard>
				</Grid>
			</Grid>

			<Grid container spacing={2} sx={{ mb: 2 }}>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsSection
						title={t('ANALYTICS_TOP_FOLDERS')}
						loading={loading}
						errorKey={error}
						empty={!data?.top_folders.length}
					>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('PATH')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_TOTAL_SPACE')}</TableCell>
									<TableCell align='right'>{t('MODIFIED')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.top_folders.map((folder) => (
									<TableRow key={folder.path} hover onClick={() => navigate('/files')} sx={{ cursor: 'pointer' }}>
										<TableCell>{folder.path}</TableCell>
										<TableCell align='right'>{formatBytes(folder.bytes)}</TableCell>
										<TableCell align='right'>{formatDate(folder.last_modified)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsSection>
				</Grid>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsSection
						title={t('ANALYTICS_EXTENSIONS')}
						loading={loading}
						errorKey={error}
						empty={!data?.extensions.length}
					>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('ANALYTICS_EXTENSION')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_FILE_COUNT')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_TOTAL_SPACE')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.extensions.map((item) => (
									<TableRow key={item.ext}>
										<TableCell>{item.ext}</TableCell>
										<TableCell align='right'>{item.count.toLocaleString()}</TableCell>
										<TableCell align='right'>{formatBytes(item.bytes)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsSection>
				</Grid>
			</Grid>

			<Grid container spacing={2} sx={{ mb: 2 }}>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsSection
						title={t('ANALYTICS_RECENT_FILES')}
						loading={loading}
						errorKey={error}
						empty={!data?.recent_files.length}
					>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('NAME')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_FILE_SIZE')}</TableCell>
									<TableCell align='right'>{t('CREATED')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.recent_files.map((file) => (
									<TableRow key={file.id} hover onClick={() => navigate('/files')} sx={{ cursor: 'pointer' }}>
										<TableCell>{file.name}</TableCell>
										<TableCell align='right'>{formatBytes(file.size_bytes)}</TableCell>
										<TableCell align='right'>{formatDate(file.created_at)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsSection>
				</Grid>
				<Grid size={{ xs: 12, lg: 6 }}>
					<AnalyticsSection
						title={t('ANALYTICS_DUPLICATE_FILES')}
						loading={loading}
						errorKey={error}
						empty={!data?.duplicates.top_groups.length}
					>
						<Box sx={{ mb: 2 }}>
							<Typography variant='body2' color='text.secondary'>
								{t('ANALYTICS_WASTED_SPACE')}: {formatBytes(data?.duplicates.reclaimable_size ?? 0)} ({formatPercent(reclaimablePercent)})
							</Typography>
						</Box>
						<Table size='small'>
							<TableHead>
								<TableRow>
									<TableCell>{t('ANALYTICS_SIGNATURE')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_COPIES')}</TableCell>
									<TableCell align='right'>{t('ANALYTICS_WASTED_SPACE')}</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data?.duplicates.top_groups.slice(0, 10).map((group) => (
									<TableRow key={group.signature}>
										<TableCell>{group.signature.slice(0, 12)}...</TableCell>
										<TableCell align='right'>{group.copies}</TableCell>
										<TableCell align='right'>{formatBytes(group.reclaimable_size)}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</AnalyticsSection>
				</Grid>
			</Grid>

			<Grid container spacing={2} sx={{ mb: 2 }}>
				<Grid size={{ xs: 12 }}>
					<AnalyticsSection title={t('ANALYTICS_INDEX_HEALTH')} loading={loading} errorKey={error}>
						<Box className={styles.healthHeader}>
							<Chip label={healthStatusLabel} color={data?.health.status === 'error' ? 'error' : data?.health.status === 'scanning' ? 'warning' : 'success'} />
							<Typography color='text.secondary'>
								{t('ANALYTICS_INDEXED_FILES')}: {(data?.health.indexed_files ?? 0).toLocaleString()} • {t('ANALYTICS_ERRORS_24H')}:{' '}
								{data?.health.errors_last_24h ?? 0}
							</Typography>
							<Typography color='text.secondary'>
								{t('ANALYTICS_LAST_SCAN')}: {formatDate(data?.health.last_scan_at ?? '')}
							</Typography>
						</Box>
						<Table size='small'>
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
				</Grid>
			</Grid>

			<div className={styles.footerRow}>
				<Typography variant='body2' color='text.secondary'>
					{t('ANALYTICS_UPDATED_MINUTES', { minutes: updatedMinutes })}
				</Typography>
				<Button variant='secondary' onClick={() => void refresh()}>
					{t('ANALYTICS_REFRESH')}
				</Button>
			</div>
		</div>
	);
};

const AnalyticsPage = () => {
	return (
		<Layout>
			<AnalyticsProvider>
				<AnalyticsContent />
			</AnalyticsProvider>
		</Layout>
	);
};

export default AnalyticsPage;
