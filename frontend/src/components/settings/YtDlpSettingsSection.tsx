import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Link from '@mui/material/Link';
import useYtDlpSettings from './useYtDlpSettings';
import styles from './YtDlpSettingsSection.module.css';

type YtDlpSettingsSectionProps = {
	className?: string;
};

const YtDlpSettingsSection = ({ className = '' }: YtDlpSettingsSectionProps) => {
	const { t, status, isLoading, hasError, isUpdating, handleUpdate } = useYtDlpSettings();
	const sectionClassName = `${className} ${styles.section}`.trim();

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_YTDLP_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_YTDLP_DESCRIPTION')}</p>
			</div>

			{isLoading ? <CircularProgress size={20} /> : null}
			{hasError ? <Alert severity="error">{t('SETTINGS_YTDLP_LOAD_ERROR')}</Alert> : null}

			{!isLoading && !hasError && status ? (
				<>
					<div className={styles.summary}>
						<Chip
							variant="outlined"
							label={`${t('SETTINGS_YTDLP_CURRENT')}: ${
								status.installed ? status.current_version : t('SETTINGS_YTDLP_NOT_INSTALLED')
							}`}
						/>
						{status.latest_version ? (
							<Chip variant="outlined" label={`${t('SETTINGS_YTDLP_LATEST')}: ${status.latest_version}`} />
						) : null}
						{status.update_available ? (
							<Chip color="warning" label={t('SETTINGS_YTDLP_UPDATE_AVAILABLE')} />
						) : (
							<Chip color="success" variant="outlined" label={t('SETTINGS_YTDLP_UP_TO_DATE')} />
						)}
					</div>

					<div className={styles.actions}>
						<Button
							variant="contained"
							disabled={!status.update_available || isUpdating}
							onClick={handleUpdate}
							startIcon={isUpdating ? <CircularProgress size={16} color="inherit" /> : undefined}
						>
							{isUpdating ? t('SETTINGS_YTDLP_UPDATING') : t('SETTINGS_YTDLP_UPDATE_BUTTON')}
						</Button>
						{status.release_url ? (
							<Link href={status.release_url} target="_blank" rel="noreferrer">
								{t('SETTINGS_YTDLP_RELEASE_NOTES')}
							</Link>
						) : null}
					</div>
				</>
			) : null}
		</section>
	);
};

export default YtDlpSettingsSection;
