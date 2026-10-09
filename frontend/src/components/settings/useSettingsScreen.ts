import useI18n from '@/components/i18n/provider/i18nContext';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import type { UpdateSettingsConfigurationRequest } from '@/service/configuration';
import { useSnackbar } from 'notistack';
import { buildDraftFromSettings } from './buildSettingsDraft';
import { useCallback, useEffect, useMemo, useState } from 'react';

const accentOptionValues = ['violet', 'cyan', 'rose'] as const;
const themeModeOptionValues = ['dark', 'light', 'system'] as const;
const slideshowOptionValues = [4, 8, 12, 20] as const;

const areSettingsEqual = (
    left: UpdateSettingsConfigurationRequest,
    right: UpdateSettingsConfigurationRequest
) => JSON.stringify(left) === JSON.stringify(right);

// The backend returns validation failures (e.g. a captures path inside a storage
// root) as an already-translated { error } string. Surface it verbatim instead
// of the generic save-error toast; the i18n rule forbids re-translating it.
const extractBackendError = (error: unknown): string | null => {
    if (error && typeof error === 'object' && 'response' in error) {
        const response = (error as { response?: { data?: { error?: unknown } } }).response;
        const message = response?.data?.error;
        if (typeof message === 'string' && message.trim()) {
            return message;
        }
    }
    return null;
};

const useSettingsScreen = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const { settings, isLoading, isSaving, hasError, saveSettings } = useSettings();
    const settingsSnapshot = useMemo(
        () => JSON.stringify(buildDraftFromSettings(settings)),
        [settings]
    );
    const baselineDraft = useMemo(
        () => JSON.parse(settingsSnapshot) as UpdateSettingsConfigurationRequest,
        [settingsSnapshot]
    );
    const [draft, setDraft] = useState<UpdateSettingsConfigurationRequest>(() => baselineDraft);

    useEffect(() => {
        setDraft(baselineDraft);
    }, [baselineDraft]);

    const hasUnsavedChanges = useMemo(
        () => !areSettingsEqual(draft, baselineDraft),
        [baselineDraft, draft]
    );

    const languageOptions = useMemo(
        () =>
            settings.language.available.map((locale) => ({
                value: locale,
                label: t(`SETTINGS_LANGUAGE_OPTION_${locale}`),
            })),
        [settings.language.available, t]
    );

    const accentOptions = useMemo(
        () =>
            accentOptionValues.map((value) => ({
                value,
                label: t(`SETTINGS_APPEARANCE_ACCENT_${value.toUpperCase()}`),
            })),
        [t]
    );

    const themeModeOptions = useMemo(
        () =>
            themeModeOptionValues.map((value) => ({
                value,
                label: t(`SETTINGS_APPEARANCE_THEME_${value.toUpperCase()}`),
            })),
        [t]
    );

    const slideshowOptions = useMemo(
        () =>
            slideshowOptionValues.map((value) => ({
                value,
                label: t('SETTINGS_PLAYERS_SLIDESHOW_OPTION', {
                    seconds: String(value),
                }),
            })),
        [t]
    );

    const setIndexingField = useCallback(
        <Key extends keyof UpdateSettingsConfigurationRequest['indexing']>(
            field: Key,
            value: UpdateSettingsConfigurationRequest['indexing'][Key]
        ) => {
            setDraft((current) => ({
                ...current,
                indexing: {
                    ...current.indexing,
                    [field]: value,
                },
            }));
        },
        []
    );

    const setAIField = useCallback(
        <Key extends keyof UpdateSettingsConfigurationRequest['ai']>(
            field: Key,
            value: UpdateSettingsConfigurationRequest['ai'][Key]
        ) => {
            setDraft((current) => ({
                ...current,
                ai: {
                    ...current.ai,
                    [field]: value,
                },
            }));
        },
        []
    );

    const setPlayersField = useCallback(
        <Key extends keyof UpdateSettingsConfigurationRequest['players']>(
            field: Key,
            value: UpdateSettingsConfigurationRequest['players'][Key]
        ) => {
            setDraft((current) => ({
                ...current,
                players: {
                    ...current.players,
                    [field]: value,
                },
            }));
        },
        []
    );

    const setAppearanceField = useCallback(
        <Key extends keyof UpdateSettingsConfigurationRequest['appearance']>(
            field: Key,
            value: UpdateSettingsConfigurationRequest['appearance'][Key]
        ) => {
            setDraft((current) => ({
                ...current,
                appearance: {
                    ...current.appearance,
                    [field]: value,
                },
            }));
        },
        []
    );

    const setCapturesField = useCallback(
        <Key extends keyof UpdateSettingsConfigurationRequest['captures']>(
            field: Key,
            value: UpdateSettingsConfigurationRequest['captures'][Key]
        ) => {
            setDraft((current) => ({
                ...current,
                captures: {
                    ...current.captures,
                    [field]: value,
                },
            }));
        },
        []
    );

    const setLanguageField = useCallback((value: string) => {
        setDraft((current) => ({
            ...current,
            language: {
                current: value,
            },
        }));
    }, []);

    const handleReset = useCallback(() => {
        setDraft(baselineDraft);
    }, [baselineDraft]);

    const handleSave = useCallback(async () => {
        try {
            await saveSettings(draft);
            enqueueSnackbar(t('SETTINGS_SAVE_SUCCESS'), { variant: 'success' });
        } catch (error) {
            const backendMessage = extractBackendError(error);
            enqueueSnackbar(backendMessage ?? t('SETTINGS_SAVE_ERROR'), { variant: 'error' });
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
    };
};

export default useSettingsScreen;
