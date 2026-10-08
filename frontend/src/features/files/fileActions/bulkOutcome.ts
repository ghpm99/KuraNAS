import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { runWithConcurrencyLimit } from '@/shared/utils/runWithConcurrencyLimit';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export { extractBackendErrorMessage };

const bulkConcurrencyLimit = 4;

export type BulkOutcome = {
    succeededFiles: FileData[];
    failedFiles: FileData[];
    firstFailureMessage?: string;
};

export const runBulkOperation = async (
    files: FileData[],
    operation: (file: FileData) => Promise<unknown>
): Promise<BulkOutcome> => {
    const settledTasks = await runWithConcurrencyLimit(files, bulkConcurrencyLimit, operation);
    const succeededFiles: FileData[] = [];
    const failedFiles: FileData[] = [];
    let firstFailureMessage: string | undefined;

    settledTasks.forEach((settledTask) => {
        if (settledTask.status === 'fulfilled') {
            succeededFiles.push(settledTask.input);
            return;
        }
        failedFiles.push(settledTask.input);
        firstFailureMessage ??= extractBackendErrorMessage(settledTask.reason);
    });

    return { succeededFiles, failedFiles, firstFailureMessage };
};
