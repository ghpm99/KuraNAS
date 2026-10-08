import useI18n from '@/components/i18n/provider/i18nContext';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import { IconButton, List } from '@mui/material';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { navigationItems } from '@/components/layout/navigationItems';
import styles from './Sidebar.module.css';

interface SidebarProps {
    mobile?: boolean;
    isCollapsed?: boolean;
    onToggleCollapsed?: () => void;
    onNavigate?: () => void;
}

const Sidebar = ({
    mobile = false,
    isCollapsed = false,
    onToggleCollapsed,
    onNavigate,
}: SidebarProps) => {
    const { t } = useI18n();
    const isIconOnly = isCollapsed && !mobile;
    const canToggle = !mobile && onToggleCollapsed !== undefined;
    const toggleLabel = t(isIconOnly ? 'SIDEBAR_EXPAND' : 'SIDEBAR_COLLAPSE');
    const ToggleIcon = isIconOnly ? PanelLeftOpen : PanelLeftClose;
    const classNames = [styles.sidebar];
    if (mobile) classNames.push(styles.mobile);
    if (isIconOnly) classNames.push(styles.collapsed);

    return (
        <nav className={classNames.join(' ')} data-collapsed={isIconOnly ? 'true' : 'false'}>
            <div className={styles.brand}>
                <span className={styles.brandMark} aria-hidden="true" />
                <div className={styles.brandText}>
                    <p className={styles.brandTitle}>{t('APP_NAME')}</p>
                </div>
                {canToggle && (
                    <IconButton
                        size="small"
                        className={styles.collapseToggle}
                        onClick={onToggleCollapsed}
                        aria-label={toggleLabel}
                        aria-expanded={!isIconOnly}
                    >
                        <ToggleIcon size={18} />
                    </IconButton>
                )}
            </div>
            <List className={styles.navList} dense>
                {navigationItems.map((item) => (
                    <NavItem
                        key={item.href}
                        href={item.href}
                        icon={item.icon}
                        label={t(item.labelKey)}
                        isIconOnly={isIconOnly}
                        onClick={onNavigate}
                    />
                ))}
            </List>
        </nav>
    );
};

export default Sidebar;
