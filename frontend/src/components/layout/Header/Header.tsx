import { Avatar, Box, Drawer, IconButton, InputAdornment, InputBase, List, Typography } from '@mui/material';
import { Bell, BookImage, Clock, Info, LayoutGrid, Menu, Music, Search, Star, Videotape } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import NavItem from '@/components/layout/Sidebar/components/navItem';

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

interface HeaderProps {
	showClock?: boolean;
	currentTime?: Date;
}

export default function Header({ showClock = false, currentTime }: HeaderProps) {
	const { t } = useI18n();
	const [mobileOpen, setMobileOpen] = useState(false);

	return (
		<>
			<Box
				component='header'
				sx={{
					gridArea: 'gh',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'space-between',
					px: 3,
					bgcolor: '#141419',
					borderBottom: '1px solid',
					borderColor: 'divider',
				}}
			>
				<Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
					<IconButton
						onClick={() => setMobileOpen(true)}
						sx={{ display: { xs: 'inline-flex', md: 'none' } }}
						size='small'
					>
						<Menu size={20} />
					</IconButton>
					<InputBase
						type='search'
						placeholder={t('SEARCH_PLACEHOLDER')}
						startAdornment={
							<InputAdornment position='start'>
								<Search size={16} style={{ color: '#a1a1aa' }} />
							</InputAdornment>
						}
						sx={{
							width: { xs: 200, sm: 384 },
							bgcolor: 'rgba(255, 255, 255, 0.04)',
							border: '1px solid',
							borderColor: 'divider',
							borderRadius: 2.5,
							px: 1.5,
							py: 0.5,
							fontSize: '0.875rem',
							transition: 'border-color 0.2s ease',
							'&:hover': {
								borderColor: 'rgba(255, 255, 255, 0.12)',
							},
							'&.Mui-focused': {
								borderColor: 'primary.main',
							},
						}}
					/>
				</Box>

				<Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
					{showClock && currentTime && (
						<Box sx={{ display: { xs: 'none', sm: 'flex' }, alignItems: 'center', gap: 1, bgcolor: 'rgba(255, 255, 255, 0.04)', border: '1px solid', borderColor: 'divider', borderRadius: 2, px: 1.25, py: 0.75 }}>
							<Clock size={16} />
							<Typography variant='body2'>{currentTime.toLocaleTimeString()}</Typography>
						</Box>
					)}
					<IconButton title={t('NOTIFICATIONS')} size='small'>
						<Bell size={16} />
					</IconButton>
					<Avatar src='/avatar.jpg' alt={t('AVATAR_ALT')} sx={{ width: 32, height: 32 }} />
				</Box>
			</Box>

			<Drawer
				open={mobileOpen}
				onClose={() => setMobileOpen(false)}
				sx={{
					display: { xs: 'block', md: 'none' },
					'& .MuiDrawer-paper': {
						width: 280,
						bgcolor: '#141419',
						borderRight: '1px solid',
						borderColor: 'divider',
					},
				}}
			>
				<Box sx={{ p: 2, textAlign: 'center', borderBottom: '1px solid', borderColor: 'divider' }}>
					<Typography variant='h6' fontWeight={700}>KuraNAS</Typography>
				</Box>
				<List sx={{ px: 1, pt: 1 }} dense onClick={() => setMobileOpen(false)}>
					<NavItem href='/' icon={<LayoutGrid size={20} />}>{t('ALL_FILES')}</NavItem>
					<NavItem href='/images' icon={<BookImage size={20} />}>Imagens</NavItem>
					<NavItem href='/music' icon={<Music size={20} />}>Musicas</NavItem>
					<NavItem href='/videos' icon={<Videotape size={20} />}>Vídeos</NavItem>
					<NavItem href='/starred' icon={<Star size={20} />}>{t('STARRED_FILES')}</NavItem>
					<NavItem href='/activity-diary' icon={<ActivityIcon />}>{t('ACTIVITY_DIARY')}</NavItem>
					<NavItem href='/analytics' icon={<AnalyticsIcon />}>{t('ANALYTICS')}</NavItem>
					<NavItem href='/about' icon={<Info size={20} />}>{t('ABOUT')}</NavItem>
				</List>
			</Drawer>
		</>
	);
}
