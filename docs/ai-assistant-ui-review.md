# Assistant UI review — 2026-10-15

This report distinguishes shared-component fixtures from the real application. It is supplementary evidence for R04–R07, R27, and R40, not certification of every requirement in the experience plan.

## Real application review

The reviewer used Chromium through Playwright against the actual desktop and mobile entry points, with the real Vue router/Framework7 router, Pinia store, authenticated HTTP API, isolated SQLite database, synthetic ledger, and deterministic OpenAI-protocol provider. No page/component/store stubs were used. The disposable environment was created by `scripts/assistant-experience-e2e-environment.mjs`. The reviewer did not implement the platform shell; the parent independently inspected the screenshots as well.

The UI used the current working tree. The review server used the backend development binary built at 18:58 on 2026-10-15, so these checks do not replace the final rebuild and backend acceptance suite. Provider text was deterministic test data, not a real-model accuracy evaluation. Screenshots contain only synthetic account data.

Desktop viewport: 1280 × 900, then 720 × 900 for the narrow drawer and memory view. Mobile viewport: 390 × 844. Browser zoom: 100%. Light theme. Detailed assertion output: [review.json](ai-assistant-ui-review/review.json).

| Check | Desktop | Mobile | Evidence |
| --- | --- | --- | --- |
| Toolbar fits without horizontal page overflow | PASS, 61 px high | PASS, approximately 95 px high | Answer screenshots and measured dimensions |
| Markdown clipboard preserves original headings, emphasis, code fences, and newlines | PASS | PASS | Browser clipboard readback compared with exact provider source |
| Code clipboard preserves indentation and trailing newline | PASS | PASS | Browser clipboard readback: two leading spaces and final newline |
| Drawer contains focus over 18 successive Tab presses | PASS at 720 px | PASS | Actual native keyboard events and activeElement assertions |
| Drawer Esc closes and restores trigger focus | PASS | PASS | Actual keyboard event and focus assertion |
| More menu Esc closes and restores summary focus | PASS | PASS | Native details state and focus assertion |
| Memory view opens, Close is reachable, and closing retains meaningful focus | PASS | PASS | Actual pointer interaction and focus assertion |
| Textarea has no native inner border/outline | PASS | PASS | Computed border `0px`, outline `none`; focus indicator belongs to outer composer |
| Unhandled page exceptions during these flows | 0 | 0 | `pageerror` listener |

The clipboard checks granted clipboard permissions to a disposable browser context. Permission-denial/fallback and click-time streaming snapshot behavior were separately covered in component tests; they were not represented as browser permission-denial tests.

### Screenshots personally inspected

| State | Desktop | Mobile |
| --- | --- | --- |
| Answer, ledger reference, and reply actions | [1280 × 900](ai-assistant-ui-review/desktop-answer.png) | [390 × 844](ai-assistant-ui-review/mobile-answer.png) |
| Narrow history drawer and keyboard focus | [720 × 900](ai-assistant-ui-review/desktop-history.png) | [390 × 844](ai-assistant-ui-review/mobile-history.png) |
| Memory management after loading completes | [720 × 900](ai-assistant-ui-review/desktop-memory.png) | [390 × 844](ai-assistant-ui-review/mobile-memory.png) |

The real-route review caught issues that the component fixture suite did not:

1. Framework7/global button widths stacked mobile toolbar and reply actions. Explicit widths restored a compact toolbar and visible action row.
2. The history focus trap included controls inside closed `details`, allowing Tab to escape to the document. The shell now filters those descendants and explicitly cycles visible controls.
3. An open More menu obscured the memory Close button. Toolbar actions now close the menu before entering the destination view.
4. Framework7's `.toggle` class collided with the memory label and overlapped its text. The component now uses a feature-specific class and an explicit checkbox size.
5. A global code color had poor contrast, and the original composer used a hard-coded green. Code inherits foreground color; the send button and outer focus indicator use the application theme.

All five findings were corrected and rechecked before these final screenshots were saved.

## Shared-component checks

