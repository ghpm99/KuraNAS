import { useMemo } from 'react';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/utils';
import { getImageDirectoryPath } from '../imageLibraryData';

type ViewerDetailItem = {
    label: string;
    value: string;
};

export type ViewerDetailSection = {
    title: string;
    items: ViewerDetailItem[];
};

type UseImageViewerModalParams = {
    activeImage: IImageData;
    activeImageDate: Date | null;
    activeIndex: number;
    totalImages: number;
    dateFormatter: Intl.DateTimeFormat;
};

const classificationLabelKeys = {
    capture: 'IMAGES_CLASSIFICATION_CAPTURE',
    photo: 'IMAGES_CLASSIFICATION_PHOTO',
    other: 'IMAGES_CLASSIFICATION_OTHER',
} as const;

const formatNumberValue = (value?: number) => {
    if (!value) {
        return '';
    }

    return Number.isInteger(value) ? String(value) : value.toFixed(1);
};

const formatExposureValue = (value?: number) => {
    if (!value) {
        return '';
    }

    if (value > 0 && value < 1) {
        return `1/${Math.round(1 / value)}s`;
    }

    return `${value}s`;
};

const buildValue = (value: string | undefined, fallback: string) => value?.trim() || fallback;

export const useImageViewerModal = ({
    activeImage,
    activeImageDate,
    activeIndex,
    totalImages,
    dateFormatter,
}: UseImageViewerModalParams) => {
    const { t } = useI18n();

    return useMemo(() => {
        const folderPath = getImageDirectoryPath(activeImage);
        const classificationKey =
            classificationLabelKeys[activeImage.metadata?.classification?.category ?? 'other'];
        const resolution =
            activeImage.metadata?.width && activeImage.metadata?.height
                ? `${activeImage.metadata.width} x ${activeImage.metadata.height}`
                : t('COMMON_NOT_AVAILABLE');
        const deviceLabel = [activeImage.metadata?.make, activeImage.metadata?.model]
            .filter(Boolean)
            .join(' ');
        const positionLabel = t('IMAGES_VIEWER_POSITION', {
            current: String(activeIndex + 1),
            total: String(totalImages),
        });
        const confidenceValue = activeImage.metadata?.classification?.confidence;
        const confidenceLabel =
            typeof confidenceValue === 'number'
                ? `${Math.round(confidenceValue * 100)}%`
                : t('COMMON_NOT_AVAILABLE');

        const details: ViewerDetailSection[] = [
            {
                title: t('IMAGES_DETAILS_SECTION_LIBRARY'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_FOLDER'),
                        value: buildValue(folderPath, t('COMMON_NOT_AVAILABLE')),
                    },
                    {
                        label: t('IMAGES_DETAIL_FORMAT'),
                        value: buildValue(activeImage.format, t('COMMON_NOT_AVAILABLE')),
                    },
                    {
                        label: t('IMAGES_DETAIL_SIZE'),
                        value: formatSize(activeImage.size),
                    },
                    { label: t('IMAGES_DETAIL_DIMENSIONS'), value: resolution },
                    { label: t('IMAGES_DETAIL_CATEGORY'), value: t(classificationKey) },
                    { label: t('IMAGES_DETAIL_CONFIDENCE'), value: confidenceLabel },
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
                    {
                        label: t('IMAGES_DETAIL_CREATED'),
                        value: activeImage.created_at
                            ? dateFormatter.format(new Date(activeImage.created_at))
                            : t('COMMON_NOT_AVAILABLE'),
                    },
                    {
                        label: t('IMAGES_DETAIL_SOFTWARE'),
                        value: buildValue(
                            activeImage.metadata?.software,
                            t('COMMON_NOT_AVAILABLE')
                        ),
                    },
                    {
                        label: t('IMAGES_DETAIL_DESCRIPTION'),
                        value: buildValue(
                            activeImage.metadata?.image_description,
                            t('COMMON_NOT_AVAILABLE')
                        ),
                    },
                ],
            },
            {
                title: t('IMAGES_DETAILS_SECTION_DEVICE'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_CAMERA'),
                        value: buildValue(deviceLabel, t('COMMON_NOT_AVAILABLE')),
                    },
                    {
                        label: t('IMAGES_DETAIL_LENS'),
                        value: buildValue(
                            activeImage.metadata?.lens_model,
                            t('COMMON_NOT_AVAILABLE')
                        ),
                    },
                    {
                        label: t('IMAGES_DETAIL_ISO'),
                        value:
                            formatNumberValue(activeImage.metadata?.iso) ||
                            t('COMMON_NOT_AVAILABLE'),
                    },
                    {
                        label: t('IMAGES_DETAIL_FOCAL'),
                        value: formatNumberValue(activeImage.metadata?.focal_length)
                            ? `${formatNumberValue(activeImage.metadata?.focal_length)}mm`
                            : t('COMMON_NOT_AVAILABLE'),
                    },
                    {
                        label: t('IMAGES_DETAIL_APERTURE'),
                        value: formatNumberValue(activeImage.metadata?.f_number)
                            ? `f/${formatNumberValue(activeImage.metadata?.f_number)}`
                            : t('COMMON_NOT_AVAILABLE'),
                    },
                    {
                        label: t('IMAGES_DETAIL_EXPOSURE'),
                        value:
                            formatExposureValue(activeImage.metadata?.exposure_time) ||
                            t('COMMON_NOT_AVAILABLE'),
                    },
                ],
            },
        ];

        return {
            details,
            folderPath,
            positionLabel,
        };
    }, [activeImage, activeImageDate, activeIndex, dateFormatter, t, totalImages]);
};
