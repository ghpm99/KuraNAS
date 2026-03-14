import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { type FileData } from '@/components/providers/fileProvider/fileContext';
import { FileType, getFileTypeInfo } from '@/utils';
import { useMemo, useState } from 'react';
import type { BreadcrumbSegment } from '@/components/files/useFilesExplorerScreen';
import { findTrailById } from '@/components/files/fileNavigation';

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
	const { files, selectedItem, handleSelectItem } = useFile();
	const [activeFilter, setActiveFilter] = useState<FavoritesFilter>('all');
	const [viewMode, setViewMode] = useState<FavoritesViewMode>('grid');

	const scopedItems = useMemo(() => {
		if (!selectedItem) {
			return files;
		}

		if (selectedItem.type === FileType.Directory) {
			return selectedItem.file_children ?? [];
		}

		return [];
	}, [files, selectedItem]);

	const filterCounts = useMemo(
		() => ({
			all: scopedItems.length,
			folders: scopedItems.filter((item) => item.type === FileType.Directory).length,
			files: scopedItems.filter((item) => matchesFavoritesFilter('files', item)).length,
			media: scopedItems.filter((item) => matchesFavoritesFilter('media', item)).length,
		}),
		[scopedItems],
	);

	const filterOptions = useMemo<FavoritesFilterOption[]>(
		() => [
			{ value: 'all', label: t('FAVORITES_FILTER_ALL'), count: filterCounts.all },
			{ value: 'folders', label: t('FAVORITES_FILTER_FOLDERS'), count: filterCounts.folders },
			{ value: 'files', label: t('FAVORITES_FILTER_FILES'), count: filterCounts.files },
			{ value: 'media', label: t('FAVORITES_FILTER_MEDIA'), count: filterCounts.media },
		],
		[filterCounts.all, filterCounts.files, filterCounts.folders, filterCounts.media, t],
	);

	const filteredItems = useMemo(
		() => scopedItems.filter((item) => matchesFavoritesFilter(activeFilter, item)),
		[activeFilter, scopedItems],
	);

	const breadcrumbSegments = useMemo<BreadcrumbSegment[]>(() => {
		const rootSegment = {
			id: null,
			label: t('STARRED_FILES'),
			isCurrent: !selectedItem,
		};

		if (!selectedItem) {
			return [rootSegment];
		}

		const trail = findTrailById(files, selectedItem.id) ?? [];
		if (trail.length === 0) {
			return [
				rootSegment,
				{
					id: selectedItem.type === FileType.File ? null : selectedItem.id,
					label: selectedItem.name,
					isCurrent: true,
				},
			];
		}

		return [
			rootSegment,
			...trail.map((item, index) => ({
				id: item.id,
				label: item.name,
				isCurrent: index === trail.length - 1,
			})),
		];
	}, [files, selectedItem, t]);

	const itemCountLabel = useMemo(() => {
		const count = filteredItems.length;
		return `${count} ${count === 1 ? t('ITEM') : t('ITENS')}`;
	}, [filteredItems.length, t]);

	const currentTitle = selectedItem ? selectedItem.name : t('STARRED_FILES');
	const contextPath = selectedItem
		? selectedItem.type === FileType.File
			? selectedItem.parent_path
			: selectedItem.path
		: t('FAVORITES_CONTEXT_ALL');
	const activeFilterLabel = filterOptions.find((option) => option.value === activeFilter)?.label ?? t('FAVORITES_FILTER_ALL');

	return {
		activeFilter,
		activeFilterLabel,
		breadcrumbSegments,
		contextPath,
		currentTitle,
		filterOptions,
		filteredItems,
		handleSelectItem,
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
