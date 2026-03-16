import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import useI18n from '@/components/i18n/provider/i18nContext';
import FolderTree from '@/components/layout/Sidebar/components/folderTree';
import Tabs from '@/components/tabs';
import { Button, Drawer, ToggleButton, ToggleButtonGroup } from '@mui/material';
import { FolderOpen, LayoutGrid, List, PanelLeft } from 'lucide-react';
import { FileType } from '@/utils';
import { useNavigate } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import useFilesExplorerScreen from './useFilesExplorerScreen';
import styles from './FilesExplorerScreen.module.css';

const FilesExplorerScreen = () => {
	const { t } = useI18n();
	const {
		breadcrumbSegments,
		closeMobileTree,
		contextLabel,
		itemCountLabel,
		mobileTreeOpen,
		openMobileTree,
		selectedItem,
		setViewMode,
		viewMode,
	} = useFilesExplorerScreen();
	const navigate = useNavigate();
	const isFileSelected = selectedItem?.type === FileType.File;
	const workspaceClassName = isFileSelected
		? `${styles.workspace} ${styles.workspaceWithPreview}`
		: styles.workspace;

	return (
		<div className={styles.page}>
			<section className={styles.hero}>
				<div className={styles.heroEyebrow}>
					<FolderOpen size={16} />
					<span>{t('FILES_EXPLORER_EYEBROW')}</span>
				</div>
				<div>
					<h1 className={styles.heroTitle}>{t('FILES_PAGE_TITLE')}</h1>
					<p className={styles.heroDescription}>{t('FILES_PAGE_DESCRIPTION')}</p>
				</div>
			</section>

			<div className={workspaceClassName}>
				<div className={styles.mainColumn}>
					<section className={styles.panel}>
						<div className={styles.contextHeader}>
							<div>
								<p className={styles.contextTitle}>{t('FILES_CURRENT_LOCATION')}</p>
								<nav className={styles.breadcrumb} aria-label={t('FILES_CURRENT_LOCATION')}>
									{breadcrumbSegments.map((segment, index) => (
										<div key={`${segment.label}-${segment.id ?? 'root'}`} className={styles.breadcrumb}>
											{segment.isCurrent ? (
												<span className={styles.breadcrumbCurrent}>{segment.label}</span>
											) : (
												<button
													type='button'
													className={styles.breadcrumbButton}
													onClick={() => {
														const url = segment.path
															? `${appRoutes.files}${segment.path}`
															: appRoutes.files;
														navigate(url);
													}}
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
								<Button
									variant='outlined'
									size='small'
									startIcon={<PanelLeft size={16} />}
									onClick={openMobileTree}
									className={styles.treeButton}
								>
									{t('FILES_OPEN_TREE')}
								</Button>
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
										<span>{t('FILES_VIEW_GRID')}</span>
									</ToggleButton>
									<ToggleButton value='list' aria-label={t('FILES_VIEW_LIST')}>
										<List size={16} />
										<span>{t('FILES_VIEW_LIST')}</span>
									</ToggleButton>
								</ToggleButtonGroup>
							</div>
						</div>

						<div className={styles.contextMeta}>
							<span>{contextLabel}</span>
							<span>{itemCountLabel}</span>
							{selectedItem ? <span>{selectedItem.name}</span> : null}
						</div>
					</section>

					<section className={`${styles.panel} ${styles.toolbarCard}`}>
						<ActionBar />
					</section>

					{!isFileSelected ? (
						<section className={`${styles.panel} ${styles.tabsCard}`}>
							<Tabs />
						</section>
					) : null}

					<section className={`${styles.panel} ${styles.contentCard}`}>
						<FileContent showHeading={false} viewMode={viewMode} />
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

			<Drawer anchor='left' open={mobileTreeOpen} onClose={closeMobileTree}>
				<div className={styles.drawerContent}>
					<p className={styles.drawerTitle}>{t('FILES_OPEN_TREE')}</p>
					<FolderTree />
				</div>
			</Drawer>
		</div>
	);
};

export default FilesExplorerScreen;
