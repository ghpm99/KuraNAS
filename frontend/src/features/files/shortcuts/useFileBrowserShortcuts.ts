import { useCallback, useEffect, useState } from 'react';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import type { FileSelection } from '@/features/files/selection/useFileSelection';
import { focusFileItem } from './fileItemFocus';
import { isModalOpen, isNativelyActivatableTarget, isTextEntryTarget } from '@/components/shortcuts/keyboardEventContext';
import { resolveNextFocusIndex, resolveShortcutAction, type ArrowDirection } from './shortcutAction';

type FileBrowserShortcutsOptions = {
    isEnabled: boolean;
    files: FileData[];
    selection: FileSelection;
    getColumnCount: () => number;
    onOpenFile: (file: FileData) => void;
    onDeleteFiles: (files: FileData[]) => void;
    onRenameFile: (file: FileData) => void;
    onGoToParent?: () => void;
    onFocusSearch?: () => void;
};

export const useFileBrowserShortcuts = ({
    isEnabled,
    files,
    selection,
    getColumnCount,
    onOpenFile,
    onDeleteFiles,
    onRenameFile,
    onGoToParent,
    onFocusSearch,
}: FileBrowserShortcutsOptions) => {
    const [focusedFileId, setFocusedFileId] = useState<number | null>(null);
    const { selectedFiles, selectedCount, isSelected, toggle, selectRange, selectAll } = selection;

    const focusedIndex = files.findIndex((file) => file.id === focusedFileId);
    const focusedFile = focusedIndex === -1 ? null : files[focusedIndex];
    const tabStopFileId = focusedFile?.id ?? files[0]?.id ?? null;

    const moveFocus = useCallback(
        (direction: ArrowDirection, isExtendingSelection: boolean) => {
            const nextIndex = resolveNextFocusIndex({
                currentIndex: focusedIndex,
                direction,
                columnCount: getColumnCount(),
                itemCount: files.length,
            });
            const nextFile = files[nextIndex];
            if (!nextFile) return;
            if (isExtendingSelection) {
                if (focusedFile && !isSelected(focusedFile.id)) toggle(focusedFile);
                selectRange(nextFile, files);
            }
            setFocusedFileId(nextFile.id);
            focusFileItem(nextFile.id);
        },
        [files, focusedFile, focusedIndex, getColumnCount, isSelected, selectRange, toggle]
    );

    const findSingleSelectedFile = useCallback(
        () => (selectedCount === 1 ? selectedFiles[0] : null),
        [selectedCount, selectedFiles]
    );

    useEffect(() => {
        if (!isEnabled) return undefined;

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.defaultPrevented || isTextEntryTarget(event.target) || isModalOpen()) return;
            const action = resolveShortcutAction(event);
            if (!action) return;

            if (action.kind === 'openItem') {
                if (isNativelyActivatableTarget(event.target)) return;
                const fileToOpen = focusedFile ?? findSingleSelectedFile();
                if (!fileToOpen) return;
                event.preventDefault();
                onOpenFile(fileToOpen);
                return;
            }
            if (action.kind === 'deleteSelection') {
                const filesToDelete = selectedCount > 0 ? selectedFiles : focusedFile ? [focusedFile] : [];
                if (filesToDelete.length === 0) return;
                event.preventDefault();
                onDeleteFiles(filesToDelete);
                return;
            }
            if (action.kind === 'renameSelection') {
                const fileToRename = selectedCount > 0 ? findSingleSelectedFile() : focusedFile;
                if (!fileToRename) return;
                event.preventDefault();
                onRenameFile(fileToRename);
                return;
            }
            if (action.kind === 'selectAll') {
                event.preventDefault();
                selectAll(files);
                return;
            }
            if (action.kind === 'goToParent') {
                if (!onGoToParent) return;
                event.preventDefault();
                onGoToParent();
                return;
            }
            if (action.kind === 'focusSearch') {
                if (!onFocusSearch) return;
                event.preventDefault();
                onFocusSearch();
                return;
            }
            event.preventDefault();
            moveFocus(action.direction, action.isExtendingSelection);
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [
        isEnabled,
        files,
        focusedFile,
        selectedFiles,
        selectedCount,
        findSingleSelectedFile,
        selectAll,
        moveFocus,
        onOpenFile,
        onDeleteFiles,
        onRenameFile,
        onGoToParent,
        onFocusSearch,
    ]);

    return {
        tabStopFileId,
        focusFile: setFocusedFileId,
    };
};

export default useFileBrowserShortcuts;
