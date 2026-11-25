/**
 * Decodes the base64 grid sent by the server into one byte per cell.
 * Returns an empty array when the text is absent or is not valid base64.
 */
export const decodeVacuumMapCells = (encodedCells: string | undefined): Uint8Array => {
    if (!encodedCells) return new Uint8Array(0);
    try {
        const binary = atob(encodedCells);
        const cells = new Uint8Array(binary.length);
        for (let index = 0; index < binary.length; index += 1) {
            cells[index] = binary.charCodeAt(index);
        }
        return cells;
    } catch {
        return new Uint8Array(0);
    }
};
