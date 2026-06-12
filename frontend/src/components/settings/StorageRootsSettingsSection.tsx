import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import useStorageRootsSettings from './useStorageRootsSettings';
import styles from './StorageRootsSettingsSection.module.css';

type StorageRootsSettingsSectionProps = {
	className?: string;
};

const StorageRootsSettingsSection = ({ className = '' }: StorageRootsSettingsSectionProps) => {
	const {
		t,
		roots,
		primaryRootId,
		isLoading,
		isSaving,
		hasError,
		path,
		label,
		setPath,
		setLabel,
		handleAdd,
		handleToggle,
		handleDelete,
	} = useStorageRootsSettings();

	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_STORAGE_ROOTS_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_STORAGE_ROOTS_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_STORAGE_ROOTS_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_STORAGE_ROOTS_DESCRIPTION')}</p>
			</div>
			{hasError ? (
				<Alert severity="error">{t('SETTINGS_STORAGE_ROOTS_LOAD_ERROR')}</Alert>
			) : null}
			<div className={styles.form}>
				<TextField
					size="small"
					label={t('SETTINGS_STORAGE_ROOTS_PATH_LABEL')}
					value={path}
					onChange={(event) => setPath(event.target.value)}
					disabled={isSaving}
				/>
				<TextField
					size="small"
					label={t('SETTINGS_STORAGE_ROOTS_LABEL_LABEL')}
					value={label}
					onChange={(event) => setLabel(event.target.value)}
					disabled={isSaving}
				/>
				<Button
					variant="contained"
					onClick={() => void handleAdd()}
					disabled={isSaving || path.trim().length === 0}
				>
					{t('SETTINGS_STORAGE_ROOTS_ADD')}
				</Button>
			</div>
			{roots.length === 0 ? (
				<Alert severity="warning">{t('SETTINGS_STORAGE_ROOTS_EMPTY')}</Alert>
			) : (
				roots.map((root) => {
					const isPrimary = root.id === primaryRootId;
					return (
						<div key={root.id} className={styles.row}>
							<div className={styles.entry}>
								<span className={styles.label}>
									{root.label}
									{isPrimary ? (
										<Chip
											size="small"
											variant="outlined"
											label={t('SETTINGS_STORAGE_ROOTS_PRIMARY')}
											className={styles.primaryChip}
										/>
									) : null}
								</span>
								<span className={styles.path}>{root.path}</span>
							</div>
							<Switch
								checked={root.enabled}
								onChange={(_, checked) => void handleToggle(root.id, checked)}
								disabled={isSaving || isPrimary}
								slotProps={{
									input: { 'aria-label': t('SETTINGS_STORAGE_ROOTS_ENABLED') },
								}}
							/>
							<Button
								variant="text"
								color="error"
								onClick={() => void handleDelete(root.id)}
								disabled={isSaving || isPrimary}
							>
								{t('SETTINGS_STORAGE_ROOTS_REMOVE')}
							</Button>
						</div>
					);
				})
			)}
			<Alert severity="info">{t('SETTINGS_STORAGE_ROOTS_HELP')}</Alert>
		</section>
	);
};

export default StorageRootsSettingsSection;
