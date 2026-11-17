import { Tab, Tabs as MuiTabs } from '@mui/material';
import { useSearchParams } from 'react-router-dom';
import { FileType } from '@/utils';
import useFile, {
    type FileListCategoryType,
} from '@/features/files/providers/fileProvider/fileContext';
import { fileListFilterSearchParam } from '@/features/files/files/fileListFilterFromSearch';
import useI18n from '../i18n/provider/i18nContext';

const Tabs = () => {
    const { t } = useI18n();
    const { selectedItem, fileListFilter } = useFile();
    const [searchParams, setSearchParams] = useSearchParams();

    if (selectedItem?.type === FileType.File) return null;

    const selectFilter = (filter: FileListCategoryType) => {
        const nextSearchParams = new URLSearchParams(searchParams);
        if (filter === 'all') {
            nextSearchParams.delete(fileListFilterSearchParam);
        } else {
            nextSearchParams.set(fileListFilterSearchParam, filter);
        }
        setSearchParams(nextSearchParams);
    };

    return (
        <MuiTabs value={fileListFilter} onChange={(_, filter) => selectFilter(filter)}>
            <Tab label={t('ALL_FILES')} value="all" />
            <Tab label={t('RECENT_FILES')} value="recent" />
            <Tab label={t('STARRED_FILES')} value="starred" />
        </MuiTabs>
    );
};

export default Tabs;
