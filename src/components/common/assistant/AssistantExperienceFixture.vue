<script setup lang="ts">
import { mdiChartLine, mdiChartDonut, mdiWalletOutline } from '@mdi/js';
import { computed, shallowRef } from 'vue';
import AssistantExperienceComposer from './AssistantExperienceComposer.vue';
import AssistantExperienceMessageList from './AssistantExperienceMessageList.vue';
import AssistantConversationHistory from './AssistantConversationHistory.vue';
import AssistantMemoryManager from './AssistantMemoryManager.vue';
import AssistantCapacityNotice from './AssistantCapacityNotice.vue';
import { fixtureConversation, fixtureMemory, fixtureMessage } from './experienceFixture.ts';
const props = defineProps<{
    variant: 'desktop' | 'mobile';
    state: 'empty' | 'answer' | 'failure' | 'history' | 'memory' | 'capacity';
    dark?: boolean;
}>();
const draft = shallowRef('');
const historyQuery = shallowRef('');
const historyVersions = shallowRef(false);
const lastAction = shallowRef('');
const messages = computed(() =>
    props.state === 'empty'
        ? []
        : props.state === 'failure'
          ? [
                {
                    ...fixtureMessage,
                    selectedAnswer: {
                        ...fixtureMessage.selectedAnswer!,
                        state: 'failed' as const,
                        reasonCode: 'provider_unavailable' as const
                    }
                }
            ]
          : [fixtureMessage]
);
</script>
<template>
    <main class="fixture" :class="{ dark }" :style="{ colorScheme: dark ? 'dark' : 'light' }">
        <aside v-if="state === 'history'" class="fixture-history">
            <AssistantConversationHistory
                v-model:query="historyQuery"
                v-model:include-versions="historyVersions"
                :items="[fixtureConversation]"
                :title-limit="100"
                @select="lastAction = 'Selected history'"
                @close="lastAction = 'Closed history'"
            />
        </aside>
        <section v-else-if="state === 'memory'" class="fixture-memory">
            <AssistantMemoryManager
                :items="[fixtureMemory]"
                enabled
                :byte-limit="1024"
                @close="lastAction = 'Closed memory'"
                @source="lastAction = 'Opened source'"
            />
        </section>
        <section v-else class="fixture-conversation">
            <header class="fixture-header">
                <strong>{{ fixtureConversation.title }}</strong
                ><span>AI Assistant</span>
            </header>
            <AssistantExperienceMessageList
                :messages="messages"
                :examples="[
                    { text: 'Compare this month with last month', icon: mdiChartLine },
                    { text: 'How much did I spend on dining?', icon: mdiChartDonut },
                    { text: 'Show my account balances', icon: mdiWalletOutline }
                ]"
                @example="draft = $event"
                @retry="lastAction = 'Retry requested'"
                @open-reference="lastAction = 'Opened bill'"
            /><AssistantCapacityNotice
                v-if="state === 'capacity'"
                :used="1750000"
                :reserved="65536"
                :limit="2097152"
                can-continue
                @continue="lastAction = 'Continue requested'"
            /><AssistantExperienceComposer
                v-model="draft"
                :variant="variant"
                :limits="{ inputCharacters: 4000, inputBytes: 16384 }"
                @send="lastAction = 'Send requested'"
            />
        </section>
        <output v-if="lastAction">{{ lastAction }}</output>
    </main>
</template>
<style scoped>
.fixture {
    height: 100dvh;
    display: flex;
    font-family: sans-serif;
    color: #25352d;
    background: #fff;
}
.fixture.dark {
    color: #eef4ef;
    background: #17231d;
    --assistant-surface: #17231d;
    --assistant-action: #28654c;
}
.fixture-conversation {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-width: 0;
}
.fixture-header {
    min-height: 64px;
    padding: 16px 24px;
    display: flex;
    justify-content: space-between;
    gap: 16px;
    border-bottom: 1px solid #7f7f7f60;
}
.fixture-header strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.fixture-header span {
    flex-shrink: 0;
}
.fixture-history {
    width: 280px;
    border-inline-end: 1px solid #7f7f7f60;
}
.fixture-memory {
    width: 100%;
    overflow-y: auto;
}
output {
    position: fixed;
    bottom: 0;
    background: Canvas;
    color: CanvasText;
}
@media (max-width: 480px) {
    .fixture-header {
        padding-inline: 16px;
    }
    .fixture-header span {
        display: none;
    }
    .fixture-history {
        width: 100%;
    }
}
</style>
