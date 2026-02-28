import { List, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { Disc, Folder, ListMusic, ListPlus, Tag, User } from 'lucide-react';
import { MusicView } from '@/types/music';
import { useMusic } from '@/components/hooks/musicProvider/musicProvider';
import useI18n from '@/components/i18n/provider/i18nContext';

const viewKeys: { key: MusicView; labelKey: string; icon: React.ReactNode }[] = [
	{ key: 'all', labelKey: 'MUSIC_ALL_TRACKS', icon: <ListMusic size={20} /> },
	{ key: 'artists', labelKey: 'MUSIC_ARTISTS', icon: <User size={20} /> },
	{ key: 'albums', labelKey: 'MUSIC_ALBUMS', icon: <Disc size={20} /> },
	{ key: 'genres', labelKey: 'MUSIC_GENRES', icon: <Tag size={20} /> },
	{ key: 'folders', labelKey: 'MUSIC_FOLDERS', icon: <Folder size={20} /> },
	{ key: 'playlists', labelKey: 'MUSIC_PLAYLISTS', icon: <ListPlus size={20} /> },
];

const MusicSidebar = () => {
	const { currentView, setCurrentView } = useMusic();
	const { t } = useI18n();

	return (
		<List sx={{ py: 0 }}>
			{viewKeys.map((view) => (
				<ListItemButton
					key={view.key}
					selected={currentView === view.key}
					onClick={() => setCurrentView(view.key)}
					sx={{ borderRadius: 1, mb: 0.5 }}
				>
					<ListItemIcon sx={{ minWidth: 36 }}>{view.icon}</ListItemIcon>
					<ListItemText primary={t(view.labelKey)} primaryTypographyProps={{ variant: 'body2' }} />
				</ListItemButton>
			))}
		</List>
	);
};

export default MusicSidebar;
