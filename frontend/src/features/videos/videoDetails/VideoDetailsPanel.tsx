import { Box, Button, CircularProgress, Drawer, IconButton, List, Typography } from '@mui/material';
import { Download, X } from 'lucide-react';
import { Link } from 'react-router-dom';
import { buildFilesUrl } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import DetailRow from '@/features/files/fileDetails/detailRow';
import { readOptionalDate } from '@/features/files/fileDetails/optionalDate';
import { getFileDownloadUrl } from '@/service/files';
import type { VideoFileDto, VideoLibraryResolution } from '@/service/videoPlayback';
import { formatSize } from '@/shared/utils/formatSize';
import { formatDate } from '@/utils';
import { usePlayerPortalContainer } from '@/features/videos/videoPlayer/playerPortalContainer';
import useVideoTechnicalRows from './useVideoTechnicalRows';

type VideoDetailsPanelProps = {
    video: VideoFileDto | null;
    onClose: () => void;
};

const resolutionLabelKeys: Record<VideoLibraryResolution, string> = {
    sd: 'VIDEO_BADGE_RESOLUTION_SD',
    hd: 'VIDEO_BADGE_RESOLUTION_HD',
    fhd: 'VIDEO_BADGE_RESOLUTION_FHD',
    uhd: 'VIDEO_BADGE_RESOLUTION_UHD',
};

const panelWidth = { xs: '100vw', sm: 380 };

export default function VideoDetailsPanel({ video, onClose }: VideoDetailsPanelProps) {
    if (!video) return null;
    return <OpenVideoDetailsPanel video={video} onClose={onClose} />;
}

type OpenVideoDetailsPanelProps = {
    video: VideoFileDto;
    onClose: () => void;
};

function OpenVideoDetailsPanel({ video, onClose }: OpenVideoDetailsPanelProps) {
    const { t } = useI18n();
    const portalContainer = usePlayerPortalContainer();
    const { rows, isLoading, hasFailed, backendMessage } = useVideoTechnicalRows(video);
    const createdAt = readOptionalDate(video.created_at);
    const updatedAt = readOptionalDate(video.updated_at);
    const formatLabel = (video.format ?? '').replace(/^\./, '').toUpperCase();

    return (
        <Drawer anchor="right" open onClose={onClose} ModalProps={{ container: portalContainer }}>
            <Box
                sx={{ p: 2, width: panelWidth }}
                role="region"
                aria-label={t('VIDEO_DETAILS_TITLE')}
            >
                <Box
                    sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
                >
                    <Typography variant="subtitle1" fontWeight={600}>
                        {t('VIDEO_DETAILS_TITLE')}
                    </Typography>
                    <IconButton size="small" onClick={onClose} aria-label={t('CLOSE')}>
                        <X size={18} />
                    </IconButton>
                </Box>

                <Button
                    component="a"
                    href={getFileDownloadUrl(video.id)}
                    download
                    variant="contained"
                    size="small"
                    startIcon={<Download size={14} />}
                    sx={{ my: 1.5 }}
                >
                    {t('VIDEO_ERROR_DOWNLOAD_ORIGINAL')}
                </Button>

                <List dense disablePadding>
                    {video.name ? <DetailRow label={t('NAME')} value={video.name} /> : null}
                    {video.parent_path ? (
                        <DetailRow
                            label={t('IMAGES_DETAIL_FOLDER')}
                            value={
                                <Link
                                    to={buildFilesUrl(video.parent_path)}
                                    title={t('VIDEO_DETAILS_OPEN_FOLDER')}
                                >
                                    {video.parent_path}
                                </Link>
                            }
                        />
                    ) : null}
                    {video.size > 0 ? (
                        <DetailRow label={t('SIZE')} value={formatSize(video.size)} />
                    ) : null}
                    {formatLabel ? (
                        <DetailRow label={t('IMAGES_DETAIL_FORMAT')} value={formatLabel} />
                    ) : null}
                    {rows.map((row) => (
                        <DetailRow
                            key={row.labelKey}
                            label={t(row.labelKey)}
                            value={
                                row.resolutionClass
                                    ? `${row.value} (${t(resolutionLabelKeys[row.resolutionClass])})`
                                    : row.value
                            }
                        />
                    ))}
                    {createdAt ? (
                        <DetailRow label={t('CREATED')} value={formatDate(createdAt)} />
                    ) : null}
                    {updatedAt ? (
                        <DetailRow label={t('MODIFIED')} value={formatDate(updatedAt)} />
                    ) : null}
                </List>

                {isLoading ? (
                    <Box role="status" aria-label={t('LOADING')} sx={{ mt: 1.5 }}>
                        <CircularProgress size={16} />
                    </Box>
                ) : null}
                {hasFailed ? (
                    <Box role="alert" sx={{ mt: 1.5 }}>
                        <Typography variant="caption" color="warning.main" display="block">
                            {t('VIDEO_DETAILS_METADATA_UNAVAILABLE')}
                        </Typography>
                        {backendMessage ? (
                            <Typography variant="caption" color="text.secondary" display="block">
                                {backendMessage}
                            </Typography>
                        ) : null}
                    </Box>
                ) : null}
            </Box>
        </Drawer>
    );
}
