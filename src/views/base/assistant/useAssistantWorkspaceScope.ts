import { onScopeDispose, watch } from 'vue';

/** Captures UI authority across awaits, including an away-and-back transition. */
export function useAssistantWorkspaceScope(context: () => readonly unknown[], reset?: () => void) {
    let generation = 0;
    let disposed = false;
    let previous = context();
    watch(
        context,
        (value) => {
            if (value.length === previous.length && value.every((item, index) => Object.is(item, previous[index])))
                return;
            previous = value;
            generation++;
            reset?.();
        },
        { flush: 'sync' }
    );
    onScopeDispose(() => {
        disposed = true;
        generation++;
    }, true);
    return () => {
        const captured = generation;
        return () => !disposed && captured === generation;
    };
}
