import type { ReactNode } from 'react';
import { useState } from 'react';
import Header from '@/components/layout/Header/Header';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import styles from './AppShell.module.css';
import { useAppShell } from './useAppShell';
import { Drawer } from '@mui/material';
import { BottomNav } from '@/components/layout/BottomNav/BottomNav';

interface AppShellProps {
    children: ReactNode;
}

export const AppShell = ({ children }: AppShellProps) => {
    const { hasQueue, showClock } = useAppShell();
    const [mobileOpen, setMobileOpen] = useState(false);
    
    const scrollAreaClassName = hasQueue
        ? `${styles.scrollArea} ${styles.scrollAreaWithPlayer}`
        : styles.scrollArea;

    const handleCloseMobileMenu = () => setMobileOpen(false);
    const handleOpenMobileMenu = () => setMobileOpen(true);

    return (
        <div className={styles.shell}>
            <div className={styles.sidebarPane}>
                <Sidebar />
            </div>
            <Header showClock={showClock} onOpenMobileMenu={handleOpenMobileMenu} />
            <main className={styles.mainPane}>
                <div className={scrollAreaClassName}>{children}</div>
            </main>
            
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
