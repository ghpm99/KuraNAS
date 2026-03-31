import Alert from '@mui/material/Alert';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import TextField from '@mui/material/TextField';
import type { LibraryCategory } from '@/types/libraries';
import useLibrarySettings from './useLibrarySettings';
import styles from './LibrarySettingsSection.module.css';

type LibrarySettingsSectionProps = {
	className?: string;
};

const getCategoryBadge = (category: LibraryCategory) => {
	switch (category) {
	case 'images':
		return 'I'
	case 'music':
		return 'M'
	case 'videos':
		return 'V'
	case 'documents':
		return 'D'
	default:
		return '?'
	}
};

const LibrarySettingsSection = ({ className = '' }: LibrarySettingsSectionProps) => {
	const { t, libraries, isLoading, isSaving, hasError, setPath, handleSave, getCategoryLabel } =
		useLibrarySettings()

	const sectionClassName = `${className} ${styles.section}`.trim()

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_LIBRARIES_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_LIBRARIES_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		)
	}

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_LIBRARIES_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_LIBRARIES_DESCRIPTION')}</p>
			</div>
			{hasError ? <Alert severity="error">{t('SETTINGS_LIBRARY_SAVE_ERROR')}</Alert> : null}
			{libraries.map((library) => (
				<div key={library.category} className={styles.row}>
					<div className={styles.category}>
						<Avatar className={styles.badge}>{getCategoryBadge(library.category)}</Avatar>
						<span>{getCategoryLabel(library.category)}</span>
					</div>
					<TextField
						className={styles.input}
						size="small"
						label={t('SETTINGS_LIBRARY_PATH_LABEL')}
						value={library.path}
						onChange={(event) => setPath(library.category, event.target.value)}
						disabled={isSaving}
					/>
					<Button
						variant="outlined"
						onClick={() => void handleSave(library.category)}
						disabled={isSaving || library.path.trim().length === 0}
					>
						{t('SETTINGS_LIBRARY_SAVE')}
					</Button>
				</div>
			))}
		</section>
	)
}

export default LibrarySettingsSection
