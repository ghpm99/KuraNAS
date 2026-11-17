import { Box, Button, IconButton, Tooltip, Typography } from '@mui/material';
import { Trash2, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

type DuplicateSelectionBarProps = {
    selectedCount: number;
    onClear: () => void;
    onDelete: () => void;
};

const DuplicateSelectionBar = ({
    selectedCount,
    onClear,
    onDelete,
}: DuplicateSelectionBarProps) => {
    const { t } = useI18n();

    return (
        <Box
            role="toolbar"
            aria-label={t('FILES_SELECTION_TOOLBAR')}
            sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1 }}
        >
            <Tooltip title={t('FILES_CLEAR_SELECTION')}>
                <IconButton size="small" aria-label={t('FILES_CLEAR_SELECTION')} onClick={onClear}>
                    <X size={16} />
                </IconButton>
            </Tooltip>
            <Typography variant="body2" fontWeight={600}>
                {t('FILES_SELECTION_COUNT', { count: String(selectedCount) })}
            </Typography>
            <Button
                size="small"
                variant="outlined"
                color="error"
                startIcon={<Trash2 size={16} />}
                onClick={onDelete}
            >
                {t('DELETE')}
            </Button>
        </Box>
    );
};

export default DuplicateSelectionBar;
