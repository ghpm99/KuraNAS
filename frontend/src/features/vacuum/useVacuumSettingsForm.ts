import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRef, useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { updateVacuumSettings } from '@/service/vacuum';
import type { VacuumSettingsChanges } from '@/types/vacuum';
import { invalidateVacuumQueries } from './invalidateVacuumQueries';
import type { SavedVacuumSettings } from './toSavedVacuumSettings';
import { useVacuumFeedback } from './useVacuumFeedback';
import { resolveVacuumFailureMessage } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumSettingsForm = (savedSettings: SavedVacuumSettings) => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const { notifySuccess } = useVacuumFeedback();
    const [editedAddress, setEditedAddress] = useState<string | undefined>();
    const [editedDeviceId, setEditedDeviceId] = useState<string | undefined>();
    const [localKey, setLocalKey] = useState('');
    const isSaveInFlightRef = useRef(false);

    const address = editedAddress ?? savedSettings.address;
    const deviceId = editedDeviceId ?? savedSettings.deviceId;
    const hasUnsavedChanges =
        address !== savedSettings.address ||
        deviceId !== savedSettings.deviceId ||
        localKey !== '';

    const saveMutation = useMutation({
        mutationFn: (changes: VacuumSettingsChanges) => updateVacuumSettings(changes),
        onSuccess: (confirmedSettings) => {
            queryClient.setQueryData(vacuumQueryKeys.settings, confirmedSettings);
            setEditedAddress(undefined);
            setEditedDeviceId(undefined);
            setLocalKey('');
            notifySuccess('VACUUM_SETTINGS_SAVED');
            void invalidateVacuumQueries(queryClient);
        },
        onSettled: () => {
            isSaveInFlightRef.current = false;
        },
    });

    const save = () => {
        if (isSaveInFlightRef.current) return;
        isSaveInFlightRef.current = true;
        saveMutation.mutate({ address, device_id: deviceId, local_key: localKey });
    };

    const setAddress = (typedAddress: string) =>
        setEditedAddress(typedAddress === savedSettings.address ? undefined : typedAddress);

    const setDeviceId = (typedDeviceId: string) =>
        setEditedDeviceId(typedDeviceId === savedSettings.deviceId ? undefined : typedDeviceId);

    const saveFailureMessage = saveMutation.isError
        ? resolveVacuumFailureMessage(saveMutation.error, t('CONNECTION_BANNER_UNAVAILABLE'))
        : undefined;

    const connectionBlockedMessageKey = resolveConnectionBlockedMessageKey(
        savedSettings,
        hasUnsavedChanges
    );

    return {
        address,
        deviceId,
        localKey,
        hasStoredLocalKey: savedSettings.hasLocalKey,
        hasUnsavedChanges,
        isSaving: saveMutation.isPending,
        saveFailureMessage,
        connectionBlockedMessageKey,
        setAddress,
        setDeviceId,
        setLocalKey,
        save,
    };
};

const resolveConnectionBlockedMessageKey = (
    savedSettings: SavedVacuumSettings,
    hasUnsavedChanges: boolean
): string | undefined => {
    if (savedSettings.deviceId === '' || !savedSettings.hasLocalKey) {
        return 'VACUUM_SETTINGS_CONNECTION_NEEDS_SAVED_CREDENTIALS';
    }
    if (hasUnsavedChanges) return 'VACUUM_SETTINGS_CONNECTION_SAVE_FIRST';
    return undefined;
};
