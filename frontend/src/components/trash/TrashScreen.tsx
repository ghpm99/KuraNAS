import { formatSize } from '@/shared/utils/formatSize';
import { Button, Chip, CircularProgress, TextField } from '@mui/material';
import { ArchiveRestore, ChevronLeft, ChevronRight, Trash2 } from 'lucide-react';
import styles from './TrashScreen.module.css';
import { useTrashScreen } from './useTrashScreen';

const TrashScreen = () => {
    const {
        t,
        items,
        page,
        hasNext,
        hasPrev,
        goToNextPage,
        goToPrevPage,
        isLoading,
        isError,
        isEmpty,
        isMutating,
        retentionDays,
        handleRestore,
        handleDeleteForever,
        handleEmptyTrash,
        handleRetentionChange,
    } = useTrashScreen();

    return (
        <div className={styles.page}>
            <header className={styles.hero}>
                <div className={styles.heroText}>
                    <h1 className={styles.title}>{t('TRASH_PAGE_TITLE')}</h1>
                    <p className={styles.description}>{t('TRASH_PAGE_DESCRIPTION')}</p>
                </div>
                <Button
                    variant="outlined"
                    color="error"
                    disabled={isMutating || items.length === 0}
                    onClick={() => void handleEmptyTrash()}
                    startIcon={<Trash2 size={16} />}
                >
                    {t('TRASH_EMPTY_BUTTON')}
                </Button>
            </header>

            {retentionDays !== undefined && (
                <div className={styles.retention}>
                    <TextField
                        size="small"
                        type="number"
                        label={t('TRASH_RETENTION_LABEL')}
                        defaultValue={retentionDays}
                        slotProps={{ htmlInput: { min: 1 } }}
                        onBlur={(event) =>
                            void handleRetentionChange(Number(event.target.value))
                        }
                    />
                    <span className={styles.description}>{t('TRASH_RETENTION_HINT')}</span>
                </div>
            )}

            {isLoading && (
                <div className={styles.state}>
                    <CircularProgress size={20} />
                    <span>{t('TRASH_LOADING')}</span>
                </div>
            )}

            {isError && <div className={styles.state}>{t('TRASH_LOAD_ERROR')}</div>}

            {isEmpty && <div className={styles.state}>{t('TRASH_EMPTY_STATE')}</div>}

            {!isLoading && !isError && items.length > 0 && (
                <div className={styles.list}>
                    {items.map((item) => (
                        <article key={item.id} className={styles.item}>
                            <div className={styles.itemInfo}>
                                <p className={styles.itemPath}>{item.original_path}</p>
                                <div className={styles.itemMeta}>
                                    <Chip
                                        size="small"
                                        variant="outlined"
                                        label={`${t('TRASH_DELETED_AT_LABEL')}: ${new Date(item.deleted_at).toLocaleString()}`}
                                    />
                                    <Chip
                                        size="small"
                                        variant="outlined"
                                        label={`${t('TRASH_SIZE_LABEL')}: ${formatSize(item.size)}`}
                                    />
                                </div>
                            </div>
                            <div className={styles.itemActions}>
                                <Button
                                    variant="contained"
                                    disabled={isMutating}
                                    onClick={() => void handleRestore(item.id)}
                                    startIcon={<ArchiveRestore size={16} />}
                                >
                                    {t('TRASH_RESTORE_BUTTON')}
                                </Button>
                                <Button
                                    variant="outlined"
                                    color="error"
                                    disabled={isMutating}
                                    onClick={() => void handleDeleteForever(item.id)}
                                    startIcon={<Trash2 size={16} />}
                                >
                                    {t('TRASH_DELETE_FOREVER_BUTTON')}
                                </Button>
                            </div>
                        </article>
                    ))}
                </div>
            )}

            {(hasPrev || hasNext) && (
                <div className={styles.pager}>
                    <Button
                        variant="outlined"
                        size="small"
                        disabled={!hasPrev}
                        onClick={goToPrevPage}
                        startIcon={<ChevronLeft size={16} />}
                    >
                        {t('TRASH_PREV_PAGE')}
                    </Button>
                    <span>{page}</span>
                    <Button
                        variant="outlined"
                        size="small"
                        disabled={!hasNext}
                        onClick={goToNextPage}
                        endIcon={<ChevronRight size={16} />}
                    >
                        {t('TRASH_NEXT_PAGE')}
                    </Button>
                </div>
            )}
        </div>
    );
};

export default TrashScreen;
