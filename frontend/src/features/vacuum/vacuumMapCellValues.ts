export const wallCellValue = 0;
export const firstRoomCellValue = 1;
export const lastRoomCellValue = 126;
export const unknownCellValue = 127;
export const obstacleCellValue = 255;

export const isRoomFloorCellValue = (cellValue: number): boolean =>
    cellValue >= firstRoomCellValue && cellValue <= lastRoomCellValue;
