import type { ReactNode } from 'react';
import Header from '@/components/layout/Header/Header';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import styles from './AppShell.module.css';
import { useAppShell } from './useAppShell';

interface AppShellProps {
	children: ReactNode;
}

export const AppShell = ({ children }: AppShellProps) => {
	const { hasQueue, showClock } = useAppShell();
	const scrollAreaClassName = hasQueue
		? `${styles.scrollArea} ${styles.scrollAreaWithPlayer}`
		: styles.scrollArea;

	return (
		<div className={styles.shell}>
			<div className={styles.sidebarPane}>
				<Sidebar />
			</div>
			<Header showClock={showClock} />
			<main className={styles.mainPane}>
				<div className={scrollAreaClassName}>{children}</div>
			</main>
		</div>
	);
};
