import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { discoverVacuumDevice, testVacuumConnection } from '@/service/vacuum';
import type { VacuumSettings } from '@/types/vacuum';
import { invalidateVacuumQueries } from './invalidateVacuumQueries';
import { resolveVacuumFailureMessage } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export type VacuumConnectionOutcome = {
    isSuccess: boolean;
    message: string;
};

export const useVacuumConnectionActions = () => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const [outcome, setOutcome] = useState<VacuumConnectionOutcome | undefined>();

    const reportFailure = (error: unknown) =>
        setOutcome({
            isSuccess: false,
            message: resolveVacuumFailureMessage(error, t('CONNECTION_BANNER_UNAVAILABLE')),
        });

    const discoveryMutation = useMutation({
        mutationFn: discoverVacuumDevice,
        onMutate: () => setOutcome(undefined),
        onSuccess: (discovery) => {
            const discoveredAddress = discovery?.address ?? '';
            queryClient.setQueryData<VacuumSettings | undefined>(
                vacuumQueryKeys.settings,
                (settings) => settings && { ...settings, address: discoveredAddress }
            );
            setOutcome({
                isSuccess: true,
                message: t('VACUUM_SETTINGS_DISCOVERY_FOUND', { address: discoveredAddress }),
            });
            void invalidateVacuumQueries(queryClient);
        },
        onError: reportFailure,
    });

    const testMutation = useMutation({
        mutationFn: testVacuumConnection,
        onMutate: () => setOutcome(undefined),
        onSuccess: () =>
            setOutcome({ isSuccess: true, message: t('VACUUM_SETTINGS_TEST_SUCCEEDED') }),
        onError: reportFailure,
    });

    return {
        outcome,
        isDiscovering: discoveryMutation.isPending,
        isTesting: testMutation.isPending,
        discover: () => discoveryMutation.mutate(),
        testConnection: () => testMutation.mutate(),
    };
};
