import { useState } from 'react';
import type { SyntheticEvent } from 'react';

interface UseSeekSliderParams {
    playbackPosition: number;
    seek: (positionInSeconds: number) => void;
}

interface SeekSliderBinding {
    sliderPosition: number;
    handleSliderChange: (_event: Event, newPosition: number | number[]) => void;
    handleSliderCommit: (
        _event: Event | SyntheticEvent,
        committedPosition: number | number[]
    ) => void;
}

const toSinglePosition = (position: number | number[]): number =>
    Array.isArray(position) ? (position[0] ?? 0) : position;

export const useSeekSlider = ({
    playbackPosition,
    seek,
}: UseSeekSliderParams): SeekSliderBinding => {
    const [draggedPosition, setDraggedPosition] = useState<number | null>(null);

    const handleSliderChange = (_event: Event, newPosition: number | number[]) => {
        setDraggedPosition(toSinglePosition(newPosition));
    };

    const handleSliderCommit = (
        _event: Event | SyntheticEvent,
        committedPosition: number | number[]
    ) => {
        setDraggedPosition(null);
        seek(toSinglePosition(committedPosition));
    };

    return {
        sliderPosition: draggedPosition ?? playbackPosition,
        handleSliderChange,
        handleSliderCommit,
    };
};
