import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { Divider, List, Typography } from '@mui/material';
import DetailRow from './detailRow';
import useFileTypeMetadataRows from './useFileTypeMetadataRows';

const FileTypeMetadataSection = ({ file }: { file: Pick<FileData, 'id' | 'type' | 'format'> }) => {
    const { t } = useI18n();
    const rows = useFileTypeMetadataRows(file);

    if (rows.length === 0) return null;

    return (
        <>
            <Divider sx={{ my: 1.5 }} />
            <Typography variant="overline" color="text.secondary" display="block">
                {t('FILE_METADATA_TITLE')}
            </Typography>
            <List dense disablePadding>
                {rows.map((row) => (
                    <DetailRow key={row.label} label={row.label} value={row.value} />
                ))}
            </List>
        </>
    );
};

export default FileTypeMetadataSection;
