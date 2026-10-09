import { FileType } from '@/utils';
import { formatSize } from '@/shared/utils/formatSize';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

type Translate = (key: string) => string;

export const formatModifiedDate = (isoTimestamp: string, locale?: string): string => {
    const modifiedAt = new Date(isoTimestamp);
    if (Number.isNaN(modifiedAt.getTime())) return '';
    return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' }).format(
        modifiedAt
    );
};

export const formatSizeColumn = (file: FileData, translate: Translate): string => {
    if (file.type === FileType.File) return formatSize(file.size);
    const itemCount = file.directory_content_count ?? 0;
    return `${itemCount} ${itemCount === 1 ? translate('ITEM') : translate('ITENS')}`;
};

export const formatTypeColumn = (file: FileData, translate: Translate): string => {
    if (file.type !== FileType.File) return translate('FOLDER');
    return (file.format ?? '').replace(/^\./, '').toUpperCase();
};
