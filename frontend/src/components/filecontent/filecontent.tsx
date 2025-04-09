import FileCard from '../filecard';
import useFile from '../providers/fileprovider/fileContext';
import './filecontent.css';

const FileContent = () => {
	const { status, selectedItem } = useFile();

	if (status === 'loading') {
		return <div>Carregando...</div>;
	}
	if (status === 'error') {
		return <div>Error loading files</div>;
	}

	if (!selectedItem) {
		return <div>No file selected</div>;
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
