import { FileData } from '@/components/providers/fileProvider/fileContext';
import { getFileTypeInfo } from '@/utils';
import './fileViewer.css';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';

const FileViewer = ({ file }: { file: FileData }) => {
    const { t } = useI18n();
    const blobUrl = (id: number) => `${getApiV1BaseUrl()}/files/blob/${id}`;
    const fileType = getFileTypeInfo(file.format);
    if (fileType.type === 'image') {
        return <img src={blobUrl(file.id)} alt={file.name} />;
    }

    if (fileType.type === 'audio') {
        return (
            <audio controls>
                <source src={blobUrl(file.id)} type={fileType.mime} />
                {t('AUDIO_NOT_SUPPORTED')}
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
        return (
            <embed
                title={file.name}
                className="embed"
                src={blobUrl(file.id)}
                type={fileType.mime}
            />
        );
    }

    if (fileType.type === 'archive') {
        return (
            <a className="download-file" href={blobUrl(file.id)} download={file.name}>
                {t('DOWNLOAD_FILE', { fileName: file.name })}
            </a>
        );
    }

    return (
        <p>
            {t('UNSUPPORTED_FILE_FORMAT')} {t(fileType.description)}
        </p>
    );
};

export default FileViewer;
