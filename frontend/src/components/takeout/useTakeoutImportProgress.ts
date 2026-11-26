import { getTakeoutImport } from '@/service/takeout';
import type { TakeoutImport } from '@/types/takeout';
import { useQuery } from '@tanstack/react-query';

export const TAKEOUT_IMPORT_POLL_INTERVAL_MS = 2000;

const ACTIVE_IMPORT_STATES = ['queued', 'running'];

export const isTakeoutImportActive = (takeoutImport: TakeoutImport | undefined): boolean =>
	!takeoutImport || ACTIVE_IMPORT_STATES.includes(takeoutImport.state);

const useTakeoutImportProgress = (importId: number) =>
	useQuery({
		queryKey: ['takeout-import', importId],
		queryFn: () => getTakeoutImport(importId),
		refetchInterval: (query) =>
			query.state.error || !isTakeoutImportActive(query.state.data) ? false : TAKEOUT_IMPORT_POLL_INTERVAL_MS,
	});

export default useTakeoutImportProgress;
