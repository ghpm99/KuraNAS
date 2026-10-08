import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileSearchBar from './FileSearchBar';

describe('FileSearchBar', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FileSearchBar />);
    });

    it('uses the folder placeholder and shows the subfolders toggle in folder scope', () => {
        render(<FileSearchBar isFolderScope />);

        expect(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER')).toBeInTheDocument();
        expect(screen.getByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS')).toBeChecked();
    });

    it('uses the global placeholder and hides the toggle outside a folder', () => {
        render(<FileSearchBar />);

        expect(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER_ALL')).toBeInTheDocument();
        expect(screen.queryByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS')).toBeNull();
    });

    it('reports typing, the recursive toggle and the clear button', () => {
        const onChange = jest.fn();
        const onClear = jest.fn();
        const onRecursiveChange = jest.fn();
        render(
            <FileSearchBar
                value="rel"
                onChange={onChange}
                onClear={onClear}
                isFolderScope
                isRecursive
                onRecursiveChange={onRecursiveChange}
            />
        );

        fireEvent.change(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER'), {
            target: { value: 'relatorio' },
        });
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS'));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_CLEAR' }));

        expect(onChange).toHaveBeenCalledWith('relatorio');
        expect(onRecursiveChange).toHaveBeenCalledWith(false);
        expect(onClear).toHaveBeenCalled();
    });

    it('does not show the clear button while the field is empty', () => {
        render(<FileSearchBar value="" />);

        expect(screen.queryByRole('button', { name: 'FILES_SEARCH_CLEAR' })).toBeNull();
    });

    it('tolerates handlers being absent', () => {
        render(<FileSearchBar value="rel" isFolderScope />);

        fireEvent.change(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER'), {
            target: { value: 'x' },
        });
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_INCLUDE_SUBFOLDERS'));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_CLEAR' }));
    });

    it('hands the input element to the given ref', () => {
        const inputRef = { current: null as HTMLInputElement | null };
        render(<FileSearchBar inputRef={inputRef} />);

        expect(inputRef.current).toBe(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER_ALL'));
    });
});
