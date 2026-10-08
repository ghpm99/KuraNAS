const textEntryTagNames = ['INPUT', 'TEXTAREA', 'SELECT'];
const nativelyActivatableSelector = 'a[href], button, summary, [role="button"], [role="menuitem"]';
const openModalSelector = '.MuiModal-root';

export const isTextEntryTarget = (target: EventTarget | null): boolean => {
    if (!(target instanceof HTMLElement)) return false;
    if (textEntryTagNames.includes(target.tagName)) return true;
    const contentEditableValue = target.getAttribute('contenteditable');
    return target.isContentEditable === true || contentEditableValue === '' || contentEditableValue === 'true';
};

export const isNativelyActivatableTarget = (target: EventTarget | null): boolean =>
    target instanceof HTMLElement && target.closest(nativelyActivatableSelector) !== null;

export const isModalOpen = (): boolean => document.querySelector(openModalSelector) !== null;
