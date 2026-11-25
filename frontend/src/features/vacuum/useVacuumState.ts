import { toVacuumState } from './vacuumStateCatalog';
import { useVacuumStatus } from './useVacuumStatus';

export const useVacuumState = () => toVacuumState(useVacuumStatus().data?.state);
