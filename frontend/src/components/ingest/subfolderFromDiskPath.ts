const normalizeDiskPath = (diskPath: string): string =>
    diskPath.trim().replace(/\\/g, '/').replace(/\/+$/, '');

export const subfolderFromDiskPath = (rootPath: string, folderDiskPath: string): string | null => {
    const root = normalizeDiskPath(rootPath);
    const folder = normalizeDiskPath(folderDiskPath);
    if (folder === root) return '';
    if (!folder.startsWith(`${root}/`)) return null;
    return folder.slice(root.length + 1);
};
