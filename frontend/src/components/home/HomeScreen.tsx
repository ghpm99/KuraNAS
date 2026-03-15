import { appRoutes } from '@/app/routes';
import useMediaOpener from '@/components/hooks/useMediaOpener/useMediaOpener';
import useI18n from '@/components/i18n/provider/i18nContext';
import useGlobalSearch from '@/components/search/useGlobalSearch';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useHomeScreen from './useHomeScreen';
import type { HomeRecentFile, HomeFavoriteFile, HomeRecentImage } from './useHomeScreen';
import { formatDate, formatSize, getFileTypeInfo } from '@/utils';
import {
	AlertCircle,
	ArrowRight,
	BarChart3,
	Film,
	FolderOpen,
	HardDrive,
	Heart,
	Image as ImageIcon,
	LibraryBig,
	Music2,
	Search,
	Settings2,
} from 'lucide-react';
import { Button, Chip, InputAdornment, LinearProgress, TextField } from '@mui/material';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import styles from './HomeScreen.module.css';
import HomeSectionShell from './HomeSectionShell';

type QuickAction = {
	id: string;
	label: string;
	description: string;
	route: string;
	icon: ReactNode;
};

const getAnalyticsStatusKey = (status: 'ok' | 'scanning' | 'error') => {
	switch (status) {
		case 'scanning':
			return 'ANALYTICS_STATUS_SCANNING';
		case 'error':
			return 'ANALYTICS_STATUS_ERROR';
		default:
			return 'ANALYTICS_STATUS_OK';
	}
};

const imageThumbnailUrl = (id: number) =>
	`${getApiV1BaseUrl()}/files/thumbnail/${id}?width=480&height=360`;

// --- File list rendering (shared by recent files and favorites) ---

type FileItem = { id: number; name: string; path: string; parent_path: string; format: string; size_bytes?: number; size?: number; created_at: string; updated_at?: string };

const FileListCard = ({ file, onClick, t }: { file: FileItem; onClick: () => void; t: (key: string) => string }) => {
	const fileType = getFileTypeInfo(file.format);
	return (
		<button type='button' className={styles.recentCardButton} onClick={onClick}>
			<div className={styles.recentCard}>
				<div className={styles.recentIcon}>{t(fileType.description)}</div>
				<div className={styles.recentContent}>
					<div className={styles.recentHeader}>
						<h3 className={styles.recentTitle}>{file.name}</h3>
						<Chip size='small' label={file.format || file.name.split('.').pop() || '--'} />
					</div>
					<p className={styles.recentMeta}>
						{formatSize(file.size_bytes ?? file.size ?? 0)} · {formatDate(file.updated_at || file.created_at)}
					</p>
					<p className={styles.recentPath}>{file.parent_path}</p>
				</div>
			</div>
		</button>
	);
};

