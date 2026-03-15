import useI18n from '@/components/i18n/provider/i18nContext';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import type { SettingsConfiguration, UpdateSettingsConfigurationRequest } from '@/service/configuration';
import { useSnackbar } from 'notistack';
import { useCallback, useEffect, useMemo, useState } from 'react';

const accentOptionValues = ['violet', 'cyan', 'rose'] as const;
const slideshowOptionValues = [4, 8, 12, 20] as const;

const buildDraftFromSettings = (settings: SettingsConfiguration): UpdateSettingsConfigurationRequest => ({
	library: {
		watched_paths: settings.library.watched_paths,
		remember_last_location: settings.library.remember_last_location,
		prioritize_favorites: settings.library.prioritize_favorites,
	},
	indexing: {
		scan_on_startup: settings.indexing.scan_on_startup,
		extract_metadata: settings.indexing.extract_metadata,
		generate_previews: settings.indexing.generate_previews,
	},
	players: {
		remember_music_queue: settings.players.remember_music_queue,
		remember_video_progress: settings.players.remember_video_progress,
		autoplay_next_video: settings.players.autoplay_next_video,
		image_slideshow_seconds: settings.players.image_slideshow_seconds,
	},
	appearance: {
		accent_color: settings.appearance.accent_color,
		reduce_motion: settings.appearance.reduce_motion,
	},
	language: {
		current: settings.language.current,
	},
});

const serializeWatchedPaths = (paths: string[]) => paths.join('\n');

const parseWatchedPaths = (value: string) =>
	Array.from(
		new Set(
			value
				.split('\n')
				.map((path) => path.trim())
				.filter(Boolean),
		),
	);

const areSettingsEqual = (left: UpdateSettingsConfigurationRequest, right: UpdateSettingsConfigurationRequest) =>
	JSON.stringify(left) === JSON.stringify(right);

const useSettingsScreen = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const { settings, isLoading, isSaving, hasError, saveSettings } = useSettings();
	const settingsSnapshot = useMemo(() => JSON.stringify(buildDraftFromSettings(settings)), [settings]);
	const baselineDraft = useMemo(() => JSON.parse(settingsSnapshot) as UpdateSettingsConfigurationRequest, [settingsSnapshot]);
	const [draft, setDraft] = useState<UpdateSettingsConfigurationRequest>(() => baselineDraft);

	useEffect(() => {
		setDraft(baselineDraft);
	}, [baselineDraft]);

	const watchedPathsText = useMemo(() => serializeWatchedPaths(draft.library.watched_paths), [draft.library.watched_paths]);
	const hasUnsavedChanges = useMemo(() => !areSettingsEqual(draft, baselineDraft), [baselineDraft, draft]);

	const languageOptions = useMemo(
		() =>
			settings.language.available.map((locale) => ({
				value: locale,
				label: t(`SETTINGS_LANGUAGE_OPTION_${locale}`),
			})),
		[settings.language.available, t],
	);

	const accentOptions = useMemo(
		() =>
			accentOptionValues.map((value) => ({
				value,
				label: t(`SETTINGS_APPEARANCE_ACCENT_${value.toUpperCase()}`),
			})),
		[t],
	);

	const slideshowOptions = useMemo(
		() =>
			slideshowOptionValues.map((value) => ({
				value,
				label: t('SETTINGS_PLAYERS_SLIDESHOW_OPTION', { seconds: String(value) }),
			})),
		[t],
	);

	const setLibraryField = useCallback(
		<Key extends keyof UpdateSettingsConfigurationRequest['library']>(field: Key, value: UpdateSettingsConfigurationRequest['library'][Key]) => {
			setDraft((current) => ({
				...current,
				library: {
					...current.library,
					[field]: value,
				},
			}));
		},
		[],
	);

	const setIndexingField = useCallback(
		<Key extends keyof UpdateSettingsConfigurationRequest['indexing']>(field: Key, value: UpdateSettingsConfigurationRequest['indexing'][Key]) => {
			setDraft((current) => ({
				...current,
				indexing: {
					...current.indexing,
					[field]: value,
				},
			}));
		},
		[],
	);

	const setPlayersField = useCallback(
		<Key extends keyof UpdateSettingsConfigurationRequest['players']>(field: Key, value: UpdateSettingsConfigurationRequest['players'][Key]) => {
			setDraft((current) => ({
				...current,
				players: {
					...current.players,
					[field]: value,
				},
			}));
		},
		[],
	);

	const setAppearanceField = useCallback(
		<Key extends keyof UpdateSettingsConfigurationRequest['appearance']>(field: Key, value: UpdateSettingsConfigurationRequest['appearance'][Key]) => {
			setDraft((current) => ({
				...current,
				appearance: {
					...current.appearance,
					[field]: value,
				},
			}));
		},
		[],
	);

	const setLanguageField = useCallback((value: string) => {
		setDraft((current) => ({
			...current,
			language: {
				current: value,
			},
		}));
	}, []);

	const handleWatchedPathsChange = useCallback(
		(value: string) => {
			setLibraryField('watched_paths', parseWatchedPaths(value));
		},
		[setLibraryField],
	);

	const handleReset = useCallback(() => {
		setDraft(baselineDraft);
	}, [baselineDraft]);

	const handleSave = useCallback(async () => {
		try {
			await saveSettings(draft);
			enqueueSnackbar(t('SETTINGS_SAVE_SUCCESS'), { variant: 'success' });
		} catch {
			enqueueSnackbar(t('SETTINGS_SAVE_ERROR'), { variant: 'error' });
		}
	}, [draft, enqueueSnackbar, saveSettings, t]);

	return {
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
	};
};

export default useSettingsScreen;
