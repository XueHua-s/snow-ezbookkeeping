import { APP_THEME_CSS } from '../consts/theme.mts';

// Component tests bypass the HTML entrypoints where Vite installs the application theme.
const style = document.createElement('style');
style.id = 'application-theme';
style.textContent = APP_THEME_CSS;
document.head.append(style);
