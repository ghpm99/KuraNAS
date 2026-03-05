import { act, render, screen } from '@testing-library/react';
import UpdateCard from './UpdateCard';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
	<div>
		<h2>{title}</h2>
		{children}
	</div>
));

jest.mock('@/components/providers/aboutProvider/AboutContext', () => ({
	useAbout: () => ({ version: '1.0.0' }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const enqueueSnackbar = jest.fn();
jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar }),
}));

jest.mock('@/service/update', () => ({
	getUpdateStatus: jest.fn(),
	applyUpdate: jest.fn(),
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;

describe('about/UpdateCard', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: jest.fn() });
		mockedUseQuery.mockReturnValue({
			data: undefined,
			isLoading: false,
			isError: false,
			error: undefined,
			refetch: jest.fn(),
		});
		mockedUseMutation.mockImplementation(({ onSuccess, onError }: any) => ({
			isPending: false,
			mutate: (fail?: boolean) => {
				if (fail) onError?.();
				else onSuccess?.();
			},
		}));
	});

	it('renders current version and checks updates', () => {
		render(<UpdateCard />);
		expect(screen.getByText('CURRENT_VERSION')).toBeInTheDocument();
		expect(screen.getByText('1.0.0')).toBeInTheDocument();
		screen.getByText('CHECK_FOR_UPDATES').click();
		expect(mockedUseQuery.mock.results[0]!.value.refetch).toHaveBeenCalled();
	});

	it('renders update details and applies update with success snackbar', async () => {
		mockedUseQuery.mockReturnValue({
			data: {
				latest_version: '1.1.0',
				update_available: true,
				release_date: '2026-01-01',
				asset_size: 2048,
				release_notes: 'notes',
			},
			isLoading: false,
			isError: false,
			error: undefined,
			refetch: jest.fn(),
		});

		render(<UpdateCard />);
		expect(screen.getByText('LATEST_VERSION')).toBeInTheDocument();
		expect(screen.getByText('UPDATE_NOW')).toBeInTheDocument();

		await act(async () => {
			screen.getByText('UPDATE_NOW').click();
		});
		expect(enqueueSnackbar).toHaveBeenCalledWith('UPDATE_APPLIED_SUCCESS', { variant: 'success' });
	});

	it('renders error state and mutation error path', async () => {
		const invalidateQueries = jest.fn();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries });
		mockedUseQuery.mockReturnValue({
			data: undefined,
			isLoading: false,
			isError: true,
			error: new Error('network'),
			refetch: jest.fn(),
		});
		mockedUseMutation.mockImplementation(({ onError }: any) => ({
			isPending: false,
			mutate: () => onError?.(),
		}));

		render(<UpdateCard />);
		expect(screen.getByText(/UPDATE_CHECK_ERROR/)).toBeInTheDocument();

		await act(async () => {
			mockedUseMutation.mock.results[0]!.value.mutate();
		});

		expect(enqueueSnackbar).toHaveBeenCalledWith('UPDATE_APPLIED_ERROR', { variant: 'error' });
		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['update-status'] });
	});
});
