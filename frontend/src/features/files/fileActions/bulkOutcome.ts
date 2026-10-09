import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { runBulkOperation as runBulkItemsOperation } from '@/shared/bulk/runBulkOperation';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export { extractBackendErrorMessage };

export type BulkOutcome = {
    succeededFiles: FileData[];
    failedFiles: FileData[];
    firstFailureMessage?: string;
};

export const runBulkOperation = async (
    files: FileData[],
    operation: (file: FileData) => Promise<unknown>
): Promise<BulkOutcome> => {
    const { succeededItems, failedItems, firstFailureMessage } = await runBulkItemsOperation(
        files,
        operation
    );
    return { succeededFiles: succeededItems, failedFiles: failedItems, firstFailureMessage };
};
