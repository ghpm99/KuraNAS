import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { FileType, getFileTypeInfo } from '@/utils';
import { useMemo, useState } from 'react';
import type { BreadcrumbSegment } from '@/features/files/files/useFilesExplorerScreen';
import { findTrailById } from '@/features/files/files/fileNavigation';

export type FavoritesFilter = 'all' | 'folders' | 'files' | 'media';
export type FavoritesViewMode = 'grid' | 'list';

type FavoritesFilterOption = {
    value: FavoritesFilter;
    label: string;
    count: number;
};

const mediaTypes = new Set(['audio', 'image', 'video']);

const isMediaFile = (file: FileData) => {
    if (file.type !== FileType.File) {
        return false;
    }

    return mediaTypes.has(getFileTypeInfo(file.format).type);
};

const matchesFavoritesFilter = (filter: FavoritesFilter, file: FileData) => {
    switch (filter) {
        case 'folders':
            return file.type === FileType.Directory;
        case 'files':
            return file.type === FileType.File && !isMediaFile(file);
        case 'media':
            return isMediaFile(file);
        default:
            return true;
    }
};

const useFavoritesScreen = () => {
    const { t } = useI18n();
    const { files, selectedItem, fetchNextPage, hasNextPage, isFetchingNextPage, handleSelectItem } =
        useFile();
    const [activeFilter, setActiveFilter] = useState<FavoritesFilter>('all');
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

    const filterCounts = useMemo(
        () => ({
            all: scopedItems.length,
            folders: scopedItems.filter((item) => item.type === FileType.Directory).length,
            files: scopedItems.filter((item) => matchesFavoritesFilter('files', item)).length,
            media: scopedItems.filter((item) => matchesFavoritesFilter('media', item)).length,
        }),
        [scopedItems]
    );

    const filterOptions = useMemo<FavoritesFilterOption[]>(
        () => [
            {
                value: 'all',
                label: t('FAVORITES_FILTER_ALL'),
                count: filterCounts.all,
            },
            {
                value: 'folders',
                label: t('FAVORITES_FILTER_FOLDERS'),
                count: filterCounts.folders,
            },
            {
                value: 'files',
                label: t('FAVORITES_FILTER_FILES'),
                count: filterCounts.files,
            },
            {
                value: 'media',
                label: t('FAVORITES_FILTER_MEDIA'),
                count: filterCounts.media,
            },
        ],
        [filterCounts.all, filterCounts.files, filterCounts.folders, filterCounts.media, t]
    );

    const filteredItems = useMemo(
        () => scopedItems.filter((item) => matchesFavoritesFilter(activeFilter, item)),
        [activeFilter, scopedItems]
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

    const itemCountLabel = useMemo(() => {
        const count = filteredItems.length;
        if (hasNextPage) {
            return `${count}+ ${t('ITENS')}`;
        }
        return `${count} ${count === 1 ? t('ITEM') : t('ITENS')}`;
    }, [filteredItems.length, hasNextPage, t]);

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
        filteredItems,
        hasNextPage,
        isFetchingNextPage,
        itemCountLabel,
        selectedItem,
        setActiveFilter,
        setViewMode,
        viewMode,
    };
};

export const favoritesScreenUtils = {
    matchesFavoritesFilter,
};

export default useFavoritesScreen;
