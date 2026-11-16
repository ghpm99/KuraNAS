import { useSnackbar } from 'notistack';
import { useCallback } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { resolveVacuumFailureMessage } from './vacuumFailures';

export const useVacuumFeedback = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();

    const notifySuccess = useCallback(
        (messageKey: string) => {
            enqueueSnackbar(t(messageKey), { variant: 'success' });
        },
        [enqueueSnackbar, t]
    );

    const notifyFailure = useCallback(
        (error: unknown) => {
            enqueueSnackbar(resolveVacuumFailureMessage(error, t('CONNECTION_BANNER_UNAVAILABLE')), {
                variant: 'error',
            });
        },
        [enqueueSnackbar, t]
    );

    return { notifySuccess, notifyFailure };
};
