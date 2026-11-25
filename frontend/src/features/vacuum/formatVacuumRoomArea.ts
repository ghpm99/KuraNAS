const defaultDecimalSeparator = '.';

export const formatVacuumRoomArea = (area: number | undefined, decimalSeparator: string): string => {
    const safeArea = typeof area === 'number' && Number.isFinite(area) ? Math.max(0, area) : 0;
    const separator = decimalSeparator.length === 1 ? decimalSeparator : defaultDecimalSeparator;
    return safeArea.toFixed(1).replace('.', separator);
};
