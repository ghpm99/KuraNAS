import { Outlet } from 'react-router-dom';
import styles from './MusicContent.module.css';

const MusicContent = () => {
    return (
        <div className={styles.content}>
            <Outlet />
        </div>
    );
};

export default MusicContent;
