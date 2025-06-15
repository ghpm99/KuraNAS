import { FileType, formatDate, formatSize, getFileTypeInfo } from '@/utils';
import useFile from '../hooks/fileProvider/fileContext';
import './fileDetails.css';
import useI18n from '../i18n/provider/i18nContext';
const FileDetails = () => {
	const { selectedItem, isLoadingAccessData, recentAccessFiles } = useFile();
	const { t } = useI18n();

	if (!selectedItem || selectedItem.type === FileType.Directory) return <></>;
	const fileType = getFileTypeInfo(selectedItem.format);
	return (
		<div className='file-details'>
			<div className='details-header'>
				<h2 className='details-title'>{t('FILE_DETAILS_TITLE')}</h2>
				<p className='details-subtitle'>{t('FILE_DETAILS_SUBTITLE')}</p>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>{t('PROPERTIES')}</h3>
				<div className='detail-item'>
					<span className='detail-label'>{t('TYPE')}</span>
					<span className='detail-value'>{fileType.description}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>{t('SIZE')}</span>
					<span className='detail-value'>
						{formatSize(selectedItem.size)}({selectedItem.size} B)
					</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>{t('CREATED')}</span>
					<span className='detail-value'>{formatDate(selectedItem.created_at)}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>{t('MODIFIED')}</span>
					<span className='detail-value'>{formatDate(selectedItem.updated_at)}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>{t('PATH')}</span>
					<span className='detail-value'>{selectedItem.path}</span>
				</div>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>{t('TAGS')}</h3>
				<div className='tag-list'></div>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>{t('RECENT_ACTIVITY')}</h3>
				{isLoadingAccessData ? (
					<p className='loading-message'>{t('LOADING_RECENT_ACTIVITY')}</p>
				) : (
					<ul className='activity-list'>
						{recentAccessFiles.map((access) =>
							access.file_id !== selectedItem.id ? null : (
								<li key={access.id} className='activity-item'>
									<span className='activity-ip'>{access.ip_address}</span>
									<span className='activity-date'>{formatDate(access.accessed_at)}</span>
								</li>
							)
						)}
					</ul>
				)}
			</div>
		</div>
	);
};

export default FileDetails;
