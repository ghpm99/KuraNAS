import { Button } from '@mui/material';
import { Link } from 'react-router-dom';
import { getMusicRoute } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useMusicHomeScreen } from '@/components/music/useMusicHomeScreen';
import styles from './MusicHomeScreen.module.css';

const MusicHomeScreen = () => {
	const { t } = useI18n();
	const {
		albumHighlights,
		artistHighlights,
		currentTrackArtist,
		currentTrackTitle,
		featuredPlaylists,
		hasQueue,
		isActionPending,
		isLoadingPlaylists,
		nextTracks,
		openQueue,
		playAlbum,
		playArtist,
		playPlaylist,
		playbackContext,
		queueCount,
		returnToContextHref,
		status,
		totalAlbums,
		totalArtists,
		totalPlaylists,
		totalTracks,
	} = useMusicHomeScreen();

	const playbackContextLabel = playbackContext ? t(playbackContext.labelKey, playbackContext.labelParams) : '';

	return (
		<div className={styles.content}>
			<div className={styles.heroGrid}>
				<section className={`${styles.panel} ${styles.heroPanel}`}>
					<span className={styles.panelLabel}>{t('MUSIC_HOME_CONTINUE_LABEL')}</span>
					<h2 className={styles.panelTitle}>
						{hasQueue ? t('MUSIC_HOME_QUEUE_READY') : t('MUSIC_HOME_QUEUE_EMPTY')}
					</h2>
					<p className={styles.panelDescription}>{t('MUSIC_HOME_CONTINUE_DESCRIPTION')}</p>

					<div className={styles.trackMeta}>
						<strong>{hasQueue ? currentTrackTitle : t('MUSIC_HOME_QUEUE_EMPTY_STATE')}</strong>
						<span>{hasQueue ? currentTrackArtist : t('MUSIC_HOME_QUEUE_EMPTY_HELP')}</span>
						<span className={styles.metricCaption}>{t('MUSIC_HOME_QUEUE_COUNT', { count: String(queueCount) })}</span>
					</div>

					{playbackContextLabel && (
						<p className={styles.contextLabel}>{t('MUSIC_PLAYBACK_FROM', { context: playbackContextLabel })}</p>
					)}

					{nextTracks.length > 0 && (
						<div className={styles.nextTracks}>
							<span className={styles.subsectionLabel}>{t('MUSIC_HOME_UP_NEXT')}</span>
							<div className={styles.nextTrackList}>
								{nextTracks.map((track) => (
									<div key={track.id} className={styles.nextTrackCard}>
										<strong>{track.title}</strong>
										<span>{track.artist}</span>
									</div>
								))}
							</div>
						</div>
					)}

					<div className={styles.actions}>
						<Button variant='contained' size='small' onClick={openQueue} disabled={!hasQueue}>
							{t('MUSIC_HOME_OPEN_QUEUE')}
						</Button>
						{returnToContextHref && (
							<Button component={Link} to={returnToContextHref} variant='outlined' size='small'>
								{t('MUSIC_HOME_RETURN_TO_CONTEXT')}
							</Button>
						)}
					</div>
				</section>

				<section className={styles.panel}>
					<span className={styles.panelLabel}>{t('MUSIC_HOME_LIBRARY_LABEL')}</span>
					<h2 className={styles.panelTitle}>{t('MUSIC_HOME_LIBRARY_STATUS')}</h2>
					<p className={styles.panelDescription}>
						{status === 'pending' ? t('MUSIC_HOME_LIBRARY_LOADING') : t('MUSIC_HOME_LIBRARY_READY')}
					</p>
					<div className={styles.metricGrid}>
						<div className={styles.metricItem}>
							<strong className={styles.metricValue}>{totalTracks}</strong>
							<span className={styles.metricCaption}>{t('MUSIC_HOME_LIBRARY_TRACKS')}</span>
						</div>
						<div className={styles.metricItem}>
							<strong className={styles.metricValue}>{totalArtists}</strong>
							<span className={styles.metricCaption}>{t('MUSIC_HOME_ARTISTS')}</span>
						</div>
						<div className={styles.metricItem}>
							<strong className={styles.metricValue}>{totalAlbums}</strong>
							<span className={styles.metricCaption}>{t('MUSIC_ALBUMS')}</span>
						</div>
						<div className={styles.metricItem}>
							<strong className={styles.metricValue}>{totalPlaylists}</strong>
							<span className={styles.metricCaption}>{t('MUSIC_PLAYLISTS')}</span>
						</div>
					</div>
				</section>
			</div>

			<section className={styles.section}>
				<div className={styles.sectionHeader}>
					<div>
						<span className={styles.subsectionLabel}>{t('MUSIC_PLAYLISTS')}</span>
						<h2 className={styles.sectionTitle}>{t('MUSIC_HOME_FEATURED_PLAYLISTS')}</h2>
					</div>
					<Button component={Link} to={getMusicRoute('playlists')} variant='text' size='small'>
						{t('MUSIC_HOME_OPEN_SECTION')}
					</Button>
				</div>

				<div className={styles.cardGrid}>
					{isLoadingPlaylists && featuredPlaylists.length === 0 ? (
						<div className={styles.emptyCard}>{t('LOADING')}</div>
					) : featuredPlaylists.length === 0 ? (
						<div className={styles.emptyCard}>{t('MUSIC_HOME_FEATURED_PLAYLISTS_EMPTY')}</div>
					) : (
						featuredPlaylists.map((playlist) => (
							<article key={playlist.id} className={styles.sectionCard}>
								<span className={styles.cardEyebrow}>{playlist.is_system ? t('MUSIC_HOME_SYSTEM_PLAYLIST') : t('MUSIC_PLAYLISTS')}</span>
								<h3 className={styles.cardTitle}>{playlist.name}</h3>
								<p className={styles.cardDescription}>{playlist.description || t('MUSIC_PLAYLISTS_DESCRIPTION')}</p>
								<span className={styles.metricCaption}>
									{playlist.track_count} {t('MUSIC_TRACKS_COUNT')}
								</span>
								<div className={styles.actions}>
									<Button
										variant='contained'
										size='small'
										onClick={() => playPlaylist(playlist.id, playlist.name)}
										disabled={isActionPending(playlist.actionKey)}
									>
										{isActionPending(playlist.actionKey) ? t('LOADING') : t('MUSIC_HOME_PLAY_NOW')}
									</Button>
									<Button component={Link} to={playlist.href} variant='outlined' size='small'>
										{t('MUSIC_HOME_OPEN_SECTION')}
									</Button>
								</div>
							</article>
						))
					)}
				</div>
			</section>

			<section className={styles.section}>
				<div className={styles.sectionHeader}>
					<div>
						<span className={styles.subsectionLabel}>{t('MUSIC_ARTISTS')}</span>
						<h2 className={styles.sectionTitle}>{t('MUSIC_HOME_RECENT_ARTISTS')}</h2>
					</div>
					<Button component={Link} to={getMusicRoute('artists')} variant='text' size='small'>
						{t('MUSIC_HOME_OPEN_SECTION')}
					</Button>
				</div>

				<div className={styles.cardGrid}>
					{artistHighlights.length === 0 ? (
						<div className={styles.emptyCard}>{t('MUSIC_HOME_RECENT_ARTISTS_EMPTY')}</div>
					) : (
						artistHighlights.map((artist) => (
							<article key={artist.artist} className={styles.sectionCard}>
								<span className={styles.cardEyebrow}>{t('MUSIC_ARTISTS')}</span>
								<h3 className={styles.cardTitle}>{artist.artist}</h3>
								<p className={styles.cardDescription}>
									{artist.trackCount} {t('MUSIC_TRACKS_COUNT')} · {artist.albumCount} {t('MUSIC_ALBUMS')}
								</p>
								<div className={styles.actions}>
									<Button
										variant='contained'
										size='small'
										onClick={() => playArtist(artist.artist)}
										disabled={isActionPending(artist.actionKey)}
									>
										{isActionPending(artist.actionKey) ? t('LOADING') : t('MUSIC_HOME_PLAY_NOW')}
									</Button>
									<Button component={Link} to={artist.href} variant='outlined' size='small'>
										{t('MUSIC_HOME_OPEN_SECTION')}
									</Button>
								</div>
							</article>
						))
					)}
				</div>
			</section>

			<section className={styles.section}>
				<div className={styles.sectionHeader}>
					<div>
						<span className={styles.subsectionLabel}>{t('MUSIC_ALBUMS')}</span>
						<h2 className={styles.sectionTitle}>{t('MUSIC_HOME_RECENT_ALBUMS')}</h2>
					</div>
					<Button component={Link} to={getMusicRoute('albums')} variant='text' size='small'>
						{t('MUSIC_HOME_OPEN_SECTION')}
					</Button>
				</div>

				<div className={styles.cardGrid}>
					{albumHighlights.length === 0 ? (
						<div className={styles.emptyCard}>{t('MUSIC_HOME_RECENT_ALBUMS_EMPTY')}</div>
					) : (
						albumHighlights.map((album) => (
							<article key={`${album.artist}-${album.album}`} className={styles.sectionCard}>
								<span className={styles.cardEyebrow}>{t('MUSIC_ALBUMS')}</span>
								<h3 className={styles.cardTitle}>{album.album}</h3>
								<p className={styles.cardDescription}>
									{album.artist}
									{album.year ? ` · ${album.year}` : ''}
								</p>
								<span className={styles.metricCaption}>
									{album.trackCount} {t('MUSIC_TRACKS_COUNT')}
								</span>
								<div className={styles.actions}>
									<Button
										variant='contained'
										size='small'
										onClick={() => playAlbum(album.album)}
										disabled={isActionPending(album.actionKey)}
									>
										{isActionPending(album.actionKey) ? t('LOADING') : t('MUSIC_HOME_PLAY_NOW')}
									</Button>
									<Button component={Link} to={album.href} variant='outlined' size='small'>
										{t('MUSIC_HOME_OPEN_SECTION')}
									</Button>
								</div>
							</article>
						))
					)}
				</div>
			</section>
		</div>
	);
};

export default MusicHomeScreen;
