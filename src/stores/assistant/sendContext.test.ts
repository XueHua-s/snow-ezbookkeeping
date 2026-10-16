import { expect, it } from 'vitest';
import { getBrowserTimezoneName } from '@/lib/datetime.ts';
import { mocks, preferences, store } from './testFixture.ts';

it('persists the current language and timezone with the exact outgoing command', async () => {
    await store.initialize();
    await expect(store.send({ content: 'Compare this year with last year' })).rejects.toThrow('lost response');
    const sent = mocks.api['send']!.mock.calls[0]![0];
    expect(sent).toMatchObject({ locale: 'zh_Hans', timezone: 'Asia/Shanghai' });
    expect(store.pending[0]?.payload).toEqual(sent);
    preferences.timezone.mockReturnValue('America/New_York');
    preferences.language.mockReturnValue('en');
    await expect(store.retryPending(store.pending[0]!.clientRequestId)).rejects.toThrow('lost response');
    expect(mocks.api['send']!.mock.calls[1]![0]).toEqual(sent);
    expect(store.pending[0]?.payload).toMatchObject({ locale: 'zh_Hans', timezone: 'Asia/Shanghai' });
});

it('preserves explicit date context', async () => {
    await store.initialize();
    await expect(store.send({ content: 'Report', locale: 'ja', timezone: 'Asia/Tokyo' })).rejects.toThrow('lost response');
    expect(mocks.api['send']!.mock.calls[0]![0]).toMatchObject({ locale: 'ja', timezone: 'Asia/Tokyo' });
});

it('uses the browser timezone when the stored preference is empty', async () => {
    preferences.timezone.mockReturnValueOnce('');
    await store.initialize();
    await expect(store.send({ content: 'Report' })).rejects.toThrow('lost response');
    expect(mocks.api['send']!.mock.calls[0]![0]).toMatchObject({ timezone: getBrowserTimezoneName() });
});
