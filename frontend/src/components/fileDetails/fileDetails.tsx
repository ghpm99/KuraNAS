import { FileType, formatDate, formatSize, getFileTypeInfo } from '@/utils';
import useFile from '../providers/fileProvider/fileContext';
import useI18n from '../i18n/provider/i18nContext';
import {
    Box,
    CircularProgress,
    Divider,
    IconButton,
    List,
    ListItem,
    Typography,
} from '@mui/material';
import { X } from 'lucide-react';
import type { ReactNode } from 'react';

function DetailRow({ label, value }: { label: string; value: ReactNode }) {
    return (
        <ListItem disablePadding sx={{ py: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                <Typography variant="caption" color="text.secondary">
                    {label}
                </Typography>
                <Typography
                    variant="caption"
                    sx={{ maxWidth: '60%', textAlign: 'right', wordBreak: 'break-all' }}
                >
                    {value}
                </Typography>
            </Box>
        </ListItem>
    );
}

const FileDetails = () => {
    const { selectedItem, isLoadingAccessData, recentAccessFiles, handleSelectItem } = useFile();
    const { t } = useI18n();

    if (!selectedItem || selectedItem.type === FileType.Directory) return null;

    const fileType = getFileTypeInfo(selectedItem.format);

    return (
        <Box sx={{ p: 2 }}>
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}
            >
                <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                    {t('FILE_DETAILS_TITLE')}
                </Typography>
                <IconButton
                    size="small"
                    onClick={() => handleSelectItem(null)}
                    aria-label={t('CLOSE')}
                >
                    <X size={18} />
                </IconButton>
            </Box>
            <Typography variant="caption" color="text.secondary" display="block" gutterBottom>
                {t('FILE_DETAILS_SUBTITLE')}
            </Typography>

            <Typography variant="overline" color="text.secondary" display="block" sx={{ mt: 2 }}>
                {t('PROPERTIES')}
            </Typography>
            <List dense disablePadding>
                <DetailRow label={t('TYPE')} value={fileType.description} />
                <DetailRow
                    label={t('SIZE')}
                    value={`${formatSize(selectedItem.size)} (${selectedItem.size} B)`}
                />
                <DetailRow label={t('CREATED')} value={formatDate(selectedItem.created_at)} />
                <DetailRow label={t('MODIFIED')} value={formatDate(selectedItem.updated_at)} />
                <DetailRow label={t('PATH')} value={selectedItem.path} />
            </List>

            <Divider sx={{ my: 1.5 }} />
            <Typography variant="overline" color="text.secondary" display="block">
                {t('RECENT_ACTIVITY')}
            </Typography>
            {isLoadingAccessData ? (
                <CircularProgress size={16} />
            ) : (
                <List dense disablePadding>
                    {recentAccessFiles
                        .filter((access) => access.file_id === selectedItem.id)
                        .map((access) => (
                            <ListItem key={access.id} disablePadding sx={{ py: 0.5 }}>
                                <Box
                                    sx={{
                                        display: 'flex',
                                        justifyContent: 'space-between',
                                        width: '100%',
                                    }}
                                >
                                    <Typography variant="caption">{access.ip_address}</Typography>
                                    <Typography variant="caption" color="text.secondary">
                                        {formatDate(access.accessed_at)}
                                    </Typography>
                                </Box>
                            </ListItem>
                        ))}
                </List>
            )}
        </Box>
    );
};

export default FileDetails;
