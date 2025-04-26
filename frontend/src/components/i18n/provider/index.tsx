import { apiBase } from '@/service';
import { useQuery } from '@tanstack/react-query';
import { I18nContextProvider, I18nContextType } from './i18nContext';

const I18nProvider = ({ children }: { children: React.ReactNode }) => {
	const { status, data } = useQuery({
		queryKey: ['configuration'],
		queryFn: async () => {
			const response = await apiBase.get(`/configuration/translation`);
			return response.data;
		},
	});

	const t = (key: string, options?: Record<string, string>): string => {
		if (status !== 'success' && !data) return key;

		const translation = data[key];
		if (translation) {
			return Object.entries(options || {}).reduce((acc, [k, v]) => {
				return acc.replace(`{{${k}}}`, v);
			}, translation);
		}

		return key;
	};
	console.log('status', status);
	console.log('data', data);

	const contextValue: I18nContextType = {
		t,
	};
	return <I18nContextProvider value={contextValue}>{children}</I18nContextProvider>;
};

export default I18nProvider;
