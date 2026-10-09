import { Tab, Tabs } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileSearchMode } from './useFileSearchMode';
import styles from './fileSearch.module.css';

interface FileSearchModeTabsProps {
    mode?: FileSearchMode;
    onChange?: (mode: FileSearchMode) => void;
}

const FileSearchModeTabs = ({ mode = 'name', onChange }: FileSearchModeTabsProps) => {
    const { t } = useI18n();

    return (
        <Tabs
            value={mode}
            onChange={(_, nextMode: FileSearchMode) => onChange?.(nextMode)}
            aria-label={t('FILES_SEARCH_MODE_LABEL')}
            className={styles.modeTabs}
        >
            <Tab value="name" label={t('FILES_SEARCH_MODE_NAME')} />
            <Tab value="content" label={t('FILES_SEARCH_MODE_CONTENT')} />
        </Tabs>
    );
};

export default FileSearchModeTabs;
