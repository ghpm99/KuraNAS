import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { useEffect, useMemo } from 'react';
import { findSiblingFiles, locateAmongSiblings, type SiblingPosition } from './siblingFiles';

const keyboardIgnoredSelector = 'input, textarea, select, audio, video, [contenteditable="true"]';

const isKeyboardTargetIgnored = (target: EventTarget | null): boolean =>
    target instanceof Element && target.closest(keyboardIgnoredSelector) !== null;

const hasModifierKey = (event: KeyboardEvent): boolean =>
    event.altKey || event.ctrlKey || event.metaKey || event.shiftKey;

const useSiblingNavigation = (file: FileData): SiblingPosition | null => {
    const { files, handleSelectItem } = useFile();

    const siblingPosition = useMemo(
        () => locateAmongSiblings(findSiblingFiles(files ?? [], file), file.id),
        [files, file]
    );

    useEffect(() => {
        if (!siblingPosition) return undefined;

        const navigateWithArrowKeys = (event: KeyboardEvent) => {
            if (hasModifierKey(event) || isKeyboardTargetIgnored(event.target)) return;
            if (event.key === 'ArrowLeft' && siblingPosition.previous) {
                handleSelectItem(siblingPosition.previous);
            }
            if (event.key === 'ArrowRight' && siblingPosition.next) {
                handleSelectItem(siblingPosition.next);
            }
        };

        window.addEventListener('keydown', navigateWithArrowKeys);
        return () => window.removeEventListener('keydown', navigateWithArrowKeys);
    }, [siblingPosition, handleSelectItem]);

    return siblingPosition;
};

export default useSiblingNavigation;
