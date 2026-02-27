import { Avatar, Box, IconButton, InputAdornment, InputBase, Typography } from '@mui/material';
import { Bell, Clock, Search } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

interface HeaderProps {
	showClock?: boolean;
	currentTime?: Date;
}

export default function Header({ showClock = false, currentTime }: HeaderProps) {
	const { t } = useI18n();

	return (
		<Box
			component='header'
			sx={{
				gridArea: 'gh',
				display: 'flex',
				alignItems: 'center',
				justifyContent: 'space-between',
				px: 3,
				bgcolor: '#1c1a1f',
			}}
		>
			<InputBase
				type='search'
				placeholder={t('SEARCH_PLACEHOLDER')}
				startAdornment={
					<InputAdornment position='start'>
						<Search size={16} style={{ color: '#6b7280' }} />
					</InputAdornment>
				}
				sx={{
					width: 384,
					bgcolor: '#2a282e',
					borderRadius: 1,
					px: 1.5,
					py: 0.5,
					fontSize: '0.875rem',
				}}
			/>

			<Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
				{showClock && currentTime && (
					<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#2a282e', borderRadius: 1, px: 1.25, py: 0.75 }}>
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
	);
}
