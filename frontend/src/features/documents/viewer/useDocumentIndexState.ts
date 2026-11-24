import { useQuery } from '@tanstack/react-query';
import { getDocumentIndexState } from '@/service/documents';
import { documentIndexStateQueryKey } from '../documentsQueryKeys';

const useDocumentIndexState = (fileId: number) =>
    useQuery({
        queryKey: documentIndexStateQueryKey(fileId),
        queryFn: () => getDocumentIndexState(fileId),
        retry: false,
    });

export default useDocumentIndexState;
