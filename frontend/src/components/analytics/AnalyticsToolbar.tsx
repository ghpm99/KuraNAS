import Button from '@/components/ui/Button/Button';
import { type AnalyticsScreenState } from '@/components/analytics/useAnalyticsScreenState';
import { AnalyticsPeriod } from '@/types/analytics';
import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';
import styles from './AnalyticsToolbar.module.css';

interface AnalyticsToolbarProps {
	state: AnalyticsScreenState;
}

const AnalyticsToolbar = ({ state }: AnalyticsToolbarProps) => {
	const { t, period, setPeriod, refresh, updatedMinutes } = state;

	return (
		<div className={styles.toolbar}>
			<div className={styles.summary}>
				<span className={styles.eyebrow}>{t('ANALYTICS_PERIOD')}</span>
				<span className={styles.updated}>{t('ANALYTICS_UPDATED_MINUTES', { minutes: updatedMinutes })}</span>
			</div>
			<div className={styles.actions}>
				<FormControl size='small' className={styles.selectControl}>
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
	);
};

export default AnalyticsToolbar;
