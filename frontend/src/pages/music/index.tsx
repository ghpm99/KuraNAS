import './music.css';
import ActionBar from '@/components/actionBar';
import Tabs from '@/components/tabs';
import MusicContent from '@/components/musicContent';

const MusicPage = () => {
	return (
		<div className='content'>
			<ActionBar />
			<Tabs />
			<MusicContent />
		</div>
	);
};

export default MusicPage;
