import { Alert, Button, MenuItem, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { SavedVacuumCloudSettings } from '@/features/vacuum/toSavedVacuumCloudSettings';
import { useVacuumCloudSettingsForm } from '@/features/vacuum/useVacuumCloudSettingsForm';
import { vacuumCloudRegionOptions } from '@/features/vacuum/vacuumCloudRegions';
import VacuumCloudSettingsHelp from './VacuumCloudSettingsHelp';
import VacuumCloudTestAction from './VacuumCloudTestAction';
import styles from './VacuumSettingsSection.module.css';

type VacuumCloudSettingsFormProps = {
    savedSettings: SavedVacuumCloudSettings;
};

const VacuumCloudSettingsForm = ({ savedSettings }: VacuumCloudSettingsFormProps) => {
    const { t } = useI18n();
    const form = useVacuumCloudSettingsForm(savedSettings);
    const isRegionKnown = vacuumCloudRegionOptions.some(({ region }) => region === form.region);

    return (
        <>
            <VacuumCloudSettingsHelp />
            <div className={styles.fields}>
                <TextField
                    label={t('VACUUM_CLOUD_SETTINGS_ACCESS_ID_LABEL')}
                    value={form.accessId}
                    onChange={(event) => form.setAccessId(event.target.value)}
                    disabled={form.isSaving}
                    autoComplete="off"
                />
                <TextField
                    type="password"
                    label={t('VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_LABEL')}
                    helperText={t(
                        form.hasStoredAccessSecret
                            ? 'VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_STORED_HELP'
                            : 'VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_EMPTY_HELP'
                    )}
                    value={form.accessSecret}
                    onChange={(event) => form.setAccessSecret(event.target.value)}
                    disabled={form.isSaving}
                    autoComplete="new-password"
                />
                <TextField
                    select
                    label={t('VACUUM_CLOUD_SETTINGS_REGION_LABEL')}
                    value={isRegionKnown ? form.region : ''}
                    onChange={(event) => form.setRegion(event.target.value)}
                    disabled={form.isSaving}
                >
                    {vacuumCloudRegionOptions.map(({ region, labelKey }) => (
                        <MenuItem key={region} value={region} className={styles.touchTarget}>
                            {t(labelKey)}
                        </MenuItem>
                    ))}
                </TextField>
            </div>
            {form.saveFailureMessage ? (
                <Alert severity="error">{form.saveFailureMessage}</Alert>
            ) : null}
            <div className={styles.actions}>
                <Button
                    variant="contained"
                    className={styles.touchTarget}
                    onClick={form.save}
                    disabled={form.isSaving || !form.hasUnsavedChanges}
                >
                    {form.isSaving ? t('SAVING') : t('VACUUM_CLOUD_SETTINGS_SAVE')}
                </Button>
            </div>
            <VacuumCloudTestAction blockedMessageKey={form.testBlockedMessageKey} />
        </>
    );
};

export default VacuumCloudSettingsForm;
