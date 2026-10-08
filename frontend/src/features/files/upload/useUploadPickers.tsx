import { useRef, type ChangeEvent, type InputHTMLAttributes, type ReactNode } from 'react';
import { entriesFromFileList } from './entriesFromFileList';
import useUploadToCurrentFolder from './useUploadToCurrentFolder';

const folderInputAttributes = {
    webkitdirectory: '',
} as unknown as InputHTMLAttributes<HTMLInputElement>;

const hiddenInputStyle = { display: 'none' };

export const useUploadPickers = (): {
    openFilePicker: () => void;
    openFolderPicker: () => void;
    pickerInputs: ReactNode;
} => {
    const filePickerRef = useRef<HTMLInputElement | null>(null);
    const folderPickerRef = useRef<HTMLInputElement | null>(null);
    const { uploadEntries } = useUploadToCurrentFolder();

    const uploadPickedFiles = async (event: ChangeEvent<HTMLInputElement>) => {
        const pickedFiles = event.target.files;
        if (!pickedFiles || pickedFiles.length === 0) return;
        const entries = entriesFromFileList(pickedFiles);
        event.target.value = '';
        await uploadEntries(entries);
    };

    return {
        openFilePicker: () => filePickerRef.current?.click(),
        openFolderPicker: () => folderPickerRef.current?.click(),
        pickerInputs: (
            <>
                <input
                    ref={filePickerRef}
                    type="file"
                    multiple
                    style={hiddenInputStyle}
                    onChange={uploadPickedFiles}
                />
                <input
                    ref={folderPickerRef}
                    type="file"
                    multiple
                    style={hiddenInputStyle}
                    onChange={uploadPickedFiles}
                    {...folderInputAttributes}
                />
            </>
        ),
    };
};

export default useUploadPickers;
