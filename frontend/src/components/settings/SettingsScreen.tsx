import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Alert, Button } from '@mui/material';
import { Link } from 'react-router-dom';
import styles from './SettingsScreen.module.css';

const SettingsScreen = () => {
	const { t } = useI18n();

	return (
		<div className={styles.content}>
			<header className={styles.header}>
				<h1 className={styles.title}>{t('SETTINGS_PAGE_TITLE')}</h1>
				<p className={styles.description}>{t('SETTINGS_PAGE_DESCRIPTION')}</p>
			</header>

			<section className={styles.panel}>
				<Alert severity='info'>{t('SETTINGS_PAGE_INFO')}</Alert>
				<div className={styles.actions}>
					<Button component={Link} to={appRoutes.analytics} variant='contained'>{t('ANALYTICS')}</Button>
					<Button component={Link} to={appRoutes.about} variant='outlined'>{t('ABOUT')}</Button>
					<Button component={Link} to={appRoutes.activityDiary} variant='text'>{t('ACTIVITY_DIARY')}</Button>
				</div>
			</section>
		</div>
	);
};

export default SettingsScreen;
