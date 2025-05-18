'use client';

import { LayoutGrid } from 'lucide-react';

import useI18n from '@/components/i18n/provider/i18nContext';
import FolderTree from '@/components/layout/Sidebar/components/folderTree';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import styles from './Sidebar.module.css';

export default function Sidebar() {
	const { t } = useI18n();

	return (
		<div className={styles.sidebar}>
			<nav className={styles.nav}>
				<NavItem href='/' icon={<LayoutGrid className='icon' />}>
					{t('ALL_FILES')}
				</NavItem>
				<NavItem
					href='/activity-diary'
					icon={
						<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
							<path
								d='M15 3v18M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7m0-18H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h7m0-18v18'
								strokeWidth='2'
								strokeLinecap='round'
							/>
						</svg>
					}
				>
					{t('ACTIVITY_DIARY')}
				</NavItem>
				<NavItem
					href='/analytics'
					icon={
						<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
							<path
								d='M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2M9 5h6m-3 4v6m-3-3h6'
								strokeWidth='2'
								strokeLinecap='round'
								strokeLinejoin='round'
							/>
						</svg>
					}
				>
					Analytics
				</NavItem>
				<FolderTree />
			</nav>
		</div>
	);
}
