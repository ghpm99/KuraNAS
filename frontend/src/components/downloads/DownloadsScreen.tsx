import { Button, Chip, CircularProgress } from '@mui/material';
import { Download } from 'lucide-react';
import styles from './DownloadsScreen.module.css';
import { useDownloadsScreen } from './useDownloadsScreen';

const DownloadsScreen = () => {
    const {
        t,
        items,
        isLoading,
        isError,
        isEmpty,
        hasBrowserExtension,
        formatSize,
        buildDownloadHref,
    } = useDownloadsScreen();

    return (
        <div className={styles.page}>
            <header className={styles.hero}>
                <p className={styles.eyebrow}>{t('DOWNLOADS_PAGE_EYEBROW')}</p>
                <h1 className={styles.title}>{t('DOWNLOADS_PAGE_TITLE')}</h1>
                <p className={styles.description}>{t('DOWNLOADS_PAGE_DESCRIPTION')}</p>
            </header>

            {isLoading && (
                <div className={styles.state}>
                    <CircularProgress size={20} />
                    <span>{t('DOWNLOADS_LOADING')}</span>
                </div>
            )}

            {isError && <div className={styles.state}>{t('DOWNLOADS_ERROR')}</div>}

            {isEmpty && <div className={styles.state}>{t('DOWNLOADS_EMPTY')}</div>}

            {!isLoading && !isError && items.length > 0 && (
                <div className={styles.grid}>
                    {items.map((item) => (
                        <article key={item.id} className={styles.card}>
                            <div className={styles.cardHeader}>
                                <h2 className={styles.cardTitle}>{item.name}</h2>
                                {item.description && (
                                    <p className={styles.cardDescription}>{item.description}</p>
                                )}
                            </div>

                            <div className={styles.meta}>
                                {item.version && (
                                    <Chip
                                        size="small"
                                        variant="outlined"
                                        label={`${t('DOWNLOADS_VERSION')}: ${item.version}`}
                                    />
                                )}
                                {item.min_os && (
                                    <Chip
                                        size="small"
                                        variant="outlined"
                                        label={`${t('DOWNLOADS_MIN_OS')}: ${item.min_os}`}
                                    />
                                )}
                                <Chip
                                    size="small"
                                    variant="outlined"
                                    label={`${t('DOWNLOADS_SIZE')}: ${formatSize(item.size_bytes)}`}
                                />
                            </div>

                            <div className={styles.actions}>
                                <Button
                                    component="a"
                                    href={buildDownloadHref(item)}
                                    download
                                    variant="contained"
                                    startIcon={<Download size={16} />}
                                >
                                    {t('DOWNLOADS_BUTTON')}
                                </Button>
                            </div>
                        </article>
                    ))}
                </div>
            )}

            {hasBrowserExtension && (
                <section className={styles.instructions}>
                    <h3 className={styles.instructionsTitle}>
                        {t('DOWNLOADS_PLUGIN_INSTRUCTIONS_TITLE')}
                    </h3>
                    <ol className={styles.steps}>
                        <li>{t('DOWNLOADS_PLUGIN_STEP_1')}</li>
                        <li>{t('DOWNLOADS_PLUGIN_STEP_2')}</li>
                        <li>{t('DOWNLOADS_PLUGIN_STEP_3')}</li>
                        <li>{t('DOWNLOADS_PLUGIN_STEP_4')}</li>
                    </ol>
                </section>
            )}
        </div>
    );
};

export default DownloadsScreen;
