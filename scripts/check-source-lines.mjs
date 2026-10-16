import { readdir, readFile } from 'node:fs/promises';
import { extname, join } from 'node:path';

const limit = 500;
const extensions = new Set(['.go', '.ts', '.mts', '.cts', '.vue', '.js', '.mjs', '.cjs', '.scss', '.css', '.sh', '.py', '.html']);
const violations = [];
let checked = 0;

async function check(path, recursive = true) {
    const entries = await readdir(path, { withFileTypes: true });
    for (const entry of entries) {
        const file = join(path, entry.name);
        if (entry.isDirectory()) {
            if (recursive) await check(file);
        } else if (extensions.has(extname(file))) {
            const source = await readFile(file, 'utf8');
            const lines = source.split('\n').length - Number(source.endsWith('\n'));
            checked++;
            if (lines > limit) violations.push({ file, lines });
        }
    }
}

await check('.', false);
for (const root of ['src', 'pkg', 'cmd', 'scripts']) await check(root);
for (const { file, lines } of violations.sort((a, b) => b.lines - a.lines)) {
    console.error(`${file}: ${lines} lines (limit ${limit})`);
}
console.log(`Checked ${checked} source, test and style files; ${violations.length} exceed ${limit} lines.`);
process.exitCode = violations.length ? 1 : 0;
