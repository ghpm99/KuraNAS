import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getImageMetadataSummary } from '@/service/image';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import { formatSize } from '@/utils';
import { getImageCategoryLabelKey } from '../imageCategoryLabels';

type ViewerDetailItem = {
    label: string;
    value: string;
};

export type ViewerDetailSection = {
    title: string;
    items: ViewerDetailItem[];
};

type UseImageViewerModalParams = {
    activeImage: ImageLibraryItem;
    activeImageDate: Date | null;
    activeIndex: number;
    totalImages: number;
    dateFormatter: Intl.DateTimeFormat;
};

const formatNumberValue = (numericValue?: number) => {
    if (!numericValue) {
        return '';
    }

    return Number.isInteger(numericValue) ? String(numericValue) : numericValue.toFixed(1);
};

const formatExposureValue = (exposureSeconds?: number) => {
    if (!exposureSeconds) {
        return '';
    }

    if (exposureSeconds > 0 && exposureSeconds < 1) {
        return `1/${Math.round(1 / exposureSeconds)}s`;
    }

    return `${exposureSeconds}s`;
};

const buildValue = (rawValue: string | undefined, fallback: string) => rawValue?.trim() || fallback;

export const useImageViewerModal = ({
    activeImage,
    activeImageDate,
    activeIndex,
    totalImages,
    dateFormatter,
}: UseImageViewerModalParams) => {
    const { t } = useI18n();
    const { data: metadataSummary } = useQuery({
        queryKey: ['images', 'metadata', activeImage.file_id],
        queryFn: () => getImageMetadataSummary(activeImage.file_id),
        staleTime: listingStaleTimeMs,
        refetchOnWindowFocus: false,
        retry: false,
    });

    return useMemo(() => {
        const notAvailable = t('COMMON_NOT_AVAILABLE');
        const folderPath = activeImage.parent_path;
        const imageWidth = activeImage.width || metadataSummary?.width;
        const imageHeight = activeImage.height || metadataSummary?.height;
        const resolution =
            imageWidth && imageHeight ? `${imageWidth} x ${imageHeight}` : notAvailable;
        const deviceLabel = [metadataSummary?.make, metadataSummary?.model]
            .filter(Boolean)
            .join(' ');
        const focalLength = formatNumberValue(metadataSummary?.focal_length);
        const aperture = formatNumberValue(metadataSummary?.f_number);
        const positionLabel = t('IMAGES_VIEWER_POSITION', {
            current: String(activeIndex + 1),
            total: String(totalImages),
        });

        const details: ViewerDetailSection[] = [
            {
                title: t('IMAGES_DETAILS_SECTION_LIBRARY'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_FOLDER'),
                        value: buildValue(folderPath, notAvailable),
                    },
                    {
                        label: t('IMAGES_DETAIL_FORMAT'),
                        value: buildValue(activeImage.format, notAvailable),
                    },
                    { label: t('IMAGES_DETAIL_SIZE'), value: formatSize(activeImage.size) },
                    { label: t('IMAGES_DETAIL_DIMENSIONS'), value: resolution },
                    {
                        label: t('IMAGES_DETAIL_CATEGORY'),
                        value: t(getImageCategoryLabelKey(activeImage.category)),
                    },
                ],
            },
            {
                title: t('IMAGES_DETAILS_SECTION_CAPTURE'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_DATE'),
                        value: activeImageDate
                            ? dateFormatter.format(activeImageDate)
                            : t('IMAGES_DATE_UNAVAILABLE'),
                    },
                ],
            },
            {
                title: t('IMAGES_DETAILS_SECTION_DEVICE'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_CAMERA'),
                        value: buildValue(deviceLabel, notAvailable),
                    },
                    {
                        label: t('IMAGES_DETAIL_LENS'),
                        value: buildValue(metadataSummary?.lens_model, notAvailable),
                    },
                    {
                        label: t('IMAGES_DETAIL_ISO'),
                        value: formatNumberValue(metadataSummary?.iso) || notAvailable,
                    },
                    {
                        label: t('IMAGES_DETAIL_FOCAL'),
                        value: focalLength ? `${focalLength}mm` : notAvailable,
                    },
                    {
                        label: t('IMAGES_DETAIL_APERTURE'),
                        value: aperture ? `f/${aperture}` : notAvailable,
                    },
                    {
                        label: t('IMAGES_DETAIL_EXPOSURE'),
                        value: formatExposureValue(metadataSummary?.exposure_time) || notAvailable,
                    },
                ],
            },
        ];

        return {
            details,
            folderPath,
            positionLabel,
        };
    }, [activeImage, activeImageDate, activeIndex, dateFormatter, metadataSummary, t, totalImages]);
};
