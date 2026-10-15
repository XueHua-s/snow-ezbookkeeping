package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

const pageCountForClearTransactions = 1000

const pageCountForDataExport = 1000

// DataManagementsApi represents data management api
type DataManagementsApi struct {
	ApiUsingConfig
	tokens                  *services.TokenService
	users                   *services.UserService
	accounts                *services.AccountService
	transactions            *services.TransactionService
	categories              *services.TransactionCategoryService
	tags                    *services.TransactionTagService
	tagGroups               *services.TransactionTagGroupService
	pictures                *services.TransactionPictureService
	templates               *services.TransactionTemplateService
	userCustomIcons         *services.UserCustomIconService
	userCustomExchangeRates *services.UserCustomExchangeRatesService
	insightsExploreres      *services.InsightsExplorerService
}

// Initialize a data management api singleton instance
var (
	DataManagements = &DataManagementsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		tokens:                  services.Tokens,
		users:                   services.Users,
		accounts:                services.Accounts,
		transactions:            services.Transactions,
		categories:              services.TransactionCategories,
		tags:                    services.TransactionTags,
		tagGroups:               services.TransactionTagGroups,
		pictures:                services.TransactionPictures,
		templates:               services.TransactionTemplates,
		userCustomIcons:         services.UserCustomIcons,
		userCustomExchangeRates: services.UserCustomExchangeRates,
		insightsExploreres:      services.InsightsExplorers,
	}
)

// ExportDataToEzbookkeepingCSVHandler returns exported data in csv format
func (a *DataManagementsApi) ExportDataToEzbookkeepingCSVHandler(c *core.WebContext) ([]byte, string, *errs.Error) {
	return a.getExportedFileContent(c, "csv")
}

// ExportDataToEzbookkeepingTSVHandler returns exported data in csv format
func (a *DataManagementsApi) ExportDataToEzbookkeepingTSVHandler(c *core.WebContext) ([]byte, string, *errs.Error) {
	return a.getExportedFileContent(c, "tsv")
}

// DataStatisticsHandler returns user data statistics
func (a *DataManagementsApi) DataStatisticsHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	totalAccountCount, err := a.accounts.GetTotalAccountCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total account count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalTransactionCategoryCount, err := a.categories.GetTotalCategoryCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total transaction category count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalTransactionTagCount, err := a.tags.GetTotalTagCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total transaction tag count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalTransactionCount, err := a.transactions.GetTotalTransactionCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total transaction count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalTransactionPictureCount, err := a.pictures.GetTotalTransactionPicturesCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total transaction picture count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalExplorationCount, err := a.insightsExploreres.GetTotalExplorationsCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total exploration count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalTransactionTemplateCount, err := a.templates.GetTotalNormalTemplateCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total transaction template count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalScheduledTransactionCount, err := a.templates.GetTotalScheduledTemplateCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total scheduled transaction count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	totalCustomIconCount, err := a.userCustomIcons.GetTotalCustomIconsCountByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.DataStatisticsHandler] failed to get total custom icon count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	dataStatisticsResp := &models.DataStatisticsResponse{
		TotalAccountCount:              totalAccountCount,
		TotalTransactionCategoryCount:  totalTransactionCategoryCount,
		TotalTransactionTagCount:       totalTransactionTagCount,
		TotalTransactionCount:          totalTransactionCount,
		TotalTransactionPictureCount:   totalTransactionPictureCount,
		TotalExplorationCount:          totalExplorationCount,
		TotalTransactionTemplateCount:  totalTransactionTemplateCount,
		TotalScheduledTransactionCount: totalScheduledTransactionCount,
		TotalCustomIconCount:           totalCustomIconCount,
	}

	return dataStatisticsResp, nil
}
