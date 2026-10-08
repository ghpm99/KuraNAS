import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import CreateFolderDialog from './createFolderDialog';

const nameInput = () => screen.getByLabelText('NAME') as HTMLInputElement;
const confirmButton = () => screen.getByRole('button', { name: 'NEW_FOLDER' });

describe('CreateFolderDialog', () => {
    it('renders closed without any provider', () => {
        render(<CreateFolderDialog isOpen={false} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('opens with the input focused and confirm disabled but no error shown yet', () => {
        render(<CreateFolderDialog isOpen onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(nameInput()).toHaveFocus();
        expect(confirmButton()).toBeDisabled();
        expect(screen.queryByText('FILES_NAME_ERROR_EMPTY')).not.toBeInTheDocument();
    });

    it('flags an emptied name and path separators after typing', () => {
        render(<CreateFolderDialog isOpen onClose={jest.fn()} onConfirm={jest.fn()} />);

        fireEvent.change(nameInput(), { target: { value: 'a/b' } });
        expect(screen.getByText('FILES_NAME_ERROR_INVALID_CHARACTERS')).toBeInTheDocument();
        expect(confirmButton()).toBeDisabled();

        fireEvent.change(nameInput(), { target: { value: '' } });
        expect(screen.getByText('FILES_NAME_ERROR_EMPTY')).toBeInTheDocument();
    });

    it('submits the trimmed name on enter and keeps the dialog on success handling to the caller', async () => {
        const onConfirm = jest.fn().mockResolvedValue({ errorMessage: null });
        render(<CreateFolderDialog isOpen onClose={jest.fn()} onConfirm={onConfirm} />);

        fireEvent.change(nameInput(), { target: { value: ' Docs ' } });
        fireEvent.submit(nameInput().closest('form') as HTMLFormElement);

        await waitFor(() => expect(onConfirm).toHaveBeenCalledWith('Docs'));
    });

    it('shows the backend conflict message verbatim and stays open', async () => {
        const onConfirm = jest.fn().mockResolvedValue({ errorMessage: 'Pasta já existe' });
        render(<CreateFolderDialog isOpen onClose={jest.fn()} onConfirm={onConfirm} />);

        fireEvent.change(nameInput(), { target: { value: 'Docs' } });
        fireEvent.click(confirmButton());

        expect(await screen.findByText('Pasta já existe')).toBeInTheDocument();
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
});
