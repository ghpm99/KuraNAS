import { useCallback } from 'react';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';

export type OutcomeMessages = {
    singleSuccessKey: string;
    multipleSuccessKey: string;
    failureKey: string;
};

export type OutcomeCounts = {
    succeededCount: number;
    failedCount: number;
    firstFailureMessage?: string;
};

export const useBulkOutcomeNotifier = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();

    return useCallback(
        ({ succeededCount, failedCount, firstFailureMessage }: OutcomeCounts, messages: OutcomeMessages) => {
            const failureMessage = firstFailureMessage ?? t(messages.failureKey);

            if (failedCount === 0) {
                enqueueSnackbar(
                    succeededCount === 1
                        ? t(messages.singleSuccessKey)
                        : t(messages.multipleSuccessKey, { count: String(succeededCount) }),
                    { variant: 'success' }
                );
                return;
            }
            if (succeededCount === 0 && failedCount === 1) {
                enqueueSnackbar(failureMessage, { variant: 'error' });
                return;
            }
            if (succeededCount === 0) {
                enqueueSnackbar(
                    t('FILES_BULK_FAILED_SUMMARY', { failed: String(failedCount), message: failureMessage }),
                    { variant: 'error' }
                );
                return;
            }
            enqueueSnackbar(
                t('FILES_BULK_PARTIAL_SUMMARY', {
                    succeeded: String(succeededCount),
                    failed: String(failedCount),
                    message: failureMessage,
                }),
                { variant: 'warning' }
            );
        },
        [enqueueSnackbar, t]
    );
};
