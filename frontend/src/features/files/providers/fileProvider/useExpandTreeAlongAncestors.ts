import { useQueries, useQuery } from '@tanstack/react-query';
import { useEffect, useMemo } from 'react';
import { getFilesTree } from '@/service/files';
import { FileType } from '@/utils';
import type { FileData, FilesSort, PaginationResponse } from './fileContext';
import { mergeChildrenIntoTree } from './fileProviderUtils';
import useFileAncestors from './useFileAncestors';

const treePageSize = 200;
const treeStaleTimeMs = 30_000;
const primaryRootPath = '/';

type AncestorChildren = {
    ancestorId: number;
    items: FileData[];
};

const collectLoadedAncestorChildren = (
    results: { data?: AncestorChildren }[]
): AncestorChildren[] =>
    results.flatMap((result) => (result.data ? [result.data] : []));

type Params = {
    selectedItemId: number | null;
    filesSort: FilesSort;
    setFileTree: React.Dispatch<React.SetStateAction<FileData[]>>;
};

const useExpandTreeAlongAncestors = ({ selectedItemId, filesSort, setFileTree }: Params) => {
    const { data: ancestors } = useFileAncestors(selectedItemId);

    const { data: rootLevelPage } = useQuery<PaginationResponse>({
        queryKey: ['files-tree-root', filesSort],
        queryFn: () =>
            getFilesTree({ page: 1, pageSize: treePageSize, category: 'all', sort: filesSort }),
        enabled: selectedItemId !== null,
        staleTime: treeStaleTimeMs,
        retry: false,
    });
    const rootLevelItems = rootLevelPage?.items;

    const expandableAncestorIds = useMemo(() => {
        const rootLevelIds = new Set((rootLevelItems ?? []).map((item) => item.id));
        return (ancestors ?? [])
            .filter(
                (ancestor) =>
                    ancestor.type === FileType.Directory &&
                    (ancestor.path !== primaryRootPath || rootLevelIds.has(ancestor.id))
            )
            .map((ancestor) => ancestor.id);
    }, [ancestors, rootLevelItems]);

    const loadedAncestorChildren = useQueries({
        queries: expandableAncestorIds.map((ancestorId) => ({
            queryKey: ['files-tree-children', ancestorId, filesSort],
            queryFn: async (): Promise<AncestorChildren> => {
                const page = await getFilesTree({
                    page: 1,
                    pageSize: treePageSize,
                    fileParent: ancestorId,
                    category: 'all',
                    sort: filesSort,
                });
                return { ancestorId, items: page?.items ?? [] };
            },
            staleTime: treeStaleTimeMs,
            retry: false,
        })),
        combine: collectLoadedAncestorChildren,
    });

    useEffect(() => {
        if (!rootLevelItems) return;
        let cancelled = false;
        queueMicrotask(() => {
            if (cancelled) return;
            setFileTree((currentTree) => {
                const treeWithRootLevel = currentTree.length === 0 ? rootLevelItems : currentTree;
                return expandableAncestorIds.reduce((expandedTree, ancestorId) => {
                    const loaded = loadedAncestorChildren.find(
                        (entry) => entry.ancestorId === ancestorId
                    );
                    return loaded
                        ? mergeChildrenIntoTree(expandedTree, ancestorId, loaded.items)
                        : expandedTree;
                }, treeWithRootLevel);
            });
        });
        return () => {
            cancelled = true;
        };
    }, [rootLevelItems, expandableAncestorIds, loadedAncestorChildren, setFileTree]);
};

export default useExpandTreeAlongAncestors;
