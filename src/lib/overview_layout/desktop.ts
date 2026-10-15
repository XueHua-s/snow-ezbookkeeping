import {
    DEFAULT_DESKTOP_OVERVIEW_LAYOUT,
    DESKTOP_OVERVIEW_LAYOUT_COLUMNS,
    DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS,
    DESKTOP_OVERVIEW_LAYOUT_MAX_WIDGETS,
    DESKTOP_OVERVIEW_WIDGET_DEFINITIONS
} from '@/consts/overview_layout.ts';
import {
    type DesktopOverviewLayout,
    type DesktopOverviewWidgetLayout,
    OverviewWidgetType
} from '@/core/overview_layout.ts';
import {
    isArray,
    isObject,
    isString,
    normalizeInteger
} from '@/lib/common.ts';
import { cloneWidget, isDesktopWidgetsOverlap, normalizeOverviewWidgetSettings } from './settings.ts';

export function resolveDesktopOverviewWidgetCollisions(widgets: DesktopOverviewWidgetLayout[], activeId?: string): DesktopOverviewWidgetLayout[] {
    const sorted: DesktopOverviewWidgetLayout[] = widgets.map(cloneWidget).sort((a, b) => {
        if (a.id === activeId) {
            return -1;
        } else if (b.id === activeId) {
            return 1;
        } else {
            return a.y - b.y || a.x - b.x || a.id.localeCompare(b.id);
        }
    });
    const placed: DesktopOverviewWidgetLayout[] = [];

    for (const widget of sorted) {
        let overlapping: DesktopOverviewWidgetLayout[] = placed.filter(other => isDesktopWidgetsOverlap(widget, other));

        while (overlapping.length > 0) {
            widget.y = Math.max(...overlapping.map(other => other.y + other.h));
            overlapping = placed.filter(other => isDesktopWidgetsOverlap(widget, other));
        }

        placed.push(widget);
    }

    return placed;
}

export function compactDesktopOverviewWidgets(widgets: DesktopOverviewWidgetLayout[], fixedId?: string): DesktopOverviewWidgetLayout[] {
    const result: DesktopOverviewWidgetLayout[] = widgets.map(cloneWidget).sort((a, b) => a.y - b.y || a.x - b.x || a.id.localeCompare(b.id));

    for (const widget of result) {
        if (widget.id === fixedId) {
            continue;
        }

        if (fixedId) {
            for (let y = 0; y < widget.y; y++) {
                const candidate = { ...widget, y };

                if (!result.some(other => other.id !== widget.id && isDesktopWidgetsOverlap(candidate, other))) {
                    widget.y = y;
                    break;
                }
            }

            continue;
        }

        while (widget.y > 0) {
            const candidate = { ...widget, y: widget.y - 1 };

            if (result.some(other => other.id !== widget.id && isDesktopWidgetsOverlap(candidate, other))) {
                break;
            }

            widget.y--;
        }
    }

    return result.sort((a, b) => a.y - b.y || a.x - b.x || a.id.localeCompare(b.id));
}

export function findDesktopOverviewWidgetPosition(widgets: DesktopOverviewWidgetLayout[], width: number, height: number): { x: number; y: number } {
    for (let y = 0; y < DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS - height; y++) {
        for (let x = 0; x <= DESKTOP_OVERVIEW_LAYOUT_COLUMNS - width; x++) {
            const candidate: DesktopOverviewWidgetLayout = {
                id: '',
                type: OverviewWidgetType.CurrentMonthOverview,
                x,
                y,
                w: width,
                h: height,
                settings: {}
            };

            if (!widgets.some(widget => isDesktopWidgetsOverlap(candidate, widget))) {
                return { x, y };
            }
        }
    }

    return { x: 0, y: Math.max(0, ...widgets.map(widget => widget.y + widget.h)) };
}

