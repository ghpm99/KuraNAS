import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import { Box, CircularProgress, List, Typography } from '@mui/material';
import DetailRow from './detailRow';
import useFolderStats from './useFolderStats';

const FolderStatsSection = ({ folderId }: { folderId?: number }) => {
    const { t } = useI18n();
    const { data: folderStats, isPending, isError } = useFolderStats(folderId);

    if (isError || folderId === undefined) return null;

    if (isPending || !folderStats) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, py: 0.5 }}>
                <CircularProgress size={14} aria-label={t('FOLDER_STATS_CALCULATING')} />
                <Typography variant="caption" color="text.secondary">
                    {t('FOLDER_STATS_CALCULATING')}
                </Typography>
            </Box>
        );
    }

    return (
        <List dense disablePadding>
            <DetailRow label={t('FOLDER_STATS_FILE_COUNT')} value={folderStats.file_count ?? 0} />
            <DetailRow
                label={t('FOLDER_STATS_FOLDER_COUNT')}
                value={folderStats.folder_count ?? 0}
            />
            <DetailRow
                label={t('FOLDER_STATS_TOTAL_SIZE')}
                value={formatSize(folderStats.total_size_bytes ?? 0)}
            />
        </List>
    );
};

export default FolderStatsSection;
