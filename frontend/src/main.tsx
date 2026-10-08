import './index.css';

import { applyPersistedColorScheme } from '@/components/providers/colorSchemeProvider/colorScheme';
import App from './app/App.tsx';
import { createRoot } from 'react-dom/client';

applyPersistedColorScheme();

createRoot(document.getElementById('root')!).render(<App />);
