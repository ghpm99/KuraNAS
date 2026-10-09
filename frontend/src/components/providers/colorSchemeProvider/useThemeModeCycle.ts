import { useCallback } from 'react';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useOptionalSettings } from '@/components/providers/settingsProvider/settingsContext';
import { buildDraftFromSettings } from '@/components/settings/buildSettingsDraft';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { ThemeMode } from './colorScheme';
import { useColorScheme } from './colorSchemeContext';

const themeModeCycle: readonly ThemeMode[] = ['dark', 'light', 'system'];

export const themeModeLabelKeys: Record<ThemeMode, string> = {
    dark: 'SETTINGS_APPEARANCE_THEME_DARK',
    light: 'SETTINGS_APPEARANCE_THEME_LIGHT',
    system: 'SETTINGS_APPEARANCE_THEME_SYSTEM',
};

export const getNextThemeMode = (currentThemeMode: ThemeMode): ThemeMode => {
    const currentIndex = themeModeCycle.indexOf(currentThemeMode);
    return themeModeCycle[(currentIndex + 1) % themeModeCycle.length]!;
};

export const useThemeModeCycle = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const settingsContext = useOptionalSettings();
    const { themeMode } = useColorScheme();
    const nextThemeMode = getNextThemeMode(themeMode);
    const canCycleThemeMode = settingsContext !== undefined;

    const cycleThemeMode = useCallback(async () => {
        if (!settingsContext) {
            return;
        }
        const currentDraft = buildDraftFromSettings(settingsContext.settings);
        try {
            await settingsContext.saveSettings({
                ...currentDraft,
                appearance: { ...currentDraft.appearance, theme_mode: nextThemeMode },
            });
        } catch (error) {
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('SETTINGS_SAVE_ERROR'), {
                variant: 'error',
            });
        }
    }, [enqueueSnackbar, nextThemeMode, settingsContext, t]);

    return { nextThemeMode, canCycleThemeMode, cycleThemeMode };
};
