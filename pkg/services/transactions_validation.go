package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

func (s *TransactionService) getAccountModels(sess *xorm.Session, transaction *models.Transaction) (sourceAccount *models.Account, destinationAccount *models.Account, err error) {
	sourceAccount = &models.Account{}
	destinationAccount = &models.Account{}

	has, err := sess.ID(transaction.AccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(sourceAccount)

	if err != nil {
		return nil, nil, err
	} else if !has {
		return nil, nil, errs.ErrSourceAccountNotFound
	}

	// check whether the related account is valid
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		if transaction.RelatedAccountId != 0 && transaction.RelatedAccountId != transaction.AccountId {
			return nil, nil, errs.ErrAccountIdInvalid
		} else {
			destinationAccount = sourceAccount
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME || transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
		if transaction.RelatedAccountId != 0 {
			return nil, nil, errs.ErrAccountIdInvalid
		}

		destinationAccount = nil
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		if transaction.RelatedAccountId <= 0 {
			return nil, nil, errs.ErrAccountIdInvalid
		} else {
			has, err = sess.ID(transaction.RelatedAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(destinationAccount)

			if err != nil {
				return nil, nil, err
			} else if !has {
				return nil, nil, errs.ErrDestinationAccountNotFound
			}
		}
	}

	// check whether the parent accounts are valid
	if sourceAccount.ParentAccountId > 0 && destinationAccount != nil && sourceAccount.ParentAccountId != destinationAccount.ParentAccountId && destinationAccount.ParentAccountId > 0 {
		var accounts []*models.Account
		err := sess.Where("uid=? AND deleted=? and (account_id=? or account_id=?)", transaction.Uid, false, sourceAccount.ParentAccountId, destinationAccount.ParentAccountId).Find(&accounts)

		if err != nil {
			return nil, nil, err
		}

		if len(accounts) < 2 {
			return nil, nil, errs.ErrAccountNotFound
		}

		for i := 0; i < len(accounts); i++ {
			account := accounts[i]

			if account.Hidden {
				return nil, nil, errs.ErrCannotUseHiddenAccount
			}
		}
	} else if sourceAccount.ParentAccountId > 0 && (destinationAccount == nil || sourceAccount.ParentAccountId == destinationAccount.ParentAccountId || destinationAccount.ParentAccountId == 0) {
		sourceParentAccount := &models.Account{}
		has, err = sess.ID(sourceAccount.ParentAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(sourceParentAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrSourceAccountNotFound
		}

		if sourceParentAccount.Hidden {
			return nil, nil, errs.ErrCannotUseHiddenAccount
		}
	} else if sourceAccount.ParentAccountId == 0 && destinationAccount != nil && destinationAccount.ParentAccountId > 0 {
		destinationParentAccount := &models.Account{}
		has, err = sess.ID(destinationAccount.ParentAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(destinationParentAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrDestinationAccountNotFound
		}

		if destinationParentAccount.Hidden {
			return nil, nil, errs.ErrCannotUseHiddenAccount
		}
	}

	return sourceAccount, destinationAccount, nil
}

func (s *TransactionService) getOldAccountModels(sess *xorm.Session, transaction *models.Transaction, oldTransaction *models.Transaction, sourceAccount *models.Account, destinationAccount *models.Account) (oldSourceAccount *models.Account, oldDestinationAccount *models.Account, err error) {
	oldSourceAccount = &models.Account{}
	oldDestinationAccount = &models.Account{}

	if transaction.AccountId == oldTransaction.AccountId {
		oldSourceAccount = sourceAccount
	} else {
		has, err := sess.ID(oldTransaction.AccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(oldSourceAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrSourceAccountNotFound
		}
	}

	if transaction.RelatedAccountId == oldTransaction.RelatedAccountId {
		oldDestinationAccount = destinationAccount
	} else if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		has, err := sess.ID(oldTransaction.RelatedAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(oldDestinationAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrDestinationAccountNotFound
		}
	}

	return oldSourceAccount, oldDestinationAccount, nil
}

func (s *TransactionService) getRelatedUpdateColumns(updateCols []string) []string {
	relatedUpdateCols := make([]string, len(updateCols))

	for i := 0; i < len(updateCols); i++ {
		if updateCols[i] == "account_id" {
			relatedUpdateCols[i] = "related_account_id"
		} else if updateCols[i] == "related_account_id" {
			relatedUpdateCols[i] = "account_id"
		} else if updateCols[i] == "amount" {
			relatedUpdateCols[i] = "related_account_amount"
		} else if updateCols[i] == "related_account_amount" {
			relatedUpdateCols[i] = "amount"
		} else {
			relatedUpdateCols[i] = updateCols[i]
		}
	}

	return relatedUpdateCols
}

func (s *TransactionService) isAccountIdValid(transaction *models.Transaction) error {
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		if transaction.RelatedAccountId != 0 && transaction.RelatedAccountId != transaction.AccountId {
			return errs.ErrTransactionDestinationAccountCannotBeSet
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME ||
		transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
		if transaction.RelatedAccountId != 0 {
			return errs.ErrTransactionDestinationAccountCannotBeSet
		} else if transaction.RelatedAccountAmount != 0 {
			return errs.ErrTransactionDestinationAmountCannotBeSet
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		if transaction.AccountId == transaction.RelatedAccountId {
			return errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		return errs.ErrTransactionTypeInvalid
	} else {
		return errs.ErrTransactionTypeInvalid
	}

	return nil
}

func (s *TransactionService) isCategoryValid(sess *xorm.Session, transaction *models.Transaction) error {
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		if transaction.CategoryId != 0 {
			return errs.ErrBalanceModificationTransactionCannotSetCategory
		}
	} else {
		category := &models.TransactionCategory{}
		has, err := sess.ID(transaction.CategoryId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(category)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionCategoryNotFound
		}

		if category.Hidden {
			return errs.ErrCannotUseHiddenTransactionCategory
		}

		if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
			return errs.ErrCannotUsePrimaryCategoryForTransaction
		}

		if (transaction.Type == models.TRANSACTION_DB_TYPE_INCOME && category.Type != models.CATEGORY_TYPE_INCOME) ||
			(transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE && category.Type != models.CATEGORY_TYPE_EXPENSE) ||
			((transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) && category.Type != models.CATEGORY_TYPE_TRANSFER) {
			return errs.ErrTransactionCategoryTypeInvalid
		}

		parentCategory := &models.TransactionCategory{}
		has, err = sess.ID(category.ParentCategoryId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(parentCategory)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionCategoryNotFound
		}

		if parentCategory.Hidden {
			return errs.ErrCannotUseHiddenTransactionCategory
		}
	}

	return nil
}

func (s *TransactionService) isTagsValid(sess *xorm.Session, uid int64, transactionTagIndexes []*models.TransactionTagIndex, tagIds []int64) error {
	if len(transactionTagIndexes) > 0 {
		var tags []*models.TransactionTag
		err := sess.Where("uid=? AND deleted=?", uid, false).In("tag_id", tagIds).Find(&tags)

		if err != nil {
			return err
		}

		tagMap := make(map[int64]*models.TransactionTag)

		for i := 0; i < len(tags); i++ {
			if tags[i].Hidden {
				return errs.ErrCannotUseHiddenTransactionTag
			}

			tagMap[tags[i].TagId] = tags[i]
		}

		for i := 0; i < len(transactionTagIndexes); i++ {
			if _, exists := tagMap[transactionTagIndexes[i].TagId]; !exists {
				return errs.ErrTransactionTagNotFound
			}
		}
	}

	return nil
}

func (s *TransactionService) isPicturesValid(sess *xorm.Session, transaction *models.Transaction, pictureIds []int64) error {
	if len(pictureIds) > 0 {
		var pictureInfos []*models.TransactionPictureInfo
		err := sess.Where("uid=? AND deleted=?", transaction.Uid, false).In("picture_id", pictureIds).Find(&pictureInfos)

		if err != nil {
			return err
		}

		pictureInfoMap := make(map[int64]*models.TransactionPictureInfo)

		for i := 0; i < len(pictureInfos); i++ {
			if pictureInfos[i].TransactionId != models.TransactionPictureNewPictureTransactionId && pictureInfos[i].TransactionId != transaction.TransactionId {
				return errs.ErrTransactionPictureIdInvalid
			}

			pictureInfoMap[pictureInfos[i].PictureId] = pictureInfos[i]
		}

		for i := 0; i < len(pictureIds); i++ {
			if _, exists := pictureInfoMap[pictureIds[i]]; !exists {
				return errs.ErrTransactionPictureNotFound
			}
		}
	}

	return nil
}
