import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { duplicateCategories, type DuplicateCategory } from '@/types/analytics';

type DuplicateCategorySelectProps = {
    category: DuplicateCategory | undefined;
    onChange: (category: DuplicateCategory | undefined) => void;
};

const allCategoriesValue = 'all';

const categoryLabelKeys: Record<DuplicateCategory, string> = {
    image: 'ANALYTICS_DUPLICATE_TYPE_IMAGE',
    audio: 'ANALYTICS_DUPLICATE_TYPE_AUDIO',
    video: 'ANALYTICS_DUPLICATE_TYPE_VIDEO',
    document: 'ANALYTICS_DUPLICATE_TYPE_DOCUMENT',
    other: 'ANALYTICS_DUPLICATE_TYPE_OTHER',
};

const DuplicateCategorySelect = ({ category, onChange }: DuplicateCategorySelectProps) => {
    const { t } = useI18n();

    return (
        <FormControl size="small" sx={{ minWidth: 200 }}>
            <InputLabel id="analytics-duplicates-type">
                {t('ANALYTICS_DUPLICATES_FILTER_LABEL')}
            </InputLabel>
            <Select
                labelId="analytics-duplicates-type"
                label={t('ANALYTICS_DUPLICATES_FILTER_LABEL')}
                value={category ?? allCategoriesValue}
                onChange={(event) =>
                    onChange(duplicateCategories.find((option) => option === event.target.value))
                }
            >
                <MenuItem value={allCategoriesValue}>{t('ANALYTICS_DUPLICATE_TYPE_ALL')}</MenuItem>
                {duplicateCategories.map((option) => (
                    <MenuItem key={option} value={option}>
                        {t(categoryLabelKeys[option])}
                    </MenuItem>
                ))}
            </Select>
        </FormControl>
    );
};

export default DuplicateCategorySelect;
