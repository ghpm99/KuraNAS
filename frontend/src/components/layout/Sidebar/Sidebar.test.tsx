import { render, screen } from '@testing-library/react';
import React from 'react';
import Sidebar from './Sidebar';
import { MemoryRouter } from 'react-router-dom';

const mockUseUI = jest.fn();

jest.mock('@/components/providers/uiProvider/uiContext', () => ({ useUI: () => mockUseUI() }));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));
jest.mock('@/components/layout/Sidebar/components/folderTree', () => ({
	__esModule: true,
	default: () => <div>FolderTree</div>,
}));

describe('layout/Sidebar', () => {
	it('renders navigation entries', () => {
		mockUseUI.mockReturnValue({ activePage: 'images' });
		render(
			<MemoryRouter initialEntries={['/images']}>
				<Sidebar />
			</MemoryRouter>,
		);
		expect(screen.getByText('ALL_FILES')).toBeInTheDocument();
		expect(screen.getByText('NAV_IMAGES')).toBeInTheDocument();
		expect(screen.queryByText('FolderTree')).not.toBeInTheDocument();
	});

	it('shows folder tree only for files page', () => {
		mockUseUI.mockReturnValue({ activePage: 'files' });
		render(
			<MemoryRouter initialEntries={['/']}>
				<Sidebar />
			</MemoryRouter>,
		);
		expect(screen.getByText('FolderTree')).toBeInTheDocument();
	});
});
