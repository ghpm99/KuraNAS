const repeatModesInCycleOrder = ['none', 'all', 'one'] as const;

export type RepeatMode = (typeof repeatModesInCycleOrder)[number];

export const nextRepeatMode = (currentMode: string): RepeatMode => {
    const currentIndex = repeatModesInCycleOrder.indexOf(currentMode as RepeatMode);
    if (currentIndex === -1) return repeatModesInCycleOrder[0];
    return (
        repeatModesInCycleOrder[(currentIndex + 1) % repeatModesInCycleOrder.length] ??
        repeatModesInCycleOrder[0]
    );
};
