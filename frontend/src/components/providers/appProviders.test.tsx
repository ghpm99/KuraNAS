import { render, screen } from '@testing-library/react';
import AppProviders from './appProviders';

jest.mock('react-router-dom', () => ({ BrowserRouter: ({ children }: any) => <div data-testid='router'>{children}</div> }));
jest.mock('../i18n/provider', () => ({ children }: any) => <div data-testid='i18n'>{children}</div>);
jest.mock('./settingsProvider', () => ({ children }: any) => <div data-testid='settings-provider'>{children}</div>);

describe('appProviders', () => {
	it('wraps children with app providers', () => {
		render(
			<AppProviders>
				<div>body</div>
			</AppProviders>,
		);

		expect(screen.getByTestId('router')).toBeInTheDocument();
		expect(screen.getByTestId('i18n')).toBeInTheDocument();
		expect(screen.getByTestId('settings-provider')).toBeInTheDocument();
		expect(screen.getByText('body')).toBeInTheDocument();
	});
});
