import useFile, { FileData } from '@/components/providers/fileprovider/fileContext';
import FolderItem from './components/folderitem';
import './folderTree.css';

const FolderTree = () => {
	const fileContext = useFile();

	if (fileContext.status === 'loading') {
		return <div>Loading...</div>;
	}
	if (fileContext.status === 'error') {
		return <div>Error loading files</div>;
	}

	const handleClick = (file: FileData) => {
		fileContext.handleSelectItem(file);
	}

	const renderFiles = (fileArray: FileData[]) => {
		if (!fileArray || fileArray.length === 0) {
			return <div>No files found</div>;
		}
		const fileComponent = fileArray.map((file) => (
			<FolderItem key={file.id} type={file.type} label={file.name} onClick={() => handleClick(file)}>
				{file.file_children?.length > 0 && <div className='folder-children'>{renderFiles(file.file_children)}</div>}
			</FolderItem>
		));

		return fileComponent;
	};

	return (
		<div className='nav-section'>
			<div className='nav-section-title'>Arquivos</div>
			<div className='folder-list'>{renderFiles(fileContext.files)}</div>
		</div>
	);
};

export default FolderTree;
