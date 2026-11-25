import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRef, useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { updateVacuumCloudSettings } from '@/service/vacuum';
import type { VacuumCloudSettingsChanges } from '@/types/vacuum';
import { invalidateVacuumQueries } from './invalidateVacuumQueries';
import type { SavedVacuumCloudSettings } from './toSavedVacuumCloudSettings';
import { useVacuumFeedback } from './useVacuumFeedback';
import { resolveVacuumFailureMessage } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumCloudSettingsForm = (savedSettings: SavedVacuumCloudSettings) => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const { notifySuccess } = useVacuumFeedback();
    const [editedAccessId, setEditedAccessId] = useState<string | undefined>();
    const [editedRegion, setEditedRegion] = useState<string | undefined>();
    const [accessSecret, setAccessSecret] = useState('');
    const isSaveInFlightRef = useRef(false);

    const accessId = editedAccessId ?? savedSettings.accessId;
    const region = editedRegion ?? savedSettings.region;
    const hasUnsavedChanges =
        accessId !== savedSettings.accessId ||
        region !== savedSettings.region ||
        accessSecret !== '';

    const saveMutation = useMutation({
        mutationFn: (changes: VacuumCloudSettingsChanges) => updateVacuumCloudSettings(changes),
        onSuccess: (confirmedSettings) => {
            queryClient.setQueryData(vacuumQueryKeys.cloudSettings, confirmedSettings);
            setEditedAccessId(undefined);
            setEditedRegion(undefined);
            setAccessSecret('');
            notifySuccess('VACUUM_CLOUD_SETTINGS_SAVED');
            void invalidateVacuumQueries(queryClient);
        },
        onSettled: () => {
            isSaveInFlightRef.current = false;
        },
    });

    const save = () => {
        if (isSaveInFlightRef.current) return;
        isSaveInFlightRef.current = true;
        saveMutation.mutate({ access_id: accessId, access_secret: accessSecret, region });
    };

    const setAccessId = (typedAccessId: string) =>
        setEditedAccessId(typedAccessId === savedSettings.accessId ? undefined : typedAccessId);

    const setRegion = (selectedRegion: string) =>
        setEditedRegion(selectedRegion === savedSettings.region ? undefined : selectedRegion);

    const saveFailureMessage = saveMutation.isError
        ? resolveVacuumFailureMessage(saveMutation.error, t('CONNECTION_BANNER_UNAVAILABLE'))
        : undefined;

    return {
        accessId,
        accessSecret,
        region,
        hasStoredAccessSecret: savedSettings.hasAccessSecret,
        hasUnsavedChanges,
        isSaving: saveMutation.isPending,
        saveFailureMessage,
        testBlockedMessageKey: resolveTestBlockedMessageKey(savedSettings, hasUnsavedChanges),
        setAccessId,
        setAccessSecret,
        setRegion,
        save,
    };
};

const resolveTestBlockedMessageKey = (
    savedSettings: SavedVacuumCloudSettings,
    hasUnsavedChanges: boolean
): string | undefined => {
    const isCloudConfigured =
        savedSettings.accessId !== '' &&
        savedSettings.region !== '' &&
        savedSettings.hasAccessSecret;
    if (!isCloudConfigured) return 'VACUUM_CLOUD_SETTINGS_TEST_NEEDS_SAVED_CREDENTIALS';
    if (hasUnsavedChanges) return 'VACUUM_CLOUD_SETTINGS_TEST_SAVE_FIRST';
    return undefined;
};
