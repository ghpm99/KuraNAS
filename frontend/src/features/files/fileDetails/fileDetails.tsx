import { FileType, formatDate, getFileTypeInfo } from '@/utils';
import { formatSize } from '@/shared/utils/formatSize';
import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    Box,
    Chip,
    CircularProgress,
    Divider,
    IconButton,
    List,
    ListItem,
    Typography,
} from '@mui/material';
import { Flame, Snowflake, X } from 'lucide-react';
import ChecksumRow from './checksumRow';
import DetailRow from './detailRow';
import DiskLocationRow from './diskLocationRow';
import FileTypeMetadataSection from './fileTypeMetadataSection';
import FolderStatsSection from './folderStatsSection';
import { readOptionalDate } from './optionalDate';

type FileDetailsProps = {
    file: FileData;
    onClose: () => void;
};

const FileDetails = ({ file, onClose }: FileDetailsProps) => {
    const { isLoadingAccessData, recentAccessFiles } = useFile();
    const { t } = useI18n();
    const isFolder = file.type === FileType.Directory;
    const fileTypeDescription = isFolder
        ? t('FOLDER')
        : t(getFileTypeInfo(file.format).description);

    const formatOptionalDate = (rawDate: unknown) => {
        const readDate = readOptionalDate(rawDate);
        return readDate === null ? t('FILE_DETAILS_NEVER') : formatDate(readDate);
    };

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
                <IconButton size="small" onClick={onClose} aria-label={t('CLOSE')}>
                    <X size={18} />
                </IconButton>
            </Box>
            <Typography variant="caption" color="text.secondary" display="block" gutterBottom>
                {t('FILE_DETAILS_SUBTITLE')}
            </Typography>

            {file.tier === 'cold' ? (
                <Chip
                    size="small"
                    color="info"
                    variant="outlined"
                    icon={<Snowflake size={14} />}
                    label={t('FILE_TIER_COLD')}
                    sx={{ mt: 0.5 }}
                />
            ) : null}
            {file.tier === 'hot' ? (
                <Chip
                    size="small"
                    color="warning"
                    variant="outlined"
                    icon={<Flame size={14} />}
                    label={t('FILE_TIER_HOT')}
                    sx={{ mt: 0.5 }}
                />
            ) : null}

            <Typography variant="overline" color="text.secondary" display="block" sx={{ mt: 2 }}>
                {t('PROPERTIES')}
            </Typography>
            <List dense disablePadding>
                <DetailRow label={t('NAME')} value={file.name} />
                <DetailRow label={t('TYPE')} value={fileTypeDescription} />
                {isFolder ? null : (
                    <DetailRow
                        label={t('SIZE')}
                        value={`${formatSize(file.size)} (${file.size} B)`}
                    />
                )}
                <DetailRow label={t('CREATED')} value={formatDate(file.created_at)} />
                <DetailRow label={t('MODIFIED')} value={formatDate(file.updated_at)} />
                {isFolder ? null : (
                    <>
                        <DetailRow
                            label={t('FILE_DETAILS_LAST_INTERACTION')}
                            value={formatOptionalDate(file.last_interaction)}
                        />
                        <DetailRow
                            label={t('FILE_DETAILS_LAST_BACKUP')}
                            value={formatOptionalDate(file.last_backup)}
                        />
                    </>
                )}
                <DetailRow label={t('PATH')} value={file.path} />
            </List>
            {isFolder ? null : <ChecksumRow checksum={file.check_sum} />}
            <DiskLocationRow fileId={file.id} />

            {isFolder ? (
                <>
                    <Divider sx={{ my: 1.5 }} />
                    <FolderStatsSection folderId={file.id} />
                </>
            ) : (
                <>
                    <FileTypeMetadataSection file={file} />
                    <Divider sx={{ my: 1.5 }} />
                    <Typography variant="overline" color="text.secondary" display="block">
                        {t('RECENT_ACTIVITY')}
                    </Typography>
                    {isLoadingAccessData ? (
                        <CircularProgress size={16} />
                    ) : (
                        <List dense disablePadding>
                            {(recentAccessFiles ?? [])
                                .filter((access) => access.file_id === file.id)
                                .map((access) => (
                                    <ListItem key={access.id} disablePadding sx={{ py: 0.5 }}>
                                        <Box
                                            sx={{
                                                display: 'flex',
                                                justifyContent: 'space-between',
                                                width: '100%',
                                            }}
                                        >
                                            <Typography variant="caption">
                                                {access.ip_address}
                                            </Typography>
                                            <Typography variant="caption" color="text.secondary">
                                                {formatDate(access.accessed_at)}
                                            </Typography>
                                        </Box>
                                    </ListItem>
                                ))}
                        </List>
                    )}
                </>
            )}
        </Box>
    );
};

export default FileDetails;
