import { Box, List, ListItemButton, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { Disc, Folder, ListMusic, ListPlus, Tag, User } from 'lucide-react';
import { MusicView } from '@/types/music';
import { useMusic } from '@/components/providers/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';

const viewKeys: { key: MusicView; labelKey: string; icon: React.ReactNode }[] = [
	{ key: 'all', labelKey: 'MUSIC_ALL_TRACKS', icon: <ListMusic size={18} /> },
	{ key: 'artists', labelKey: 'MUSIC_ARTISTS', icon: <User size={18} /> },
	{ key: 'albums', labelKey: 'MUSIC_ALBUMS', icon: <Disc size={18} /> },
	{ key: 'genres', labelKey: 'MUSIC_GENRES', icon: <Tag size={18} /> },
	{ key: 'folders', labelKey: 'MUSIC_FOLDERS', icon: <Folder size={18} /> },
	{ key: 'playlists', labelKey: 'MUSIC_PLAYLISTS', icon: <ListPlus size={18} /> },
];

const MusicSidebar = () => {
	const { currentView, setCurrentView } = useMusic();
	const { t } = useI18n();

	return (
		<Box>
			<Typography
				variant='overline'
				color='text.secondary'
				sx={{ px: 1.5, pt: 1, display: 'block', fontSize: '0.65rem', letterSpacing: 1.5 }}
			>
				Library
			</Typography>
			<List sx={{ py: 0.5 }}>
				{viewKeys.map((view) => {
					const isSelected = currentView === view.key;
					return (
						<ListItemButton
							key={view.key}
							selected={isSelected}
							onClick={() => setCurrentView(view.key)}
							sx={{
								borderRadius: 1.5,
								mb: 0.25,
								py: 0.75,
								pl: 1.5,
								position: 'relative',
								'&::before': isSelected
									? {
											content: '""',
											position: 'absolute',
											left: 0,
											top: '25%',
											height: '50%',
											width: 3,
											borderRadius: 2,
											bgcolor: 'primary.main',
										}
									: undefined,
							}}
						>
							<ListItemIcon sx={{ minWidth: 32, color: isSelected ? 'primary.main' : 'text.secondary' }}>
								{view.icon}
							</ListItemIcon>
							<ListItemText
								primary={t(view.labelKey)}
								primaryTypographyProps={{
									variant: 'body2',
									fontWeight: isSelected ? 600 : 400,
									color: isSelected ? 'text.primary' : 'text.secondary',
								}}
							/>
						</ListItemButton>
					);
				})}
			</List>
		</Box>
	);
};

export default MusicSidebar;
