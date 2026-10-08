import { isModalOpen, isNativelyActivatableTarget, isTextEntryTarget } from './keyboardEventContext';

describe('keyboardEventContext', () => {
    afterEach(() => {
        document.body.innerHTML = '';
    });

    it('detects text entry targets', () => {
        const input = document.createElement('input');
        const editable = document.createElement('div');
        editable.setAttribute('contenteditable', 'true');
        const plain = document.createElement('div');

        expect(isTextEntryTarget(input)).toBe(true);
        expect(isTextEntryTarget(editable)).toBe(true);
        expect(isTextEntryTarget(plain)).toBe(false);
        expect(isTextEntryTarget(null)).toBe(false);
    });

    it('detects natively activatable targets including their descendants', () => {
        const link = document.createElement('a');
        link.href = '/x';
        const label = document.createElement('span');
        link.appendChild(label);

        expect(isNativelyActivatableTarget(label)).toBe(true);
        expect(isNativelyActivatableTarget(document.createElement('div'))).toBe(false);
    });

    it('detects an open modal root', () => {
        expect(isModalOpen()).toBe(false);
        const modal = document.createElement('div');
        modal.className = 'MuiModal-root';
        document.body.appendChild(modal);
        expect(isModalOpen()).toBe(true);
    });
});
