import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import { useImageViewer } from '@/components/hooks/useImageViewer/useImageViewer';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useImage } from '@/components/providers/imageProvider/imageProvider';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import { getFileByPath } from '@/service/files';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import {
    buildTimelineCountsByMonth,
    groupImagesByMonth,
    groupImagesWithoutHeader,
    parseTakenAt,
} from './imageDateGroups';
import { buildFolderCards } from './imageFolderCards';
import { mapFileDataToLibraryItem, readImageLibraryItemId } from './imageLibraryItemMapping';
import { useDebouncedNameQuery } from './useDebouncedNameQuery';
import { useImageAlbumCards } from './useImageAlbumCards';
import { useImageLibraryControls } from './useImageLibraryControls';
import { useImageStarToggle } from './useImageStarToggle';

export type ImageEmptyKind = 'library' | 'filtered' | 'favorites';

const sectionTitleKeys = {
    library: 'IMAGES_SECTION_LIBRARY',
    recent: 'IMAGES_SECTION_RECENT',
    captures: 'IMAGES_SECTION_CAPTURES',
    photos: 'IMAGES_SECTION_PHOTOS',
    favorites: 'IMAGES_SECTION_FAVORITES',
    folders: 'IMAGES_SECTION_FOLDERS',
    albums: 'IMAGES_SECTION_ALBUMS',
} as const;

const resolveEmptyKind = (
    section: string,
    hasUserFilters: boolean,
    hasSectionFilter: boolean
): ImageEmptyKind => {
    if (hasUserFilters) {
        return 'filtered';
    }
    if (section === 'favorites') {
        return 'favorites';
    }
    return hasSectionFilter ? 'filtered' : 'library';
};