export function findDesktopOverviewWidgetDuplicatePosition(widgets: DesktopOverviewWidgetLayout[], widget: DesktopOverviewWidgetLayout): { x: number; y: number } {
    const rightPosition = { x: widget.x + widget.w, y: widget.y };
    const rightCandidate = { ...widget, ...rightPosition };

    if (rightCandidate.x + rightCandidate.w <= DESKTOP_OVERVIEW_LAYOUT_COLUMNS &&
        !widgets.some(currentWidget => isDesktopWidgetsOverlap(rightCandidate, currentWidget))) {
        return rightPosition;
    }

    return { x: widget.x, y: widget.y + widget.h };
}

export function isDefaultDesktopOverviewLayout(layout: DesktopOverviewLayout): boolean {
    return serializeDesktopOverviewLayout(layout) === serializeDesktopOverviewLayout(DEFAULT_DESKTOP_OVERVIEW_LAYOUT);
}

export function normalizeDesktopOverviewLayout(input: unknown): DesktopOverviewLayout {
    if (!isObject(input)) {
        throw new Error('input is not an object');
    }

    const source = input as Record<string, unknown>;
    const sourceWidgets = source['widgets'];

    if (!isArray(sourceWidgets)) {
        throw new Error('widgets is not an array');
    }

    const finalWidgets: DesktopOverviewWidgetLayout[] = [];
    const existsIds: Record<string, boolean> = {};

    for (const item of sourceWidgets) {
        if (!isObject(item)) {
            continue;
        }

        if (finalWidgets.length >= DESKTOP_OVERVIEW_LAYOUT_MAX_WIDGETS) {
            break;
        }

        const sourceWidget = item as Record<string, unknown>;
        const type = sourceWidget['type'] as OverviewWidgetType;
        const definition = DESKTOP_OVERVIEW_WIDGET_DEFINITIONS[type];
        const widgetId = sourceWidget['id'];

        if (!definition || !widgetId || !isString(widgetId) || widgetId.length > 100 || existsIds[widgetId]) {
            continue;
        }

        const w: number = normalizeInteger(sourceWidget['w'], definition.defaultWidth, definition.minWidth, Math.min(DESKTOP_OVERVIEW_LAYOUT_COLUMNS, definition.maxWidth ?? DESKTOP_OVERVIEW_LAYOUT_COLUMNS));
        const h: number = normalizeInteger(sourceWidget['h'], definition.defaultHeight, definition.minHeight, Math.min(DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS, definition.maxHeight ?? DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS));
        const x: number = normalizeInteger(sourceWidget['x'], 0, 0, DESKTOP_OVERVIEW_LAYOUT_COLUMNS - w);
        const y: number = normalizeInteger(sourceWidget['y'], 0, 0, DESKTOP_OVERVIEW_LAYOUT_MAX_ROWS - h);

        const finalWidget: DesktopOverviewWidgetLayout = {
            id: widgetId,
            type: type,
            x: x,
            y: y,
            w: w,
            h: h,
            settings: normalizeOverviewWidgetSettings(definition, sourceWidget['settings'])
        };

        existsIds[widgetId] = true;
        finalWidgets.push(finalWidget);
    }

    return {
        widgets: compactDesktopOverviewWidgets(resolveDesktopOverviewWidgetCollisions(finalWidgets))
    };
}

export function cloneDesktopOverviewLayout(original: DesktopOverviewLayout): DesktopOverviewLayout {
    return {
        widgets: original.widgets.map(cloneWidget)
    };
}

export function parseDesktopOverviewLayout(value: string): DesktopOverviewLayout {
    if (!value) {
        return cloneDesktopOverviewLayout(DEFAULT_DESKTOP_OVERVIEW_LAYOUT);
    }

    return normalizeDesktopOverviewLayout(JSON.parse(value));
}

export function serializeDesktopOverviewLayout(layout: DesktopOverviewLayout, pretty?: boolean): string {
    const normalized = normalizeDesktopOverviewLayout(layout);
    return JSON.stringify(normalized, null, pretty ? 4 : undefined);
}
