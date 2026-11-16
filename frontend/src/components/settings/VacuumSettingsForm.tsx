import { Alert, Button, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { SavedVacuumSettings } from '@/features/vacuum/toSavedVacuumSettings';
import { useVacuumSettingsForm } from '@/features/vacuum/useVacuumSettingsForm';
import VacuumConnectionActions from './VacuumConnectionActions';
import VacuumSettingsHelp from './VacuumSettingsHelp';
import styles from './VacuumSettingsSection.module.css';

type VacuumSettingsFormProps = {
    savedSettings: SavedVacuumSettings;
};

const VacuumSettingsForm = ({ savedSettings }: VacuumSettingsFormProps) => {
    const { t } = useI18n();
    const form = useVacuumSettingsForm(savedSettings);

    return (
        <>
            <VacuumSettingsHelp />
            <div className={styles.fields}>
                <TextField
                    size="small"
                    label={t('VACUUM_SETTINGS_DEVICE_ID_LABEL')}
                    value={form.deviceId}
                    onChange={(event) => form.setDeviceId(event.target.value)}
                    disabled={form.isSaving}
                    autoComplete="off"
                />
                <TextField
                    size="small"
                    type="password"
                    label={t('VACUUM_SETTINGS_LOCAL_KEY_LABEL')}
                    helperText={t(
                        form.hasStoredLocalKey
                            ? 'VACUUM_SETTINGS_LOCAL_KEY_STORED_HELP'
                            : 'VACUUM_SETTINGS_LOCAL_KEY_EMPTY_HELP'
                    )}
                    value={form.localKey}
                    onChange={(event) => form.setLocalKey(event.target.value)}
                    disabled={form.isSaving}
                    autoComplete="new-password"
                />
                <TextField
                    size="small"
                    label={t('VACUUM_SETTINGS_ADDRESS_LABEL')}
                    helperText={t('VACUUM_SETTINGS_ADDRESS_HELP')}
                    value={form.address}
                    onChange={(event) => form.setAddress(event.target.value)}
                    disabled={form.isSaving}
                    autoComplete="off"
                />
            </div>
            {form.saveFailureMessage ? (
                <Alert severity="error">{form.saveFailureMessage}</Alert>
            ) : null}
            <div className={styles.actions}>
                <Button
                    variant="contained"
                    onClick={form.save}
                    disabled={form.isSaving || !form.hasUnsavedChanges}
                >
                    {form.isSaving ? t('SAVING') : t('VACUUM_SETTINGS_SAVE')}
                </Button>
            </div>
            <VacuumConnectionActions blockedMessageKey={form.connectionBlockedMessageKey} />
        </>
    );
};

export default VacuumSettingsForm;
