import useFile from '@/features/files/providers/fileProvider/fileContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import { FileType } from '@/utils';
import { useMemo, useState } from 'react';
import { findTrailById } from './fileNavigation';

export type FilesViewMode = 'grid' | 'list';

export type BreadcrumbSegment = {
    id: number | null;
    label: string;
    path: string | null;
    isCurrent: boolean;
};

const useFilesExplorerScreen = () => {
    const { t } = useI18n();
    const { files, selectedItem, fileListFilter } = useFile();
    const [viewMode, setViewMode] = useState<FilesViewMode>('grid');
    const [mobileTreeOpen, setMobileTreeOpen] = useState(false);

    const currentListTitle = useMemo(() => {
        if (fileListFilter === 'starred') {
            return t('STARRED_FILES');
        }

        if (fileListFilter === 'recent') {
            return t('RECENT_FILES');
        }

        return t('FILES');
    }, [fileListFilter, t]);

    const currentItems = useMemo(() => {
        if (!selectedItem) {
            return files;
        }

        if (selectedItem.type === FileType.Directory) {
            return selectedItem.file_children ?? [];
        }

        return [];
    }, [files, selectedItem]);

    const breadcrumbSegments = useMemo<BreadcrumbSegment[]>(() => {
        const rootSegment: BreadcrumbSegment = {
            id: null,
            label: t('FILES'),
            path: null,
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
                    id: selectedItem.id,
                    label: selectedItem.name,
                    path: selectedItem.path,
                    isCurrent: true,
                },
            ];
        }

        return [
            rootSegment,
            ...trail.map((item, index) => ({
                id: item.id,
                label: item.name,
                path: item.path,
                isCurrent: index === trail.length - 1,
            })),
        ];
    }, [files, selectedItem, t]);

    const itemCountLabel = useMemo(() => {
        const count = currentItems.length;
        return `${count} ${count === 1 ? t('ITEM') : t('ITENS')}`;
    }, [currentItems.length, t]);

    const contextLabel = selectedItem
        ? selectedItem.type === FileType.File
            ? selectedItem.parent_path
            : selectedItem.path
        : currentListTitle;

    return {
        breadcrumbSegments,
        contextLabel,
        currentListTitle,
        itemCountLabel,
        mobileTreeOpen,
        openMobileTree: () => setMobileTreeOpen(true),
        closeMobileTree: () => setMobileTreeOpen(false),
        selectedItem,
        setViewMode,
        viewMode,
    };
};

export default useFilesExplorerScreen;
