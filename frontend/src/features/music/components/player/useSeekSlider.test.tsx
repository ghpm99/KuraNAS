import { act, renderHook } from '@testing-library/react';
import { useSeekSlider } from './useSeekSlider';

const dragEvent = new Event('change');

describe('useSeekSlider', () => {
    it('follows the playback position without any service mock', () => {
        const { result } = renderHook(() =>
            useSeekSlider({ playbackPosition: 12, seek: () => undefined })
        );

        expect(result.current.sliderPosition).toBe(12);
    });

    it('shows the dragged position without calling seek', () => {
        const seek = jest.fn();
        const { result } = renderHook(() => useSeekSlider({ playbackPosition: 12, seek }));

        act(() => result.current.handleSliderChange(dragEvent, 80));

        expect(result.current.sliderPosition).toBe(80);
        expect(seek).not.toHaveBeenCalled();
    });

    it('seeks once on commit and follows playback again afterwards', () => {
        const seek = jest.fn();
        const { result, rerender } = renderHook(
            ({ playbackPosition }) => useSeekSlider({ playbackPosition, seek }),
            { initialProps: { playbackPosition: 12 } }
        );

        act(() => result.current.handleSliderChange(dragEvent, 80));
        act(() => result.current.handleSliderChange(dragEvent, 90));
        act(() => result.current.handleSliderCommit(dragEvent, 90));

        expect(seek).toHaveBeenCalledTimes(1);
        expect(seek).toHaveBeenCalledWith(90);

        rerender({ playbackPosition: 91 });
        expect(result.current.sliderPosition).toBe(91);
    });
});
