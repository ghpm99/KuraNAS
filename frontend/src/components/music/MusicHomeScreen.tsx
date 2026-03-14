import { Button } from '@mui/material';
import { Link } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useMusicHomeScreen } from '@/components/music/useMusicHomeScreen';
import styles from './MusicHomeScreen.module.css';

const MusicHomeScreen = () => {
	const { t } = useI18n();
	const { currentTrackArtist, currentTrackTitle, hasQueue, queueCount, sections, status, totalTracks } = useMusicHomeScreen();

	return (
		<div className={styles.content}>
			<div className={styles.heroGrid}>
				<section className={styles.panel}>
					<span className={styles.panelLabel}>{t('MUSIC_HOME_QUEUE_LABEL')}</span>
					<h2 className={styles.panelTitle}>
						{hasQueue ? t('MUSIC_HOME_QUEUE_READY') : t('MUSIC_HOME_QUEUE_EMPTY')}
					</h2>
					<p className={styles.panelDescription}>{t('MUSIC_HOME_QUEUE_DESCRIPTION')}</p>
					<div className={styles.trackMeta}>
						<strong>{hasQueue ? currentTrackTitle : t('MUSIC_HOME_QUEUE_EMPTY_STATE')}</strong>
						<span>{hasQueue ? currentTrackArtist : t('MUSIC_HOME_QUEUE_EMPTY_HELP')}</span>
						<span className={styles.metricCaption}>{t('MUSIC_HOME_QUEUE_COUNT', { count: String(queueCount) })}</span>
					</div>
				</section>

				<section className={styles.panel}>
					<span className={styles.panelLabel}>{t('MUSIC_HOME_LIBRARY_LABEL')}</span>
					<h2 className={styles.panelTitle}>{t('MUSIC_HOME_LIBRARY_STATUS')}</h2>
					<p className={styles.panelDescription}>
						{status === 'pending' ? t('MUSIC_HOME_LIBRARY_LOADING') : t('MUSIC_HOME_LIBRARY_READY')}
					</p>
					<strong className={styles.metricValue}>{totalTracks}</strong>
					<span className={styles.metricCaption}>{t('MUSIC_HOME_LIBRARY_TRACKS')}</span>
				</section>
			</div>

			<section className={styles.sectionGrid} aria-label={t('MUSIC_NAVIGATION_LABEL')}>
				{sections.map((section) => (
					<article key={section.key} className={styles.sectionCard}>
						<h2 className={styles.sectionTitle}>{t(section.labelKey)}</h2>
						<p className={styles.sectionDescription}>{t(section.descriptionKey)}</p>
						<Button
							component={Link}
							to={section.href}
							variant='outlined'
							size='small'
							className={styles.sectionAction}
						>
							{t('MUSIC_HOME_OPEN_SECTION')}
						</Button>
					</article>
				))}
			</section>
		</div>
	);
};

export default MusicHomeScreen;
