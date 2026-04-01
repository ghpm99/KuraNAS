export const formatSize = (size: number): string => {
    if (size < 1024) return `${size} B`;
    const units = ['KB', 'MB', 'GB', 'TB'];
    let unitIndex = -1;
    let formattedSize = size;

    do {
        formattedSize /= 1024;
        unitIndex++;
    } while (formattedSize >= 1024 && unitIndex < units.length - 1);

    return `${formattedSize.toFixed(2)} ${units[unitIndex]}`;
};
