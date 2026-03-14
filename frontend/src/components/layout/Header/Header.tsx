import { Avatar, Drawer, IconButton, InputBase } from '@mui/material';
import { Bell, Clock3, Menu, Search } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import styles from './Header.module.css';
import { useHeader } from './useHeader';

interface HeaderProps {
	showClock?: boolean;
}

export default function Header({ showClock = false }: HeaderProps) {
	const { t } = useI18n();
	const { currentTime, mobileOpen, closeMobileMenu, openMobileMenu } = useHeader(showClock);

	return (
		<>
			<div className={styles.wrapper}>
				<header className={styles.header}>
					<div className={styles.searchGroup}>
						<IconButton
							onClick={openMobileMenu}
							className={styles.menuButton}
							size='small'
							title={t('OPEN_NAVIGATION_MENU')}
							aria-label={t('OPEN_NAVIGATION_MENU')}
						>
							<Menu size={18} />
						</IconButton>
						<div className={styles.searchField}>
							<Search size={16} className={styles.searchIcon} />
							<InputBase
								type='search'
								placeholder={t('SEARCH_PLACEHOLDER')}
								className={styles.searchInput}
							/>
						</div>
					</div>

					<div className={styles.actions}>
						{showClock && (
							<div className={styles.clock}>
								<Clock3 size={16} />
								<span className={styles.clockLabel}>{currentTime.toLocaleTimeString()}</span>
							</div>
						)}

						<IconButton
							title={t('NOTIFICATIONS')}
							aria-label={t('NOTIFICATIONS')}
							size='small'
							className={styles.iconButton}
						>
							<Bell size={16} />
						</IconButton>
						<Avatar src='/avatar.jpg' alt={t('AVATAR_ALT')} className={styles.avatar} />
					</div>
				</header>
			</div>

			<Drawer open={mobileOpen} onClose={closeMobileMenu} PaperProps={{ className: styles.drawerPaper }}>
				<Sidebar mobile onNavigate={closeMobileMenu} />
			</Drawer>
		</>
	);
}
