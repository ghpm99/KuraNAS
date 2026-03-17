import { render, screen, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
    addChildrenToTree,
    buildFilesUrl,
    extractFilePath,
    findItemInTree,
    findTrailByIdInTree,
} from './fileProviderUtils';
import type { FileData, FileContextType, PaginationResponse } from './fileContext';

// --- mocks ---

const mockNavigate = jest.fn();
let mockPathname = '/files';

jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useLocation: () => ({
        pathname: mockPathname,
        search: '',
        hash: '',
        state: null,
        key: 'default',
    }),
    useNavigate: () => mockNavigate,
}));

const mockGetFileByPath = jest.fn<Promise<FileData | null>, [string]>();
const mockGetFilesTree = jest.fn<Promise<PaginationResponse>, [any]>();
const mockGetRecentAccessByFileId = jest.fn<Promise<any[]>, [number]>();
const mockToggleStarredFile = jest.fn<Promise<void>, [number]>();
const mockRescanFiles = jest.fn<Promise<void>, []>();
const mockUploadFilesToPath = jest.fn<Promise<void>, [FileList, string?]>();
const mockCreateFolderAtPath = jest.fn<Promise<void>, [string, string?]>();
const mockMoveFilePath = jest.fn<Promise<void>, [string, string]>();
const mockCopyFilePath = jest.fn<Promise<void>, [string, string]>();
const mockRenameFilePath = jest.fn<Promise<void>, [string, string]>();
const mockDeleteFilePath = jest.fn<Promise<void>, [string]>();

jest.mock('@/service/files', () => ({
    getFileByPath: mockGetFileByPath,
    getFilesTree: mockGetFilesTree,
    getRecentAccessByFileId: mockGetRecentAccessByFileId,
    toggleStarredFile: mockToggleStarredFile,
    rescanFiles: mockRescanFiles,
    uploadFilesToPath: mockUploadFilesToPath,
    createFolderAtPath: mockCreateFolderAtPath,
    moveFilePath: mockMoveFilePath,
    copyFilePath: mockCopyFilePath,
    renameFilePath: mockRenameFilePath,
    deleteFilePath: mockDeleteFilePath,
}));

// Capture context from the provider
let capturedContext: FileContextType | null = null;

function ContextCapture() {
    const { useFile } = require('./fileContext');
    capturedContext = useFile();
    return <div data-testid="context-ready">ready</div>;
}

// helpers
const createNode = (
    id: number,
    overrides?: Partial<FileData>,
    children?: FileData[]
): FileData => ({
    id,
    name: `node-${id}`,
    path: `/node-${id}`,
    parent_path: '/',
    type: 1,
    format: 'folder',
    size: 0,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    file_children: children,
    ...overrides,
});

const makePaginationResponse = (
    items: FileData[],
    hasNext = false,
    page = 1
): PaginationResponse => ({
    items,
    pagination: { hasNext, hasPrevious: false, page, pageSize: 200 },
});

function createQueryClient() {
    return new QueryClient({
        defaultOptions: {
            queries: { retry: false, gcTime: 0 },
            mutations: { retry: false },
        },
    });
}

function renderProvider(queryClient?: QueryClient) {
    const qc = queryClient ?? createQueryClient();
    const FileProvider = require('./index').default;

    return render(
        <QueryClientProvider client={qc}>
            <FileProvider>
                <ContextCapture />
            </FileProvider>
        </QueryClientProvider>
    );
}

// ========== UTILITY TESTS ==========

