import type { UploadConflictPolicy } from '@/service/files';

export type UploadConflictChoice = 'ask' | 'rename' | 'replace' | 'skip';

export const defaultConflictPolicy: UploadConflictChoice = 'ask';

export const conflictPolicyStorageKey = 'kuranas.files.upload.conflictPolicy';

const validChoices: UploadConflictChoice[] = ['ask', 'rename', 'replace', 'skip'];

const isConflictChoice = (storedValue: string | null): storedValue is UploadConflictChoice =>
    validChoices.includes(storedValue as UploadConflictChoice);

export const toWireConflictPolicy = (choice: UploadConflictChoice): UploadConflictPolicy =>
    choice === 'ask' ? 'fail' : choice;

export const loadConflictPolicy = (): UploadConflictChoice => {
    try {
        const storedChoice = window.localStorage.getItem(conflictPolicyStorageKey);
        return isConflictChoice(storedChoice) ? storedChoice : defaultConflictPolicy;
    } catch {
        return defaultConflictPolicy;
    }
};

export const saveConflictPolicy = (choice: UploadConflictChoice): void => {
    try {
        window.localStorage.setItem(conflictPolicyStorageKey, choice);
    } catch {
        return;
    }
};
