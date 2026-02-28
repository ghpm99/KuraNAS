import MusicLayout from '@/components/music/musicLayout';
import MusicContent from '@/components/musicContent';
import MusicSidebar from '@/components/music/MusicSidebar';
import './music.css';

const MusicPage = () => {
	return (
		<MusicLayout>
			<div className='content'>
				<div className='music-views-sidebar'>
					<MusicSidebar />
				</div>
				<MusicContent />
			</div>
		</MusicLayout>
	);
};

export default MusicPage;
