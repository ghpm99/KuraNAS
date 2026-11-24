import { Button } from '@mui/material';
import { ListPlus } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicRoute } from '@/app/routes';

interface CreateRulePlaylistFromGenreButtonProps {
    genreKey: string;
}

export default function CreateRulePlaylistFromGenreButton({
    genreKey,
}: CreateRulePlaylistFromGenreButtonProps) {
    const { t } = useI18n();
    const navigate = useNavigate();

    return (
        <Button
            size="small"
            color="inherit"
            startIcon={<ListPlus size={16} />}
            onClick={() =>
                navigate(
                    `${getMusicRoute('playlists')}?create_rule_genre=${encodeURIComponent(genreKey)}`
                )
            }
        >
            {t('MUSIC_GENRE_CREATE_RULE_PLAYLIST')}
        </Button>
    );
}
