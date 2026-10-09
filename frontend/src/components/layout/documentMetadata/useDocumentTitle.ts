import { useEffect } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';

export const useDocumentTitle = (pageTitle: string) => {
    const { t } = useI18n();
    const appName = t('APP_NAME');

    useEffect(() => {
        const previousTitle = document.title;
        document.title = pageTitle ? `${pageTitle} · ${appName}` : appName;
        return () => {
            document.title = previousTitle;
        };
    }, [pageTitle, appName]);
};
