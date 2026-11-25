import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumMapLegend.module.css';
import { vacuumMapPalette } from './vacuumMapPalette';
import { toCssColor } from './vacuumRoomColor';

type LegendEntry = {
    labelKey: string;
    swatchColor: string;
    shape: 'square' | 'circle' | 'line';
    outlineColor?: string;
};

const legendEntries: readonly LegendEntry[] = [
    {
        labelKey: 'VACUUM_MAP_LEGEND_WALL',
        swatchColor: toCssColor(vacuumMapPalette.wall),
        shape: 'square',
    },
    {
        labelKey: 'VACUUM_MAP_LEGEND_OBSTACLE',
        swatchColor: toCssColor(vacuumMapPalette.obstacle),
        shape: 'square',
    },
    {
        labelKey: 'VACUUM_MAP_LEGEND_FORBIDDEN_ZONE',
        swatchColor: vacuumMapPalette.forbiddenFill,
        outlineColor: vacuumMapPalette.forbiddenStroke,
        shape: 'square',
    },
    { labelKey: 'VACUUM_MAP_LEGEND_DOCK', swatchColor: vacuumMapPalette.dock, shape: 'circle' },
    { labelKey: 'VACUUM_MAP_LEGEND_PATH', swatchColor: vacuumMapPalette.path, shape: 'line' },
    { labelKey: 'VACUUM_MAP_LEGEND_ROBOT', swatchColor: vacuumMapPalette.robot, shape: 'circle' },
];

const VacuumMapLegend = () => {
    const { t } = useI18n();

    return (
        <ul className={styles.legend} aria-label={t('VACUUM_MAP_LEGEND_TITLE')}>
            {legendEntries.map((entry) => (
                <li key={entry.labelKey} className={styles.entry}>
                    <span
                        aria-hidden="true"
                        className={`${styles.swatch} ${styles[entry.shape]}`}
                        style={{
                            background: entry.swatchColor,
                            borderColor: entry.outlineColor ?? entry.swatchColor,
                        }}
                    />
                    {t(entry.labelKey)}
                </li>
            ))}
        </ul>
    );
};

export default VacuumMapLegend;
