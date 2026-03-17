import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useAbout } from '@/components/providers/aboutProvider/AboutContext';
import { applyUpdate, getUpdateStatus } from '@/service/update';
import { UpdateStatus } from '@/types/update';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

type AboutDetail = {
    label: string;
    value: string;
};

type AboutTool = {
    title: string;
    description: string;
    href: string;
};

const COPY_FEEDBACK_TIMEOUT_MS = 2000;

const formatTechnicalTimestamp = (value: string) => {
    if (!value) {
        return '-';
    }

    const parsedDate = new Date(value);
    if (Number.isNaN(parsedDate.getTime())) {
        return value;
    }

    return parsedDate.toISOString();
};

export const useAboutScreen = () => {
    const about = useAbout();
    const { t } = useI18n();
    const [copied, setCopied] = useState(false);
    const copyFeedbackTimeoutRef = useRef<number | null>(null);

    useEffect(() => {
        return () => {
            if (copyFeedbackTimeoutRef.current !== null) {
                window.clearTimeout(copyFeedbackTimeoutRef.current);
            }
        };
    }, []);

    const runtimeDetails = useMemo<AboutDetail[]>(
        () => [
            { label: t('ABOUT_RUNTIME_VERSION'), value: about.version || '-' },
            { label: t('ABOUT_RUNTIME_ROOT'), value: about.path || '-' },
            {
                label: t('ABOUT_RUNTIME_WORKERS'),
                value: about.enable_workers
                    ? t('SETTINGS_STATUS_ENABLED')
                    : t('SETTINGS_STATUS_DISABLED'),
            },
            { label: t('ABOUT_RUNTIME_UPTIME'), value: about.uptime || t('LOADING') },
            {
                label: t('ABOUT_RUNTIME_STARTED_AT'),
                value: formatTechnicalTimestamp(about.statup_time),
            },
        ],
        [about.enable_workers, about.path, about.statup_time, about.uptime, about.version, t]
    );

    const buildDetails = useMemo<AboutDetail[]>(
        () => [
            { label: t('ABOUT_BUILD_COMMIT'), value: about.commit_hash || '-' },
            { label: t('ABOUT_BUILD_PLATFORM'), value: about.platform || '-' },
            { label: t('ABOUT_BUILD_LANGUAGE'), value: about.lang || '-' },
            { label: t('ABOUT_BUILD_MODE'), value: about.gin_mode || '-' },
            { label: t('ABOUT_BUILD_BACKEND'), value: about.gin_version || '-' },
            { label: t('ABOUT_BUILD_GO'), value: about.go_version || '-' },
            { label: t('ABOUT_BUILD_NODE'), value: about.node_version || '-' },
        ],
        [
            about.commit_hash,
            about.gin_mode,
            about.gin_version,
            about.go_version,
            about.lang,
            about.node_version,
            about.platform,
            t,
        ]
    );

    const tools = useMemo<AboutTool[]>(
        () => [
            {
                title: t('ABOUT_TOOL_ANALYTICS_TITLE'),
                description: t('ABOUT_TOOL_ANALYTICS_DESCRIPTION'),
                href: appRoutes.analytics,
            },
            {
                title: t('ABOUT_TOOL_SETTINGS_TITLE'),
                description: t('ABOUT_TOOL_SETTINGS_DESCRIPTION'),
                href: appRoutes.settings,
            },
        ],
        [t]
    );

    const { enqueueSnackbar } = useSnackbar();

    const {
        data: updateStatus,
        isLoading: isCheckingUpdate,
        isError: isUpdateError,
    } = useQuery<UpdateStatus>({
        queryKey: ['updateStatus'],
        queryFn: getUpdateStatus,
        refetchOnWindowFocus: false,
    });

    const { mutate: triggerUpdate, isPending: isApplyingUpdate } = useMutation({
        mutationFn: applyUpdate,
        onSuccess: () => {
            enqueueSnackbar(t('UPDATE_APPLIED'), { variant: 'success' });
        },
        onError: () => {
            enqueueSnackbar(t('ERROR_UPDATE_APPLY'), { variant: 'error' });
        },
    });

    const formatFileSize = (bytes: number): string => {
        if (bytes <= 0) return '-';
        const units = ['B', 'KB', 'MB', 'GB'];
        const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
        return `${(bytes / Math.pow(1024, index)).toFixed(1)} ${units[index]}`;
    };

    const updateDetails = useMemo<AboutDetail[]>(() => {
        if (!updateStatus) return [];
        return [
            { label: t('ABOUT_UPDATE_CURRENT'), value: updateStatus.current_version || '-' },
            { label: t('ABOUT_UPDATE_LATEST'), value: updateStatus.latest_version || '-' },
            { label: t('ABOUT_UPDATE_RELEASED'), value: updateStatus.release_date || '-' },
            { label: t('ABOUT_UPDATE_ASSET'), value: updateStatus.asset_name || '-' },
            { label: t('ABOUT_UPDATE_SIZE'), value: formatFileSize(updateStatus.asset_size) },
        ];
    }, [updateStatus, t]);

    const copyCommitHash = useCallback(async () => {
        if (!about.commit_hash) {
            return;
        }

        try {
            await navigator.clipboard.writeText(about.commit_hash);
            setCopied(true);
            if (copyFeedbackTimeoutRef.current !== null) {
                window.clearTimeout(copyFeedbackTimeoutRef.current);
            }
            copyFeedbackTimeoutRef.current = window.setTimeout(() => {
                setCopied(false);
                copyFeedbackTimeoutRef.current = null;
            }, COPY_FEEDBACK_TIMEOUT_MS);
        } catch {
            setCopied(false);
        }
    }, [about.commit_hash]);

    return {
        t,
        version: about.version || '-',
        workersEnabled: about.enable_workers,
        runtimeDetails,
        buildDetails,
        tools,
        copied,
        copyCommitHash,
        updateStatus,
        updateDetails,
        isCheckingUpdate,
        isUpdateError,
        isApplyingUpdate,
        triggerUpdate,
    };
};
