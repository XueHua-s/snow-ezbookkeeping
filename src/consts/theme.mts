/** Shared cream-green palette for framework themes, charts and generated CSS tokens. */
export const APP_THEME = {
    light: {
        primary: '#27845c', primaryDarken: '#206e4c', onPrimary: '#ffffff',
        background: '#fbfcf9', surface: '#ffffff', raised: '#ffffff',
        tint: '#e8f3e9', border: '#e5ece4', text: '#233e30', muted: '#617568', shadow: '#233e30'
    },
    dark: {
        primary: '#89d8b0', primaryDarken: '#69bd94', onPrimary: '#18251f',
        background: '#18251f', surface: '#213229', raised: '#2b4436',
        tint: '#2b4436', border: '#354a3d', text: '#e5eee7', muted: '#afc1b4', shadow: '#000000'
    }
} as const;

function rgb(hex: string): string {
    return [1, 3, 5].map(offset => parseInt(hex.slice(offset, offset + 2), 16)).join(', ');
}

function tokens(theme: typeof APP_THEME.light | typeof APP_THEME.dark): string {
    return Object.entries(theme).map(([name, value]) => `--app-${name}: ${value}; --app-${name}-rgb: ${rgb(value)};`).join('\n');
}

/** Injected before application startup, including the initial loading page. */
export const APP_THEME_CSS = `
:root { ${tokens(APP_THEME.light)} color-scheme: light; }
.v-theme--dark, :root.dark, .dark { ${tokens(APP_THEME.dark)} color-scheme: dark; }
`;
