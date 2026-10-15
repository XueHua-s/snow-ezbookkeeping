<script setup lang="ts">
import { useNumberPadSheet } from './NumberPadSheet/useNumberPadSheet.ts';

const props = defineProps<{
    modelValue: number;
    minValue?: number;
    maxValue?: number;
    currency?: string;
    flipNegative?: boolean;
    hint?: string;
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: number): void;
    (e: 'update:show', value: boolean): void;
}>();

const {
    tt, currentSymbol, showPastePopover, digits, decimalSeparator, supportDecimalSeparator, currentDisplay,
    currentDisplayNumClass, confirmText, inputNum, inputDoubleNum, inputDecimalSeparator, setSymbol, backspace, clear, paste,
    confirm, onSheetOpen, onSheetClosed, onDisplayValueClick,
} = useNumberPadSheet(props, emit);

</script>

<template>
    <f7-sheet swipe-to-close swipe-handler=".swipe-handler" class="numpad-sheet" style="height: auto"
              :opened="show" @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <div class="swipe-handler"></div>
        <f7-page-content class="margin-top no-padding-top">
            <div class="margin-top padding-horizontal" v-if="hint">
                <span>{{ hint }}</span>
            </div>
            <div class="numpad-values" @click="onDisplayValueClick">
                <span id="numpad-value" class="numpad-value" :class="currentDisplayNumClass">{{ currentDisplay }}</span>
            </div>

            <f7-popover class="paste-context-menu-popover" target-el="#numpad-value"
                        v-model:opened="showPastePopover">
                <f7-list class="paste-context-menu">
                    <f7-list-item link="#" no-chevron popover-close
                                  :title="tt('Paste')" @click="paste"></f7-list-item>
                </f7-list>
            </f7-popover>

            <div class="numpad-buttons">
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(7)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[7] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(8)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[8] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(9)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[9] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-function no-right-border" @pointerup="setSymbol('×')">
                    <span class="numpad-button-text numpad-button-text-normal">&times;</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(4)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[4] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(5)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[5] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(6)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[6] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-function no-right-border" @pointerup="setSymbol('−')">
                    <span class="numpad-button-text numpad-button-text-normal">&minus;</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(1)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[1] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(2)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[2] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(3)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[3] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-function no-right-border" @pointerup="setSymbol('+')">
                    <span class="numpad-button-text numpad-button-text-normal">&plus;</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputDecimalSeparator()"
                        v-if="supportDecimalSeparator">
                    <span class="numpad-button-text numpad-button-text-normal">{{ decimalSeparator }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputDoubleNum(0)"
                        v-if="!supportDecimalSeparator">
                    <span class="numpad-button-text numpad-button-text-normal">{{ `${digits[0]}${digits[0]}` }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" @pointerup="inputNum(0)">
                    <span class="numpad-button-text numpad-button-text-normal">{{ digits[0] }}</span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-num" :aria-label="tt('Delete')" @pointerup="backspace" @taphold="clear()">
                    <span class="numpad-button-text numpad-button-text-normal">
                        <f7-icon class="icon-with-direction" f7="delete_left"></f7-icon>
                    </span>
                </f7-button>
                <f7-button class="numpad-button numpad-button-confirm no-right-border no-bottom-border" fill @click="confirm()">
                    <span :class="{ 'numpad-button-text': true, 'numpad-button-text-confirm': !currentSymbol }">{{ confirmText }}</span>
                </f7-button>
            </div>
        </f7-page-content>
    </f7-sheet>
</template>



<style>
.numpad-sheet {
    height: auto;
}

.numpad-values {
    border-bottom: 1px solid var(--f7-page-bg-color);
}

.numpad-value {
    display: flex;
    position: relative;
    padding-inline-start: 16px;
    line-height: 1;
    height: var(--ebk-numpad-value-height);
    align-items: center;
    box-sizing: border-box;
    user-select: none;
}

.numpad-value-extra-small {
    font-size: var(--ebk-numpad-value-extra-small-font-size);
}

.numpad-value-small {
    font-size: var(--ebk-numpad-value-small-font-size);
}

.numpad-value-normal {
    font-size: var(--ebk-numpad-value-normal-font-size);
}

.numpad-value-large {
    font-size: var(--ebk-numpad-value-large-font-size);
}

.numpad-buttons {
    display: flex;
    flex-wrap: wrap;
}

.numpad-button {
    display: flex;
    position: relative;
    text-align: center;
    border-radius: 0;
    border-right: 1px solid var(--f7-page-bg-color);
    border-bottom: 1px solid var(--f7-page-bg-color);
    height: 60px;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    box-sizing: border-box;
    user-select: none;
    touch-action: manipulation;
    transition: transform 0.01s ease;
}

.numpad-button-num {
    width: calc(80% / 3);
}

.numpad-button-num:active,
.numpad-button-function:active {
    background-color: var(--f7-button-pressed-bg-color, rgba(var(--f7-theme-color-rgb), 0.15));
}

.numpad-button-function, .numpad-button-confirm {
    width: 20%;
}

.numpad-button-num.active-state, .numpad-button-function.active-state {
    background-color: rgba(var(--f7-color-black-rgb), .15);
}

.numpad-button-text {
    display: block;
    font-size: var(--ebk-numpad-normal-button-font-size);
    font-weight: normal;
    line-height: 1;
}

.numpad-button-text-normal {
    color: var(--f7-color-black);
}

.dark .numpad-button-text-normal {
    color: var(--f7-color-white);
}

.numpad-button-text-confirm {
    font-size: var(--ebk-numpad-confirm-button-font-size);
}
</style>
