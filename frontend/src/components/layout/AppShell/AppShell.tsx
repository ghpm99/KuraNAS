import type { MouseEvent, ReactNode } from 'react';
import { useRef, useState } from 'react';
import Header from '@/components/layout/Header/Header';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import styles from './AppShell.module.css';
import { useAppShell } from './useAppShell';
import { Drawer } from '@mui/material';
import { BottomNav } from '@/components/layout/BottomNav/BottomNav';
import { BackToTopButton } from './BackToTopButton';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useRouteFocus } from './useRouteFocus';
import { useScrollRestoration } from './useScrollRestoration';
import { useSidebarCollapse } from './useSidebarCollapse';

const mainContentId = 'main-content';

interface AppShellProps {
    children: ReactNode;
    banner?: ReactNode;
}

export const AppShell = ({ children, banner }: AppShellProps) => {
    const { t } = useI18n();
    const { hasQueue, showClock } = useAppShell();
    const [mobileOpen, setMobileOpen] = useState(false);
    const scrollAreaRef = useRef<HTMLDivElement>(null);
    const mainRef = useRef<HTMLElement>(null);
    const announcerRef = useRef<HTMLDivElement>(null);
    useRouteFocus(mainRef, announcerRef);
    useScrollRestoration(scrollAreaRef);
    const { isCollapsed, toggleCollapsed } = useSidebarCollapse();

    const scrollAreaClassName = hasQueue
        ? `${styles.scrollArea} ${styles.scrollAreaWithPlayer}`
        : styles.scrollArea;

    const handleCloseMobileMenu = () => setMobileOpen(false);
    const handleOpenMobileMenu = () => setMobileOpen(true);
    const handleSkipToContent = (event: MouseEvent<HTMLAnchorElement>) => {
        event.preventDefault();
        mainRef.current?.focus();
    };

    return (
        <div className={styles.shell} data-has-player={hasQueue ? 'true' : 'false'}>
            <a href={`#${mainContentId}`} className={styles.skipLink} onClick={handleSkipToContent}>
                {t('SKIP_TO_CONTENT')}
            </a>
            <div className={styles.sidebarPane}>
                <Sidebar isCollapsed={isCollapsed} onToggleCollapsed={toggleCollapsed} />
            </div>
            <Header showClock={showClock} onOpenMobileMenu={handleOpenMobileMenu} />
            <main id={mainContentId} ref={mainRef} tabIndex={-1} className={styles.mainPane}>
                {banner}
                <div ref={scrollAreaRef} className={scrollAreaClassName}>
                    {children}
                </div>
            </main>
            <div ref={announcerRef} aria-live="polite" className={styles.visuallyHidden} />
            <BackToTopButton scrollElementRef={scrollAreaRef} isPlayerVisible={hasQueue} />

            <BottomNav onOpenMenu={handleOpenMobileMenu} />

            <Drawer
                open={mobileOpen}
                onClose={handleCloseMobileMenu}
                PaperProps={{ className: styles.drawerPaper }}
            >
                <Sidebar mobile onNavigate={handleCloseMobileMenu} />
            </Drawer>
        </div>
    );
};
