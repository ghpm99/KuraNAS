import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { buildFilesUrl } from '@/app/routes';
import { parentFolderOf } from '@/components/trash/trashItemPresentation';
import { getFileLocation } from '@/service/files';

const folderUrlOf = (logicalPath: string): string => {
    const folderPath = parentFolderOf(logicalPath);
    return buildFilesUrl(folderPath === '/' ? '' : folderPath);
};

/**
 * Leva um item de lista (download, captura…) ao arquivo da biblioteca que ele gerou:
 * `openFile` abre o arquivo em Arquivos e `openFolder` abre a pasta que o contém.
 * Quando a localização não pode ser resolvida, `failedItemId` guarda o id do item da lista.
 */
export const useOpenLibraryFileLocation = () => {
    const navigate = useNavigate();
    const [failedItemId, setFailedItemId] = useState<number | null>(null);

    const navigateToFile = async (
        itemId: number,
        fileId: number,
        buildUrl: (logicalPath: string) => string
    ) => {
        setFailedItemId(null);
        try {
            const location = await getFileLocation(fileId);
            navigate(buildUrl(location.logical_path));
        } catch {
            setFailedItemId(itemId);
        }
    };

    return {
        failedItemId,
        openFile: (itemId: number, fileId: number) => navigateToFile(itemId, fileId, buildFilesUrl),
        openFolder: (itemId: number, fileId: number) => navigateToFile(itemId, fileId, folderUrlOf),
    };
};
