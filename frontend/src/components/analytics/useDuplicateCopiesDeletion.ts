import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import useI18n from '@/components/i18n/provider/i18nContext';
import { deleteFile } from '@/service/files';
import { bulkConcurrencyLimit } from '@/shared/bulk/runBulkOperation';
import { allFileQueryKeys } from '@/shared/queryKeys/fileQueryKeys';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { runWithConcurrencyLimit } from '@/shared/utils/runWithConcurrencyLimit';
import type { DuplicateCopy } from '@/types/analytics';

export type DuplicateCopyDeletionFailure = {
    copy: DuplicateCopy;
    message: string;
};

export type DuplicateCopiesDeletionOutcome = {
    trashedCount: number;
    failures: DuplicateCopyDeletionFailure[];
};

type UseDuplicateCopiesDeletionParams = {
    onCopiesTrashed: (trashedCopies: DuplicateCopy[]) => void;
};

export const useDuplicateCopiesDeletion = ({
    onCopiesTrashed,
}: UseDuplicateCopiesDeletionParams) => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const [copiesAwaitingConfirmation, setCopiesAwaitingConfirmation] = useState<DuplicateCopy[]>(
        []
    );
    const [outcome, setOutcome] = useState<DuplicateCopiesDeletionOutcome | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);

    const requestDeletion = (copies: DuplicateCopy[]) => setCopiesAwaitingConfirmation(copies);

    const cancelDeletion = () => setCopiesAwaitingConfirmation([]);

    const dismissOutcome = () => setOutcome(null);

    const refreshListsAfterDeletion = () =>
        Promise.all([
            queryClient.invalidateQueries({ queryKey: ['analytics'] }),
            ...allFileQueryKeys.map((queryKey) =>
                queryClient.invalidateQueries({ queryKey: [queryKey] })
            ),
        ]);

    const confirmDeletion = async () => {
        const copiesToTrash = copiesAwaitingConfirmation;
        setIsDeleting(true);
        try {
            const settledDeletions = await runWithConcurrencyLimit(
                copiesToTrash,
                bulkConcurrencyLimit,
                (copy) => deleteFile(copy.id)
            );
            const trashedCopies = settledDeletions.flatMap((settled) =>
                settled.status === 'fulfilled' ? [settled.input] : []
            );
            const failures = settledDeletions.flatMap((settled) =>
                settled.status === 'rejected'
                    ? [
                          {
                              copy: settled.input,
                              message:
                                  extractBackendErrorMessage(settled.reason) ??
                                  t('ERROR_DELETE_FAILED'),
                          },
                      ]
                    : []
            );
            setCopiesAwaitingConfirmation([]);
            setOutcome({ trashedCount: trashedCopies.length, failures });
            onCopiesTrashed(trashedCopies);
            await refreshListsAfterDeletion();
        } finally {
            setIsDeleting(false);
        }
    };

    return {
        copiesAwaitingConfirmation,
        outcome,
        isDeleting,
        requestDeletion,
        cancelDeletion,
        confirmDeletion,
        dismissOutcome,
    };
};
