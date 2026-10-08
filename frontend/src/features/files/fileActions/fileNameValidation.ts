type FileNameIssue = 'empty' | 'invalidCharacters' | 'unchanged';

const forbiddenCharacters = ['/', '\\'];

export const findFileNameIssue = (typedName: string, currentName?: string): FileNameIssue | null => {
    const trimmedName = typedName.trim();
    if (trimmedName === '') return 'empty';
    if (forbiddenCharacters.some((character) => trimmedName.includes(character))) {
        return 'invalidCharacters';
    }
    if (currentName !== undefined && trimmedName === currentName) return 'unchanged';
    return null;
};

export const fileNameIssueMessageKeys: Record<FileNameIssue, string> = {
    empty: 'FILES_NAME_ERROR_EMPTY',
    invalidCharacters: 'FILES_NAME_ERROR_INVALID_CHARACTERS',
    unchanged: 'FILES_NAME_HINT_UNCHANGED',
};

export const findBaseNameLength = (name: string, isFolder: boolean): number => {
    if (isFolder) return name.length;
    const extensionStart = name.lastIndexOf('.');
    return extensionStart > 0 ? extensionStart : name.length;
};
