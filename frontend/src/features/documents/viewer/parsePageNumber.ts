export const parsePageNumber = (rawPageNumber: string | null): number | undefined => {
    const pageNumber = Number(rawPageNumber);
    return Number.isInteger(pageNumber) && pageNumber > 0 ? pageNumber : undefined;
};
