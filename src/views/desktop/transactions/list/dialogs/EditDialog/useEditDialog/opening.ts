import { ScheduledTemplateFrequencyType, TemplateType } from '@/core/template.ts';
import { TransactionType } from '@/core/transaction.ts';
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
import type { useEditDialogLocation } from './location.ts';
import type { useEditDialogPictures } from './pictures.ts';
import type { useEditDialogRecognition } from './recognition.ts';
import type { useEditDialogState } from './state.ts';
import { type TransactionEditOptions, type TransactionEditResponse } from './types.ts';

interface Context {
    addByTemplateId: ReturnType<typeof useEditDialogState>['addByTemplateId'];
    duplicateFromId: ReturnType<typeof useEditDialogState>['duplicateFromId'];
    showState: ReturnType<typeof useEditDialogState>['showState'];
    activeTab: ReturnType<typeof useEditDialogState>['activeTab'];
    loading: ReturnType<typeof useEditDialogState>['loading'];
    submitting: ReturnType<typeof useEditDialogState>['submitting'];
    submitted: ReturnType<typeof useEditDialogState>['submitted'];
    geoLocationStatus: ReturnType<typeof useEditDialogState>['geoLocationStatus'];
    setGeoLocationByClickMap: ReturnType<typeof useEditDialogState>['setGeoLocationByClickMap'];
    originalTransactionEditable: ReturnType<typeof useEditDialogState>['originalTransactionEditable'];
    noTransactionDraft: ReturnType<typeof useEditDialogState>['noTransactionDraft'];
    initOptions: ReturnType<typeof useEditDialogState>['initOptions'];
    createNewTransactionModel: ReturnType<typeof useEditDialogState>['createNewTransactionModel'];
    setTransactionModel: ReturnType<typeof useEditDialogState>['setTransactionModel'];
    initTransaction: ReturnType<typeof useEditDialogState>['initTransaction'];
    transaction: ReturnType<typeof useEditDialogState>['transaction'];
    accountsStore: ReturnType<typeof useEditDialogState>['accountsStore'];
    transactionCategoriesStore: ReturnType<typeof useEditDialogState>['transactionCategoriesStore'];
    transactionTagsStore: ReturnType<typeof useEditDialogState>['transactionTagsStore'];
    props: SetupProps<{
    type: TransactionEditPageType;
    persistent?: boolean;
    show?: boolean;
}>;
    mode: ReturnType<typeof useEditDialogState>['mode'];
    editId: ReturnType<typeof useEditDialogState>['editId'];
    transactionsStore: ReturnType<typeof useEditDialogState>['transactionsStore'];
    settingsStore: ReturnType<typeof useEditDialogState>['settingsStore'];
    updateGeoLocation: ReturnType<typeof useEditDialogLocation>['updateGeoLocation'];
    initTemplate: ReturnType<typeof useEditDialogState>['initTemplate'];
    transactionTemplatesStore: ReturnType<typeof useEditDialogState>['transactionTemplatesStore'];
    clientSessionId: ReturnType<typeof useEditDialogState>['clientSessionId'];
    rejectFunc: ReturnType<typeof useEditDialogState>['rejectFunc'];
    uploadPicture: ReturnType<typeof useEditDialogPictures>['uploadPicture'];
    pastedText: ReturnType<typeof useEditDialogState>['pastedText'];
    recognizeText: ReturnType<typeof useEditDialogRecognition>['recognizeText'];
    showPasteTextDialog: ReturnType<typeof useEditDialogState>['showPasteTextDialog'];
    resolveFunc: ReturnType<typeof useEditDialogState>['resolveFunc'];
}

