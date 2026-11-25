import { useQuery } from '@tanstack/react-query';
import { getVacuumPath } from '@/service/vacuum';
import type { MillimeterPoint } from './vacuumMapGeometry';
import { vacuumStatusPollingIntervalMs } from './useVacuumStatus';
import { useVacuumState } from './useVacuumState';
import { vacuumQueryKeys } from './vacuumQueryKeys';

const noPathPoints: [number, number][] = [];

const movingStates: ReadonlySet<string> = new Set(['cleaning', 'returning']);

const isMillimeterPoint = (position: unknown): position is MillimeterPoint =>
    typeof position === 'object' &&
    position !== null &&
    typeof (position as MillimeterPoint).x === 'number' &&
    typeof (position as MillimeterPoint).y === 'number';

export const useVacuumPath = (isEnabled: boolean) => {
    const isRobotMoving = movingStates.has(useVacuumState());
    const pathQuery = useQuery({
        queryKey: vacuumQueryKeys.path,
        queryFn: getVacuumPath,
        enabled: isEnabled,
        retry: false,
        refetchInterval: isRobotMoving ? vacuumStatusPollingIntervalMs : false,
        refetchIntervalInBackground: false,
    });
    const reportedPosition = pathQuery.data?.position;

    return {
        pathPoints: Array.isArray(pathQuery.data?.points) ? pathQuery.data.points : noPathPoints,
        robotPosition: isMillimeterPoint(reportedPosition) ? reportedPosition : null,
    };
};