describe('fileProvider utilities', () => {
    it('extracts paths from the file route', () => {
        expect(extractFilePath('/files')).toBe('');
        expect(extractFilePath('/files/')).toBe('');
        expect(extractFilePath('/files/docs/new')).toBe('/docs/new');
        expect(extractFilePath('/other')).toBe('');
    });

    it('builds encoded file URLs', () => {
        expect(buildFilesUrl('docs/new')).toBe('/files/docs/new');
        expect(buildFilesUrl('folder name/file')).toBe('/files/folder%20name/file');
        expect(buildFilesUrl('')).toBe('/files');
    });

    it('finds items in nested trees', () => {
        const tree: FileData[] = [
            createNode(1, {}, [createNode(2), createNode(3, {}, [createNode(4)])]),
            createNode(5),
        ];

        expect(findItemInTree(tree, 3)?.id).toBe(3);
        expect(findItemInTree(tree, 4)?.id).toBe(4);
        expect(findItemInTree(tree, 99)).toBeNull();
    });

    it('returns null when itemId is null or 0', () => {
        const tree: FileData[] = [createNode(1)];
        expect(findItemInTree(tree, null)).toBeNull();
        expect(findItemInTree(tree, 0)).toBeNull();
    });

    it('returns null when children exist but target is not found in them', () => {
        // This exercises the branch where file_children exists and has items,
        // but findItemInTree recurse returns null (itemChildren is falsy)
        const tree: FileData[] = [
            createNode(1, {}, [createNode(2), createNode(3)]),
            createNode(4, {}, [createNode(5)]),
        ];
        expect(findItemInTree(tree, 99)).toBeNull();
    });

    it('skips children search when file_children is empty', () => {
        const tree: FileData[] = [createNode(1, {}, [])];
        expect(findItemInTree(tree, 99)).toBeNull();
    });

    it('adds children to the correct parent node', () => {
        const tree: FileData[] = [createNode(1), createNode(2)];
        const updated = addChildrenToTree(tree, 2, [createNode(6)]);
        expect(updated[1]?.file_children?.[0]?.id).toBe(6);
    });

    it('recurses into children when adding to a deeply nested parent', () => {
        const tree: FileData[] = [createNode(1, {}, [createNode(2, {}, [createNode(3)])])];
        const updated = addChildrenToTree(tree, 3, [createNode(10)]);
        expect(updated[0]?.file_children?.[0]?.file_children?.[0]?.file_children?.[0]?.id).toBe(
            10
        );
    });

    it('returns the node unchanged when parentId does not match and no children', () => {
        const tree: FileData[] = [createNode(1), createNode(2)];
        const updated = addChildrenToTree(tree, 99, [createNode(6)]);
        expect(updated[0]?.file_children).toBeUndefined();
        expect(updated[1]?.file_children).toBeUndefined();
    });

    it('builds the trail of nodes to a target id', () => {
        const nested = [createNode(1, {}, [createNode(2, {}, [createNode(3)])]), createNode(4)];
        const trail = findTrailByIdInTree(nested, 3);
        expect(trail?.map((node) => node.id)).toEqual([1, 2, 3]);
        expect(findTrailByIdInTree(nested, 99)).toBeNull();
    });

    it('returns single-element trail for a root node', () => {
        const nodes = [createNode(5)];
        const trail = findTrailByIdInTree(nodes, 5);
        expect(trail?.map((n) => n.id)).toEqual([5]);
    });

    it('decodes encoded path segments in extractFilePath', () => {
        expect(extractFilePath('/files/my%20folder/sub')).toBe('/my folder/sub');
    });

    it('buildFilesUrl handles leading slash correctly', () => {
        expect(buildFilesUrl('/docs/readme')).toBe('/files/docs/readme');
    });
});

// ========== useFile error test ==========

describe('useFile', () => {
    it('throws when used outside of FileProvider', () => {
        const { useFile } = require('./fileContext');
        const ErrorComponent = () => {
            useFile();
            return null;
        };
        expect(() => render(<ErrorComponent />)).toThrow(
            'useFile must be used within a FileProvider'
        );
    });
});

// ========== PROVIDER INTEGRATION TESTS ==========

