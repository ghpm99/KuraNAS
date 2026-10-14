import { Plus } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import VideoPlaylistNameDialog from './VideoPlaylistNameDialog';
import styles from '../videoContent.module.css';

type VideoNewPlaylistButtonProps = {
    isCreating: boolean;
    onCreate: (name: string) => void;
};

export default function VideoNewPlaylistButton({
    isCreating,
    onCreate,
}: VideoNewPlaylistButtonProps) {
    const { t } = useI18n();
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    return (
        <>
            <button
                type="button"
                className={styles.actionBtn}
                onClick={() => setIsDialogOpen(true)}
            >
                <Plus size={14} />
                {t('VIDEO_NEW_PLAYLIST')}
            </button>
            <VideoPlaylistNameDialog
                isOpen={isDialogOpen}
                isSubmitting={isCreating}
                confirmLabelKey="ACTION_CREATE"
                onClose={() => setIsDialogOpen(false)}
                onSubmit={(name) => {
                    setIsDialogOpen(false);
                    onCreate(name);
                }}
            />
        </>
    );
}