`src/components/common/assistant/AssistantExperience.test.ts` passed 9 tests after adding the server-authoritative tail check. They cover effective Unicode/code-point and UTF-8 limits, IME confirmation, editable next draft during generation, clipboard snapshots and refusal fallback, Markdown safety, raw code copying, message-relative reading anchors, exact legacy amount fallback, and historical-window action availability.

`src/components/common/assistant/AssistantExperience.browser.test.ts` passed 13 Chromium cases before final real-platform CSS refinements. These fixtures cover desktop/mobile empty, answer/table/bill, failure, history, memory, and capacity states in light/dark themes, plus RTL with enlarged text. The final platform CSS refinements were checked in the real routes above. Fixture screenshots alone are not used to claim actual routing or device behavior.

Language audit at the implementation handoff found no untranslated newly added keys in Simplified or Traditional Chinese. English and all other language catalogs contain fallback entries. Run-state and controlled failure-reason labels include queue timeout, run timeout, execution limit, and generation unavailable; raw reason codes are not intended as UI labels. Source-line and targeted ESLint checks passed during implementation; final whole-repository gates are recorded in the main verification report.

## Remaining scope for final acceptance

The checks above do not establish physical iOS/Android soft-keyboard behavior, screen-reader behavior, every UI state under every theme/viewport, all failure/recovery routes, or real-model correctness. Those must retain their own PASS/FAIL/NOT RUN results in the main verification report. The complete R01–R40 audit, production build, database matrix, and release decision remain the parent delivery audit's responsibility.

## Extended real-route matrix

The follow-up used the refreshed isolated backend, actual application routes and settings, real store/API, and deterministic provider. Desktop dark mode, RTL, narrow 640 × 900, and genuine 200% browser zoom passed horizontal-fit, composer-visibility, and page-exception assertions. Images: [dark](ai-assistant-ui-review/desktop-dark.png), [RTL](ai-assistant-ui-review/desktop-rtl.png), [narrow](ai-assistant-ui-review/desktop-narrow.png), and [200% zoom](ai-assistant-ui-review/desktop-zoom200.png). The reviewer personally inspected the dark, RTL, and zoom images.

The 200% case used Chromium `chrome.tabs.setZoom(2)` in a disposable profile with an isolated helper extension. A 1280 × 900 physical viewport reported a 640 × 450 CSS viewport and device pixel ratio 2, rather than CSS zoom or text enlargement. RTL used the application's developer language-preview configuration, enabled only in the isolated review backend, and computed workspace direction was `rtl`. Dark mode used the application's stored theme preference and computed a dark workspace background. No DOM direction or theme classes were mocked. The matrix sends and renders an actual API-backed answer in each case. Detailed measured results are in [extended-review.json](ai-assistant-ui-review/extended-review.json).

Mobile dark and RTL at 390 × 844 and small portrait at 320 × 568 passed the same measured assertions and personal screenshot inspection: [dark](ai-assistant-ui-review/mobile-dark.png), [RTL](ai-assistant-ui-review/mobile-rtl.png), [small portrait](ai-assistant-ui-review/mobile-small.png). Mobile RTL uses the application's `?rtl` entry-point parameter alongside language preview, loading the native RTL stylesheet.

**Landscape correction and recheck:** the initial 568 × 320 layout left almost no readable message space. A mobile-landscape-only CSS rule now keeps the toolbar on one line, retains 44 px touch controls, reduces vertical padding, and caps the textarea at 44 px with internal scrolling. The [final screenshot](ai-assistant-ui-review/mobile-landscape.png) was personally inspected after a real API-backed answer and a five-line unsent draft: the message viewport is **130 px**, the textarea is **44 px**, and horizontal fit, composer visibility, and page-exception assertions pass. This final recheck used a fresh isolated review environment. Production UI is frozen after this correction.

The real desktop confirmation flow was also checked: More → Clear all assistant data → Cancel. After the focus-restoration correction and a 500 ms settle, the active element is the visible More `summary`. No clear operation was executed. Long-history prepend review identified a measurement issue: newly inserted previous content can enter the blank space above the original first message; checking the original message ID's offset gives less than 0.5 px drift across three real API-backed prepends, as independently recorded by the store verification runner.
