import useI18n from '@/components/i18n/provider/i18nContext';
import { buildImageFolderUrl, buildVideoFolderUrl } from '@/app/routes';
import type { TakeoutImport } from '@/types/takeout';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { Link } from 'react-router-dom';
import useTakeoutDestinationFolders from './useTakeoutDestinationFolders';

type TakeoutImportSummaryProps = {
	takeoutImport: TakeoutImport;
};

const TakeoutImportSummary = ({ takeoutImport }: TakeoutImportSummaryProps) => {
	const { t } = useI18n();
	const { imagesFolder, videosFolder } = useTakeoutDestinationFolders();

	return (
		<Stack spacing={1}>
			<Typography variant="body2">{t('TAKEOUT_IMPORT_FINISHED')}</Typography>
			<Typography variant="body2">
				{t('TAKEOUT_SUMMARY_IMPORTED', { count: String(takeoutImport.imported) })}
			</Typography>
			<Typography variant="body2">
				{t('TAKEOUT_SUMMARY_ALREADY_EXISTING', {
					count: String(takeoutImport.already_existing),
				})}
			</Typography>
			<Typography variant="body2">
				{t('TAKEOUT_SUMMARY_IGNORED', { count: String(takeoutImport.ignored) })}
			</Typography>
			<Typography variant="body2">
				{t('TAKEOUT_SUMMARY_FAILED', { count: String(takeoutImport.failed_entries) })}
			</Typography>
			<Stack direction="row" spacing={1}>
				<Button
					component={Link}
					to={buildImageFolderUrl(imagesFolder)}
					variant="outlined"
					size="small"
				>
					{t('TAKEOUT_VIEW_IMPORTED_PHOTOS')}
				</Button>
				<Button
					component={Link}
					to={buildVideoFolderUrl(videosFolder)}
					variant="outlined"
					size="small"
				>
					{t('TAKEOUT_VIEW_IMPORTED_VIDEOS')}
				</Button>
			</Stack>
		</Stack>
	);
};

export default TakeoutImportSummary;
