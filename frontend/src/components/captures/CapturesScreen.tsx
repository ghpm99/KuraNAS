import { Button, Chip, CircularProgress, IconButton } from '@mui/material';
import { ExternalLink, Film, Trash2 } from 'lucide-react';
import type { Capture, CaptureStatus } from '@/types/captures';
import styles from './CapturesScreen.module.css';
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
    const {
        t,
        items,
        isLoading,
        isError,
        isEmpty,
        deletingId,
        removeCapture,
        formatSize,
        captureThumbnailHref,
        captureStatusLabelKey,
    } = useCapturesScreen();

    return (
        <div className={styles.page}>
            <header className={styles.hero}>
                <p className={styles.eyebrow}>{t('CAPTURES_PAGE_EYEBROW')}</p>
                <h1 className={styles.title}>{t('CAPTURES_PAGE_TITLE')}</h1>
                <p className={styles.description}>{t('CAPTURES_PAGE_DESCRIPTION')}</p>
            </header>

            {isLoading && (
                <div className={styles.state}>
                    <CircularProgress size={20} />
                    <span>{t('CAPTURES_LOADING')}</span>
                </div>
            )}

            {isError && <div className={styles.state}>{t('CAPTURES_ERROR')}</div>}

            {isEmpty && <div className={styles.state}>{t('CAPTURES_EMPTY')}</div>}

            {!isLoading && !isError && items.length > 0 && (
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
                                            {new Date(capture.created_at).toLocaleDateString()}
                                        </span>
                                    </div>
                                </div>

                                <div className={styles.actions}>
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
                                    <IconButton
                                        aria-label={t('CAPTURES_DELETE')}
                                        size="small"
                                        disabled={deletingId === capture.id}
                                        onClick={() => removeCapture(capture.id)}
                                    >
                                        <Trash2 size={16} />
                                    </IconButton>
                                </div>
                            </li>
                        );
                    })}
                </ul>
            )}
        </div>
    );
};

export default CapturesScreen;
