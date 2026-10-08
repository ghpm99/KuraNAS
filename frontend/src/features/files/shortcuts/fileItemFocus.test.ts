import { countRenderedGridColumns, focusFileItem } from './fileItemFocus';

const appendItem = (fileId: number, top: number) => {
    const item = document.createElement('a');
    item.setAttribute('data-file-id', String(fileId));
    item.href = `/f/${fileId}`;
    Object.defineProperty(item, 'offsetTop', { value: top });
    document.body.appendChild(item);
    return item;
};

describe('fileItemFocus', () => {
    afterEach(() => {
        document.body.innerHTML = '';
    });

    it('focuses the element tagged with the file id', () => {
        appendItem(1, 0);
        const second = appendItem(2, 0);

        focusFileItem(2);

        expect(document.activeElement).toBe(second);
    });

    it('does nothing when the file is not rendered', () => {
        expect(() => focusFileItem(99)).not.toThrow();
    });

    it('counts the items sharing the first row offset', () => {
        appendItem(1, 0);
        appendItem(2, 0);
        appendItem(3, 0);
        appendItem(4, 200);

        expect(countRenderedGridColumns()).toBe(3);
    });

    it('falls back to one column when nothing is rendered', () => {
        expect(countRenderedGridColumns()).toBe(1);
    });
});
