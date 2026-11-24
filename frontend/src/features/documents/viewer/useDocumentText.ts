import { useQuery } from '@tanstack/react-query';
import { getDocumentText } from '@/service/documents';
import { documentTextQueryKey } from '../documentsQueryKeys';

const useDocumentText = (fileId: number, isEnabled: boolean) =>
    useQuery({
        queryKey: documentTextQueryKey(fileId),
        queryFn: () => getDocumentText(fileId),
        enabled: isEnabled,
        retry: false,
    });

export default useDocumentText;
