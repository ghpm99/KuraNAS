import { appRoutes } from '@/app/routes';
import {
    BookImage,
    DownloadCloud,
    Film,
    House,
    Info,
    LayoutGrid,
    MessageSquare,
    Music,
    Settings,
    Star,
    Trash2,
    Upload,
    Videotape,
} from 'lucide-react';
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
    descriptionKey: string;
};

export const navigationItems: NavigationItem[] = [
    {
        href: appRoutes.home,
        icon: <House size={20} />,
        labelKey: 'HOME',
        descriptionKey: 'HOME_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.files,
        icon: <LayoutGrid size={20} />,
        labelKey: 'FILES',
        descriptionKey: 'FILES_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.favorites,
        icon: <Star size={20} />,
        labelKey: 'STARRED_FILES',
        descriptionKey: 'FAVORITES_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.images,
        icon: <BookImage size={20} />,
        labelKey: 'NAV_IMAGES',
        descriptionKey: 'IMAGES_SECTION_LIBRARY_DESCRIPTION',
    },
    {
        href: appRoutes.music,
        icon: <Music size={20} />,
        labelKey: 'NAV_MUSIC',
        descriptionKey: 'MUSIC_HOME_DESCRIPTION',
    },
    {
        href: appRoutes.videos,
        icon: <Videotape size={20} />,
        labelKey: 'NAV_VIDEOS',
        descriptionKey: 'VIDEO_SECTION_HOME_DESCRIPTION',
    },
    {
        href: appRoutes.assistant,
        icon: <MessageSquare size={20} />,
        labelKey: 'NAV_ASSISTANT',
        descriptionKey: 'GLOBAL_SEARCH_ACTION_ASSISTANT_DESCRIPTION',
    },
    {
        href: appRoutes.takeout,
        icon: <Upload size={20} />,
        labelKey: 'NAV_TAKEOUT',
        descriptionKey: 'TAKEOUT_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.captures,
        icon: <Film size={20} />,
        labelKey: 'NAV_CAPTURES',
        descriptionKey: 'CAPTURES_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.downloads,
        icon: <DownloadCloud size={20} />,
        labelKey: 'NAV_DOWNLOADS',
        descriptionKey: 'DOWNLOADS_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.trash,
        icon: <Trash2 size={20} />,
        labelKey: 'NAV_TRASH',
        descriptionKey: 'TRASH_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.analytics,
        icon: analyticsIcon,
        labelKey: 'ANALYTICS',
        descriptionKey: 'ANALYTICS_SECTION_OVERVIEW_DESCRIPTION',
    },
    {
        href: appRoutes.settings,
        icon: <Settings size={20} />,
        labelKey: 'SETTINGS',
        descriptionKey: 'SETTINGS_PAGE_DESCRIPTION',
    },
    {
        href: appRoutes.about,
        icon: <Info size={20} />,
        labelKey: 'ABOUT',
        descriptionKey: 'ABOUT_PAGE_DESCRIPTION',
    },
];
