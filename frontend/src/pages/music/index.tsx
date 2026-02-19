import ImageContent from '@/components/imageContent';
import './music.css';
import ActionBar from '@/components/actionBar';
import Tabs from '@/components/tabs';

const ImagesPage = () => {
	return (
		<div className='content'>
			<ActionBar />
			<Tabs />
			<ImageContent />
		</div>
	);
};

export default ImagesPage;
