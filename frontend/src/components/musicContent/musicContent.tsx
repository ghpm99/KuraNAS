import { Outlet } from 'react-router-dom';
import QueueDrawer from '@/components/playlist/QueueDrawer';
import styles from './MusicContent.module.css';

const MusicContent = () => {
	return (
		<>
			<div className={styles.content}>
				<Outlet />
			</div>
			<QueueDrawer />
		</>
	);
};

export default MusicContent;
