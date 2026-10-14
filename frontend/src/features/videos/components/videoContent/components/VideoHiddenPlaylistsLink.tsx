import { Button } from '@mui/material';
import { Link } from 'react-router-dom';
import { getVideoRoute } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from '../videoContent.module.css';

export default function VideoHiddenPlaylistsLink() {
    const { t } = useI18n();

    return (
        <div className={styles.hiddenPlaylistsLinkRow}>
            <Button component={Link} to={getVideoRoute('hidden')} variant="text" size="small">
                {t('VIDEO_HIDDEN_PLAYLISTS_LINK')}
            </Button>
        </div>
    );
}
