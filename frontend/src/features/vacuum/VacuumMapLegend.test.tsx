import { render, screen } from '@testing-library/react';
import VacuumMapLegend from './VacuumMapLegend';

describe('VacuumMapLegend', () => {
    it('renders without any provider or service', () => {
        expect(() => render(<VacuumMapLegend />)).not.toThrow();
    });

    it('lists wall, obstacle, forbidden zone, dock, path and robot', () => {
        render(<VacuumMapLegend />);

        const legend = screen.getByRole('list', { name: 'VACUUM_MAP_LEGEND_TITLE' });
        expect(Array.from(legend.querySelectorAll('li')).map((entry) => entry.textContent)).toEqual([
            'VACUUM_MAP_LEGEND_WALL',
            'VACUUM_MAP_LEGEND_OBSTACLE',
            'VACUUM_MAP_LEGEND_FORBIDDEN_ZONE',
            'VACUUM_MAP_LEGEND_DOCK',
            'VACUUM_MAP_LEGEND_PATH',
            'VACUUM_MAP_LEGEND_ROBOT',
        ]);
    });
});
