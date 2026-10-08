import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { Box, IconButton, Typography } from '@mui/material';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import useSiblingNavigation from './useSiblingNavigation';

const FileViewerNavigation = ({ file }: { file: FileData }) => {
    const { t } = useI18n();
    const { handleSelectItem } = useFile();
    const siblingPosition = useSiblingNavigation(file);

    if (!siblingPosition) return null;

    const { previous, next, position, total } = siblingPosition;

    return (
        <Box
            sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 1,
                py: 0.5,
            }}
        >
            <IconButton
                size="small"
                aria-label={t('FILE_VIEWER_PREVIOUS')}
                disabled={previous === null}
                onClick={() => previous && handleSelectItem(previous)}
            >
                <ChevronLeft size={18} />
            </IconButton>
            <Typography variant="caption" color="text.secondary">
                {t('FILE_VIEWER_POSITION', { position: String(position), total: String(total) })}
            </Typography>
            <IconButton
                size="small"
                aria-label={t('FILE_VIEWER_NEXT')}
                disabled={next === null}
                onClick={() => next && handleSelectItem(next)}
            >
                <ChevronRight size={18} />
            </IconButton>
        </Box>
    );
};

export default FileViewerNavigation;
