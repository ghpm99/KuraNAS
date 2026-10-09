const fileItemSelector = '[data-file-id]';

export const focusFileItem = (fileId: number): void => {
    document.querySelector<HTMLElement>(`[data-file-id="${fileId}"]`)?.focus();
};

export const countRenderedGridColumns = (): number => {
    const renderedItems = Array.from(document.querySelectorAll<HTMLElement>(fileItemSelector));
    const [firstItem] = renderedItems;
    if (!firstItem) return 1;
    const itemsInFirstRow = renderedItems.filter((item) => item.offsetTop === firstItem.offsetTop);
    return Math.max(itemsInFirstRow.length, 1);
};
