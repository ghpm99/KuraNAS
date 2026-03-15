import { useQuery } from '@tanstack/react-query';
import { useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import useFile from '@/components/providers/fileProvider/fileContext';
import { getFileByPath } from '@/service/files';

export default function FilePathSync() {
	const { selectResolvedItem, selectedItem } = useFile();
	const [searchParams, setSearchParams] = useSearchParams();
	const requestedPath = searchParams.get('path')?.trim() ?? '';
	const isSyncingFromUrl = useRef(false);

	const { data: requestedItem, isFetched } = useQuery({
		queryKey: ['files-path', requestedPath],
		queryFn: () => getFileByPath(requestedPath),
		enabled: requestedPath.length > 0,
		staleTime: 0,
	});

	// URL → State: resolve ?path= and sync into file context
	useEffect(() => {
		if (!requestedPath || !isFetched) {
			return;
		}

		if (requestedItem?.id && requestedItem.id !== selectedItem?.id) {
			isSyncingFromUrl.current = true;
			selectResolvedItem(requestedItem);
		}
	}, [isFetched, requestedItem, requestedPath, selectResolvedItem, selectedItem?.id]);

	// State → URL: reflect current selection in ?path=
	useEffect(() => {
		if (isSyncingFromUrl.current) {
			isSyncingFromUrl.current = false;
			return;
		}

		const currentUrlPath = searchParams.get('path')?.trim() ?? '';
		const selectedPath = selectedItem?.path ?? '';

		// Don't clear URL path while we're still resolving it
		if (!selectedPath && currentUrlPath && !isFetched) return;

		if (selectedPath && selectedPath !== currentUrlPath) {
			const next = new URLSearchParams(searchParams);
			next.set('path', selectedPath);
			setSearchParams(next, { replace: true });
		} else if (!selectedPath && currentUrlPath) {
			const next = new URLSearchParams(searchParams);
			next.delete('path');
			setSearchParams(next, { replace: true });
		}
	}, [selectedItem?.path, searchParams, setSearchParams, isFetched]);

	return null;
}
