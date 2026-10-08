import type { Ref } from 'react';
import { Checkbox, FormControlLabel, IconButton, InputBase } from '@mui/material';
import { Search, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './fileSearch.module.css';

interface FileSearchBarProps {
    value?: string;
    onChange?: (value: string) => void;
    onClear?: () => void;
    isFolderScope?: boolean;
    isRecursive?: boolean;
    onRecursiveChange?: (isRecursive: boolean) => void;
    inputRef?: Ref<HTMLInputElement>;
}

const FileSearchBar = ({
    value = '',
    onChange,
    onClear,
    isFolderScope = false,
    isRecursive = true,
    onRecursiveChange,
    inputRef,
}: FileSearchBarProps) => {
    const { t } = useI18n();
    const placeholder = isFolderScope ? t('FILES_SEARCH_PLACEHOLDER') : t('FILES_SEARCH_PLACEHOLDER_ALL');

    return (
        <div className={styles.searchBar}>
            <div className={styles.searchField}>
                <Search size={16} className={styles.searchIcon} />
                <InputBase
                    value={value}
                    inputRef={inputRef}
                    onChange={(event) => onChange?.(event.target.value)}
                    placeholder={placeholder}
                    className={styles.searchInput}
                    inputProps={{ 'aria-label': placeholder }}
                />
                {value ? (
                    <IconButton size="small" aria-label={t('FILES_SEARCH_CLEAR')} onClick={onClear}>
                        <X size={14} />
                    </IconButton>
                ) : null}
            </div>
            {isFolderScope ? (
                <FormControlLabel
                    className={styles.recursiveToggle}
                    label={t('FILES_SEARCH_INCLUDE_SUBFOLDERS')}
                    control={
                        <Checkbox
                            size="small"
                            checked={isRecursive}
                            onChange={(event) => onRecursiveChange?.(event.target.checked)}
                        />
                    }
                />
            ) : null}
        </div>
    );
};

export default FileSearchBar;
