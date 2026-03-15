import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useAbout } from '@/components/providers/aboutProvider/AboutContext';
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
				value: about.enable_workers ? t('SETTINGS_STATUS_ENABLED') : t('SETTINGS_STATUS_DISABLED'),
			},
			{ label: t('ABOUT_RUNTIME_UPTIME'), value: about.uptime || t('LOADING') },
			{ label: t('ABOUT_RUNTIME_STARTED_AT'), value: formatTechnicalTimestamp(about.statup_time) },
		],
		[about.enable_workers, about.path, about.statup_time, about.uptime, about.version, t],
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
		[about.commit_hash, about.gin_mode, about.gin_version, about.go_version, about.lang, about.node_version, about.platform, t],
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
			{
				title: t('ABOUT_TOOL_DIARY_TITLE'),
				description: t('ABOUT_TOOL_DIARY_DESCRIPTION'),
				href: appRoutes.activityDiary,
			},
		],
		[t],
	);

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
	};
};
