import MusicLayout from '@/components/music/musicLayout';
import MusicContent from '@/components/musicContent';
import MusicSidebar from '@/components/music/MusicSidebar';
import styles from './music.module.css';

const MusicPage = () => {
	return (
		<MusicLayout>
			<div className={styles['content']}>
				<div className={styles['music-views-sidebar']}>
					<MusicSidebar />
				</div>
				<MusicContent />
			</div>
		</MusicLayout>
	);
};

export default MusicPage;
