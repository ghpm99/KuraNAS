import { useMusic } from '../hooks/musicProvider/musicProvider';
import './musicContent.css';
import Playlist from '../playlist/paylist';
import AllTracksView from '@/pages/music/views/AllTracksView';
import ArtistsView from '@/pages/music/views/ArtistsView';
import AlbumsView from '@/pages/music/views/AlbumsView';
import GenresView from '@/pages/music/views/GenresView';
import FoldersView from '@/pages/music/views/FoldersView';
import { useGlobalMusic } from '../providers/GlobalMusicProvider';

const MusicContent = () => {
	const { currentView } = useMusic();
	const { hasQueue } = useGlobalMusic();

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
			{hasQueue && <Playlist />}
		</>
	);
};

export default MusicContent;
