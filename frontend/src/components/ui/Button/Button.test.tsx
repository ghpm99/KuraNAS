import { fireEvent, render, screen } from '@testing-library/react';
import Button from './Button';
import { Plus } from 'lucide-react';

describe('components/ui/Button', () => {
	it('renders primary variant and calls click', () => {
		const onClick = jest.fn();
		render(<Button onClick={onClick}>Run</Button>);
		fireEvent.click(screen.getByRole('button', { name: 'Run' }));
		expect(onClick).toHaveBeenCalled();
	});

	it('renders secondary variant with icon and disabled state', () => {
		render(
			<Button variant='secondary' icon={Plus} type='submit' disabled>
				Create
			</Button>,
		);
		const button = screen.getByRole('button', { name: 'Create' });
		expect(button).toBeDisabled();
		expect(button.querySelector('svg')).toBeTruthy();
	});
});
