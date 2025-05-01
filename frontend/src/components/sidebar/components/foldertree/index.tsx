import useFile, { FileData } from '@/components/providers/fileprovider/fileContext';
import FolderItem from './components/folderitem';
import './folderTree.css';
import useI18n from '@/components/i18n/provider/i18nContext';

const FolderTree = () => {
	const { status, handleSelectItem, files, expandedItems, selectedItem } = useFile();
	const { t } = useI18n();

	if (status === 'loading') {
		return <div>{t('LOADING')}</div>;
	}
	if (status === 'error' && files.length === 0) {
		return <div>{t('ERROR_LOADING_FILES')}</div>;
	}

	const handleClick = (file: FileData) => {
		handleSelectItem(file.id);
	};

	const renderFiles = (fileArray: FileData[]) => {
		if (!fileArray || fileArray.length === 0) {
			return <div>{t('EMPTY_FILE_LIST')}</div>;
		}
		const fileComponent = fileArray.map((file) => (
			<FolderItem
				key={file.id}
				type={file.type}
				label={file.name}
				onClick={() => handleClick(file)}
				expanded={expandedItems.includes(file.id)}
				selected={selectedItem?.id === file.id}
			>
				{file.file_children?.length > 0 && <div className='folder-children'>{renderFiles(file.file_children)}</div>}
			</FolderItem>
		));

		return fileComponent;
	};

	return (
		<div className='nav-section'>
			<div className='nav-section-title'>{t('FILES')}</div>
			<div className='folder-list'>{renderFiles(files)}</div>
		</div>
	);
};

export default FolderTree;
