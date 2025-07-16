import { Bell, Clock, Search } from 'lucide-react';

import styles from './Header.module.css';
import useI18n from '@/components/i18n/provider/i18nContext';

interface HeaderProps {
	showClock?: boolean;
	currentTime?: Date;
}

export default function Header({ showClock = false, currentTime }: HeaderProps) {
	const { t } = useI18n();
	return (
		<header className={styles.header}>
			<div className={styles.searchContainer}>
				<Search className={styles.searchIcon} />
				<input type='search' placeholder={t('SEARCH_PLACEHOLDER')} className={styles.searchInput} />
			</div>
			<div className={styles.actions}>
				{showClock && currentTime && (
					<div className={styles.timeDisplay}>
						<Clock className={styles.icon} />
						<span>{currentTime.toLocaleTimeString()}</span>
					</div>
				)}
				<button className={styles.iconButton} title={t('NOTIFICATIONS')}>
					<Bell className={styles.icon} />
				</button>
				<div className={styles.avatar}>
					<img src='/avatar.jpg' alt={t('AVATAR_ALT')} width={32} height={32} />
				</div>
			</div>
		</header>
	);
}
