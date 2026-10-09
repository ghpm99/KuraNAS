import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import { getFileLocation } from '@/service/files';
import { getImageMetadataSummary } from '@/service/image';
import type { ImageLibraryItem, ImageMetadataSummary } from '@/types/imageLibrary';
import { formatSize } from '@/utils';
import { getImageCategoryLabelKey } from '../imageCategoryLabels';
import { parseTakenAt } from '../imageDateGroups';

export type ViewerDetailItem = {
    label: string;
    value: string;
    link?: { href: string; label: string };
    copyValue?: string;
};

export type ViewerDetailSection = {
    title: string;
    items: ViewerDetailItem[];
};

export type ViewerContent = {
    caption: string;
    tags: string[];
    ocrText: string;
};

type UseImageViewerModalParams = {
    activeImage: ImageLibraryItem;
    activeImageDate: Date | null;
    activeIndex: number;
    totalImages: number | null;
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

const gpsCoordinateDecimals = 6;

const buildOpenStreetMapUrl = (latitude: number, longitude: number) => {
    const latitudeText = latitude.toFixed(gpsCoordinateDecimals);
    const longitudeText = longitude.toFixed(gpsCoordinateDecimals);
    return `https://www.openstreetmap.org/?mlat=${latitudeText}&mlon=${longitudeText}#map=15/${latitudeText}/${longitudeText}`;
};

type GpsCoordinates = { latitude: number; longitude: number };

const readGpsCoordinates = (summary?: ImageMetadataSummary): GpsCoordinates | null => {
    const latitude = summary?.gps_latitude;
    const longitude = summary?.gps_longitude;
    if (typeof latitude !== 'number' || typeof longitude !== 'number') {
        return null;
    }
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
        return null;
    }
    return latitude === 0 && longitude === 0 ? null : { latitude, longitude };
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

    const { data: fileLocation } = useQuery({
        queryKey: ['images', 'location', activeImage.file_id],
        queryFn: () => getFileLocation(activeImage.file_id),
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
        const positionLabel =
            totalImages === null
                ? ''
                : t('IMAGES_VIEWER_POSITION', {
                      current: String(activeIndex + 1),
                      total: String(totalImages),
                  });
        const diskPath = fileLocation?.disk_path?.trim() ?? '';
        const capturedAt = activeImageDate ?? parseTakenAt(metadataSummary?.taken_at ?? null);
        const gpsCoordinates = readGpsCoordinates(metadataSummary);
        const gpsItem: ViewerDetailItem = gpsCoordinates
            ? {
                  label: t('IMAGES_DETAIL_GPS'),
                  value: `${gpsCoordinates.latitude.toFixed(gpsCoordinateDecimals)}, ${gpsCoordinates.longitude.toFixed(gpsCoordinateDecimals)}`,
                  link: {
                      href: buildOpenStreetMapUrl(
                          gpsCoordinates.latitude,
                          gpsCoordinates.longitude
                      ),
                      label: t('IMAGES_DETAIL_GPS_OPEN_MAP'),
                  },
              }
            : { label: t('IMAGES_DETAIL_GPS'), value: notAvailable };
        const classificationConfidence = metadataSummary?.classification_confidence;

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
                    {
                        label: t('IMAGES_DETAIL_DISK_LOCATION'),
                        value: buildValue(diskPath, notAvailable),
                        copyValue: diskPath || undefined,
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
                        value: capturedAt
                            ? dateFormatter.format(capturedAt)
                            : t('IMAGES_DATE_UNAVAILABLE'),
                    },
                    gpsItem,
                    {
                        label: t('IMAGES_DETAIL_SOFTWARE'),
                        value: buildValue(metadataSummary?.software, notAvailable),
                    },
                    {
                        label: t('IMAGES_DETAIL_DESCRIPTION'),
                        value: buildValue(metadataSummary?.image_description, notAvailable),
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
            {
                title: t('IMAGES_DETAILS_SECTION_AI'),
                items: [
                    {
                        label: t('IMAGES_DETAIL_CONFIDENCE'),
                        value: classificationConfidence
                            ? `${Math.round(classificationConfidence * 100)}%`
                            : notAvailable,
                    },
                    {
                        label: t('IMAGES_DETAIL_SUGGESTED_NAME'),
                        value: buildValue(metadataSummary?.suggested_name, notAvailable),
                    },
                ],
            },
        ];

        const content: ViewerContent = {
            caption: metadataSummary?.caption?.trim() ?? '',
            tags: metadataSummary?.tags ?? [],
            ocrText: metadataSummary?.ocr_text?.trim() ?? '',
        };

        return {
            details,
            content,
            folderPath,
            positionLabel,
        };
    }, [
        activeImage,
        activeImageDate,
        activeIndex,
        dateFormatter,
        fileLocation,
        metadataSummary,
        t,
        totalImages,
    ]);
};
