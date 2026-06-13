import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import { formatSize } from '@/utils';
import useTieringSettings, { tieringStatusKey } from './useTieringSettings';
import styles from './TieringSettingsSection.module.css';

type TieringSettingsSectionProps = {
	className?: string;
};

const BYTES_PER_MIB = 1024 * 1024;

const TieringSettingsSection = ({ className = '' }: TieringSettingsSectionProps) => {
	const { t, form, status, usage, isLoading, isSaving, hasError, hasUnsavedChanges, setField, handleSave } =
		useTieringSettings();

	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_TIERING_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_TIERING_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	const lastRunAt = status?.ended_at ?? status?.started_at;

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_TIERING_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_TIERING_DESCRIPTION')}</p>
			</div>
			{hasError ? <Alert severity="error">{t('SETTINGS_TIERING_LOAD_ERROR')}</Alert> : null}
			<div className={styles.summary}>
				{status?.has_run ? (
					<Chip
						variant="outlined"
						label={`${t('SETTINGS_TIERING_LAST_RUN')}: ${t(tieringStatusKey(status.status))}${
							lastRunAt ? ` · ${new Date(lastRunAt).toLocaleString()}` : ''
						}`}
					/>
				) : (
					<Chip variant="outlined" label={t('SETTINGS_TIERING_NEVER_RAN')} />
				)}
				{usage ? (
					<>
						<Chip
							variant="outlined"
							label={t('SETTINGS_TIERING_HOT_USAGE', {
								count: String(usage.hot_files),
								size: formatSize(usage.hot_bytes),
							})}
						/>
						<Chip
							variant="outlined"
							label={t('SETTINGS_TIERING_COLD_USAGE', {
								count: String(usage.cold_files),
								size: formatSize(usage.cold_bytes),
							})}
						/>
					</>
				) : null}
			</div>
			{status?.last_error ? <Alert severity="error">{status.last_error}</Alert> : null}
			<FormControlLabel
				control={
					<Switch
						checked={form.enabled}
						onChange={(_, checked) => setField('enabled', checked)}
						disabled={isSaving}
					/>
				}
				label={t('SETTINGS_TIERING_ENABLED')}
			/>
			<TextField
				size="small"
				fullWidth
				label={t('SETTINGS_TIERING_COLD_DIR')}
				value={form.cold_dir_path}
				onChange={(event) => setField('cold_dir_path', event.target.value)}
				disabled={isSaving}
			/>
			<div className={styles.numbers}>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_TIERING_MIN_AGE_DAYS')}
					value={String(form.min_age_days)}
					onChange={(event) => setField('min_age_days', Number(event.target.value))}
					disabled={isSaving}
				/>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_TIERING_MIN_SIZE_MIB')}
					value={String(Math.round(form.min_size_bytes / BYTES_PER_MIB))}
					onChange={(event) => setField('min_size_bytes', Number(event.target.value) * BYTES_PER_MIB)}
					disabled={isSaving}
				/>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_TIERING_INTERVAL_HOURS')}
					value={String(form.interval_hours)}
					onChange={(event) => setField('interval_hours', Number(event.target.value))}
					disabled={isSaving}
				/>
			</div>
			<div className={styles.actions}>
				<Button
					variant="contained"
					onClick={() => void handleSave()}
					disabled={isSaving || !hasUnsavedChanges}
				>
					{t('SETTINGS_TIERING_SAVE')}
				</Button>
			</div>
			<Alert severity="info">{t('SETTINGS_TIERING_HELP')}</Alert>
		</section>
	);
};

export default TieringSettingsSection;
