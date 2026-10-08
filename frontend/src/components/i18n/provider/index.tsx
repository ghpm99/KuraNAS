import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getTranslations } from '@/service/configuration';
import { I18nContextProvider, I18nContextType } from './i18nContext';
import { fallbackCatalog } from './fallbackCatalog';
import { getTranslationRetryDelay, translationRetryCount } from './translationRetry';

const I18nProvider = ({ children }: { children: React.ReactNode }) => {
    const { data } = useQuery({
        queryKey: ['configuration'],
        queryFn: getTranslations,
        retry: translationRetryCount,
        retryDelay: getTranslationRetryDelay,
    });

    const catalog = useMemo(() => ({ ...fallbackCatalog, ...data }), [data]);

    const t = (key: string, options?: Record<string, string>): string => {
        const translation = catalog[key];
        if (!translation) return key;

        return Object.entries(options || {}).reduce(
            (translated, [placeholder, replacement]) =>
                translated.replace(`{{${placeholder}}}`, replacement),
            translation
        );
    };

    const contextValue: I18nContextType = {
        t,
    };
    return <I18nContextProvider value={contextValue}>{children}</I18nContextProvider>;
};

export default I18nProvider;
