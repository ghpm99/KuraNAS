import { getFileBlobUrl } from '@/service/files';
import styles from './DocumentViewer.module.css';

type DocumentPdfFrameProps = {
    fileId: number;
    title: string;
    pageNumber?: number;
};

const DocumentPdfFrame = ({ fileId, title, pageNumber }: DocumentPdfFrameProps) => {
    const pageFragment = pageNumber === undefined ? '' : `#page=${pageNumber}`;

    return (
        <iframe
            className={styles.pdfFrame}
            title={title}
            src={`${getFileBlobUrl(fileId)}${pageFragment}`}
        />
    );
};

export default DocumentPdfFrame;
