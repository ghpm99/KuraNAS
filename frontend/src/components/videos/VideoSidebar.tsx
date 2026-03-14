import { Link } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoNavigation } from '@/components/videos/useVideoNavigation';
import styles from './VideoSidebar.module.css';

const VideoSidebar = () => {
	const { t } = useI18n();
	const { items } = useVideoNavigation();

	return (
		<nav className={styles.nav} aria-label={t('VIDEO_NAVIGATION_LABEL')}>
			<p className={styles.label}>{t('VIDEO_NAVIGATION_LABEL')}</p>
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

export default VideoSidebar;
