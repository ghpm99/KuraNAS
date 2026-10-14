import { useQuery } from '@tanstack/react-query';
import { getVideoSummary } from '@/service/fileTypeMetadata';
import type { VideoFileDto } from '@/service/videoPlayback';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { buildVideoTechnicalRows, type VideoTechnicalRow } from './videoTechnicalRows';

type VideoTechnicalRowsState = {
    rows: VideoTechnicalRow[];
    isLoading: boolean;
    hasFailed: boolean;
    backendMessage?: string;
};

const summaryStaleTimeMs = 60_000;

const useVideoTechnicalRows = (video: VideoFileDto): VideoTechnicalRowsState => {
    const summaryQuery = useQuery({
        queryKey: ['videos', 'details', 'summary', video.id],
        queryFn: () => getVideoSummary(video.id),
        retry: false,
        staleTime: summaryStaleTimeMs,
    });

    return {
        rows: buildVideoTechnicalRows(video, summaryQuery.data),
        isLoading: summaryQuery.isLoading,
        hasFailed: summaryQuery.isError,
        backendMessage: extractBackendErrorMessage(summaryQuery.error),
    };
};

export default useVideoTechnicalRows;
