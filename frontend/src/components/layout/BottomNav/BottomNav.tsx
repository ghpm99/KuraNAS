import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { BookImage, House, LayoutGrid, Menu, Music } from 'lucide-react';
import { NavLink } from 'react-router-dom';
import styles from './BottomNav.module.css';

interface BottomNavProps {
    onOpenMenu: () => void;
}

const buildNavItemClassName = ({ isActive }: { isActive: boolean }) =>
    `${styles.navItem} ${isActive ? styles.navItemActive : ''}`;

export const BottomNav = ({ onOpenMenu }: BottomNavProps) => {
    const { t } = useI18n();

    const linkItems = [
        { id: 'home', label: t('HOME'), icon: <House size={20} />, path: appRoutes.home },
        { id: 'files', label: t('FILES'), icon: <LayoutGrid size={20} />, path: appRoutes.files },
        { id: 'images', label: t('NAV_IMAGES'), icon: <BookImage size={20} />, path: appRoutes.images },
        { id: 'music', label: t('NAV_MUSIC'), icon: <Music size={20} />, path: appRoutes.music },
    ];

    return (
        <nav className={styles.bottomNav} aria-label={t('NAV_BOTTOM_LABEL')}>
            {linkItems.map((item) => (
                <NavLink key={item.id} to={item.path} className={buildNavItemClassName}>
                    {item.icon}
                    <span className={styles.navLabel}>{item.label}</span>
                </NavLink>
            ))}
            <button className={styles.navItem} onClick={onOpenMenu} type="button">
                <Menu size={20} />
                <span className={styles.navLabel}>{t('MENU')}</span>
            </button>
        </nav>
    );
};
