import { IconButton, InputBase } from '@mui/material';
import { Search, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useMusicSearchField } from '@/features/music/components/useMusicSearchField';
import styles from './MusicSearchField.module.css';

const MusicSearchField = () => {
    const { t } = useI18n();
    const { inputText, setInputText, submit, clear } = useMusicSearchField();
    const placeholder = t('MUSIC_SEARCH_PLACEHOLDER');

    return (
        <div className={styles.searchField}>
            <Search size={16} className={styles.searchIcon} />
            <InputBase
                value={inputText}
                onChange={(event) => setInputText(event.target.value)}
                onKeyDown={(event) => {
                    if (event.key === 'Enter') {
                        submit();
                    }
                }}
                placeholder={placeholder}
                className={styles.searchInput}
                inputProps={{ 'aria-label': placeholder }}
            />
            {inputText ? (
                <IconButton size="small" aria-label={t('MUSIC_SEARCH_CLEAR')} onClick={clear}>
                    <X size={14} />
                </IconButton>
            ) : null}
        </div>
    );
};

export default MusicSearchField;
