import useI18n from '@/components/i18n/provider/i18nContext';
import { Box, Button, IconButton, Tooltip, Typography } from '@mui/material';
import { Copy, Flame } from 'lucide-react';
import { useSnackbar } from 'notistack';
import useFileLocation from './useFileLocation';
import usePromoteFileToHot from './usePromoteFileToHot';

const DiskLocationRow = ({ fileId }: { fileId?: number }) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const location = useFileLocation(fileId);
    const { promoteToHot, isPromoting } = usePromoteFileToHot(fileId);

    if (!location?.disk_path) return null;

    const copyDiskPath = async () => {
        try {
            await navigator.clipboard.writeText(location.disk_path);
            enqueueSnackbar(t('FILE_DISK_LOCATION_COPIED'), { variant: 'success' });
        } catch {
            enqueueSnackbar(t('FILE_DISK_LOCATION_COPY_FAILED'), { variant: 'error' });
        }
    };

    return (
        <Box sx={{ py: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="caption" color="text.secondary">
                    {t('FILE_DISK_LOCATION')}
                </Typography>
                <Tooltip title={t('FILE_DISK_LOCATION_COPY')}>
                    <IconButton
                        size="small"
                        onClick={copyDiskPath}
                        aria-label={t('FILE_DISK_LOCATION_COPY')}
                    >
                        <Copy size={14} />
                    </IconButton>
                </Tooltip>
            </Box>
            <Typography
                variant="caption"
                component="code"
                display="block"
                sx={{ fontFamily: 'monospace', overflowWrap: 'anywhere', wordBreak: 'break-all' }}
            >
                {location.disk_path}
            </Typography>
            {location.tier === 'cold' ? (
                <Button
                    size="small"
                    variant="outlined"
                    startIcon={<Flame size={14} />}
                    disabled={isPromoting}
                    onClick={promoteToHot}
                    sx={{ mt: 0.5 }}
                >
                    {t('FILE_PROMOTE_TO_HOT')}
                </Button>
            ) : null}
            {location.exists_on_disk ? null : (
                <Typography variant="caption" color="warning.main" display="block" sx={{ mt: 0.5 }}>
                    {t('FILE_DISK_LOCATION_MISSING')}
                </Typography>
            )}
        </Box>
    );
};

export default DiskLocationRow;
