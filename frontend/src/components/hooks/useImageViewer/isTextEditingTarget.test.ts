import { isTextEditingTarget, isTypingInTextField } from './isTextEditingTarget';

describe('isTextEditingTarget', () => {
    it('is false for null and non-element targets', () => {
        expect(isTextEditingTarget(null)).toBe(false);
        expect(isTextEditingTarget(window)).toBe(false);
    });

    it('is false for plain elements and explicitly non-editable containers', () => {
        const plainButton = document.createElement('button');
        const lockedContainer = document.createElement('div');
        lockedContainer.setAttribute('contenteditable', 'false');

        expect(isTextEditingTarget(plainButton)).toBe(false);
        expect(isTextEditingTarget(lockedContainer)).toBe(false);
    });

    it('is true for form fields', () => {
        expect(isTextEditingTarget(document.createElement('input'))).toBe(true);
        expect(isTextEditingTarget(document.createElement('textarea'))).toBe(true);
        expect(isTextEditingTarget(document.createElement('select'))).toBe(true);
    });

    it('detects the focused field when the event target is not the field', () => {
        const field = document.createElement('input');
        document.body.appendChild(field);
        field.focus();

        expect(isTypingInTextField(new KeyboardEvent('keydown', { key: 'i' }))).toBe(true);

        field.remove();
        expect(isTypingInTextField(new KeyboardEvent('keydown', { key: 'i' }))).toBe(false);
    });
});
