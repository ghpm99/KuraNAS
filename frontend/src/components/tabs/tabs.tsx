import { Tab, Tabs as MuiTabs } from '@mui/material';
import { FileType } from '@/utils';
import useFile from '../providers/fileProvider/fileContext';
import useI18n from '../i18n/provider/i18nContext';

const Tabs = () => {
    const { t } = useI18n();
    const { selectedItem, fileListFilter, setFileListFilter } = useFile();

    if (selectedItem?.type === FileType.File) return null;

    return (
        <MuiTabs value={fileListFilter} onChange={(_, val) => setFileListFilter(val)}>
            <Tab label={t('ALL_FILES')} value="all" />
            <Tab label={t('RECENT_FILES')} value="recent" />
            <Tab label={t('STARRED_FILES')} value="starred" />
        </MuiTabs>
    );
};

export default Tabs;
