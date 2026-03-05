import { BookImage, Info, LayoutGrid, Music, Star, Videotape } from 'lucide-react';
import { useUI } from '@/components/providers/uiProvider/uiContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import FolderTree from '@/components/layout/Sidebar/components/folderTree';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import { Box, List } from '@mui/material';

const ActivityIcon = () => (
	<svg width={20} height={20} viewBox='0 0 24 24' fill='none' stroke='currentColor'>
		<path d='M15 3v18M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7m0-18H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h7m0-18v18' strokeWidth='2' strokeLinecap='round' />
	</svg>
);

const AnalyticsIcon = () => (
	<svg width={20} height={20} viewBox='0 0 24 24' fill='none' stroke='currentColor'>
		<path d='M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2M9 5h6m-3 4v6m-3-3h6' strokeWidth='2' strokeLinecap='round' strokeLinejoin='round' />
	</svg>
);

const Sidebar = () => {
	const { t } = useI18n();
	const { activePage } = useUI();

	return (
		<Box
			component='nav'
			sx={{
				gridArea: 'left-nav',
				height: '100%',
				bgcolor: '#141419',
				borderRight: '1px solid',
				borderColor: 'divider',
				display: 'flex',
				flexDirection: 'column',
				overflow: 'hidden',
				'@media (max-width: 900px)': { display: 'none' },
			}}
		>
			<List sx={{ px: 1, pt: 0.5, flexShrink: 0 }} dense>
				<NavItem href='/' icon={<LayoutGrid size={20} />}>{t('ALL_FILES')}</NavItem>
				<NavItem href='/images' icon={<BookImage size={20} />}>{t('NAV_IMAGES')}</NavItem>
				<NavItem href='/music' icon={<Music size={20} />}>{t('NAV_MUSIC')}</NavItem>
				<NavItem href='/videos' icon={<Videotape size={20} />}>{t('NAV_VIDEOS')}</NavItem>
				<NavItem href='/starred' icon={<Star size={20} />}>{t('STARRED_FILES')}</NavItem>
				<NavItem href='/activity-diary' icon={<ActivityIcon />}>{t('ACTIVITY_DIARY')}</NavItem>
				<NavItem href='/analytics' icon={<AnalyticsIcon />}>{t('ANALYTICS')}</NavItem>
				<NavItem href='/about' icon={<Info size={20} />}>{t('ABOUT')}</NavItem>
			</List>
			{activePage === 'files' && (
				<Box sx={{ flex: 1, overflowY: 'auto', minHeight: 0 }}>
					<FolderTree />
				</Box>
			)}
		</Box>
	);
};

export default Sidebar;
