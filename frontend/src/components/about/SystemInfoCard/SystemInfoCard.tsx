import Card from '@/components/ui/Card/Card';
import { Box, Typography } from '@mui/material';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ReactNode } from 'react';

function InfoRow({ label, value }: { label: string; value: ReactNode }) {
	return (
		<Box sx={{ display: 'flex', justifyContent: 'space-between', py: 0.75, borderBottom: 1, borderColor: 'divider' }}>
			<Typography variant='body2' color='text.secondary'>{label}</Typography>
			<Typography variant='body2'>{value}</Typography>
		</Box>
	);
}

const SystemInfoCard = () => {
	const { version, platform, lang } = useAbout();
	const { t } = useI18n();
	const buildDate = new Date().toLocaleDateString('pt-BR', {
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
	});

	return (
		<Card title={t('SYSTEM_INFO_TITLE')}>
			<InfoRow label={t('PROJECT_NAME')} value='Kuranas' />
			<InfoRow label={t('PROGRAM_VERSION')} value={version} />
			<InfoRow label={t('BUILD_VERSION')} value='20231001' />
			<InfoRow label={t('PLATFORM')} value={`${platform === 'windows' ? '🪟' : '🐧'} ${platform}`} />
			<InfoRow label={t('LANGUAGE')} value={lang} />
			<InfoRow label={t('BUILD_DATE')} value={buildDate} />
		</Card>
	);
};

export default SystemInfoCard;
