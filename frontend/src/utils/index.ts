export enum FileType {
    Directory = 1,
    File = 2,
}

export { formatSize } from '@/shared/utils/formatSize';

export const formatDate = (dateString: string): string => {
    try {
        const date = new Date(dateString);
        return date.toLocaleString();
    } catch (error) {
        console.error('Failed to format date', error);
        return dateString;
    }
};

export const formatDuration = (seconds: number | undefined): string => {
    if (!seconds || seconds <= 0) return 'Em andamento';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
        return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
        return `${minutes}m ${secs}s`;
    } else {
        return `${secs}s`;
    }
};

export const formatDateTime = (date: Date): string => {
    return date.toLocaleString('pt-BR', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    });
};

export { getFileTypeInfo, hasDedicatedMediaScreen, isPreviewOnlyImageFormat } from './fileTypeInfo';
export type { FileTypeInfo } from './fileTypeInfo';
