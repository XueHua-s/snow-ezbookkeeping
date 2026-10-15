import { ScheduledTemplateFrequencyType, TemplateType } from '@/core/template.ts';
import {
    TransactionType
} from '@/core/transaction.ts';
import { isDefined } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import type { SetupProps } from '@/lib/ui/props.ts';
import { Transaction } from '@/models/transaction.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { Router } from 'framework7/types';
import type { useEditPagePictures } from './pictures.ts';
import type { useEditPageRecognition } from './recognition.ts';
import type { useEditPageState } from './state.ts';

interface Context {
    pageTypeAndMode: ReturnType<typeof useEditPageState>['pageTypeAndMode'];
    showToast: ReturnType<typeof useEditPageState>['showToast'];
    loadingError: ReturnType<typeof useEditPageState>['loadingError'];
    loading: ReturnType<typeof useEditPageState>['loading'];
    accountsStore: ReturnType<typeof useEditPageState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useEditPageState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useEditPageState>['transactionTagsStore'];
    transactionTemplatesStore: ReturnType<typeof useEditPageState>['transactionTemplatesStore'];
    query: ReturnType<typeof useEditPageState>['query'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    editId: ReturnType<typeof useEditPageState>['editId'];
    duplicateFromId: ReturnType<typeof useEditPageState>['duplicateFromId'];
    transactionsStore: ReturnType<typeof useEditPageState>['transactionsStore'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    getQueryTransactionOptions: ReturnType<typeof useEditPageState>['getQueryTransactionOptions'];
    clientSessionId: ReturnType<typeof useEditPageState>['clientSessionId'];
    addByTemplateId: ReturnType<typeof useEditPageState>['addByTemplateId'];
    settingsStore: ReturnType<typeof useEditPageState>['settingsStore'];
    setTransactionModel: ReturnType<typeof useEditPageState>['setTransactionModel'];
    props: SetupProps<{
    f7route: Router.Route;
    f7router: Router.Router;
    autoUploadPicture?: File;
    autoRecognizeClipboardText?: string;
}>;
    showTransactionPictures: ReturnType<typeof useEditPageState>['showTransactionPictures'];
    uploadPicture: ReturnType<typeof useEditPagePictures>['uploadPicture'];
    pastedText: ReturnType<typeof useEditPageState>['pastedText'];
    recognizeText: ReturnType<typeof useEditPageRecognition>['recognizeText'];
    showAITextRecognitionSheet: ReturnType<typeof useEditPageState>['showAITextRecognitionSheet'];
}

export function useEditPageInitialization(context: Context) {
    const {
        pageTypeAndMode, showToast, loadingError, loading, accountsStore, transactionCategoriesStore, transactionTagsStore,
        transactionTemplatesStore, query, mode, editId, duplicateFromId, transactionsStore, transaction,
        getQueryTransactionOptions, clientSessionId, addByTemplateId, settingsStore, setTransactionModel, props,
        showTransactionPictures, uploadPicture, pastedText, recognizeText, showAITextRecognitionSheet,
    } = context;

    function init(): void {
        if (!pageTypeAndMode) {
            showToast('Parameter Invalid');
            loadingError.value = 'Parameter Invalid';
            return;
        }

        loading.value = true;

        const promises: Promise<unknown>[] = [
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false }),
            transactionTemplatesStore.loadAllTemplates({ force: false, templateType: TemplateType.Normal.type })
        ];

        if (pageTypeAndMode.type === TransactionEditPageType.Transaction) {
            if (query['id']) {
                if (mode.value === TransactionEditPageMode.Edit) {
                    editId.value = query['id'];
                } else if (mode.value === TransactionEditPageMode.Add) {
                    duplicateFromId.value = query['id'];
                }

                promises.push(transactionsStore.getTransaction({ transactionId: query['id'], withPictures: mode.value !== TransactionEditPageMode.Add }));
            }
        } else if (pageTypeAndMode.type === TransactionEditPageType.Template) {
            const template = TransactionTemplate.createNewTransactionTemplate(transaction.value);
            template.name = '';

            if (query['templateType']) {
                template.templateType = parseInt(query['templateType']);
            }

            if (template.templateType === TemplateType.Schedule.type) {
                template.scheduledFrequencyType = ScheduledTemplateFrequencyType.Disabled.type;
                template.scheduledFrequency = '';
            }

            transaction.value = template;

            if (query['id']) {
                if (mode.value === TransactionEditPageMode.Edit) {
                    editId.value = query['id'];
                }

                promises.push(transactionTemplatesStore.getTemplate({ templateId: query['id'] }));
            }
        }

        const initOptions = getQueryTransactionOptions();

        if (initOptions.type &&
            initOptions.type >= TransactionType.Income &&
            initOptions.type <= TransactionType.Transfer) {
            transaction.value.type = initOptions.type;
        } else if (initOptions.type === TransactionType.ModifyBalance &&
            pageTypeAndMode.type === TransactionEditPageType.Transaction &&
            mode.value === TransactionEditPageMode.View) {
            transaction.value.type = initOptions.type;
        }

        if (mode.value === TransactionEditPageMode.Add) {
            clientSessionId.value = generateRandomUUID();
        }

        Promise.all(promises).then(function (responses) {
            if (query['id'] && !responses[4]) {
                if (pageTypeAndMode.type === TransactionEditPageType.Transaction) {
                    showToast('Unable to retrieve transaction');
                    loadingError.value = 'Unable to retrieve transaction';
                } else if (pageTypeAndMode.type === TransactionEditPageType.Template) {
                    showToast('Unable to retrieve template');
                    loadingError.value = 'Unable to retrieve template';
                }

                return;
            }

            let fromTransaction: Transaction | TransactionTemplate | null = null;

            if (pageTypeAndMode.type === TransactionEditPageType.Transaction) {
                if (query['id'] && responses[4] instanceof Transaction) {
                    fromTransaction = responses[4];
                } else if (query['templateId'] && transactionTemplatesStore.allTransactionTemplatesMap && transactionTemplatesStore.allTransactionTemplatesMap[TemplateType.Normal.type]) {
                    fromTransaction = (transactionTemplatesStore.allTransactionTemplatesMap[TemplateType.Normal.type] as Record<string, TransactionTemplate>)[query['templateId']] ?? null;

                    if (fromTransaction) {
                        addByTemplateId.value = fromTransaction.id;
                    }
                } else if (query['noTransactionDraft'] !== 'true' && (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled' || settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') && transactionsStore.transactionDraft) {
                    fromTransaction = Transaction.ofDraft(transactionsStore.transactionDraft);
                }
            } else if (pageTypeAndMode.type === TransactionEditPageType.Template && responses[4] instanceof TransactionTemplate) {
                if (query['id']) {
                    fromTransaction = responses[4];
                }
            }

            setTransactionModel(
                fromTransaction,
                initOptions,
                pageTypeAndMode.type === TransactionEditPageType.Transaction && (mode.value === TransactionEditPageMode.Edit || mode.value === TransactionEditPageMode.View)
            );

            if (pageTypeAndMode.type === TransactionEditPageType.Transaction && query['id'] && responses[4] instanceof Transaction) {
                if (fromTransaction && query['withTime'] && query['withTime'] === 'true') {
                    transaction.value.time = fromTransaction.time;
                    transaction.value.timeZone = fromTransaction.timeZone;
                    transaction.value.utcOffset = fromTransaction.utcOffset;
                }

                if (fromTransaction && query['withGeoLocation'] && query['withGeoLocation'] === 'true') {
                    transaction.value.setGeoLocation(fromTransaction.geoLocation);
                }
            } else if (pageTypeAndMode.type === TransactionEditPageType.Template && query['id'] && responses[4] instanceof TransactionTemplate) {
                const template = responses[4];
                transaction.value.id = template.id;

                if (!(transaction.value instanceof TransactionTemplate)) {
                    transaction.value = TransactionTemplate.createNewTransactionTemplate(transaction.value);
                }

                (transaction.value as TransactionTemplate).fillFrom(template);
            }

            if (props.autoUploadPicture) {
                showTransactionPictures.value = true;
                uploadPicture(props.autoUploadPicture);
            }

            loading.value = false;

            if (isDefined(props.autoRecognizeClipboardText)) {
                pastedText.value = props.autoRecognizeClipboardText;

                if (pastedText.value && !settingsStore.appSettings.alwaysRequireConfirmationOfClipboardContentBeforeSubmission) {
                    recognizeText(pastedText.value);
                } else {
                    showAITextRecognitionSheet.value = true;
                }
            }
        }).catch(error => {
            logger.error('failed to load essential data for editing transaction', error);

            if (error.processed) {
                loading.value = false;
            } else {
                loadingError.value = error;
                showToast(error.message || error);
            }
        });
    }

    return { init };
}
