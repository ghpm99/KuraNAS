import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import RenameFileDialog from './renameFileDialog';

const reportFile = createTestFile(7, { name: 'report.final.pdf' });
const photosFolder = createTestFile(8, { name: 'photos.2024', type: 1 });

const nameInput = () => screen.getByLabelText('NAME') as HTMLInputElement;

describe('RenameFileDialog', () => {
    it('renders closed with no file and without any provider', () => {
        render(<RenameFileDialog file={null} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('focuses the input and pre-selects only the base name of a file', () => {
        render(<RenameFileDialog file={reportFile} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(nameInput()).toHaveFocus();
        expect(nameInput().selectionStart).toBe(0);
        expect(nameInput().selectionEnd).toBe('report.final'.length);
    });

    it('pre-selects the whole name of a folder', () => {
        render(<RenameFileDialog file={photosFolder} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(nameInput().selectionEnd).toBe('photos.2024'.length);
    });

    it('disables confirm with a hint while the name is unchanged', () => {
        render(<RenameFileDialog file={reportFile} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.getByRole('button', { name: 'RENAME' })).toBeDisabled();
        expect(screen.getByText('FILES_NAME_HINT_UNCHANGED')).toBeInTheDocument();
    });

    it.each([
        ['', 'FILES_NAME_ERROR_EMPTY'],
        ['a/b.pdf', 'FILES_NAME_ERROR_INVALID_CHARACTERS'],
        ['a\\b.pdf', 'FILES_NAME_ERROR_INVALID_CHARACTERS'],
    ])('blocks %j with helper text', (typedName, expectedMessageKey) => {
        render(<RenameFileDialog file={reportFile} onClose={jest.fn()} onConfirm={jest.fn()} />);

        fireEvent.change(nameInput(), { target: { value: typedName } });

        expect(screen.getByText(expectedMessageKey)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'RENAME' })).toBeDisabled();
    });

    it('submits the trimmed new name', async () => {
        const onConfirm = jest.fn().mockResolvedValue({ errorMessage: null });
        render(<RenameFileDialog file={reportFile} onClose={jest.fn()} onConfirm={onConfirm} />);

        fireEvent.change(nameInput(), { target: { value: '  summary.pdf ' } });
        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));

        await waitFor(() => expect(onConfirm).toHaveBeenCalledWith(reportFile, 'summary.pdf'));
    });

    it('shows the backend error verbatim, stays open and clears it when typing again', async () => {
        const onConfirm = jest
            .fn()
            .mockResolvedValue({ errorMessage: 'Já existe um item com este nome' });
        render(<RenameFileDialog file={reportFile} onClose={jest.fn()} onConfirm={onConfirm} />);

        fireEvent.change(nameInput(), { target: { value: 'taken.pdf' } });
        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));

        expect(await screen.findByText('Já existe um item com este nome')).toBeInTheDocument();
        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'RENAME' })).toBeEnabled();

        fireEvent.change(nameInput(), { target: { value: 'free.pdf' } });
        expect(screen.queryByText('Já existe um item com este nome')).not.toBeInTheDocument();
    });

    it('cancels through the cancel button', () => {
        const onClose = jest.fn();
        render(<RenameFileDialog file={reportFile} onClose={onClose} onConfirm={jest.fn()} />);

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onClose).toHaveBeenCalledTimes(1);
    });
});
