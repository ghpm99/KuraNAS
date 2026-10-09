import { getFileTextPreview, textPreviewMaxBytes } from '@/service/files';
import { useQuery } from '@tanstack/react-query';

const previewStaleTimeMs = 60_000;

type TextFilePreview = {
    text: string;
    isTruncated: boolean;
};

const useTextFilePreview = (fileId: number, fileSizeBytes: number | undefined) =>
    useQuery<TextFilePreview>({
        queryKey: ['files', 'text-preview', fileId],
        queryFn: async () => {
            const body = await getFileTextPreview(fileId);
            const text = typeof body === 'string' ? body.slice(0, textPreviewMaxBytes) : '';
            const isTruncated =
                (fileSizeBytes ?? 0) > textPreviewMaxBytes ||
                (typeof body === 'string' && body.length > textPreviewMaxBytes);
            return { text, isTruncated };
        },
        staleTime: previewStaleTimeMs,
        retry: false,
    });

export default useTextFilePreview;
