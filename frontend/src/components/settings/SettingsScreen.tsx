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
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import AccessControlSettingsSection from './AccessControlSettingsSection';
import AIProvidersSettingsSection from './AIProvidersSettingsSection';
import ImageClassificationBackfill from './ImageClassificationBackfill';
import AutoShutdownSettingsSection from './AutoShutdownSettingsSection';
import BackupSettingsSection from './BackupSettingsSection';
import EmailSettingsSection from './EmailSettingsSection';
import LibrarySettingsSection from './LibrarySettingsSection';
import StorageRootsSettingsSection from './StorageRootsSettingsSection';
import TieringSettingsSection from './TieringSettingsSection';
import YtDlpSettingsSection from './YtDlpSettingsSection';
import SettingsSectionAnchor from './SettingsSectionAnchor';
import useSettingsScreen from './useSettingsScreen';
import { useSettingsHashScroll } from './useSettingsHashScroll';
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
        themeModeOptions,
        slideshowOptions,
        setIndexingField,
        setCapturesField,
        setAIField,
        setPlayersField,
        setAppearanceField,
        setLanguageField,
        handleReset,
        handleSave,
    } = useSettingsScreen();

    useSettingsHashScroll();

    const disableActions = isLoading || isSaving;

    return (
        <PageContainer>
            <PageHeader
                title={t('SETTINGS_PAGE_TITLE')}
                subtitle={t('SETTINGS_PAGE_DESCRIPTION')}
                actions={
                    <>
                        <Chip
                            label={`${t('SETTINGS_SUMMARY_WORKERS')}: ${settings.indexing.workers_enabled ? t('SETTINGS_STATUS_ENABLED') : t('SETTINGS_STATUS_DISABLED')}`}
                            variant="outlined"
                        />
                        <Chip
                            label={`${t('LANGUAGE')}: ${draft.language.current}`}
                            variant="outlined"
                        />
                    </>
                }
            />

            {hasError ? <Alert severity="error">{t('SETTINGS_LOAD_ERROR')}</Alert> : null}

            <div className={styles.grid}>
                <SettingsSectionAnchor sectionId="storage-roots">
                    <StorageRootsSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="backup">
                    <BackupSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="tiering">
                    <TieringSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="auto-shutdown">
                    <AutoShutdownSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="access-control">
                    <AccessControlSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="libraries">
                    <LibrarySettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="ai-providers">
                    <AIProvidersSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="yt-dlp">
                    <YtDlpSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="email">
                    <EmailSettingsSection className={styles.panel} />
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="indexing">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_INDEXING')}</h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_INDEXING_DESCRIPTION')}
                            </p>
                        </div>
                        <div className={styles.summary}>
                            <Chip
                                label={t('SETTINGS_INDEXING_SCAN_ON_STARTUP')}
                                variant={draft.indexing.scan_on_startup ? 'filled' : 'outlined'}
                            />
                            <Chip
                                label={t('SETTINGS_INDEXING_EXTRACT_METADATA')}
                                variant={draft.indexing.extract_metadata ? 'filled' : 'outlined'}
                            />
                            <Chip
                                label={t('SETTINGS_INDEXING_GENERATE_PREVIEWS')}
                                variant={draft.indexing.generate_previews ? 'filled' : 'outlined'}
                            />
                        </div>
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.indexing.scan_on_startup}
                                    onChange={(_, checked) =>
                                        setIndexingField('scan_on_startup', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_INDEXING_SCAN_ON_STARTUP')}
                        />
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.indexing.extract_metadata}
                                    onChange={(_, checked) =>
                                        setIndexingField('extract_metadata', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_INDEXING_EXTRACT_METADATA')}
                        />
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.indexing.generate_previews}
                                    onChange={(_, checked) =>
                                        setIndexingField('generate_previews', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_INDEXING_GENERATE_PREVIEWS')}
                        />
                        <Alert severity={settings.indexing.workers_enabled ? 'success' : 'warning'}>
                            {settings.indexing.workers_enabled
                                ? t('SETTINGS_INDEXING_WORKERS_ON')
                                : t('SETTINGS_INDEXING_WORKERS_OFF')}
                        </Alert>
                    </section>
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="captures">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_CAPTURES')}</h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_CAPTURES_DESCRIPTION')}
                            </p>
                        </div>
                        <TextField
                            fullWidth
                            label={t('SETTINGS_CAPTURES_SAVE_PATH')}
                            value={draft.captures?.save_path ?? ''}
                            onChange={(event) => setCapturesField('save_path', event.target.value)}
                            placeholder={settings.captures?.default_path ?? ''}
                            disabled={disableActions}
                        />
                        <p className={styles.hint}>
                            {t('SETTINGS_CAPTURES_DEFAULT_PATH', {
                                path: settings.captures?.default_path || '-',
                            })}
                        </p>
                        {(settings.captures?.storage_roots ?? []).length > 0 ? (
                            <div className={styles.summary}>
                                {settings.captures.storage_roots.map((root) => (
                                    <Chip key={root} label={root} variant="outlined" />
                                ))}
                            </div>
                        ) : null}
                        <Alert severity="warning">
                            {t('SETTINGS_CAPTURES_OUTSIDE_ROOTS_HELP')}
                        </Alert>
                    </section>
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="ai">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_AI')}</h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_AI_DESCRIPTION')}
                            </p>
                        </div>
                        <div className={styles.summary}>
                            <Chip
                                label={t('SETTINGS_AI_IMAGE_CLASSIFICATION')}
                                variant={draft.ai.image_classification ? 'filled' : 'outlined'}
                            />
                        </div>
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.ai.image_classification}
                                    onChange={(_, checked) =>
                                        setAIField('image_classification', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_AI_IMAGE_CLASSIFICATION')}
                        />
                        <Alert severity="info">{t('SETTINGS_AI_IMAGE_CLASSIFICATION_HELP')}</Alert>
                        <ImageClassificationBackfill
                            disabled={!draft.ai.image_classification || disableActions}
                        />
                    </section>
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="players">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_PLAYERS')}</h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_PLAYERS_DESCRIPTION')}
                            </p>
                        </div>
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.players.remember_music_queue}
                                    onChange={(_, checked) =>
                                        setPlayersField('remember_music_queue', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_PLAYERS_REMEMBER_MUSIC_QUEUE')}
                        />
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.players.remember_video_progress}
                                    onChange={(_, checked) =>
                                        setPlayersField('remember_video_progress', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_PLAYERS_REMEMBER_VIDEO_PROGRESS')}
                        />
                        <FormControlLabel
                            control={
                                <Switch
                                    checked={draft.players.autoplay_next_video}
                                    onChange={(_, checked) =>
                                        setPlayersField('autoplay_next_video', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_PLAYERS_AUTOPLAY_NEXT_VIDEO')}
                        />
                        <FormControl fullWidth>
                            <InputLabel id="settings-slideshow-label">
                                {t('SETTINGS_PLAYERS_SLIDESHOW_INTERVAL')}
                            </InputLabel>
                            <Select
                                labelId="settings-slideshow-label"
                                value={String(draft.players.image_slideshow_seconds)}
                                label={t('SETTINGS_PLAYERS_SLIDESHOW_INTERVAL')}
                                onChange={(event) =>
                                    setPlayersField(
                                        'image_slideshow_seconds',
                                        Number(event.target.value)
                                    )
                                }
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
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="appearance">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>
                                {t('SETTINGS_SECTION_APPEARANCE')}
                            </h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_APPEARANCE_DESCRIPTION')}
                            </p>
                        </div>
                        <FormControl fullWidth>
                            <InputLabel id="settings-accent-label">
                                {t('SETTINGS_APPEARANCE_ACCENT')}
                            </InputLabel>
                            <Select
                                labelId="settings-accent-label"
                                value={draft.appearance.accent_color}
                                label={t('SETTINGS_APPEARANCE_ACCENT')}
                                onChange={(event) =>
                                    setAppearanceField(
                                        'accent_color',
                                        event.target.value as 'violet' | 'cyan' | 'rose'
                                    )
                                }
                                disabled={disableActions}
                            >
                                {accentOptions.map((option) => (
                                    <MenuItem key={option.value} value={option.value}>
                                        {option.label}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                        <FormControl fullWidth>
                            <InputLabel id="settings-theme-label">
                                {t('SETTINGS_APPEARANCE_THEME')}
                            </InputLabel>
                            <Select
                                labelId="settings-theme-label"
                                value={draft.appearance.theme_mode}
                                label={t('SETTINGS_APPEARANCE_THEME')}
                                onChange={(event) =>
                                    setAppearanceField(
                                        'theme_mode',
                                        event.target.value as 'dark' | 'light' | 'system'
                                    )
                                }
                                disabled={disableActions}
                            >
                                {themeModeOptions.map((option) => (
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
                                    onChange={(_, checked) =>
                                        setAppearanceField('reduce_motion', checked)
                                    }
                                    disabled={disableActions}
                                />
                            }
                            label={t('SETTINGS_APPEARANCE_REDUCE_MOTION')}
                        />
                        <div className={styles.swatches}>
                            {accentOptions.map((option) => (
                                <div key={option.value} className={styles.swatchRow}>
                                    <span
                                        className={`${styles.swatch} ${styles[`swatch${option.value}`]}`}
                                        aria-hidden="true"
                                    />
                                    <span>{option.label}</span>
                                </div>
                            ))}
                        </div>
                    </section>
                </SettingsSectionAnchor>

                <SettingsSectionAnchor sectionId="language">
                    <section className={styles.panel}>
                        <div className={styles.panelHeader}>
                            <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_LANGUAGE')}</h2>
                            <p className={styles.panelDescription}>
                                {t('SETTINGS_SECTION_LANGUAGE_DESCRIPTION')}
                            </p>
                        </div>
                        <FormControl fullWidth>
                            <InputLabel id="settings-language-label">{t('LANGUAGE')}</InputLabel>
                            <Select
                                labelId="settings-language-label"
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
                </SettingsSectionAnchor>
            </div>

            <footer className={styles.footer}>
                <div className={styles.links}>
                    <Button component={Link} to={appRoutes.analytics} variant="outlined">
                        {t('ANALYTICS')}
                    </Button>
                    <Button component={Link} to={appRoutes.about} variant="text">
                        {t('ABOUT')}
                    </Button>
                </div>
                <div className={styles.actions}>
                    <Button
                        variant="text"
                        onClick={handleReset}
                        disabled={disableActions || !hasUnsavedChanges}
                    >
                        {t('SETTINGS_RESET')}
                    </Button>
                    <Button
                        variant="contained"
                        onClick={() => void handleSave()}
                        disabled={disableActions || !hasUnsavedChanges}
                    >
                        {isSaving ? t('SAVING') : t('SETTINGS_SAVE')}
                    </Button>
                </div>
            </footer>
        </PageContainer>
    );
};

export default SettingsScreen;
