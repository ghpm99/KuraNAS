import useI18n from '@/components/i18n/provider/i18nContext';
import useFile from '@/features/files/providers/fileProvider/fileContext';
import { FileType } from '@/utils';
import { useQuery } from '@tanstack/react-query';
import { getStarredCounts, type StarredCounts, type StarredSort } from '@/service/files';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { starredCountsQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { useMemo, useState } from 'react';
import type { BreadcrumbSegment } from '@/features/files/files/useFilesExplorerScreen';
import { findTrailById } from '@/features/files/files/fileNavigation';

export type FavoritesFilter = 'all' | 'folders' | 'files' | 'media';
export type FavoritesViewMode = 'grid' | 'list';

export type FavoritesSortOption = {
    value: StarredSort;
    label: string;
};

type FavoritesFilterOption = {
    value: FavoritesFilter;
    label: string;
    count: number | null;
};

const favoritesCountKeyByFilter: Record<FavoritesFilter, keyof StarredCounts> = {
    all: 'total',
    folders: 'folders',
    files: 'files',
    media: 'media',
};

const useFavoritesScreen = () => {
    const { t } = useI18n();
    const {
        files,
        selectedItem,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        handleSelectItem,
        starredKind,
        setStarredKind,
        starredSort,
        setStarredSort,
    } = useFile();
    const [viewMode, setViewMode] = useState<FavoritesViewMode>('grid');

    const favoritesTrail = useMemo(() => {
        if (!selectedItem) {
            return [];
        }

        const fullTrail = findTrailById(files, selectedItem.id) ?? [];
        const firstStarredIndex = fullTrail.findIndex((trailItem) => trailItem.starred);
        if (firstStarredIndex < 0) {
            return [selectedItem];
        }

        return fullTrail.slice(firstStarredIndex);
    }, [files, selectedItem]);

    const parentFolder = favoritesTrail[favoritesTrail.length - 2] ?? null;

    const scopedItems = useMemo(() => {
        if (!selectedItem) {
            return files;
        }

        if (selectedItem.type === FileType.Directory) {
            return selectedItem.file_children ?? [];
        }

        return parentFolder ? (parentFolder.file_children ?? []) : files;
    }, [files, selectedItem, parentFolder]);

    const isRootListing =
        !selectedItem || (selectedItem.type === FileType.File && parentFolder === null);

    const { data: starredCounts } = useQuery({
        queryKey: starredCountsQueryKey,
        queryFn: getStarredCounts,
        enabled: isRootListing,
        staleTime: listingStaleTimeMs,
    });

    const activeFilter: FavoritesFilter = isRootListing ? starredKind : 'all';

    const filterOptions = useMemo<FavoritesFilterOption[]>(
        () =>
            (
                [
                    ['all', 'FAVORITES_FILTER_ALL'],
                    ['folders', 'FAVORITES_FILTER_FOLDERS'],
                    ['files', 'FAVORITES_FILTER_FILES'],
                    ['media', 'FAVORITES_FILTER_MEDIA'],
                ] as const
            ).map(([value, labelKey]) => ({
                value,
                label: t(labelKey),
                count: starredCounts?.[favoritesCountKeyByFilter[value]] ?? null,
            })),
        [starredCounts, t]
    );

    const sortOptions = useMemo<FavoritesSortOption[]>(
        () => [
            { value: 'name', label: t('FAVORITES_SORT_NAME') },
            { value: 'starred_at', label: t('FAVORITES_SORT_STARRED_AT') },
        ],
        [t]
    );

    const breadcrumbSegments = useMemo<BreadcrumbSegment[]>(() => {
        const rootSegment: BreadcrumbSegment = {
            id: null,
            label: t('STARRED_FILES'),
            path: null,
            isCurrent: !selectedItem,
        };

        if (!selectedItem) {
            return [rootSegment];
        }

        return [
            rootSegment,
            ...favoritesTrail.map((trailItem, index) => ({
                id: trailItem.id,
                label: trailItem.name,
                path: trailItem.path,
                isCurrent: index === favoritesTrail.length - 1,
            })),
        ];
    }, [favoritesTrail, selectedItem, t]);

    const totalItems = isRootListing
        ? starredCounts?.[favoritesCountKeyByFilter[activeFilter]]
        : undefined;

    const itemCountLabel = useMemo(() => {
        const count = totalItems ?? scopedItems.length;
        if (totalItems === undefined && hasNextPage) {
            return t('FAVORITES_ITEMS_OPEN_ENDED', { count: String(count) });
        }
        return `${count} ${count === 1 ? t('ITEM') : t('ITENS')}`;
    }, [scopedItems.length, hasNextPage, t, totalItems]);

    const currentTitle = selectedItem ? selectedItem.name : t('STARRED_FILES');
    const contextPath = selectedItem
        ? selectedItem.type === FileType.File
            ? selectedItem.parent_path
            : selectedItem.path
        : t('FAVORITES_CONTEXT_ALL');
    const activeFilterLabel =
        filterOptions.find((option) => option.value === activeFilter)?.label ??
        t('FAVORITES_FILTER_ALL');

    const closeSelectedItem = () => handleSelectItem(parentFolder);

    return {
        activeFilter,
        activeFilterLabel,
        breadcrumbSegments,
        closeSelectedItem,
        contextPath,
        currentTitle,
        fetchNextPage,
        filterOptions,
        listedItems: scopedItems,
        hasNextPage,
        isRootListing,
        isFetchingNextPage,
        itemCountLabel,
        selectedItem,
        setActiveFilter: setStarredKind,
        setStarredSort,
        setViewMode,
        sortOptions,
        starredSort,
        viewMode,
    };
};

export default useFavoritesScreen;
