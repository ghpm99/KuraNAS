import { useEffect } from 'react';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';

const fallbackLanguageTag = 'en-US';

export const toLanguageTag = (language: string | undefined): string => {
    const normalizedLanguage = language?.trim().replace(/_/g, '-');
    return normalizedLanguage || fallbackLanguageTag;
};

export const useDocumentLanguage = () => {
    const { settings } = useSettings();
    const languageTag = toLanguageTag(settings.language?.current);

    useEffect(() => {
        document.documentElement.lang = languageTag;
    }, [languageTag]);
};
