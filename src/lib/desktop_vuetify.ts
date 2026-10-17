import { APP_THEME } from '@/consts/theme.mts';
import { DecimalSeparator, DigitGroupingSymbol } from '@/core/numeral.ts';
import { getRtlLocales } from '@/locales/helpers.ts';
import { ref } from 'vue';
import { type Composer } from 'vue-i18n';
import { createVuetify } from 'vuetify';
import { VAlert } from 'vuetify/components/VAlert';
import { VApp } from 'vuetify/components/VApp';
import { VAutocomplete } from 'vuetify/components/VAutocomplete';
import { VAvatar } from 'vuetify/components/VAvatar';
import { VBadge } from 'vuetify/components/VBadge';
import { VBtn } from 'vuetify/components/VBtn';
import { VBtnGroup } from 'vuetify/components/VBtnGroup';
import { VBtnToggle } from 'vuetify/components/VBtnToggle';
import { VCard, VCardActions, VCardItem, VCardSubtitle, VCardText, VCardTitle } from 'vuetify/components/VCard';
import { VCheckbox, VCheckboxBtn } from 'vuetify/components/VCheckbox';
import { VChip } from 'vuetify/components/VChip';
import { VColorPicker } from 'vuetify/components/VColorPicker';
import { VDataTable } from 'vuetify/components/VDataTable';
import { VDialog } from 'vuetify/components/VDialog';
import { VDivider } from 'vuetify/components/VDivider';
import { VExpansionPanel, VExpansionPanelText, VExpansionPanelTitle, VExpansionPanels } from 'vuetify/components/VExpansionPanel';
import { VForm } from 'vuetify/components/VForm';
import { VCol, VContainer, VRow, VSpacer } from 'vuetify/components/VGrid';
import { VIcon } from 'vuetify/components/VIcon';
import { VImg } from 'vuetify/components/VImg';
import { VInput } from 'vuetify/components/VInput';
import { VLabel } from 'vuetify/components/VLabel';
import { VLayout } from 'vuetify/components/VLayout';
import { VList, VListGroup, VListImg, VListItem, VListItemAction, VListItemMedia, VListItemSubtitle, VListItemTitle, VListSubheader } from 'vuetify/components/VList';
import { VMain } from 'vuetify/components/VMain';
import { VMenu } from 'vuetify/components/VMenu';
import { VNavigationDrawer } from 'vuetify/components/VNavigationDrawer';
import { VOverlay } from 'vuetify/components/VOverlay';
import { VPagination } from 'vuetify/components/VPagination';
import { VProgressCircular } from 'vuetify/components/VProgressCircular';
import { VProgressLinear } from 'vuetify/components/VProgressLinear';
import { VSelect } from 'vuetify/components/VSelect';
import { VSkeletonLoader } from 'vuetify/components/VSkeletonLoader';
import { VSlideGroup, VSlideGroupItem } from 'vuetify/components/VSlideGroup';
import { VSnackbar } from 'vuetify/components/VSnackbar';
import { VSwitch } from 'vuetify/components/VSwitch';
import { VTable } from 'vuetify/components/VTable';
import { VTab, VTabs } from 'vuetify/components/VTabs';
import { VTextarea } from 'vuetify/components/VTextarea';
import { VTextField } from 'vuetify/components/VTextField';
import { VToolbar } from 'vuetify/components/VToolbar';
import { VTooltip } from 'vuetify/components/VTooltip';
import { VWindow, VWindowItem } from 'vuetify/components/VWindow';
import { aliases, mdi } from 'vuetify/iconsets/mdi-svg';
import type { LocaleInstance } from 'vuetify/lib/framework.d.ts';



// key is in the original i18n text of vuetify (in vuetify/lib/locale/en.js), value is the text in the ezBookkeeping i18n files
const vuetifyI18nTextKeyMap: Record<string, string> = {
    'open': 'Open',
    'close': 'Close'
}

