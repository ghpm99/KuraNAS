import useFile, { FileData } from '@/components/hooks/fileProvider/fileContext';
import FolderItem from './components/folderItem';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Box, CircularProgress, List, Typography } from '@mui/material';

const FolderTree = () => {
	const { status, handleSelectItem, files, expandedItems, selectedItem } = useFile();
	const { t } = useI18n();

	if (status === 'loading') {
		return <Box sx={{ p: 1.5 }}><CircularProgress size={16} /></Box>;
	}
	if (status === 'error' && files.length === 0) {
		return <Typography variant='caption' color='error' sx={{ px: 1.5 }}>{t('ERROR_LOADING_FILES')}</Typography>;
	}

	const handleClick = (file: FileData) => {
		handleSelectItem(file.id);
	};

	const renderFiles = (fileArray: FileData[]): React.ReactNode => {
		if (!fileArray || fileArray.length === 0) {
			return <Typography variant='caption' sx={{ px: 1.5 }}>{t('EMPTY_FILE_LIST')}</Typography>;
		}
		return fileArray.map((file) => (
			<FolderItem
				key={file.id}
				type={file.type}
				label={file.name}
				onClick={() => handleClick(file)}
				expanded={expandedItems.includes(file.id)}
				selected={selectedItem?.id === file.id}
			>
				{file.file_children && file.file_children.length > 0 && (
					<Box sx={{ pl: 2 }}>{renderFiles(file.file_children)}</Box>
				)}
			</FolderItem>
		));
	};

	return (
		<Box sx={{ mt: 1 }}>
			<Typography
				variant='overline'
				sx={{ px: 1.5, color: 'text.secondary', cursor: 'pointer', display: 'block' }}
				onClick={() => handleSelectItem(null)}
			>
				{t('FILES')}
			</Typography>
			<List dense disablePadding>{renderFiles(files)}</List>
		</Box>
	);
};

export default FolderTree;