const HomeScreen = () => {
	const { t } = useI18n();
	const navigate = useNavigate();
	const { openMediaItem } = useMediaOpener();
	const { openSearch, shortcut } = useGlobalSearch();
	const {
		recentFiles,
		favoriteItems = [],
		recentImages = [],
		videoContinueItems,
		videoResume,
		musicResume,
		analytics,
		isAnalyticsLoading,
		isFavoritesLoading = false,
		isImagesLoading = false,
		isVideoLoading,
		isMusicLoading,
	} = useHomeScreen();

	const quickActions = useMemo<QuickAction[]>(() => [
		{ id: 'files', label: t('FILES'), description: t('HOME_LIBRARY_DESCRIPTION'), route: appRoutes.files, icon: <FolderOpen size={18} /> },
		{ id: 'favorites', label: t('STARRED_FILES'), description: t('HOME_LIBRARY_DESCRIPTION'), route: appRoutes.favorites, icon: <Heart size={18} /> },
		{ id: 'images', label: t('NAV_IMAGES'), description: t('HOME_MEDIA_DESCRIPTION'), route: appRoutes.images, icon: <ImageIcon size={18} /> },
		{ id: 'music', label: t('NAV_MUSIC'), description: t('HOME_MEDIA_DESCRIPTION'), route: appRoutes.music, icon: <Music2 size={18} /> },
		{ id: 'videos', label: t('NAV_VIDEOS'), description: t('HOME_MEDIA_DESCRIPTION'), route: appRoutes.videos, icon: <Film size={18} /> },
		{ id: 'analytics', label: t('ANALYTICS'), description: t('HOME_SYSTEM_DESCRIPTION'), route: appRoutes.analytics, icon: <BarChart3 size={18} /> },
		{ id: 'settings', label: t('SETTINGS'), description: t('HOME_SYSTEM_DESCRIPTION'), route: appRoutes.settings, icon: <Settings2 size={18} /> },
	], [t]);

	const storageUsedLabel = analytics
		? `${formatSize(analytics.storage.used_bytes)} / ${formatSize(analytics.storage.total_bytes)}`
		: '--';
	const storageFreeLabel = analytics ? formatSize(analytics.storage.free_bytes) : '--';
	const analyticsStatusLabel = analytics ? t(getAnalyticsStatusKey(analytics.health.status)) : t('LOADING');
	const indexedFilesLabel = analytics ? analytics.health.indexed_files.toLocaleString() : '--';
	const recentErrorsLabel = analytics ? analytics.health.errors_last_24h.toLocaleString() : '--';
	const lastScanLabel = analytics?.health.last_scan_at ? formatDate(analytics.health.last_scan_at) : t('HOME_LAST_SCAN_EMPTY');

	const musicTitle = musicResume?.track.metadata?.title || musicResume?.track.name || '';
	const musicArtist = musicResume?.track.metadata?.artist || t('HOME_UNKNOWN_ARTIST');
	const featuredVideoItems = videoResume
		? videoContinueItems.filter((item) => item.video.id !== videoResume.video.id)
		: videoContinueItems;

	const handleOpenFile = (file: HomeRecentFile | HomeFavoriteFile) => {
		if (!openMediaItem(file)) {
			navigate({
				pathname: appRoutes.files,
				search: `?path=${encodeURIComponent(file.path)}`,
			});
		}
	};

	const handleOpenRecentImage = (image: HomeRecentImage) => {
		openMediaItem(image);
	};

	return (
		<div className={styles.page}>
			{/* Hero section */}
			<section className={styles.hero}>
				<div className={styles.heroCopy}>
					<div className={styles.heroEyebrow}>
						<LibraryBig size={16} />
						<span>{t('HOME_HERO_EYEBROW')}</span>
					</div>
					<h1 className={styles.heroTitle}>{t('HOME_PAGE_TITLE')}</h1>
					<p className={styles.heroDescription}>{t('HOME_PAGE_DESCRIPTION')}</p>
				</div>

				<div className={styles.heroSearch}>
					<TextField
						fullWidth
						value=''
						onClick={openSearch}
						onFocus={openSearch}
						placeholder={t('SEARCH_PLACEHOLDER')}
						inputProps={{ readOnly: true, 'aria-label': t('GLOBAL_SEARCH_OPEN') }}
						InputProps={{
							startAdornment: (
								<InputAdornment position='start'>
									<Search size={18} />
								</InputAdornment>
							),
						}}
					/>
					<div className={styles.searchMeta}>
						<p className={styles.searchHint}>{t('GLOBAL_SEARCH_SHORTCUT', { shortcut })}</p>
						<Button variant='text' onClick={openSearch}>{t('HOME_OPEN_SECTION')}</Button>
					</div>
				</div>

				<div className={styles.heroMetrics}>
					<div className={styles.metricCard}>
						<div className={styles.metricLabel}>{t('HOME_STORAGE_LABEL')}</div>
						<div className={styles.metricValue}>{storageUsedLabel}</div>
						<div className={styles.metricHelp}>{storageFreeLabel}</div>
					</div>
					<div className={styles.metricCard}>
						<div className={styles.metricLabel}>{t('HOME_INDEX_LABEL')}</div>
						<div className={styles.metricValue}>{indexedFilesLabel}</div>
						<div className={styles.metricHelp}>{analyticsStatusLabel}</div>
					</div>
					<div className={styles.metricCard}>
						<div className={styles.metricLabel}>{t('HOME_LAST_SCAN_LABEL')}</div>
						<div className={styles.metricValue}>{lastScanLabel}</div>
						<div className={styles.metricHelp}>{t('HOME_ERRORS_LABEL')}: {recentErrorsLabel}</div>
					</div>
				</div>

				<div className={styles.actionsSection}>
					<div className={styles.sectionHeader}>
						<div>
							<h2 className={styles.sectionTitle}>{t('HOME_LIBRARY_TITLE')}</h2>
							<p className={styles.sectionDescription}>{t('HOME_LIBRARY_DESCRIPTION')}</p>
						</div>
					</div>
					<div className={styles.actionsGrid}>
						{quickActions.map((action) => (
							<article key={action.id} className={styles.actionCard}>
								<div className={styles.actionIcon}>{action.icon}</div>
								<div className={styles.actionContent}>
									<h3 className={styles.actionTitle}>{action.label}</h3>
									<p className={styles.actionDescription}>{action.description}</p>
								</div>
								<Button component={RouterLink} to={action.route} variant='text' endIcon={<ArrowRight size={16} />}>
									{t('HOME_OPEN_SECTION')}
								</Button>
							</article>
						))}
					</div>
				</div>
			</section>

			{/* Content sections */}
			<div className={styles.contentGrid}>
				{/* Recent files */}
				<HomeSectionShell
					title={t('RECENT_FILES')}
					description={t('HOME_RECENT_DESCRIPTION')}
					linkLabel={t('FILES')}
					linkTo={appRoutes.files}
					isLoading={isAnalyticsLoading}
					isEmpty={recentFiles.length === 0}
					emptyMessage={t('HOME_RECENT_EMPTY')}
				>
					<div className={styles.recentList}>
						{recentFiles.map((file) => (
							<FileListCard key={file.id} file={{ ...file, size_bytes: file.size_bytes }} onClick={() => handleOpenFile(file)} t={t} />
						))}
					</div>
				</HomeSectionShell>

				{/* Recent images */}
				<HomeSectionShell
					title={t('NAV_IMAGES')}
					description={t('HOME_IMAGES_DESCRIPTION')}
					linkLabel={t('NAV_IMAGES')}
					linkTo={appRoutes.images}
					isLoading={isImagesLoading}
					isEmpty={recentImages.length === 0}
					emptyMessage={t('HOME_IMAGES_EMPTY')}
					skeletonVariant='rounded'
					skeletonHeight={132}
				>
					<div className={styles.imageGrid}>
						{recentImages.map((image) => (
							<button
								key={image.id}
								type='button'
								className={styles.imageCard}
								onClick={() => handleOpenRecentImage(image)}
								aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: image.name })}
							>
								<img src={imageThumbnailUrl(image.id)} alt={image.name} className={styles.imageCardImage} loading='lazy' />
								<div className={styles.imageCardOverlay}>
									<strong>{image.name}</strong>
									<span>{formatDate(image.created_at)}</span>
								</div>
							</button>
						))}
					</div>
				</HomeSectionShell>

				{/* Favorites */}
				<HomeSectionShell
					title={t('STARRED_FILES')}
					description={t('HOME_FAVORITES_DESCRIPTION')}
					linkLabel={t('STARRED_FILES')}
					linkTo={appRoutes.favorites}
					isLoading={isFavoritesLoading}
					isEmpty={favoriteItems.length === 0}
					emptyMessage={t('HOME_FAVORITES_EMPTY')}
				>
					<div className={styles.recentList}>
						{favoriteItems.map((file) => (
							<FileListCard key={file.id} file={file} onClick={() => handleOpenFile(file)} t={t} />
						))}
					</div>
				</HomeSectionShell>

				{/* Music */}
				<HomeSectionShell
					title={t('MUSIC_NOW_PLAYING')}
					description={t('HOME_MUSIC_DESCRIPTION')}
					linkLabel={t('NAV_MUSIC')}
					linkTo={appRoutes.music}
					isLoading={isMusicLoading}
					isEmpty={!musicResume}
					emptyMessage={t('HOME_MUSIC_EMPTY')}
					skeletonCount={1}
					skeletonVariant='rounded'
					skeletonHeight={180}
				>
					{musicResume ? (
						<div className={styles.mediaCard}>
							<div className={styles.mediaHeader}>
								<div>
									<h3 className={styles.mediaTitle}>{musicTitle}</h3>
									<p className={styles.mediaSubtitle}>{musicArtist}</p>
								</div>
								<Chip
									size='small'
									color={musicResume.isPlaying ? 'primary' : 'default'}
									label={musicResume.isPlaying ? t('IN_PROGRESS') : t('HOME_RESUME_ACTION')}
								/>
							</div>
							<div className={styles.mediaMeta}>
								<span>{formatSize(musicResume.track.size)}</span>
								<span>{t('HOME_QUEUE_COUNT', { count: String(musicResume.queueCount) })}</span>
							</div>
							<LinearProgress variant='determinate' value={musicResume.progressPercent} className={styles.progress} />
							<div className={styles.mediaMeta}>
								<span>{formatDate(musicResume.track.updated_at)}</span>
								<span>{Math.round(musicResume.progressPercent)}%</span>
							</div>
							<Button component={RouterLink} to={appRoutes.music} variant='contained'>
								{t('HOME_RESUME_ACTION')}
							</Button>
						</div>
					) : null}
				</HomeSectionShell>

				{/* Videos */}
				<HomeSectionShell
					title={t('VIDEO_CONTINUE_WATCHING')}
					description={t('HOME_VIDEO_DESCRIPTION')}
					linkLabel={t('NAV_VIDEOS')}
					linkTo={appRoutes.videos}
					isLoading={isVideoLoading}
					isEmpty={!videoResume && videoContinueItems.length === 0}
					emptyMessage={t('HOME_VIDEO_EMPTY')}
					skeletonCount={1}
					skeletonVariant='rounded'
					skeletonHeight={220}
					className={styles.videoPanel}
				>
					<div className={styles.videoStack}>
						{videoResume ? (
							<article className={styles.videoHeroCard}>
								<img
									className={styles.videoHeroImage}
									src={`${getApiV1BaseUrl()}/files/video-thumbnail/${videoResume.video.id}?width=960&height=540`}
									alt={videoResume.video.name}
								/>
								<div className={styles.videoHeroOverlay}>
									<div>
										<Chip size='small' label={t('VIDEO_CONTINUE_BADGE_RESUME')} />
										<h3 className={styles.mediaTitle}>{videoResume.video.name}</h3>
										<p className={styles.mediaSubtitle}>{videoResume.video.parent_path}</p>
									</div>
									<div className={styles.videoHeroFooter}>
										<LinearProgress variant='determinate' value={videoResume.progressPercent} className={styles.progress} />
										<Button
											component={RouterLink}
											to={`${appRoutes.videoPlayerBase}/${videoResume.video.id}${videoResume.playlistId ? `?playlist=${videoResume.playlistId}` : ''}`}
											state={{ from: appRoutes.home }}
											variant='contained'
										>
											{t('HOME_RESUME_ACTION')}
										</Button>
									</div>
								</div>
							</article>
						) : null}

						<div className={styles.videoList}>
							{featuredVideoItems.map((item) => (
								<RouterLink
									key={item.video.id}
									className={styles.videoCard}
									to={`${appRoutes.videoPlayerBase}/${item.video.id}`}
									state={{ from: appRoutes.home }}
								>
									<img
										className={styles.videoCardImage}
										src={`${getApiV1BaseUrl()}/files/video-thumbnail/${item.video.id}?width=400&height=225`}
										alt={item.video.name}
									/>
									<div className={styles.videoCardBody}>
										<h3 className={styles.videoCardTitle}>{item.video.name}</h3>
										<p className={styles.videoCardMeta}>{item.video.parent_path}</p>
										<LinearProgress variant='determinate' value={item.progress_pct} className={styles.progress} />
									</div>
								</RouterLink>
							))}
						</div>
					</div>
				</HomeSectionShell>

				{/* System status */}
				<HomeSectionShell
					title={t('STATUS_SYSTEM_TITLE')}
					description={t('HOME_STATUS_DESCRIPTION')}
					linkLabel={t('ANALYTICS')}
					linkTo={appRoutes.analytics}
					isLoading={isAnalyticsLoading}
					isEmpty={!analytics}
					emptyMessage={t('HOME_STATUS_DESCRIPTION')}
					skeletonCount={1}
					skeletonVariant='rounded'
					skeletonHeight={180}
					className={styles.statusPanel}
				>
					{analytics ? (
						<div className={styles.statusGrid}>
							<div className={styles.statusCard}>
								<div className={styles.statusLabel}>
									<HardDrive size={16} />
									<span>{t('HOME_STORAGE_LABEL')}</span>
								</div>
								<strong className={styles.statusValue}>{storageUsedLabel}</strong>
								<span className={styles.statusHelp}>{storageFreeLabel}</span>
							</div>
							<div className={styles.statusCard}>
								<div className={styles.statusLabel}>
									<LibraryBig size={16} />
									<span>{t('HOME_INDEX_LABEL')}</span>
								</div>
								<strong className={styles.statusValue}>{indexedFilesLabel}</strong>
								<span className={styles.statusHelp}>{analyticsStatusLabel}</span>
							</div>
							<div className={styles.statusCard}>
								<div className={styles.statusLabel}>
									<AlertCircle size={16} />
									<span>{t('HOME_ERRORS_LABEL')}</span>
								</div>
								<strong className={styles.statusValue}>{recentErrorsLabel}</strong>
								<span className={styles.statusHelp}>{lastScanLabel}</span>
							</div>
						</div>
					) : null}
				</HomeSectionShell>
			</div>
		</div>
	);
};

export default HomeScreen;
