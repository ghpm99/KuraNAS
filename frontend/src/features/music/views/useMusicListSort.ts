import { useCallback, useState } from 'react';
import type { MusicListSort, MusicListSortField } from '@/types/music';
import {
    getDefaultOrderOfField,
    loadMusicListSort,
    saveMusicListSort,
    type MusicSortableView,
} from './musicListSortPreference';

export const useMusicListSort = (view: MusicSortableView) => {
    const [listSort, setListSort] = useState<MusicListSort>(() => loadMusicListSort(view));

    const updateListSort = useCallback(
        (nextSort: MusicListSort) => {
            setListSort(nextSort);
            saveMusicListSort(view, nextSort);
        },
        [view]
    );

    const changeField = useCallback(
        (field: MusicListSortField) =>
            updateListSort({ sort: field, order: getDefaultOrderOfField(field) }),
        [updateListSort]
    );

    const toggleOrder = useCallback(
        () =>
            updateListSort({
                sort: listSort.sort,
                order: listSort.order === 'asc' ? 'desc' : 'asc',
            }),
        [listSort, updateListSort]
    );

    return { listSort, changeField, toggleOrder };
};
