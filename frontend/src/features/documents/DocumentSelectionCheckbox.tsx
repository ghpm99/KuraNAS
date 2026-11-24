import { Checkbox } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type DocumentSelectionCheckboxProps = {
    documentName: string;
    isSelected: boolean;
    onToggle: () => void;
};

const DocumentSelectionCheckbox = ({
    documentName,
    isSelected,
    onToggle,
}: DocumentSelectionCheckboxProps) => {
    const { t } = useI18n();

    return (
        <Checkbox
            size="small"
            checked={isSelected}
            onChange={onToggle}
            slotProps={{ input: { 'aria-label': t('FILES_SELECT_ITEM', { name: documentName }) } }}
            sx={{ p: 0.5 }}
        />
    );
};

export default DocumentSelectionCheckbox;
