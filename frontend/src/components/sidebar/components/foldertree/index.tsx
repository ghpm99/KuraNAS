import useFile, { FileData } from '@/components/providers/fileprovider/fileContext';
import FolderItem from './components/folderitem';
import './folderTree.css';

const FolderTree = () => {
	const { status, handleSelectItem, files, expandedItems, selectedItem } = useFile();

	if (status === 'loading') {
		return <div>Loading...</div>;
	}
	if (status === 'error') {
		return <div>Error loading files</div>;
	}

	const handleClick = (file: FileData) => {
		handleSelectItem(file);
	};

	const renderFiles = (fileArray: FileData[]) => {
		if (!fileArray || fileArray.length === 0) {
			return <div>No files found</div>;
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
			<div className='nav-section-title'>Arquivos</div>
			<div className='folder-list'>{renderFiles(files)}</div>
		</div>
	);
};

export default FolderTree;
