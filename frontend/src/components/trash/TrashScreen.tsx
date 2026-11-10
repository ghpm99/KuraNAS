import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import { Button, CircularProgress } from '@mui/material';
import { ChevronLeft, ChevronRight, Trash2 } from 'lucide-react';
import TrashConfirmDialog from './TrashConfirmDialog';
import TrashItemCard from './TrashItemCard';
import TrashListingControls from './TrashListingControls';
import TrashRetentionField from './TrashRetentionField';
import TrashSummaryLine from './TrashSummaryLine';
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
        isSearching,
        sort,
        changeSort,
        searchInput,
        changeSearchInput,
        isMutating,
        handleRestore,
        pendingConfirmation,
        requestDeleteForever,
        requestEmptyTrash,
        cancelPendingAction,
        confirmPendingAction,
    } = useTrashScreen();
    const isEmptyingTrash = pendingConfirmation?.kind === 'emptyTrash';

    return (
        <PageContainer>
            <PageHeader
                title={t('TRASH_PAGE_TITLE')}
                subtitle={t('TRASH_PAGE_DESCRIPTION')}
                actions={
                    <Button
                        variant="outlined"
                        color="error"
                        disabled={isMutating || items.length === 0}
                        onClick={requestEmptyTrash}
                        startIcon={<Trash2 size={16} />}
                    >
                        {t('TRASH_EMPTY_BUTTON')}
                    </Button>
                }
            />

            <TrashSummaryLine />

            <TrashRetentionField />

            <TrashListingControls
                sort={sort}
                onSortChange={changeSort}
                searchInput={searchInput}
                onSearchInputChange={changeSearchInput}
            />

            {isLoading && (
                <div className={styles.state}>
                    <CircularProgress size={20} />
                    <span>{t('TRASH_LOADING')}</span>
                </div>
            )}

            {isError && <div className={styles.state}>{t('TRASH_LOAD_ERROR')}</div>}

            {isEmpty && (
                <div className={styles.state}>
                    {t(isSearching ? 'TRASH_SEARCH_NO_RESULTS' : 'TRASH_EMPTY_STATE')}
                </div>
            )}

            {!isLoading && !isError && items.length > 0 && (
                <div className={styles.list}>
                    {items.map((item) => (
                        <TrashItemCard
                            key={item.id}
                            item={item}
                            isBusy={isMutating}
                            onRestore={(itemId) => void handleRestore(itemId)}
                            onDeleteForever={requestDeleteForever}
                        />
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
            <TrashConfirmDialog
                isOpen={pendingConfirmation !== null}
                title={t(isEmptyingTrash ? 'TRASH_EMPTY_BUTTON' : 'TRASH_DELETE_FOREVER_BUTTON')}
                message={t(isEmptyingTrash ? 'TRASH_EMPTY_CONFIRM' : 'TRASH_DELETE_CONFIRM')}
                onCancel={cancelPendingAction}
                onConfirm={() => void confirmPendingAction()}
            />
        </PageContainer>
    );
};

export default TrashScreen;
