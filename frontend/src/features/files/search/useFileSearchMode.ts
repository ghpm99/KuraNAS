import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';

export type FileSearchMode = 'name' | 'content';

const searchModeParam = 'in';
const contentModeValue = 'content';

const useFileSearchMode = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const searchMode: FileSearchMode =
        searchParams.get(searchModeParam) === contentModeValue ? 'content' : 'name';

    const setSearchMode = useCallback(
        (nextMode: FileSearchMode) => {
            setSearchParams(
                (currentParams) => {
                    const nextParams = new URLSearchParams(currentParams);
                    if (nextMode === 'content') {
                        nextParams.set(searchModeParam, contentModeValue);
                    } else {
                        nextParams.delete(searchModeParam);
                    }
                    return nextParams;
                },
                { replace: true }
            );
        },
        [setSearchParams]
    );

    return { searchMode, setSearchMode };
};

export default useFileSearchMode;
