import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import VacuumVolumePreference from './VacuumVolumePreference';

const renderVolume = (onCommit = jest.fn().mockResolvedValue(undefined), volume = 4) => {
    const { container } = render(
        <VacuumVolumePreference label="Volume" volume={volume} isDisabled={false} onCommit={onCommit} />
    );
    const sliderTrack = container.querySelector('.MuiSlider-root') as HTMLElement;
    sliderTrack.getBoundingClientRect = () =>
        ({ left: 0, width: 100, top: 0, height: 10, right: 100, bottom: 10 }) as DOMRect;
    return { onCommit, sliderTrack };
};

describe('VacuumVolumePreference', () => {
    it('renders without any provider or service showing the current volume', () => {
        renderVolume();

        expect(screen.getByRole('slider', { name: 'Volume' })).toHaveValue('4');
    });

    it('does not commit while the slider is dragged, only when it is released', async () => {
        const { onCommit, sliderTrack } = renderVolume();
        const slider = screen.getByRole('slider', { name: 'Volume' });

        fireEvent.mouseDown(sliderTrack, { clientX: 70, buttons: 1 });
        expect(slider).toHaveValue('7');
        expect(onCommit).not.toHaveBeenCalled();

        fireEvent.mouseUp(document);
        await waitFor(() => expect(onCommit).toHaveBeenCalledWith(7));
        expect(onCommit).toHaveBeenCalledTimes(1);
    });

    it('goes back to the confirmed volume once the commit settles without a new volume', async () => {
        const { sliderTrack } = renderVolume();
        const slider = screen.getByRole('slider', { name: 'Volume' });

        fireEvent.mouseDown(sliderTrack, { clientX: 90, buttons: 1 });
        fireEvent.mouseUp(document);

        await waitFor(() => expect(slider).toHaveValue('4'));
    });
});
