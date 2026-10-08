import { Search } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import useGlobalSearch from '@/components/search/useGlobalSearch';
import NotificationBell from '@/components/notifications/NotificationBell';
import styles from './Header.module.css';

export default function Header() {
    const { t } = useI18n();
    const { openSearch, shortcut } = useGlobalSearch();

    return (
        <div className={styles.wrapper}>
            <header className={styles.header}>
                <div className={styles.searchGroup}>
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
                    <NotificationBell className={styles.iconButton} />
                </div>
            </header>
        </div>
    );
}
