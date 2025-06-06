'use client';

import type { ReactNode } from 'react';

import Header from '../Header/Header';
import Sidebar from '../Sidebar/Sidebar';
import styles from './Layout.module.css';
import { useUI } from '@/components/hooks/UI/uiContext';
import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';

interface LayoutProps {
	children: ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
	const { activePage } = useUI();
	const { currentTime } = useActivityDiary();

	const showClock = activePage === 'activity';

	return (
		<div className={styles.layout}>
			<div className='sidebar-header'>
				<h1 className='app-title'>KuraNAS</h1>
			</div>
			<Header showClock={showClock} currentTime={currentTime} />
			<Sidebar />
			{children}
		</div>
	);
};

export default Layout;
