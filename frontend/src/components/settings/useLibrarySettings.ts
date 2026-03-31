import useI18n from '@/components/i18n/provider/i18nContext';
import { getLibraries, updateLibrary } from '@/service/libraries';
import type { LibraryCategory, LibraryDto } from '@/types/libraries';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useMemo, useState } from 'react';

const orderedCategories: LibraryCategory[] = ['images', 'music', 'videos', 'documents'];

const useLibrarySettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();
	const librariesQuery = useQuery({
		queryKey: ['libraries'],
		queryFn: getLibraries,
		retry: false,
	});
	const updateMutation = useMutation({
		mutationFn: ({ category, path }: { category: LibraryCategory; path: string }) =>
			updateLibrary(category, { path }),
		onSuccess: (updatedLibrary) => {
			queryClient.setQueryData<LibraryDto[]>(['libraries'], (current = []) => {
				let found = false;
				const next = current.map((library) => {
					if (library.category !== updatedLibrary.category) {
						return library;
					}
					found = true;
					return updatedLibrary;
				});
				if (!found) {
					next.push(updatedLibrary);
				}
				return next;
			});
		},
	});
	const [editedPaths, setEditedPaths] = useState<Partial<Record<LibraryCategory, string>>>({});

	const libraries = useMemo(() => {
		const resolved = librariesQuery.data ?? [];
		const map = new Map<LibraryCategory, LibraryDto>();
		for (const library of resolved) {
			map.set(library.category, library);
		}
		return orderedCategories.map((category) => ({
			category,
			path: editedPaths[category] ?? map.get(category)?.path ?? '',
			originalPath: map.get(category)?.path ?? '',
		}));
	}, [editedPaths, librariesQuery.data]);

	const getCategoryLabel = useCallback(
		(category: LibraryCategory) => {
			switch (category) {
				case 'images':
					return t('LIBRARY_IMAGES');
				case 'music':
					return t('LIBRARY_MUSIC');
				case 'videos':
					return t('LIBRARY_VIDEOS');
				case 'documents':
					return t('LIBRARY_DOCUMENTS');
				default:
					return category;
			}
		},
		[t]
	);

	const setPath = useCallback((category: LibraryCategory, path: string) => {
		setEditedPaths((current) => ({
			...current,
			[category]: path,
		}));
	}, []);

	const handleSave = useCallback(
		async (category: LibraryCategory) => {
			const path =
				(editedPaths[category] ??
					librariesQuery.data?.find((library) => library.category === category)?.path ??
					'').trim();
			try {
				await updateMutation.mutateAsync({ category, path });
				setEditedPaths((current) => ({
					...current,
					[category]: path,
				}));
				enqueueSnackbar(t('SETTINGS_LIBRARY_SAVED'), { variant: 'success' });
			} catch {
				enqueueSnackbar(t('SETTINGS_LIBRARY_SAVE_ERROR'), { variant: 'error' });
			}
		},
		[editedPaths, enqueueSnackbar, librariesQuery.data, t, updateMutation]
	);

	return {
		t,
		libraries,
		isLoading: librariesQuery.isLoading,
		isSaving: updateMutation.isPending,
		hasError: librariesQuery.isError,
		setPath,
		handleSave,
		getCategoryLabel,
	};
};

export default useLibrarySettings;
