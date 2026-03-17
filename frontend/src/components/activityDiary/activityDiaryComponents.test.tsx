import { act, fireEvent, render, screen } from '@testing-library/react';
import ActivityDiaryActionBar from './ActivityDiaryActionBar/ActivityDiaryActionBar';
import ActivityDiaryForm from './ActivityDiaryForm/ActivityDiaryForm';
import ActivityList from './ActivityList/ActivityList';
import ActivitySummary from './ActivitySummary/ActivitySummary';
import ActivityDiaryLayout from './activityDiaryLayout';
import ActivityDiaryPage from '@/pages/activityDiary';

jest.mock('@/components/providers/activityDiaryProvider/ActivityDiaryContext', () => ({
    useActivityDiary: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
    <div>
        <h2>{title}</h2>
        {children}
    </div>
));

jest.mock('../layout/Layout', () => ({ children }: any) => (
    <div data-testid="layout">{children}</div>
));
jest.mock('@/components/activityDiary/activityDiaryLayout', () => ({ children }: any) => (
    <div data-testid="activity-layout">{children}</div>
));
jest.mock('@/components/activityDiary/ActivityDiaryActionBar', () => () => <div>ActionBar</div>);
jest.mock('@/components/activityDiary/ActivityDiaryForm', () => () => <div>Form</div>);
jest.mock('@/components/activityDiary/ActivitySummary', () => () => <div>Summary</div>);
jest.mock('@/components/activityDiary/ActivityList', () => () => <div>List</div>);

const { useActivityDiary } = jest.requireMock(
    '@/components/providers/activityDiaryProvider/ActivityDiaryContext'
);

const defaultCtx = {
    message: undefined,
    form: { name: 'Task', description: 'Desc' },
    handleSubmit: jest.fn((e: any) => e.preventDefault()),
    handleNameChange: jest.fn(),
    handleDescriptionChange: jest.fn(),
    data: {
        summary: {
            total_activities: 2,
            total_time_spent_seconds: 3600,
            longest_activity: {
                name: 'Task',
                duration_seconds: 1800,
                duration_formatted: '30m',
            },
        },
        entries: {
            items: [
                {
                    id: 1,
                    name: 'A1',
                    description: 'D1',
                    start_time: '2026-01-01T00:00:00Z',
                    end_time: { HasValue: true, Value: '2026-01-01T00:30:00Z' },
                    duration: 1800,
                },
            ],
        },
    },
    getCurrentDuration: jest.fn(() => 12),
    copyActivity: jest.fn(),
};

describe('activity diary components', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        useActivityDiary.mockReturnValue(defaultCtx);
    });

    it('renders action bar with and without message', () => {
        render(<ActivityDiaryActionBar />);
        expect(screen.getByText('ACTIVITY_DIARY_TITLE')).toBeInTheDocument();

        useActivityDiary.mockReturnValueOnce({
            ...defaultCtx,
            message: { text: 'Erro', type: 'error' },
        });
        render(<ActivityDiaryActionBar />);
        expect(screen.getByText('Erro')).toBeInTheDocument();
    });

    it('submits form and triggers field handlers', () => {
        render(<ActivityDiaryForm />);
        fireEvent.change(screen.getByRole('textbox', { name: /ACTIVITY_NAME_LABEL/i }), {
            target: { value: 'New Name' },
        });
        fireEvent.change(screen.getAllByRole('textbox')[1]!, {
            target: { value: 'New Description' },
        });
        fireEvent.click(screen.getByText('ADD_ACTIVITY'));

        expect(defaultCtx.handleNameChange).toHaveBeenCalled();
        expect(defaultCtx.handleDescriptionChange).toHaveBeenCalled();
        expect(defaultCtx.handleSubmit).toHaveBeenCalled();
    });

    it('renders list entries and copy action', () => {
        render(<ActivityList />);
        expect(screen.getByText('A1')).toBeInTheDocument();
        expect(screen.getByText('D1')).toBeInTheDocument();

        act(() => {
            screen.getAllByRole('button')[0]!.click();
        });
        expect(defaultCtx.copyActivity).toHaveBeenCalled();
    });

    it('renders no activities branch and in-progress duration branch', () => {
        useActivityDiary.mockReturnValueOnce({
            ...defaultCtx,
            data: {
                ...defaultCtx.data,
                entries: { items: [] },
            },
        });
        render(<ActivityList />);
        expect(screen.getByText('NO_ACTIVITIES')).toBeInTheDocument();

        useActivityDiary.mockReturnValueOnce({
            ...defaultCtx,
            data: {
                ...defaultCtx.data,
                entries: {
                    items: [
                        {
                            id: 2,
                            name: 'A2',
                            description: '',
                            start_time: '2026-01-01T00:00:00Z',
                            end_time: { HasValue: false, Value: '' },
                            duration: undefined,
                        },
                    ],
                },
            },
        });
        render(<ActivityList />);
        expect(screen.getByText('IN_PROGRESS')).toBeInTheDocument();
    });

    it('renders summary values', () => {
        render(<ActivitySummary />);
        expect(screen.getByText('TOTAL_ACTIVITIES')).toBeInTheDocument();
        expect(screen.getByText('LONGEST_ACTIVITY')).toBeInTheDocument();
        expect(screen.getByText('Task')).toBeInTheDocument();
    });

    it('renders summary without longest activity subtitle', () => {
        useActivityDiary.mockReturnValueOnce({
            ...defaultCtx,
            data: {
                ...defaultCtx.data,
                summary: {
                    ...defaultCtx.data.summary,
                    longest_activity: undefined,
                },
            },
        });
        render(<ActivitySummary />);
        expect(screen.getByText('LONGEST_ACTIVITY')).toBeInTheDocument();
    });

    it('renders activity layouts and page composition', () => {
        render(
            <ActivityDiaryLayout>
                <div>child</div>
            </ActivityDiaryLayout>
        );
        expect(screen.getByTestId('activity-layout')).toBeInTheDocument();
        expect(screen.getByText('child')).toBeInTheDocument();

        render(<ActivityDiaryPage />);
        expect(screen.getAllByTestId('activity-layout').length).toBeGreaterThan(0);
        expect(screen.getByText('ActionBar')).toBeInTheDocument();
        expect(screen.getByText('Form')).toBeInTheDocument();
        expect(screen.getByText('Summary')).toBeInTheDocument();
        expect(screen.getByText('List')).toBeInTheDocument();
    });
});
