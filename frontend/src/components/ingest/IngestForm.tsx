import { Alert, Button, MenuItem, TextField, Typography } from '@mui/material';
import FolderPicker from '@/components/folderPicker/folderPicker';
import styles from './IngestForm.module.css';
import { useIngestForm } from './useIngestForm';

const IngestForm = () => {
    const form = useIngestForm();
    const { t } = form;

    return (
        <form
            className={styles.form}
            noValidate
            onSubmit={(event) => {
                event.preventDefault();
                form.submit();
            }}
        >
            <div className={styles.row}>
                <TextField
                    label={t('INGEST_FORM_URL_LABEL')}
                    placeholder={t('INGEST_FORM_URL_PLACEHOLDER')}
                    value={form.url}
                    onChange={(event) => form.changeUrl(event.target.value)}
                    error={form.urlError !== null}
                    helperText={form.urlError ?? undefined}
                    fullWidth
                />
                {form.isClipboardAvailable && (
                    <Button variant="outlined" onClick={form.pasteUrlFromClipboard}>
                        {t('INGEST_FORM_PASTE')}
                    </Button>
                )}
            </div>

            <TextField
                select
                label={t('INGEST_FORM_TARGET_ROOT_LABEL')}
                value={form.targetRoot}
                onChange={(event) => form.changeTargetRoot(event.target.value)}
                fullWidth
            >
                {form.targets.map((target) => (
                    <MenuItem key={target.path} value={target.path}>
                        {target.label}
                    </MenuItem>
                ))}
            </TextField>

            <div className={styles.subfolder}>
                <Button
                    variant="outlined"
                    onClick={form.openFolderPicker}
                    disabled={form.targetRoot === ''}
                >
                    {t('INGEST_FORM_CHOOSE_SUBFOLDER')}
                </Button>
                <Typography
                    variant="body2"
                    color={form.subfolderError ? 'error' : 'text.secondary'}
                >
                    {form.subfolderError ?? (form.subfolder || t('INGEST_FORM_SUBFOLDER_ROOT'))}
                </Typography>
            </div>

            <TextField
                select
                label={t('INGEST_FORM_PRESET_LABEL')}
                value={form.preset}
                onChange={(event) => form.changePreset(event.target.value)}
                fullWidth
            >
                {form.presets.map((preset) => (
                    <MenuItem key={preset.key} value={preset.key}>
                        {preset.label}
                    </MenuItem>
                ))}
            </TextField>

            {form.serverError && <Alert severity="error">{form.serverError}</Alert>}

            <div>
                <Button type="submit" variant="contained" disabled={!form.canSubmit}>
                    {t('INGEST_FORM_SUBMIT')}
                </Button>
            </div>

            <FolderPicker
                open={form.isFolderPickerOpen}
                onClose={form.closeFolderPicker}
                onSelect={form.selectFolder}
            />
        </form>
    );
};

export default IngestForm;
