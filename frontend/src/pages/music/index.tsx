import MusicDomainHeader from '@/components/music/MusicDomainHeader';
import MusicLayout from '@/components/music/musicLayout';
import MusicContent from '@/components/musicContent';
import MusicSidebar from '@/components/music/MusicSidebar';
import styles from './music.module.css';

const MusicPage = () => {
	return (
		<MusicLayout>
			<div className={styles.page}>
				<MusicDomainHeader />
				<div className={styles.content}>
					<MusicSidebar />
					<MusicContent />
				</div>
			</div>
		</MusicLayout>
	);
};

export default MusicPage;
