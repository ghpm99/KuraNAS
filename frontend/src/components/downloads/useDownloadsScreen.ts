import useI18n from '@/components/i18n/provider/i18nContext';
import { buildDownloadHref, getDownloads } from '@/service/downloads';
import { formatSize } from '@/shared/utils/formatSize';
import type { DownloadItem } from '@/types/downloads';
import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';

// Browser extensions can't be installed by a plain double-click like an APK, so
// the screen shows manual "load unpacked" steps for them.
export const isBrowserExtension = (item: DownloadItem): boolean => item.platform === 'browser';

export const useDownloadsScreen = () => {
    const { t } = useI18n();

    const { data, isLoading, isError } = useQuery<DownloadItem[]>({
        queryKey: ['downloads'],
        queryFn: getDownloads,
        refetchOnWindowFocus: false,
    });

    const items = useMemo(() => data ?? [], [data]);
    const hasBrowserExtension = useMemo(() => items.some(isBrowserExtension), [items]);

    return {
        t,
        items,
        isLoading,
        isError,
        isEmpty: !isLoading && !isError && items.length === 0,
        hasBrowserExtension,
        formatSize,
        buildDownloadHref,
        isBrowserExtension,
    };
};
