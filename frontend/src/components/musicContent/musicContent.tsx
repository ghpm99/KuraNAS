import { useMusic } from '../providers/musicProvider/musicProvider';
import './musicContent.css';
import AllTracksView from '@/pages/music/views/AllTracksView';
import ArtistsView from '@/pages/music/views/ArtistsView';
import AlbumsView from '@/pages/music/views/AlbumsView';
import GenresView from '@/pages/music/views/GenresView';
import FoldersView from '@/pages/music/views/FoldersView';
import PlaylistsView from '@/pages/music/views/PlaylistsView';
import QueueDrawer from '@/components/playlist/QueueDrawer';

const MusicContent = () => {
	const { currentView } = useMusic();

	const renderView = () => {
		switch (currentView) {
			case 'artists':
				return <ArtistsView />;
			case 'albums':
				return <AlbumsView />;
			case 'genres':
				return <GenresView />;
			case 'folders':
				return <FoldersView />;
			case 'playlists':
				return <PlaylistsView />;
			case 'all':
			default:
				return <AllTracksView />;
		}
	};

	return (
		<>
			<div className='music-content'>
				{renderView()}
			</div>
			<QueueDrawer />
		</>
	);
};

export default MusicContent;
