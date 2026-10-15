package api

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/converters"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math"
	"strings"
	"time"
)

func (a *DataManagementsApi) getExportedFileContent(c *core.WebContext, fileType string) ([]byte, string, *errs.Error) {
	if !a.CurrentConfig().EnableDataExport {
		return nil, "", errs.ErrDataExportNotAllowed
	}

	var exportTransactionDataReq models.ExportTransactionDataRequest
	err := c.ShouldBindQuery(&exportTransactionDataReq)

	if err != nil {
		log.Warnf(c, "[data_managements.getExportedFileContent] parse request failed, because %s", err.Error())
		return nil, "", errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[data_managements.getExportedFileContent] cannot get client timezone, because %s", err.Error())
		clientTimezone = time.Local
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[data_managements.getExportedFileContent] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, "", errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_EXPORT_TRANSACTION) {
		return nil, "", errs.ErrNotPermittedToPerformThisAction
	}

	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	categories, err := a.categories.GetAllCategoriesByUid(c, uid, 0, -1)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	tags, err := a.tags.GetAllTagsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to get tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	tagIndexes, err := a.tags.GetAllTagIdsMapOfAllTransactions(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to get tag index for user \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	accountMap := a.accounts.GetAccountMapByList(accounts)
	categoryMap := a.categories.GetCategoryMapByList(categories)
	tagMap := a.tags.GetTagMapByList(tags)

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, exportTransactionDataReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[data_managements.getExportedFileContent] get account error, because %s", err.Error())
		return nil, "", errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.categories.GetCategoryOrSubCategoryIds(c, exportTransactionDataReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[data_managements.getExportedFileContent] get transaction category error, because %s", err.Error())
		return nil, "", errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := exportTransactionDataReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(exportTransactionDataReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[data_managements.getExportedFileContent] parse transaction tag filters error, because %s", err.Error())
			return nil, "", errs.Or(err, errs.ErrOperationFailed)
		}
	}

	maxTransactionTime := int64(math.MaxInt64)
	minTransactionTime := int64(0)

	if exportTransactionDataReq.MaxTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(exportTransactionDataReq.MaxTime)
	}

	if exportTransactionDataReq.MinTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(exportTransactionDataReq.MinTime)
	}

	allTransactions, err := a.transactions.GetAllSpecifiedTransactions(c, uid, maxTransactionTime, minTransactionTime, exportTransactionDataReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, exportTransactionDataReq.AmountFilter, exportTransactionDataReq.Keyword, exportTransactionDataReq.MatchMode, false, pageCountForDataExport, true)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to all transactions user \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	dataExporter := converters.GetTransactionDataExporter(fileType)

	if dataExporter == nil {
		return nil, "", errs.ErrNotImplemented
	}

	result, err := dataExporter.ToExportedContent(c, uid, allTransactions, accountMap, categoryMap, tagMap, tagIndexes)

	if err != nil {
		log.Errorf(c, "[data_managements.getExportedFileContent] failed to get exported data for \"uid:%d\", because %s", uid, err.Error())
		return nil, "", errs.Or(err, errs.ErrOperationFailed)
	}

	fileName := a.getFileName(user, clientTimezone, fileType)

	return result, fileName, nil
}

func (a *DataManagementsApi) getFileName(user *models.User, clientTimezone *time.Location, fileExtension string) string {
	currentTime := utils.FormatUnixTimeToLongDateTimeWithoutSecond(time.Now().Unix(), clientTimezone)
	currentTime = strings.Replace(currentTime, "-", "_", -1)
	currentTime = strings.Replace(currentTime, " ", "_", -1)
	currentTime = strings.Replace(currentTime, ":", "_", -1)

	return fmt.Sprintf("%s_%s.%s", user.Username, currentTime, fileExtension)
}
