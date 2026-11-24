import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { createEmptyPlaylistRules } from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';
import PlaylistCreateDialog from './PlaylistCreateDialog';

const renderDialog = (
    overrides: Partial<React.ComponentProps<typeof PlaylistCreateDialog>> = {}
) => {
    const props = {
        open: true,
        newName: 'Rock mix',
        newDescription: '',
        newKind: 'manual' as const,
        newRules: createEmptyPlaylistRules(),
        errorField: null,
        isSubmitting: false,
        onKindChange: jest.fn(),
        onRulesChange: jest.fn(),
        onClose: jest.fn(),
        onNameChange: jest.fn(),
        onDescriptionChange: jest.fn(),
        onSubmit: jest.fn(),
        ...overrides,
    };
    renderWithoutBackend(<PlaylistCreateDialog {...props} />);
    return props;
};

describe('PlaylistCreateDialog', () => {
    it('renders without any backend', () => {
        expect(() => renderDialog()).not.toThrow();
        expect(() => renderDialog({ newKind: 'rule' })).not.toThrow();
    });

    it('hides the rule form for a manual playlist and submits', () => {
        const { onSubmit } = renderDialog();

        expect(screen.queryByLabelText('MUSIC_PLAYLIST_RULE_GENRES')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE' }));

        expect(onSubmit).toHaveBeenCalledTimes(1);
    });

    it('switches kind through the toggle', () => {
        const { onKindChange } = renderDialog();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_KIND_RULE' }));

        expect(onKindChange).toHaveBeenCalledWith('rule');
    });

    it('shows the rule form and blocks submit until a criterion is chosen', () => {
        renderDialog({ newKind: 'rule' });

        expect(screen.getByLabelText('MUSIC_PLAYLIST_RULE_GENRES')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'ACTION_CREATE' })).toBeDisabled();
    });

    it('allows submit once the rule has a criterion', () => {
        renderDialog({
            newKind: 'rule',
            newRules: { ...createEmptyPlaylistRules(), favorites_only: true },
        });

        expect(screen.getByRole('button', { name: 'ACTION_CREATE' })).toBeEnabled();
    });

    it('blocks submit with a blank name', () => {
        renderDialog({ newName: '  ' });

        expect(screen.getByRole('button', { name: 'ACTION_CREATE' })).toBeDisabled();
    });
});
