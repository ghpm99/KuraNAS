import { ArrowLeft, Download, MoreHorizontal, Plus, Share2, Star } from 'lucide-react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../hooks/fileProvider/fileContext';
import { FileType } from '@/utils';
import { Box, Button, IconButton, Typography } from '@mui/material';
import { Link } from 'react-router-dom';

export const ActionBar = () => {
	const { selectedItem } = useFile();
	const { t } = useI18n();

	if (!selectedItem || selectedItem?.type === FileType.Directory) {
		return (
			<Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
				<Button variant='contained' size='small' startIcon={<Plus size={16} />}>
					{t('NEW_FILE')}
				</Button>
				<Button
					variant='outlined'
					size='small'
					startIcon={
						<svg viewBox='0 0 24 24' fill='none' stroke='currentColor' width={16} height={16}>
							<path d='M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3' strokeWidth='2' strokeLinecap='round' strokeLinejoin='round' />
						</svg>
					}
				>
					{t('UPLOAD_FILE')}
				</Button>
				<Button
					variant='outlined'
					size='small'
					startIcon={
						<svg viewBox='0 0 24 24' fill='none' stroke='currentColor' width={16} height={16}>
							<path d='M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' strokeWidth='2' strokeLinecap='round' strokeLinejoin='round' />
						</svg>
					}
				>
					{t('NEW_FOLDER')}
				</Button>
			</Box>
		);
	}

	return (
		<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
				<IconButton component={Link} to='/' size='small'>
					<ArrowLeft size={16} />
				</IconButton>
				<Typography variant='h6'>{selectedItem.name}</Typography>
			</Box>
			<Box sx={{ display: 'flex', gap: 1 }}>
				<IconButton size='small'><Star size={16} /></IconButton>
				<Button variant='outlined' size='small' startIcon={<Download size={16} />}>{t('DOWNLOAD')}</Button>
				<Button variant='contained' size='small' startIcon={<Share2 size={16} />}>{t('SHARE')}</Button>
				<IconButton size='small'><MoreHorizontal size={16} /></IconButton>
			</Box>
		</Box>
	);
};

export default ActionBar;
