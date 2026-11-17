import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { IMusicData } from '@/types/music';
import TrackSelectionControls from './TrackSelectionControls';
import { buildTracks } from './trackSelectionTestFactory';
import { useTrackSelection } from './useTrackSelection';

const ControlsOverLoadedTracks = ({ loadedTracks }: { loadedTracks: IMusicData[] }) => {
    const selection = useTrackSelection('list');
    return (
        <>
            <input aria-label="free-text" />
            <TrackSelectionControls selection={selection} loadedTracks={loadedTracks} />
            <span data-testid="selected-ids">{selection.selectedFileIds.join(',')}</span>
        </>
    );
};

const enterSelectionMode = () =>
    fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECT_TRACKS' }));

const selectAllLoaded = () =>
    fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_SELECT_ALL_LOADED' }));

describe('TrackSelectionControls', () => {
    it('offers the select action first and the toolbar only after entering the mode', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(2)} />);
        expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();

        enterSelectionMode();

        expect(screen.getByRole('toolbar', { name: 'MUSIC_SELECTION_TOOLBAR' })).toBeVisible();
        expect(screen.getByText('MUSIC_SELECTION_COUNT')).toBeVisible();
        expect(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'MUSIC_SELECTION_CLEAR' })).toBeDisabled();
    });

    it('selects all loaded tracks, enables editing and clears them', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(3)} />);
        enterSelectionMode();

        selectAllLoaded();
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('1,2,3');
        expect(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' })).toBeEnabled();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_CLEAR' }));
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('');
        expect(screen.getByRole('toolbar')).toBeVisible();
    });

    it('leaves the mode on Escape and on the cancel action', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(2)} />);

        enterSelectionMode();
        selectAllLoaded();
        fireEvent.keyDown(document, { key: 'Escape' });
        expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('');

        enterSelectionMode();
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_CANCEL' }));
        expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
    });

    it('disables editing above 500 tracks and explains why', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(501)} />);
        enterSelectionMode();

        selectAllLoaded();

        expect(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' })).toBeDisabled();
        expect(screen.getByRole('alert')).toHaveTextContent('MUSIC_SELECTION_LIMIT_EXCEEDED');
    });

    it('opens the batch dialog for the selected ids and keeps the mode when it is cancelled', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(2)} />);
        enterSelectionMode();
        selectAllLoaded();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' }));
        expect(screen.getByRole('dialog')).toHaveTextContent('MUSIC_EDIT_SELECTION');

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('1,2');
    });

    it('selects every loaded track on Ctrl+A and Cmd+A while the mode is on', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(3)} />);
        enterSelectionMode();

        fireEvent.keyDown(document.body, { key: 'a', ctrlKey: true });
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('1,2,3');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SELECTION_CLEAR' }));
        fireEvent.keyDown(document.body, { key: 'a', metaKey: true });
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('1,2,3');
    });

    it('does not select all on Ctrl+A typed inside a text field or outside the mode', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(3)} />);

        fireEvent.keyDown(document.body, { key: 'a', ctrlKey: true });
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('');

        enterSelectionMode();
        fireEvent.keyDown(screen.getByLabelText('free-text'), { key: 'a', ctrlKey: true });
        expect(screen.getByTestId('selected-ids')).toHaveTextContent('');
    });

    it('keeps editing disabled when Ctrl+A selects more than 500 loaded tracks', () => {
        renderWithoutBackend(<ControlsOverLoadedTracks loadedTracks={buildTracks(501)} />);
        enterSelectionMode();

        fireEvent.keyDown(document.body, { key: 'a', ctrlKey: true });

        expect(screen.getByRole('button', { name: 'MUSIC_SELECTION_EDIT' })).toBeDisabled();
        expect(screen.getByRole('alert')).toHaveTextContent('MUSIC_SELECTION_LIMIT_EXCEEDED');
    });
});
