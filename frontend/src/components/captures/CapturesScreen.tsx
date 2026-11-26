import ListFeedback from '@/components/listFeedback/ListFeedback';
import { formatRelativeDate } from '@/shared/utils/formatRelativeDate';
import EmptyState from '@/components/emptyState/emptyState';
import { appRoutes } from '@/app/routes';
import { Link } from 'react-router-dom';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import { Button, Chip, CircularProgress, IconButton, MenuItem, TextField } from '@mui/material';
import { ExternalLink, Film, RefreshCw, RotateCcw, Trash2 } from 'lucide-react';
import type { Capture, CaptureStatus } from '@/types/captures';
import { useRef } from 'react';
import CapturePromotedActions from './CapturePromotedActions';
import CaptureDeleteDialog from './CaptureDeleteDialog';
import styles from './CapturesScreen.module.css';
import { captureStatuses } from './capturesFilters';
import { useCapturesFilters } from './useCapturesFilters';
import { useCapturesScreen } from './useCapturesScreen';

const statusClass: Record<CaptureStatus, string | undefined> = {
    uploaded: styles.statusUploaded,
    promoting: styles.statusPromoting,
    promoted: styles.statusPromoted,
    failed: styles.statusFailed,
};

const formatEpisode = (capture: Capture): string => {
    if (capture.season != null && capture.episode != null) {
        return `S${capture.season}E${capture.episode}`;
    }
    if (capture.episode != null) {
        return `E${capture.episode}`;
    }
    return '';
};

