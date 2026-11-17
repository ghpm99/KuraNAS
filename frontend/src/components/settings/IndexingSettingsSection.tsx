import { Alert, Chip, FormControlLabel, Switch, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { UpdateSettingsConfigurationRequest } from '@/service/configuration';
import {
    isScanIntervalHoursValid,
    maxScanIntervalHours,
    minScanIntervalHours,
} from './scanIntervalHours';
import styles from './SettingsScreen.module.css';

type IndexingDraft = UpdateSettingsConfigurationRequest['indexing'];

type BooleanIndexingField = 'extract_metadata' | 'generate_previews' | 'index_document_text';

type IndexingSettingsSectionProps = {
    className?: string;
    indexing: IndexingDraft;
    isWorkersEnabled: boolean;
    isDisabled: boolean;
    onChange: <Key extends keyof IndexingDraft>(field: Key, value: IndexingDraft[Key]) => void;
};

const booleanPreferences: {
    field: BooleanIndexingField;
    labelKey: string;
    effectKey: string;
}[] = [
    {
        field: 'extract_metadata',
        labelKey: 'SETTINGS_INDEXING_EXTRACT_METADATA',
        effectKey: 'SETTINGS_INDEXING_EXTRACT_METADATA_EFFECT',
    },
    {
        field: 'generate_previews',
        labelKey: 'SETTINGS_INDEXING_GENERATE_PREVIEWS',
        effectKey: 'SETTINGS_INDEXING_GENERATE_PREVIEWS_EFFECT',
    },
    {
        field: 'index_document_text',
        labelKey: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT',
        effectKey: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT_EFFECT',
    },
];

const IndexingSettingsSection = ({
    className,
    indexing,
    isWorkersEnabled,
    isDisabled,
    onChange,
}: IndexingSettingsSectionProps) => {
    const { t } = useI18n();
    const isIntervalInvalid = !isScanIntervalHoursValid(indexing.scan_interval_hours);

    return (
        <section className={className}>
            <div className={styles.panelHeader}>
                <h2 className={styles.panelTitle}>{t('SETTINGS_SECTION_INDEXING')}</h2>
                <p className={styles.panelDescription}>
                    {t('SETTINGS_SECTION_INDEXING_DESCRIPTION')}
                </p>
            </div>
            <div className={styles.summary}>
                {booleanPreferences.map(({ field, labelKey }) => (
                    <Chip
                        key={field}
                        label={t(labelKey)}
                        variant={indexing[field] ? 'filled' : 'outlined'}
                    />
                ))}
            </div>
            {booleanPreferences.map(({ field, labelKey, effectKey }) => (
                <div key={field}>
                    <FormControlLabel
                        control={
                            <Switch
                                checked={indexing[field]}
                                onChange={(_, isChecked) => onChange(field, isChecked)}
                                disabled={isDisabled}
                            />
                        }
                        label={t(labelKey)}
                    />
                    <p className={styles.panelDescription}>{t(effectKey)}</p>
                </div>
            ))}
            <TextField
                fullWidth
                type="number"
                label={t('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS')}
                value={
                    Number.isNaN(indexing.scan_interval_hours) ? '' : indexing.scan_interval_hours
                }
                onChange={(event) =>
                    onChange(
                        'scan_interval_hours',
                        event.target.value === '' ? Number.NaN : Number(event.target.value)
                    )
                }
                disabled={isDisabled}
                error={isIntervalInvalid}
                helperText={
                    isIntervalInvalid
                        ? t('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS_INVALID', {
                              min: String(minScanIntervalHours),
                              max: String(maxScanIntervalHours),
                          })
                        : t('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS_EFFECT')
                }
                slotProps={{
                    htmlInput: { min: minScanIntervalHours, max: maxScanIntervalHours },
                }}
            />
            <Alert severity={isWorkersEnabled ? 'success' : 'warning'}>
                {isWorkersEnabled
                    ? t('SETTINGS_INDEXING_WORKERS_ON')
                    : t('SETTINGS_INDEXING_WORKERS_OFF')}
            </Alert>
        </section>
    );
};

export default IndexingSettingsSection;
