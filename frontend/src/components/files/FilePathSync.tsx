import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import useFile from '@/components/providers/fileProvider/fileContext';
import { getFileByPath } from '@/service/files';

export default function FilePathSync() {
	const { selectResolvedItem, selectedItem } = useFile();
	const [searchParams, setSearchParams] = useSearchParams();
	const requestedPath = searchParams.get('path')?.trim() ?? '';

	const { data: requestedItem, isFetched } = useQuery({
		queryKey: ['files-path', requestedPath],
		queryFn: () => getFileByPath(requestedPath),
		enabled: requestedPath.length > 0,
		staleTime: 0,
	});

	useEffect(() => {
		if (!requestedPath || !isFetched) {
			return;
		}

		if (requestedItem?.id && requestedItem.id !== selectedItem?.id) {
			selectResolvedItem(requestedItem);
		}

		const nextSearchParams = new URLSearchParams(searchParams);
		nextSearchParams.delete('path');
		setSearchParams(nextSearchParams, { replace: true });
	}, [isFetched, requestedItem, requestedPath, searchParams, selectResolvedItem, selectedItem?.id, setSearchParams]);

	return null;
}