const CapturesScreen = () => {
    const filters = useCapturesFilters();
    const {
        t,
        items,
        isLoading,
        isError,
        isReloading,
        reload,
        errorMessage,
        retryLoading,
        isNextPageError,
        hasNextPage,
        isFetchingNextPage,
        fetchNextPage,
        retryPromotion,
        isRetryingPromotion,
        deletingId,
        captureAwaitingDeletion,
        requestRemoval,
        cancelRemoval,
        confirmRemoval,
        formatSize,
        captureThumbnailHref,
        captureStatusLabelKey,
    } = useCapturesScreen(filters.appliedFilters);

    const screenTopRef = useRef<HTMLDivElement>(null);

    const reloadFromTop = () => {
        screenTopRef.current?.scrollIntoView({ block: 'start' });
        void reload();
    };

    return (
        <PageContainer>
            <div ref={screenTopRef} />
            <PageHeader
                title={t('CAPTURES_PAGE_TITLE')}
                subtitle={t('CAPTURES_PAGE_DESCRIPTION')}
                actions={
                    <Button
                        size="small"
                        startIcon={<RefreshCw size={14} />}
                        disabled={isLoading || isReloading}
                        onClick={reloadFromTop}
                    >
                        {t('CAPTURES_RELOAD')}
                    </Button>
                }
            />

            <div className={styles.filters}>
                <TextField
                    type="search"
                    size="small"
                    placeholder={t('CAPTURES_SEARCH_PLACEHOLDER')}
                    value={filters.nameInput}
                    onChange={(event) => filters.changeNameInput(event.target.value)}
                    slotProps={{ htmlInput: { 'aria-label': t('CAPTURES_SEARCH_PLACEHOLDER') } }}
                />
                <TextField
                    select
                    size="small"
                    label={t('CAPTURES_FILTER_STATUS')}
                    value={filters.appliedFilters.status}
                    onChange={(event) =>
                        filters.changeStatus(event.target.value as CaptureStatus | '')
                    }
                    sx={{ minWidth: 180 }}
                >
                    <MenuItem value="">{t('CAPTURES_FILTER_STATUS_ALL')}</MenuItem>
                    {captureStatuses.map((status) => (
                        <MenuItem key={status} value={status}>
                            {t(captureStatusLabelKey(status))}
                        </MenuItem>
                    ))}
                </TextField>
            </div>

            {isLoading && (
                <div className={styles.state}>
                    <CircularProgress size={20} />
                    <span>{t('CAPTURES_LOADING')}</span>
                </div>
            )}

            {!isLoading && (
                <ListFeedback
                    isEmpty={items.length === 0}
                    isError={isError}
                    isNextPageError={isNextPageError}
                    errorTitle={t('CAPTURES_ERROR')}
                    errorMessage={errorMessage}
                    onRetry={retryLoading}
                    hasNextPage={hasNextPage}
                    isFetchingNextPage={isFetchingNextPage}
                    onLoadMore={fetchNextPage}
                    emptyState={
                        filters.hasActiveFilters ? (
                            <div className={styles.state}>
                                <span>{t('CAPTURES_FILTER_EMPTY')}</span>
                                <Button size="small" onClick={filters.clearFilters}>
                                    {t('CAPTURES_FILTER_CLEAR')}
                                </Button>
                            </div>
                        ) : (
                            <EmptyState
                                icon={<Film size={28} aria-hidden />}
                                title={t('CAPTURES_EMPTY')}
                                description={t('CAPTURES_EMPTY_DESCRIPTION')}
                                actions={
                                    <Button
                                        size="small"
                                        variant="outlined"
                                        component={Link}
                                        to={appRoutes.downloads}
                                    >
                                        {t('CAPTURES_EMPTY_INSTALL_EXTENSION')}
                                    </Button>
                                }
                            />
                        )
                    }
                >
                    <ul className={styles.list}>
                        {items.map((capture) => {
                            const thumb = captureThumbnailHref(capture);
                            const episode = formatEpisode(capture);
                            const heading = capture.title || capture.name;

                            return (
                                <li key={capture.id} className={styles.card}>
                                    <div className={styles.thumb}>
                                        {thumb ? (
                                            <img
                                                className={styles.thumbImg}
                                                src={thumb}
                                                alt={heading}
                                                loading="lazy"
                                            />
                                        ) : (
                                            <Film size={28} aria-hidden />
                                        )}
                                    </div>

                                    <div className={styles.body}>
                                        <div className={styles.headingRow}>
                                            <h2 className={styles.cardTitle}>{heading}</h2>
                                            <Chip
                                                size="small"
                                                className={statusClass[capture.status]}
                                                label={t(captureStatusLabelKey(capture.status))}
                                            />
                                        </div>

                                        {(capture.episode_title || episode) && (
                                            <p className={styles.subtitle}>
                                                {[episode, capture.episode_title]
                                                    .filter(Boolean)
                                                    .join(' · ')}
                                            </p>
                                        )}

                                        {capture.status === 'failed' && capture.failure_reason && (
                                            <p className={styles.failureReason}>
                                                {capture.failure_reason}
                                            </p>
                                        )}

                                        <div className={styles.meta}>
                                            {capture.platform && (
                                                <Chip
                                                    size="small"
                                                    variant="outlined"
                                                    label={capture.platform}
                                                />
                                            )}
                                            <Chip
                                                size="small"
                                                variant="outlined"
                                                label={`${t('CAPTURES_SIZE')}: ${formatSize(capture.size)}`}
                                            />
                                            <span className={styles.date}>
                                                {formatRelativeDate(capture.created_at)}
                                            </span>
                                        </div>
                                    </div>

                                    <div className={styles.actions}>
                                        {capture.status === 'promoted' &&
                                            typeof capture.file_id === 'number' && (
                                                <CapturePromotedActions
                                                    captureId={capture.id}
                                                    fileId={capture.file_id}
                                                />
                                            )}
                                        {capture.source_url && (
                                            <Button
                                                component="a"
                                                href={capture.source_url}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                size="small"
                                                startIcon={<ExternalLink size={14} />}
                                            >
                                                {t('CAPTURES_SOURCE_LINK')}
                                            </Button>
                                        )}
                                        {capture.status === 'failed' && (
                                            <Button
                                                size="small"
                                                startIcon={<RotateCcw size={14} />}
                                                disabled={isRetryingPromotion(capture.id)}
                                                onClick={() => void retryPromotion(capture.id)}
                                            >
                                                {t('CAPTURES_RETRY_PROMOTION')}
                                            </Button>
                                        )}
                                        <IconButton
                                            aria-label={t('CAPTURES_DELETE')}
                                            size="small"
                                            disabled={deletingId === capture.id}
                                            onClick={() => requestRemoval(capture)}
                                        >
                                            <Trash2 size={16} />
                                        </IconButton>
                                    </div>
                                </li>
                            );
                        })}
                    </ul>
                    {!hasNextPage && !isNextPageError ? (
                        <div className={styles.state}>{t('CAPTURES_END')}</div>
                    ) : null}
                </ListFeedback>
            )}

            <CaptureDeleteDialog
                capture={captureAwaitingDeletion}
                onCancel={cancelRemoval}
                onConfirm={confirmRemoval}
            />
        </PageContainer>
    );
};

export default CapturesScreen;
