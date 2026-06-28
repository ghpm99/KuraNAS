import {
    Alert,
    Button,
    Checkbox,
    Chip,
    CircularProgress,
    FormControlLabel,
    Step,
    StepLabel,
    Stepper,
    Switch,
    TextField,
} from '@mui/material';
import { Link } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import { EnvField } from '@/service/configuration';
import useConfigWizard, { generateTokenKey } from './useConfigWizard';
import styles from './ConfigWizardScreen.module.css';

const PATH_TEST_KEYS = ['ENTRY_POINT', 'YTDLP_PATH'];

const ConfigWizardScreen = () => {
    const {
        t,
        isLoading,
        isError,
        restartRequired,
        steps,
        activeStep,
        fieldsByGroup,
        fieldValue,
        setField,
        confirmed,
        setConfirmed,
        changedDangerous,
        pendingChanges,
        fieldMap,
        isLastStep,
        goNext,
        goBack,
        dbTestResult,
        pathTestResult,
        runDbTest,
        runPathTest,
        save,
        isSaving,
    } = useConfigWizard();

    const renderField = (field: EnvField) => {
        const label = t(`ENV_FIELD_${field.key}`);
        const value = fieldValue(field);

        if (field.kind === 'bool') {
            return (
                <FormControlLabel
                    key={field.key}
                    control={
                        <Switch
                            checked={value === 'true'}
                            onChange={(_, checked) =>
                                setField(field.key, checked ? 'true' : 'false')
                            }
                        />
                    }
                    label={label}
                />
            );
        }

        if (field.kind === 'secret') {
            return (
                <div key={field.key} className={styles.field}>
                    <TextField
                        fullWidth
                        type="password"
                        label={label}
                        value={value}
                        placeholder={t('ENV_SECRET_PLACEHOLDER')}
                        onChange={(event) => setField(field.key, event.target.value)}
                    />
                    <div className={styles.fieldMeta}>
                        <Chip
                            size="small"
                            label={
                                field.configured
                                    ? t('ENV_SECRET_CONFIGURED')
                                    : t('ENV_SECRET_EMPTY')
                            }
                            color={field.configured ? 'success' : 'default'}
                            variant="outlined"
                        />
                        {field.key === 'EMAIL_TOKEN_KEY' ? (
                            <Button
                                size="small"
                                onClick={() => setField(field.key, generateTokenKey())}
                            >
                                {t('ENV_GENERATE_TOKEN_KEY')}
                            </Button>
                        ) : null}
                    </div>
                    {field.key === 'EMAIL_TOKEN_KEY' ? (
                        <Alert severity="warning">{t('ENV_TOKEN_KEY_WARNING')}</Alert>
                    ) : null}
                </div>
            );
        }

        return (
            <div key={field.key} className={styles.field}>
                <TextField
                    fullWidth
                    type={field.kind === 'int' ? 'number' : 'text'}
                    label={label}
                    value={value}
                    onChange={(event) => setField(field.key, event.target.value)}
                />
                {PATH_TEST_KEYS.includes(field.key) ? (
                    <Button size="small" onClick={() => void runPathTest(field.key)}>
                        {t('ENV_TEST_PATH_BUTTON')}
                    </Button>
                ) : null}
            </div>
        );
    };

    const renderReview = () => {
        if (pendingChanges.length === 0) {
            return <p className={styles.hint}>{t('ENV_REVIEW_NO_CHANGES')}</p>;
        }

        return (
            <ul className={styles.reviewList}>
                {pendingChanges.map((key) => {
                    const field = fieldMap[key];
                    if (!field) {
                        return null;
                    }
                    const display = field.kind === 'secret' ? '••••••' : fieldValue(field);
                    return (
                        <li key={key} className={styles.reviewItem}>
                            <span className={styles.reviewKey}>{t(`ENV_FIELD_${key}`)}</span>
                            <span className={styles.reviewValue}>{display}</span>
                        </li>
                    );
                })}
            </ul>
        );
    };

    if (isLoading) {
        return (
            <div className={styles.center}>
                <CircularProgress />
            </div>
        );
    }

    if (isError) {
        return (
            <div className={styles.page}>
                <Alert severity="warning">{t('ENV_WIZARD_LOAD_ERROR')}</Alert>
                <p className={styles.hint}>{t('ENV_WIZARD_LOOPBACK_HELP')}</p>
                <Button component={Link} to={appRoutes.about} variant="text">
                    {t('ABOUT')}
                </Button>
            </div>
        );
    }

    const currentStep = steps[activeStep];
    if (!currentStep) {
        return null;
    }
    const isReview = currentStep.group === 'review';

    return (
        <div className={styles.page}>
            <header className={styles.header}>
                <h1 className={styles.title}>{t('ENV_WIZARD_TITLE')}</h1>
                <p className={styles.description}>{t('ENV_WIZARD_DESCRIPTION')}</p>
            </header>

            {restartRequired ? <Alert severity="info">{t('ENV_RESTART_REQUIRED')}</Alert> : null}

            <Stepper activeStep={activeStep} alternativeLabel className={styles.stepper}>
                {steps.map((step) => (
                    <Step key={step.group}>
                        <StepLabel>{step.title}</StepLabel>
                    </Step>
                ))}
            </Stepper>

            <section className={styles.stepContent}>
                {isReview ? (
                    <>
                        <p className={styles.description}>{t('ENV_REVIEW_DESCRIPTION')}</p>
                        {renderReview()}
                    </>
                ) : (
                    (fieldsByGroup[currentStep.group] ?? []).map(renderField)
                )}

                {currentStep.group === 'database' ? (
                    <div className={styles.testRow}>
                        <Button variant="outlined" onClick={() => void runDbTest()}>
                            {t('ENV_TEST_DB_BUTTON')}
                        </Button>
                        {dbTestResult ? (
                            <Alert severity={dbTestResult.ok ? 'success' : 'error'}>
                                {dbTestResult.message}
                            </Alert>
                        ) : null}
                    </div>
                ) : null}

                {pathTestResult ? (
                    <Alert severity={pathTestResult.ok ? 'success' : 'error'}>
                        {pathTestResult.message}
                    </Alert>
                ) : null}

                {changedDangerous ? (
                    <FormControlLabel
                        control={
                            <Checkbox
                                checked={confirmed}
                                onChange={(_, checked) => setConfirmed(checked)}
                            />
                        }
                        label={t('ENV_CONFIRM_DANGEROUS')}
                    />
                ) : null}
            </section>

            <footer className={styles.footer}>
                <Button variant="text" disabled={activeStep === 0} onClick={goBack}>
                    {t('ENV_BACK')}
                </Button>
                {isLastStep ? (
                    <Button
                        variant="contained"
                        onClick={save}
                        disabled={
                            isSaving ||
                            pendingChanges.length === 0 ||
                            (changedDangerous && !confirmed)
                        }
                    >
                        {isSaving ? t('SAVING') : t('ENV_SAVE')}
                    </Button>
                ) : (
                    <Button variant="contained" onClick={goNext}>
                        {t('ENV_NEXT')}
                    </Button>
                )}
            </footer>
        </div>
    );
};

export default ConfigWizardScreen;
