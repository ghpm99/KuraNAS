import { IconButton } from '@mui/material';
import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

type DocumentStarButtonProps = {
    documentName: string;
    isStarred: boolean;
    onToggle: () => void;
};

const DocumentStarButton = ({ documentName, isStarred, onToggle }: DocumentStarButtonProps) => {
    const { t } = useI18n();

    return (
        <IconButton
            size="small"
            aria-label={t(isStarred ? 'FILES_UNFAVORITE_ITEM' : 'FILES_FAVORITE_ITEM', {
                name: documentName,
            })}
            aria-pressed={isStarred}
            onClick={onToggle}
            sx={{ color: 'var(--app-color-warning, #f5b301)' }}
        >
            <Star size={16} fill={isStarred ? 'currentColor' : 'none'} />
        </IconButton>
    );
};

export default DocumentStarButton;
