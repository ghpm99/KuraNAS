import { Box, IconButton, Typography } from '@mui/material';
import { ArrowLeft, Play, Shuffle } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

interface CategoryHeaderProps {
	title: string;
	subtitle?: string;
	trackCount: number;
	icon: React.ReactNode;
	gradientFrom?: string;
	gradientTo?: string;
	onBack: () => void;
	onPlayAll: () => void;
	onShuffleAll: () => void;
}

const CategoryHeader = ({
	title,
	subtitle,
	trackCount,
	icon,
	gradientFrom = '#4f46e5',
	gradientTo = '#1a1a24',
	onBack,
	onPlayAll,
	onShuffleAll,
}: CategoryHeaderProps) => {
	const { t } = useI18n();

	return (
		<Box
			sx={{
				background: `linear-gradient(180deg, ${gradientFrom}33 0%, ${gradientTo} 100%)`,
				borderRadius: 3,
				p: 2.5,
				mb: 2,
			}}
		>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
				<IconButton onClick={onBack} size='small' sx={{ color: 'text.primary' }}>
					<ArrowLeft size={20} />
				</IconButton>
			</Box>

			<Box sx={{ display: 'flex', alignItems: 'flex-end', gap: 2.5 }}>
				<Box
					sx={{
						width: 120,
						height: 120,
						borderRadius: 2,
						bgcolor: `${gradientFrom}44`,
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						flexShrink: 0,
						boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
					}}
				>
					{icon}
				</Box>
				<Box sx={{ minWidth: 0, flex: 1 }}>
					<Typography variant='h5' fontWeight={700} noWrap sx={{ mb: 0.5 }}>
						{title}
					</Typography>
					{subtitle && (
						<Typography variant='body2' color='text.secondary' noWrap sx={{ mb: 0.5 }}>
							{subtitle}
						</Typography>
					)}
					<Typography variant='caption' color='text.secondary'>
						{trackCount} {t('MUSIC_TRACKS_COUNT')}
					</Typography>
				</Box>
			</Box>

			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mt: 2 }}>
				<IconButton
					onClick={onPlayAll}
					sx={{
						bgcolor: 'primary.main',
						color: 'white',
						width: 44,
						height: 44,
						'&:hover': { bgcolor: 'primary.light', transform: 'scale(1.05)' },
						transition: 'all 0.2s ease',
					}}
				>
					<Play size={22} fill='white' />
				</IconButton>
				<IconButton
					onClick={onShuffleAll}
					sx={{
						color: 'text.secondary',
						'&:hover': { color: 'text.primary' },
					}}
				>
					<Shuffle size={20} />
				</IconButton>
			</Box>
		</Box>
	);
};

export default CategoryHeader;
