import FileContent from '@/features/files/fileContent';
import FileDetails from '@/features/files/fileDetails';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import { FileType } from '@/utils';
import { LayoutGrid, List, Sparkles } from 'lucide-react';
import { ToggleButton, ToggleButtonGroup } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import useFavoritesScreen from './useFavoritesScreen';
import styles from './FavoritesScreen.module.css';

const FavoritesScreen = () => {
    const { t } = useI18n();
    const {
        activeFilter,
        activeFilterLabel,
        breadcrumbSegments,
        contextPath,
        currentTitle,
        filterOptions,
        filteredItems,
        itemCountLabel,
        selectedItem,
        setActiveFilter,
        setViewMode,
        viewMode,
    } = useFavoritesScreen();
    const navigate = useNavigate();
    const isFileSelected = selectedItem?.type === FileType.File;
    const workspaceClassName = isFileSelected
        ? `${styles.workspace} ${styles.workspaceWithPreview}`
        : styles.workspace;

    return (
        <PageContainer>
            <PageHeader
                title={t('FAVORITES_PAGE_TITLE')}
                subtitle={t('FAVORITES_PAGE_DESCRIPTION')}
            />

            <div className={workspaceClassName}>
                <div className={styles.mainColumn}>
                    <section className={styles.panel}>
                        <div className={styles.contextHeader}>
                            <div>
                                <p className={styles.contextTitle}>
                                    {t('FAVORITES_CONTEXT_LABEL')}
                                </p>
                                <nav
                                    className={styles.breadcrumb}
                                    aria-label={t('FAVORITES_CONTEXT_LABEL')}
                                >
                                    {breadcrumbSegments.map((segment, index) => (
                                        <div
                                            key={`${segment.label}-${segment.id ?? 'root'}`}
                                            className={styles.breadcrumb}
                                        >
                                            {segment.isCurrent ? (
                                                <span className={styles.breadcrumbCurrent}>
                                                    {segment.label}
                                                </span>
                                            ) : (
                                                <button
                                                    type="button"
                                                    className={styles.breadcrumbButton}
                                                    onClick={() => {
                                                        const url = segment.path
                                                            ? `${appRoutes.files}${segment.path}`
                                                            : appRoutes.favorites;
                                                        navigate(url);
                                                    }}
                                                >
                                                    {segment.label}
                                                </button>
                                            )}
                                            {index < breadcrumbSegments.length - 1 ? (
                                                <span className={styles.breadcrumbSeparator}>
                                                    /
                                                </span>
                                            ) : null}
                                        </div>
                                    ))}
                                </nav>
                            </div>

                            <div className={styles.contextActions}>
                                <ToggleButtonGroup
                                    size="small"
                                    value={activeFilter}
                                    exclusive
                                    onChange={(_, nextFilter) => {
                                        if (nextFilter) {
                                            setActiveFilter(nextFilter);
                                        }
                                    }}
                                    aria-label={t('FAVORITES_FILTER_SWITCH')}
                                >
                                    {filterOptions.map((option) => (
                                        <ToggleButton
                                            key={option.value}
                                            value={option.value}
                                            aria-label={option.label}
                                        >
                                            <span className={styles.filterButton}>
                                                <span>{option.label}</span>
                                                <span className={styles.filterCount}>
                                                    {option.count}
                                                </span>
                                            </span>
                                        </ToggleButton>
                                    ))}
                                </ToggleButtonGroup>

                                <ToggleButtonGroup
                                    size="small"
                                    value={viewMode}
                                    exclusive
                                    onChange={(_, nextViewMode) => {
                                        if (nextViewMode) {
                                            setViewMode(nextViewMode);
                                        }
                                    }}
                                    aria-label={t('FILES_VIEW_SWITCH')}
                                >
                                    <ToggleButton value="grid" aria-label={t('FILES_VIEW_GRID')}>
                                        <LayoutGrid size={16} />
                                    </ToggleButton>
                                    <ToggleButton value="list" aria-label={t('FILES_VIEW_LIST')}>
                                        <List size={16} />
                                    </ToggleButton>
                                </ToggleButtonGroup>
                            </div>
                        </div>

                        <div className={styles.contextMeta}>
                            <span>{contextPath}</span>
                            <span>{itemCountLabel}</span>
                            <span>
                                <Sparkles size={14} />
                                <span className={styles.filterValue}>{activeFilterLabel}</span>
                            </span>
                        </div>
                    </section>

                    <section className={`${styles.panel} ${styles.contentCard}`}>
                        <FileContent
                            showHeading={false}
                            viewMode={viewMode}
                            items={filteredItems}
                            title={currentTitle}
                            emptyStateMessage={t('FAVORITES_EMPTY_STATE')}
                        />
                    </section>
                </div>

                {isFileSelected ? (
                    <aside className={styles.previewColumn}>
                        <section className={`${styles.panel} ${styles.previewCard}`}>
                            <FileDetails />
                        </section>
                    </aside>
                ) : null}
            </div>
        </PageContainer>
    );
};

export default FavoritesScreen;