export const useImageContent = () => {
    const { t } = useI18n();
    const { settings } = useSettings();
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const {
        view,
        items,
        status,
        error,
        isFetchNextPageError,
        total,
        timeline,
        fetchNextPage,
        refetch,
        hasNextPage,
        isFetchingNextPage,
    } = useImage();
    const controls = useImageLibraryControls(view);
    const { openImageParam, closeImageParam } = controls;
    const { toggleStar, isStarTogglePending } = useImageStarToggle();
    const { typedNameQuery, setTypedNameQuery } = useDebouncedNameQuery(
        view.filters.nameQuery,
        controls.setNameQuery
    );
    const locale = t('LOCALE');
    const { section, selectedFolder, selectedAlbum, ordering } = view;

    const requestedImageId = Number(searchParams.get('image') ?? '');
    const requestedImagePath = searchParams.get('imagePath')?.trim() ?? '';
    const isRequestedImageLoaded = items.some((image) => image.file_id === requestedImageId);
    const shouldResolveRequestedImage =
        Number.isFinite(requestedImageId) &&
        requestedImageId > 0 &&
        requestedImagePath.length > 0 &&
        !isRequestedImageLoaded;
    const requestedImageQuery = useQuery({
        queryKey: ['images', 'path', requestedImagePath],
        queryFn: () => getFileByPath(requestedImagePath),
        enabled: shouldResolveRequestedImage,
    });
    const resolvedRequestedImage =
        requestedImageQuery.data && requestedImageQuery.data.id === requestedImageId
            ? mapFileDataToLibraryItem(requestedImageQuery.data)
            : null;
    const viewerImages = useMemo(
        () => (resolvedRequestedImage ? [...items, resolvedRequestedImage] : items),
        [items, resolvedRequestedImage]
    );

    const monthFormatter = useMemo(
        () => new Intl.DateTimeFormat(locale, { month: 'long', year: 'numeric', timeZone: 'UTC' }),
        [locale]
    );
    const dateFormatter = useMemo(
        () => new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' }),
        [locale]
    );

    const viewMode =
        section === 'folders' && !selectedFolder
            ? 'folders'
            : section === 'albums' && !selectedAlbum
              ? 'albums'
              : 'grid';

    const groups = useMemo(() => {
        if (viewMode !== 'grid') {
            return [];
        }
        if (ordering.sort !== 'taken_at') {
            return groupImagesWithoutHeader(items);
        }
        return groupImagesByMonth({
            items,
            formatMonth: (date) => monthFormatter.format(date),
            undatedLabel: t('IMAGES_GROUP_NO_DATE'),
            countsByMonth: buildTimelineCountsByMonth(timeline),
        });
    }, [viewMode, ordering.sort, items, monthFormatter, t, timeline]);

    const folderCards = useMemo(
        () => (viewMode === 'folders' ? buildFolderCards(items) : []),
        [viewMode, items]
    );
    const albumCards = useImageAlbumCards(viewMode === 'albums');

    const viewer = useImageViewer(
        viewerImages,
        settings.players.image_slideshow_seconds * 1000,
        readImageLibraryItemId
    );
    const { activeImage, openImage, closeViewer } = viewer;
    const activeImageDate = activeImage ? parseTakenAt(activeImage.taken_at) : null;

    const handleOpenImage = useCallback(
        (fileId: number) => {
            openImage(fileId);
            openImageParam(fileId);
        },
        [openImage, openImageParam]
    );

    const handleCloseViewer = useCallback(() => {
        closeViewer();
        closeImageParam();
    }, [closeViewer, closeImageParam]);

    const handleToggleFavoriteOfActiveImage = useCallback(() => {
        if (!activeImage || isStarTogglePending) {
            return;
        }
        toggleStar(activeImage.file_id, activeImage.starred);
    }, [activeImage, isStarTogglePending, toggleStar]);

    const handleOpenActiveImageFolder = useCallback(() => {
        if (!activeImage?.parent_path) {
            return;
        }
        handleCloseViewer();
        navigate({
            pathname: appRoutes.files,
            search: `?${new URLSearchParams({ path: activeImage.parent_path }).toString()}`,
        });
    }, [activeImage, handleCloseViewer, navigate]);

    useEffect(() => {
        if (Number.isFinite(requestedImageId) && requestedImageId > 0) {
            if (viewerImages.some((image) => image.file_id === requestedImageId)) {
                openImage(requestedImageId);
            }
        }
    }, [viewerImages, openImage, requestedImageId]);

    const hasLoadError = status === 'error';
    const loadErrorMessage = hasLoadError ? extractBackendErrorMessage(error) : undefined;
    const retryLoading = useCallback(() => {
        if (isFetchNextPageError) {
            fetchNextPage();
            return;
        }
        refetch();
    }, [isFetchNextPageError, fetchNextPage, refetch]);

    const loadNextPage = useCallback(() => {
        fetchNextPage();
    }, [fetchNextPage]);

    const selectedCollectionTitle =
        section === 'albums' && selectedAlbum ? t(selectedAlbum.titleKey) : selectedFolder;
    const title = selectedCollectionTitle || t(sectionTitleKeys[section]);

    const summary =
        viewMode === 'folders'
            ? t('IMAGES_FOLDERS_SUMMARY', { count: String(folderCards.length) })
            : viewMode === 'albums'
              ? t('IMAGES_ALBUMS_SUMMARY', { count: String(albumCards.length) })
              : total === null
                ? ''
                : t('IMAGES_PHOTOS_COUNT', { count: String(total) });

    const hasSectionFilter = section === 'recent' || section === 'captures' || section === 'photos';

    return {
        view,
        viewMode,
        title,
        summary,
        groups,
        folderCards,
        albumCards,
        timeline,
        status,
        hasLoadError,
        loadErrorMessage,
        isFetchNextPageError,
        retryLoading,
        hasNextPage,
        isFetchingNextPage,
        loadNextPage,
        isEmpty: status === 'success' && items.length === 0,
        emptyKind: resolveEmptyKind(section, view.hasUserFilters, hasSectionFilter),
        typedNameQuery,
        setTypedNameQuery,
        controls,
        dateFormatter,
        monthFormatter,
        viewerImages,
        viewer,
        activeImageDate,
        isFavoritePending: isStarTogglePending,
        toggleStar,
        handleOpenImage,
        handleCloseViewer,
        handleToggleFavoriteOfActiveImage,
        handleOpenActiveImageFolder,
    };
};
