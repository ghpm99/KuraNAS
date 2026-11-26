import { Button } from '@mui/material';
import { FolderOpen, Play } from 'lucide-react';
import { Link } from 'react-router-dom';
import { buildVideoPlayerUrl } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useOpenLibraryFileLocation } from '@/shared/hooks/useOpenLibraryFileLocation';
import styles from './CapturesScreen.module.css';

type CapturePromotedActionsProps = {
    captureId: number;
    fileId: number;
};

const CapturePromotedActions = ({ captureId, fileId }: CapturePromotedActionsProps) => {
    const { t } = useI18n();
    const fileLocation = useOpenLibraryFileLocation();

    return (
        <>
            <Button
                size="small"
                component={Link}
                to={buildVideoPlayerUrl(fileId)}
                startIcon={<Play size={14} />}
            >
                {t('CAPTURES_WATCH')}
            </Button>
            <Button
                size="small"
                startIcon={<FolderOpen size={14} />}
                onClick={() => void fileLocation.openFolder(captureId, fileId)}
            >
                {t('CAPTURES_OPEN_FOLDER')}
            </Button>
            {fileLocation.failedItemId === captureId && (
                <p className={styles.failureReason}>{t('CAPTURES_OPEN_FOLDER_ERROR')}</p>
            )}
        </>
    );
};

export default CapturePromotedActions;
