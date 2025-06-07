import { ArrowLeft, Download, MoreHorizontal, Plus, Share2, Star } from 'lucide-react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../hooks/fileProvider/fileContext';
import { FileType } from '@/utils';
import './actionBar.css';

export const ActionBar = () => {
	const { selectedItem } = useFile();
	const { t } = useI18n();
	if (!selectedItem || selectedItem?.type === FileType.Directory) {
		return (
			<div className='action-bar'>
				<button className='button primary-button'>
					<Plus className='icon' />
					{t('NEW_FILE')}
				</button>
				<button className='button secondary-button'>
					<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
						<path
							d='M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3'
							strokeWidth='2'
							strokeLinecap='round'
							strokeLinejoin='round'
						/>
					</svg>
					{t('UPLOAD_FILE')}
				</button>
				<button className='button secondary-button'>
					<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
						<path
							d='M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
							strokeWidth='2'
							strokeLinecap='round'
							strokeLinejoin='round'
						/>
					</svg>
					{t('NEW_FOLDER')}
				</button>
			</div>
		);
	}

	return (
		<div className='action-bar-file'>
			<div className='file-title-container'>
				<a href='/' className='icon-button'>
					<ArrowLeft className='icon' />
				</a>
				<h1 className='file-title-action'>{selectedItem.name}</h1>
			</div>
			<div className='file-actions'>
				<button className='icon-button'>
					<Star className='icon' />
				</button>
				<button className='secondary-button'>
					<Download className='icon' />
					{t('DOWNLOAD')}
				</button>
				<button className='primary-button'>
					<Share2 className='icon' />
					{t('SHARE')}
				</button>
				<button className='icon-button'>
					<MoreHorizontal className='icon' />
				</button>
			</div>
		</div>
	);
};

export default ActionBar;
