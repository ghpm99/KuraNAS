import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import useAutoShutdownSettings from './useAutoShutdownSettings';
import { computeNextShutdownExecution } from './nextShutdownExecution';
import { maxGracePeriodSeconds, minGracePeriodSeconds } from './autoShutdownGraceRange';
import styles from './AutoShutdownSettingsSection.module.css';

type AutoShutdownSettingsSectionProps = {
	className?: string;
};

const AutoShutdownSettingsSection = ({ className = '' }: AutoShutdownSettingsSectionProps) => {
	const {
		t,
		form,
		suggestion,
		isLoading,
		isSaving,
		isSuggesting,
		hasError,
		hasUnsavedChanges,
		isGracePeriodInvalid,
		setField,
		handleSave,
		handleSuggest,
	} = useAutoShutdownSettings();

	const nextExecution = form.enabled ? computeNextShutdownExecution(form.time, new Date()) : null;
	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('SETTINGS_AUTO_SHUTDOWN_TITLE')}</h2>
					<p className={styles.description}>{t('SETTINGS_AUTO_SHUTDOWN_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('SETTINGS_AUTO_SHUTDOWN_TITLE')}</h2>
				<p className={styles.description}>{t('SETTINGS_AUTO_SHUTDOWN_DESCRIPTION')}</p>
			</div>
			{hasError ? (
				<Alert severity="error">{t('SETTINGS_AUTO_SHUTDOWN_LOAD_ERROR')}</Alert>
			) : null}
			<FormControlLabel
				control={
					<Switch
						checked={form.enabled}
						onChange={(_, checked) => setField('enabled', checked)}
						disabled={isSaving}
					/>
				}
				label={t('SETTINGS_AUTO_SHUTDOWN_ENABLE_LABEL')}
			/>
			<div className={styles.fields}>
				<TextField
					size="small"
					type="time"
					label={t('SETTINGS_AUTO_SHUTDOWN_TIME_LABEL')}
					value={form.time}
					onChange={(event) => setField('time', event.target.value)}
					disabled={isSaving || !form.enabled}
					InputLabelProps={{ shrink: true }}
				/>
				<TextField
					size="small"
					type="number"
					label={t('SETTINGS_AUTO_SHUTDOWN_GRACE_LABEL')}
					error={isGracePeriodInvalid}
					helperText={
						isGracePeriodInvalid
							? t('SETTINGS_AUTO_SHUTDOWN_GRACE_INVALID', {
									min: String(minGracePeriodSeconds),
									max: String(maxGracePeriodSeconds),
								})
							: t('SETTINGS_AUTO_SHUTDOWN_GRACE_HELPER')
					}
					slotProps={{
						htmlInput: { min: minGracePeriodSeconds, max: maxGracePeriodSeconds },
					}}
					value={String(form.grace_period_seconds)}
					onChange={(event) =>
						setField('grace_period_seconds', Number(event.target.value))
					}
					disabled={isSaving || !form.enabled}
				/>
			</div>
			<p className={styles.description}>{t('SETTINGS_AUTO_SHUTDOWN_WARNING_NOTICE')}</p>
			<p className={styles.description}>
				{nextExecution
					? t('SETTINGS_AUTO_SHUTDOWN_NEXT_RUN', {
							dateTime: nextExecution.toLocaleString(),
						})
					: t('SETTINGS_AUTO_SHUTDOWN_NEXT_RUN_NONE')}
			</p>
			<div className={styles.suggestRow}>
				<Button
					variant="outlined"
					onClick={() => void handleSuggest()}
					disabled={isSaving || isSuggesting}
				>
					{t('SETTINGS_AUTO_SHUTDOWN_SUGGEST_BUTTON')}
				</Button>
				{suggestion ? (
					<p className={styles.suggestHint}>
						{suggestion.available
							? t('SETTINGS_AUTO_SHUTDOWN_SUGGESTION', {
									time: suggestion.time,
									count: String(suggestion.sample_size),
								})
							: t('SETTINGS_AUTO_SHUTDOWN_SUGGESTION_EMPTY')}
					</p>
				) : null}
			</div>
			<div className={styles.actions}>
				<Button
					variant="contained"
					onClick={() => void handleSave()}
					disabled={isSaving || !hasUnsavedChanges || isGracePeriodInvalid}
				>
					{t('SETTINGS_AUTO_SHUTDOWN_SAVE')}
				</Button>
			</div>
		</section>
	);
};

export default AutoShutdownSettingsSection;
