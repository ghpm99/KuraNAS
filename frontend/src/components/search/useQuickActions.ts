import { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { requestShortcutsHelp, requestSidebarToggle } from '@/components/layout/appCommandEvents';
import {
    themeModeLabelKeys,
    useThemeModeCycle,
} from '@/components/providers/colorSchemeProvider/useThemeModeCycle';
import { quickActionDestinations } from './quickActionDestinations';
import type { SearchDialogItem } from './useGlobalSearchProvider';

export const useQuickActions = (): SearchDialogItem[] => {
    const { t } = useI18n();
    const navigate = useNavigate();
    const { nextThemeMode, canCycleThemeMode, cycleThemeMode } = useThemeModeCycle();

    return useMemo(() => {
        const destinationActions = quickActionDestinations.map<SearchDialogItem>((destination) => ({
            id: `action-destination-${destination.route}`,
            kind: 'action',
            label: destination.parentLabelKey
                ? t('GLOBAL_SEARCH_ACTION_SECTION_LABEL', {
                      domain: t(destination.parentLabelKey),
                      section: t(destination.labelKey),
                  })
                : t(destination.labelKey),
            description: t(destination.descriptionKey),
            onSelect: () => navigate(destination.route),
        }));

        const commandActions: SearchDialogItem[] = [
            {
                id: 'action-toggle-sidebar',
                kind: 'action',
                label: t('GLOBAL_SEARCH_ACTION_TOGGLE_SIDEBAR'),
                description: t('SHORTCUT_TOGGLE_SIDEBAR'),
                onSelect: requestSidebarToggle,
            },
            {
                id: 'action-show-shortcuts',
                kind: 'action',
                label: t('SHORTCUTS_DIALOG_TITLE'),
                description: t('SHORTCUT_SHOW_HELP'),
                onSelect: requestShortcutsHelp,
            },
        ];

        if (canCycleThemeMode) {
            commandActions.unshift({
                id: 'action-toggle-theme',
                kind: 'action',
                label: t('GLOBAL_SEARCH_ACTION_TOGGLE_THEME'),
                description: t('GLOBAL_SEARCH_ACTION_TOGGLE_THEME_DESCRIPTION', {
                    theme: t(themeModeLabelKeys[nextThemeMode]),
                }),
                onSelect: () => void cycleThemeMode(),
            });
        }

        return [...destinationActions, ...commandActions];
    }, [canCycleThemeMode, cycleThemeMode, navigate, nextThemeMode, t]);
};
