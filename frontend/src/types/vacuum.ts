export type VacuumState =
    | 'idle'
    | 'cleaning'
    | 'paused'
    | 'returning'
    | 'charging'
    | 'charged'
    | 'emptying'
    | 'sleeping'
    | 'remote_control'
    | 'fault'
    | 'unknown';

export type VacuumCommand = 'start' | 'pause' | 'resume' | 'stop' | 'dock' | 'locate';

export type VacuumDirection = 'forward' | 'backward' | 'left' | 'right' | 'stop';

export type VacuumConsumablePart = 'filter' | 'main_brush' | 'side_brush';

export type VacuumSettings = {
    address: string;
    device_id: string;
    has_local_key: boolean;
};

export type VacuumSettingsChanges = {
    address: string;
    device_id: string;
    local_key: string;
};

export type VacuumDiscovery = {
    address: string;
};

export type VacuumStatus = {
    state: string;
    battery_percent: number;
    cleaning_seconds: number;
    cleaning_area: number;
    fault_codes: number[];
    read_at: string;
};

export type VacuumPreferences = {
    suction: string | null;
    water_level: string | null;
    volume: number | null;
    is_do_not_disturb: boolean | null;
    is_carpet_boost: boolean | null;
    is_resume_after_charge: boolean | null;
    is_auto_empty: boolean | null;
    auto_empty_frequency: string | null;
};

export type VacuumPreferencesPatch = Partial<VacuumPreferences>;

export type VacuumConsumable = {
    part: string;
    seconds: number;
    max_seconds: number;
};

export type VacuumConsumables = {
    items: VacuumConsumable[];
};

export type VacuumTotals = {
    cleaning_seconds: number;
    cleaning_area: number;
    cleaning_count: number;
};

export type VacuumCloudRegion = 'us' | 'eu' | 'cn' | 'in';

export type VacuumCloudSettings = {
    access_id: string;
    region: string;
    has_access_secret: boolean;
};

export type VacuumCloudSettingsChanges = {
    access_id: string;
    access_secret: string;
    region: string;
};

export type VacuumMapPoint = {
    x: number;
    y: number;
};

export type VacuumForbiddenZone = {
    vertices: [number, number][];
};

export type VacuumMap = {
    map_id: number;
    width: number;
    height: number;
    resolution: number;
    origin_x: number;
    origin_y: number;
    cells: string;
    dock: VacuumMapPoint | null;
    forbidden_zones: VacuumForbiddenZone[];
    read_at: string;
};

export type VacuumRoom = {
    id: number;
    name: string | null;
    area: number;
};

export type VacuumRooms = {
    map_id: number;
    items: VacuumRoom[];
};

export type VacuumRoomRename = {
    map_id: number;
    name: string;
};

export type VacuumPath = {
    path_id: number | null;
    points: [number, number][];
    position: VacuumMapPoint | null;
};

export type VacuumRoomCleaningRequest = {
    map_id: number;
    room_ids: number[];
};

export type VacuumSpotCleaningRequest = {
    map_id: number;
    x: number;
    y: number;
};
