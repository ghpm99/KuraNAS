import { LayoutGrid } from 'lucide-react';
import NavItem from './components/navItem';
import FolderTree from './components/folderTree';
import useI18n from '../i18n/provider/i18nContext';
import './sidebar.css';

const Sidebar = () => {
	const { t } = useI18n();
	return (
		<div className='sidebar'>
			<nav className='nav'>
				<NavItem href='#' icon={<LayoutGrid className='icon' />} active>
					{t('ALL_FILES')}
				</NavItem>
				<NavItem
					href='#'
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
					href='#'
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
};

export default Sidebar;