// Vuetify component defaults, theme tokens and locale adapter share one owner.
export function createDesktopVuetify(i18nGlobal: Composer) {
    return createVuetify({
    components: {
        VAlert,
        VApp,
        VAvatar,
        VAutocomplete,
        VBadge,
        VBtn,
        VBtnGroup,
        VBtnToggle,
        VCard,
        VCardActions,
        VCardItem,
        VCardSubtitle,
        VCardText,
        VCardTitle,
        VCheckbox,
        VCheckboxBtn,
        VChip,
        VColorPicker,
        VDataTable,
        VDialog,
        VDivider,
        VExpansionPanel,
        VExpansionPanelText,
        VExpansionPanelTitle,
        VExpansionPanels,
        VForm,
        VContainer,
        VCol,
        VRow,
        VSpacer,
        VIcon,
        VImg,
        VInput,
        VLabel,
        VLayout,
        VList,
        VListGroup,
        VListImg,
        VListItem,
        VListItemAction,
        VListItemMedia,
        VListItemSubtitle,
        VListItemTitle,
        VListSubheader,
        VMain,
        VMenu,
        VNavigationDrawer,
        VOverlay,
        VPagination,
        VProgressCircular,
        VProgressLinear,
        VSelect,
        VSkeletonLoader,
        VSlideGroup,
        VSlideGroupItem,
        VSnackbar,
        VSwitch,
        VTabs,
        VTab,
        VTable,
        VTextarea,
        VTextField,
        VToolbar,
        VTooltip,
        VWindow,
        VWindowItem
    },
    icons: {
        defaultSet: 'mdi',
        aliases,
        sets: {
            mdi
        }
    },
    defaults: {
        VAlert: {
            density: 'comfortable',
            VBtn: {
                color: undefined
            }
        },
        VAutocomplete: {
            variant: 'outlined',
            density: 'comfortable',
            color: 'primary',
            hideDetails: 'auto'
        },
        VAvatar: {
            variant: 'flat',
            VIcon: {
                size: 24,
            },
        },
        VBadge: {
            color: 'primary'
        },
        VBtn: {
            variant: 'flat',
            color: 'primary'
        },
        VCheckbox: {
            color: 'primary',
            density: 'compact',
            hideDetails: 'auto'
        },
        VChip: {
            elevation: 0
        },
        VList: {
            color: 'primary',
            density: 'compact'
        },
        VPagination: {
            density: 'comfortable',
            activeColor: 'primary'
        },
        VRadio: {
            density: 'compact',
            color: 'primary',
            hideDetails: 'auto'
        },
        VSelect: {
            variant: 'outlined',
            density: 'comfortable',
            color: 'primary',
            hideDetails: 'auto'
        },
        VSlider: {
            color: 'primary',
            hideDetails: 'auto'
        },
        VSwitch: {
            inset: true,
            density: 'compact',
            color: 'primary',
            hideDetails: 'auto'
        },
        VProgressCircular: {
            size: 40
        },
        VSnackbar: {
            timeout: 3000,
            zIndex: 9000
        },
        VTable: {
            hover: true,
            density: 'comfortable'
        },
        VTabs: {
            color: 'primary',
            VSlideGroup: {
                showArrows: true
            }
        },
        VTextarea: {
            variant: 'outlined',
            density: 'comfortable',
            color: 'primary',
            hideDetails: 'auto'
        },
        VTextField: {
            variant: 'outlined',
            density: 'comfortable',
            color: 'primary',
            hideDetails: 'auto'
        },
        VToolbar: {
            color: 'primary'
        },
        VTooltip: {
            location: 'top',
            zIndex: 9000
        },
        VWindow: {
            touch: false
        },
        VWindowItem: {
            transition: false,
            reverseTransition: false
        }
    },
    theme: {
        defaultTheme: 'light',
        themes: {
            light: {
                dark: false,
                colors: {
                    'primary': APP_THEME.light.primary,
                    'primary-darken-1': APP_THEME.light.primaryDarken,
                    'on-primary': APP_THEME.light.onPrimary,
                    'secondary': '#8c8c8c',
                    'secondary-darken-1': '#595754',
                    'on-secondary': '#ffffff',
                    'success': '#4cd964',
                    'success-darken-1': '#40b654',
                    'on-success': '#ffffff',
                    'info': '#2196f3',
                    'info-darken-1': '#1e85d7',
                    'on-info': '#ffffff',
                    'warning': '#ff9500',
                    'warning-darken-1': '#de8201',
                    'on-warning': '#ffffff',
                    'error': '#ff3b30',
                    'error-darken-1': '#e1342b',
                    'on-error': '#ffffff',
                    'teal': '#009688',
                    'background': APP_THEME.light.background,
                    'on-background': APP_THEME.light.text,
                    'surface': APP_THEME.light.surface,
                    'on-surface': APP_THEME.light.text,
                    'notification-background': APP_THEME.light.surface,
                    'on-notification-background': APP_THEME.light.text,
                    'grey': '#8c8c8c',
                    'grey-50': '#fafafa',
                    'grey-100': '#f0f2f8',
                    'grey-200': '#eeeeee',
                    'grey-300': '#e0e0e0',
                    'grey-400': '#bdbdbd',
                    'grey-500': '#9e9e9e',
                    'grey-600': '#757575',
                    'grey-700': '#616161',
                    'grey-800': '#424242',
                    'grey-900': '#212121',
                    'skin-bordered-background': APP_THEME.light.background,
                    'skin-bordered-surface': APP_THEME.light.surface,
                    'expansion-panel-text-custom-bg': APP_THEME.light.tint
                },
                variables: {
                    'code-color': '#ff8000',
                    'overlay-scrim-background': APP_THEME.light.shadow,
                    'tooltip-background': '#212121',
                    'tooltip-color': '#ffffff',
                    'overlay-scrim-opacity': 0.2,
                    'hover-opacity': 0.08,
                    'focus-opacity': 0.1,
                    'selected-opacity': 0.08,
                    'activated-opacity': 0.08,
                    'pressed-opacity': 0.14,
                    'dragged-opacity': 0.1,
                    'disabled-opacity': 0.4,
                    'border-color': APP_THEME.light.text,
                    'border-opacity': 0.12,
                    'table-header-background': APP_THEME.light.background,
                    'high-emphasis-opacity': 0.9,
                    'medium-emphasis-opacity': 0.7,

                    // shadows
                    'shadow-key-umbra-color': APP_THEME.light.shadow,
                    'shadow-xs-opacity': '0.16',
                    'shadow-sm-opacity': '0.18',
                    'shadow-md-opacity': '0.20',
                    'shadow-lg-opacity': '0.22',
                    'shadow-xl-opacity': '0.24',
                }
            },
            dark: {
                dark: true,
                colors: {
                    'primary': APP_THEME.dark.primary,
                    'primary-darken-1': APP_THEME.dark.primaryDarken,
                    'on-primary': APP_THEME.dark.onPrimary,
                    'secondary': '#9d9b99',
                    'secondary-darken-1': '#3e3d3c',
                    'on-secondary': '#fff',
                    'success': '#4cd964',
                    'success-darken-1': '#40b654',
                    'on-success': '#ffffff',
                    'info': '#2196f3',
                    'info-darken-1': '#1e85d7',
                    'on-info': '#ffffff',
                    'warning': '#ff9500',
                    'warning-darken-1': '#de8201',
                    'on-warning': '#ffffff',
                    'error': '#ff3b30',
                    'error-darken-1': '#e1342b',
                    'on-error': '#ffffff',
                    'teal': '#009688',
                    'background': APP_THEME.dark.background,
                    'on-background': APP_THEME.dark.text,
                    'surface': APP_THEME.dark.surface,
                    'on-surface': APP_THEME.dark.text,
                    'notification-background': APP_THEME.dark.surface,
                    'on-notification-background': APP_THEME.dark.text,
                    'grey': '#4d4c4b',
                    'grey-50': '#212121',
                    'grey-100': '#424242',
                    'grey-200': '#616161',
                    'grey-300': '#757575',
                    'grey-400': '#909090',
                    'grey-500': '#a2a2a2',
                    'grey-600': '#b4b4b4',
                    'grey-700': '#c6c6c6',
                    'grey-800': '#d8d8d8',
                    'grey-900': '#eaeaea',
                    'skin-bordered-background': APP_THEME.dark.background,
                    'skin-bordered-surface': APP_THEME.dark.surface,
                    'expansion-panel-text-custom-bg': APP_THEME.dark.tint
                },
                variables: {
                    'code-color': '#ff8000',
                    'overlay-scrim-background': APP_THEME.dark.shadow,
                    'tooltip-background': '#333333',
                    'tooltip-color': '#eeeeee',
                    'overlay-scrim-opacity': 0.2,
                    'hover-opacity': 0.12,
                    'focus-opacity': 0.1,
                    'selected-opacity': 0.08,
                    'activated-opacity': 0.08,
                    'pressed-opacity': 0.14,
                    'disabled-opacity': 0.4,
                    'dragged-opacity': 0.1,
                    'border-color': APP_THEME.dark.text,
                    'border-opacity': 0.12,
                    'table-header-background': APP_THEME.dark.background,
                    'high-emphasis-opacity': 0.9,
                    'medium-emphasis-opacity': 0.7,

                    // Shadows
                    'shadow-key-umbra-color': APP_THEME.dark.shadow,
                    'shadow-xs-opacity': '0.20',
                    'shadow-sm-opacity': '0.22',
                    'shadow-md-opacity': '0.24',
                    'shadow-lg-opacity': '0.26',
                    'shadow-xl-opacity': '0.28',
                }
            }
        }
    },
    locale: {
        rtl: getRtlLocales(),
        adapter: ((i18nGlobal: Composer) => {
            const instance: LocaleInstance = {
                name: 'ezBookkeeping i18n',
                decimalSeparator: ref<string>(DecimalSeparator.Default.symbol), // should never use vuetify to format numbers
                numericGroupSeparator: ref<string>(DigitGroupingSymbol.Default.symbol), // should never use vuetify to format numbers
                messages: i18nGlobal.messages,
                current: i18nGlobal.locale,
                fallback: i18nGlobal.locale, // no need to let vuetify know what fallback locale is
                t: (key: string, ...params: unknown[]): string => {
                    if (!key) {
                        return '';
                    }

                    if (!key.startsWith('$vuetify.')) {
                        return key;
                    }

                    key = key.substring(9); // remove '$vuetify.' prefix
                    const mappedTextKey = vuetifyI18nTextKeyMap[key];

                    if (!mappedTextKey) {
                        return key;
                    }

                    if (params && params.length > 0) {
                        // @ts-expect-error the arguments passed in are compatible with vue-i18n method arguments
                        return i18nGlobal.t(mappedTextKey, ...params);
                    } else {
                        return i18nGlobal.t(mappedTextKey);
                    }
                },
                n: (value: number): string => {
                    return i18nGlobal.n(value);
                },
                provide: (): LocaleInstance => {
                    return instance;
                }
            };

            return instance;
        })(i18nGlobal) as LocaleInstance,
    }
});
}
