// @vitest-environment happy-dom
import { describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import Actions from './AssistantExperienceActions.vue';
import MessageList from './AssistantExperienceMessageList.vue';
import Evidence from './AssistantExperienceEvidence.vue';
import { fixtureMessage } from './experienceFixture.ts';
vi.mock('@/locales/helpers.ts', () => ({ useI18n: () => ({ tt: (key: string) => key }) }));
const versions = Array.from({ length: 40 }, (_, i) => ({ id: String(i + 1), attemptNumber: i + 1, state: 'completed' as const, createdAt: '' }));
describe('answer browsing', () => {
    it('navigates both directions through all loaded attempts without a select menu', async () => {
        const wrapper = mount(Actions, { props: { content: '', versions, selectedVersionId: '40', selectedAttemptNumber: 40 } });
        expect(wrapper.find('select').exists()).toBe(false);
        for (let i = 40; i > 1; i--) {
            await wrapper.setProps({ selectedVersionId: String(i), selectedAttemptNumber: i });
            await wrapper.get('[aria-label="Earlier answer"]').trigger('click');
            expect(wrapper.emitted('selectVersion')?.at(-1)).toEqual([String(i - 1)]);
        }
        for (let i = 1; i < 40; i++) {
            await wrapper.setProps({ selectedVersionId: String(i), selectedAttemptNumber: i });
            await wrapper.get('[aria-label="Newer answer"]').trigger('click');
            expect(wrapper.emitted('selectVersion')?.at(-1)).toEqual([String(i + 1)]);
        }
        expect(wrapper.emitted('useAnswer')).toBeUndefined();
        wrapper.unmount();
    });
    it('includes a search-selected attempt outside the loaded page and disables actions while busy', async () => {
        const wrapper = mount(Actions, { props: { content: '', versions: versions.slice(20), selectedVersionId: '5', selectedAttemptNumber: 5, hasMoreVersions: true, canUseAnswer: true } });
        expect(wrapper.text()).toContain('5 / 40');
        await wrapper.get('[aria-label="Newer answer"]').trigger('click');
        expect(wrapper.emitted('selectVersion')).toEqual([['21']]);
        await wrapper.setProps({ disabledReason: 'Working' });
        expect(wrapper.findAll('button').every(button => button.attributes('disabled') !== undefined)).toBe(true);
        wrapper.unmount();
    });
    it.each(['completed', 'failed'] as const)('offers adoption only for a completed latest alternative (%s)', state => {
        const message = { ...fixtureMessage, selectedAssistantMessageId: 'other', selectedAnswer: { ...fixtureMessage.selectedAnswer!, state, pageableQueryId: 'query' } };
        const wrapper = mount(MessageList, { props: { messages: [message], latestMessageId: message.id }, global: { stubs: { AssistantMarkdownContent: true, AssistantExperienceEvidence: true } } });
        const actions = wrapper.findAllComponents(Actions).at(-1)!;
        expect(actions.props('canUseAnswer')).toBe(state === 'completed');
        expect(wrapper.findComponent(Evidence).props('canPage')).toBe(false);
        wrapper.unmount();
    });
    it('allows a historical follow-up as a new conversation without changing its context', () => {
        const message = { ...fixtureMessage, selectedAssistantMessageId: 'other', selectedAnswer: { ...fixtureMessage.selectedAnswer!, pageableQueryId: 'query' } };
        const wrapper = mount(MessageList, { props: { messages: [message], latestMessageId: 'latest' }, global: { stubs: { AssistantMarkdownContent: true, AssistantExperienceEvidence: true } } });
        expect(wrapper.findAllComponents(Actions).at(-1)!.props('canUseAnswer')).toBe(false);
        expect(wrapper.findComponent(Evidence).props()).toMatchObject({ canPage: true, pageLabel: 'Continue finding bills in a new conversation' });
        wrapper.unmount();
    });
});

it.each(['Offline. Drafts are saved on this device.', 'Unarchive this conversation to continue.'])('explains disabled writes while leaving answer browsing available: %s', async reason => {
    const message = { ...fixtureMessage, selectedAssistantMessageId: 'other', selectedAnswer: { ...fixtureMessage.selectedAnswer!, pageableQueryId: 'query' } };
    const wrapper = mount(MessageList, { props: { messages: [message], latestMessageId: message.id, writeDisabledReason: reason }, global: { stubs: { AssistantMarkdownContent: true, AssistantExperienceEvidence: true } } });
    const actions = wrapper.findAllComponents(Actions).at(-1)!;
    const adopt = actions.findAll('button').find(button => button.text() === 'Continue with this answer')!;
    expect(adopt.attributes('disabled')).toBeDefined();
    expect(adopt.attributes('title')).toBe(reason);
    expect(actions.get('[aria-label="Regenerate"]').attributes('disabled')).toBeDefined();
    expect(actions.get('[aria-label="Newer answer"]').attributes('disabled')).toBeUndefined();
    await actions.get('[aria-label="Newer answer"]').trigger('click');
    expect(wrapper.emitted('selectVersion')).toHaveLength(1);
    expect(wrapper.findComponent(Evidence).props('disabledReason')).toBe(reason);
    wrapper.unmount();
    const evidence = mount(Evidence, { props: { answer: { ...message.selectedAnswer!, references: [] }, canPage: true, disabledReason: reason } });
    const followUp = evidence.get('.more-bills');
    expect(followUp.attributes('disabled')).toBeDefined();
    expect(followUp.attributes('title')).toBe(reason);
    await followUp.trigger('click');
    expect(evidence.emitted('nextPage')).toBeUndefined();
    evidence.unmount();
});
