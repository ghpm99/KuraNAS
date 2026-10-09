import useI18n from '@/components/i18n/provider/i18nContext';
import {
    EnvConfig,
    EnvField,
    EnvTestResult,
    getEnvConfig,
    testEnvDatabase,
    testEnvPath,
    updateEnvConfig,
} from '@/service/configuration';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useMemo, useState } from 'react';

export const ENV_STEP_GROUPS = ['general', 'database', 'access', 'email', 'ai', 'workers'] as const;

const STEP_TITLE_KEYS: Record<string, string> = {
    general: 'ENV_STEP_GENERAL',
    database: 'ENV_STEP_DATABASE',
    access: 'ENV_STEP_ACCESS',
    email: 'ENV_STEP_EMAIL',
    ai: 'ENV_STEP_AI',
    workers: 'ENV_STEP_WORKERS',
};

const ENV_QUERY_KEY = ['envConfig'];

// generateTokenKey builds a base64-encoded 32-byte key for EMAIL_TOKEN_KEY,
// matching the AES-256 contract the backend validates.
export const generateTokenKey = (): string => {
    const bytes = new Uint8Array(32);
    crypto.getRandomValues(bytes);
    let binary = '';
    bytes.forEach((value) => {
        binary += String.fromCharCode(value);
    });
    return btoa(binary);
};

// resolveBackendError renders the already-translated server message verbatim and
// falls back to a generic client key only when the response carries none.
const resolveBackendError = (error: unknown, fallback: string): string => {
    const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error;
    return message ?? fallback;
};

export const useConfigWizard = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const { data, isLoading, isError } = useQuery<EnvConfig>({
        queryKey: ENV_QUERY_KEY,
        queryFn: getEnvConfig,
        retry: false,
        refetchOnWindowFocus: false,
    });

    const [activeStep, setActiveStep] = useState(0);
    const [changes, setChanges] = useState<Record<string, string>>({});
    const [confirmed, setConfirmed] = useState(false);
    const [dbTestResult, setDbTestResult] = useState<EnvTestResult | null>(null);
    const [pathTestResult, setPathTestResult] = useState<EnvTestResult | null>(null);

    const fields = useMemo(() => data?.fields ?? [], [data]);

    const fieldMap = useMemo(() => {
        const map: Record<string, EnvField> = {};
        fields.forEach((field) => {
            map[field.key] = field;
        });
        return map;
    }, [fields]);

    const steps = useMemo(
        () => [
            ...ENV_STEP_GROUPS.map((group) => ({ group, title: t(STEP_TITLE_KEYS[group] ?? group) })),
            { group: 'review', title: t('ENV_STEP_REVIEW') },
        ],
        [t]
    );

    const fieldsByGroup = useMemo(() => {
        const map: Record<string, EnvField[]> = {};
        fields.forEach((field) => {
            const list = map[field.group] ?? [];
            list.push(field);
            map[field.group] = list;
        });
        return map;
    }, [fields]);

    const setField = (key: string, value: string) => {
        setChanges((previous) => ({ ...previous, [key]: value }));
    };

    const fieldValue = (field: EnvField): string => {
        const change = changes[field.key];
        if (change !== undefined) {
            return change;
        }
        return field.kind === 'secret' ? '' : field.value;
    };

    const valueForKey = (key: string): string => {
        const field = fieldMap[key];
        if (!field) {
            return changes[key] ?? '';
        }
        return fieldValue(field);
    };

    const pendingChanges = useMemo(
        () =>
            Object.keys(changes).filter((key) => {
                const field = fieldMap[key];
                if (!field) {
                    return false;
                }
                if (field.kind === 'secret') {
                    return changes[key] !== '';
                }
                return changes[key] !== field.value;
            }),
        [changes, fieldMap]
    );

    const changedDangerous = useMemo(
        () => pendingChanges.some((key) => fieldMap[key]?.dangerous),
        [pendingChanges, fieldMap]
    );

    const isLastStep = activeStep === steps.length - 1;
    const goNext = () => setActiveStep((step) => Math.min(step + 1, steps.length - 1));
    const goBack = () => setActiveStep((step) => Math.max(step - 1, 0));

    const updateMutation = useMutation({
        mutationFn: () => updateEnvConfig({ changes, confirmed }),
        onSuccess: (config) => {
            queryClient.setQueryData(ENV_QUERY_KEY, config);
            setChanges({});
            setConfirmed(false);
            enqueueSnackbar(t('ENV_RESTART_REQUIRED'), { variant: 'warning' });
        },
        onError: (error) => {
            enqueueSnackbar(resolveBackendError(error, t('ERROR_ENV_UPDATE_FAILED')), {
                variant: 'error',
            });
        },
    });

    const runDbTest = async () => {
        const result = await testEnvDatabase({
            host: valueForKey('DB_HOST'),
            port: valueForKey('DB_PORT'),
            user: valueForKey('DB_USER'),
            name: valueForKey('DB_NAME'),
            password: changes['DB_PASSWORD'] ?? '',
        });
        setDbTestResult(result);
    };

    const runPathTest = async (key: string) => {
        const result = await testEnvPath({ path: valueForKey(key) });
        setPathTestResult(result);
    };

    const save = () => {
        if (pendingChanges.length === 0) {
            return;
        }
        updateMutation.mutate();
    };

    return {
        t,
        isLoading,
        isError,
        restartRequired: data?.restart_required ?? false,
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
        isSaving: updateMutation.isPending,
    };
};

export default useConfigWizard;
