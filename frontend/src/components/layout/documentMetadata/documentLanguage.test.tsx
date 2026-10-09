import { render } from '@testing-library/react';
import {
    SettingsContextProvider,
    defaultSettingsConfiguration,
    type SettingsContextType,
} from '@/components/providers/settingsProvider/settingsContext';
import type { SettingsConfiguration } from '@/service/configuration';
import DocumentLanguageSync from './DocumentLanguageSync';
import { toLanguageTag } from './documentLanguage';

const renderWithSettings = (settings: Partial<SettingsConfiguration>) => {
    const contextValue: SettingsContextType = {
        settings: { ...defaultSettingsConfiguration, ...settings } as SettingsConfiguration,
        isLoading: false,
        isSaving: false,
        hasError: false,
        refresh: jest.fn(),
        saveSettings: jest.fn(),
    };
    return render(
        <SettingsContextProvider value={contextValue}>
            <DocumentLanguageSync />
        </SettingsContextProvider>
    );
};

describe('DocumentLanguageSync', () => {
    it('falls back to en-US when the settings carry no language group', () => {
        renderWithSettings({ language: undefined });

        expect(document.documentElement.lang).toBe('en-US');
    });

    it('follows the active language from the settings', () => {
        renderWithSettings({ language: { current: 'pt-BR', available: ['pt-BR', 'en-US'] } });

        expect(document.documentElement.lang).toBe('pt-BR');
    });

    it('normalizes underscores into BCP-47 tags', () => {
        expect(toLanguageTag('pt_BR')).toBe('pt-BR');
        expect(toLanguageTag('  ')).toBe('en-US');
        expect(toLanguageTag(undefined)).toBe('en-US');
    });
});
