import { Box, Button, Typography } from '@mui/material';
import { CheckCheck, CheckSquare, Pencil, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import BatchTrackMetadataDialog from '@/features/music/components/trackMetadata/BatchTrackMetadataDialog';
import type { IMusicData } from '@/types/music';
import { MAX_TRACKS_PER_BATCH_EDIT, type TrackSelection } from './useTrackSelection';
import useTrackSelectionControls from './useTrackSelectionControls';

interface TrackSelectionControlsProps {
    selection: TrackSelection;
    loadedTracks: IMusicData[];
}

export default function TrackSelectionControls({
    selection,
    loadedTracks,
}: TrackSelectionControlsProps) {
    const { t } = useI18n();
    const { editedFileIds, openEditDialog, closeEditDialog } = useTrackSelectionControls(
        selection,
        loadedTracks
    );

    return (
        <>
            {selection.isSelecting ? (
                <Box
                    role="toolbar"
                    aria-label={t('MUSIC_SELECTION_TOOLBAR')}
                    sx={{
                        position: 'sticky',
                        bottom: 0,
                        zIndex: 1,
                        display: 'flex',
                        flexWrap: 'wrap',
                        alignItems: 'center',
                        gap: 1,
                        p: 1,
                        bgcolor: 'background.paper',
                        borderTop: 1,
                        borderColor: 'divider',
                    }}
                >
                    <Button
                        size="small"
                        color="inherit"
                        startIcon={<X size={16} />}
                        onClick={selection.stopSelecting}
                    >
                        {t('MUSIC_SELECTION_CANCEL')}
                    </Button>
                    <Typography variant="body2" fontWeight={600} sx={{ mr: 1 }}>
                        {t('MUSIC_SELECTION_COUNT', { count: String(selection.selectedCount) })}
                    </Typography>
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<CheckCheck size={16} />}
                        onClick={() => selection.selectAllLoaded(loadedTracks)}
                    >
                        {t('MUSIC_SELECTION_SELECT_ALL_LOADED')}
                    </Button>
                    <Button
                        size="small"
                        variant="outlined"
                        onClick={selection.clear}
                        disabled={selection.selectedCount === 0}
                    >
                        {t('MUSIC_SELECTION_CLEAR')}
                    </Button>
                    <Button
                        size="small"
                        variant="contained"
                        startIcon={<Pencil size={16} />}
                        onClick={openEditDialog}
                        disabled={!selection.canEditSelection}
                    >
                        {t('MUSIC_SELECTION_EDIT')}
                    </Button>
                    {selection.isOverBatchLimit && (
                        <Typography variant="caption" color="error" role="alert">
                            {t('MUSIC_SELECTION_LIMIT_EXCEEDED', {
                                max: String(MAX_TRACKS_PER_BATCH_EDIT),
                            })}
                        </Typography>
                    )}
                </Box>
            ) : (
                loadedTracks.length > 0 && (
                    <Button
                        size="small"
                        color="inherit"
                        startIcon={<CheckSquare size={16} />}
                        onClick={selection.startSelecting}
                    >
                        {t('MUSIC_SELECT_TRACKS')}
                    </Button>
                )
            )}
            {editedFileIds && (
                <BatchTrackMetadataDialog
                    selector={{ fileIds: editedFileIds }}
                    onClose={closeEditDialog}
                    onUpdated={selection.stopSelecting}
                />
            )}
        </>
    );
}
