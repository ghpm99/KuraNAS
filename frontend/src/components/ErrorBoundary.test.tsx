import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import ErrorBoundary from './ErrorBoundary';

const ThrowingChild = ({ shouldThrow }: { shouldThrow: boolean }) => {
	if (shouldThrow) {
		throw new Error('boom');
	}

	return <div>Recovered child</div>;
};

const RecoverableHarness = () => {
	const [shouldThrow, setShouldThrow] = useState(true);

	return (
		<>
			<button type='button' onClick={() => setShouldThrow(false)}>
				Recover
			</button>
			<ErrorBoundary>
				<ThrowingChild shouldThrow={shouldThrow} />
			</ErrorBoundary>
		</>
	);
};

describe('components/ErrorBoundary', () => {
	let consoleErrorSpy: jest.SpyInstance;

	beforeEach(() => {
		consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => undefined);
	});

	afterEach(() => {
		consoleErrorSpy.mockRestore();
	});

	it('renders children when no error is thrown', () => {
		render(
			<ErrorBoundary>
				<div>Safe child</div>
			</ErrorBoundary>,
		);

		expect(screen.getByText('Safe child')).toBeInTheDocument();
	});

	it('shows the fallback UI and can recover after reset', () => {
		render(<RecoverableHarness />);

		expect(screen.getByText('Something went wrong')).toBeInTheDocument();
		expect(screen.getByText('boom')).toBeInTheDocument();
		expect(consoleErrorSpy).toHaveBeenCalledWith(
			'ErrorBoundary caught:',
			expect.objectContaining({ message: 'boom' }),
			expect.any(Object),
		);

		fireEvent.click(screen.getByRole('button', { name: 'Recover' }));
		fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

		expect(screen.getByText('Recovered child')).toBeInTheDocument();
	});
});
