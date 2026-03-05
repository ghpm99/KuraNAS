import type { ReactNode } from 'react';

import Header from '../Header/Header';
import Sidebar from '../Sidebar/Sidebar';
import styles from './Layout.module.css';
import { useUI } from '@/components/providers/uiProvider/uiContext';
import { useActivityDiary } from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

interface LayoutProps {
	children: ReactNode;
}

export const Layout = ({ children }: LayoutProps) => {
	const { activePage } = useUI();
	const { currentTime } = useActivityDiary();
	const { hasQueue } = useGlobalMusic();

	const showClock = activePage === 'activity';

	return (
		<div className={styles.layout}>
			<div className={styles.sidebarHeader}>
				<h1 className='app-title'>KuraNAS</h1>
			</div>
			<Header showClock={showClock} currentTime={currentTime} />
			<Sidebar />
			<div className={styles.mainContent} style={hasQueue ? { paddingBottom: 80 } : undefined}>
				{children}
			</div>
		</div>
	);
};
