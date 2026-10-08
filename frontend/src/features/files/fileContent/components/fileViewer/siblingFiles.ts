import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { FileType } from '@/utils';

export type SiblingPosition = {
    previous: FileData | null;
    next: FileData | null;
    position: number;
    total: number;
};

const findNodeByPath = (nodes: FileData[], path: string): FileData | null => {
    for (const node of nodes) {
        if (node.path === path) return node;
        const nodeInChildren = findNodeByPath(node.file_children ?? [], path);
        if (nodeInChildren) return nodeInChildren;
    }
    return null;
};

const findLoadedSiblings = (tree: FileData[], file: FileData): FileData[] => {
    if (tree.some((node) => node.id === file.id)) return tree;
    return findNodeByPath(tree, file.parent_path)?.file_children ?? [];
};

export const findSiblingFiles = (tree: FileData[], file: FileData): FileData[] =>
    findLoadedSiblings(tree, file).filter((sibling) => sibling.type === FileType.File);

export const locateAmongSiblings = (
    siblingFiles: FileData[],
    fileId: number
): SiblingPosition | null => {
    const currentIndex = siblingFiles.findIndex((sibling) => sibling.id === fileId);
    if (currentIndex === -1) return null;
    return {
        previous: siblingFiles[currentIndex - 1] ?? null,
        next: siblingFiles[currentIndex + 1] ?? null,
        position: currentIndex + 1,
        total: siblingFiles.length,
    };
};
