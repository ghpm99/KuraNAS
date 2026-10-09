const textEditingTagNames = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

export const isTextEditingTarget = (target: EventTarget | null): boolean => {
    if (!(target instanceof HTMLElement)) {
        return false;
    }
    if (textEditingTagNames.has(target.tagName) || target.isContentEditable) {
        return true;
    }
    return target.closest('[contenteditable]:not([contenteditable="false"])') !== null;
};

export const isTypingInTextField = (event: KeyboardEvent): boolean =>
    isTextEditingTarget(event.target) || isTextEditingTarget(document.activeElement);
