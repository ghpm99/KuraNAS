import FileCard from '../filecard';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../providers/fileprovider/fileContext';
import './filecontent.css';

const FileContent = () => {
	const { status, selectedItem } = useFile();
	const { t } = useI18n();

	if (status === 'loading') {
		return <div>Carregando...</div>;
	}
	if (status === 'error') {
		return <div>Error loading files</div>;
	}

	if (!selectedItem) {
		return <div>{t('NO_FILE_SELECTED')}</div>;
	}

	return (
		<>
			<h1>{selectedItem.name}</h1>
			<div className='file-grid'>
				{selectedItem?.file_children?.map((file) => (
					<FileCard title={file.name} metadata='Shared folder • 8 presentations' thumbnail='/placeholder.svg' />
				))}
			</div>
		</>
	);
};

export default FileContent;
