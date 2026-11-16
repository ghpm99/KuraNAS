import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getVacuumPreferences, updateVacuumPreferences } from '@/service/vacuum';
import type { VacuumPreferencesPatch } from '@/types/vacuum';
import { useVacuumFeedback } from './useVacuumFeedback';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumPreferences = () => {
    const queryClient = useQueryClient();
    const { notifySuccess, notifyFailure } = useVacuumFeedback();

    const preferencesQuery = useQuery({
        queryKey: vacuumQueryKeys.preferences,
        queryFn: getVacuumPreferences,
        retry: false,
    });

    const preferencesMutation = useMutation({
        mutationFn: (changes: VacuumPreferencesPatch) => updateVacuumPreferences(changes),
        onSuccess: (confirmedPreferences) => {
            queryClient.setQueryData(vacuumQueryKeys.preferences, confirmedPreferences);
            notifySuccess('VACUUM_PREFERENCE_SAVED');
        },
        onError: notifyFailure,
    });

    const savePreference = async (changes: VacuumPreferencesPatch): Promise<void> => {
        try {
            await preferencesMutation.mutateAsync(changes);
        } catch {
            return;
        }
    };

    const pendingPreferenceName = preferencesMutation.isPending
        ? Object.keys(preferencesMutation.variables ?? {})[0]
        : undefined;

    return { preferencesQuery, pendingPreferenceName, savePreference };
};
