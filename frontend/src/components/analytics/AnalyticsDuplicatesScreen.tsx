import { useSearchParams } from 'react-router-dom';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import useI18n from '@/components/i18n/provider/i18nContext';
import DuplicateCategorySelect from '@/components/analytics/DuplicateCategorySelect';
import DuplicateDeleteDialog from '@/components/analytics/DuplicateDeleteDialog';
import DuplicateDeleteResultDialog from '@/components/analytics/DuplicateDeleteResultDialog';
import DuplicateGroupCard from '@/components/analytics/DuplicateGroupCard';
import DuplicateSelectionBar from '@/components/analytics/DuplicateSelectionBar';
import {
    duplicateCategoryQueryParam,
    parseDuplicateCategory,
} from '@/components/analytics/duplicateCategoryFilter';
import { countGroupsLeftWithoutCopies } from '@/components/analytics/emptiedDuplicateGroups';
import { useDuplicateCopiesDeletion } from '@/components/analytics/useDuplicateCopiesDeletion';
import { useDuplicateGroupsPage } from '@/components/analytics/useDuplicateGroupsPage';
import { useItemSelection } from '@/shared/selection/useItemSelection';
import type { DuplicateCategory, DuplicateCopy } from '@/types/analytics';
import styles from './AnalyticsDuplicatesScreen.module.css';

const readCopyId = (copy: DuplicateCopy) => copy.id;

const AnalyticsDuplicatesScreen = () => {
    const { t } = useI18n();
    const [searchParams, setSearchParams] = useSearchParams();
    const category = parseDuplicateCategory(searchParams.toString());
    const {
        groups,
        isPending,
        isError,
        isFetchNextPageError,
        isFetchingNextPage,
        hasNextPage,
        fetchNextPage,
        errorMessage,
        retry,
    } = useDuplicateGroupsPage(category);
    const selection = useItemSelection<DuplicateCopy>(category ?? 'all', readCopyId);
    const deletion = useDuplicateCopiesDeletion({ onCopiesTrashed: selection.deselect });

    const changeCategory = (nextCategory: DuplicateCategory | undefined) => {
        const nextSearchParams = new URLSearchParams(searchParams);
        if (nextCategory) {
            nextSearchParams.set(duplicateCategoryQueryParam, nextCategory);
        } else {
            nextSearchParams.delete(duplicateCategoryQueryParam);
        }
        setSearchParams(nextSearchParams);
    };

    const renderList = () => {
        if (isPending) {
            return <p role="status">{t('LOADING')}</p>;
        }
        return (
            <ListFeedback
                isEmpty={groups.length === 0}
                isError={isError}
                isNextPageError={isFetchNextPageError}
                errorTitle={t('ANALYTICS_DUPLICATES_ERROR_TITLE')}
                errorMessage={errorMessage}
                onRetry={retry}
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                onLoadMore={fetchNextPage}
                emptyState={<p>{t('ANALYTICS_DUPLICATES_EMPTY')}</p>}
            >
                <ul className={styles.groups} aria-label={t('ANALYTICS_SECTION_DUPLICATES')}>
                    {groups.map((group) => (
                        <DuplicateGroupCard
                            key={group.signature}
                            group={group}
                            selection={selection}
                            onDeleteCopy={(copy) => deletion.requestDeletion([copy])}
                        />
                    ))}
                </ul>
            </ListFeedback>
        );
    };

    return (
        <div className={styles.screen}>
            <DuplicateCategorySelect category={category} onChange={changeCategory} />
            {selection.hasSelection ? (
                <DuplicateSelectionBar
                    selectedCount={selection.selectedCount}
                    onClear={selection.clear}
                    onDelete={() => deletion.requestDeletion(selection.selectedItems)}
                />
            ) : null}
            {renderList()}
            <DuplicateDeleteDialog
                copiesCount={deletion.copiesAwaitingConfirmation.length}
                groupsLeftWithoutCopiesCount={countGroupsLeftWithoutCopies(
                    groups,
                    deletion.copiesAwaitingConfirmation
                )}
                isDeleting={deletion.isDeleting}
                onCancel={deletion.cancelDeletion}
                onConfirm={() => void deletion.confirmDeletion()}
            />
            <DuplicateDeleteResultDialog
                outcome={deletion.outcome}
                onClose={deletion.dismissOutcome}
            />
        </div>
    );
};

export default AnalyticsDuplicatesScreen;
