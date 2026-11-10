import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import { Button, CircularProgress } from '@mui/material';
import { ChevronLeft, ChevronRight, Trash2 } from 'lucide-react';
import TrashConfirmDialog from './TrashConfirmDialog';
import TrashItemCard from './TrashItemCard';
import TrashListingControls from './TrashListingControls';
import TrashRetentionField from './TrashRetentionField';
import TrashSelectionToolbar from './TrashSelectionToolbar';
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
        isItemBusy,
        isWholeTrashBusy,
        isAnyActionRunning,
        isBatchRunning,
        handleRestore,
        pendingConfirmation,
        requestDeleteForever,
        requestEmptyTrash,
        cancelPendingAction,
        confirmPendingAction,
        confirmationText,
        selectedCount,
        isItemSelected,
        toggleItemSelection,
        selectWholePage,
        clearSelection,
        restoreSelectedItems,
        requestDeleteSelected,
    } = useTrashScreen();

    return (
        <PageContainer>
            <PageHeader
                title={t('TRASH_PAGE_TITLE')}
                subtitle={t('TRASH_PAGE_DESCRIPTION')}
                actions={
                    <Button
                        variant="outlined"
                        color="error"
                        disabled={isAnyActionRunning || items.length === 0}
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
                    <TrashSelectionToolbar
                        selectedCount={selectedCount}
                        isBusy={isWholeTrashBusy || isBatchRunning}
                        onSelectAll={selectWholePage}
                        onClearSelection={clearSelection}
                        onRestoreSelected={restoreSelectedItems}
                        onDeleteSelected={requestDeleteSelected}
                    />
                    {items.map((item) => (
                        <TrashItemCard
                            key={item.id}
                            item={item}
                            isBusy={isItemBusy(item.id)}
                            isSelected={isItemSelected(item.id)}
                            onToggleSelection={toggleItemSelection}
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
                title={t(confirmationText.titleKey)}
                message={t(confirmationText.messageKey, confirmationText.messageVariables)}
                onCancel={cancelPendingAction}
                onConfirm={() => void confirmPendingAction()}
            />
        </PageContainer>
    );
};

export default TrashScreen;
