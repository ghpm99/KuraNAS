export const MUSIC_COLLECTION_PAGE_SIZE = 50;

export const handleKeyboardActivation = (
    event: React.KeyboardEvent<HTMLElement>,
    onActivate: () => void
) => {
    if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        onActivate();
    }
};

export const getFolderName = (path: string) => {
    const parts = path.split('/').filter(Boolean);
    return parts[parts.length - 1] || path;
};
