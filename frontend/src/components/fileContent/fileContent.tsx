import { FileType, formatSize } from '@/utils';
import FileCard from '../fileCard';
import useI18n from '../i18n/provider/i18nContext';
import useFile, { FileData } from '../providers/fileProvider/fileContext';
import FileViewer from './components/fileViewer/fileViewer';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import './fileContent.css';

const FileContent = () => {
	const { status, handleSelectItem, selectedItem, files, handleStarredItem, fileListFilter } = useFile();
	const { t } = useI18n();
	const currentListTitle =
		fileListFilter === 'starred' ? t('STARRED_FILES') : fileListFilter === 'recent' ? t('RECENT_FILES') : t('FILES');

	if (status === 'pending') {
		return <div className='file-content'>{t('LOADING')}</div>;
	}
	if (status === 'error') {
		return <div className='file-content'>{t('ERROR_LOADING_FILES')}</div>;
	}

	const fileMetadata = (file: FileData): string => {
		if (file.type === FileType.File) {
			const format = file.format ? `${file.format} - ` : '';
			const fileSize = formatSize(file.size);

			return `${format}${fileSize}`;
		}
		const directoryContentCount = file.directory_content_count;
		const countText = directoryContentCount > 1 ? t('ITENS') : t('ITEM');
		return `${t('FOLDER')} - ${directoryContentCount} ${countText}`;
	};

	const thumbnailUrl = (id: number) => `${getApiV1BaseUrl()}/files/thumbnail/${id}`;

	if (!selectedItem) {
		return (
			<div className='file-content'>
				<h1>{currentListTitle}</h1>
				<div className='file-grid'>
					{files?.map((file) => (
						<FileCard
							key={file.id}
							title={file.name}
							starred={file.starred}
							metadata={fileMetadata(file)}
							thumbnail={thumbnailUrl(file.id)}
							onClick={() => handleSelectItem(file.id)}
							onClickStar={() => handleStarredItem(file.id)}
						/>
					))}
				</div>
			</div>
		);
	}

	if (selectedItem.type === FileType.Directory) {
		return (
			<div className='file-content'>
				<h1>{selectedItem.name}</h1>
				<div className='file-grid'>
					{selectedItem?.file_children?.map((file) => (
						<FileCard
							key={file.id}
							title={file.name}
							starred={file.starred}
							metadata={fileMetadata(file)}
							thumbnail={thumbnailUrl(file.id)}
							onClick={() => handleSelectItem(file.id)}
							onClickStar={() => handleStarredItem(file.id)}
						/>
					))}
				</div>
			</div>
		);
	}

	return (
		<div className='preview-container'>
			<FileViewer file={selectedItem} />
		</div>
	);
};

export default FileContent;
