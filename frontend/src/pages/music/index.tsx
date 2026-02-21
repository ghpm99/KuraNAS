import './music.css';
import ActionBar from '@/components/actionBar';
import Tabs from '@/components/tabs';
import MusicContent from '@/components/musicContent';
import { MusicProvider } from '@/components/hooks/musicProvider/musicProvider';

const MusicPage = () => {
	return (
		<MusicProvider>
			<div className='content'>
				<ActionBar />
				<Tabs />
				<MusicContent />
			</div>
		</MusicProvider>
	);
};

export default MusicPage;
