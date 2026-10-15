package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/mayswind/ezbookkeeping/pkg/api"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/middlewares"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// registerAPIRoutes owns API authentication middleware and versioned endpoint bindings.
func registerAPIRoutes(router *gin.Engine, config *settings.Config) {
	apiRoute := router.Group("/api")

	apiRoute.Use(bindMiddleware(middlewares.RequestId(config), config))
	apiRoute.Use(bindMiddleware(middlewares.RequestLog, config))
	{
		if config.EnableInternalAuth {
			apiRoute.POST("/authorize.json", bindApiWithTokenUpdate(api.Authorizations.AuthorizeHandler, config))
		}

		if config.EnableInternalAuth && config.EnableTwoFactor {
			twoFactorRoute := apiRoute.Group("/2fa")
			twoFactorRoute.Use(bindMiddleware(middlewares.JWTTwoFactorAuthorization(config), config))
			{
				twoFactorRoute.POST("/authorize.json", bindApiWithTokenUpdate(api.Authorizations.TwoFactorAuthorizeHandler, config))
				twoFactorRoute.POST("/recovery.json", bindApiWithTokenUpdate(api.Authorizations.TwoFactorAuthorizeByRecoveryCodeHandler, config))
			}
		}

		if config.EnableOAuth2Login {
			oauth2Route := apiRoute.Group("/oauth2")
			oauth2Route.Use(bindMiddleware(middlewares.JWTOAuth2CallbackAuthorization(config), config))
			{
				oauth2Route.POST("/authorize.json", bindApiWithTokenUpdate(api.Authorizations.OAuth2CallbackAuthorizeHandler, config))
			}
		}

		if config.EnableInternalAuth && config.EnableUserRegister {
			apiRoute.POST("/register.json", bindApiWithTokenUpdate(api.Users.UserRegisterHandler, config))
		}

		if config.EnableUserVerifyEmail {
			apiRoute.POST("/verify_email/resend.json", bindApi(api.Users.UserSendVerifyEmailByUnloginUserHandler, config))

			emailVerifyRoute := apiRoute.Group("/verify_email")
			emailVerifyRoute.Use(bindMiddleware(middlewares.JWTEmailVerifyAuthorization(config), config))
			{
				emailVerifyRoute.POST("/by_token.json", bindApi(api.Users.UserEmailVerifyHandler, config))
			}
		}

		if config.EnableInternalAuth && config.EnableUserForgetPassword {
			apiRoute.POST("/forget_password/request.json", bindApi(api.ForgetPasswords.UserForgetPasswordRequestHandler, config))

			resetPasswordRoute := apiRoute.Group("/forget_password/reset")
			resetPasswordRoute.Use(bindMiddleware(middlewares.JWTResetPasswordAuthorization(config), config))
			{
				resetPasswordRoute.POST("/by_token.json", bindApi(api.ForgetPasswords.UserResetPasswordHandler, config))
			}
		}

		apiRoute.GET("/logout.json", bindApiWithTokenUpdate(api.Tokens.TokenRevokeCurrentHandler, config))

		apiV1Route := apiRoute.Group("/v1")
		apiV1Route.Use(bindMiddleware(middlewares.JWTAuthorizationByHeader(config), config))
		apiV1Route.Use(bindMiddleware(middlewares.APITokenIpLimit(config), config))
		{
			// Tokens
			apiV1Route.GET("/tokens/list.json", bindApi(api.Tokens.TokenListHandler, config))
			apiV1Route.POST("/tokens/generate/api.json", bindApi(api.Tokens.TokenGenerateAPIHandler, config))
			apiV1Route.POST("/tokens/generate/mcp.json", bindApi(api.Tokens.TokenGenerateMCPHandler, config))
			apiV1Route.POST("/tokens/revoke.json", bindApi(api.Tokens.TokenRevokeHandler, config))
			apiV1Route.POST("/tokens/revoke_all.json", bindApi(api.Tokens.TokenRevokeAllHandler, config))
			apiV1Route.POST("/tokens/refresh.json", bindApiWithTokenUpdate(api.Tokens.TokenRefreshHandler, config))

			// Users
			apiV1Route.GET("/users/profile/get.json", bindApi(api.Users.UserProfileHandler, config))
			apiV1Route.POST("/users/profile/update.json", bindApiWithTokenUpdate(api.Users.UserUpdateProfileHandler, config))

			if config.AvatarProvider == core.USER_AVATAR_PROVIDER_INTERNAL {
				apiV1Route.POST("/users/avatar/update.json", bindApi(api.Users.UserUpdateAvatarHandler, config))
				apiV1Route.POST("/users/avatar/remove.json", bindApi(api.Users.UserRemoveAvatarHandler, config))
			}

			if config.EnableUserVerifyEmail {
				apiV1Route.POST("/users/verify_email/resend.json", bindApi(api.Users.UserSendVerifyEmailByLoginedUserHandler, config))
			}

			// External Authentications
			if config.EnableOAuth2Login {
				apiV1Route.GET("/users/external_auth/list.json", bindApi(api.UserExternalAuths.ExternalAuthListHandler, config))
				apiV1Route.POST("/users/external_auth/unlink.json", bindApi(api.UserExternalAuths.UnlinkExternalAuthHandler, config))
			}

			// Application Cloud Settings
			apiV1Route.GET("/users/settings/cloud/get.json", bindApi(api.UserApplicationCloudSettings.ApplicationSettingsGetHandler, config))
			apiV1Route.POST("/users/settings/cloud/update.json", bindApi(api.UserApplicationCloudSettings.ApplicationSettingsUpdateHandler, config))
			apiV1Route.POST("/users/settings/cloud/disable.json", bindApi(api.UserApplicationCloudSettings.ApplicationSettingsDisableHandler, config))

			// Two-Factor Authorization
			if config.EnableTwoFactor {
				apiV1Route.GET("/users/2fa/status.json", bindApi(api.TwoFactorAuthorizations.TwoFactorStatusHandler, config))
				apiV1Route.POST("/users/2fa/enable/request.json", bindApi(api.TwoFactorAuthorizations.TwoFactorEnableRequestHandler, config))
				apiV1Route.POST("/users/2fa/enable/confirm.json", bindApiWithTokenUpdate(api.TwoFactorAuthorizations.TwoFactorEnableConfirmHandler, config))
				apiV1Route.POST("/users/2fa/disable.json", bindApi(api.TwoFactorAuthorizations.TwoFactorDisableHandler, config))
				apiV1Route.POST("/users/2fa/recovery/regenerate.json", bindApi(api.TwoFactorAuthorizations.TwoFactorRecoveryCodeRegenerateHandler, config))
			}

			// Data
			apiV1Route.GET("/data/statistics.json", bindApi(api.DataManagements.DataStatisticsHandler, config))
			apiV1Route.POST("/data/clear/all.json", bindApi(api.DataManagements.ClearAllDataHandler, config))
			apiV1Route.POST("/data/clear/transactions.json", bindApi(api.DataManagements.ClearAllTransactionsHandler, config))
			apiV1Route.POST("/data/clear/transactions/by_account.json", bindApi(api.DataManagements.ClearAllTransactionsByAccountHandler, config))

			if config.EnableDataExport {
				apiV1Route.GET("/data/export.csv", bindCsv(api.DataManagements.ExportDataToEzbookkeepingCSVHandler, config))
				apiV1Route.GET("/data/export.tsv", bindTsv(api.DataManagements.ExportDataToEzbookkeepingTSVHandler, config))
			}

			// Accounts
			apiV1Route.GET("/accounts/list.json", bindApi(api.Accounts.AccountListHandler, config))
			apiV1Route.GET("/accounts/get.json", bindApi(api.Accounts.AccountGetHandler, config))
			apiV1Route.POST("/accounts/add.json", bindApi(api.Accounts.AccountCreateHandler, config))
			apiV1Route.POST("/accounts/modify.json", bindApi(api.Accounts.AccountModifyHandler, config))
			apiV1Route.POST("/accounts/update/last_reconciled_time.json", bindApi(api.Accounts.AccountUpdateLastReconciledTimeHandler, config))
			apiV1Route.POST("/accounts/hide.json", bindApi(api.Accounts.AccountHideHandler, config))
			apiV1Route.POST("/accounts/move.json", bindApi(api.Accounts.AccountMoveHandler, config))
			apiV1Route.POST("/accounts/delete.json", bindApi(api.Accounts.AccountDeleteHandler, config))
			apiV1Route.POST("/accounts/sub_account/delete.json", bindApi(api.Accounts.SubAccountDeleteHandler, config))

			// Transactions
			apiV1Route.GET("/transactions/count.json", bindApi(api.Transactions.TransactionCountHandler, config))
			apiV1Route.GET("/transactions/list.json", bindApi(api.Transactions.TransactionListHandler, config))
			apiV1Route.GET("/transactions/list/by_month.json", bindApi(api.Transactions.TransactionMonthListHandler, config))
			apiV1Route.GET("/transactions/list/all.json", bindApi(api.Transactions.TransactionListAllHandler, config))
			apiV1Route.GET("/transactions/reconciliation_statements.json", bindApi(api.Transactions.TransactionReconciliationStatementHandler, config))
			apiV1Route.GET("/transactions/unreconciled_transaction_counts.json", bindApi(api.Transactions.TransactionUnreconciledTransactionCountsHandler, config))
			apiV1Route.GET("/transactions/statistics.json", bindApi(api.Transactions.TransactionStatisticsHandler, config))
			apiV1Route.GET("/transactions/statistics/trends.json", bindApi(api.Transactions.TransactionStatisticsTrendsHandler, config))
			apiV1Route.GET("/transactions/statistics/asset_trends.json", bindApi(api.Transactions.TransactionStatisticsAssetTrendsHandler, config))
			apiV1Route.GET("/transactions/amounts.json", bindApi(api.Transactions.TransactionAmountsHandler, config))
			apiV1Route.GET("/transactions/amounts/daily.json", bindApi(api.Transactions.TransactionDailyAmountsHandler, config))
			apiV1Route.GET("/transactions/get.json", bindApi(api.Transactions.TransactionGetHandler, config))
			apiV1Route.POST("/transactions/add.json", bindApi(api.Transactions.TransactionCreateHandler, config))
			apiV1Route.POST("/transactions/modify.json", bindApi(api.Transactions.TransactionModifyHandler, config))
			apiV1Route.POST("/transactions/batch_update/category.json", bindApi(api.Transactions.TransactionBatchUpdateCategoriesHandler, config))
			apiV1Route.POST("/transactions/batch_update/account.json", bindApi(api.Transactions.TransactionBatchUpdateAccountsHandler, config))
			apiV1Route.POST("/transactions/batch_update/tag/add.json", bindApi(api.Transactions.TransactionBatchAddTagsHandler, config))
			apiV1Route.POST("/transactions/batch_update/tag/remove.json", bindApi(api.Transactions.TransactionBatchRemoveTagsHandler, config))
			apiV1Route.POST("/transactions/batch_update/tag/clear.json", bindApi(api.Transactions.TransactionBatchClearTagsHandler, config))
			apiV1Route.POST("/transactions/move/all.json", bindApi(api.Transactions.TransactionMoveAllBetweenAccountsHandler, config))
			apiV1Route.POST("/transactions/delete.json", bindApi(api.Transactions.TransactionDeleteHandler, config))
			apiV1Route.POST("/transactions/batch_delete.json", bindApi(api.Transactions.TransactionBatchDeleteHandler, config))

			if config.EnableDataImport {
				apiV1Route.POST("/transactions/parse_custom_file.json", bindApi(api.Transactions.TransactionParseImportCustomFileDataHandler, config))
				apiV1Route.POST("/transactions/parse_import.json", bindApi(api.Transactions.TransactionParseImportFileHandler, config))
				apiV1Route.POST("/transactions/import.json", bindApi(api.Transactions.TransactionImportHandler, config))
				apiV1Route.GET("/transactions/import/process.json", bindApi(api.Transactions.TransactionImportProcessHandler, config))
			}

			// Transaction Pictures
			if config.EnableTransactionPictures {
				apiV1Route.POST("/transaction/pictures/upload.json", bindApi(api.TransactionPictures.TransactionPictureUploadHandler, config))
				apiV1Route.POST("/transaction/pictures/remove_unused.json", bindApi(api.TransactionPictures.TransactionPictureRemoveUnusedHandler, config))
			}

			// Transaction Categories
			apiV1Route.GET("/transaction/categories/list.json", bindApi(api.TransactionCategories.CategoryListHandler, config))
			apiV1Route.GET("/transaction/categories/get.json", bindApi(api.TransactionCategories.CategoryGetHandler, config))
			apiV1Route.POST("/transaction/categories/add.json", bindApi(api.TransactionCategories.CategoryCreateHandler, config))
			apiV1Route.POST("/transaction/categories/add_batch.json", bindApi(api.TransactionCategories.CategoryCreateBatchHandler, config))
			apiV1Route.POST("/transaction/categories/modify.json", bindApi(api.TransactionCategories.CategoryModifyHandler, config))
			apiV1Route.POST("/transaction/categories/hide.json", bindApi(api.TransactionCategories.CategoryHideHandler, config))
			apiV1Route.POST("/transaction/categories/move.json", bindApi(api.TransactionCategories.CategoryMoveHandler, config))
			apiV1Route.POST("/transaction/categories/delete.json", bindApi(api.TransactionCategories.CategoryDeleteHandler, config))

			// Transaction Tag Groups
			apiV1Route.GET("/transaction/tags/groups/list.json", bindApi(api.TransactionTagGroups.TagGroupListHandler, config))
			apiV1Route.GET("/transaction/tags/groups/get.json", bindApi(api.TransactionTagGroups.TagGroupGetHandler, config))
			apiV1Route.POST("/transaction/tags/groups/add.json", bindApi(api.TransactionTagGroups.TagGroupCreateHandler, config))
			apiV1Route.POST("/transaction/tags/groups/modify.json", bindApi(api.TransactionTagGroups.TagGroupModifyHandler, config))
			apiV1Route.POST("/transaction/tags/groups/move.json", bindApi(api.TransactionTagGroups.TagGroupMoveHandler, config))
			apiV1Route.POST("/transaction/tags/groups/delete.json", bindApi(api.TransactionTagGroups.TagGroupDeleteHandler, config))

			// Transaction Tags
			apiV1Route.GET("/transaction/tags/list.json", bindApi(api.TransactionTags.TagListHandler, config))
			apiV1Route.GET("/transaction/tags/get.json", bindApi(api.TransactionTags.TagGetHandler, config))
			apiV1Route.POST("/transaction/tags/add.json", bindApi(api.TransactionTags.TagCreateHandler, config))
			apiV1Route.POST("/transaction/tags/add_batch.json", bindApi(api.TransactionTags.TagCreateBatchHandler, config))
			apiV1Route.POST("/transaction/tags/modify.json", bindApi(api.TransactionTags.TagModifyHandler, config))
			apiV1Route.POST("/transaction/tags/hide.json", bindApi(api.TransactionTags.TagHideHandler, config))
			apiV1Route.POST("/transaction/tags/move.json", bindApi(api.TransactionTags.TagMoveHandler, config))
			apiV1Route.POST("/transaction/tags/delete.json", bindApi(api.TransactionTags.TagDeleteHandler, config))

			// Transaction Templates
			apiV1Route.GET("/transaction/templates/list.json", bindApi(api.TransactionTemplates.TemplateListHandler, config))
			apiV1Route.GET("/transaction/templates/get.json", bindApi(api.TransactionTemplates.TemplateGetHandler, config))
			apiV1Route.POST("/transaction/templates/add.json", bindApi(api.TransactionTemplates.TemplateCreateHandler, config))
			apiV1Route.POST("/transaction/templates/modify.json", bindApi(api.TransactionTemplates.TemplateModifyHandler, config))
			apiV1Route.POST("/transaction/templates/hide.json", bindApi(api.TransactionTemplates.TemplateHideHandler, config))
			apiV1Route.POST("/transaction/templates/move.json", bindApi(api.TransactionTemplates.TemplateMoveHandler, config))
			apiV1Route.POST("/transaction/templates/delete.json", bindApi(api.TransactionTemplates.TemplateDeleteHandler, config))

			// Insights Explorers
			apiV1Route.GET("/insights/explorers/list.json", bindApi(api.InsightsExplorers.InsightsExplorerListHandler, config))
			apiV1Route.GET("/insights/explorers/get.json", bindApi(api.InsightsExplorers.InsightsExplorerGetHandler, config))
			apiV1Route.POST("/insights/explorers/add.json", bindApi(api.InsightsExplorers.InsightsExplorerCreateHandler, config))
			apiV1Route.POST("/insights/explorers/modify.json", bindApi(api.InsightsExplorers.InsightsExplorerModifyHandler, config))
			apiV1Route.POST("/insights/explorers/hide.json", bindApi(api.InsightsExplorers.InsightsExplorerHideHandler, config))
			apiV1Route.POST("/insights/explorers/move.json", bindApi(api.InsightsExplorers.InsightsExplorerMoveHandler, config))
			apiV1Route.POST("/insights/explorers/delete.json", bindApi(api.InsightsExplorers.InsightsExplorerDeleteHandler, config))

			// Large Language Models
			if config.TextRecognitionLLMConfig != nil && config.TextRecognitionLLMConfig.LLMProvider != "" {
				if config.TransactionFromAITextRecognition {
					apiV1Route.POST("/llm/transactions/recognize_text.json", bindApi(api.LargeLanguageModels.RecognizeTransactionTextHandler, config))
				}
			}

			if config.EnableAIAssistant && config.AIAssistantLLMConfig != nil && config.AIAssistantLLMConfig.LLMProvider != "" {
				apiV1Route.POST("/llm/assistant/chat.json", bindApi(api.LargeLanguageModels.AssistantChatHandler, config))
				apiV1Route.POST("/llm/assistant/chat/stream.json", bindEventStreamApi(api.LargeLanguageModels.AssistantChatStreamHandler, config))
			}

			if config.ReceiptImageRecognitionLLMConfig != nil && config.ReceiptImageRecognitionLLMConfig.LLMProvider != "" {
				if config.TransactionFromAIImageRecognition {
					apiV1Route.POST("/llm/transactions/recognize_receipt_image.json", bindApi(api.LargeLanguageModels.RecognizeReceiptImageHandler, config))
				}
			}

			if config.CodingAssistantLLMConfig != nil && config.CodingAssistantLLMConfig.LLMProvider != "" {
				if config.InsightsExplorerCodingAssistant {
					apiV1Route.POST("/llm/code/generate.json", bindApi(api.LargeLanguageModels.GenerateCodeHandler, config))
				}
			}

			// User Custom Icons
			if config.EnableUserCustomIcon {
				apiV1Route.GET("/custom_icons/list.json", bindApi(api.UserCustomIcons.CustomIconListHandler, config))
				apiV1Route.POST("/custom_icons/upload.json", bindApi(api.UserCustomIcons.CustomIconUploadHandler, config))
				apiV1Route.POST("/custom_icons/move.json", bindApi(api.UserCustomIcons.CustomIconMoveHandler, config))
				apiV1Route.POST("/custom_icons/delete.json", bindApi(api.UserCustomIcons.CustomIconDeleteHandler, config))
			}

			// Exchange Rates
			apiV1Route.GET("/exchange_rates/latest.json", bindApi(api.ExchangeRates.LatestExchangeRateHandler, config))
			apiV1Route.POST("/exchange_rates/user_custom/update.json", bindApi(api.ExchangeRates.UserCustomExchangeRateUpdateHandler, config))
			apiV1Route.POST("/exchange_rates/user_custom/delete.json", bindApi(api.ExchangeRates.UserCustomExchangeRateDeleteHandler, config))

			// System
			apiV1Route.GET("/systems/version.json", bindApi(api.Systems.VersionHandler, config))
		}
	}

}
