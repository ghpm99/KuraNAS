import { fireEvent, render, screen } from '@testing-library/react';
import FileViewerNavigation from './fileViewerNavigation';

const mockUseFile = jest.fn();
const mockHandleSelectItem = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options ? `${key}:${options.position}/${options.total}` : key,
    }),
}));

const createFile = (id: number, overrides = {}) =>
    ({
        id,
        name: `f${id}.txt`,
        path: `/docs/f${id}.txt`,
        parent_path: '/docs',
        type: 2,
        format: '.txt',
        size: 1,
        ...overrides,
    }) as any;

const folderWithFiles = (...files: any[]) => [
    { id: 100, name: 'docs', path: '/docs', parent_path: '/', type: 1, file_children: files },
];

describe('FileViewerNavigation', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue({ files: [], handleSelectItem: mockHandleSelectItem });
    });

    it('renders nothing when the siblings were not loaded', () => {
        const { container } = render(<FileViewerNavigation file={createFile(2)} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('survives a context without a loaded tree', () => {
        mockUseFile.mockReturnValue({ handleSelectItem: mockHandleSelectItem });

        const { container } = render(<FileViewerNavigation file={createFile(2)} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('navigates to the previous and next sibling with the buttons', () => {
        mockUseFile.mockReturnValue({
            files: folderWithFiles(createFile(1), createFile(2), createFile(3)),
            handleSelectItem: mockHandleSelectItem,
        });
        render(<FileViewerNavigation file={createFile(2)} />);

        expect(screen.getByText('FILE_VIEWER_POSITION:2/3')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILE_VIEWER_PREVIOUS' }));
        expect(mockHandleSelectItem).toHaveBeenLastCalledWith(expect.objectContaining({ id: 1 }));

        fireEvent.click(screen.getByRole('button', { name: 'FILE_VIEWER_NEXT' }));
        expect(mockHandleSelectItem).toHaveBeenLastCalledWith(expect.objectContaining({ id: 3 }));
    });

    it('disables the buttons at both ends of the listing', () => {
        mockUseFile.mockReturnValue({
            files: folderWithFiles(createFile(1), createFile(2)),
            handleSelectItem: mockHandleSelectItem,
        });
        const { rerender } = render(<FileViewerNavigation file={createFile(1)} />);

        expect(screen.getByRole('button', { name: 'FILE_VIEWER_PREVIOUS' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'FILE_VIEWER_NEXT' })).toBeEnabled();

        rerender(<FileViewerNavigation file={createFile(2)} />);

        expect(screen.getByRole('button', { name: 'FILE_VIEWER_PREVIOUS' })).toBeEnabled();
        expect(screen.getByRole('button', { name: 'FILE_VIEWER_NEXT' })).toBeDisabled();
    });

    it('navigates with the arrow keys and stops at the ends', () => {
        mockUseFile.mockReturnValue({
            files: folderWithFiles(createFile(1), createFile(2), createFile(3)),
            handleSelectItem: mockHandleSelectItem,
        });
        const { rerender } = render(<FileViewerNavigation file={createFile(2)} />);

        fireEvent.keyDown(window, { key: 'ArrowRight' });
        expect(mockHandleSelectItem).toHaveBeenLastCalledWith(expect.objectContaining({ id: 3 }));

        fireEvent.keyDown(window, { key: 'ArrowLeft' });
        expect(mockHandleSelectItem).toHaveBeenLastCalledWith(expect.objectContaining({ id: 1 }));

        mockHandleSelectItem.mockClear();
        rerender(<FileViewerNavigation file={createFile(1)} />);
        fireEvent.keyDown(window, { key: 'ArrowLeft' });
        expect(mockHandleSelectItem).not.toHaveBeenCalled();

        rerender(<FileViewerNavigation file={createFile(3)} />);
        fireEvent.keyDown(window, { key: 'ArrowRight' });
        expect(mockHandleSelectItem).not.toHaveBeenCalled();
    });

    it('ignores arrows with modifiers, other keys and typing in inputs or media controls', () => {
        mockUseFile.mockReturnValue({
            files: folderWithFiles(createFile(1), createFile(2), createFile(3)),
            handleSelectItem: mockHandleSelectItem,
        });
        render(
            <>
                <FileViewerNavigation file={createFile(2)} />
                <input aria-label="field" />
                <video aria-label="player" />
            </>
        );

        fireEvent.keyDown(window, { key: 'ArrowRight', ctrlKey: true });
        fireEvent.keyDown(window, { key: 'ArrowRight', shiftKey: true });
        fireEvent.keyDown(window, { key: 'Enter' });
        fireEvent.keyDown(screen.getByLabelText('field'), { key: 'ArrowRight' });
        fireEvent.keyDown(screen.getByLabelText('player'), { key: 'ArrowLeft' });

        expect(mockHandleSelectItem).not.toHaveBeenCalled();
    });

    it('stops listening to the keyboard after unmounting', () => {
        mockUseFile.mockReturnValue({
            files: folderWithFiles(createFile(1), createFile(2)),
            handleSelectItem: mockHandleSelectItem,
        });
        const { unmount } = render(<FileViewerNavigation file={createFile(1)} />);

        unmount();
        fireEvent.keyDown(window, { key: 'ArrowRight' });

        expect(mockHandleSelectItem).not.toHaveBeenCalled();
    });
});
