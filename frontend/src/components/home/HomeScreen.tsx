import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useHomeScreen from './useHomeScreen';
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
import { Button, Chip, InputAdornment, LinearProgress, Skeleton, TextField } from '@mui/material';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import styles from './HomeScreen.module.css';

type QuickAction = {
	id: string;
	label: string;
	description: string;
	route: string;
	icon: ReactNode;
};

const matchesSearch = (query: string, ...values: Array<string | undefined>) => {
	const normalizedQuery = query.trim().toLowerCase();
	if (!normalizedQuery) {
		return true;
	}

	return values.some((value) => value?.toLowerCase().includes(normalizedQuery));
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

const getRecentFileRoute = (file: { id: number; format: string }) => {
	const fileType = getFileTypeInfo(file.format);

	if (fileType.type === 'video') {
		return `${appRoutes.videoPlayerBase}/${file.id}`;
	}

	if (fileType.type === 'image') {
		return appRoutes.images;
	}

	if (fileType.type === 'audio') {
		return appRoutes.music;
	}

	return appRoutes.files;
};

const HomeScreen = () => {
	const { t } = useI18n();
	const {
		searchQuery,
		setSearchQuery,
		recentFiles,
		videoContinueItems,
		videoResume,
		musicResume,
		analytics,
		isAnalyticsLoading,
		isVideoLoading,
		isMusicLoading,
	} = useHomeScreen();

	const quickActions = useMemo<QuickAction[]>(() => [
		{
			id: 'files',
			label: t('FILES'),
			description: t('HOME_LIBRARY_DESCRIPTION'),
			route: appRoutes.files,
			icon: <FolderOpen size={18} />,
		},
		{
			id: 'favorites',
			label: t('STARRED_FILES'),
			description: t('HOME_LIBRARY_DESCRIPTION'),
			route: appRoutes.favorites,
			icon: <Heart size={18} />,
		},
		{
			id: 'images',
			label: t('NAV_IMAGES'),
			description: t('HOME_MEDIA_DESCRIPTION'),
			route: appRoutes.images,
			icon: <ImageIcon size={18} />,
		},
		{
			id: 'music',
			label: t('NAV_MUSIC'),
			description: t('HOME_MEDIA_DESCRIPTION'),
			route: appRoutes.music,
			icon: <Music2 size={18} />,
		},
		{
			id: 'videos',
			label: t('NAV_VIDEOS'),
			description: t('HOME_MEDIA_DESCRIPTION'),
			route: appRoutes.videos,
			icon: <Film size={18} />,
		},
		{
			id: 'analytics',
			label: t('ANALYTICS'),
			description: t('HOME_SYSTEM_DESCRIPTION'),
			route: appRoutes.analytics,
			icon: <BarChart3 size={18} />,
		},
		{
			id: 'settings',
			label: t('SETTINGS'),
			description: t('HOME_SYSTEM_DESCRIPTION'),
			route: appRoutes.settings,
			icon: <Settings2 size={18} />,
		},
	], [t]);

	const filteredActions = useMemo(
		() => quickActions.filter((action) => matchesSearch(searchQuery, action.label, action.description)),
		[quickActions, searchQuery],
	);

	const filteredRecentFiles = useMemo(
		() => recentFiles.filter((file) => matchesSearch(searchQuery, file.name, file.parent_path, file.format)),
		[recentFiles, searchQuery],
	);

	const filteredVideoItems = useMemo(
		() => videoContinueItems.filter((item) => matchesSearch(searchQuery, item.video.name, item.video.parent_path)),
		[searchQuery, videoContinueItems],
	);

	const supportsMusicSection = matchesSearch(
		searchQuery,
		t('MUSIC_NOW_PLAYING'),
		t('HOME_MUSIC_DESCRIPTION'),
		musicResume?.track.name,
		musicResume?.track.metadata?.title,
		musicResume?.track.metadata?.artist,
	);

	const supportsStatusSection = matchesSearch(
		searchQuery,
		t('STATUS_SYSTEM_TITLE'),
		t('HOME_STATUS_DESCRIPTION'),
		analytics?.health.status,
	);

	const searchHasResults = filteredActions.length > 0
		|| filteredRecentFiles.length > 0
		|| filteredVideoItems.length > 0
		|| (supportsMusicSection && Boolean(musicResume))
		|| (supportsStatusSection && Boolean(analytics));

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
		? filteredVideoItems.filter((item) => item.video.id !== videoResume.video.id)
		: filteredVideoItems;

	return (
		<div className={styles.page}>
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
						value={searchQuery}
						onChange={(event) => setSearchQuery(event.target.value)}
						placeholder={t('SEARCH_PLACEHOLDER')}
						InputProps={{
							startAdornment: (
								<InputAdornment position='start'>
									<Search size={18} />
								</InputAdornment>
							),
						}}
					/>
					<div className={styles.searchMeta}>
						<p className={styles.searchHint}>{t('HOME_SEARCH_HELP')}</p>
						{searchQuery ? (
							<Button variant='text' onClick={() => setSearchQuery('')}>
								{t('HOME_CLEAR_SEARCH')}
							</Button>
						) : null}
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
						{filteredActions.map((action) => (
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

					{searchQuery && !searchHasResults ? (
						<div className={styles.emptyState}>
							<p className={styles.emptyTitle}>{t('HOME_SEARCH_EMPTY')}</p>
						</div>
					) : null}
				</div>
			</section>

			<div className={styles.contentGrid}>
				<section className={styles.panel}>
					<div className={styles.sectionHeader}>
						<div>
							<h2 className={styles.sectionTitle}>{t('RECENT_FILES')}</h2>
							<p className={styles.sectionDescription}>{t('HOME_RECENT_DESCRIPTION')}</p>
						</div>
						<Button component={RouterLink} to={appRoutes.files} variant='text'>
							{t('FILES')}
						</Button>
					</div>

					{isAnalyticsLoading ? (
						<div className={styles.recentList}>
							{Array.from({ length: 3 }).map((_, index) => (
								<div key={index} className={styles.recentCard}>
									<Skeleton variant='rectangular' height={72} />
								</div>
							))}
						</div>
					) : filteredRecentFiles.length > 0 ? (
						<div className={styles.recentList}>
							{filteredRecentFiles.map((file) => {
								const fileType = getFileTypeInfo(file.format);
								return (
									<RouterLink key={file.id} className={styles.recentCard} to={getRecentFileRoute(file)}>
										<div className={styles.recentIcon}>{t(fileType.description)}</div>
										<div className={styles.recentContent}>
											<div className={styles.recentHeader}>
												<h3 className={styles.recentTitle}>{file.name}</h3>
												<Chip size='small' label={file.format || file.name.split('.').pop() || '--'} />
											</div>
											<p className={styles.recentMeta}>{formatSize(file.size_bytes)} · {formatDate(file.created_at)}</p>
											<p className={styles.recentPath}>{file.parent_path}</p>
										</div>
									</RouterLink>
								);
							})}
						</div>
					) : (
						<div className={styles.emptyState}>
							<p className={styles.emptyTitle}>{t('HOME_RECENT_EMPTY')}</p>
						</div>
					)}
				</section>

				<section className={styles.panel}>
					<div className={styles.sectionHeader}>
						<div>
							<h2 className={styles.sectionTitle}>{t('MUSIC_NOW_PLAYING')}</h2>
							<p className={styles.sectionDescription}>{t('HOME_MUSIC_DESCRIPTION')}</p>
						</div>
						<Button component={RouterLink} to={appRoutes.music} variant='text'>
							{t('NAV_MUSIC')}
						</Button>
					</div>

					{isMusicLoading ? (
						<Skeleton variant='rounded' height={180} />
					) : supportsMusicSection && musicResume ? (
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
					) : (
						<div className={styles.emptyState}>
							<p className={styles.emptyTitle}>{t('HOME_MUSIC_EMPTY')}</p>
						</div>
					)}
				</section>

				<section className={`${styles.panel} ${styles.videoPanel}`}>
					<div className={styles.sectionHeader}>
						<div>
							<h2 className={styles.sectionTitle}>{t('VIDEO_CONTINUE_WATCHING')}</h2>
							<p className={styles.sectionDescription}>{t('HOME_VIDEO_DESCRIPTION')}</p>
						</div>
						<Button component={RouterLink} to={appRoutes.videos} variant='text'>
							{t('NAV_VIDEOS')}
						</Button>
					</div>

					{isVideoLoading ? (
						<Skeleton variant='rounded' height={220} />
					) : (videoResume || filteredVideoItems.length > 0) ? (
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
												to={`${appRoutes.videoPlayerBase}/${videoResume.video.id}`}
												state={{ from: appRoutes.home, playlistId: videoResume.playlistId ?? undefined }}
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
					) : (
						<div className={styles.emptyState}>
							<p className={styles.emptyTitle}>{t('HOME_VIDEO_EMPTY')}</p>
						</div>
					)}
				</section>

				{supportsStatusSection ? (
					<section className={`${styles.panel} ${styles.statusPanel}`}>
						<div className={styles.sectionHeader}>
							<div>
								<h2 className={styles.sectionTitle}>{t('STATUS_SYSTEM_TITLE')}</h2>
								<p className={styles.sectionDescription}>{t('HOME_STATUS_DESCRIPTION')}</p>
							</div>
							<Button component={RouterLink} to={appRoutes.analytics} variant='text'>
								{t('ANALYTICS')}
							</Button>
						</div>

						{isAnalyticsLoading ? (
							<Skeleton variant='rounded' height={180} />
						) : analytics ? (
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
						) : (
							<div className={styles.emptyState}>
								<p className={styles.emptyTitle}>{t('HOME_STATUS_DESCRIPTION')}</p>
							</div>
						)}
					</section>
				) : null}
			</div>
		</div>
	);
};

export default HomeScreen;
