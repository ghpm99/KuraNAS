import useI18n from '@/components/i18n/provider/i18nContext';
import VacuumBlock from './VacuumBlock';
import VacuumMapCloudNotConfigured from './VacuumMapCloudNotConfigured';
import styles from './VacuumMapBlock.module.css';
import VacuumMapWorkspace from './VacuumMapWorkspace';
import { useVacuumCloudReadiness } from './useVacuumCloudReadiness';
import { useVacuumMap } from './useVacuumMap';
import { useVacuumPath } from './useVacuumPath';
import { useVacuumRooms } from './useVacuumRooms';
import { useVacuumState } from './useVacuumState';
import { isVacuumCommandAvailable } from './vacuumCommandAvailability';

const VacuumMapBlock = () => {
    const { t } = useI18n();
    const cloud = useVacuumCloudReadiness();
    const { mapQuery, map } = useVacuumMap(cloud.isConfigured);
    const { roomsQuery, rooms } = useVacuumRooms(cloud.isConfigured);
    const { pathPoints, robotPosition } = useVacuumPath(cloud.isConfigured);
    const state = useVacuumState();

    const hasMapLoadError = mapQuery.isError && mapQuery.data === undefined;
    const hasRoomsLoadError = roomsQuery.isError && roomsQuery.data === undefined;
    const isLoadingContent =
        cloud.isChecking ||
        (cloud.isConfigured && (mapQuery.isPending || roomsQuery.isPending));
    const hasLoadError = cloud.hasLoadError || hasMapLoadError || hasRoomsLoadError;

    const retryFailedLoads = () => {
        if (cloud.hasLoadError) cloud.refetch();
        if (hasMapLoadError) void mapQuery.refetch();
        if (hasRoomsLoadError) void roomsQuery.refetch();
    };

    const renderContent = () => {
        if (!cloud.isConfigured) return <VacuumMapCloudNotConfigured />;
        if (map === null) {
            return <p className={styles.empty}>{t('VACUUM_MAP_EMPTY')}</p>;
        }
        return (
            <VacuumMapWorkspace
                map={map}
                rooms={rooms}
                pathPoints={pathPoints}
                robotPosition={robotPosition}
                canStartCleaning={isVacuumCommandAvailable('start', state)}
            />
        );
    };

    return (
        <VacuumBlock
            className={styles.block}
            title={t('VACUUM_MAP_TITLE')}
            isLoading={isLoadingContent}
            hasLoadError={hasLoadError}
            loadError={cloud.loadError ?? mapQuery.error ?? roomsQuery.error}
            onRetry={retryFailedLoads}
        >
            {renderContent()}
        </VacuumBlock>
    );
};

export default VacuumMapBlock;
