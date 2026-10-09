import useI18n from '@/components/i18n/provider/i18nContext';
import { Box, IconButton, Tooltip, Typography } from '@mui/material';
import { Copy } from 'lucide-react';
import { useSnackbar } from 'notistack';

type ChecksumRowProps = {
    checksum?: string;
};

const ChecksumRow = ({ checksum }: ChecksumRowProps) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const hasChecksum = Boolean(checksum);

    const copyChecksum = async () => {
        try {
            await navigator.clipboard.writeText(checksum ?? '');
            enqueueSnackbar(t('FILE_DETAILS_CHECKSUM_COPIED'), { variant: 'success' });
        } catch {
            enqueueSnackbar(t('FILE_DETAILS_CHECKSUM_COPY_FAILED'), { variant: 'error' });
        }
    };

    return (
        <Box sx={{ py: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="caption" color="text.secondary">
                    {t('FILE_DETAILS_CHECKSUM')}
                </Typography>
                {hasChecksum ? (
                    <Tooltip title={t('FILE_DETAILS_CHECKSUM_COPY')}>
                        <IconButton
                            size="small"
                            onClick={copyChecksum}
                            aria-label={t('FILE_DETAILS_CHECKSUM_COPY')}
                        >
                            <Copy size={14} />
                        </IconButton>
                    </Tooltip>
                ) : null}
            </Box>
            {hasChecksum ? (
                <Typography
                    variant="caption"
                    component="code"
                    display="block"
                    sx={{
                        fontFamily: 'monospace',
                        overflowWrap: 'anywhere',
                        wordBreak: 'break-all',
                    }}
                >
                    {checksum}
                </Typography>
            ) : (
                <Typography variant="caption" color="text.secondary" display="block">
                    {t('FILE_DETAILS_CHECKSUM_CALCULATING')}
                </Typography>
            )}
        </Box>
    );
};

export default ChecksumRow;
