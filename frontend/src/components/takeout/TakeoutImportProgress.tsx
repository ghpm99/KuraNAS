import ErrorState from '@/components/errorState/errorState';
import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import TakeoutImportSummary from './TakeoutImportSummary';
import useTakeoutImportProgress, { isTakeoutImportActive } from './useTakeoutImportProgress';

type TakeoutImportProgressProps = {
	importId: number;
};

const TakeoutImportProgress = ({ importId }: TakeoutImportProgressProps) => {
	const { t } = useI18n();
	const { data: takeoutImport, error, refetch } = useTakeoutImportProgress(importId);

	if (error) {
		return (
			<ErrorState
				title={t('TAKEOUT_IMPORT_PROGRESS_ERROR')}
				backendMessage={extractBackendErrorMessage(error)}
				onRetry={() => void refetch()}
			/>
		);
	}

	if (!takeoutImport || isTakeoutImportActive(takeoutImport)) {
		const processedEntries = takeoutImport?.processed_entries ?? 0;
		const totalEntries = takeoutImport?.total_entries ?? 0;
		const percentage =
			totalEntries > 0
				? Math.min(100, Math.round((processedEntries / totalEntries) * 100))
				: 0;
		return (
			<>
				<LinearProgress variant="determinate" value={percentage} />
				<Typography variant="body2">
					{t('TAKEOUT_IMPORT_PROGRESS', {
						processed: String(processedEntries),
						total: String(totalEntries),
					})}
				</Typography>
			</>
		);
	}

	if (takeoutImport.state === 'completed') {
		return <TakeoutImportSummary takeoutImport={takeoutImport} />;
	}

	return <Typography variant="body2">{t('TAKEOUT_IMPORT_FAILED')}</Typography>;
};

export default TakeoutImportProgress;
