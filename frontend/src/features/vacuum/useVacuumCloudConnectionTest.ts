import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { testVacuumCloudConnection } from '@/service/vacuum';
import type { VacuumConnectionOutcome } from './useVacuumConnectionActions';
import { resolveVacuumFailureMessage } from './vacuumFailures';

export const useVacuumCloudConnectionTest = () => {
    const { t } = useI18n();
    const [outcome, setOutcome] = useState<VacuumConnectionOutcome | undefined>();

    const testMutation = useMutation({
        mutationFn: testVacuumCloudConnection,
        onMutate: () => setOutcome(undefined),
        onSuccess: () =>
            setOutcome({ isSuccess: true, message: t('VACUUM_CLOUD_SETTINGS_TEST_SUCCEEDED') }),
        onError: (error) =>
            setOutcome({
                isSuccess: false,
                message: resolveVacuumFailureMessage(error, t('CONNECTION_BANNER_UNAVAILABLE')),
            }),
    });

    return {
        outcome,
        isTesting: testMutation.isPending,
        testCloud: () => testMutation.mutate(),
    };
};
