import {
    DEFAULT_MOBILE_OVERVIEW_LAYOUT,
    MOBILE_OVERVIEW_LAYOUT_MAX_WIDGETS,
    MOBILE_OVERVIEW_WIDGET_DEFINITIONS
} from '@/consts/overview_layout.ts';
import type { MobileOverviewLayout, MobileOverviewWidgetLayout, OverviewWidgetType } from '@/core/overview_layout.ts';
import {
    isArray,
    isObject,
    isString
} from '@/lib/common.ts';
import { cloneWidget, normalizeOverviewWidgetSettings } from './settings.ts';

export function isDefaultMobileOverviewLayout(layout: MobileOverviewLayout): boolean {
    return serializeMobileOverviewLayout(layout) === serializeMobileOverviewLayout(DEFAULT_MOBILE_OVERVIEW_LAYOUT);
}

export function normalizeMobileOverviewLayout(input: unknown): MobileOverviewLayout {
    if (!isObject(input)) {
        throw new Error('input is not an object');
    }

    const source = input as Record<string, unknown>;
    const sourceWidgets = source['widgets'];

    if (!isArray(sourceWidgets)) {
        throw new Error('widgets is not an array');
    }

    const finalWidgets: MobileOverviewWidgetLayout[] = [];
    const existsIds: Record<string, boolean> = {};

    for (const item of sourceWidgets) {
        if (!isObject(item)) {
            continue;
        }

        if (finalWidgets.length >= MOBILE_OVERVIEW_LAYOUT_MAX_WIDGETS) {
            break;
        }

        const sourceWidget = item as Record<string, unknown>;
        const type = sourceWidget['type'] as OverviewWidgetType;
        const definition = MOBILE_OVERVIEW_WIDGET_DEFINITIONS[type];
        const widgetId = sourceWidget['id'];

        if (!definition || !widgetId || !isString(widgetId) || widgetId.length > 100 || existsIds[widgetId]) {
            continue;
        }

        const finalWidget: MobileOverviewWidgetLayout = {
            id: widgetId,
            type: type,
            settings: normalizeOverviewWidgetSettings(definition, sourceWidget['settings'])
        }

        existsIds[widgetId] = true;
        finalWidgets.push(finalWidget);
    }

    return {
        widgets: finalWidgets
    };
}

export function cloneMobileOverviewLayout(original: MobileOverviewLayout): MobileOverviewLayout {
    return {
        widgets: original.widgets.map(cloneWidget)
    };
}

export function parseMobileOverviewLayout(value: string): MobileOverviewLayout {
    if (!value) {
        return cloneMobileOverviewLayout(DEFAULT_MOBILE_OVERVIEW_LAYOUT);
    }

    return normalizeMobileOverviewLayout(JSON.parse(value));
}

export function serializeMobileOverviewLayout(layout: MobileOverviewLayout, pretty?: boolean): string {
    const normalized = normalizeMobileOverviewLayout(layout);
    return JSON.stringify(normalized, null, pretty ? 2 : undefined);
}
