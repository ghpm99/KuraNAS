import { Link } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useMusicNavigation } from '@/components/music/useMusicNavigation';
import styles from './MusicSidebar.module.css';

const MusicSidebar = () => {
	const { t } = useI18n();
	const { items } = useMusicNavigation();

	return (
		<nav className={styles.nav} aria-label={t('MUSIC_NAVIGATION_LABEL')}>
			<p className={styles.label}>{t('MUSIC_NAVIGATION_LABEL')}</p>
			<div className={styles.list}>
				{items.map((item) => (
					<Link
						key={item.key}
						to={item.href}
						className={item.isActive ? `${styles.link} ${styles.linkActive}` : styles.link}
					>
						<span className={styles.title}>{t(item.labelKey)}</span>
						<span className={styles.description}>{t(item.descriptionKey)}</span>
					</Link>
				))}
			</div>
		</nav>
	);
};

export default MusicSidebar;
