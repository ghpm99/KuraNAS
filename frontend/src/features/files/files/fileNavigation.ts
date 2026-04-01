import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export const findTrailById = (
    nodes: FileData[],
    targetId: number | null,
    parents: FileData[] = []
): FileData[] | null => {
    if (!targetId) {
        return null;
    }

    for (const node of nodes) {
        const nextParents = [...parents, node];
        if (node.id === targetId) {
            return nextParents;
        }

        if (node.file_children?.length) {
            const branch = findTrailById(node.file_children, targetId, nextParents);
            if (branch) {
                return branch;
            }
        }
    }

    return null;
};
