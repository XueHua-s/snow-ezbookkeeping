package cmd

import (
	clis "github.com/mayswind/ezbookkeeping/pkg/cli"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"os"
)

func checkUserTransactionAndAccount(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")

	log.CliInfof(c, "[user_data.checkUserTransactionAndAccount] starting checking user \"%s\" data", username)

	_, err = clis.UserData.CheckTransactionAndAccount(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.checkUserTransactionAndAccount] error occurs when checking user data")
		return err
	}

	log.CliInfof(c, "[user_data.checkUserTransactionAndAccount] user transactions and accounts data has been checked successfully, there is no problem with user data")

	return nil
}

func fixTransactionTagIndexNotHaveTransactionTime(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")

	log.CliInfof(c, "[user_data.fixTransactionTagIndexNotHaveTransactionTime] starting fixing user \"%s\" transaction tag index data", username)

	_, err = clis.UserData.FixTransactionTagIndexWithTransactionTime(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.fixTransactionTagIndexNotHaveTransactionTime] error occurs when fixing user data")
		return err
	}

	log.CliInfof(c, "[user_data.fixTransactionTagIndexNotHaveTransactionTime] user transaction tag index data has been fixed successfully")

	return nil
}

func exportUserTransaction(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	filePath := c.String("file")
	fileType := c.String("type")

	if fileType == "" {
		fileType = "csv"
	}

	if fileType != "csv" && fileType != "tsv" {
		log.CliErrorf(c, "[user_data.exportUserTransaction] export file type is not supported")
		return errs.ErrNotSupported
	}

	if filePath == "" {
		log.CliErrorf(c, "[user_data.exportUserTransaction] export file path is unspecified")
		return os.ErrNotExist
	}

	fileExists, err := utils.IsExists(filePath)

	if fileExists {
		log.CliErrorf(c, "[user_data.exportUserTransaction] specified file path already exists")
		return os.ErrExist
	}

	log.CliInfof(c, "[user_data.exportUserTransaction] starting exporting user \"%s\" data", username)

	content, err := clis.UserData.ExportTransaction(c, username, fileType)

	if err != nil {
		log.CliErrorf(c, "[user_data.exportUserTransaction] error occurs when exporting user data")
		return err
	}

	err = utils.WriteFile(filePath, content)

	if err != nil {
		log.CliErrorf(c, "[user_data.exportUserTransaction] failed to write to %s", filePath)
		return err
	}

	log.CliInfof(c, "[user_data.exportUserTransaction] user transactions have been exported to %s", filePath)

	return nil
}

func importUserTransaction(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	filePath := c.String("file")
	filetype := c.String("type")

	if filePath == "" {
		log.CliErrorf(c, "[user_data.importUserTransaction] import file path is not specified")
		return os.ErrNotExist
	}

	fileExists, err := utils.IsExists(filePath)

	if !fileExists {
		log.CliErrorf(c, "[user_data.importUserTransaction] import file does not exist")
		return os.ErrExist
	}

	if filetype != "ezbookkeeping_csv" && filetype != "ezbookkeeping_tsv" {
		log.CliErrorf(c, "[user_data.importUserTransaction] unknown file type \"%s\"", filetype)
		return errs.ErrImportFileTypeNotSupported
	}

	data, err := os.ReadFile(filePath)

	if err != nil {
		log.CliErrorf(c, "[user_data.importUserTransaction] failed to load import file")
		return err
	}

	log.CliInfof(c, "[user_data.importUserTransaction] start importing transactions to user \"%s\"", username)

	err = clis.UserData.ImportTransaction(c, username, filetype, data)

	if err != nil {
		log.CliErrorf(c, "[user_data.importUserTransaction] error occurs when importing user data")
		return err
	}

	log.CliInfof(c, "[user_data.importUserTransaction] transactions have been imported to user \"%s\"", username)

	return nil
}
