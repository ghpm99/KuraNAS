import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { buildFilesUrl } from '@/app/routes';
import { parentFolderOf } from '@/components/trash/trashItemPresentation';
import { getFileLocation } from '@/service/files';

const folderUrlOf = (logicalPath: string): string => {
    const folderPath = parentFolderOf(logicalPath);
    return buildFilesUrl(folderPath === '/' ? '' : folderPath);
};

export const useOpenIngestDownloadLocation = () => {
    const navigate = useNavigate();
    const [failedDownloadId, setFailedDownloadId] = useState<number | null>(null);

    const navigateToFile = async (
        downloadId: number,
        fileId: number,
        buildUrl: (logicalPath: string) => string
    ) => {
        setFailedDownloadId(null);
        try {
            const location = await getFileLocation(fileId);
            navigate(buildUrl(location.logical_path));
        } catch {
            setFailedDownloadId(downloadId);
        }
    };

    return {
        failedDownloadId,
        openFile: (downloadId: number, fileId: number) =>
            navigateToFile(downloadId, fileId, buildFilesUrl),
        openFolder: (downloadId: number, fileId: number) =>
            navigateToFile(downloadId, fileId, folderUrlOf),
    };
};
