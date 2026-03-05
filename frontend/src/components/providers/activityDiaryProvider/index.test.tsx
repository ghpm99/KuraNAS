import { fireEvent, render, screen } from '@testing-library/react';
import ActivityDiaryProvider from '.';
import { useActivityDiary } from './ActivityDiaryContext';

const mockUseQuery = jest.fn();
const mockUseMutation = jest.fn();
const mockEnqueueSnackbar = jest.fn();
const mockApiGet = jest.fn();
const mockApiPost = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useQuery: (...args: any[]) => mockUseQuery(...args),
	useMutation: (...args: any[]) => mockUseMutation(...args),
}));

jest.mock('@/service', () => ({
	apiBase: {
		get: (...args: any[]) => mockApiGet(...args),
		post: (...args: any[]) => mockApiPost(...args),
	},
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

const Harness = () => {
	const context = useActivityDiary();
	return (
		<div>
			<form onSubmit={context.handleSubmit} aria-label='diary-form'>
				<input aria-label='name' value={context.form.name} onChange={(e) => context.handleNameChange(e as any)} />
				<textarea
					aria-label='description'
					value={context.form.description}
					onChange={(e) => context.handleDescriptionChange(e as any)}
				/>
				<button type='submit'>submit</button>
			</form>
			<button
				type='button'
				onClick={() =>
						context.copyActivity({
							id: 9,
							name: 'copiable',
							description: '',
							start_time: '2026-01-01T00:00:00.000Z',
							end_time: { Value: '', HasValue: false },
							duration: 0,
							duration_formatted: null,
						})
					}
				>
				copy
			</button>
			<div data-testid='message'>{context.message?.text ?? ''}</div>
			<div data-testid='error'>{context.error ?? ''}</div>
			<div data-testid='entries'>{context.data?.entries.items.length ?? 0}</div>
			<div data-testid='duration'>{context.getCurrentDuration('2026-01-01T00:00:00.000Z')}</div>
		</div>
	);
};

describe('providers/activityDiaryProvider/index', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		jest.spyOn(console, 'error').mockImplementation(() => {});

		mockApiGet.mockImplementation((path: string) => {
			if (path === '/diary/summary') {
				return Promise.resolve({
					data: {
						date: '2026-03-04',
						total_activities: 1,
						total_time_spent_seconds: 120,
					},
				});
			}
			return Promise.resolve({
				data: {
					items: [
						{
							id: 1,
							name: 'entry',
							description: 'x',
							start_time: '2026-03-04T00:00:00.000Z',
							end_time: null,
							duration: 1,
							duration_formatted: '1s',
						},
					],
					pagination: { page: 1, page_size: 10, has_next: false, has_prev: false },
				},
			});
		});
		mockApiPost.mockResolvedValue({ data: { id: 99 } });

		mockUseQuery.mockImplementation((options: any) => {
			options.queryFn?.();
			if (options.queryKey[0] === 'activity-diary-summary') {
				return {
					data: {
						date: '2026-03-04',
						total_activities: 1,
						total_time_spent_seconds: 120,
					},
					error: undefined,
					refetch: jest.fn(),
				};
			}

			return {
				data: {
					items: [
						{
							id: 1,
							name: 'entry',
							description: 'x',
							start_time: '2026-03-04T00:00:00.000Z',
							end_time: null,
							duration: 1,
							duration_formatted: '1s',
						},
					],
					pagination: { page: 1, page_size: 10, has_next: false, has_prev: false },
				},
				error: undefined,
				refetch: jest.fn(),
			};
		});

		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (value: any) => {
				options.mutationFn?.(value);
				if (value === -1 || (typeof value === 'object' && value?.name === 'force-error')) {
					options.onError?.(new Error('forced error'));
					return;
				}
				options.onSuccess?.();
			},
		}));
	});

	afterEach(() => {
		jest.restoreAllMocks();
	});

	it('validates form and submits normalized payload', () => {
		render(
			<ActivityDiaryProvider>
				<Harness />
			</ActivityDiaryProvider>,
		);

		fireEvent.change(screen.getByLabelText('name'), { target: { value: 'ab' } });
		fireEvent.submit(screen.getByLabelText('diary-form'));
		expect(screen.getByTestId('message').textContent).toContain('mínimo 3');

		fireEvent.change(screen.getByLabelText('name'), { target: { value: '#invalid' } });
		fireEvent.submit(screen.getByLabelText('diary-form'));
		expect(screen.getByTestId('message').textContent).toContain('só pode conter');

		fireEvent.change(screen.getByLabelText('name'), { target: { value: ' '.repeat(3) } });
		fireEvent.submit(screen.getByLabelText('diary-form'));
		expect(screen.getByTestId('message').textContent).toContain('não pode ser vazio');

		fireEvent.change(screen.getByLabelText('name'), { target: { value: 'A'.repeat(51) } });
		fireEvent.submit(screen.getByLabelText('diary-form'));
		expect(screen.getByTestId('message').textContent).toContain('máximo 50');

		fireEvent.change(screen.getByLabelText('name'), { target: { value: '  Valid Name  ' } });
		fireEvent.change(screen.getByLabelText('description'), { target: { value: 'desc' } });
		fireEvent.submit(screen.getByLabelText('diary-form'));

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Atividade adicionada com sucesso!', { variant: 'success' });
		expect(screen.getByLabelText('name')).toHaveValue('');
		expect(screen.getByLabelText('description')).toHaveValue('');
		expect(Number(screen.getByTestId('duration').textContent)).toBeGreaterThanOrEqual(0);
		expect(screen.getByTestId('entries')).toHaveTextContent('1');
		expect(mockApiPost).toHaveBeenCalled();
	});

	it('copies activity and handles mutation errors', () => {
		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (value: any) => {
				options.mutationFn?.(value);
				if (typeof value === 'number') {
					options.onError?.(new Error('copy error'));
					return;
				}
				options.onError?.(new Error('create error'));
			},
		}));

		render(
			<ActivityDiaryProvider>
				<Harness />
			</ActivityDiaryProvider>,
		);

		fireEvent.change(screen.getByLabelText('name'), { target: { value: 'ValidName' } });
		fireEvent.submit(screen.getByLabelText('diary-form'));
		fireEvent.click(screen.getByRole('button', { name: 'copy' }));

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Erro ao adicionar atividade.', { variant: 'error' });
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Erro ao duplicar atividade.', { variant: 'error' });
	});

	it('exposes query error message when data fetch fails', () => {
		mockUseQuery.mockImplementation(({ queryKey, queryFn }: any) => {
			queryFn?.();
			if (queryKey[0] === 'activity-diary-summary') {
				return { data: undefined, error: new Error('summary down'), refetch: jest.fn() };
			}
			return { data: undefined, error: new Error('list down'), refetch: jest.fn() };
		});

		render(
			<ActivityDiaryProvider>
				<Harness />
			</ActivityDiaryProvider>,
		);

		expect(screen.getByTestId('error')).toHaveTextContent('list down');
	});
});
