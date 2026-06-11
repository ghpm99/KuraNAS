import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import useAccessControlSettings from './useAccessControlSettings';
import styles from './AccessControlSettingsSection.module.css';

type AccessControlSettingsSectionProps = {
	className?: string;
};

const AccessControlSettingsSection = ({ className = '' }: AccessControlSettingsSectionProps) => {
	const {
		t,
		entries,
		clientIP,
		isLoading,
		isSaving,
		hasError,
		cidr,
		label,
		setCidr,
		setLabel,
		handleAdd,
		handleAddCurrentDevice,
		handleToggle,
		handleDelete,
	} = useAccessControlSettings()

	const sectionClassName = `${className} ${styles.section}`.trim()

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_ACCESS_CONTROL_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_ACCESS_CONTROL_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		)
	}

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_ACCESS_CONTROL_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_ACCESS_CONTROL_DESCRIPTION')}</p>
			</div>
			{hasError ? (
				<Alert severity="error">{t('SETTINGS_ACCESS_CONTROL_LOAD_ERROR')}</Alert>
			) : null}
			{clientIP ? (
				<div className={styles.clientIp}>
					<span>{t('SETTINGS_ACCESS_CONTROL_YOUR_IP', { ip: clientIP })}</span>
					<Button
						variant="outlined"
						size="small"
						onClick={() => void handleAddCurrentDevice()}
						disabled={isSaving}
					>
						{t('SETTINGS_ACCESS_CONTROL_ADD_CURRENT')}
					</Button>
				</div>
			) : null}
			<div className={styles.form}>
				<TextField
					size="small"
					label={t('SETTINGS_ACCESS_CONTROL_CIDR_LABEL')}
					value={cidr}
					onChange={(event) => setCidr(event.target.value)}
					disabled={isSaving}
				/>
				<TextField
					size="small"
					label={t('SETTINGS_ACCESS_CONTROL_LABEL_LABEL')}
					value={label}
					onChange={(event) => setLabel(event.target.value)}
					disabled={isSaving}
				/>
				<Button
					variant="contained"
					onClick={() => void handleAdd()}
					disabled={isSaving || cidr.trim().length === 0}
				>
					{t('SETTINGS_ACCESS_CONTROL_ADD')}
				</Button>
			</div>
			{entries.length === 0 ? (
				<Alert severity="warning">{t('SETTINGS_ACCESS_CONTROL_EMPTY')}</Alert>
			) : (
				entries.map((entry) => (
					<div key={entry.id} className={styles.row}>
						<div className={styles.entry}>
							<span className={styles.cidr}>{entry.cidr}</span>
							{entry.label ? <span className={styles.label}>{entry.label}</span> : null}
						</div>
						<Switch
							checked={entry.enabled}
							onChange={(_, checked) => void handleToggle(entry.id, checked)}
							disabled={isSaving}
							slotProps={{
								input: { 'aria-label': t('SETTINGS_ACCESS_CONTROL_ENABLED') },
							}}
						/>
						<Button
							variant="text"
							color="error"
							onClick={() => void handleDelete(entry.id)}
							disabled={isSaving}
						>
							{t('SETTINGS_ACCESS_CONTROL_REMOVE')}
						</Button>
					</div>
				))
			)}
		</section>
	)
}

export default AccessControlSettingsSection
