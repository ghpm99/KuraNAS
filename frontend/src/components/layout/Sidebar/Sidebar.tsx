import useI18n from '@/components/i18n/provider/i18nContext';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import { List } from '@mui/material';
import { navigationItems } from '@/components/layout/navigationItems';
import styles from './Sidebar.module.css';

interface SidebarProps {
    mobile?: boolean;
    onNavigate?: () => void;
}

const Sidebar = ({ mobile = false, onNavigate }: SidebarProps) => {
    const { t } = useI18n();
    const sidebarClassName = mobile ? `${styles.sidebar} ${styles.mobile}` : styles.sidebar;

    return (
        <nav className={sidebarClassName}>
            <div className={styles.brand}>
                <span className={styles.brandMark} aria-hidden="true" />
                <div className={styles.brandText}>
                    <p className={styles.brandTitle}>{t('APP_NAME')}</p>
                </div>
            </div>
            <List className={styles.navList} dense>
                {navigationItems.map((item) => (
                    <NavItem key={item.href} href={item.href} icon={item.icon} onClick={onNavigate}>
                        {t(item.labelKey)}
                    </NavItem>
                ))}
            </List>
        </nav>
    );
};

export default Sidebar;