describe('FileProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        capturedContext = null;
        mockPathname = '/files';
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
        mockGetRecentAccessByFileId.mockResolvedValue([]);
        mockGetFileByPath.mockResolvedValue(null);
        mockToggleStarredFile.mockResolvedValue(undefined);
        mockRescanFiles.mockResolvedValue(undefined);
        mockUploadFilesToPath.mockResolvedValue(undefined);
        mockCreateFolderAtPath.mockResolvedValue(undefined);
        mockMoveFilePath.mockResolvedValue(undefined);
        mockCopyFilePath.mockResolvedValue(undefined);
        mockRenameFilePath.mockResolvedValue(undefined);
        mockDeleteFilePath.mockResolvedValue(undefined);
    });

    it('renders children and provides context at root path', async () => {
        renderProvider();
        await waitFor(() => expect(screen.getByTestId('context-ready')).toBeInTheDocument());
        expect(capturedContext).not.toBeNull();
        expect(capturedContext!.selectedItem).toBeNull();
        expect(capturedContext!.files).toEqual([]);
        expect(capturedContext!.expandedItems).toEqual([]);
    });

    it('fetches the file tree on mount at root', async () => {
        const items = [createNode(1), createNode(2)];
        mockGetFilesTree.mockResolvedValue(makePaginationResponse(items));

        renderProvider();
        await waitFor(() => expect(capturedContext!.files.length).toBe(2));
        expect(capturedContext!.files[0]?.id).toBe(1);
    });

    it('resolves selected item from URL path', async () => {
        mockPathname = '/files/docs';
        const resolved = createNode(10, { name: 'docs', path: '/docs', type: 1 });
        mockGetFileByPath.mockResolvedValue(resolved);
        const childItems = [
            createNode(20, { name: 'readme.txt', path: '/docs/readme.txt', type: 2 }),
        ];
        mockGetFilesTree.mockResolvedValue(makePaginationResponse(childItems));

        renderProvider();
        await waitFor(() => expect(capturedContext!.selectedItem).not.toBeNull());
        expect(mockGetFileByPath).toHaveBeenCalledWith('/docs');
    });

    it('does not call getFileByPath when at root', async () => {
        mockPathname = '/files';
        renderProvider();
        await waitFor(() => expect(screen.getByTestId('context-ready')).toBeInTheDocument());
        expect(mockGetFileByPath).not.toHaveBeenCalled();
    });

    it('handleSelectItem navigates to the file URL', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        act(() => {
            capturedContext!.handleSelectItem(createNode(5, { path: '/my docs/report' }));
        });
        expect(mockNavigate).toHaveBeenCalledWith('/files/my%20docs/report');
    });

    it('handleSelectItem navigates to root when null', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        act(() => {
            capturedContext!.handleSelectItem(null);
        });
        expect(mockNavigate).toHaveBeenCalledWith('/files');
    });

    it('handleStarredItem calls toggleStarredFile and refetches', async () => {
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([createNode(1)]));
        renderProvider();
        await waitFor(() => expect(capturedContext!.files.length).toBe(1));

        await act(async () => {
            capturedContext!.handleStarredItem(1);
        });
        await waitFor(() => expect(mockToggleStarredFile).toHaveBeenCalledWith(1));
    });

    it('createFolder calls service and refetches', async () => {
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.createFolder('new-folder', '/parent');
        });
        expect(mockCreateFolderAtPath).toHaveBeenCalledWith('new-folder', '/parent');
    });

    it('deletePath calls service and refetches', async () => {
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.deletePath('/old-file');
        });
        expect(mockDeleteFilePath).toHaveBeenCalledWith('/old-file');
    });

    it('movePath calls service with correct args', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.movePath('/src', '/dest');
        });
        expect(mockMoveFilePath).toHaveBeenCalledWith('/src', '/dest');
    });

    it('copyPath calls service with correct args', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.copyPath('/src', '/dest');
        });
        expect(mockCopyFilePath).toHaveBeenCalledWith('/src', '/dest');
    });

    it('renamePath calls service with correct args', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.renamePath('/old', 'new-name');
        });
        expect(mockRenameFilePath).toHaveBeenCalledWith('/old', 'new-name');
    });

    it('uploadFiles calls service with correct args', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        const files = { length: 0, item: () => null } as unknown as FileList;
        await act(async () => {
            await capturedContext!.uploadFiles(files, '/target');
        });
        expect(mockUploadFilesToPath).toHaveBeenCalledWith(files, '/target');
    });

    it('rescanFiles calls service and refetches', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());

        await act(async () => {
            await capturedContext!.rescanFiles();
        });
        expect(mockRescanFiles).toHaveBeenCalled();
    });

    it('setFileListFilter updates the filter state', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());
        expect(capturedContext!.fileListFilter).toBe('all');

        act(() => {
            capturedContext!.setFileListFilter('starred');
        });
        await waitFor(() => expect(capturedContext!.fileListFilter).toBe('starred'));
    });

    it('provides default status as pending initially', async () => {
        mockGetFilesTree.mockReturnValue(new Promise(() => {}));
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());
        expect(capturedContext!.status).toBe('pending');
    });

    it('status becomes success after data loads', async () => {
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
        renderProvider();
        await waitFor(() => expect(capturedContext!.status).toBe('success'));
    });

    it('expandedItems is empty when no item is selected', async () => {
        mockPathname = '/files';
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([createNode(1)]));

        renderProvider();
        await waitFor(() => expect(capturedContext!.files.length).toBe(1));
        expect(capturedContext!.expandedItems).toEqual([]);
    });

    it('expandedItems contains trail ids when selected item is in tree', async () => {
        // Build a tree at root that contains node 10 with child 20
        mockPathname = '/files';
        const rootTree = [
            createNode(10, { name: 'docs', path: '/docs', type: 1 }, [
                createNode(20, { name: 'sub', path: '/docs/sub', type: 1 }),
            ]),
        ];
        mockGetFilesTree.mockResolvedValue(makePaginationResponse(rootTree));
        mockGetFileByPath.mockResolvedValue(null);

        const qc = createQueryClient();
        const FileProvider = require('./index').default;

        const { rerender } = render(
            <QueryClientProvider client={qc}>
                <FileProvider>
                    <ContextCapture />
                </FileProvider>
            </QueryClientProvider>
        );

        // Wait for the tree to load at root
        await waitFor(() => expect(capturedContext!.files.length).toBe(1));
        expect(capturedContext!.files[0]?.id).toBe(10);

        // Now navigate to a child path - mock the path resolution
        mockPathname = '/files/docs/sub';
        mockGetFileByPath.mockResolvedValue(
            createNode(20, { name: 'sub', path: '/docs/sub', type: 1 })
        );
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));

        rerender(
            <QueryClientProvider client={qc}>
                <FileProvider>
                    <ContextCapture />
                </FileProvider>
            </QueryClientProvider>
        );

        await waitFor(() => {
            expect(capturedContext!.expandedItems).toContain(10);
            expect(capturedContext!.expandedItems).toContain(20);
        });
    });

    it('isLoadingAccessData is provided in context', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());
        expect(typeof capturedContext!.isLoadingAccessData).toBe('boolean');
    });

    it('recentAccessFiles defaults to empty array', async () => {
        renderProvider();
        await waitFor(() => expect(capturedContext).not.toBeNull());
        expect(capturedContext!.recentAccessFiles).toEqual([]);
    });

    it('fetches recent access data when selected item is a file', async () => {
        mockPathname = '/files/readme.txt';
        // type: 2 = FileType.File
        const resolved = createNode(10, {
            name: 'readme.txt',
            path: '/readme.txt',
            type: 2,
        });
        mockGetFileByPath.mockResolvedValue(resolved);
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));
        const accessData = [
            {
                id: 1,
                ip_address: '127.0.0.1',
                file_id: 10,
                accessed_at: '2026-01-01',
            },
        ];
        mockGetRecentAccessByFileId.mockResolvedValue(accessData);

        renderProvider();
        await waitFor(() => expect(capturedContext!.recentAccessFiles.length).toBe(1));
        expect(mockGetRecentAccessByFileId).toHaveBeenCalledWith(10);
    });

    it('does not fetch recent access data when selected item is a directory', async () => {
        mockPathname = '/files/docs';
        // type: 1 = FileType.Directory
        const resolved = createNode(10, { name: 'docs', path: '/docs', type: 1 });
        mockGetFileByPath.mockResolvedValue(resolved);
        mockGetFilesTree.mockResolvedValue(makePaginationResponse([]));

        renderProvider();
        await waitFor(() => expect(capturedContext!.selectedItem).not.toBeNull());
        // Allow queries to settle
        await new Promise((r) => setTimeout(r, 100));
        expect(mockGetRecentAccessByFileId).not.toHaveBeenCalled();
    });

    it('supports pagination when hasNext is true', async () => {
        const firstPage = [createNode(1), createNode(2)];
        mockGetFilesTree.mockResolvedValueOnce(makePaginationResponse(firstPage, true, 1));

        renderProvider();
        await waitFor(() => expect(capturedContext!.files.length).toBe(2));
        // The infinite query should have been called once, and because hasNext=true
        // the getNextPageParam should return 2
        expect(mockGetFilesTree).toHaveBeenCalled();
    });

    it('builds effectiveSelectedItem with children from snapshot when item is a directory', async () => {
        mockPathname = '/files/docs';
        // type 1 = Directory
        const resolved = createNode(10, { name: 'docs', path: '/docs', type: 1 });
        mockGetFileByPath.mockResolvedValue(resolved);

        const childItems = [createNode(20), createNode(30)];
        mockGetFilesTree.mockResolvedValue(makePaginationResponse(childItems));

        renderProvider();
        await waitFor(() => {
            expect(capturedContext!.selectedItem).not.toBeNull();
            expect(capturedContext!.selectedItem!.file_children).toBeDefined();
            expect(capturedContext!.selectedItem!.file_children!.length).toBe(2);
        });
    });

    it('updates file tree when data has selected item with children', async () => {
        mockPathname = '/files/docs';
        const resolved = createNode(10, { name: 'docs', path: '/docs', type: 1 });
        mockGetFileByPath.mockResolvedValue(resolved);

        const childItems = [createNode(20, { name: 'file.txt', path: '/docs/file.txt' })];
        mockGetFilesTree.mockResolvedValue(makePaginationResponse(childItems));

        renderProvider();
        await waitFor(() => expect(capturedContext!.selectedItem).not.toBeNull());
        // The effective selected item should include children from the data
        expect(capturedContext!.selectedItem!.file_children).toBeDefined();
    });
});
