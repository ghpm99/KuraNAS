import { useQuery } from '@tanstack/react-query';
import { AboutContext, AboutContextType } from './AboutContext';
import { apiBase } from '@/service';
import { useCallback, useEffect, useState } from 'react';
import { formatDuration } from '@/utils';
import useI18n from '@/components/i18n/provider/i18nContext';

const initialAboutContext: AboutContextType = {
	version: '',
	commit_hash: '',
	platform: '',
	enable_workers: false,
	gin_mode: '',
	lang: '',
	path: '',
	uptime: '',
	statup_time: new Date().toISOString(),
	gin_version: '',
	go_version: '',
	node_version: '',
};

export function AboutProvider({ children }: { children: React.ReactNode }) {
	const [currentTime, setCurrentTime] = useState(new Date());
	const { t } = useI18n();

	useEffect(() => {
		const timer = setInterval(() => {
			setCurrentTime(new Date());
		}, 1000);

		return () => clearInterval(timer);
	}, []);

	const { data } = useQuery({
		queryKey: ['about'],
		queryFn: async () => {
			const response = await apiBase.get<AboutContextType>('configuration/about');
			if (response.status !== 200) {
				throw new Error('Network response was not ok');
			}
			return response.data;
		},
		refetchOnWindowFocus: false,
	});

	const getCurrentUptime = useCallback((): string => {
		if (!data?.statup_time) {
			return t('LOADING');
		}
		const date = new Date(data.statup_time);
		const uptimeInSeconds = Math.floor((currentTime.getTime() - date.getTime()) / 1000);
		return formatDuration(uptimeInSeconds);
	}, [currentTime, data?.statup_time]);

	const value = {
		...(data || initialAboutContext),
		uptime: getCurrentUptime() || t('LOADING'),
	};

	return <AboutContext.Provider value={value}>{children}</AboutContext.Provider>;
}
