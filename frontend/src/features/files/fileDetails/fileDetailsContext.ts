import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { createContext } from 'react';

export type FileDetailsContextValue = {
    isAvailable: boolean;
    explicitTarget: FileData | null;
    openDetails: (file: FileData) => void;
    closeDetails: () => void;
};

export const unavailableFileDetails: FileDetailsContextValue = {
    isAvailable: false,
    explicitTarget: null,
    openDetails: () => undefined,
    closeDetails: () => undefined,
};

export const FileDetailsContext = createContext<FileDetailsContextValue>(unavailableFileDetails);
