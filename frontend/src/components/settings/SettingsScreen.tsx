import { appRoutes } from '@/app/routes';
import {
	Alert,
	Button,
	Chip,
	FormControl,
	InputLabel,
	MenuItem,
	Select,
	Switch,
	FormControlLabel,
	TextField,
} from '@mui/material';
import { Link } from 'react-router-dom';
import useSettingsScreen from './useSettingsScreen';
import styles from './SettingsScreen.module.css';

const SettingsScreen = () => {
	const {
		t,
		settings,
		draft,
		isLoading,
		isSaving,
		hasError,
		hasUnsavedChanges,
			languageOptions,
			accentOptions,
			slideshowOptions,
			watchedPathsText,
			setLibraryField,
			setIndexingField,
			setPlayersField,
			setAppearanceField,
			setLanguageField,
			handleWatchedPathsChange,
			handleReset,
			handleSave,
		} = useSettingsScreen();

	const disableActions = isLoading || isSaving;

	return (
		<div className={styles.content}>
			<header className={styles.header}>
				<div className={styles.intro}>
					<h1 className={styles.title}>{t('SETTINGS_PAGE_TITLE')}</h1>
					<p className={styles.description}>{t('SETTINGS_PAGE_DESCRIPTION')}</p>
				</div>
				<div className={styles.summary}>
					<Chip label={`${t('SETTINGS_SUMMARY_ROOT')}: ${settings.library.runtime_root_path || '-'}`} variant='outlined' />
					<Chip
						label={`${t('SETTINGS_SUMMARY_WORKERS')}: ${settings.indexing.workers_enabled ? t('SETTINGS_STATUS_ENABLED') : t('SETTINGS_STATUS_DISABLED')}`}
						variant='outlined'
					/>
					<Chip label={`${t('LANGUAGE')}: ${draft.language.current}`} variant='outlined' />
				</div>
			</header>

			{hasError ? <Alert severity='error'>{t('SETTINGS_LOAD_ERROR')}</Alert> : null}

			<div className={styles.grid}>
					<section className={styles.panel}>
						<div className={styles.panelHeader}>
							<h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_LIBRARY')}</h2>
							<p className={styles.panelDescription}>{t('SETTINGS_SECTION_LIBRARY_DESCRIPTION')}</p>
						</div>
						<div className={styles.summary}>
							{draft.library.watched_paths.map((path) => (
								<Chip key={path} label={path} variant='outlined' />
							))}
						</div>
						<TextField
							fullWidth
							multiline
							minRows={4}
							label={t('SETTINGS_LIBRARY_WATCHED_PATHS')}
							value={watchedPathsText}
							onChange={(event) => handleWatchedPathsChange(event.target.value)}
							disabled={disableActions}
						/>
						<FormControlLabel
							control={
								<Switch
									checked={draft.library.remember_last_location}
									onChange={(_, checked) => setLibraryField('remember_last_location', checked)}
									disabled={disableActions}
								/>
							}
							label={t('SETTINGS_LIBRARY_REMEMBER_LAST_LOCATION')}
						/>
						<FormControlLabel
							control={
								<Switch
									checked={draft.library.prioritize_favorites}
									onChange={(_, checked) => setLibraryField('prioritize_favorites', checked)}
									disabled={disableActions}
								/>
							}
							label={t('SETTINGS_LIBRARY_PRIORITIZE_FAVORITES')}
						/>
						<Alert severity='info'>{t('SETTINGS_LIBRARY_WATCHED_PATHS_HELP')}</Alert>
						<p className={styles.hint}>{t('SETTINGS_LIBRARY_RUNTIME_ROOT', { path: settings.library.runtime_root_path || '-' })}</p>
					</section>

					<section className={styles.panel}>
						<div className={styles.panelHeader}>
							<h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_INDEXING')}</h2>
							<p className={styles.panelDescription}>{t('SETTINGS_SECTION_INDEXING_DESCRIPTION')}</p>
						</div>
						<div className={styles.summary}>
							<Chip label={t('SETTINGS_INDEXING_SCAN_ON_STARTUP')} variant={draft.indexing.scan_on_startup ? 'filled' : 'outlined'} />
							<Chip label={t('SETTINGS_INDEXING_EXTRACT_METADATA')} variant={draft.indexing.extract_metadata ? 'filled' : 'outlined'} />
							<Chip label={t('SETTINGS_INDEXING_GENERATE_PREVIEWS')} variant={draft.indexing.generate_previews ? 'filled' : 'outlined'} />
						</div>
						<FormControlLabel
							control={
								<Switch
									checked={draft.indexing.scan_on_startup}
									onChange={(_, checked) => setIndexingField('scan_on_startup', checked)}
									disabled={disableActions}
								/>
							}
							label={t('SETTINGS_INDEXING_SCAN_ON_STARTUP')}
						/>
						<FormControlLabel
							control={
								<Switch
									checked={draft.indexing.extract_metadata}
									onChange={(_, checked) => setIndexingField('extract_metadata', checked)}
									disabled={disableActions}
								/>
							}
							label={t('SETTINGS_INDEXING_EXTRACT_METADATA')}
						/>
						<FormControlLabel
							control={
								<Switch
									checked={draft.indexing.generate_previews}
									onChange={(_, checked) => setIndexingField('generate_previews', checked)}
									disabled={disableActions}
								/>
							}
							label={t('SETTINGS_INDEXING_GENERATE_PREVIEWS')}
						/>
						<Alert severity={settings.indexing.workers_enabled ? 'success' : 'warning'}>
							{settings.indexing.workers_enabled ? t('SETTINGS_INDEXING_WORKERS_ON') : t('SETTINGS_INDEXING_WORKERS_OFF')}
						</Alert>
				</section>

				<section className={styles.panel}>
					<div className={styles.panelHeader}>
						<h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_PLAYERS')}</h2>
						<p className={styles.panelDescription}>{t('SETTINGS_SECTION_PLAYERS_DESCRIPTION')}</p>
					</div>
					<FormControlLabel
						control={
							<Switch
								checked={draft.players.remember_music_queue}
								onChange={(_, checked) => setPlayersField('remember_music_queue', checked)}
								disabled={disableActions}
							/>
						}
						label={t('SETTINGS_PLAYERS_REMEMBER_MUSIC_QUEUE')}
					/>
					<FormControlLabel
						control={
							<Switch
								checked={draft.players.remember_video_progress}
								onChange={(_, checked) => setPlayersField('remember_video_progress', checked)}
								disabled={disableActions}
							/>
						}
						label={t('SETTINGS_PLAYERS_REMEMBER_VIDEO_PROGRESS')}
					/>
					<FormControlLabel
						control={
							<Switch
								checked={draft.players.autoplay_next_video}
								onChange={(_, checked) => setPlayersField('autoplay_next_video', checked)}
								disabled={disableActions}
							/>
						}
						label={t('SETTINGS_PLAYERS_AUTOPLAY_NEXT_VIDEO')}
					/>
					<FormControl fullWidth>
						<InputLabel id='settings-slideshow-label'>{t('SETTINGS_PLAYERS_SLIDESHOW_INTERVAL')}</InputLabel>
						<Select
							labelId='settings-slideshow-label'
							value={String(draft.players.image_slideshow_seconds)}
							label={t('SETTINGS_PLAYERS_SLIDESHOW_INTERVAL')}
							onChange={(event) => setPlayersField('image_slideshow_seconds', Number(event.target.value))}
							disabled={disableActions}
						>
							{slideshowOptions.map((option) => (
								<MenuItem key={option.value} value={String(option.value)}>
									{option.label}
								</MenuItem>
							))}
						</Select>
					</FormControl>
				</section>

				<section className={styles.panel}>
					<div className={styles.panelHeader}>
						<h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_APPEARANCE')}</h2>
						<p className={styles.panelDescription}>{t('SETTINGS_SECTION_APPEARANCE_DESCRIPTION')}</p>
					</div>
					<FormControl fullWidth>
						<InputLabel id='settings-accent-label'>{t('SETTINGS_APPEARANCE_ACCENT')}</InputLabel>
						<Select
							labelId='settings-accent-label'
							value={draft.appearance.accent_color}
							label={t('SETTINGS_APPEARANCE_ACCENT')}
							onChange={(event) => setAppearanceField('accent_color', event.target.value as 'violet' | 'cyan' | 'rose')}
							disabled={disableActions}
						>
							{accentOptions.map((option) => (
								<MenuItem key={option.value} value={option.value}>
									{option.label}
								</MenuItem>
							))}
						</Select>
					</FormControl>
					<FormControlLabel
						control={
							<Switch
								checked={draft.appearance.reduce_motion}
								onChange={(_, checked) => setAppearanceField('reduce_motion', checked)}
								disabled={disableActions}
							/>
						}
						label={t('SETTINGS_APPEARANCE_REDUCE_MOTION')}
					/>
					<div className={styles.swatches}>
						{accentOptions.map((option) => (
							<div key={option.value} className={styles.swatchRow}>
								<span className={`${styles.swatch} ${styles[`swatch${option.value}`]}`} aria-hidden='true' />
								<span>{option.label}</span>
							</div>
						))}
					</div>
				</section>

				<section className={styles.panel}>
					<div className={styles.panelHeader}>
						<h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_LANGUAGE')}</h2>
						<p className={styles.panelDescription}>{t('SETTINGS_SECTION_LANGUAGE_DESCRIPTION')}</p>
					</div>
					<FormControl fullWidth>
						<InputLabel id='settings-language-label'>{t('LANGUAGE')}</InputLabel>
						<Select
							labelId='settings-language-label'
							value={draft.language.current}
							label={t('LANGUAGE')}
							onChange={(event) => setLanguageField(event.target.value)}
							disabled={disableActions}
						>
							{languageOptions.map((option) => (
								<MenuItem key={option.value} value={option.value}>
									{option.label}
								</MenuItem>
							))}
						</Select>
					</FormControl>
					<p className={styles.hint}>{t('SETTINGS_LANGUAGE_HELP')}</p>
				</section>
			</div>

			<footer className={styles.footer}>
				<div className={styles.links}>
					<Button component={Link} to={appRoutes.analytics} variant='outlined'>
						{t('ANALYTICS')}
					</Button>
					<Button component={Link} to={appRoutes.about} variant='text'>
						{t('ABOUT')}
					</Button>
				</div>
				<div className={styles.actions}>
					<Button variant='text' onClick={handleReset} disabled={disableActions || !hasUnsavedChanges}>
						{t('SETTINGS_RESET')}
					</Button>
					<Button variant='contained' onClick={() => void handleSave()} disabled={disableActions || !hasUnsavedChanges}>
						{isSaving ? t('SAVING') : t('SETTINGS_SAVE')}
					</Button>
				</div>
			</footer>
		</div>
	);
};

export default SettingsScreen;
