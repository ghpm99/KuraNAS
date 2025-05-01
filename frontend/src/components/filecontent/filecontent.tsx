import FileCard from '../filecard';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../providers/fileprovider/fileContext';
import './filecontent.css';

const FileContent = () => {
	const { status, selectedItem, files } = useFile();
	const { t } = useI18n();

	if (status === 'loading') {
		return <div>{t('LOADING')}</div>;
	}
	if (status === 'error') {
		return <div>{t('ERROR_LOADING_FILES')}</div>;
	}

	if (!selectedItem) {
		return (
			<>
				<h1>{t('FILES')}</h1>
				<div className='file-grid'>
					{files?.map((file) => (
						<FileCard
							title={file.name}
							metadata='Shared folder • 8 presentations'
							thumbnail={`${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${file.id}`}
						/>
					))}
				</div>
			</>
		);
	}

	return (
		<>
			<h1>{selectedItem.name}</h1>
			<div className='file-grid'>
				{selectedItem?.file_children?.map((file) => (
					<FileCard
						title={file.name}
						metadata='Shared folder • 8 presentations'
						thumbnail={`${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${file.id}`}
					/>
				))}
			</div>
		</>
	);
};

export default FileContent;
