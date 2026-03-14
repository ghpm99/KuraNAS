import { render, screen } from '@testing-library/react';
import Layout from './index';

jest.mock('../AppShell', () => ({
	__esModule: true,
	default: ({ children }: any) => <div data-testid='app-shell'>{children}</div>,
}));

describe('layout/Layout/index', () => {
	it('delegates children to the app shell', () => {
		render(
			<Layout>
				<div>content</div>
			</Layout>,
		);

		expect(screen.getByTestId('app-shell')).toBeInTheDocument();
		expect(screen.getByText('content')).toBeInTheDocument();
	});
});
