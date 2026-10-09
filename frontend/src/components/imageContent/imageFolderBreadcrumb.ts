export type ImageFolderCrumb = {
    name: string;
    path: string;
};

export const buildFolderCrumbs = (folderPath: string): ImageFolderCrumb[] => {
    const segments = folderPath.replace(/\\/g, '/').split('/').filter(Boolean);
    return segments.map((segment, position) => ({
        name: segment,
        path: `/${segments.slice(0, position + 1).join('/')}`,
    }));
};
