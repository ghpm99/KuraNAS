import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import PlaylistEditDialog from './PlaylistEditDialog';

const renderDialog = (overrides: Partial<React.ComponentProps<typeof PlaylistEditDialog>> = {}) => {
    const props = {
        currentName: 'Mix',
        currentDescription: 'old',
        isSubmitting: false,
        onClose: jest.fn(),
        onSubmit: jest.fn(),
        ...overrides,
    };
    renderWithoutBackend(<PlaylistEditDialog {...props} />);
    return props;
};

describe('PlaylistEditDialog', () => {
    it('renders without any backend', () => {
        expect(() => renderDialog()).not.toThrow();
    });

    it('starts from the current values and submits the trimmed edits', () => {
        const { onSubmit } = renderDialog();
        expect(screen.getByLabelText('NAME')).toHaveValue('Mix');

        fireEvent.change(screen.getByLabelText('NAME'), { target: { value: '  Road trip ' } });
        fireEvent.change(screen.getByLabelText('MUSIC_DESCRIPTION_OPTIONAL'), {
            target: { value: ' long drives ' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_SAVE' }));

        expect(onSubmit).toHaveBeenCalledWith('Road trip', 'long drives');
    });

    it('blocks saving a blank name and while submitting', () => {
        renderDialog();
        fireEvent.change(screen.getByLabelText('NAME'), { target: { value: '   ' } });
        expect(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_SAVE' })).toBeDisabled();
    });

    it('disables saving while a submit is in flight and cancels through the close handler', () => {
        const { onClose } = renderDialog({ isSubmitting: true });
        expect(screen.queryByRole('button', { name: 'MUSIC_PLAYLIST_SAVE' })).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onClose).toHaveBeenCalled();
    });
});
