import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getFileDownloadUrl } from '@/service/files';
import { formatSize } from '@/shared/utils/formatSize';
import { getFileTypeInfo } from '@/utils';
import { Box, Button, Typography } from '@mui/material';
import { Download, FileQuestion } from 'lucide-react';

type UnsupportedFileCardProps = {
    file: Pick<FileData, 'id' | 'name' | 'size' | 'format'>;
    reasonKey?: string;
};

const UnsupportedFileCard = ({
    file,
    reasonKey = 'FILE_PREVIEW_UNAVAILABLE',
}: UnsupportedFileCardProps) => {
    const { t } = useI18n();
    const fileType = getFileTypeInfo(file.format);

    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 1.5,
                p: 3,
                textAlign: 'center',
                minWidth: 0,
            }}
        >
            <FileQuestion size={48} aria-hidden="true" />
            <Typography variant="subtitle1" fontWeight={600} sx={{ wordBreak: 'break-all' }}>
                {file.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
                {t(fileType.description)} · {formatSize(file.size ?? 0)}
            </Typography>
            <Typography variant="body2">{t(reasonKey)}</Typography>
            <Button
                variant="contained"
                startIcon={<Download size={16} />}
                href={getFileDownloadUrl(file.id)}
                download={file.name}
            >
                {t('DOWNLOAD')}
            </Button>
        </Box>
    );
};

export default UnsupportedFileCard;
