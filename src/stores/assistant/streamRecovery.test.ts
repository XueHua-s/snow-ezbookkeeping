import { expect, it, vi } from 'vitest';
import { mocks, response, conversation, snapshot, deferred, store } from './testFixture.ts';

async function liveConversation() {
    vi.useFakeTimers();
    vi.spyOn(Math, 'random').mockReturnValue(0);
    await store.initialize();
    const stream = deferred<void>();
    mocks.api['get']!.mockResolvedValue(response({ ...conversation('1', { activeRunId: '3' }), activeRun: snapshot() }));
    mocks.api['state']!.mockResolvedValue(response({ id: '1', accessible: true, contextVersion: 1, metadataVersion: 1, activeRunId: '3', activeRunRevision: 1 }));
    mocks.api['subscribe']!.mockReturnValue(stream.promise);
    await store.open('1');
    store.attach();
    await store.synchronize();
    await vi.advanceTimersByTimeAsync(0);
    return stream;
}

it.each(['reject', 'resolve'])('replaces the live-stream watchdog with prompt polling on stream %s', async ending => {
    const stream = await liveConversation();
    const checks = mocks.api['state']!.mock.calls.length;
    // The next connection stays open so only the first disconnect is measured.
    mocks.api['subscribe']!.mockReturnValue(deferred<void>().promise);
    if (ending === 'reject') stream.reject(new Error('network disconnected'));
    else stream.resolve();
    await vi.advanceTimersByTimeAsync(1999);
    expect(mocks.api['state']!.mock.calls.length).toBe(checks);
    await vi.advanceTimersByTimeAsync(1);
    expect(mocks.api['state']!.mock.calls.length).toBeGreaterThan(checks);
});

it.each(['navigation', 'identity', 'detach', 'dispose'])('does not revive polling after %s cancels the connection', async boundary => {
    const stream = await liveConversation();
    const signal = mocks.api['subscribe']!.mock.calls.at(-1)![3] as AbortSignal;
    if (boundary === 'navigation') {
        mocks.api['get']!.mockResolvedValue(response(conversation('2')));
        await store.open('2');
    } else if (boundary === 'identity') store.reset(false);
    else if (boundary === 'detach') store.detach();
    else store.$dispose();
    expect(signal.aborted).toBe(true);
    const timers = vi.getTimerCount();
    const checks = mocks.api['state']!.mock.calls.length;
    stream.reject(new Error('late disconnect'));
    await vi.advanceTimersByTimeAsync(0);
    expect(vi.getTimerCount()).toBe(timers);
    await vi.advanceTimersByTimeAsync(2000);
    expect(mocks.api['state']!.mock.calls.length).toBe(checks);
    if (boundary !== 'navigation') {
        await vi.advanceTimersByTimeAsync(18000);
        expect(mocks.api['state']!.mock.calls.length).toBe(checks);
    }
});
