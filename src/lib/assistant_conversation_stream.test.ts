import { describe, expect, it, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { ASSISTANT_SNAPSHOT_FRAME_BYTES, isAssistantRunSnapshot, readAssistantConversationStream } from './assistant_conversation_stream.ts';
import { assistantReasonCodes, type AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';

const fixture = JSON.parse(readFileSync('testdata/assistant_snapshot_v1.json', 'utf8')) as AssistantRunSnapshot;
const encoder = new TextEncoder();
function frame(snapshot = fixture, event = 'run.terminal'): string {
    return `event: ${event}\nid: ${snapshot.runId}:${snapshot.revision}\ndata: ${JSON.stringify(snapshot)}\n\n`;
}
function body(text: string, chunk = 17): ReadableStream<Uint8Array> {
    const bytes = encoder.encode(text);
    return new ReadableStream({ start(controller) {
        for (let offset = 0; offset < bytes.length; offset += chunk) { controller.enqueue(bytes.slice(offset, offset + chunk)); }
        controller.close();
    } });
}
describe('persistent assistant snapshots', () => {
    it('accepts every declared reason code and rejects unknown wire values', () => {
        for (const reasonCode of assistantReasonCodes) {
            expect(isAssistantRunSnapshot({ ...fixture, reasonCode })).toBe(true);
        }
        expect(isAssistantRunSnapshot({ ...fixture, reasonCode: 'future_unknown' })).toBe(false);
        expect(isAssistantRunSnapshot({ ...fixture, reasonCode: 1 })).toBe(false);
    });
    it('consumes the Go/TS fixture without rounding signed int64 amounts or IDs', async () => {
        const onSnapshot = vi.fn();
        await readAssistantConversationStream(body(frame()), fixture.runId, { onSnapshot });
        expect(onSnapshot.mock.calls[0]![0].references[0].sourceAmountMinor).toBe('-9223372036854775808');
        expect(onSnapshot.mock.calls[0]![0].references[0].destinationAmountMinor).toBe('9223372036854775807');
        expect(onSnapshot.mock.calls[0]![0].runId).toBe('9223372036854775807');
    });
    it.each(['\n', '\r\n', '\r'])('handles split UTF8 and %j delimiters', async delimiter => {
        const onTerminal = vi.fn();
        await readAssistantConversationStream(body(frame().replaceAll('\n', delimiter), 1), fixture.runId, { onSnapshot: vi.fn(), onTerminal });
        expect(onTerminal).toHaveBeenCalledExactlyOnceWith(fixture);
    });
    it('replaces snapshots and ignores duplicate/out-of-order revisions', async () => {
        const onSnapshot = vi.fn();
        const running = { ...fixture, revision: 10, state: 'running' as const, content: '早' };
        await readAssistantConversationStream(body(frame(running, 'message.snapshot') + frame(running, 'message.snapshot') +
            frame({ ...running, revision: 9 }, 'message.snapshot') + frame()), fixture.runId, { onSnapshot });
        expect(onSnapshot.mock.calls.map(call => call[0].content)).toEqual(['早', fixture.content]);
    });
    it.each([
        frame().replace('schemaVersion":1', 'schemaVersion":2'),
        frame().replace(`id: ${fixture.runId}:12`, 'id: 1:12'),
        frame({ ...fixture, state: 'running' }, 'run.terminal'),
        frame(fixture, 'message.snapshot'),
        frame({ ...fixture, content: '  ' }),
        frame() + frame(),
        'data: {broken}\n\n'
    ])('rejects invalid frames without a fabricated terminal', async text => {
        const onTerminal = vi.fn();
        await expect(readAssistantConversationStream(body(text), fixture.runId, { onSnapshot: vi.fn(), onTerminal })).rejects.toMatchObject({ kind: 'protocol' });
        expect(onTerminal).not.toHaveBeenCalled();
    });
    it('rejects truncated and unterminated streams', async () => {
        for (const text of [frame().slice(0, -1), frame({ ...fixture, state: 'running' }, 'message.snapshot')]) {
            await expect(readAssistantConversationStream(body(text), fixture.runId, { onSnapshot: vi.fn() })).rejects.toMatchObject({ kind: 'disconnected' });
        }
    });
    it('limits actual UTF8 bytes including headers and unfinished lines', async () => {
        for (const text of [': ' + '中'.repeat(ASSISTANT_SNAPSHOT_FRAME_BYTES / 2), frame({ ...fixture, content: '中'.repeat(90000) })]) {
            await expect(readAssistantConversationStream(body(text, 4096), fixture.runId, { onSnapshot: vi.fn() })).rejects.toMatchObject({ kind: 'protocol' });
        }
    });
    it('does not accept numeric money or IDs and stops a canceled subscription', async () => {
        expect(isAssistantRunSnapshot({ ...fixture, runId: 123 })).toBe(false);
        expect(isAssistantRunSnapshot({ ...fixture, references: [{ ...fixture.references[0], sourceAmountMinor: 123 }] })).toBe(false);
        const controller = new AbortController();
        controller.abort();
        await expect(readAssistantConversationStream(body(frame()), fixture.runId, { onSnapshot: vi.fn() }, controller.signal)).rejects.toMatchObject({ kind: 'canceled' });
    });
});
