import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Button } from '@mui/material';
import { Link } from 'react-router-dom';
import styles from './HomeScreen.module.css';

const HomeScreen = () => {
	const { t } = useI18n();

	return (
		<div className={styles.content}>
			<section className={styles.hero}>
				<h1 className={styles.heroTitle}>{t('HOME_PAGE_TITLE')}</h1>
				<p className={styles.heroDescription}>{t('HOME_PAGE_DESCRIPTION')}</p>
			</section>

			<section className={styles.grid}>
				<article className={styles.card}>
					<div>
						<h2 className={styles.cardTitle}>{t('HOME_LIBRARY_TITLE')}</h2>
						<p className={styles.cardDescription}>{t('HOME_LIBRARY_DESCRIPTION')}</p>
					</div>
					<div className={styles.actions}>
						<Button component={Link} to={appRoutes.files} variant='contained'>{t('FILES')}</Button>
						<Button component={Link} to={appRoutes.favorites} variant='outlined'>{t('STARRED_FILES')}</Button>
					</div>
				</article>

				<article className={styles.card}>
					<div>
						<h2 className={styles.cardTitle}>{t('HOME_MEDIA_TITLE')}</h2>
						<p className={styles.cardDescription}>{t('HOME_MEDIA_DESCRIPTION')}</p>
					</div>
					<div className={styles.actions}>
						<Button component={Link} to={appRoutes.images} variant='outlined'>{t('NAV_IMAGES')}</Button>
						<Button component={Link} to={appRoutes.music} variant='outlined'>{t('NAV_MUSIC')}</Button>
						<Button component={Link} to={appRoutes.videos} variant='outlined'>{t('NAV_VIDEOS')}</Button>
					</div>
				</article>

				<article className={styles.card}>
					<div>
						<h2 className={styles.cardTitle}>{t('HOME_SYSTEM_TITLE')}</h2>
						<p className={styles.cardDescription}>{t('HOME_SYSTEM_DESCRIPTION')}</p>
					</div>
					<div className={styles.actions}>
						<Button component={Link} to={appRoutes.analytics} variant='outlined'>{t('ANALYTICS')}</Button>
						<Button component={Link} to={appRoutes.settings} variant='contained'>{t('SETTINGS')}</Button>
						<Button component={Link} to={appRoutes.about} variant='outlined'>{t('ABOUT')}</Button>
						<Button component={Link} to={appRoutes.activityDiary} variant='text'>{t('ACTIVITY_DIARY')}</Button>
					</div>
				</article>
			</section>
		</div>
	);
};

export default HomeScreen;
