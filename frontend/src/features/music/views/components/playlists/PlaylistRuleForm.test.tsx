import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { createEmptyPlaylistRules } from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';
import PlaylistRuleForm from './PlaylistRuleForm';

describe('PlaylistRuleForm', () => {
    it('renders without any backend', () => {
        expect(() =>
            renderWithoutBackend(
                <PlaylistRuleForm rules={createEmptyPlaylistRules()} onChange={jest.fn()} />
            )
        ).not.toThrow();
    });

    it('reports every control change as a complete rules object', () => {
        const onChange = jest.fn();
        renderWithoutBackend(
            <PlaylistRuleForm rules={createEmptyPlaylistRules()} onChange={onChange} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_RULE_MATCH_ALL' }));
        expect(onChange).toHaveBeenLastCalledWith({ ...createEmptyPlaylistRules(), match: 'all' });

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_RULE_YEAR_FROM'), {
            target: { value: '1990' },
        });
        expect(onChange).toHaveBeenLastCalledWith({
            ...createEmptyPlaylistRules(),
            year_from: 1990,
        });

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_RULE_YEAR_TO'), {
            target: { value: '2000' },
        });
        expect(onChange).toHaveBeenLastCalledWith({ ...createEmptyPlaylistRules(), year_to: 2000 });

        fireEvent.click(screen.getByLabelText('MUSIC_PLAYLIST_RULE_FAVORITES_ONLY'));
        expect(onChange).toHaveBeenLastCalledWith({
            ...createEmptyPlaylistRules(),
            favorites_only: true,
        });
    });

    it('clears a year when the field is emptied', () => {
        const onChange = jest.fn();
        renderWithoutBackend(
            <PlaylistRuleForm
                rules={{ ...createEmptyPlaylistRules(), year_from: 1990 }}
                onChange={onChange}
            />
        );

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_RULE_YEAR_FROM'), {
            target: { value: '' },
        });

        expect(onChange).toHaveBeenLastCalledWith({
            ...createEmptyPlaylistRules(),
            year_from: null,
        });
    });

    it('asks for a criterion instead of previewing an empty rule', () => {
        renderWithoutBackend(
            <PlaylistRuleForm rules={createEmptyPlaylistRules()} onChange={jest.fn()} />
        );

        expect(screen.getByRole('status')).toHaveTextContent('MUSIC_PLAYLIST_RULE_PREVIEW_IDLE');
    });
});
