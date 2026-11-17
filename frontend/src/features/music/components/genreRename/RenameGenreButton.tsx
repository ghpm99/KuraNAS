import { Button } from '@mui/material';
import { Pencil } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import RenameGenreDialog from './RenameGenreDialog';

interface RenameGenreButtonProps {
    genreKey: string;
    genreName: string;
    trackCount: number;
    onRenamed: (renamedGenreKey: string) => void;
}

export default function RenameGenreButton({
    genreKey,
    genreName,
    trackCount,
    onRenamed,
}: RenameGenreButtonProps) {
    const { t } = useI18n();
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    return (
        <>
            <Button
                size="small"
                color="inherit"
                startIcon={<Pencil size={16} />}
                onClick={() => setIsDialogOpen(true)}
            >
                {t('MUSIC_GENRE_RENAME')}
            </Button>
            {isDialogOpen && (
                <RenameGenreDialog
                    genreKey={genreKey}
                    genreName={genreName}
                    trackCount={trackCount}
                    onRenamed={onRenamed}
                    onClose={() => setIsDialogOpen(false)}
                />
            )}
        </>
    );
}
