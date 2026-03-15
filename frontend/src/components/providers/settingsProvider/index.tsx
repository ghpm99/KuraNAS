import {
	getSettingsConfiguration,
	type SettingsConfiguration,
	type UpdateSettingsConfigurationRequest,
	updateSettingsConfiguration,
} from '@/service/configuration';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo } from 'react';
import { defaultSettingsConfiguration, SettingsContextProvider } from './settingsContext';

const accentPalette: Record<SettingsConfiguration['appearance']['accent_color'], { primary: string; hover: string; glow: string }> = {
	violet: {
		primary: '#6D5DF6',
		hover: '#7C70FF',
		glow: '0 0 0 1px rgba(109, 93, 246, 0.6), 0 8px 30px rgba(109, 93, 246, 0.18)',
	},
	cyan: {
		primary: '#06B6D4',
		hover: '#22D3EE',
		glow: '0 0 0 1px rgba(6, 182, 212, 0.58), 0 8px 30px rgba(6, 182, 212, 0.18)',
	},
	rose: {
		primary: '#E11D48',
		hover: '#FB7185',
		glow: '0 0 0 1px rgba(225, 29, 72, 0.58), 0 8px 30px rgba(225, 29, 72, 0.18)',
	},
};

const applyAppearanceSettings = (settings: SettingsConfiguration) => {
	const root = document.documentElement;
	const accent = accentPalette[settings.appearance.accent_color] ?? accentPalette.violet;

	root.style.setProperty('--app-color-primary', accent.primary);
	root.style.setProperty('--app-color-primary-hover', accent.hover);
	root.style.setProperty('--app-shadow-active-primary', accent.glow);

	if (settings.appearance.reduce_motion) {
		root.dataset.appMotion = 'reduced';
		return;
	}

	delete root.dataset.appMotion;
};

export const SettingsProvider = ({ children }: { children: React.ReactNode }) => {
	const queryClient = useQueryClient();
	const settingsQuery = useQuery({
		queryKey: ['configuration', 'settings'],
		queryFn: getSettingsConfiguration,
		retry: false,
	});

	const saveMutation = useMutation({
		mutationFn: (request: UpdateSettingsConfigurationRequest) => updateSettingsConfiguration(request),
		onSuccess: async (settings) => {
			queryClient.setQueryData(['configuration', 'settings'], settings);
			await queryClient.invalidateQueries({ queryKey: ['configuration'] });
		},
	});

	const settings = settingsQuery.data ?? defaultSettingsConfiguration;
	const { isLoading, isError, refetch } = settingsQuery;
	const { isPending, mutateAsync } = saveMutation;

	useEffect(() => {
		applyAppearanceSettings(settings);
	}, [settings]);

	const value = useMemo(
		() => ({
			settings,
			isLoading,
			isSaving: isPending,
			hasError: isError,
			refresh: async () => {
				await refetch();
			},
			saveSettings: async (request: UpdateSettingsConfigurationRequest) => mutateAsync(request),
		}),
		[isError, isLoading, isPending, mutateAsync, refetch, settings],
	);

	return <SettingsContextProvider value={value}>{children}</SettingsContextProvider>;
};

export default SettingsProvider;