export function useEditDialogOpening(context: Context) {
    const {
        addByTemplateId, duplicateFromId, showState, activeTab, loading, submitting, submitted, geoLocationStatus,
        setGeoLocationByClickMap, originalTransactionEditable, noTransactionDraft, initOptions, createNewTransactionModel,
        setTransactionModel, initTransaction, transaction, accountsStore, transactionCategoriesStore, transactionTagsStore, props,
        mode, editId, transactionsStore, settingsStore, updateGeoLocation, initTemplate, transactionTemplatesStore,
        clientSessionId, rejectFunc, uploadPicture, pastedText, recognizeText, showPasteTextDialog, resolveFunc,
    } = context;

    function open(options: TransactionEditOptions): Promise<TransactionEditResponse | undefined> {
        addByTemplateId.value = null;
        duplicateFromId.value = null;
        showState.value = true;
        activeTab.value = 'basicInfo';
        loading.value = true;
        submitting.value = false;
        submitted.value = false;
        geoLocationStatus.value = null;
        setGeoLocationByClickMap.value = false;
        originalTransactionEditable.value = false;
        noTransactionDraft.value = options.noTransactionDraft || false;

        initOptions.value = options;

        const newTransaction = createNewTransactionModel(options.type);
        setTransactionModel(newTransaction, options, true);
        initTransaction.value = Transaction.of(transaction.value);

        const promises: Promise<unknown>[] = [
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false })
        ];

        if (props.type === TransactionEditPageType.Transaction) {
            if (options && options.id) {
                if (options.currentTransaction) {
                    setTransactionModel(options.currentTransaction, options, true);
                }

                mode.value = TransactionEditPageMode.View;
                editId.value = options.id;

                promises.push(transactionsStore.getTransaction({ transactionId: editId.value }));
            } else {
                mode.value = TransactionEditPageMode.Add;
                editId.value = null;

                if (options.template) {
                    setTransactionModel(options.template, options, false);
                    initTransaction.value = Transaction.of(transaction.value);
                    addByTemplateId.value = options.template.id;
                } else if (!options.noTransactionDraft && (settingsStore.appSettings.autoSaveTransactionDraft === 'enabled' || settingsStore.appSettings.autoSaveTransactionDraft === 'confirmation') && transactionsStore.transactionDraft) {
                    setTransactionModel(Transaction.ofDraft(transactionsStore.transactionDraft), options, false);
                }

                if (settingsStore.appSettings.autoGetCurrentGeoLocation
                    && !geoLocationStatus.value && !transaction.value.geoLocation) {
                    updateGeoLocation(false);
                }
            }
        } else if (props.type === TransactionEditPageType.Template) {
            initTemplate.value = TransactionTemplate.createNewTransactionTemplate(transaction.value);
            initTemplate.value.name = '';

            if (options && options.templateType) {
                initTemplate.value.templateType = options.templateType;
            }

            if (initTemplate.value.templateType === TemplateType.Schedule.type) {
                initTemplate.value.scheduledFrequencyType = ScheduledTemplateFrequencyType.Disabled.type;
                initTemplate.value.scheduledFrequency = '';
            }

            transaction.value = TransactionTemplate.ofTemplate(initTemplate.value);

            if (options && options.id) {
                if (options.currentTemplate) {
                    setTransactionModel(options.currentTemplate, options, false);
                    (transaction.value as TransactionTemplate).fillFrom(options.currentTemplate);
                }

                mode.value = TransactionEditPageMode.Edit;
                editId.value = options.id;
                transaction.value.id = options.id;

                promises.push(transactionTemplatesStore.getTemplate({ templateId: editId.value }));
            } else {
                mode.value = TransactionEditPageMode.Add;
                editId.value = null;
                transaction.value.id = '';
            }
        }

        if (options.type &&
            options.type >= TransactionType.Income &&
            options.type <= TransactionType.Transfer) {
            transaction.value.type = options.type;
        }

        if (mode.value === TransactionEditPageMode.Add) {
            clientSessionId.value = generateRandomUUID();
        }

        Promise.all(promises).then(function (responses) {
            if (editId.value && !responses[3]) {
                if (rejectFunc.value) {
                    if (props.type === TransactionEditPageType.Transaction) {
                        rejectFunc.value('Unable to retrieve transaction');
                    } else if (props.type === TransactionEditPageType.Template) {
                        rejectFunc.value('Unable to retrieve template');
                    }
                }

                return;
            }

            if (props.type === TransactionEditPageType.Transaction && options && options.id && responses[3] && responses[3] instanceof Transaction) {
                const transaction: Transaction = responses[3];
                setTransactionModel(transaction, options, true);
                initTransaction.value = Transaction.of(transaction);
                originalTransactionEditable.value = transaction.editable;
            } else if (props.type === TransactionEditPageType.Template && options && options.id && responses[3] && responses[3] instanceof TransactionTemplate) {
                const template: TransactionTemplate = responses[3];
                setTransactionModel(template, options, false);

                if (!(transaction.value instanceof TransactionTemplate)) {
                    transaction.value = TransactionTemplate.createNewTransactionTemplate(transaction.value);
                }

                (transaction.value as TransactionTemplate).fillFrom(template);
                initTemplate.value = TransactionTemplate.ofTemplate(template);
            } else {
                setTransactionModel(null, options, true);

                if (props.type === TransactionEditPageType.Transaction) {
                    initTransaction.value = Transaction.of(transaction.value);
                } else if (props.type === TransactionEditPageType.Template && transaction.value instanceof TransactionTemplate) {
                    initTemplate.value = TransactionTemplate.ofTemplate(transaction.value);
                }
            }

            if (options.autoUploadPicture) {
                uploadPicture(options.autoUploadPicture);
            }

            loading.value = false;

            if (isDefined(options.autoRecognizeClipboardText)) {
                pastedText.value = options.autoRecognizeClipboardText;

                if (pastedText.value && !settingsStore.appSettings.alwaysRequireConfirmationOfClipboardContentBeforeSubmission) {
                    recognizeText(pastedText.value);
                } else {
                    showPasteTextDialog.value = true;
                }
            }
        }).catch(error => {
            logger.error('failed to load essential data for editing transaction', error);

            loading.value = false;
            showState.value = false;

            if (!error.processed) {
                if (rejectFunc.value) {
                    rejectFunc.value(error);
                }
            }
        });

        return new Promise((resolve, reject) => {
            resolveFunc.value = resolve;
            rejectFunc.value = reject;
        });
    }

    return { open };
}
