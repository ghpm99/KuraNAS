import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import TrackSelectionControls from './TrackSelectionControls';
import { useTrackSelection } from './useTrackSelection';

const ControlsOverEmptyList = () => (
    <TrackSelectionControls selection={useTrackSelection('empty')} loadedTracks={[]} />
);

describe('TrackSelectionControls with no service mock and no backend', () => {
    it('mounts over an empty list without offering to select', () => {
        renderWithoutBackend(<ControlsOverEmptyList />);

        expect(screen.queryByRole('button')).not.toBeInTheDocument();
        expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
    });
});
