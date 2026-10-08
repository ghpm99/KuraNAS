import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getAudioSummary, getImageSummary, getVideoSummary } from '@/service/fileTypeMetadata';
import { FileType, getFileTypeInfo } from '@/utils';
import { useQuery } from '@tanstack/react-query';
import { buildAudioRows, buildImageRows, buildVideoRows } from './fileTypeMetadataRows';

export type TranslatedMetadataRow = {
    label: string;
    value: string;
};

const summaryStaleTimeMs = 60_000;

const useFileTypeMetadataRows = (
    file: Pick<FileData, 'id' | 'type' | 'format'>
): TranslatedMetadataRow[] => {
    const { t } = useI18n();
    const category = getFileTypeInfo(file.format).type;
    const isFile = file.type === FileType.File;

    const queryOptions = (isMatchingCategory: boolean) => ({
        enabled: isFile && isMatchingCategory,
        retry: false,
        staleTime: summaryStaleTimeMs,
    });

    const imageSummary = useQuery({
        queryKey: ['files', 'type-metadata', 'image', file.id],
        queryFn: () => getImageSummary(file.id),
        ...queryOptions(category === 'image'),
    });
    const audioSummary = useQuery({
        queryKey: ['files', 'type-metadata', 'audio', file.id],
        queryFn: () => getAudioSummary(file.id),
        ...queryOptions(category === 'audio'),
    });
    const videoSummary = useQuery({
        queryKey: ['files', 'type-metadata', 'video', file.id],
        queryFn: () => getVideoSummary(file.id),
        ...queryOptions(category === 'video'),
    });

    const rows = (() => {
        if (!isFile) return [];
        if (category === 'image' && imageSummary.data) return buildImageRows(imageSummary.data);
        if (category === 'audio' && audioSummary.data) return buildAudioRows(audioSummary.data);
        if (category === 'video' && videoSummary.data) return buildVideoRows(videoSummary.data);
        return [];
    })();

    return rows.map((row) => ({ label: t(row.labelKey), value: row.value }));
};

export default useFileTypeMetadataRows;
