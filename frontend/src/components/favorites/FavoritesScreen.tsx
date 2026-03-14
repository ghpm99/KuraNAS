import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import useI18n from '@/components/i18n/provider/i18nContext';
import { FileType } from '@/utils';
import { Heart, LayoutGrid, List, Sparkles } from 'lucide-react';
import { ToggleButton, ToggleButtonGroup } from '@mui/material';
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
		handleSelectItem,
		itemCountLabel,
		selectedItem,
		setActiveFilter,
		setViewMode,
		viewMode,
	} = useFavoritesScreen();
	const isFileSelected = selectedItem?.type === FileType.File;
	const workspaceClassName = isFileSelected
		? `${styles.workspace} ${styles.workspaceWithPreview}`
		: styles.workspace;

	return (
		<div className={styles.page}>
			<section className={styles.hero}>
				<div className={styles.heroCopy}>
					<div className={styles.heroEyebrow}>
						<Heart size={16} />
						<span>{t('FAVORITES_EYEBROW')}</span>
					</div>
					<h1 className={styles.heroTitle}>{t('FAVORITES_PAGE_TITLE')}</h1>
					<p className={styles.heroDescription}>{t('FAVORITES_PAGE_DESCRIPTION')}</p>
				</div>

				<div className={styles.heroMeta}>
					<div className={styles.heroMetric}>
						<span className={styles.heroMetricLabel}>{t('FAVORITES_SCOPE_LABEL')}</span>
						<span className={styles.heroMetricValue}>{currentTitle}</span>
						<span className={styles.heroMetricHelp}>{contextPath}</span>
					</div>
					<div className={styles.heroMetric}>
						<span className={styles.heroMetricLabel}>{t('FAVORITES_ACTIVE_FILTER_LABEL')}</span>
						<span className={styles.heroMetricValue}>{activeFilterLabel}</span>
						<span className={styles.heroMetricHelp}>{itemCountLabel}</span>
					</div>
				</div>
			</section>

			<div className={workspaceClassName}>
				<div className={styles.mainColumn}>
					<section className={styles.panel}>
						<div className={styles.contextHeader}>
							<div>
								<p className={styles.contextTitle}>{t('FAVORITES_CONTEXT_LABEL')}</p>
								<nav className={styles.breadcrumb} aria-label={t('FAVORITES_CONTEXT_LABEL')}>
									{breadcrumbSegments.map((segment, index) => (
										<div key={`${segment.label}-${segment.id ?? 'root'}`} className={styles.breadcrumb}>
											{segment.isCurrent ? (
												<span className={styles.breadcrumbCurrent}>{segment.label}</span>
											) : (
												<button
													type='button'
													className={styles.breadcrumbButton}
													onClick={() => handleSelectItem(segment.id)}
												>
													{segment.label}
												</button>
											)}
											{index < breadcrumbSegments.length - 1 ? (
												<span className={styles.breadcrumbSeparator}>/</span>
											) : null}
										</div>
									))}
								</nav>
							</div>

							<div className={styles.contextActions}>
								<ToggleButtonGroup
									size='small'
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
										<ToggleButton key={option.value} value={option.value} aria-label={option.label}>
											<span className={styles.filterButton}>
												<span>{option.label}</span>
												<span className={styles.filterCount}>{option.count}</span>
											</span>
										</ToggleButton>
									))}
								</ToggleButtonGroup>

								<ToggleButtonGroup
									size='small'
									value={viewMode}
									exclusive
									onChange={(_, nextViewMode) => {
										if (nextViewMode) {
											setViewMode(nextViewMode);
										}
									}}
									aria-label={t('FILES_VIEW_SWITCH')}
								>
									<ToggleButton value='grid' aria-label={t('FILES_VIEW_GRID')}>
										<LayoutGrid size={16} />
									</ToggleButton>
									<ToggleButton value='list' aria-label={t('FILES_VIEW_LIST')}>
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
		</div>
	);
};

export default FavoritesScreen;
