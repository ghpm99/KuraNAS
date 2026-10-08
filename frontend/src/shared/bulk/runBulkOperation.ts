import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { runWithConcurrencyLimit } from '@/shared/utils/runWithConcurrencyLimit';

const bulkConcurrencyLimit = 4;

export type BulkItemsOutcome<TItem> = {
    succeededItems: TItem[];
    failedItems: TItem[];
    firstFailureMessage?: string;
};

export const runBulkOperation = async <TItem>(
    items: TItem[],
    operation: (item: TItem) => Promise<unknown>
): Promise<BulkItemsOutcome<TItem>> => {
    const settledTasks = await runWithConcurrencyLimit(items, bulkConcurrencyLimit, operation);
    const succeededItems: TItem[] = [];
    const failedItems: TItem[] = [];
    let firstFailureMessage: string | undefined;

    settledTasks.forEach((settledTask) => {
        if (settledTask.status === 'fulfilled') {
            succeededItems.push(settledTask.input);
            return;
        }
        failedItems.push(settledTask.input);
        firstFailureMessage ??= extractBackendErrorMessage(settledTask.reason);
    });

    return { succeededItems, failedItems, firstFailureMessage };
};
