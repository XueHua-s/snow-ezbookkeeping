// @env browser
import { expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-vue';
import { defineComponent, h, nextTick, shallowRef, watch } from 'vue';
import { createMemoryHistory, createRouter, RouterView, useRoute } from 'vue-router';
import MessageList from './AssistantExperienceMessageList.vue';
import { fixtureMessage } from './experienceFixture.ts';
import type { AssistantReadingAnchor } from '@/views/base/assistant/useAssistantAutoScroll.ts';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (text: string) => text }) }));

it('reuses a message list: unread A → fresh B follows latest → back restores A', async () => {
    const saved = new Map<string, AssistantReadingAnchor>();
    let list!: InstanceType<typeof MessageList>;
    const Host = defineComponent({
        setup() {
            const route = useRoute();
            const ref = shallowRef<InstanceType<typeof MessageList>>();
            watch(
                () => route.query['conversationId'],
                async () => {
                    await nextTick();
                    ref.value?.restoreAnchor(saved.get(String(route.query['conversationId'])));
                },
            );
            return () =>
                h('div', { style: 'height: 400px; display:flex; width:700px' }, [
                    h(MessageList, {
                        ref: (value: unknown) => {
                            ref.value = value as InstanceType<typeof MessageList>;
                            list = ref.value;
                        },
                        messages: Array.from({ length: 24 }, (_, index) => ({
                            ...fixtureMessage,
                            id: `${route.query['conversationId']}-${index}`,
                            selectedAnswer: undefined,
                            content: `Conversation ${route.query['conversationId']} message ${index}\nA long enough message for real browser layout.`,
                        })),
                        anchor: saved.get(String(route.query['conversationId'])),
                    }),
                ]);
        },
    });
    const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/assistant', component: Host }] });
    await router.push('/assistant?conversationId=A');
    const screen = await render(RouterView, { global: { plugins: [router] } });
    const panel = screen.container.querySelector<HTMLElement>('.experience-messages')!;
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    await new Promise(requestAnimationFrame);
    await new Promise(requestAnimationFrame);
    await new Promise(requestAnimationFrame);
    panel.scrollTop = 250;
    await expect.poll(() => list.captureAnchor()?.following).toBe(false);
    const anchor = list.captureAnchor()!;
    saved.set('A', anchor);
    const oldTop = panel.scrollTop;
    await router.push('/assistant?conversationId=B');
    await nextTick();
    // Same DOM element proves route reuse, and measurements use Chromium layout.
    expect(screen.container.querySelector('.experience-messages')).toBe(panel);
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    router.back();
    await expect.poll(() => router.currentRoute.value.query['conversationId']).toBe('A');
    await expect.poll(() => Math.abs(panel.scrollTop - oldTop)).toBeLessThan(2);
    expect(list.captureAnchor()?.messageId).toBe(anchor.messageId);
    router.forward();
    await expect.poll(() => router.currentRoute.value.query['conversationId']).toBe('B');
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
});

it('retains selected text in the real message list while new messages arrive', async () => {
    const messages = shallowRef(Array.from({ length: 16 }, (_, index) => ({
        ...fixtureMessage,
        id: String(index),
        selectedAnswer: undefined,
        content: `Message ${index}\nSelected reading content.`,
    })));
    const screen = await render(defineComponent({
        setup: () => () => h('div', { style: 'height:400px;display:flex;width:700px' }, [
            h(MessageList, { messages: messages.value })
        ])
    }));
    const panel = screen.container.querySelector<HTMLElement>('.experience-messages')!;
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    for (let frame = 0; frame < 4; frame++) await new Promise(requestAnimationFrame);
    const selected = panel.querySelector('[data-message-id="15"] .question-text')!;
    const range = document.createRange();
    range.selectNodeContents(selected);
    const selection = window.getSelection()!;
    selection.removeAllRanges();
    selection.addRange(range);
    const top = selected.getBoundingClientRect().top;
    try {
        messages.value = [...messages.value, { ...messages.value[0]!, id: '16', content: 'New message\n'.repeat(20) }];
        await expect.element(screen.getByRole('button', { name: 'New content · Back to latest' })).toBeVisible();
        for (let frame = 0; frame < 4; frame++) await new Promise(requestAnimationFrame);
        expect(Math.abs(selected.getBoundingClientRect().top - top)).toBeLessThan(2);
        expect(selection.toString()).toContain('Selected reading content.');
        selection.removeAllRanges();
        await screen.getByRole('button', { name: 'New content · Back to latest' }).click();
        await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    } finally {
        selection.removeAllRanges();
    }
});

it('follows rendered Markdown growth and preserves older reading after scrolling up', async () => {
    const messages = shallowRef([{ ...fixtureMessage, selectedAnswer: { ...fixtureMessage.selectedAnswer!, references: [] } }]);
    let list!: InstanceType<typeof MessageList>;
    const screen = await render(defineComponent({
        setup: () => () => h('div', { style: 'height:300px;display:flex;width:700px' }, [
            h(MessageList, {
                ref: (value: unknown) => { list = value as InstanceType<typeof MessageList>; },
                messages: messages.value
            })
        ])
    }));
    const panel = screen.container.querySelector<HTMLElement>('.experience-messages')!;
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    const oldHeight = panel.scrollHeight;
    messages.value = [{ ...fixtureMessage, selectedAnswer: {
        ...fixtureMessage.selectedAnswer!, references: [], content: fixtureMessage.selectedAnswer!.content + '\n\nExtra paragraph.\n\n'.repeat(20)
    } }];
    await expect.poll(() => panel.scrollHeight).toBeGreaterThan(oldHeight + 300);
    await expect.poll(() => panel.scrollHeight - panel.clientHeight - panel.scrollTop).toBeLessThan(2);
    for (let frame = 0; frame < 4; frame++) await new Promise(requestAnimationFrame);
    panel.scrollTop = 100;
    await expect.poll(() => list.captureAnchor()?.following).toBe(false);
    const anchor = list.captureAnchor();
    const grownHeight = panel.scrollHeight;
    messages.value = [{ ...messages.value[0]!, selectedAnswer: {
        ...messages.value[0]!.selectedAnswer!, content: messages.value[0]!.selectedAnswer!.content + '\n\nLater streamed paragraph.\n\n'.repeat(20)
    } }];
    await expect.poll(() => panel.scrollHeight).toBeGreaterThan(grownHeight + 300);
    for (let frame = 0; frame < 4; frame++) await new Promise(requestAnimationFrame);
    expect(list.captureAnchor()).toEqual(anchor);
});
