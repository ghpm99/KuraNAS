import { Avatar, IconButton } from '@mui/material';
import { Clock3, Menu, Search } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import useGlobalSearch from '@/components/search/useGlobalSearch';
import NotificationBell from '@/components/notifications/NotificationBell';
import styles from './Header.module.css';
import { useHeader } from './useHeader';

interface HeaderProps {
    showClock?: boolean;
    onOpenMobileMenu: () => void;
}

export default function Header({ showClock = false, onOpenMobileMenu }: HeaderProps) {
    const { t } = useI18n();
    const { openSearch, shortcut } = useGlobalSearch();
    const { currentTime } = useHeader(showClock);

    return (
        <div className={styles.wrapper}>
            <header className={styles.header}>
                <div className={styles.searchGroup}>
                    <IconButton
                        onClick={onOpenMobileMenu}
                        className={styles.menuButton}
                        size="small"
                        title={t('OPEN_NAVIGATION_MENU')}
                        aria-label={t('OPEN_NAVIGATION_MENU')}
                    >
                        <Menu size={22} />
                    </IconButton>
                    <button
                        type="button"
                        className={styles.searchField}
                        onClick={openSearch}
                        aria-label={t('GLOBAL_SEARCH_OPEN')}
                    >
                        <Search size={16} className={styles.searchIcon} />
                        <span className={styles.searchPlaceholder}>
                            {t('SEARCH_PLACEHOLDER')}
                        </span>
                        <span className={styles.searchShortcut}>
                            {t('GLOBAL_SEARCH_SHORTCUT', { shortcut })}
                        </span>
                    </button>
                </div>

                <div className={styles.actions}>
                    {showClock && (
                        <div className={styles.clock}>
                            <Clock3 size={16} />
                            <span className={styles.clockLabel}>
                                {currentTime.toLocaleTimeString()}
                            </span>
                        </div>
                    )}

                    <NotificationBell className={styles.iconButton} />
                    <Avatar src="/avatar.jpg" alt={t('AVATAR_ALT')} className={styles.avatar} />
                </div>
            </header>
        </div>
    );
}
