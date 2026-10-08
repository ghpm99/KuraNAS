import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { ImageLibrarySort } from '@/types/imageLibrary';
import {
    buildJumpBeforeDate,
    imageSearchParamNames,
    type ImageLibraryView,
} from './imageLibraryView';

type ParamsMutation = (nextParams: URLSearchParams) => void;

const writeOrDelete = (nextParams: URLSearchParams, name: string, value: string | null) => {
    if (value) {
        nextParams.set(name, value);
        return;
    }
    nextParams.delete(name);
};

export const useImageLibraryControls = (view: ImageLibraryView) => {
    const [, setSearchParams] = useSearchParams();

    const updateParams = useCallback(
        (mutation: ParamsMutation, shouldReplaceHistory = false) => {
            setSearchParams(
                (currentParams) => {
                    const nextParams = new URLSearchParams(currentParams);
                    mutation(nextParams);
                    return nextParams;
                },
                { replace: shouldReplaceHistory }
            );
        },
        [setSearchParams]
    );

    const updateFilterParams = useCallback(
        (mutation: ParamsMutation, shouldReplaceHistory = false) => {
            updateParams((nextParams) => {
                mutation(nextParams);
                nextParams.delete(imageSearchParamNames.jumpBefore);
                nextParams.delete('image');
                nextParams.delete('imagePath');
            }, shouldReplaceHistory);
        },
        [updateParams]
    );

    const setNameQuery = useCallback(
        (nameQuery: string) =>
            updateFilterParams(
                (nextParams) =>
                    writeOrDelete(nextParams, imageSearchParamNames.query, nameQuery.trim()),
                true
            ),
        [updateFilterParams]
    );

    const setTakenFrom = useCallback(
        (takenFrom: string) =>
            updateFilterParams((nextParams) =>
                writeOrDelete(nextParams, imageSearchParamNames.takenFrom, takenFrom)
            ),
        [updateFilterParams]
    );

    const setTakenTo = useCallback(
        (takenTo: string) =>
            updateFilterParams((nextParams) =>
                writeOrDelete(nextParams, imageSearchParamNames.takenTo, takenTo)
            ),
        [updateFilterParams]
    );

    const toggleFormat = useCallback(
        (format: string) =>
            updateFilterParams((nextParams) => {
                const selectedFormats = nextParams.getAll(imageSearchParamNames.format);
                nextParams.delete(imageSearchParamNames.format);
                const nextFormats = selectedFormats.includes(format)
                    ? selectedFormats.filter((selectedFormat) => selectedFormat !== format)
                    : [...selectedFormats, format];
                nextFormats.forEach((nextFormat) =>
                    nextParams.append(imageSearchParamNames.format, nextFormat)
                );
            }),
        [updateFilterParams]
    );

    const setSort = useCallback(
        (sort: ImageLibrarySort) =>
            updateFilterParams((nextParams) => {
                writeOrDelete(
                    nextParams,
                    imageSearchParamNames.sort,
                    sort === 'taken_at' ? null : sort
                );
                nextParams.delete(imageSearchParamNames.order);
            }),
        [updateFilterParams]
    );

    const toggleSortOrder = useCallback(
        () =>
            updateFilterParams((nextParams) =>
                nextParams.set(
                    imageSearchParamNames.order,
                    view.ordering.order === 'asc' ? 'desc' : 'asc'
                )
            ),
        [updateFilterParams, view.ordering.order]
    );

    const clearUserFilters = useCallback(
        () =>
            updateFilterParams((nextParams) => {
                nextParams.delete(imageSearchParamNames.query);
                nextParams.delete(imageSearchParamNames.takenFrom);
                nextParams.delete(imageSearchParamNames.takenTo);
                nextParams.delete(imageSearchParamNames.format);
            }),
        [updateFilterParams]
    );

    const jumpToMonth = useCallback(
        (year: number, month: number) =>
            updateParams((nextParams) => {
                nextParams.set(imageSearchParamNames.jumpBefore, buildJumpBeforeDate(year, month));
                nextParams.delete('image');
                nextParams.delete('imagePath');
            }),
        [updateParams]
    );

    const clearJump = useCallback(
        () => updateParams((nextParams) => nextParams.delete(imageSearchParamNames.jumpBefore)),
        [updateParams]
    );

    const selectFolder = useCallback(
        (folderPath: string | null) =>
            updateFilterParams((nextParams) =>
                writeOrDelete(nextParams, imageSearchParamNames.folder, folderPath)
            ),
        [updateFilterParams]
    );

    const selectAlbum = useCallback(
        (albumId: string | null) =>
            updateFilterParams((nextParams) =>
                writeOrDelete(nextParams, imageSearchParamNames.album, albumId)
            ),
        [updateFilterParams]
    );

    const openImageParam = useCallback(
        (imageId: number) => updateParams((nextParams) => nextParams.set('image', String(imageId))),
        [updateParams]
    );

    const closeImageParam = useCallback(
        () =>
            updateParams((nextParams) => {
                nextParams.delete('image');
                nextParams.delete('imagePath');
            }),
        [updateParams]
    );

    return {
        setNameQuery,
        setTakenFrom,
        setTakenTo,
        toggleFormat,
        setSort,
        toggleSortOrder,
        clearUserFilters,
        jumpToMonth,
        clearJump,
        selectFolder,
        selectAlbum,
        openImageParam,
        closeImageParam,
    };
};
