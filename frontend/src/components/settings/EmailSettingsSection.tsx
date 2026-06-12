import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Link from '@mui/material/Link';
import Switch from '@mui/material/Switch';
import type { EmailAccountStatus, EmailDeviceCodeStatus } from '@/types/email';
import useEmailSettings from './useEmailSettings';
import styles from './EmailSettingsSection.module.css';

type EmailSettingsSectionProps = {
	className?: string;
};

const statusKeyByAccountStatus: Record<EmailAccountStatus, string> = {
	linked: 'SETTINGS_EMAIL_STATUS_LINKED',
	error: 'SETTINGS_EMAIL_STATUS_ERROR',
	reauth_required: 'SETTINGS_EMAIL_STATUS_REAUTH_REQUIRED',
};

const deviceStatusKeyByStatus: Record<EmailDeviceCodeStatus, string> = {
	idle: 'SETTINGS_EMAIL_DEVICE_PENDING',
	pending: 'SETTINGS_EMAIL_DEVICE_PENDING',
	linked: 'SETTINGS_EMAIL_DEVICE_LINKED',
	expired: 'SETTINGS_EMAIL_DEVICE_EXPIRED',
	error: 'SETTINGS_EMAIL_DEVICE_ERROR',
};

const EmailSettingsSection = ({ className = '' }: EmailSettingsSectionProps) => {
	const {
		t,
		accounts,
		isLoading,
		isSaving,
		hasError,
		loadErrorMessage,
		deviceCode,
		deviceStatus,
		handleLinkGoogle,
		handleLinkMicrosoft,
		handleToggleSync,
		handleRemove,
	} = useEmailSettings();

	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_EMAIL_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_EMAIL_HELP')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_EMAIL_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_EMAIL_HELP')}</p>
			</div>
			{hasError ? <Alert severity="error">{loadErrorMessage}</Alert> : null}
			<div className={styles.actions}>
				<Button variant="contained" onClick={() => void handleLinkGoogle()} disabled={isSaving}>
					{t('SETTINGS_EMAIL_ADD_GOOGLE')}
				</Button>
				<Button
					variant="contained"
					onClick={() => void handleLinkMicrosoft()}
					disabled={isSaving}
				>
					{t('SETTINGS_EMAIL_ADD_MICROSOFT')}
				</Button>
			</div>
			<Alert severity="info">{t('SETTINGS_EMAIL_GOOGLE_HINT')}</Alert>
			{deviceCode && deviceStatus ? (
				<Alert severity={deviceStatus === 'linked' ? 'success' : 'info'}>
					<div className={styles.deviceCode}>
						{/* The prompt arrives already translated from the backend. */}
						<span>{deviceCode.message}</span>
						<span className={styles.userCode}>{deviceCode.user_code}</span>
						<Link href={deviceCode.verification_uri} target="_blank" rel="noopener">
							{deviceCode.verification_uri}
						</Link>
						<span>{t(deviceStatusKeyByStatus[deviceStatus])}</span>
					</div>
				</Alert>
			) : null}
			{accounts.length === 0 ? (
				<Alert severity="warning">{t('SETTINGS_EMAIL_NO_ACCOUNTS')}</Alert>
			) : (
				accounts.map((account) => (
					<div key={account.id} className={styles.row}>
						<div className={styles.entry}>
							<span className={styles.address}>
								{account.address}
								<Chip size="small" variant="outlined" label={account.provider} />
								<Chip
									size="small"
									color={account.status === 'linked' ? 'success' : 'warning'}
									variant="outlined"
									label={t(statusKeyByAccountStatus[account.status])}
								/>
							</span>
							{account.last_error ? (
								<span className={styles.lastError}>{account.last_error}</span>
							) : null}
						</div>
						<Switch
							checked={account.sync_enabled}
							onChange={(_, checked) => void handleToggleSync(account.id, checked)}
							disabled={isSaving}
							slotProps={{
								input: { 'aria-label': t('SETTINGS_EMAIL_SYNC_ENABLED') },
							}}
						/>
						<Button
							variant="text"
							color="error"
							onClick={() => void handleRemove(account.id)}
							disabled={isSaving}
						>
							{t('SETTINGS_EMAIL_REMOVE')}
						</Button>
					</div>
				))
			)}
		</section>
	);
};

export default EmailSettingsSection;
