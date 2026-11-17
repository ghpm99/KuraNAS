import { Button } from '@mui/material';
import { Pencil } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { BatchTrackMetadataSelector } from '@/types/music';
import BatchTrackMetadataDialog from './BatchTrackMetadataDialog';

interface EditCollectionMetadataButtonProps {
    selector: BatchTrackMetadataSelector;
}

export default function EditCollectionMetadataButton({
    selector,
}: EditCollectionMetadataButtonProps) {
    const { t } = useI18n();
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const labelKey = 'albumKey' in selector ? 'MUSIC_EDIT_ALBUM' : 'MUSIC_EDIT_FOLDER';

    return (
        <>
            <Button
                size="small"
                color="inherit"
                startIcon={<Pencil size={16} />}
                onClick={() => setIsDialogOpen(true)}
            >
                {t(labelKey)}
            </Button>
            {isDialogOpen && (
                <BatchTrackMetadataDialog
                    selector={selector}
                    onClose={() => setIsDialogOpen(false)}
                />
            )}
        </>
    );
}
