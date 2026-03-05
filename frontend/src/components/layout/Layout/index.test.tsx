import { render, screen } from '@testing-library/react';
import React from 'react';
import Layout from './index';

jest.mock('@/components/providers/uiProvider', () => ({
	UIProvider: ({ children }: any) => <div data-testid='ui-provider'>{children}</div>,
}));

jest.mock('@/components/providers/activityDiaryProvider', () => ({
	__esModule: true,
	default: ({ children }: any) => <div data-testid='activity-provider'>{children}</div>,
}));

jest.mock('./Layout', () => ({
	Layout: ({ children }: any) => <div data-testid='layout-component'>{children}</div>,
}));

jest.mock('@/components/activePageListener', () => ({
	__esModule: true,
	default: () => <div data-testid='active-page-listener'>listener</div>,
}));

describe('layout/Layout/index', () => {
	it('wraps children with providers and active page listener', () => {
		render(
			<Layout>
				<div>content</div>
			</Layout>,
		);

		expect(screen.getByTestId('ui-provider')).toBeInTheDocument();
		expect(screen.getByTestId('activity-provider')).toBeInTheDocument();
		expect(screen.getByTestId('layout-component')).toBeInTheDocument();
		expect(screen.getByTestId('active-page-listener')).toBeInTheDocument();
		expect(screen.getByText('content')).toBeInTheDocument();
	});
});
