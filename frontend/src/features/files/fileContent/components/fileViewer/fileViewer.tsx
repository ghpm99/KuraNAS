import { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getFileTypeInfo } from '@/utils';
import './fileViewer.css';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getFileBlobUrl, getFileThumbnailUrl } from '@/service/files';
import { useState } from 'react';
import TextFileViewer from './textFileViewer';
import UnsupportedFileCard from './unsupportedFileCard';

const pdfMime = 'application/pdf';

const FileViewer = ({ file }: { file: FileData }) => {
    const { t } = useI18n();
    const [failedPlaybackFileId, setFailedPlaybackFileId] = useState<number | null>(null);
    const fileType = getFileTypeInfo(file.format);
    const hasPlaybackFailed = failedPlaybackFileId === file.id;
    const markPlaybackFailed = () => setFailedPlaybackFileId(file.id);

    if (fileType.type === 'image') {
        const imageUrl = fileType.isThumbnailOnly
            ? getFileThumbnailUrl(file.id)
            : getFileBlobUrl(file.id);
        return <img src={imageUrl} alt={file.name} />;
    }

    if (fileType.type === 'audio' && !hasPlaybackFailed) {
        return (
            <audio controls onError={markPlaybackFailed}>
                <source
                    src={getFileBlobUrl(file.id)}
                    type={fileType.mime}
                    onError={markPlaybackFailed}
                />
                {t('AUDIO_NOT_SUPPORTED')}
            </audio>
        );
    }

    if (fileType.type === 'video' && !hasPlaybackFailed) {
        return (
            <video controls id={file.id.toString()} onError={markPlaybackFailed}>
                <source
                    src={getFileBlobUrl(file.id)}
                    type={fileType.mime}
                    onError={markPlaybackFailed}
                />
            </video>
        );
    }

    if (fileType.type === 'document' && fileType.mime === pdfMime) {
        return (
            <embed
                title={file.name}
                className="embed"
                src={getFileBlobUrl(file.id)}
                type={fileType.mime}
            />
        );
    }

    if (fileType.type === 'text') {
        return <TextFileViewer file={file} />;
    }

    return (
        <UnsupportedFileCard
            file={file}
            reasonKey={hasPlaybackFailed ? 'FILE_PREVIEW_PLAYBACK_FAILED' : undefined}
        />
    );
};

export default FileViewer;
