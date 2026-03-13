import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Box, Typography, Button } from '@mui/material';

interface Props {
	children: ReactNode;
}

interface State {
	hasError: boolean;
	error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = { hasError: false, error: null };
	}

	static getDerivedStateFromError(error: Error): State {
		return { hasError: true, error };
	}

	componentDidCatch(error: Error, errorInfo: ErrorInfo) {
		console.error('ErrorBoundary caught:', error, errorInfo);
	}

	handleReset = () => {
		this.setState({ hasError: false, error: null });
	};

	render() {
		if (this.state.hasError) {
			return (
				<Box
					sx={{
						display: 'flex',
						flexDirection: 'column',
						alignItems: 'center',
						justifyContent: 'center',
						minHeight: '100vh',
						gap: 2,
						p: 4,
					}}
				>
					<Typography variant='h5'>Something went wrong</Typography>
					<Typography variant='body2' color='text.secondary' sx={{ maxWidth: 600, textAlign: 'center' }}>
						{this.state.error?.message}
					</Typography>
					<Button variant='contained' onClick={this.handleReset}>
						Try again
					</Button>
				</Box>
			);
		}

		return this.props.children;
	}
}

export default ErrorBoundary;
