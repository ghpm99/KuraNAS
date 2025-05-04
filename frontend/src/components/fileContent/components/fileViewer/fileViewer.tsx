import { FileData } from '@/components/providers/fileProvider/fileContext';
import { getFileTypeInfo } from '@/utils';
import './fileViewer.css';

const FileViewer = ({ file }: { file: FileData }) => {
	const blobUrl = (id: number) => `${import.meta.env.VITE_API_URL}/api/v1/files/blob/${id}`;
	const fileType = getFileTypeInfo(file.format);
	if (fileType.type === 'image') {
		return <img src={blobUrl(file.id)} alt={file.name} />;
	}

	if (fileType.type === 'audio') {
		return (
			<audio controls>
				<source src={blobUrl(file.id)} type={fileType.mime} />
				Your browser does not support the audio element.
			</audio>
		);
	}

	if (fileType.type === 'video') {
		return (
			<video controls id={file.id.toString()}>
				<source src={blobUrl(file.id)} type={fileType.mime} />
			</video>
		);
	}

	if (fileType.type === 'document') {
		return <embed title={file.name} className='embed' src={blobUrl(file.id)} type={fileType.mime} />;
	}

	if (fileType.type === 'archive') {
		return (
			<a className='download-file' href={blobUrl(file.id)} download={file.name}>
				Baixar {file.name}
			</a>
		);
	}

	return <p>Formato de arquivo não suportado: {fileType.description}</p>;
};

export default FileViewer;
