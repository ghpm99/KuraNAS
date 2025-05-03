import FileCard from '../filecard';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../providers/fileprovider/fileContext';
import './filecontent.css';

const imageFormat = ['.jpg', '.jpeg', '.png', '.gif', '.bmp'];

const audioFormat = ['.mp3'];

const videoFormat = ['.mp4', '.webm'];

const FileContent = () => {
	const { status, handleSelectItem, selectedItem, files } = useFile();
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
							metadata={file.type === 1 ? '' : `${file.size}`}
							thumbnail={`${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${file.id}`}
							onClick={() => handleSelectItem(file.id)}
						/>
					))}
				</div>
			</>
		);
	}

	if (selectedItem.type === 1) {
		return (
			<>
				<h1>{selectedItem.name}</h1>
				<div className='file-grid'>
					{selectedItem?.file_children?.map((file) => (
						<FileCard
							title={file.name}
							metadata={file.type === 1 ? '' : `${file.size}`}
							thumbnail={`${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${file.id}`}
							onClick={() => handleSelectItem(file.id)}
						/>
					))}
				</div>
			</>
		);
	}

	if (imageFormat.includes(selectedItem.format)) {
		return (
			<>
				<h1>{selectedItem.name}</h1>
				<img src={`${import.meta.env.VITE_API_URL}/api/v1/files/blob/${selectedItem.id}`} />
			</>
		);
	}

	if (audioFormat.includes(selectedItem.format)) {
		return (
			<>
				<h1>{selectedItem.name}</h1>
				<audio controls>
					<source src={`${import.meta.env.VITE_API_URL}/api/v1/files/blob/${selectedItem.id}`} type='audio/mpeg' />
					Your browser does not support the audio element.
				</audio>
			</>
		);
	}

	if (videoFormat.includes(selectedItem.format)) {
		return (
			<>
				<h1>{selectedItem.name}</h1>
				<video controls id={selectedItem.id.toString()}>
					<source src={`${import.meta.env.VITE_API_URL}/api/v1/files/blob/${selectedItem.id}`} type='video/mp4' />
				</video>
			</>
		);
	}

	if (selectedItem.format === '.pdf') {
		return (
			<>
				<h1>{selectedItem.name}</h1>
				<embed
					title={selectedItem.name}
					className='embed'
					src={`${import.meta.env.VITE_API_URL}/api/v1/files/blob/${selectedItem.id}`}
				/>
			</>
		);
	}
};

export default FileContent;
