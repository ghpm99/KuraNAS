import { appRoutes } from '@/app/routes';
import { BookImage, House, Info, LayoutGrid, Music, Settings, Star, Videotape } from 'lucide-react';
import type { ReactNode } from 'react';

const analyticsIcon = (
    <svg width={20} height={20} viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path
            d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2M9 5h6m-3 4v6m-3-3h6"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export type NavigationItem = {
    href: string;
    icon: ReactNode;
    labelKey: string;
};

export const navigationItems: NavigationItem[] = [
    { href: appRoutes.home, icon: <House size={20} />, labelKey: 'HOME' },
    { href: appRoutes.files, icon: <LayoutGrid size={20} />, labelKey: 'FILES' },
    {
        href: appRoutes.favorites,
        icon: <Star size={20} />,
        labelKey: 'STARRED_FILES',
    },
    {
        href: appRoutes.images,
        icon: <BookImage size={20} />,
        labelKey: 'NAV_IMAGES',
    },
    { href: appRoutes.music, icon: <Music size={20} />, labelKey: 'NAV_MUSIC' },
    {
        href: appRoutes.videos,
        icon: <Videotape size={20} />,
        labelKey: 'NAV_VIDEOS',
    },
    { href: appRoutes.analytics, icon: analyticsIcon, labelKey: 'ANALYTICS' },
    {
        href: appRoutes.settings,
        icon: <Settings size={20} />,
        labelKey: 'SETTINGS',
    },
    { href: appRoutes.about, icon: <Info size={20} />, labelKey: 'ABOUT' },
];
