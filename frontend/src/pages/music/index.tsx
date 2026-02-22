import MusicLayout from '@/components/music/musicLayout';
import MusicContent from '@/components/musicContent';
import './music.css';

const MusicPage = () => {
	return (
		<MusicLayout>
			<div className='content'>
				<MusicContent />
			</div>
		</MusicLayout>
	);
};

export default MusicPage;
