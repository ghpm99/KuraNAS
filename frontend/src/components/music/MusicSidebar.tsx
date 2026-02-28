import { List, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { Disc, Folder, ListMusic, Tag, User } from 'lucide-react';
import { MusicView } from '@/types/music';
import { useMusic } from '@/components/hooks/musicProvider/musicProvider';

const views: { key: MusicView; label: string; icon: React.ReactNode }[] = [
	{ key: 'all', label: 'All Tracks', icon: <ListMusic size={20} /> },
	{ key: 'artists', label: 'Artists', icon: <User size={20} /> },
	{ key: 'albums', label: 'Albums', icon: <Disc size={20} /> },
	{ key: 'genres', label: 'Genres', icon: <Tag size={20} /> },
	{ key: 'folders', label: 'Folders', icon: <Folder size={20} /> },
];

const MusicSidebar = () => {
	const { currentView, setCurrentView } = useMusic();

	return (
		<List sx={{ py: 0 }}>
			{views.map((view) => (
				<ListItemButton
					key={view.key}
					selected={currentView === view.key}
					onClick={() => setCurrentView(view.key)}
					sx={{ borderRadius: 1, mb: 0.5 }}
				>
					<ListItemIcon sx={{ minWidth: 36 }}>{view.icon}</ListItemIcon>
					<ListItemText primary={view.label} primaryTypographyProps={{ variant: 'body2' }} />
				</ListItemButton>
			))}
		</List>
	);
};

export default MusicSidebar;
