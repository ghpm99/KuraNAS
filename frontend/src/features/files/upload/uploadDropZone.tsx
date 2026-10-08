import { useRef, useState, type DragEvent, type ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { collectDroppedEntries } from './collectDroppedEntries';
import useUploadToCurrentFolder from './useUploadToCurrentFolder';
import styles from './uploadDropZone.module.css';

const carriesFiles = (event: DragEvent<HTMLElement>): boolean =>
    Array.from(event.dataTransfer?.types ?? []).includes('Files');

const UploadDropZone = ({ children }: { children: ReactNode }) => {
    const { t } = useI18n();
    const { uploadEntries } = useUploadToCurrentFolder();
    const [isDragging, setIsDragging] = useState(false);
    const dragDepthRef = useRef(0);

    const handleDragEnter = (event: DragEvent<HTMLElement>) => {
        if (!carriesFiles(event)) return;
        event.preventDefault();
        dragDepthRef.current += 1;
        setIsDragging(true);
    };

    const handleDragOver = (event: DragEvent<HTMLElement>) => {
        if (!carriesFiles(event)) return;
        event.preventDefault();
    };

    const handleDragLeave = (event: DragEvent<HTMLElement>) => {
        if (!carriesFiles(event)) return;
        dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);
        if (dragDepthRef.current === 0) setIsDragging(false);
    };

    const handleDrop = async (event: DragEvent<HTMLElement>) => {
        if (!carriesFiles(event)) return;
        event.preventDefault();
        dragDepthRef.current = 0;
        setIsDragging(false);
        const droppedEntries = await collectDroppedEntries(event.dataTransfer);
        await uploadEntries(droppedEntries);
    };

    return (
        <div
            className={styles.dropZone}
            onDragEnter={handleDragEnter}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
        >
            {children}
            {isDragging ? (
                <div className={styles.overlay} role="presentation">
                    <span className={styles.overlayText}>{t('FILES_UPLOAD_DROP_HINT')}</span>
                </div>
            ) : null}
        </div>
    );
};

export default UploadDropZone;
