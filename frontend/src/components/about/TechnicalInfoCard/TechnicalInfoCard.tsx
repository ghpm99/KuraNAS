import Card from '@/components/ui/Card/Card';
import { useState } from 'react';
import { Box, Button, Divider, Typography } from '@mui/material';
import { Copy } from 'lucide-react';
import { useAbout } from '@/components/providers/aboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';

function BuildRow({ label, value }: { label: string; value: string }) {
	return (
		<Box sx={{ display: 'flex', justifyContent: 'space-between', py: 0.75, borderBottom: 1, borderColor: 'divider' }}>
			<Typography variant='body2' color='text.secondary'>{label}</Typography>
			<Typography variant='body2'>{value}</Typography>
		</Box>
	);
}

const TechnicalInfoCard = () => {
	const { commit_hash, gin_mode, gin_version, go_version, node_version } = useAbout();
	const [copied, setCopied] = useState(false);
	const { t } = useI18n();

	const copyCommitHash = async () => {
		try {
			await navigator.clipboard.writeText(commit_hash);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch (error) {
			console.error('Failed to copy commit hash', error);
		}
	};

	return (
		<Card title={t('TECHNICAL_INFO_TITLE')}>
			<Box sx={{ mb: 2 }}>
				<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
					<Typography variant='body2' fontWeight={500}>{t('COMMIT_HASH')}</Typography>
					<Button variant='outlined' size='small' startIcon={<Copy size={14} />} onClick={copyCommitHash}>
						{copied ? t('COPIED') : t('COPY')}
					</Button>
				</Box>
				<Typography variant='body2' sx={{ fontFamily: 'monospace', mb: 0.5 }}>{commit_hash}</Typography>
				<Typography variant='caption' color='text.secondary'>{t('COMMIT_DESCRIPTION')}</Typography>
			</Box>

			<Divider sx={{ my: 1.5 }} />
			<Typography variant='subtitle2' gutterBottom>{t('BUILD_DETAILS_TITLE')}</Typography>
			<BuildRow label={t('ENVIRONMENT')} value={gin_mode} />
			<BuildRow label={t('COMPILER')} value={go_version} />
			<BuildRow label={t('BACKEND')} value={gin_version} />
			<BuildRow label={t('NODEJS')} value={node_version} />
		</Card>
	);
};

export default TechnicalInfoCard;
