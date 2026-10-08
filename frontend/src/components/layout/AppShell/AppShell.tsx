import type { ReactNode } from 'react';
import { useRef, useState } from 'react';
import Header from '@/components/layout/Header/Header';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import styles from './AppShell.module.css';
import { useAppShell } from './useAppShell';
import { Drawer } from '@mui/material';
import { BottomNav } from '@/components/layout/BottomNav/BottomNav';
import { BackToTopButton } from './BackToTopButton';
import { useScrollRestoration } from './useScrollRestoration';
import { useSidebarCollapse } from './useSidebarCollapse';

interface AppShellProps {
    children: ReactNode;
    banner?: ReactNode;
}

export const AppShell = ({ children, banner }: AppShellProps) => {
    const { hasQueue, showClock } = useAppShell();
    const [mobileOpen, setMobileOpen] = useState(false);
    const scrollAreaRef = useRef<HTMLDivElement>(null);
    useScrollRestoration(scrollAreaRef);
    const { isCollapsed, toggleCollapsed } = useSidebarCollapse();

    const scrollAreaClassName = hasQueue
        ? `${styles.scrollArea} ${styles.scrollAreaWithPlayer}`
        : styles.scrollArea;

    const handleCloseMobileMenu = () => setMobileOpen(false);
    const handleOpenMobileMenu = () => setMobileOpen(true);

    return (
        <div className={styles.shell} data-has-player={hasQueue ? 'true' : 'false'}>
            <div className={styles.sidebarPane}>
                <Sidebar isCollapsed={isCollapsed} onToggleCollapsed={toggleCollapsed} />
            </div>
            <Header showClock={showClock} onOpenMobileMenu={handleOpenMobileMenu} />
            <main className={styles.mainPane}>
                {banner}
                <div ref={scrollAreaRef} className={scrollAreaClassName}>
                    {children}
                </div>
            </main>
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
