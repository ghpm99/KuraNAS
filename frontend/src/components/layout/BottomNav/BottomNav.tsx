import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { BookImage, House, LayoutGrid, Menu, Music } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import styles from './BottomNav.module.css';

interface BottomNavProps {
    onOpenMenu: () => void;
}

export const BottomNav = ({ onOpenMenu }: BottomNavProps) => {
    const { t } = useI18n();
    const navigate = useNavigate();
    const location = useLocation();

    const items = [
        {
            id: 'home',
            label: t('HOME'),
            icon: <House size={20} />,
            path: appRoutes.home,
            action: () => navigate(appRoutes.home),
            isActive: location.pathname === appRoutes.home,
        },
        {
            id: 'files',
            label: t('FILES'),
            icon: <LayoutGrid size={20} />,
            path: appRoutes.files,
            action: () => navigate(appRoutes.files),
            isActive: location.pathname.startsWith(appRoutes.files),
        },
        {
            id: 'images',
            label: t('NAV_IMAGES'),
            icon: <BookImage size={20} />,
            path: appRoutes.images,
            action: () => navigate(appRoutes.images),
            isActive: location.pathname.startsWith(appRoutes.images),
        },
        {
            id: 'music',
            label: t('NAV_MUSIC'),
            icon: <Music size={20} />,
            path: appRoutes.music,
            action: () => navigate(appRoutes.music),
            isActive: location.pathname.startsWith(appRoutes.music),
        },
        {
            id: 'menu',
            label: t('MENU'),
            icon: <Menu size={20} />,
            path: '#menu',
            action: onOpenMenu,
            isActive: false,
        },
    ];

    return (
        <nav className={styles.bottomNav}>
            {items.map((item) => (
                <button
                    key={item.id}
                    className={`${styles.navItem} ${item.isActive ? styles.navItemActive : ''}`}
                    onClick={item.action}
                    type="button"
                >
                    {item.icon}
                    <span className={styles.navLabel}>{item.label}</span>
                </button>
            ))}
        </nav>
    );
};
