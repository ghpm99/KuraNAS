import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import useBackupSettings, { backupStatusKey } from './useBackupSettings';
import styles from './BackupSettingsSection.module.css';

type BackupSettingsSectionProps = {
	className?: string;
};

const BackupSettingsSection = ({ className = '' }: BackupSettingsSectionProps) => {
	const {
		t,
		form,
		status,
		pendingFiles,
		isLoading,
		isSaving,
		hasError,
		hasUnsavedChanges,
		setField,
		handleSave,
	} = useBackupSettings();

	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_BACKUP_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_BACKUP_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	const lastRunAt = status?.ended_at ?? status?.started_at;

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_BACKUP_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_BACKUP_DESCRIPTION')}</p>
			</div>
			{hasError ? <Alert severity="error">{t('SETTINGS_BACKUP_LOAD_ERROR')}</Alert> : null}
			<div className={styles.summary}>
				{status?.has_run ? (
					<Chip
						variant="outlined"
						label={`${t('SETTINGS_BACKUP_LAST_RUN')}: ${t(backupStatusKey(status.status))}${
							lastRunAt ? ` · ${new Date(lastRunAt).toLocaleString()}` : ''
						}`}
					/>
				) : (
					<Chip variant="outlined" label={t('SETTINGS_BACKUP_NEVER_RAN')} />
				)}
				{pendingFiles !== undefined ? (
					<Chip
						variant="outlined"
						label={t('SETTINGS_BACKUP_PENDING_FILES', { count: String(pendingFiles) })}
					/>
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
				label={t('SETTINGS_BACKUP_ENABLED')}
			/>
			<TextField
				size="small"
				fullWidth
				label={t('SETTINGS_BACKUP_DESTINATION')}
				value={form.destination_path}
				onChange={(event) => setField('destination_path', event.target.value)}
				disabled={isSaving}
			/>
			<div className={styles.numbers}>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_BACKUP_RETENTION_DAYS')}
					value={String(form.retention_days)}
					onChange={(event) => setField('retention_days', Number(event.target.value))}
					disabled={isSaving}
				/>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_BACKUP_INTERVAL_HOURS')}
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
					{t('SETTINGS_BACKUP_SAVE')}
				</Button>
			</div>
			<Alert severity="info">{t('SETTINGS_BACKUP_HELP')}</Alert>
		</section>
	);
};

export default BackupSettingsSection;
